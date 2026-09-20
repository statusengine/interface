package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// ErrNotFound is returned when a named host or service does not exist.
var ErrNotFound = errors.New("not found")

// Hosts reads statusengine_hoststatus.
type Hosts struct {
	db *sql.DB
}

// NewHosts returns a repository backed by db.
func NewHosts(db *sql.DB) *Hosts { return &Hosts{db: db} }

// HostSortColumns is the whitelist of things a caller may sort hosts by.
//
// `severity` is not a column. Sorting by the raw state number puts
// UNREACHABLE above DOWN, which is backwards: an unreachable host is
// usually a consequence of some other host being down, while a down host
// is the thing to look at. The CASE puts them in the order an operator
// triages in.
var HostSortColumns = map[string]string{
	"hostname":              "hostname",
	"severity":              "CASE current_state WHEN 1 THEN 2 WHEN 2 THEN 1 ELSE 0 END",
	"state":                 "current_state",
	"last_check":            "last_check",
	"next_check":            "next_check",
	"last_state_change":     "last_state_change",
	"current_check_attempt": "current_check_attempt",
	"latency":               "latency",
	"execution_time":        "execution_time",
	"output":                "output",
}

// Nearly every column in statusengine_hoststatus is declared NULL-able,
// even the ones the worker always writes. COALESCE here rather than a
// NullString per field: a single unexpected NULL would otherwise fail the
// whole page, and "no performance data" is a perfectly ordinary plugin
// result, not an error.
const hostListColumns = `hostname,
	COALESCE(current_state, 0), COALESCE(is_hardstate, 0),
	COALESCE(output, ''), COALESCE(perfdata, ''),
	COALESCE(current_check_attempt, 0), COALESCE(max_check_attempts, 0),
	last_check, next_check, last_state_change, last_hard_state_change, status_update_time,
	COALESCE(problem_has_been_acknowledged, 0), COALESCE(acknowledgement_type, 0),
	COALESCE(scheduled_downtime_depth, 0),
	COALESCE(is_flapping, 0), COALESCE(notifications_enabled, 0),
	COALESCE(active_checks_enabled, 0), COALESCE(passive_checks_enabled, 0),
	COALESCE(is_passive_check, 0), COALESCE(event_handler_enabled, 0),
	COALESCE(flap_detection_enabled, 0),
	COALESCE(latency, 0), COALESCE(execution_time, 0)`

// long_output is varchar(8192); a list of ten thousand rows does not need
// to carry it, and the detail view is the only place it is shown.
const hostDetailColumns = hostListColumns + `,
	COALESCE(long_output, ''), COALESCE(check_command, ''), COALESCE(event_handler, ''),
	COALESCE(check_timeperiod, ''), COALESCE(node_name, ''),
	COALESCE(normal_check_interval, 0), COALESCE(retry_check_interval, 0),
	COALESCE(percent_state_change, 0),
	last_time_up, last_time_down, last_time_unreachable,
	last_notification, next_notification, COALESCE(current_notification_number, 0)`

// List returns a page of hosts and the total the filter matches.
func (r *Hosts) List(ctx context.Context, f StatusFilter, p Page) ([]domain.HostStatus, int64, error) {
	var c conditions
	f.apply(&c, []string{"hostname"}, int(domain.HostUp))

	var total int64
	countQuery := "SELECT COUNT(*) FROM statusengine_hoststatus" + c.clause()
	if err := r.db.QueryRowContext(ctx, countQuery, c.params()...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting hosts: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := "SELECT " + hostListColumns + " FROM statusengine_hoststatus" +
		c.clause() + orderBy(p.Sort, p.Desc, HostSortColumns, "hostname") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(c.params(), p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing hosts: %w", err)
	}
	defer rows.Close()

	var out []domain.HostStatus
	for rows.Next() {
		h, err := scanHostList(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("listing hosts: %w", err)
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing hosts: %w", err)
	}
	return out, total, nil
}

// Get returns one host with the fields a detail page needs.
func (r *Hosts) Get(ctx context.Context, hostname string) (domain.HostStatus, error) {
	query := "SELECT " + hostDetailColumns + " FROM statusengine_hoststatus WHERE hostname = ?"

	var h domain.HostStatus
	err := r.db.QueryRowContext(ctx, query, hostname).Scan(
		hostListTargets(&h,
			&h.LongOutput, &h.CheckCommand, &h.EventHandler, &h.CheckTimeperiod, &h.NodeName,
			&h.CheckInterval, &h.RetryInterval, &h.PercentChange,
			&h.LastTimeUp, &h.LastTimeDown, &h.LastTimeUnreach,
			&h.LastNotified, &h.NextNotify, &h.NotifyNumber,
		)...)
	if errors.Is(err, sql.ErrNoRows) {
		return h, ErrNotFound
	}
	if err != nil {
		return h, fmt.Errorf("reading host %q: %w", hostname, err)
	}
	finishHost(&h)
	return h, nil
}

// Names returns every host name, for filter dropdowns. Cheap: one indexed
// column, no row bodies.
func (r *Hosts) Names(ctx context.Context, limit int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT hostname FROM statusengine_hoststatus ORDER BY hostname LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("listing host names: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("listing host names: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// hostListTargets returns the Scan destinations for hostListColumns, plus
// whatever extra destinations the caller appends. Keeping the order in
// one place means the column list and the scan cannot drift apart
// silently - they drift apart loudly, at the first query.
func hostListTargets(h *domain.HostStatus, extra ...any) []any {
	base := []any{
		&h.Hostname, &h.State, &h.IsHard, &h.Output, &h.Perfdata,
		&h.CurrentAttempt, &h.MaxAttempts, &h.LastCheck, &h.NextCheck,
		&h.LastStateChange, &h.LastHardChange, &h.StatusUpdated,
		&h.Acknowledged, &h.AckType, &h.DowntimeDepth,
		&h.IsFlapping, &h.NotificationsOn, &h.ActiveChecksOn, &h.PassiveChecksOn,
		&h.IsPassiveCheck, &h.EventHandlerOn, &h.FlapDetectionOn,
		&h.Latency, &h.ExecutionTime,
	}
	return append(base, extra...)
}

func scanHostList(rows *sql.Rows) (domain.HostStatus, error) {
	var h domain.HostStatus
	if err := rows.Scan(hostListTargets(&h)...); err != nil {
		return h, err
	}
	finishHost(&h)
	return h, nil
}

// finishHost fills the fields the database does not store: the derived
// ones the client would otherwise have to compute identically in two
// places.
func finishHost(h *domain.HostStatus) {
	h.StateText = domain.HostState(h.State).String()
	if h.Pending() {
		h.StateText = "pending"
	}
	h.InDowntime = h.DowntimeDepth > 0
}
