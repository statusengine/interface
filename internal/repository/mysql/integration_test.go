package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/statusengine/interface/internal/domain"
)

// These exercise the real SQL against a real Statusengine schema. A query
// builder test can prove the clause reads correctly; only a database can
// prove MySQL accepts it, that the column list and the Scan targets line
// up, and that a UNION of two tables with different shapes actually runs.
//
// Skipped unless SEI_TEST_DSN points at a database with the worker's
// tables. They only read.
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("SEI_TEST_DSN")
	if dsn == "" {
		t.Skip("set SEI_TEST_DSN to run the repository integration tests")
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
	return db
}

func ctxFor(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func page(sort string, desc bool) Page {
	return Page{Limit: 50, Offset: 0, Sort: sort, Desc: desc}
}

func TestIntegrationHosts(t *testing.T) {
	r := NewHosts(testDB(t))
	ctx := ctxFor(t)

	all, total, err := r.List(ctx, StatusFilter{}, page("hostname", false))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total == 0 {
		t.Skip("no hosts in the test database")
	}
	if int64(len(all)) > total {
		t.Errorf("got %d rows but a total of %d", len(all), total)
	}

	// Every sortable column has to be accepted by MySQL, including the
	// CASE expressions - a typo there is invisible until someone clicks
	// that header.
	for name := range HostSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.List(ctx, StatusFilter{}, page(name, true)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}

	t.Run("detail matches the list row", func(t *testing.T) {
		want := all[0]
		got, err := r.Get(ctx, want.Hostname)
		if err != nil {
			t.Fatalf("Get(%q): %v", want.Hostname, err)
		}
		if got.Hostname != want.Hostname || got.State != want.State {
			t.Errorf("detail says %q/%d, the list said %q/%d",
				got.Hostname, got.State, want.Hostname, want.State)
		}
		if got.StateText == "" {
			t.Error("state_text is empty; the client should not have to map numbers")
		}
	})

	t.Run("unknown host", func(t *testing.T) {
		if _, err := r.Get(ctx, "no-such-host-9f3a"); err != ErrNotFound {
			t.Errorf("Get of a missing host = %v, want ErrNotFound", err)
		}
	})

	t.Run("a wildcard in the search is a literal", func(t *testing.T) {
		_, withWildcard, err := r.List(ctx, StatusFilter{Search: "%"}, page("hostname", false))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if withWildcard == total && total > 0 {
			t.Error(`searching for "%" returned everything, so LIKE escaping is not working`)
		}
	})

	t.Run("problems only", func(t *testing.T) {
		rows, _, err := r.List(ctx, StatusFilter{ProblemsOnly: true}, page("hostname", false))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, h := range rows {
			if h.State == int(domain.HostUp) {
				t.Errorf("%s is UP but came back in a problems-only list", h.Hostname)
			}
		}
	})
}

func TestIntegrationServices(t *testing.T) {
	r := NewServices(testDB(t))
	ctx := ctxFor(t)

	all, total, err := r.List(ctx, StatusFilter{}, page("hostname", false))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total == 0 {
		t.Skip("no services in the test database")
	}

	for name := range ServiceSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.List(ctx, StatusFilter{}, page(name, true)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}

	// A Naemon service description is free text. This installation has
	// one called `C:\ Drive Space`; a round trip through the repository
	// must not mangle it.
	t.Run("detail round trip", func(t *testing.T) {
		want := all[0]
		got, err := r.Get(ctx, want.Hostname, want.Description)
		if err != nil {
			t.Fatalf("Get(%q, %q): %v", want.Hostname, want.Description, err)
		}
		if got.Description != want.Description {
			t.Errorf("description came back as %q, want %q", got.Description, want.Description)
		}
	})

	t.Run("filtered to one host", func(t *testing.T) {
		host := all[0].Hostname
		rows, _, err := r.List(ctx, StatusFilter{Host: host}, page("hostname", false))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, s := range rows {
			if s.Hostname != host {
				t.Errorf("filtered to %q but got %q", host, s.Hostname)
			}
		}
	})
}

func TestIntegrationProblems(t *testing.T) {
	r := NewProblems(testDB(t))
	ctx := ctxFor(t)

	rows, total, err := r.List(ctx, ProblemFilter{}, page("severity", true))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total == 0 {
		t.Skip("nothing is broken in the test database")
	}

	for _, p := range rows {
		if p.Kind != domain.KindHost && p.Kind != domain.KindService {
			t.Errorf("row has kind %q", p.Kind)
		}
		if p.State == 0 {
			t.Errorf("%s/%s is in an OK state but appears as a problem", p.Hostname, p.Description)
		}
		if p.StateText == "" {
			t.Errorf("%s/%s has no state_text", p.Hostname, p.Description)
		}
	}

	// The whole point of the severity sort: worst first.
	t.Run("severity descending puts hosts that are down first", func(t *testing.T) {
		if len(rows) < 2 {
			t.Skip("need at least two problems")
		}
		if rows[0].Kind == domain.KindService {
			for _, p := range rows {
				if p.Kind == domain.KindHost && p.State == int(domain.HostDown) {
					t.Error("a host that is down was sorted below a service")
					break
				}
			}
		}
	})

	t.Run("each kind on its own", func(t *testing.T) {
		hostsOnly, hostTotal, err := r.List(ctx, ProblemFilter{Kind: domain.KindHost}, page("severity", true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, p := range hostsOnly {
			if p.Kind != domain.KindHost {
				t.Errorf("asked for hosts, got a %s", p.Kind)
			}
		}

		svcOnly, svcTotal, err := r.List(ctx, ProblemFilter{Kind: domain.KindService}, page("severity", true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, p := range svcOnly {
			if p.Kind != domain.KindService {
				t.Errorf("asked for services, got a %s", p.Kind)
			}
		}
		if hostTotal+svcTotal != total {
			t.Errorf("hosts (%d) plus services (%d) is %d, but the union said %d",
				hostTotal, svcTotal, hostTotal+svcTotal, total)
		}
	})

	t.Run("hiding services of down hosts never grows the list", func(t *testing.T) {
		_, filtered, err := r.List(ctx, ProblemFilter{HideServicesOfDownHosts: true}, page("severity", true))
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if filtered > total {
			t.Errorf("hiding rows produced more of them: %d > %d", filtered, total)
		}
	})

	for name := range ProblemSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.List(ctx, ProblemFilter{}, page(name, false)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}
}

func TestIntegrationDowntimes(t *testing.T) {
	r := NewDowntimes(testDB(t))
	ctx := ctxFor(t)
	now := time.Now().Unix()

	if _, _, err := r.Current(ctx, DowntimeFilter{Now: now}, page("scheduled_start_time", true)); err != nil {
		t.Fatalf("Current: %v", err)
	}
	history, _, err := r.History(ctx, DowntimeFilter{Now: now}, page("scheduled_start_time", true))
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	for _, d := range history {
		if d.Kind != domain.KindHost && d.Kind != domain.KindService {
			t.Errorf("downtime has kind %q", d.Kind)
		}
	}

	running := true
	if _, _, err := r.Current(ctx, DowntimeFilter{Now: now, Running: &running}, page("scheduled_start_time", true)); err != nil {
		t.Errorf("filtering to running downtimes: %v", err)
	}

	for name := range DowntimeSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.History(ctx, DowntimeFilter{Now: now}, page(name, true)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}
}

func TestIntegrationAcknowledgements(t *testing.T) {
	r := NewAcknowledgements(testDB(t))
	ctx := ctxFor(t)

	if _, _, err := r.List(ctx, AckFilter{}, page("entry_time", true)); err != nil {
		t.Fatalf("List: %v", err)
	}
	for name := range AckSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.List(ctx, AckFilter{}, page(name, true)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}
}

func TestIntegrationLogEntries(t *testing.T) {
	r := NewLogEntries(testDB(t))
	ctx := ctxFor(t)

	now := time.Now().Unix()
	f := LogFilter{From: now - 7*86400, To: now}

	rows, _, err := r.List(ctx, f, page("entry_time", true))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, e := range rows {
		if e.EntryTime < f.From || e.EntryTime > f.To {
			t.Errorf("entry at %d is outside the window %d..%d", e.EntryTime, f.From, f.To)
		}
	}

	for name := range LogSortColumns {
		t.Run("sort by "+name, func(t *testing.T) {
			if _, _, err := r.List(ctx, f, page(name, true)); err != nil {
				t.Errorf("sorting by %s: %v", name, err)
			}
		})
	}
}

func TestIntegrationSummary(t *testing.T) {
	r := NewSummary(testDB(t))
	ctx := ctxFor(t)

	s, err := r.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	for label, counts := range map[string]domain.StateCounts{"hosts": s.Hosts, "services": s.Services} {
		var summed int64
		for _, n := range counts.ByState {
			summed += n
		}
		// by_state only counts what has been checked; pending is
		// everything else. Together they are the population.
		if summed+counts.Pending != counts.Total {
			t.Errorf("%s: by_state (%d) plus pending (%d) is %d, but total is %d",
				label, summed, counts.Pending, summed+counts.Pending, counts.Total)
		}
		if counts.Unhandled > counts.Problems {
			t.Errorf("%s: more unhandled (%d) than problems (%d)", label, counts.Unhandled, counts.Problems)
		}
	}
}
