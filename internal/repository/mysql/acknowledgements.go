package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Acknowledgements reads the two acknowledgement tables.
type Acknowledgements struct {
	db *sql.DB
}

// NewAcknowledgements returns a repository backed by db.
func NewAcknowledgements(db *sql.DB) *Acknowledgements { return &Acknowledgements{db: db} }

// AckFilter narrows an acknowledgement list.
type AckFilter struct {
	Search string
	Host   string
	Author string
	Kind   domain.Kind

	// From and To bound entry_time. Both zero means no bound.
	From int64
	To   int64
}

// AckSortColumns is the whitelist for acknowledgement lists.
var AckSortColumns = map[string]string{
	"entry_time":          "entry_time",
	"hostname":            "hostname",
	"service_description": "service_description",
	"author":              "author_name",
	"state":               "state",
	"kind":                "kind",
}

const ackCommonColumns = `entry_time, COALESCE(state, 0) AS state,
	COALESCE(author_name, '') AS author_name, COALESCE(comment_data, '') AS comment_data,
	COALESCE(acknowledgement_type, 0) AS acknowledgement_type,
	COALESCE(is_sticky, 0) AS is_sticky,
	COALESCE(persistent_comment, 0) AS persistent_comment,
	COALESCE(notify_contacts, 0) AS notify_contacts`

// List returns a page of acknowledgements across hosts and services.
func (r *Acknowledgements) List(ctx context.Context, f AckFilter, p Page) ([]domain.Acknowledgement, int64, error) {
	body, args := f.union()

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ("+body+") AS a", args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting acknowledgements: %w", err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	query := "SELECT * FROM (" + body + ") AS a" +
		orderBy(p.Sort, p.Desc, AckSortColumns, "entry_time") + p.limitClause()

	rows, err := r.db.QueryContext(ctx, query, append(args, p.limitArgs()...)...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing acknowledgements: %w", err)
	}
	defer rows.Close()

	var out []domain.Acknowledgement
	for rows.Next() {
		var a domain.Acknowledgement
		var kind string
		if err := rows.Scan(&kind, &a.Hostname, &a.Description, &a.EntryTime, &a.State,
			&a.Author, &a.Comment, &a.AckType, &a.IsSticky, &a.Persistent,
			&a.NotifyContacts); err != nil {
			return nil, 0, fmt.Errorf("listing acknowledgements: %w", err)
		}
		a.Kind = domain.Kind(kind)
		if a.Kind == domain.KindHost {
			a.StateText = domain.HostState(a.State).String()
		} else {
			a.StateText = domain.ServiceState(a.State).String()
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("listing acknowledgements: %w", err)
	}
	return out, total, nil
}

func (f AckFilter) union() (string, []any) {
	var parts []string
	var args []any

	if f.Kind != domain.KindService {
		var c conditions
		f.applyCommon(&c, []string{"hostname"})
		parts = append(parts, "SELECT 'host' AS kind, hostname, '' AS service_description, "+
			ackCommonColumns+" FROM statusengine_host_acknowledgements"+c.clause())
		args = append(args, c.params()...)
	}
	if f.Kind != domain.KindHost {
		var c conditions
		f.applyCommon(&c, []string{"hostname", "service_description"})
		parts = append(parts, "SELECT 'service' AS kind, COALESCE(hostname, '') AS hostname, "+
			"service_description, "+ackCommonColumns+
			" FROM statusengine_service_acknowledgements"+c.clause())
		args = append(args, c.params()...)
	}

	switch len(parts) {
	case 0:
		return "SELECT 'host' AS kind, '' AS hostname, '' AS service_description," +
			" 0 AS entry_time, 0 AS state, '' AS author_name, '' AS comment_data," +
			" 0 AS acknowledgement_type, 0 AS is_sticky, 0 AS persistent_comment," +
			" 0 AS notify_contacts FROM DUAL WHERE FALSE", nil
	case 1:
		return parts[0], args
	default:
		return parts[0] + " UNION ALL " + parts[1], args
	}
}

func (f AckFilter) applyCommon(c *conditions, searchColumns []string) {
	if f.Host != "" {
		c.add("hostname = ?", f.Host)
	}
	if f.Author != "" {
		c.add("author_name = ?", f.Author)
	}
	if f.From > 0 {
		c.add("entry_time >= ?", f.From)
	}
	if f.To > 0 {
		c.add("entry_time <= ?", f.To)
	}
	if f.Search != "" {
		pattern := likeEscape(f.Search)
		var exprs []string
		var args []any
		for _, col := range searchColumns {
			exprs = append(exprs, col+" LIKE ? ESCAPE '\\\\'")
			args = append(args, pattern)
		}
		// The comment is the point of an acknowledgement, so it is
		// searchable alongside the object it is attached to.
		exprs = append(exprs, "comment_data LIKE ? ESCAPE '\\\\'", "author_name LIKE ? ESCAPE '\\\\'")
		args = append(args, pattern, pattern)
		c.add("("+joinOr(exprs)+")", args...)
	}
}
