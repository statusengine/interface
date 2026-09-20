package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/statusengine/interface/internal/domain"
	repo "github.com/statusengine/interface/internal/repository/mysql"
)

// sortableNames turns a whitelist map into the list an error message
// shows, so the API can tell a caller what it would have accepted.
func sortableNames(columns map[string]string) []string {
	out := make([]string, 0, len(columns))
	for name := range columns {
		out = append(out, name)
	}
	// Map order is random; a stable error message is worth the sort.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func (s *Server) page(r *http.Request, defaultSort string, columns map[string]string) (repo.Page, *apiError) {
	p, apiErr := parsePage(r, s.cfg.DefaultPageSize, s.cfg.MaxPageSize, defaultSort, sortableNames(columns))
	if apiErr != nil {
		return repo.Page{}, apiErr
	}
	return repo.Page{Limit: p.Limit, Offset: p.Offset, Sort: p.Sort, Desc: p.Desc}, nil
}

func metaFor(p repo.Page, total int64) ListMeta {
	spec := p.Sort
	if p.Desc {
		spec += ":desc"
	} else {
		spec += ":asc"
	}
	return ListMeta{Total: total, Limit: p.Limit, Offset: p.Offset, Sort: spec}
}

// internalError logs the detail and tells the client only that it
// failed. A SQL error in a response body is a description of our schema.
//
// A deadline is pulled out separately. It is not an internal error: the
// query was valid and the database was reachable, it just could not
// answer in time, and the useful reply says what to do differently.
func (s *Server) internalError(w http.ResponseWriter, r *http.Request, what string, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		loggerFrom(r.Context()).Warn("query exceeded the deadline",
			"what", what, "path", r.URL.Path, "query", r.URL.RawQuery, "timeout", s.cfg.QueryTimeout)
		writeError(w, http.StatusGatewayTimeout, CodeTimeout,
			"reading "+what+" took longer than "+s.cfg.QueryTimeout.String()+
				". Narrow the time window, name a host, or ask for fewer rows.")
		return
	}
	if errors.Is(err, context.Canceled) {
		// The browser navigated away mid-request. Nothing to report and
		// nobody to report it to.
		return
	}
	loggerFrom(r.Context()).Error(what, "error", err)
	writeError(w, http.StatusInternalServerError, CodeInternal, "could not read "+what)
}

// --- hosts -----------------------------------------------------------------

func (s *Server) statusFilter(r *http.Request, maxState int) (repo.StatusFilter, *apiError) {
	var f repo.StatusFilter
	q := r.URL.Query()

	f.Search = strings.TrimSpace(q.Get("q"))
	f.Host = strings.TrimSpace(q.Get("host"))
	f.ProblemsOnly = q.Get("problems") == "true"

	states, apiErr := parseIntList(r, "state", 0, maxState)
	if apiErr != nil {
		return f, apiErr
	}
	f.States = states

	for _, spec := range []struct {
		name string
		dst  **bool
	}{
		{"acknowledged", &f.Acknowledged},
		{"in_downtime", &f.InDowntime},
		{"flapping", &f.Flapping},
		{"notifications_enabled", &f.NotificationsEnabled},
		{"active_checks_enabled", &f.ActiveChecksEnabled},
		{"hard_state", &f.HardState},
		{"passive", &f.Passive},
		{"handled", &f.Handled},
	} {
		v, apiErr := parseTriBool(r, spec.name)
		if apiErr != nil {
			return f, apiErr
		}
		*spec.dst = v
	}
	return f, nil
}

func (s *Server) handleListHosts(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.statusFilter(r, int(domain.HostUnreachable))
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	p, apiErr := s.page(r, "hostname", repo.HostSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	hosts, total, err := s.hosts.List(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "hosts", err)
		return
	}
	writeList(w, hosts, metaFor(p, total))
}

// hostDetail is a host plus the records that explain why it is quiet.
//
// Composed here rather than left to the client. The two extra reads hit
// small tables keyed by hostname, and a detail page that has to make
// three requests makes them again on every live refresh.
type hostDetail struct {
	domain.HostStatus
	// Downtimes the core is holding on this host right now.
	Downtimes []domain.Downtime `json:"downtimes"`
	// The acknowledgement in force, when the host is acknowledged. The
	// table keeps removed ones too, so it is only read when the status
	// says one applies.
	Acknowledgement *domain.Acknowledgement `json:"acknowledgement,omitempty"`
}

func (s *Server) handleGetHost(w http.ResponseWriter, r *http.Request) {
	hostname := r.PathValue("host")

	host, err := s.hosts.Get(r.Context(), hostname)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, http.StatusNotFound, CodeNotFound, "no host named "+hostname)
		return
	}
	if err != nil {
		s.internalError(w, r, "the host", err)
		return
	}

	detail := hostDetail{HostStatus: host, Downtimes: []domain.Downtime{}}
	if host.InDowntime {
		downtimes, err := s.downtimes.ForObject(r.Context(), domain.KindHost, hostname, "")
		if err != nil {
			s.internalError(w, r, "the host's downtimes", err)
			return
		}
		detail.Downtimes = downtimes
	}
	if host.Acknowledged {
		ack, err := s.acks.LatestFor(r.Context(), domain.KindHost, hostname, "")
		if err != nil {
			s.internalError(w, r, "the host's acknowledgement", err)
			return
		}
		detail.Acknowledgement = ack
	}

	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleListHostNames(w http.ResponseWriter, r *http.Request) {
	names, err := s.hosts.Names(r.Context(), s.cfg.MaxPageSize)
	if err != nil {
		s.internalError(w, r, "host names", err)
		return
	}
	writeList(w, names, ListMeta{Total: int64(len(names)), Limit: s.cfg.MaxPageSize})
}

// --- services --------------------------------------------------------------

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.statusFilter(r, int(domain.ServiceUnknown))
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	// The nested route carries the host in the path; the flat one takes
	// it as a filter.
	if host := r.PathValue("host"); host != "" {
		f.Host = host
	}

	p, apiErr := s.page(r, "hostname", repo.ServiceSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	services, total, err := s.services.List(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "services", err)
		return
	}
	writeList(w, services, metaFor(p, total))
}

// handleGetService identifies the service by query parameters rather than
// by path segments.
//
// A Naemon service description is free text: this very installation has
// one called `C:\ Drive Space`, and others routinely contain slashes.
// Putting that in a path means percent-encoded separators, which proxies
// and routers disagree about. A query parameter has none of that problem.
func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request) {
	host, apiErr := requiredQuery(r, "host")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	description, apiErr := requiredQuery(r, "service")
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	service, err := s.services.Get(r.Context(), host, description)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, http.StatusNotFound, CodeNotFound,
			"no service "+description+" on "+host)
		return
	}
	if err != nil {
		s.internalError(w, r, "the service", err)
		return
	}

	detail := serviceDetail{ServiceStatus: service, Downtimes: []domain.Downtime{}}
	if service.InDowntime {
		downtimes, err := s.downtimes.ForObject(r.Context(), domain.KindService, host, description)
		if err != nil {
			s.internalError(w, r, "the service's downtimes", err)
			return
		}
		detail.Downtimes = downtimes
	}
	if service.Acknowledged {
		ack, err := s.acks.LatestFor(r.Context(), domain.KindService, host, description)
		if err != nil {
			s.internalError(w, r, "the service's acknowledgement", err)
			return
		}
		detail.Acknowledgement = ack
	}

	writeJSON(w, http.StatusOK, detail)
}

// serviceDetail is the service equivalent of hostDetail.
type serviceDetail struct {
	domain.ServiceStatus
	Downtimes       []domain.Downtime       `json:"downtimes"`
	Acknowledgement *domain.Acknowledgement `json:"acknowledgement,omitempty"`
}

// --- problems --------------------------------------------------------------

func (s *Server) handleListProblems(w http.ResponseWriter, r *http.Request) {
	var f repo.ProblemFilter
	q := r.URL.Query()
	f.Search = strings.TrimSpace(q.Get("q"))
	f.HideServicesOfDownHosts = q.Get("hide_services_of_down_hosts") == "true"

	kind, apiErr := parseKind(r)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	f.Kind = kind

	for _, spec := range []struct {
		name string
		dst  **bool
	}{
		{"acknowledged", &f.Acknowledged},
		{"in_downtime", &f.InDowntime},
		{"hard_state", &f.HardState},
		{"flapping", &f.Flapping},
		{"handled", &f.Handled},
	} {
		v, apiErr := parseTriBool(r, spec.name)
		if apiErr != nil {
			fail(w, apiErr)
			return
		}
		*spec.dst = v
	}

	p, apiErr := s.page(r, "severity", repo.ProblemSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	// Worst first unless the caller said otherwise: the list exists to be
	// worked from the top.
	if q.Get("sort") == "" {
		p.Desc = true
	}

	problems, total, err := s.problems.List(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "problems", err)
		return
	}
	writeList(w, problems, metaFor(p, total))
}

func parseKind(r *http.Request) (domain.Kind, *apiError) {
	switch v := r.URL.Query().Get("kind"); v {
	case "":
		return "", nil
	case string(domain.KindHost):
		return domain.KindHost, nil
	case string(domain.KindService):
		return domain.KindService, nil
	default:
		return "", &apiError{
			Code:    CodeInvalidFilter,
			Message: `kind must be "host" or "service"`,
			Field:   "kind",
		}
	}
}

// --- downtimes -------------------------------------------------------------

func (s *Server) downtimeFilter(r *http.Request) (repo.DowntimeFilter, *apiError) {
	var f repo.DowntimeFilter
	q := r.URL.Query()
	f.Search = strings.TrimSpace(q.Get("q"))
	f.Host = strings.TrimSpace(q.Get("host"))
	f.Service = strings.TrimSpace(q.Get("service"))
	f.Now = time.Now().Unix()

	if f.Service != "" && f.Host == "" {
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

	running, apiErr := parseTriBool(r, "running")
	if apiErr != nil {
		return f, apiErr
	}
	f.Running = running
	return f, nil
}

func (s *Server) handleListDowntimes(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.downtimeFilter(r)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	p, apiErr := s.page(r, "scheduled_start_time", repo.DowntimeSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	downtimes, total, err := s.downtimes.Current(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "downtimes", err)
		return
	}
	writeList(w, downtimes, metaFor(p, total))
}

func (s *Server) handleListDowntimeHistory(w http.ResponseWriter, r *http.Request) {
	f, apiErr := s.downtimeFilter(r)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	p, apiErr := s.page(r, "scheduled_start_time", repo.DowntimeSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if r.URL.Query().Get("sort") == "" {
		p.Desc = true // newest first, like every other history list
	}

	downtimes, total, err := s.downtimes.History(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "the downtime history", err)
		return
	}
	writeList(w, downtimes, metaFor(p, total))
}

// --- acknowledgements ------------------------------------------------------

func (s *Server) handleListAcknowledgements(w http.ResponseWriter, r *http.Request) {
	var f repo.AckFilter
	q := r.URL.Query()
	f.Search = strings.TrimSpace(q.Get("q"))
	f.Host = strings.TrimSpace(q.Get("host"))
	f.Service = strings.TrimSpace(q.Get("service"))
	f.Author = strings.TrimSpace(q.Get("author"))

	if f.Service != "" && f.Host == "" {
		fail(w, &apiError{
			Code:    CodeBadRequest,
			Message: "a service filter needs a host as well; a service description is not unique on its own",
			Field:   "host",
		})
		return
	}

	kind, apiErr := parseKind(r)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	f.Kind = kind

	// Acknowledgements are few and the table is small, so a window is
	// optional here. It is still offered, because "what did we
	// acknowledge last week" is a real question.
	if q.Get("from") != "" || q.Get("to") != "" {
		tr, apiErr := parseTimeRange(r, 30*24*time.Hour, 0)
		if apiErr != nil {
			fail(w, apiErr)
			return
		}
		f.From, f.To = tr.From, tr.To
	}

	p, apiErr := s.page(r, "entry_time", repo.AckSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if q.Get("sort") == "" {
		p.Desc = true
	}

	acks, total, err := s.acks.List(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "acknowledgements", err)
		return
	}
	writeList(w, acks, metaFor(p, total))
}

// --- log entries -----------------------------------------------------------

func (s *Server) handleListLogEntries(w http.ResponseWriter, r *http.Request) {
	// The worker ages log entries out after days rather than months
	// (age_logentries defaults to 5), so a day is a sensible default
	// window and a month is a generous ceiling.
	tr, apiErr := parseTimeRange(r, 24*time.Hour, 31*24*time.Hour)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	f := repo.LogFilter{
		Search: strings.TrimSpace(r.URL.Query().Get("q")),
		Node:   strings.TrimSpace(r.URL.Query().Get("node")),
		From:   tr.From,
		To:     tr.To,
	}

	p, apiErr := s.page(r, "entry_time", repo.LogSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	if r.URL.Query().Get("sort") == "" {
		p.Desc = true
	}

	entries, total, err := s.logs.List(r.Context(), f, p)
	if err != nil {
		s.internalError(w, r, "log entries", err)
		return
	}
	writeList(w, entries, metaFor(p, total))
}

// --- summary ---------------------------------------------------------------

// summaryWindowHours bounds how far back the dashboard's recent figures
// reach. A week is the longest span the notification index makes cheap,
// and a manager asking for a quarter wants a report, not a dashboard.
const (
	defaultSummaryHours = 24
	maxSummaryHours     = 168
)

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	hours := defaultSummaryHours
	if v := r.URL.Query().Get("hours"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxSummaryHours {
			writeFieldError(w, http.StatusBadRequest, CodeBadRequest,
				fmt.Sprintf("hours must be between 1 and %d", maxSummaryHours), "hours")
			return
		}
		hours = n
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
	summary, err := s.summary.Get(r.Context(), since)
	if err != nil {
		s.internalError(w, r, "the summary", err)
		return
	}
	summary.Window.Hours = hours
	writeJSON(w, http.StatusOK, summary)
}
