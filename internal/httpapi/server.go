package httpapi

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/commands"
	"github.com/statusengine/interface/internal/config"
	"github.com/statusengine/interface/internal/events"
	"github.com/statusengine/interface/internal/metrics"
	"github.com/statusengine/interface/internal/metrics/graphite"
	"github.com/statusengine/interface/internal/metrics/mysqlprov"
	repo "github.com/statusengine/interface/internal/repository/mysql"
)

// Authenticator is the part of the auth service the HTTP layer uses.
//
// An interface rather than the concrete service so a handler test can
// make authentication fail in a specific way - which is the only way to
// pin down that a database outage answers 503 and not 401.
type Authenticator interface {
	Login(ctx context.Context, username, password, userAgent, ip string) (string, auth.Identity, error)
	LoginDemo(ctx context.Context, userAgent, ip string) (string, auth.Identity, error)
	Authenticate(ctx context.Context, token string) (auth.Identity, error)
	Logout(ctx context.Context, token string) error
	SessionTTL() time.Duration
}

// Server holds everything the handlers need. Dependencies are values, not
// package globals, so a test can build one with fakes.
type Server struct {
	cfg  config.Config
	log  *slog.Logger
	db   *sql.DB
	auth Authenticator

	// Repositories over the worker's tables. Constructed here rather
	// than injected: they are thin and stateless, and a handler test
	// exercises them against a real schema anyway.
	hosts     *repo.Hosts
	services  *repo.Services
	problems  *repo.Problems
	downtimes *repo.Downtimes
	acks      *repo.Acknowledgements
	logs      *repo.LogEntries
	summary   *repo.Summary
	history   *repo.History

	// metrics is chosen by configuration. The interface is the point:
	// the endpoint, the chart and the downsampling contract do not know
	// which backend answered.
	metrics metrics.Provider

	// commands carries operator actions to the monitoring core, and
	// audit records every one of them - including the refusals.
	commands commands.Transport
	audit    *commands.Audit

	// events fans the worker's stream out to browsers. Nil when live
	// updates are switched off, in which case the endpoint says so and
	// the UI polls.
	events *events.Hub

	// ui is the built frontend. It may be nil during development, when
	// the Angular dev server serves the UI and proxies /api here.
	ui fs.FS

	handler http.Handler
}

// Options are the Server's dependencies.
type Options struct {
	Config config.Config
	Logger *slog.Logger
	DB     *sql.DB
	Auth   Authenticator
	UI     fs.FS
	Events *events.Hub
}

// New builds the Server and its routing table.
func New(opt Options) *Server {
	s := &Server{
		cfg:  opt.Config,
		log:  opt.Logger,
		db:   opt.DB,
		auth: opt.Auth,
		ui:   opt.UI,

		hosts:     repo.NewHosts(opt.DB),
		services:  repo.NewServices(opt.DB),
		problems:  repo.NewProblems(opt.DB),
		downtimes: repo.NewDowntimes(opt.DB),
		acks:      repo.NewAcknowledgements(opt.DB),
		logs:      repo.NewLogEntries(opt.DB),
		summary:   repo.NewSummary(opt.DB),
		history:   repo.NewHistory(opt.DB),

		metrics: newMetricsProvider(opt.Config, opt.DB),

		commands: commands.NewClient(
			opt.Config.WorkerCommandURL, opt.Config.WorkerCommandKey, opt.Config.WorkerTimeout),
		audit:  commands.NewAudit(opt.DB),
		events: opt.Events,
	}
	s.handler = s.routes()
	return s
}

// ServeHTTP makes the Server an http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) routes() http.Handler {
	api := http.NewServeMux()

	// Unauthenticated: health for orchestrators, and the handful of facts
	// the login page needs before anyone has logged in.
	api.HandleFunc("GET /api/v1/healthz", s.handleHealthz)
	api.HandleFunc("GET /api/v1/readyz", s.handleReadyz)
	api.HandleFunc("GET /api/v1/meta", s.handleMeta)

	api.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	api.HandleFunc("POST /api/v1/auth/login/demo", s.handleLoginDemo)
	api.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)

	// Authenticated. Each route names the permission it needs right here,
	// so the access rules are readable as a table rather than scattered
	// through the handlers.
	authed := func(perm string, h http.HandlerFunc) http.Handler {
		if perm == "" {
			return chain(h, s.authMiddleware)
		}
		return chain(h, s.authMiddleware, requirePermission(perm))
	}

	api.Handle("GET /api/v1/auth/me", authed("", s.handleMe))

	api.Handle("GET /api/v1/summary", authed(auth.PermHostsRead, s.handleSummary))

	api.Handle("GET /api/v1/hosts", authed(auth.PermHostsRead, s.handleListHosts))
	api.Handle("GET /api/v1/hosts/names", authed(auth.PermHostsRead, s.handleListHostNames))
	api.Handle("GET /api/v1/hosts/{host}", authed(auth.PermHostsRead, s.handleGetHost))
	api.Handle("GET /api/v1/hosts/{host}/services", authed(auth.PermServicesRead, s.handleListServices))

	api.Handle("GET /api/v1/services", authed(auth.PermServicesRead, s.handleListServices))
	// Singular, and identified by query parameters: a Naemon service
	// description is free text and routinely contains slashes.
	api.Handle("GET /api/v1/service", authed(auth.PermServicesRead, s.handleGetService))

	api.Handle("GET /api/v1/problems", authed(auth.PermProblemsRead, s.handleListProblems))

	api.Handle("GET /api/v1/downtimes", authed(auth.PermDowntimesRead, s.handleListDowntimes))
	api.Handle("GET /api/v1/downtimes/history", authed(auth.PermDowntimesRead, s.handleListDowntimeHistory))

	api.Handle("GET /api/v1/acknowledgements", authed(auth.PermAcksRead, s.handleListAcknowledgements))

	api.Handle("GET /api/v1/logentries", authed(auth.PermLogEntriesRead, s.handleListLogEntries))

	api.Handle("GET /api/v1/history/checks", authed(auth.PermHistoryRead, s.handleHistoryChecks))
	api.Handle("GET /api/v1/history/statechanges", authed(auth.PermHistoryRead, s.handleHistoryStateChanges))
	api.Handle("GET /api/v1/history/notifications", authed(auth.PermHistoryRead, s.handleHistoryNotifications))

	api.Handle("GET /api/v1/metrics/labels", authed(auth.PermMetricsRead, s.handleMetricLabels))
	api.Handle("GET /api/v1/metrics/series", authed(auth.PermMetricsRead, s.handleMetricSeries))

	// Commands. Each route names the permission it needs right here, so
	// the rules read as a table rather than hiding inside the handlers.
	api.Handle("POST /api/v1/commands/acknowledge", authed(auth.PermCmdAcknowledge, s.handleAcknowledge))
	api.Handle("POST /api/v1/commands/remove-acknowledgement", authed(auth.PermCmdAcknowledge, s.handleRemoveAcknowledgement))
	api.Handle("POST /api/v1/commands/downtime", authed(auth.PermCmdDowntime, s.handleScheduleDowntime))
	api.Handle("POST /api/v1/commands/downtime/delete", authed(auth.PermCmdDowntime, s.handleDeleteDowntime))
	api.Handle("POST /api/v1/commands/reschedule", authed(auth.PermCmdReschedule, s.handleReschedule))
	api.Handle("POST /api/v1/commands/submit-result", authed(auth.PermCmdPassiveResult, s.handleSubmitResult))
	api.Handle("POST /api/v1/commands/notify", authed(auth.PermCmdNotification, s.handleNotify))
	api.Handle("POST /api/v1/commands/toggle-notifications", authed(auth.PermCmdToggle, s.handleToggleNotifications))
	api.Handle("POST /api/v1/commands/toggle-active-checks", authed(auth.PermCmdToggle, s.handleToggleActiveChecks))

	api.Handle("GET /api/v1/commands/audit", authed(auth.PermAuditRead, s.handleListAudit))

	// Live change notifications. Any signed-in user may watch: it says
	// what moved, never what it moved to, so it grants nothing the
	// read permissions do not.
	api.Handle("GET /api/v1/events", authed("", s.handleEvents))

	root := http.NewServeMux()
	root.Handle("/api/", api)
	root.Handle("/", s.uiHandler())

	return chain(root,
		requestIDMiddleware,
		loggingMiddleware(s.log),
		recoverMiddleware(s.log),
		securityHeadersMiddleware,
		timeoutMiddleware(s.cfg.QueryTimeout, isStreamingRequest),
	)
}

// isStreamingRequest marks the responses that are meant to stay open,
// so the request timeout does not cut them off.
func isStreamingRequest(r *http.Request) bool {
	return r.URL.Path == "/api/v1/events"
}

// handleEvents streams change notifications, or explains why it cannot.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if s.events == nil {
		// Not an error: a deployment without a worker events key is a
		// choice, and the client falls back to polling. Saying so beats
		// a dead stream the browser keeps retrying.
		writeError(w, http.StatusNotImplemented, CodeUnavailable,
			"live updates are switched off; set worker_events_key to enable them")
		return
	}
	events.NewHandler(s.events).ServeHTTP(w, r)
}

// newMetricsProvider picks the backend named in the configuration. The
// config is validated at startup, so an unknown name cannot reach here;
// MySQL is the fallback because it is the one that always works when a
// database is configured at all.
func newMetricsProvider(cfg config.Config, db *sql.DB) metrics.Provider {
	if cfg.MetricsProvider == "graphite" {
		return graphite.New(cfg.GraphiteURL, cfg.GraphitePrefix)
	}
	return mysqlprov.New(db)
}

// ListenAndServe runs the HTTP server until ctx is cancelled, then shuts
// it down gracefully.
func (s *Server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      s,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	errc := make(chan error, 1)
	go func() {
		s.log.Info("listening", "addr", s.cfg.ListenAddr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errc <- err
			return
		}
		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		s.log.Info("shutting down")
		return srv.Shutdown(shutdownCtx)
	}
}
