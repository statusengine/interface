package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// The audit query is the one piece of SQL in this package, and the parts
// worth proving - the LIKE escape, the failed-only clause, the direction
// of the ORDER BY - are parts only a database can confirm.
//
// Skipped unless SEI_TEST_DSN points at a database that already has the
// sei_* schema. The test writes its own rows, marks them with a username
// nothing else uses, and deletes exactly those again.
func auditForTest(t *testing.T) (*Audit, string) {
	t.Helper()

	dsn := os.Getenv("SEI_TEST_DSN")
	if dsn == "" {
		t.Skip("set SEI_TEST_DSN to run the audit integration test")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("opening the test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connecting to the test database: %v", err)
	}
	var name string
	err = db.QueryRowContext(ctx, "SHOW TABLES LIKE 'sei_command_audit'").Scan(&name)
	if err != nil {
		t.Skip("sei_command_audit is not present; run the migrations against SEI_TEST_DSN first")
	}

	// Unique per run, so two runs against the same database cannot see
	// each other's rows and a crashed run leaves nothing that a later
	// one will count.
	user := fmt.Sprintf("audit-it-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sei_command_audit WHERE username = ?", user)
	})
	return NewAudit(db), user
}

func TestIntegrationAuditListFiltersAndOrders(t *testing.T) {
	audit, user := auditForTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	entries := []Entry{
		{Username: user, Action: ActionAcknowledge, Target: "web01", HTTPStatus: 202, Response: "accepted"},
		{Username: user, Action: ActionScheduleDowntime, Target: "web01/HTTP", HTTPStatus: 202, Response: "accepted"},
		{Username: user, Action: ActionReschedule, Target: "db%_01", HTTPStatus: 503, Response: "the worker is not reachable"},
	}
	for _, e := range entries {
		e.Payload = map[string]any{"comment": "integration test"}
		if err := audit.Record(ctx, e); err != nil {
			t.Fatalf("recording an entry: %v", err)
		}
	}

	all, total, err := audit.List(ctx, AuditFilter{Username: user}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if total != 3 || len(all) != 3 {
		t.Fatalf("got %d rows and total %d, want 3 and 3", len(all), total)
	}
	if all[0].Payload == nil {
		t.Error("the payload came back empty; it is the reason the log is worth reading")
	}

	// Newest first by default, and the tiebreaker has to decide: all
	// three rows were written in the same second.
	asc, _, err := audit.List(ctx, AuditFilter{Username: user}, AuditPage{Limit: 50})
	if err != nil {
		t.Fatalf("listing ascending: %v", err)
	}
	if asc[0].ID >= all[0].ID {
		t.Errorf("ascending starts at %d and descending at %d; they should be opposite ends",
			asc[0].ID, all[0].ID)
	}

	failed, total, err := audit.List(ctx, AuditFilter{Username: user, FailedOnly: true}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("listing failures: %v", err)
	}
	if total != 1 || len(failed) != 1 || failed[0].HTTPStatus != 503 {
		t.Fatalf("failed-only returned %d rows (total %d), want the one 503", len(failed), total)
	}

	byAction, _, err := audit.List(ctx,
		AuditFilter{Username: user, Action: string(ActionAcknowledge)}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("filtering by action: %v", err)
	}
	if len(byAction) != 1 || byAction[0].Target != "web01" {
		t.Fatalf("action filter returned %v, want the one acknowledgement", byAction)
	}

	// A target holding LIKE's own wildcards must be searchable for
	// literally, or a search for db%_01 returns half the log.
	search, total, err := audit.List(ctx,
		AuditFilter{Username: user, Search: "db%_01"}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("searching: %v", err)
	}
	if total != 1 || len(search) != 1 || search[0].Target != "db%_01" {
		t.Fatalf("search returned %d rows, want the one whose target contains the wildcards", len(search))
	}

	// A window that ends before the rows were written excludes them.
	old, total, err := audit.List(ctx,
		AuditFilter{Username: user, From: 1, To: 1000}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("listing an old window: %v", err)
	}
	if total != 0 || len(old) != 0 {
		t.Fatalf("a window in 1970 returned %d rows", len(old))
	}
}

func TestIntegrationAuditRecordBatchWritesOneRowPerTarget(t *testing.T) {
	audit, user := auditForTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var batch []Entry
	for _, target := range []string{"a", "b", "c", "d"} {
		batch = append(batch, Entry{
			Username: user, Action: ActionToggle, Target: target,
			HTTPStatus: 202, Response: "accepted",
			Payload: map[string]any{"switch": "notifications", "enable": false},
		})
	}
	if err := audit.RecordBatch(ctx, batch); err != nil {
		t.Fatalf("recording a batch: %v", err)
	}

	_, total, err := audit.List(ctx, AuditFilter{Username: user}, AuditPage{Limit: 50, Desc: true})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	if total != 4 {
		t.Fatalf("a bulk command over four objects wrote %d rows, want 4", total)
	}
}
