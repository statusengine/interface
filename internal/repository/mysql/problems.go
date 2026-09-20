package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Problems reads the two status tables together.
type Problems struct {
	db *sql.DB
}

// NewProblems returns a repository backed by db.
func NewProblems(db *sql.DB) *Problems { return &Problems{db: db} }

// ProblemFilter narrows the triage list.
type ProblemFilter struct {
	Search string

	// Kind restricts to "host" or "service"; empty means both.
	Kind domain.Kind

	Acknowledged *bool
	InDowntime   *bool
	HardState    *bool
	Flapping     *bool

	// Handled is the distinction the list exists for: a problem that is
	// acknowledged or in a downtime is already somebody's. Unhandled is
	// what is actually waiting.
	Handled *bool

	// HideServicesOfDownHosts drops services whose host is down. They are
	// almost always a consequence rather than a separate incident, and
	// during an outage they bury everything else.
	HideServicesOfDownHosts bool
}

// ProblemSortColumns operates on the union's derived columns.
//
// `severity` is a single scale across both kinds, in the order an
// operator works through an incident: a host that is down first, then
// critical services, then unreachable hosts - which are usually a
// consequence of the first - then warnings, then unknowns.
var ProblemSortColumns = map[string]string{
	"severity":            "severity",
	"hostname":            "hostname",
	"service_description": "service_description",
	"state":               "state",
	"last_check":          "last_check",
	"last_state_change":   "last_state_change",
	"kind":                "kind",
}

const problemSeverityHost = `CASE current_state WHEN 1 THEN 5 WHEN 2 THEN 3 ELSE 0 END`
const problemSeverityService = `CASE s.current_state WHEN 2 THEN 4 WHEN 1 THEN 2 WHEN 3 THEN 1 ELSE 0 END`

// union builds the combined query body and its arguments, without an
// ORDER BY or LIMIT, so the count and the page can share it exactly.
func (f ProblemFilter) union() (string, []any) {
	var parts []string
	var args []any

	if f.Kind != domain.KindService {
		var c conditions
		c.add("current_state <> 0")
		c.add("last_check > 0") // a host never checked is pending, not a problem
		f.applyCommon(&c, "", []string{"hostname"})

		parts = append(parts, `SELECT 'host' AS kind, hostname, '' AS service_description,
			COALESCE(current_state, 0) AS state, COALESCE(is_hardstate, 0) AS is_hard,
			COALESCE(output, '') AS output,
			COALESCE(current_check_attempt, 0) AS current_check_attempt,
			COALESCE(max_check_attempts, 0) AS max_check_attempts,
			last_check, last_state_change,
			COALESCE(problem_has_been_acknowledged, 0) AS acknowledged,
			COALESCE(scheduled_downtime_depth, 0) AS downtime_depth,
			COALESCE(is_flapping, 0) AS is_flapping,
			COALESCE(notifications_enabled, 0) AS notifications_enabled,
			0 AS host_is_down,
			`+problemSeverityHost+` AS severity
			FROM statusengine_hoststatus`+c.clause())
		args = append(args, c.params()...)
	}

	if f.Kind != domain.KindHost {
		var c conditions
		c.add("s.current_state <> 0")
		c.add("s.last_check > 0")
		f.applyCommon(&c, "s.", []string{"s.hostname", "s.service_description"})
		if f.HideServicesOfDownHosts {
			c.add("COALESCE(h.current_state, 0) = 0")
		}

		parts = append(parts, `SELECT 'service' AS kind, s.hostname, s.service_description,
			COALESCE(s.current_state, 0) AS state, COALESCE(s.is_hardstate, 0) AS is_hard,
			COALESCE(s.output, '') AS output,
			COALESCE(s.current_check_attempt, 0) AS current_check_attempt,
			COALESCE(s.max_check_attempts, 0) AS max_check_attempts,
			s.last_check, s.last_state_change,
			COALESCE(s.problem_has_been_acknowledged, 0) AS acknowledged,
			COALESCE(s.scheduled_downtime_depth, 0) AS downtime_depth,
			COALESCE(s.is_flapping, 0) AS is_flapping,
			COALESCE(s.notifications_enabled, 0) AS notifications_enabled,
			CASE WHEN COALESCE(h.current_state, 0) <> 0 THEN 1 ELSE 0 END AS host_is_down,
			`+problemSeverityService+` AS severity
			FROM statusengine_servicestatus s
			LEFT JOIN statusengine_hoststatus h ON h.hostname = s.hostname`+c.clause())
		args = append(args, c.params()...)
	}

	if len(parts) == 0 {
		// Both kinds excluded. Return something that selects nothing but
		// still has the right shape, so the caller does not branch.
		return "SELECT 'host' AS kind, '' AS hostname, '' AS service_description, 0 AS state," +
			" 0 AS is_hard, '' AS output, 0 AS current_check_attempt, 0 AS max_check_attempts," +
			" 0 AS last_check, 0 AS last_state_change, 0 AS acknowledged, 0 AS downtime_depth," +
			" 0 AS is_flapping, 0 AS notifications_enabled, 0 AS host_is_down, 0 AS severity" +
			" FROM DUAL WHERE FALSE", nil
	}
	if len(parts) == 1 {
		return parts[0], args
	}
	return parts[0] + " UNION ALL " + parts[1], args
}

func (f ProblemFilter) applyCommon(c *conditions, prefix string, searchColumns []string) {
	if f.Search != "" {
		pattern := likeEscape(f.Search)
		var exprs []string
		var args []any
		for _, col := range searchColumns {
			exprs = append(exprs, col+" LIKE ? ESCAPE '\\\\'")
			args = append(args, pattern)
		}
		c.add("("+joinOr(exprs)+")", args...)
	}

	c.addBool(prefix+"problem_has_been_acknowledged", f.Acknowledged)
	c.addBool(prefix+"is_hardstate", f.HardState)
	c.addBool(prefix+"is_flapping", f.Flapping)

	if f.InDowntime != nil {
		if *f.InDowntime {
			c.add(prefix + "scheduled_downtime_depth > 0")
		} else {
			c.add(prefix + "scheduled_downtime_depth = 0")
		}
	}
	if f.Handled != nil {
		ack := prefix + "problem_has_been_acknowledged"
		dt := prefix + "scheduled_downtime_depth"
		if *f.Handled {
			c.add("(" + ack + " = 1 OR " + dt + " > 0)")
		} else {
			c.add(ack + " = 0 AND " + dt + " = 0")
		}
	}
}

// List returns a page of problems across hosts and services.
func (r *Problems) List(ctx context.Context, f ProblemFilter, p Page) ([]domain.Problem, int64, error) {
	body, args := f.union()

	var total int64
	countQuery := "SELECT COUNT(*) FROM (" + body + ") AS p"
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting problems: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	// Descending severity is the default: the list exists to be worked
	// from the top.
	query := "SELECT * FROM (" + body + ") AS p" +
		orderBy(p.Sort, p.Desc, ProblemSortColumns, "severity") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing problems: %w", err)
	}
	defer rows.Close()

	var out []domain.Problem
	for rows.Next() {
		var p domain.Problem
		var kind string
		var severity, downtimeDepth int
		if err := rows.Scan(&kind, &p.Hostname, &p.Description, &p.State, &p.IsHard,
			&p.Output, &p.CurrentAttempt, &p.MaxAttempts, &p.LastCheck, &p.LastStateChange,
			&p.Acknowledged, &downtimeDepth, &p.IsFlapping, &p.NotificationsOn,
			&p.HostIsDown, &severity); err != nil {
			return nil, 0, fmt.Errorf("listing problems: %w", err)
		}
		p.Kind = domain.Kind(kind)
		p.InDowntime = downtimeDepth > 0
		if p.Kind == domain.KindHost {
			p.StateText = domain.HostState(p.State).String()
		} else {
			p.StateText = domain.ServiceState(p.State).String()
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing problems: %w", err)
	}
	return out, total, nil
}
