package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// LogEntries reads statusengine_logentries.
type LogEntries struct {
	db *sql.DB
}

// NewLogEntries returns a repository backed by db.
func NewLogEntries(db *sql.DB) *LogEntries { return &LogEntries{db: db} }

// LogFilter narrows the log list.
//
// From and To are required rather than optional: logentries_se leads with
// entry_time, so a window turns this into a range scan instead of walking
// a table whose retention is measured in days.
type LogFilter struct {
	Search string
	From   int64
	To     int64
	Types  []int
	Node   string
}

// LogSortColumns is the whitelist for the log list.
var LogSortColumns = map[string]string{
	"entry_time":    "entry_time",
	"logentry_type": "logentry_type",
	"id":            "id",
}

// List returns a page of log entries.
func (r *LogEntries) List(ctx context.Context, f LogFilter, p Page) ([]domain.LogEntry, int64, error) {
	var c conditions
	// Always bound entry_time, even when the caller passed nothing: it is
	// the leading column of logentries_se, and an unbounded read is never
	// what anyone meant.
	c.add("entry_time >= ?", f.From)
	c.add("entry_time <= ?", f.To)
	c.addIn("logentry_type", f.Types)
	if f.Node != "" {
		c.add("node_name = ?", f.Node)
	}
	if f.Search != "" {
		c.add("logentry_data LIKE ? ESCAPE '\\\\'", likeEscape(f.Search))
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM statusengine_logentries" + c.clause()
	if err := r.db.QueryRowContext(ctx, countQuery, c.params()...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting log entries: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := `SELECT id, entry_time, COALESCE(logentry_type, 0),
		COALESCE(logentry_data, ''), COALESCE(node_name, '')
		FROM statusengine_logentries` + c.clause() +
		orderBy(p.Sort, p.Desc, LogSortColumns, "entry_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(c.params(), p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing log entries: %w", err)
	}
	defer rows.Close()

	var out []domain.LogEntry
	for rows.Next() {
		var e domain.LogEntry
		if err := rows.Scan(&e.ID, &e.EntryTime, &e.Type, &e.Data, &e.NodeName); err != nil {
			return nil, 0, fmt.Errorf("listing log entries: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing log entries: %w", err)
	}
	return out, total, nil
}
