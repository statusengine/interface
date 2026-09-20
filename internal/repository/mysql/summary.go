package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Summary produces the dashboard's counts.
type Summary struct {
	db *sql.DB
}

// NewSummary returns a repository backed by db.
func NewSummary(db *sql.DB) *Summary { return &Summary{db: db} }

// Get counts both populations in one pass each.
//
// These are aggregates over the two status tables, which hold one row per
// object rather than per event - a few thousand rows on a large
// installation, not a few million. Counting them directly is cheaper and
// far less surprising than maintaining a cache that can go stale.
func (r *Summary) Get(ctx context.Context) (domain.Summary, error) {
	var out domain.Summary

	hosts, err := r.counts(ctx, "statusengine_hoststatus", hostStateNames)
	if err != nil {
		return out, err
	}
	services, err := r.counts(ctx, "statusengine_servicestatus", serviceStateNames)
	if err != nil {
		return out, err
	}
	out.Hosts = hosts.StateCounts
	out.Services = services.StateCounts

	if hosts.LastUpdate() > services.LastUpdate() {
		out.LastUpdate = hosts.LastUpdate()
	} else {
		out.LastUpdate = services.LastUpdate()
	}

	nodes, err := r.nodes(ctx)
	if err != nil {
		return out, err
	}
	out.Nodes = nodes
	return out, nil
}

var hostStateNames = map[int]string{0: "up", 1: "down", 2: "unreachable"}
var serviceStateNames = map[int]string{0: "ok", 1: "warning", 2: "critical", 3: "unknown"}

// stateCounts carries the newest status_update_time alongside the counts
// so the caller can say how fresh the dashboard is.
type stateCounts struct {
	domain.StateCounts
	newest int64
}

func (s stateCounts) LastUpdate() int64 { return s.newest }

func (r *Summary) counts(ctx context.Context, table string, names map[int]string) (stateCounts, error) {
	var out stateCounts
	out.ByState = make(map[string]int64, len(names))
	for _, name := range names {
		out.ByState[name] = 0
	}

	// One scan, every number. A separate query per tile would be six
	// round trips and six chances for the tiles to disagree with each
	// other by a few seconds.
	query := `SELECT
		COUNT(*),
		SUM(last_check = 0),
		SUM(last_check > 0 AND COALESCE(current_state, 0) <> 0),
		SUM(last_check > 0 AND COALESCE(current_state, 0) <> 0
		    AND COALESCE(problem_has_been_acknowledged, 0) = 0
		    AND COALESCE(scheduled_downtime_depth, 0) = 0),
		SUM(COALESCE(problem_has_been_acknowledged, 0) = 1),
		SUM(COALESCE(scheduled_downtime_depth, 0) > 0),
		SUM(COALESCE(is_flapping, 0) = 1),
		SUM(COALESCE(notifications_enabled, 0) = 0),
		SUM(COALESCE(active_checks_enabled, 0) = 0),
		COALESCE(MAX(status_update_time), 0)
		FROM ` + table

	// SUM over an empty table is NULL, so every total needs a nullable
	// destination even though the values are counts.
	var pending, problems, unhandled, acked, downtime, flapping, notifOff, activeOff sql.NullInt64
	err := r.db.QueryRowContext(ctx, query).Scan(
		&out.Total, &pending, &problems, &unhandled, &acked, &downtime,
		&flapping, &notifOff, &activeOff, &out.newest)
	if err != nil {
		return out, fmt.Errorf("summarising %s: %w", table, err)
	}

	out.Pending = pending.Int64
	out.Problems = problems.Int64
	out.Unhandled = unhandled.Int64
	out.Acknowledged = acked.Int64
	out.InDowntime = downtime.Int64
	out.Flapping = flapping.Int64
	out.NotificationsDisabled = notifOff.Int64
	out.ActiveChecksDisabled = activeOff.Int64

	byState, err := r.byState(ctx, table, names)
	if err != nil {
		return out, err
	}
	for name, n := range byState {
		out.ByState[name] = n
	}
	return out, nil
}

func (r *Summary) byState(ctx context.Context, table string, names map[int]string) (map[string]int64, error) {
	query := "SELECT COALESCE(current_state, 0), COUNT(*) FROM " + table +
		" WHERE last_check > 0 GROUP BY COALESCE(current_state, 0)"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("summarising %s by state: %w", table, err)
	}
	defer rows.Close()

	out := make(map[string]int64)
	for rows.Next() {
		var state int
		var n int64
		if err := rows.Scan(&state, &n); err != nil {
			return nil, fmt.Errorf("summarising %s by state: %w", table, err)
		}
		name, ok := names[state]
		if !ok {
			// A state the core invented that this build does not know.
			// Counting it under its number beats dropping it, because a
			// total that does not add up is worse than an odd label.
			name = fmt.Sprintf("state_%d", state)
		}
		out[name] += n
	}
	return out, rows.Err()
}

// nodes reports the monitoring cores feeding this database.
//
// statusengine_nodes is populated by the worker but can be empty on an
// installation that predates it, so the names come from the status tables
// themselves, which always carry them.
func (r *Summary) nodes(ctx context.Context) ([]domain.Node, error) {
	const query = `SELECT name,
		SUM(hosts) AS hosts, SUM(services) AS services, MAX(last_update) AS last_update
		FROM (
			SELECT COALESCE(node_name, '') AS name, COUNT(*) AS hosts, 0 AS services,
			       COALESCE(MAX(status_update_time), 0) AS last_update
			FROM statusengine_hoststatus GROUP BY COALESCE(node_name, '')
			UNION ALL
			SELECT COALESCE(node_name, '') AS name, 0 AS hosts, COUNT(*) AS services,
			       COALESCE(MAX(status_update_time), 0) AS last_update
			FROM statusengine_servicestatus GROUP BY COALESCE(node_name, '')
		) AS n
		GROUP BY name ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	defer rows.Close()

	var out []domain.Node
	for rows.Next() {
		var n domain.Node
		if err := rows.Scan(&n.Name, &n.Hosts, &n.Services, &n.LastUpdate); err != nil {
			return nil, fmt.Errorf("listing nodes: %w", err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
