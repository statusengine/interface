// Package graphite is the second metrics provider: not wired up yet, but
// the shape it has to fit is fixed by the Provider interface and by the
// worker, which can already be told to route performance data here with
// `perfdata_route: graphite` instead of to MySQL.
//
// What is left to do is the HTTP client. The rest of the interface -
// the endpoint, the chart, the downsampling contract - does not change
// when this is finished, which is the point of the abstraction.
package graphite

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/statusengine/interface/internal/domain"
)

// ErrNotImplemented is returned by every method. It is a distinct error
// so the API can answer 501 with something an operator can act on,
// rather than a generic failure that looks like a bug.
var ErrNotImplemented = errors.New(
	"the graphite metrics provider is not implemented yet; set metrics_provider to mysql")

// Provider will read from a Graphite render API.
type Provider struct {
	baseURL string
	prefix  string
}

// New returns a Provider for a Graphite instance.
func New(baseURL, prefix string) *Provider {
	return &Provider{baseURL: strings.TrimRight(baseURL, "/"), prefix: prefix}
}

// Name identifies this provider in a response.
func (p *Provider) Name() string { return "graphite" }

// Labels will list a service's series by walking the metrics tree.
func (p *Provider) Labels(context.Context, string, string) ([]domain.MetricMeta, error) {
	return nil, ErrNotImplemented
}

// Query will fetch and downsample series from the render API.
func (p *Provider) Query(context.Context, domain.MetricQuery) (domain.MetricResult, error) {
	return domain.MetricResult{}, ErrNotImplemented
}

// MetricPath builds the Graphite path for one series, using the same
// scheme the worker writes with: prefix.host.service.label, with every
// segment sanitised the same way.
//
// This is here now because it is the one piece that has to agree with
// the worker exactly, and agreeing with it is easier to get right while
// the worker's code is in front of us than later.
func (p *Provider) MetricPath(hostname, description, label string) string {
	parts := []string{p.prefix, sanitise(hostname), sanitise(description), sanitise(label)}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, ".")
}

// sanitise replaces the characters Graphite treats as path structure.
// A service called `C:\ Drive Space` has to survive this, and a dot in
// a hostname must not silently become another level of the tree.
func sanitise(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

// renderURL is the request shape the finished client will use.
func (p *Provider) renderURL(targets []string, from, to int64, maxPoints int) string {
	q := url.Values{}
	for _, t := range targets {
		q.Add("target", t)
	}
	q.Set("from", fmt.Sprintf("%d", from))
	q.Set("until", fmt.Sprintf("%d", to))
	q.Set("maxDataPoints", fmt.Sprintf("%d", maxPoints))
	q.Set("format", "json")
	return p.baseURL + "/render?" + q.Encode()
}
