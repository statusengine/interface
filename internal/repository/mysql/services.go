package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Services reads statusengine_servicestatus.
type Services struct {
	db *sql.DB
}

// NewServices returns a repository backed by db.
func NewServices(db *sql.DB) *Services { return &Services{db: db} }

// ServiceSortColumns is the whitelist of things a caller may sort
// services by.
//
// `severity` orders the way an operator triages - CRITICAL, WARNING,
// UNKNOWN, OK - rather than by the raw state number, which would put
// UNKNOWN above CRITICAL because it happens to be 3.
var ServiceSortColumns = map[string]string{
	"hostname":              "hostname",
	"service_description":   "service_description",
	"severity":              "CASE current_state WHEN 2 THEN 3 WHEN 1 THEN 2 WHEN 3 THEN 1 ELSE 0 END",
	"state":                 "current_state",
	"last_check":            "last_check",
	"next_check":            "next_check",
	"last_state_change":     "last_state_change",
	"current_check_attempt": "current_check_attempt",
	"latency":               "latency",
	"execution_time":        "execution_time",
	"output":                "output",
}

const serviceListColumns = `hostname, service_description,
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

const serviceDetailColumns = serviceListColumns + `,
	COALESCE(long_output, ''), COALESCE(check_command, ''), COALESCE(event_handler, ''),
	COALESCE(check_timeperiod, ''), COALESCE(node_name, ''),
	COALESCE(normal_check_interval, 0), COALESCE(retry_check_interval, 0),
	COALESCE(percent_state_change, 0),
	last_time_ok, last_time_warning, last_time_critical, last_time_unknown,
	last_notification, next_notification, COALESCE(current_notification_number, 0)`

// List returns a page of services and the total the filter matches.
func (r *Services) List(ctx context.Context, f StatusFilter, p Page) ([]domain.ServiceStatus, int64, error) {
	var c conditions
	f.apply(&c, []string{"hostname", "service_description"}, int(domain.ServiceOK))

	var total int64
	countQuery := "SELECT COUNT(*) FROM statusengine_servicestatus" + c.clause()
	if err := r.db.QueryRowContext(ctx, countQuery, c.params()...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting services: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := "SELECT " + serviceListColumns + " FROM statusengine_servicestatus" +
		c.clause() + orderBy(p.Sort, p.Desc, ServiceSortColumns, "hostname") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(c.params(), p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing services: %w", err)
	}
	defer rows.Close()

	var out []domain.ServiceStatus
	for rows.Next() {
		s, err := scanServiceList(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("listing services: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing services: %w", err)
	}
	return out, total, nil
}

// Get returns one service with the fields a detail page needs.
func (r *Services) Get(ctx context.Context, hostname, description string) (domain.ServiceStatus, error) {
	query := "SELECT " + serviceDetailColumns +
		" FROM statusengine_servicestatus WHERE hostname = ? AND service_description = ?"

	var s domain.ServiceStatus
	err := r.db.QueryRowContext(ctx, query, hostname, description).Scan(
		serviceListTargets(&s,
			&s.LongOutput, &s.CheckCommand, &s.EventHandler, &s.CheckTimeperiod, &s.NodeName,
			&s.CheckInterval, &s.RetryInterval, &s.PercentChange,
			&s.LastTimeOK, &s.LastTimeWarning, &s.LastTimeCrit, &s.LastTimeUnknown,
			&s.LastNotified, &s.NextNotify, &s.NotifyNumber,
		)...)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}
	if err != nil {
		return s, fmt.Errorf("reading service %q on %q: %w", description, hostname, err)
	}
	finishService(&s)
	return s, nil
}

func serviceListTargets(s *domain.ServiceStatus, extra ...any) []any {
	base := []any{
		&s.Hostname, &s.Description, &s.State, &s.IsHard, &s.Output, &s.Perfdata,
		&s.CurrentAttempt, &s.MaxAttempts, &s.LastCheck, &s.NextCheck,
		&s.LastStateChange, &s.LastHardChange, &s.StatusUpdated,
		&s.Acknowledged, &s.AckType, &s.DowntimeDepth,
		&s.IsFlapping, &s.NotificationsOn, &s.ActiveChecksOn, &s.PassiveChecksOn,
		&s.IsPassiveCheck, &s.EventHandlerOn, &s.FlapDetectionOn,
		&s.Latency, &s.ExecutionTime,
	}
	return append(base, extra...)
}

func scanServiceList(rows *sql.Rows) (domain.ServiceStatus, error) {
	var s domain.ServiceStatus
	if err := rows.Scan(serviceListTargets(&s)...); err != nil {
		return s, err
	}
	finishService(&s)
	return s, nil
}

func finishService(s *domain.ServiceStatus) {
	s.StateText = domain.ServiceState(s.State).String()
	if s.Pending() {
		s.StateText = "pending"
	}
	s.InDowntime = s.DowntimeDepth > 0
}
