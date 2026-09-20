package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Downtimes reads the scheduled-downtime and downtime-history tables.
type Downtimes struct {
	db *sql.DB
}

// NewDowntimes returns a repository backed by db.
func NewDowntimes(db *sql.DB) *Downtimes { return &Downtimes{db: db} }

// DowntimeFilter narrows a downtime list.
type DowntimeFilter struct {
	Search string
	Host   string

	// Kind restricts to "host" or "service"; empty means both.
	Kind domain.Kind

	// Running keeps only windows that are open at Now, or only those that
	// are not. Nil means both.
	Running *bool

	// Now is the instant "running" is measured against. Zero means the
	// caller did not set it and the repository uses the wall clock.
	Now int64
}

// DowntimeSortColumns is the whitelist for downtime lists.
var DowntimeSortColumns = map[string]string{
	"scheduled_start_time": "scheduled_start_time",
	"scheduled_end_time":   "scheduled_end_time",
	"entry_time":           "entry_time",
	"hostname":             "hostname",
	"service_description":  "service_description",
	"author":               "author_name",
	"kind":                 "kind",
}

const downtimeCommonColumns = `internal_downtime_id, node_name, entry_time,
	COALESCE(author_name, '') AS author_name, COALESCE(comment_data, '') AS comment_data,
	COALESCE(triggered_by_id, 0) AS triggered_by_id, COALESCE(is_fixed, 0) AS is_fixed,
	COALESCE(duration, 0) AS duration,
	scheduled_start_time, scheduled_end_time,
	COALESCE(was_started, 0) AS was_started, actual_start_time`

// Current lists the downtimes the core is holding right now: scheduled,
// running, or scheduled and not yet started.
func (r *Downtimes) Current(ctx context.Context, f DowntimeFilter, p Page) ([]domain.Downtime, int64, error) {
	body, args := f.union(false)
	return r.query(ctx, body, args, p, false)
}

// History lists downtimes that have finished, including the ones that
// were cancelled rather than allowed to expire.
func (r *Downtimes) History(ctx context.Context, f DowntimeFilter, p Page) ([]domain.Downtime, int64, error) {
	body, args := f.union(true)
	return r.query(ctx, body, args, p, true)
}

func (f DowntimeFilter) union(history bool) (string, []any) {
	hostTable := "statusengine_host_scheduleddowntimes"
	svcTable := "statusengine_service_scheduleddowntimes"
	extra := ", 0 AS actual_end_time, 0 AS was_cancelled"
	if history {
		hostTable = "statusengine_host_downtimehistory"
		svcTable = "statusengine_service_downtimehistory"
		extra = ", actual_end_time, COALESCE(was_cancelled, 0) AS was_cancelled"
	}

	var parts []string
	var args []any

	if f.Kind != domain.KindService {
		var c conditions
		f.applyCommon(&c, []string{"hostname"})
		parts = append(parts, "SELECT 'host' AS kind, hostname, '' AS service_description, "+
			downtimeCommonColumns+extra+" FROM "+hostTable+c.clause())
		args = append(args, c.params()...)
	}
	if f.Kind != domain.KindHost {
		var c conditions
		f.applyCommon(&c, []string{"hostname", "service_description"})
		parts = append(parts, "SELECT 'service' AS kind, hostname, service_description, "+
			downtimeCommonColumns+extra+" FROM "+svcTable+c.clause())
		args = append(args, c.params()...)
	}

	switch len(parts) {
	case 0:
		return "SELECT 'host' AS kind, '' AS hostname, '' AS service_description," +
			" 0 AS internal_downtime_id, '' AS node_name, 0 AS entry_time," +
			" '' AS author_name, '' AS comment_data, 0 AS triggered_by_id, 0 AS is_fixed," +
			" 0 AS duration, 0 AS scheduled_start_time, 0 AS scheduled_end_time," +
			" 0 AS was_started, 0 AS actual_start_time, 0 AS actual_end_time," +
			" 0 AS was_cancelled FROM DUAL WHERE FALSE", nil
	case 1:
		return parts[0], args
	default:
		return parts[0] + " UNION ALL " + parts[1], args
	}
}

func (f DowntimeFilter) applyCommon(c *conditions, searchColumns []string) {
	if f.Host != "" {
		c.add("hostname = ?", f.Host)
	}
	if f.Search != "" {
		pattern := likeEscape(f.Search)
		var exprs []string
		var args []any
		for _, col := range searchColumns {
			exprs = append(exprs, col+" LIKE ? ESCAPE '\\\\'")
			args = append(args, pattern)
		}
		// The comment is worth searching: it is where the change ticket
		// number ends up.
		exprs = append(exprs, "comment_data LIKE ? ESCAPE '\\\\'")
		args = append(args, pattern)
		c.add("("+joinOr(exprs)+")", args...)
	}
	if f.Running != nil {
		now := f.Now
		if *f.Running {
			c.add("was_started = 1 AND scheduled_start_time <= ? AND scheduled_end_time > ?", now, now)
		} else {
			c.add("NOT (was_started = 1 AND scheduled_start_time <= ? AND scheduled_end_time > ?)", now, now)
		}
	}
}

func (r *Downtimes) query(ctx context.Context, body string, args []any, p Page, history bool) ([]domain.Downtime, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+body+") AS d", args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting downtimes: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := "SELECT * FROM (" + body + ") AS d" +
		orderBy(p.Sort, p.Desc, DowntimeSortColumns, "scheduled_start_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing downtimes: %w", err)
	}
	defer rows.Close()

	var out []domain.Downtime
	for rows.Next() {
		var d domain.Downtime
		var kind string
		if err := rows.Scan(&kind, &d.Hostname, &d.Description, &d.InternalID, &d.NodeName,
			&d.EntryTime, &d.Author, &d.Comment, &d.TriggeredBy, &d.IsFixed, &d.Duration,
			&d.StartTime, &d.EndTime, &d.WasStarted, &d.ActualStart,
			&d.ActualEnd, &d.WasCancelled); err != nil {
			return nil, 0, fmt.Errorf("listing downtimes: %w", err)
		}
		d.Kind = domain.Kind(kind)
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing downtimes: %w", err)
	}
	return out, total, nil
}
