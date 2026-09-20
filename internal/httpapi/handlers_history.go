package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/statusengine/interface/internal/domain"
	"github.com/statusengine/interface/internal/metrics/graphite"
	repo "github.com/statusengine/interface/internal/repository/mysql"
)

// Window policy for the history endpoints.
//
// These tables are clustered on an object-first primary key, so naming a
// host turns the read into one contiguous range and the window can be
// generous. Without one, the rows for a given hour are scattered across
// the table in small per-object runs, and the only thing keeping the
// query honest is a much shorter window.
const (
	scopedHistoryDefault   = 24 * time.Hour
	scopedHistoryMax       = 90 * 24 * time.Hour
	unscopedHistoryDefault = time.Hour
	unscopedHistoryMax     = 6 * time.Hour
)

// historyFilter reads the parameters every history list shares.
func (s *Server) historyFilter(r *http.Request, maxState int) (repo.HistoryFilter, *apiError) {
	var f repo.HistoryFilter
	q := r.URL.Query()

	f.Host = strings.TrimSpace(q.Get("host"))
	f.Description = strings.TrimSpace(q.Get("service"))
	f.Search = strings.TrimSpace(q.Get("q"))

	if f.Description != "" && f.Host == "" {
		return f, &apiError{
			Code:    CodeBadRequest,
			Message: "a service filter needs a host as well; a service description is not unique on its own",
			Field:   "host",
		}
	}

	kind, apiErr := parseKind(r)
	if apiErr != nil {
		return f, apiErr
	}
	f.Kind = kind

	states, apiErr := parseIntList(r, "state", 0, maxState)
	if apiErr != nil {
		return f, apiErr
	}
	f.States = states

	f.HardOnly = q.Get("hard_only") == "true"
	f.TransitionsOnly = q.Get("transitions_only") == "true"

	defaultWindow, maxWindow := unscopedHistoryDefault, unscopedHistoryMax
	if f.Host != "" {
		defaultWindow, maxWindow = scopedHistoryDefault, scopedHistoryMax
	}
	tr, apiErr := parseTimeRange(r, defaultWindow, maxWindow)
	if apiErr != nil {
		// The ceiling depends on whether an object was named, so the
		// message says how to lift it rather than just stating the limit.
		if apiErr.Field == "from" && f.Host == "" {
			apiErr.Message += ". Naming a host raises the limit to 90 days, " +
				"because the history tables are clustered per object"
		}
		return f, apiErr
	}
	f.From, f.To = tr.From, tr.To

	return f, nil
}

// historyMeta adds the window and the scope to a list response, so the
// UI can state both instead of implying it is showing everything.
type historyMeta struct {
	ListMeta
	From   int64 `json:"from"`
	To     int64 `json:"to"`
	Scoped bool  `json:"scoped"`
}

type historyResponse struct {
	Data any         `json:"data"`
	Meta historyMeta `json:"meta"`
}

func writeHistory(w http.ResponseWriter, data any, meta ListMeta, f repo.HistoryFilter) {
	if isNilSlice(data) {
		data = []any{}
	}
	writeJSON(w, http.StatusOK, historyResponse{
		Data: data,
		Meta: historyMeta{ListMeta: meta, From: f.From, To: f.To, Scoped: f.Scoped()},
	})
}

func (s *Server) handleHistoryChecks(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.historyFilter(r, int(domain.ServiceUnknown))
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	p, apiErr := s.page(r, "start_time", repo.CheckSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if r.URL.Query().Get("sort") == "" {
		p.Desc = true
	}

	checks, total, err := s.history.Checks(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "the check history", err)
		return
	}
	writeHistory(w, checks, metaFor(p, total), f)
}

func (s *Server) handleHistoryStateChanges(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.historyFilter(r, int(domain.ServiceUnknown))
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	p, apiErr := s.page(r, "state_time", repo.StateChangeSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if r.URL.Query().Get("sort") == "" {
		p.Desc = true
	}

	changes, total, err := s.history.StateChanges(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "the state history", err)
		return
	}
	writeHistory(w, changes, metaFor(p, total), f)
}

func (s *Server) handleHistoryNotifications(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.historyFilter(r, int(domain.ServiceUnknown))
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	// The notification tables have no is_hardstate column.
	f.HardOnly = false

	p, apiErr := s.page(r, "start_time", repo.NotificationSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if r.URL.Query().Get("sort") == "" {
		p.Desc = true
	}

	notifications, total, err := s.history.Notifications(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "the notification history", err)
		return
	}
	writeHistory(w, notifications, metaFor(p, total), f)
}

// --- metrics ---------------------------------------------------------------

func (s *Server) handleMetricLabels(w http.ResponseWriter, r *http.Request) {
	host, apiErr := requiredQuery(r, "host")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	service, apiErr := requiredQuery(r, "service")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	labels, err := s.metrics.Labels(r.Context(), host, service)
	if err != nil {
		s.metricsError(w, r, err)
		return
	}
	writeList(w, labels, ListMeta{Total: int64(len(labels))})
}

func (s *Server) handleMetricSeries(w http.ResponseWriter, r *http.Request) {
	host, apiErr := requiredQuery(r, "host")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	service, apiErr := requiredQuery(r, "service")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	// Performance data is retained for months (age_perfdata defaults to
	// 90 days), and the query hits the `metric` index for one object, so
	// a long window is cheap here in a way it is not for the history
	// tables.
	tr, apiErr := parseTimeRange(r, 6*time.Hour, 366*24*time.Hour)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	maxPoints := 500
	if v := r.URL.Query().Get("max_points"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 10 || n > 5000 {
			fail(w, &apiError{
				Code:    CodeBadRequest,
				Message: "max_points must be between 10 and 5000",
				Field:   "max_points",
			})
			return
		}
		maxPoints = n
	}

	var labels []string
	if v := r.URL.Query().Get("label"); v != "" {
		for _, label := range strings.Split(v, ",") {
			if label = strings.TrimSpace(label); label != "" {
				labels = append(labels, label)
			}
		}
	}

	result, err := s.metrics.Query(r.Context(), domain.MetricQuery{
		Hostname:    host,
		Description: service,
		Labels:      labels,
		From:        tr.From,
		To:          tr.To,
		MaxPoints:   maxPoints,
	})
	if err != nil {
		s.metricsError(w, r, err)
		return
	}
	if result.Series == nil {
		result.Series = []domain.Series{}
	}
	writeJSON(w, http.StatusOK, result)
}

// metricsError turns a provider failure into something an operator can
// act on. A configured-but-unimplemented backend is a deployment
// decision, not a bug, and saying so beats a generic 500.
func (s *Server) metricsError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, graphite.ErrNotImplemented) {
		writeError(w, http.StatusNotImplemented, CodeUnavailable, err.Error())
		return
	}
	s.internalError(w, r, "performance data", err)
}
