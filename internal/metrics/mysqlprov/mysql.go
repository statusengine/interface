// Package mysqlprov reads performance data from statusengine_perfdata.
package mysqlprov

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/statusengine/interface/internal/domain"
	"github.com/statusengine/interface/internal/metrics"
)

// Provider reads statusengine_perfdata.
type Provider struct {
	db *sql.DB
}

// New returns a Provider backed by db.
func New(db *sql.DB) *Provider { return &Provider{db: db} }

// Name identifies this provider in a response.
func (p *Provider) Name() string { return "mysql" }

// Labels lists a service's series.
//
// Reads only the leading three columns of the `metric` index
// (hostname, service_description, label, timestamp_unix), so the MIN and
// MAX come out of the index without touching a row.
func (p *Provider) Labels(ctx context.Context, hostname, description string) ([]domain.MetricMeta, error) {
	const query = `SELECT label, COALESCE(unit, ''),
		MIN(timestamp_unix), MAX(timestamp_unix)
		FROM statusengine_perfdata
		WHERE hostname = ? AND service_description = ?
		GROUP BY label, unit
		ORDER BY label`

	rows, err := p.db.QueryContext(ctx, query, hostname, description)
	if err != nil {
		return nil, fmt.Errorf("listing metrics: %w", err)
	}
	defer rows.Close()

	var out []domain.MetricMeta
	for rows.Next() {
		meta := domain.MetricMeta{Hostname: hostname, Description: description}
		var label sql.NullString
		if err := rows.Scan(&label, &meta.Unit, &meta.FirstSeen, &meta.LastSeen); err != nil {
			return nil, fmt.Errorf("listing metrics: %w", err)
		}
		meta.Label = label.String
		out = append(out, meta)
	}
	return out, rows.Err()
}

// Query returns downsampled series.
func (p *Provider) Query(ctx context.Context, q domain.MetricQuery) (domain.MetricResult, error) {
	if err := metrics.Validate(q); err != nil {
		return domain.MetricResult{}, err
	}

	bucket := metrics.BucketFor(q.From, q.To, q.MaxPoints)
	result := domain.MetricResult{
		BucketSeconds: bucket,
		From:          q.From,
		To:            q.To,
		Source:        p.Name(),
	}

	// Column order matches the `metric` index exactly:
	// (hostname, service_description, label, timestamp_unix).
	var sb strings.Builder
	// DIV, not FLOOR(a / b): `/` produces a DECIMAL in MySQL and FLOOR of
	// a DECIMAL stays a DECIMAL, which arrives as
	// "1789896600.0000000..." and fails to scan into an int64. DIV is
	// integer division and yields a BIGINT. MariaDB has it too.
	sb.WriteString(`SELECT label, COALESCE(unit, ''),
		(timestamp_unix DIV ?) * ? AS bucket,
		AVG(value), MIN(value), MAX(value)
		FROM statusengine_perfdata
		WHERE hostname = ? AND service_description = ?`)

	args := []any{bucket, bucket, q.Hostname, q.Description}

	if len(q.Labels) > 0 {
		sb.WriteString(" AND label IN (")
		sb.WriteString(strings.TrimSuffix(strings.Repeat("?,", len(q.Labels)), ","))
		sb.WriteString(")")
		for _, label := range q.Labels {
			args = append(args, label)
		}
	}

	sb.WriteString(` AND timestamp_unix BETWEEN ? AND ?
		GROUP BY label, unit, bucket
		ORDER BY label, bucket`)
	args = append(args, q.From, q.To)

	rows, err := p.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return result, fmt.Errorf("querying metrics: %w", err)
	}
	defer rows.Close()

	// Rows arrive grouped by label, so one pass builds the series without
	// a map and without sorting afterwards.
	var current *domain.Series
	for rows.Next() {
		var label sql.NullString
		var unit string
		var point domain.Point
		var avg, min, max sql.NullFloat64

		if err := rows.Scan(&label, &unit, &point.Time, &avg, &min, &max); err != nil {
			return result, fmt.Errorf("querying metrics: %w", err)
		}
		point.Avg, point.Min, point.Max = avg.Float64, min.Float64, max.Float64

		if current == nil || current.Label != label.String {
			result.Series = append(result.Series, domain.Series{Label: label.String, Unit: unit})
			current = &result.Series[len(result.Series)-1]
		}
		current.Points = append(current.Points, point)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("querying metrics: %w", err)
	}
	return result, nil
}
