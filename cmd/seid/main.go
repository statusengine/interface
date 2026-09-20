// Command seid is the Statusengine Web Interface daemon. It serves the API
// and the bundled frontend, and carries the handful of subcommands an
// operator needs before a browser is useful.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/commands"
	"github.com/statusengine/interface/internal/config"
	"github.com/statusengine/interface/internal/database"
	"github.com/statusengine/interface/internal/events"
	"github.com/statusengine/interface/internal/httpapi"
	"github.com/statusengine/interface/internal/logging"
	"github.com/statusengine/interface/internal/migrate"
	"github.com/statusengine/interface/internal/webui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "seid: "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	// A subcommand is a bare first argument; anything starting with "-"
	// is a flag for the default action, serve.
	cmd := "serve"
	rest := args
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		cmd, rest = args[0], args[1:]
	}

	switch cmd {
	case "serve":
		return serve(rest)
	case "migrate":
		return migrateOnly(rest)
	case "user":
		return userCommand(rest)
	case "version":
		fmt.Println(httpapi.Version)
		return nil
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `seid - Statusengine Web Interface

Usage:
  seid [serve] [flags]        Run the API and web server (default)
  seid migrate [flags]        Apply the sei_* schema and exit
  seid user <sub> [flags]     Manage accounts: create, passwd, role, list
  seid version                Print the build version

Run "seid serve -h" or "seid user create -h" for the flags of each.
`)
}

func serve(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}
	log, err := logging.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, database.Options{
		DSN:          cfg.MySQLDSN,
		MaxOpenConns: cfg.MySQLMaxOpenConns,
		ConnMaxLife:  cfg.MySQLConnMaxLife,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrate.Run(ctx, db, log); err != nil {
		return err
	}

	store := auth.NewStore(db)
	if err := auth.Bootstrap(ctx, store, log, auth.BootstrapOptions{
		DemoMode: cfg.DemoMode, DemoUser: cfg.DemoUser, DemoCommands: cfg.DemoCommands,
	}); err != nil {
		return err
	}

	authSvc := auth.NewService(store, log, auth.Options{
		SessionTTL:      cfg.SessionTTL,
		LoginRateLimit:  cfg.LoginRateLimit,
		LoginRateWindow: cfg.LoginRateWindow,
		DemoUser:        demoUserOrEmpty(cfg),
		DemoCommands:    cfg.DemoCommands,
	})
	go authSvc.PruneSessions(ctx, time.Hour)

	users, err := store.ListUsers(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		log.Warn("no accounts exist yet - create the first one with: seid user create -username <name> -role admin")
	}
	if !cfg.CommandsEnabled() {
		log.Warn("external commands are disabled: set worker_command_key to enable them")
	}
	if !cfg.EventsEnabled() {
		log.Warn("live updates are disabled: set worker_events_key to enable them; the UI will poll instead")
	}
	if cfg.DemoMode {
		if len(cfg.DemoCommands) > 0 {
			log.Warn("the public demo account may submit external commands",
				"commands", strings.Join(cfg.DemoCommands, ","),
				"per_visitor_per_minute", cfg.DemoCommandRateLimit,
				"max_objects_per_command", cfg.DemoMaxTargets)
		} else {
			log.Info("demo mode is on and read-only", "account", cfg.DemoUser)
		}
		if !cfg.SecureCookies && !isLoopback(cfg.ListenAddr) {
			log.Warn("demo mode is reachable off this machine without secure_cookies: " +
				"the session cookie will travel in the clear unless something in front sets it right")
		}
	}

	// One WebSocket to the worker for the whole process, fanned out to
	// browsers over SSE. Nil when no key is configured, which the
	// endpoint reports so the UI can poll instead.
	var hub *events.Hub
	if cfg.EventsEnabled() {
		hub = events.NewHub(log, events.HubOptions{MaxClients: cfg.MaxEventClients})
		go hub.Run(ctx)
		go events.NewWorkerSource(cfg.WorkerEventsURL, cfg.WorkerEventsKey, hub, log).Run(ctx)
	}

	if cfg.AuditRetentionDays > 0 {
		keep := time.Duration(cfg.AuditRetentionDays) * 24 * time.Hour
		go commands.NewAudit(db, log).PruneEvery(ctx, 6*time.Hour, keep)
	}

	srv := httpapi.New(httpapi.Options{
		Config: cfg,
		Logger: log,
		DB:     db,
		Auth:   authSvc,
		UI:     resolveUI(cfg, log),
		Events: hub,
	})
	return srv.ListenAndServe(ctx)
}

func migrateOnly(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}
	log, err := logging.New(os.Stderr, cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.Open(ctx, database.Options{
		DSN:          cfg.MySQLDSN,
		MaxOpenConns: 2,
		ConnMaxLife:  cfg.MySQLConnMaxLife,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrate.Run(ctx, db, log); err != nil {
		return err
	}
	store := auth.NewStore(db)
	if err := auth.Bootstrap(ctx, store, log, auth.BootstrapOptions{
		DemoMode: cfg.DemoMode, DemoUser: cfg.DemoUser, DemoCommands: cfg.DemoCommands,
	}); err != nil {
		return err
	}
	log.Info("schema is up to date")
	return nil
}

// demoUserOrEmpty keeps the demo login path unreachable unless demo mode
// is actually on, rather than relying on the handler to remember.
func demoUserOrEmpty(cfg config.Config) string {
	if !cfg.DemoMode {
		return ""
	}
	return cfg.DemoUser
}

func resolveUI(cfg config.Config, log *slog.Logger) fs.FS {
	if cfg.UIDir != "" {
		dir := webui.Dir(cfg.UIDir)
		if _, err := fs.Stat(dir, "index.html"); err != nil {
			log.Warn("ui_dir has no index.html; serving no frontend", "dir", cfg.UIDir)
			return nil
		}
		log.Info("serving frontend from disk", "dir", cfg.UIDir)
		return dir
	}
	ui := webui.FS()
	if ui == nil {
		log.Info("no frontend embedded in this build; API only")
	}
	return ui
}

var errNoSuchUser = errors.New("no such user")

// isLoopback reports whether a listen address only accepts connections
// from this machine, which is the case where the plaintext warnings do
// not apply.
func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	if host == "" {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
