package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Entry is one submission, recorded whatever the outcome.
type Entry struct {
	UserID     *uint32
	Username   string
	Action     Action
	Target     string
	Payload    any
	HTTPStatus int
	Response   string
	RemoteIP   string
}

// Audit records every external command this interface submits.
//
// Refusals are recorded too. "Who tried to acknowledge that and why did
// it not take" is a question that comes up during a post-mortem, and an
// audit trail that only holds successes cannot answer it.
type Audit struct {
	db *sql.DB
}

// NewAudit returns an Audit backed by db.
func NewAudit(db *sql.DB) *Audit { return &Audit{db: db} }

// RecordBatch writes one entry per target in a single statement.
//
// One row per object, not one per submission: a bulk downtime over two
// hundred services is two hundred things that happened, and "which of
// them did not take" is the question an audit trail exists to answer.
func (a *Audit) RecordBatch(ctx context.Context, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}

	const columns = "(user_id, username, action, target, payload, http_status, response, remote_ip, created_at)"
	values := make([]string, 0, len(entries))
	args := make([]any, 0, len(entries)*9)
	now := time.Now().Unix()

	for _, e := range entries {
		payload, err := json.Marshal(e.Payload)
		if err != nil {
			payload = []byte(`{"error":"payload could not be encoded"}`)
		}
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args, e.UserID, e.Username, string(e.Action), truncate(e.Target, 512),
			payload, e.HTTPStatus, truncate(e.Response, 1024), truncate(e.RemoteIP, 45), now)
	}

	query := "INSERT INTO sei_command_audit " + columns + " VALUES " + strings.Join(values, ", ")
	if _, err := a.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("commands: recording %d audit entries: %w", len(entries), err)
	}
	return nil
}

// Record writes one entry.
func (a *Audit) Record(ctx context.Context, e Entry) error {
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		payload = []byte(`{"error":"payload could not be encoded"}`)
	}

	const query = `INSERT INTO sei_command_audit
		(user_id, username, action, target, payload, http_status, response, remote_ip, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = a.db.ExecContext(ctx, query,
		e.UserID, e.Username, string(e.Action), truncate(e.Target, 512), payload,
		e.HTTPStatus, truncate(e.Response, 1024), truncate(e.RemoteIP, 45),
		time.Now().Unix())
	if err != nil {
		return fmt.Errorf("commands: recording audit entry: %w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Record is what the audit list returns.
type Record struct {
	ID         uint64 `json:"id"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	Target     string `json:"target"`
	Payload    any    `json:"payload,omitempty"`
	HTTPStatus int    `json:"http_status"`
	Response   string `json:"response"`
	RemoteIP   string `json:"remote_ip,omitempty"`
	CreatedAt  int64  `json:"created_at"`
}

// AuditFilter narrows the audit list.
type AuditFilter struct {
	Username string
	Action   string
	Search   string
	From     int64
	To       int64
	// FailedOnly keeps the submissions that did not reach the broker,
	// which is the interesting half during a post-mortem.
	FailedOnly bool
}

// List returns a page of audit records, newest first.
func (a *Audit) List(ctx context.Context, f AuditFilter, limit, offset int) ([]Record, int64, error) {
	where := " WHERE 1=1"
	var args []any

	if f.Username != "" {
		where += " AND username = ?"
		args = append(args, f.Username)
	}
	if f.Action != "" {
		where += " AND action = ?"
		args = append(args, f.Action)
	}
	if f.From > 0 {
		where += " AND created_at >= ?"
		args = append(args, f.From)
	}
	if f.To > 0 {
		where += " AND created_at <= ?"
		args = append(args, f.To)
	}
	if f.FailedOnly {
		where += " AND http_status <> 202"
	}
	if f.Search != "" {
		where += " AND (target LIKE ? OR response LIKE ? OR username LIKE ?)"
		pattern := "%" + escapeLike(f.Search) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	var total int64
	if err := a.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sei_command_audit"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("commands: counting audit entries: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := `SELECT id, username, action, target, payload, http_status, response,
		remote_ip, created_at FROM sei_command_audit` + where +
		" ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?"

	rows, err := a.db.QueryContext(ctx, query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("commands: listing audit entries: %w", err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		var payload []byte
		if err := rows.Scan(&rec.ID, &rec.Username, &rec.Action, &rec.Target, &payload,
			&rec.HTTPStatus, &rec.Response, &rec.RemoteIP, &rec.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("commands: listing audit entries: %w", err)
		}
		if len(payload) > 0 {
			var decoded any
			if err := json.Unmarshal(payload, &decoded); err == nil {
				rec.Payload = decoded
			}
		}
		out = append(out, rec)
	}
	return out, total, rows.Err()
}

// escapeLike keeps a search string from acting as a LIKE pattern.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
