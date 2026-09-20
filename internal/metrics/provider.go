// Package metrics reads performance data from whichever backend holds
// it. MySQL today; the worker can already route perfdata to Graphite
// instead, so the second implementation is a real one waiting rather
// than a speculative abstraction.
package metrics

import (
	"context"
	"fmt"

	"github.com/statusengine/interface/internal/domain"
)

// Provider is a source of performance data.
type Provider interface {
	// Name identifies the provider in a response, so the UI can say where
	// a chart's data came from.
	Name() string

	// Labels lists the series a service has, with the window each one
	// actually covers.
	Labels(ctx context.Context, hostname, description string) ([]domain.MetricMeta, error)

	// Query returns downsampled series for a window.
	Query(ctx context.Context, q domain.MetricQuery) (domain.MetricResult, error)
}

// Buckets are the resolutions a provider may settle on. Round numbers,
// because "3m 17s average" is a label nobody can reason about, and they
// line up with the check intervals people actually configure.
var Buckets = []int64{
	10, 30, 60, 300, 900, 1800, 3600, 6 * 3600, 12 * 3600, 86400, 7 * 86400,
}

// BucketFor picks the smallest resolution that keeps a window within the
// caller's point budget.
//
// Downsampling is not optional: a week of ten-second checks is sixty
// thousand samples, which no browser draws usefully and no eye reads.
// The alternative - returning everything and letting the client thin it -
// moves the same cost onto the slowest machine in the chain.
func BucketFor(from, to int64, maxPoints int) int64 {
	if maxPoints < 1 {
		maxPoints = 500
	}
	span := to - from
	if span <= 0 {
		return Buckets[0]
	}
	for _, bucket := range Buckets {
		if span/bucket <= int64(maxPoints) {
			return bucket
		}
	}
	// Longer than a week per point: give the budget exactly, rounded to
	// whole seconds.
	if n := span / int64(maxPoints); n > 0 {
		return n
	}
	return Buckets[len(Buckets)-1]
}

// Validate checks a query before it reaches a backend.
func Validate(q domain.MetricQuery) error {
	if q.Hostname == "" {
		return fmt.Errorf("metrics: hostname is required")
	}
	if q.Description == "" {
		return fmt.Errorf("metrics: service_description is required")
	}
	if q.To < q.From {
		return fmt.Errorf("metrics: to is before from")
	}
	return nil
}
