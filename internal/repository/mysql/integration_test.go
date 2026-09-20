package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/statusengine/interface/internal/domain"
	"github.com/statusengine/interface/internal/metrics/mysqlprov"
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

func TestIntegrationHistory(t *testing.T) {
	db := testDB(t)
	r := NewHistory(db)
	now := time.Now().Unix()

	// Six hours, because that is the widest unscoped window the API
	// allows. A thirty-day unscoped query is a shape the product
	// forbids, and measuring it only proves that forbidding it was
	// right.
	window := HistoryFilter{From: now - 6*3600, To: now}

	// Each subtest gets its own deadline. Sharing one across a dozen
	// queries against a large table turns a slow first read on a cold
	// buffer pool into a dozen confusing failures.
	ctxFor := func(t *testing.T) context.Context {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		t.Cleanup(cancel)
		return ctx
	}

	t.Run("checks", func(t *testing.T) {
		ctx := ctxFor(t)
		rows, _, err := r.Checks(ctx, window, page("start_time", true))
		if err != nil {
			t.Fatalf("Checks: %v", err)
		}
		for _, c := range rows {
			if c.StartTime < window.From || c.StartTime > window.To {
				t.Errorf("check at %d is outside the window", c.StartTime)
			}
			if c.StateText == "" {
				t.Error("check has no state_text")
			}
		}
		for name := range CheckSortColumns {
			if _, _, err := r.Checks(ctx, window, page(name, true)); err != nil {
				t.Errorf("sorting checks by %s: %v", name, err)
			}
		}
	})

	t.Run("state changes", func(t *testing.T) {
		ctx := ctxFor(t)
		if _, _, err := r.StateChanges(ctx, window, page("state_time", true)); err != nil {
			t.Fatalf("StateChanges: %v", err)
		}
		for name := range StateChangeSortColumns {
			if _, _, err := r.StateChanges(ctx, window, page(name, true)); err != nil {
				t.Errorf("sorting state changes by %s: %v", name, err)
			}
		}
	})

	t.Run("notifications", func(t *testing.T) {
		ctx := ctxFor(t)
		rows, _, err := r.Notifications(ctx, window, page("start_time", true))
		if err != nil {
			t.Fatalf("Notifications: %v", err)
		}
		for _, n := range rows {
			if n.Reason == "" {
				t.Error("notification has no decoded reason")
			}
		}
		for name := range NotificationSortColumns {
			if _, _, err := r.Notifications(ctx, window, page(name, true)); err != nil {
				t.Errorf("sorting notifications by %s: %v", name, err)
			}
		}
	})

	// hard_only and transitions_only touch columns that exist on some of
	// these tables and not others, which is exactly the kind of thing
	// only a real database catches.
	t.Run("optional predicates parse on every table", func(t *testing.T) {
		ctx := ctxFor(t)
		hard := window
		hard.HardOnly = true
		if _, _, err := r.Checks(ctx, hard, page("start_time", true)); err != nil {
			t.Errorf("hard_only on checks: %v", err)
		}
		if _, _, err := r.StateChanges(ctx, hard, page("state_time", true)); err != nil {
			t.Errorf("hard_only on state changes: %v", err)
		}
		// The notification tables have no is_hardstate column; the
		// handler clears the flag, and the repository must not add it.
		if _, _, err := r.Notifications(ctx, window, page("start_time", true)); err != nil {
			t.Errorf("notifications: %v", err)
		}

		transitions := window
		transitions.TransitionsOnly = true
		if _, _, err := r.StateChanges(ctx, transitions, page("state_time", true)); err != nil {
			t.Errorf("transitions_only: %v", err)
		}
	})

	t.Run("scoping to one object", func(t *testing.T) {
		ctx := ctxFor(t)

		// Whatever this database actually has, rather than a name that
		// only exists on one installation.
		all, _, err := r.Checks(ctx, window, page("start_time", true))
		if err != nil {
			t.Fatalf("Checks: %v", err)
		}
		if len(all) == 0 {
			t.Skip("no checks in the window")
		}
		sample := all[0]

		scoped := window
		scoped.Host = sample.Hostname
		scoped.Description = sample.Description
		if !scoped.Scoped() {
			t.Error("a filter naming a host should report itself as scoped")
		}

		rows, _, err := r.Checks(ctx, scoped, page("start_time", true))
		if err != nil {
			t.Fatalf("Checks: %v", err)
		}
		for _, c := range rows {
			if c.Hostname != sample.Hostname || c.Description != sample.Description {
				t.Errorf("got %s/%s, want %s/%s",
					c.Hostname, c.Description, sample.Hostname, sample.Description)
			}
		}
	})
}

func TestIntegrationMetrics(t *testing.T) {
	p := mysqlprov.New(testDB(t))
	ctx := ctxFor(t)
	now := time.Now().Unix()

	// Find a service that actually has performance data.
	services := NewServices(testDB(t))
	all, total, err := services.List(ctx, StatusFilter{}, page("hostname", false))
	if err != nil {
		t.Fatalf("listing services: %v", err)
	}
	if total == 0 {
		t.Skip("no services in the test database")
	}

	var withData []domain.MetricMeta
	var host, description string
	for _, s := range all {
		labels, err := p.Labels(ctx, s.Hostname, s.Description)
		if err != nil {
			t.Fatalf("Labels(%q, %q): %v", s.Hostname, s.Description, err)
		}
		if len(labels) > 0 {
			withData, host, description = labels, s.Hostname, s.Description
			break
		}
	}
	if withData == nil {
		t.Skip("no performance data in the test database")
	}

	for _, meta := range withData {
		if meta.LastSeen < meta.FirstSeen {
			t.Errorf("%s: last_seen is before first_seen", meta.Label)
		}
	}

	t.Run("downsamples to the budget", func(t *testing.T) {
		const budget = 20
		result, err := p.Query(ctx, domain.MetricQuery{
			Hostname: host, Description: description,
			From: now - 86400, To: now, MaxPoints: budget,
		})
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if result.Source != "mysql" {
			t.Errorf("source = %q", result.Source)
		}
		if result.BucketSeconds <= 0 {
			t.Errorf("bucket = %d", result.BucketSeconds)
		}
		for _, series := range result.Series {
			if len(series.Points) > budget+1 {
				t.Errorf("%s: %d points for a budget of %d", series.Label, len(series.Points), budget)
			}
			for _, point := range series.Points {
				// The bucket boundary has to be a whole timestamp; a
				// DECIMAL here is what a FLOOR(a/b) would produce.
				if point.Time%result.BucketSeconds != 0 {
					t.Errorf("%s: point at %d is not aligned to a %ds bucket",
						series.Label, point.Time, result.BucketSeconds)
				}
				if point.Min > point.Avg || point.Avg > point.Max {
					t.Errorf("%s: min %v, avg %v, max %v are out of order",
						series.Label, point.Min, point.Avg, point.Max)
				}
			}
		}
	})

	t.Run("filters to one label", func(t *testing.T) {
		want := withData[0].Label
		result, err := p.Query(ctx, domain.MetricQuery{
			Hostname: host, Description: description,
			Labels: []string{want}, From: now - 86400, To: now, MaxPoints: 500,
		})
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		for _, series := range result.Series {
			if series.Label != want {
				t.Errorf("asked for %q, got %q", want, series.Label)
			}
		}
	})

	t.Run("a service with no data is empty, not an error", func(t *testing.T) {
		result, err := p.Query(ctx, domain.MetricQuery{
			Hostname: "no-such-host-9f3a", Description: "nothing",
			From: now - 3600, To: now, MaxPoints: 500,
		})
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if len(result.Series) != 0 {
			t.Errorf("got %d series for a host that does not exist", len(result.Series))
		}
	})
}
