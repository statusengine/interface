package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// History reads the check, state-change and notification tables.
type History struct {
	db *sql.DB
}

// NewHistory returns a repository backed by db.
func NewHistory(db *sql.DB) *History { return &History{db: db} }

// HistoryFilter narrows a history list.
//
// Host and Description are not just filters here, they decide how the
// query performs. These tables are clustered on an object-first primary
// key, so naming an object turns the read into one contiguous range;
// leaving it out means walking small per-object runs scattered across
// the whole table. The handler caps the window much harder in the second
// case, and the UI says why.
type HistoryFilter struct {
	Host        string
	Description string
	Kind        domain.Kind
	Search      string
	States      []int

	// From and To are always set by the handler.
	From int64
	To   int64

	// HardOnly drops soft states, which is what someone asking "when did
	// this actually break" means.
	HardOnly bool

	// TransitionsOnly drops statehistory rows that repeat the previous
	// state at a new check attempt.
	TransitionsOnly bool
}

// Scoped reports whether the filter names an object, which is what puts
// the query on the clustered primary key's fast path.
func (f HistoryFilter) Scoped() bool { return f.Host != "" }

// CheckSortColumns is the whitelist for the check history.
var CheckSortColumns = map[string]string{
	"start_time":          "start_time",
	"hostname":            "hostname",
	"service_description": "service_description",
	"state":               "state",
	"execution_time":      "execution_time",
	"latency":             "latency",
}

// StateChangeSortColumns is the whitelist for the state-change history.
var StateChangeSortColumns = map[string]string{
	"state_time":          "state_time",
	"hostname":            "hostname",
	"service_description": "service_description",
	"state":               "state",
}

// NotificationSortColumns is the whitelist for the notification history.
var NotificationSortColumns = map[string]string{
	"start_time":          "start_time",
	"hostname":            "hostname",
	"service_description": "service_description",
	"contact_name":        "contact_name",
	"state":               "state",
}

// --- checks ----------------------------------------------------------------

const hostCheckColumns = `'host' AS kind, hostname, '' AS service_description,
	start_time, end_time, COALESCE(state, 0) AS state, COALESCE(is_hardstate, 0) AS is_hard,
	COALESCE(output, '') AS output, COALESCE(long_output, '') AS long_output,
	COALESCE(perfdata, '') AS perfdata, COALESCE(command, '') AS command,
	COALESCE(current_check_attempt, 0) AS current_check_attempt,
	COALESCE(max_check_attempts, 0) AS max_check_attempts,
	COALESCE(latency, 0) AS latency, COALESCE(execution_time, 0) AS execution_time,
	COALESCE(timeout, 0) AS timeout, COALESCE(early_timeout, 0) AS early_timeout`

const serviceCheckColumns = `'service' AS kind, hostname, service_description,
	start_time, end_time, COALESCE(state, 0) AS state, COALESCE(is_hardstate, 0) AS is_hard,
	COALESCE(output, '') AS output, COALESCE(long_output, '') AS long_output,
	COALESCE(perfdata, '') AS perfdata, COALESCE(command, '') AS command,
	COALESCE(current_check_attempt, 0) AS current_check_attempt,
	COALESCE(max_check_attempts, 0) AS max_check_attempts,
	COALESCE(latency, 0) AS latency, COALESCE(execution_time, 0) AS execution_time,
	COALESCE(timeout, 0) AS timeout, COALESCE(early_timeout, 0) AS early_timeout`

// Checks returns a page of executed checks.
func (r *History) Checks(ctx context.Context, f HistoryFilter, p Page) ([]domain.CheckResult, int64, error) {
	body, args := f.checkUnion()

	total, err := r.count(ctx, body, args, "checks")
	if err != nil || total == 0 {
		return nil, total, err
	}

	query := "SELECT * FROM (" + body + ") AS c" +
		orderBy(p.Sort, p.Desc, CheckSortColumns, "start_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing checks: %w", err)
	}
	defer rows.Close()

	var out []domain.CheckResult
	for rows.Next() {
		var c domain.CheckResult
		var kind string
		if err := rows.Scan(&kind, &c.Hostname, &c.Description, &c.StartTime, &c.EndTime,
			&c.State, &c.IsHard, &c.Output, &c.LongOutput, &c.Perfdata, &c.Command,
			&c.CurrentAttempt, &c.MaxAttempts, &c.Latency, &c.ExecutionTime,
			&c.Timeout, &c.EarlyTimeout); err != nil {
			return nil, 0, fmt.Errorf("listing checks: %w", err)
		}
		c.Kind = domain.Kind(kind)
		c.StateText = stateText(c.Kind, c.State)
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing checks: %w", err)
	}
	return out, total, nil
}

func (f HistoryFilter) checkUnion() (string, []any) {
	var parts []string
	var args []any

	if f.Kind != domain.KindService && f.Description == "" {
		var c conditions
		f.applyTime(&c, "start_time")
		f.applyObject(&c, false)
		f.applyCommon(&c, []string{"hostname", "output"})
		f.applyHardOnly(&c)
		parts = append(parts, "SELECT "+hostCheckColumns+" FROM statusengine_hostchecks"+c.clause())
		args = append(args, c.params()...)
	}
	if f.Kind != domain.KindHost {
		var c conditions
		f.applyTime(&c, "start_time")
		f.applyObject(&c, true)
		f.applyCommon(&c, []string{"hostname", "service_description", "output"})
		f.applyHardOnly(&c)
		parts = append(parts, "SELECT "+serviceCheckColumns+" FROM statusengine_servicechecks"+c.clause())
		args = append(args, c.params()...)
	}

	switch len(parts) {
	case 0:
		return emptyCheckSelect, nil
	case 1:
		return parts[0], args
	default:
		return parts[0] + " UNION ALL " + parts[1], args
	}
}

const emptyCheckSelect = `SELECT 'host' AS kind, '' AS hostname, '' AS service_description,
	0 AS start_time, 0 AS end_time, 0 AS state, 0 AS is_hard, '' AS output,
	'' AS long_output, '' AS perfdata, '' AS command, 0 AS current_check_attempt,
	0 AS max_check_attempts, 0 AS latency, 0 AS execution_time, 0 AS timeout,
	0 AS early_timeout FROM DUAL WHERE FALSE`

// --- state changes ---------------------------------------------------------

const hostStateColumns = `'host' AS kind, hostname, '' AS service_description,
	state_time, COALESCE(state, 0) AS state, COALESCE(last_state, 0) AS last_state,
	COALESCE(last_hard_state, 0) AS last_hard_state, COALESCE(is_hardstate, 0) AS is_hard,
	COALESCE(state_change, 0) AS is_transition,
	COALESCE(current_check_attempt, 0) AS current_check_attempt,
	COALESCE(max_check_attempts, 0) AS max_check_attempts,
	COALESCE(output, '') AS output, COALESCE(long_output, '') AS long_output`

const serviceStateColumns = `'service' AS kind, hostname, service_description,
	state_time, COALESCE(state, 0) AS state, COALESCE(last_state, 0) AS last_state,
	COALESCE(last_hard_state, 0) AS last_hard_state, COALESCE(is_hardstate, 0) AS is_hard,
	COALESCE(state_change, 0) AS is_transition,
	COALESCE(current_check_attempt, 0) AS current_check_attempt,
	COALESCE(max_check_attempts, 0) AS max_check_attempts,
	COALESCE(output, '') AS output, COALESCE(long_output, '') AS long_output`

// StateChanges returns a page of state-history entries.
func (r *History) StateChanges(ctx context.Context, f HistoryFilter, p Page) ([]domain.StateChange, int64, error) {
	body, args := f.stateUnion()

	total, err := r.count(ctx, body, args, "state changes")
	if err != nil || total == 0 {
		return nil, total, err
	}

	query := "SELECT * FROM (" + body + ") AS s" +
		orderBy(p.Sort, p.Desc, StateChangeSortColumns, "state_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing state changes: %w", err)
	}
	defer rows.Close()

	var out []domain.StateChange
	for rows.Next() {
		var s domain.StateChange
		var kind string
		if err := rows.Scan(&kind, &s.Hostname, &s.Description, &s.StateTime,
			&s.State, &s.LastState, &s.LastHardState, &s.IsHard, &s.IsTransition,
			&s.CurrentAttempt, &s.MaxAttempts, &s.Output, &s.LongOutput); err != nil {
			return nil, 0, fmt.Errorf("listing state changes: %w", err)
		}
		s.Kind = domain.Kind(kind)
		s.StateText = stateText(s.Kind, s.State)
		s.LastStateText = stateText(s.Kind, s.LastState)
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing state changes: %w", err)
	}
	return out, total, nil
}

func (f HistoryFilter) stateUnion() (string, []any) {
	var parts []string
	var args []any

	if f.Kind != domain.KindService && f.Description == "" {
		var c conditions
		f.applyTime(&c, "state_time")
		f.applyObject(&c, false)
		f.applyCommon(&c, []string{"hostname", "output"})
		f.applyHardOnly(&c)
		if f.TransitionsOnly {
			c.add("state_change = 1")
		}
		parts = append(parts, "SELECT "+hostStateColumns+" FROM statusengine_host_statehistory"+c.clause())
		args = append(args, c.params()...)
	}
	if f.Kind != domain.KindHost {
		var c conditions
		f.applyTime(&c, "state_time")
		f.applyObject(&c, true)
		f.applyCommon(&c, []string{"hostname", "service_description", "output"})
		f.applyHardOnly(&c)
		if f.TransitionsOnly {
			c.add("state_change = 1")
		}
		parts = append(parts, "SELECT "+serviceStateColumns+" FROM statusengine_service_statehistory"+c.clause())
		args = append(args, c.params()...)
	}

	switch len(parts) {
	case 0:
		return emptyStateSelect, nil
	case 1:
		return parts[0], args
	default:
		return parts[0] + " UNION ALL " + parts[1], args
	}
}

const emptyStateSelect = `SELECT 'host' AS kind, '' AS hostname, '' AS service_description,
	0 AS state_time, 0 AS state, 0 AS last_state, 0 AS last_hard_state, 0 AS is_hard,
	0 AS is_transition, 0 AS current_check_attempt, 0 AS max_check_attempts,
	'' AS output, '' AS long_output FROM DUAL WHERE FALSE`

// --- notifications ---------------------------------------------------------

const hostNotificationColumns = `'host' AS kind, hostname, '' AS service_description,
	start_time, end_time, COALESCE(contact_name, '') AS contact_name,
	COALESCE(command_name, '') AS command_name, COALESCE(command_args, '') AS command_args,
	COALESCE(state, 0) AS state, COALESCE(reason_type, 0) AS reason_type,
	COALESCE(output, '') AS output, COALESCE(ack_author, '') AS ack_author,
	COALESCE(ack_data, '') AS ack_data`

const serviceNotificationColumns = `'service' AS kind, hostname, service_description,
	start_time, end_time, COALESCE(contact_name, '') AS contact_name,
	COALESCE(command_name, '') AS command_name, COALESCE(command_args, '') AS command_args,
	COALESCE(state, 0) AS state, COALESCE(reason_type, 0) AS reason_type,
	COALESCE(output, '') AS output, COALESCE(ack_author, '') AS ack_author,
	COALESCE(ack_data, '') AS ack_data`

// Notifications returns a page of per-contact notification records.
func (r *History) Notifications(ctx context.Context, f HistoryFilter, p Page) ([]domain.Notification, int64, error) {
	body, args := f.notificationUnion()

	total, err := r.count(ctx, body, args, "notifications")
	if err != nil || total == 0 {
		return nil, total, err
	}

	query := "SELECT * FROM (" + body + ") AS n" +
		orderBy(p.Sort, p.Desc, NotificationSortColumns, "start_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing notifications: %w", err)
	}
	defer rows.Close()

	var out []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var kind string
		if err := rows.Scan(&kind, &n.Hostname, &n.Description, &n.StartTime, &n.EndTime,
			&n.Contact, &n.CommandName, &n.CommandArgs, &n.State, &n.ReasonType,
			&n.Output, &n.AckAuthor, &n.AckData); err != nil {
			return nil, 0, fmt.Errorf("listing notifications: %w", err)
		}
		n.Kind = domain.Kind(kind)
		n.StateText = stateText(n.Kind, n.State)
		n.Reason = domain.NotificationReason(n.ReasonType)
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing notifications: %w", err)
	}
	return out, total, nil
}

func (f HistoryFilter) notificationUnion() (string, []any) {
	var parts []string
	var args []any

	if f.Kind != domain.KindService && f.Description == "" {
		var c conditions
		f.applyTime(&c, "start_time")
		f.applyObject(&c, false)
		f.applyCommon(&c, []string{"hostname", "contact_name", "output"})
		parts = append(parts, "SELECT "+hostNotificationColumns+" FROM statusengine_host_notifications"+c.clause())
		args = append(args, c.params()...)
	}
	if f.Kind != domain.KindHost {
		var c conditions
		f.applyTime(&c, "start_time")
		f.applyObject(&c, true)
		f.applyCommon(&c, []string{"hostname", "service_description", "contact_name", "output"})
		parts = append(parts, "SELECT "+serviceNotificationColumns+" FROM statusengine_service_notifications"+c.clause())
		args = append(args, c.params()...)
	}

	switch len(parts) {
	case 0:
		return emptyNotificationSelect, nil
	case 1:
		return parts[0], args
	default:
		return parts[0] + " UNION ALL " + parts[1], args
	}
}

const emptyNotificationSelect = `SELECT 'host' AS kind, '' AS hostname, '' AS service_description,
	0 AS start_time, 0 AS end_time, '' AS contact_name, '' AS command_name,
	'' AS command_args, 0 AS state, 0 AS reason_type, '' AS output, '' AS ack_author,
	'' AS ack_data FROM DUAL WHERE FALSE`

// --- shared ----------------------------------------------------------------

func (r *History) count(ctx context.Context, body string, args []any, what string) (int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+body+") AS x", args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("counting %s: %w", what, err)
	}
	return total, nil
}

// applyTime writes the window. It comes first so it reads in the same
// order it is enforced: there is no unbounded history query here.
func (f HistoryFilter) applyTime(c *conditions, column string) {
	c.add(column+" >= ?", f.From)
	c.add(column+" <= ?", f.To)
}

// applyHardOnly drops soft states. Separate from applyCommon because
// only the check and statehistory tables have is_hardstate - the
// notification tables do not, and a comment asking callers to remember
// that is a rule nothing enforces.
func (f HistoryFilter) applyHardOnly(c *conditions) {
	if f.HardOnly {
		c.add("is_hardstate = 1")
	}
}

// applyObject names the host and, for a service table, the service. This
// is what puts the query on the leading columns of the clustered primary
// key.
func (f HistoryFilter) applyObject(c *conditions, isService bool) {
	if f.Host != "" {
		c.add("hostname = ?", f.Host)
	}
	if isService && f.Description != "" {
		c.add("service_description = ?", f.Description)
	}
}

func (f HistoryFilter) applyCommon(c *conditions, searchColumns []string) {
	c.addIn("state", f.States)
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
}

func stateText(kind domain.Kind, state int) string {
	if kind == domain.KindHost {
		return domain.HostState(state).String()
	}
	return domain.ServiceState(state).String()
}
