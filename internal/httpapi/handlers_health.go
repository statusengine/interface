package httpapi

import (
	"context"
	"net/http"
	"time"
)

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// handleHealthz answers as long as the process is up. An orchestrator uses
// it to decide whether to restart us, so it must not fail on a dependency
// being briefly unavailable.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// handleReadyz reports whether we can actually serve: MySQL must answer.
// The worker is checked too but does not make us unready - the monitoring
// data still reads fine without it, only commands and live updates stop.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	status := http.StatusOK

	if err := s.db.PingContext(ctx); err != nil {
		checks["mysql"] = "unreachable: " + err.Error()
		status = http.StatusServiceUnavailable
	} else {
		checks["mysql"] = "ok"
	}

	switch {
	case !s.cfg.CommandsEnabled():
		checks["worker_commands"] = "disabled: no worker_command_key configured"
	default:
		checks["worker_commands"] = "configured"
	}
	if !s.cfg.EventsEnabled() {
		checks["worker_events"] = "disabled: no worker_events_key configured"
	} else {
		checks["worker_events"] = "configured"
	}

	body := healthResponse{Status: "ok", Checks: checks}
	if status != http.StatusOK {
		body.Status = "degraded"
	}
	writeJSON(w, status, body)
}

// metaResponse is what the login page needs before anyone is logged in.
// It deliberately says nothing about which accounts exist.
type metaResponse struct {
	Product  string `json:"product"`
	Version  string `json:"version"`
	DemoMode bool   `json:"demo_mode"`
	// DemoCommands names what the public account may submit, so the
	// login page can describe it instead of promising read-only.
	DemoCommands    []string `json:"demo_commands,omitempty"`
	CommandsEnabled bool     `json:"commands_enabled"`
	EventsEnabled   bool     `json:"events_enabled"`
	MetricsProvider string   `json:"metrics_provider"`
	DefaultPageSize int      `json:"default_page_size"`
	MaxPageSize     int      `json:"max_page_size"`
}

// Version is set at build time with -ldflags "-X ...Version=v1.2.3".
var Version = "dev"

// demoCommands is what the public account may submit, for a login page
// that has to describe it accurately. Empty unless demo mode is on, so
// the list cannot claim anything about an account nobody can reach.
func (s *Server) demoCommands() []string {
	if !s.cfg.DemoMode {
		return nil
	}
	return s.cfg.DemoCommands
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, metaResponse{
		Product:         "Statusengine Web Interface",
		Version:         Version,
		DemoMode:        s.cfg.DemoMode,
		DemoCommands:    s.demoCommands(),
		CommandsEnabled: s.cfg.CommandsEnabled(),
		EventsEnabled:   s.cfg.EventsEnabled(),
		MetricsProvider: s.cfg.MetricsProvider,
		DefaultPageSize: s.cfg.DefaultPageSize,
		MaxPageSize:     s.cfg.MaxPageSize,
	})
}
