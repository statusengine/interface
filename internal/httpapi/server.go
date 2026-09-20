package httpapi

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/config"
)

// Server holds everything the handlers need. Dependencies are values, not
// package globals, so a test can build one with fakes.
type Server struct {
	cfg  config.Config
	log  *slog.Logger
	db   *sql.DB
	auth *auth.Service

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
	Auth   *auth.Service
	UI     fs.FS
}

// New builds the Server and its routing table.
func New(opt Options) *Server {
	s := &Server{
		cfg:  opt.Config,
		log:  opt.Logger,
		db:   opt.DB,
		auth: opt.Auth,
		ui:   opt.UI,
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

	root := http.NewServeMux()
	root.Handle("/api/", api)
	root.Handle("/", s.uiHandler())

	return chain(root,
		requestIDMiddleware,
		loggingMiddleware(s.log),
		recoverMiddleware(s.log),
		securityHeadersMiddleware,
	)
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
