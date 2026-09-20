// Package config loads the daemon's settings from, in increasing order of
// precedence, built-in defaults, a YAML file, the environment and command
// line flags. That is the same order the Statusengine worker uses, so an
// operator who knows one knows the other.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the fully resolved configuration. Every field has a usable
// default except MySQLDSN and the worker API keys, which have no sensible
// value we could invent.
type Config struct {
	// HTTP server
	ListenAddr   string        `yaml:"listen_addr"`
	BaseURL      string        `yaml:"base_url"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`

	// MySQL
	MySQLDSN          string        `yaml:"mysql_dsn"`
	MySQLMaxOpenConns int           `yaml:"mysql_max_open_conns"`
	MySQLConnMaxLife  time.Duration `yaml:"mysql_conn_max_lifetime"`

	// QueryTimeout bounds one API request's database work. Without it a
	// query that cannot finish holds a pooled connection and a browser
	// tab for as long as the server is willing to wait, which is
	// forever. The live event stream is exempt: it is meant to stay open.
	QueryTimeout time.Duration `yaml:"query_timeout"`

	// Statusengine worker
	WorkerCommandURL string        `yaml:"worker_command_url"`
	WorkerCommandKey string        `yaml:"worker_command_key"`
	WorkerEventsURL  string        `yaml:"worker_events_url"`
	WorkerEventsKey  string        `yaml:"worker_events_key"`
	WorkerTimeout    time.Duration `yaml:"worker_timeout"`

	// Sessions and auth
	SessionTTL      time.Duration `yaml:"session_ttl"`
	SessionCookie   string        `yaml:"session_cookie"`
	SecureCookies   bool          `yaml:"secure_cookies"`
	DemoMode        bool          `yaml:"demo_mode"`
	DemoUser        string        `yaml:"demo_user"`
	LoginRateLimit  int           `yaml:"login_rate_limit"`
	LoginRateWindow time.Duration `yaml:"login_rate_window"`

	// DemoCommands is the allowlist of external commands the demo
	// account may submit. Empty, the default, means none: a
	// passwordless account on the open internet gets nothing that
	// reaches the monitoring core unless somebody says so by name.
	//
	// "notify" is refused here whatever it is set to. A custom
	// notification leaves the building - it mails and pages the real
	// contacts of whatever is being monitored - and no demo is worth
	// handing that to strangers.
	DemoCommands []string `yaml:"demo_commands"`

	// CommandRateLimit bounds how many command requests one session may
	// send per minute, and DemoCommandRateLimit does the same for the
	// demo account, which is shared by everybody who clicks "try it".
	// Zero means the default; a negative number switches the limit off.
	CommandRateLimit     int `yaml:"command_rate_limit"`
	DemoCommandRateLimit int `yaml:"demo_command_rate_limit"`

	// DemoMaxTargets caps how many objects one command from the demo
	// account may address. A bulk of a thousand is a feature for an
	// operator and a stampede from a stranger.
	DemoMaxTargets int `yaml:"demo_max_targets"`

	// AuditClientIP records the caller's address with each command.
	// True is right for an internal deployment, where the question is
	// who did this. A public demo records strangers, so an operator can
	// turn it off and keep the rest of the trail.
	AuditClientIP bool `yaml:"audit_client_ip"`

	// AuditRetentionDays drops audit rows older than this. Zero keeps
	// them forever, which is what an internal deployment usually wants
	// and a public one usually does not.
	AuditRetentionDays int `yaml:"audit_retention_days"`

	// MaxEventClients caps concurrent event-stream connections. Each is
	// cheap, but nothing else stops one client from opening them until
	// the process runs out of file descriptors.
	MaxEventClients int `yaml:"max_event_clients"`

	// Metrics
	MetricsProvider string `yaml:"metrics_provider"`
	GraphiteURL     string `yaml:"graphite_url"`
	GraphitePrefix  string `yaml:"graphite_prefix"`

	// Behaviour
	DefaultPageSize int    `yaml:"default_page_size"`
	MaxPageSize     int    `yaml:"max_page_size"`
	LogLevel        string `yaml:"log_level"`
	LogFormat       string `yaml:"log_format"`

	// UIDir serves the frontend from disk instead of the embedded bundle,
	// for iterating on the UI against a production-mode backend.
	UIDir string `yaml:"ui_dir"`
}

// Default returns the built-in configuration. The listen address stays on
// loopback: this process holds a key that can drive the monitoring core, so
// reaching the network is a decision someone makes, not one that happens by
// leaving a setting alone.
func Default() Config {
	return Config{
		ListenAddr:   "127.0.0.1:8090",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,

		MySQLMaxOpenConns: 25,
		MySQLConnMaxLife:  5 * time.Minute,
		QueryTimeout:      20 * time.Second,

		WorkerCommandURL: "http://127.0.0.1:8081/commands",
		WorkerEventsURL:  "ws://127.0.0.1:8080/ws",
		WorkerTimeout:    10 * time.Second,

		SessionTTL:      12 * time.Hour,
		SessionCookie:   "sei_session",
		SecureCookies:   false,
		DemoMode:        false,
		DemoUser:        "guest",
		LoginRateLimit:  10,
		LoginRateWindow: time.Minute,

		AuditClientIP:        true,
		AuditRetentionDays:   0,
		DemoCommands:         nil,
		CommandRateLimit:     60,
		DemoCommandRateLimit: 10,
		DemoMaxTargets:       25,
		MaxEventClients:      500,

		MetricsProvider: "mysql",
		GraphitePrefix:  "statusengine",

		DefaultPageSize: 50,
		MaxPageSize:     500,
		LogLevel:        "info",
		LogFormat:       "text",
	}
}

// Load resolves the configuration from every source. args are the command
// line arguments without the program name; pass os.Args[1:].
func Load(args []string) (Config, error) {
	cfg := Default()

	// The config file path itself can only come from a flag or the
	// environment - it cannot live inside the file it names.
	path := os.Getenv("SEI_CONFIG")
	fs := flag.NewFlagSet("seid", flag.ContinueOnError)
	fs.StringVar(&path, "config", path, "path to a YAML configuration file")

	// Parse once only to discover -config, ignoring everything else; the
	// real parse happens below, after the file has been merged in.
	pre := flag.NewFlagSet("seid-pre", flag.ContinueOnError)
	pre.SetOutput(nopWriter{})
	pre.StringVar(&path, "config", path, "")
	_ = pre.Parse(args)

	if path != "" {
		if err := cfg.mergeFile(path); err != nil {
			return cfg, err
		}
	}
	if err := cfg.mergeEnv(); err != nil {
		return cfg, err
	}
	cfg.bindFlags(fs)
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c *Config) mergeFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: reading %s: %w", path, err)
	}
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true) // a typo in a key is a startup error, not a silent default
	if err := dec.Decode(c); err != nil {
		return fmt.Errorf("config: parsing %s: %w", path, err)
	}
	return nil
}

// envBindings maps an environment variable to the field it sets. Keeping it
// as data rather than a run of if-statements means the list can be printed,
// and a new setting is one line in three places instead of five.
func (c *Config) mergeEnv() error {
	strs := map[string]*string{
		"SEI_LISTEN_ADDR":        &c.ListenAddr,
		"SEI_BASE_URL":           &c.BaseURL,
		"SEI_MYSQL_DSN":          &c.MySQLDSN,
		"SEI_WORKER_COMMAND_URL": &c.WorkerCommandURL,
		"SEI_WORKER_COMMAND_KEY": &c.WorkerCommandKey,
		"SEI_WORKER_EVENTS_URL":  &c.WorkerEventsURL,
		"SEI_WORKER_EVENTS_KEY":  &c.WorkerEventsKey,
		"SEI_SESSION_COOKIE":     &c.SessionCookie,
		"SEI_DEMO_USER":          &c.DemoUser,
		"SEI_METRICS_PROVIDER":   &c.MetricsProvider,
		"SEI_GRAPHITE_URL":       &c.GraphiteURL,
		"SEI_GRAPHITE_PREFIX":    &c.GraphitePrefix,
		"SEI_LOG_LEVEL":          &c.LogLevel,
		"SEI_LOG_FORMAT":         &c.LogFormat,
		"SEI_UI_DIR":             &c.UIDir,
	}
	for k, p := range strs {
		if v, ok := os.LookupEnv(k); ok {
			*p = v
		}
	}

	ints := map[string]*int{
		"SEI_MYSQL_MAX_OPEN_CONNS":    &c.MySQLMaxOpenConns,
		"SEI_LOGIN_RATE_LIMIT":        &c.LoginRateLimit,
		"SEI_COMMAND_RATE_LIMIT":      &c.CommandRateLimit,
		"SEI_DEMO_COMMAND_RATE_LIMIT": &c.DemoCommandRateLimit,
		"SEI_DEMO_MAX_TARGETS":        &c.DemoMaxTargets,
		"SEI_MAX_EVENT_CLIENTS":       &c.MaxEventClients,
		"SEI_AUDIT_RETENTION_DAYS":    &c.AuditRetentionDays,
		"SEI_DEFAULT_PAGE_SIZE":       &c.DefaultPageSize,
		"SEI_MAX_PAGE_SIZE":           &c.MaxPageSize,
	}
	for k, p := range ints {
		v, ok := os.LookupEnv(k)
		if !ok {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("config: %s: %q is not a number", k, v)
		}
		*p = n
	}

	if v, ok := os.LookupEnv("SEI_DEMO_COMMANDS"); ok {
		c.DemoCommands = nil
		for _, name := range strings.Split(v, ",") {
			if name = strings.TrimSpace(name); name != "" {
				c.DemoCommands = append(c.DemoCommands, name)
			}
		}
	}

	bools := map[string]*bool{
		"SEI_SECURE_COOKIES":  &c.SecureCookies,
		"SEI_DEMO_MODE":       &c.DemoMode,
		"SEI_AUDIT_CLIENT_IP": &c.AuditClientIP,
	}
	for k, p := range bools {
		v, ok := os.LookupEnv(k)
		if !ok {
			continue
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("config: %s: %q is not a boolean", k, v)
		}
		*p = b
	}

	durs := map[string]*time.Duration{
		"SEI_SESSION_TTL":    &c.SessionTTL,
		"SEI_WORKER_TIMEOUT": &c.WorkerTimeout,
		"SEI_QUERY_TIMEOUT":  &c.QueryTimeout,
	}
	for k, p := range durs {
		v, ok := os.LookupEnv(k)
		if !ok {
			continue
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("config: %s: %q is not a duration", k, v)
		}
		*p = d
	}
	return nil
}

// bindFlags registers flags whose defaults are the values resolved so far,
// which is what makes a flag win over the file and the environment.
func (c *Config) bindFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.ListenAddr, "listen-addr", c.ListenAddr, "address the HTTP server binds to")
	fs.StringVar(&c.MySQLDSN, "mysql-dsn", c.MySQLDSN, "MySQL data source name (go-sql-driver format)")
	fs.StringVar(&c.WorkerCommandURL, "worker-command-url", c.WorkerCommandURL, "full URL of the worker's /commands endpoint")
	fs.StringVar(&c.WorkerCommandKey, "worker-command-key", c.WorkerCommandKey, "API key for the worker's /commands endpoint")
	fs.StringVar(&c.WorkerEventsURL, "worker-events-url", c.WorkerEventsURL, "full URL of the worker's /ws endpoint")
	fs.StringVar(&c.WorkerEventsKey, "worker-events-key", c.WorkerEventsKey, "API key for the worker's /ws endpoint")
	fs.BoolVar(&c.DemoMode, "demo-mode", c.DemoMode, "offer the read-only demo account on the login page")
	fs.BoolVar(&c.SecureCookies, "secure-cookies", c.SecureCookies, "set the Secure attribute on the session cookie")
	fs.StringVar(&c.MetricsProvider, "metrics-provider", c.MetricsProvider, `performance data source: "mysql" or "graphite"`)
	fs.StringVar(&c.UIDir, "ui-dir", c.UIDir, "serve the frontend from this directory instead of the embedded bundle")
	fs.StringVar(&c.LogLevel, "log-level", c.LogLevel, "debug, info, warn or error")
	fs.StringVar(&c.LogFormat, "log-format", c.LogFormat, "text or json")
}

// Validate rejects a configuration that cannot work, at startup rather than
// at the first request that depends on it.
func (c *Config) Validate() error {
	var errs []error
	if c.MySQLDSN == "" {
		errs = append(errs, errors.New("mysql_dsn is required"))
	}
	if c.ListenAddr == "" {
		errs = append(errs, errors.New("listen_addr is required"))
	}
	switch c.MetricsProvider {
	case "mysql":
	case "graphite":
		if c.GraphiteURL == "" {
			errs = append(errs, errors.New(`metrics_provider is "graphite" but graphite_url is empty`))
		}
	default:
		errs = append(errs, fmt.Errorf("metrics_provider: %q is not one of mysql, graphite", c.MetricsProvider))
	}
	switch c.LogFormat {
	case "text", "json":
	default:
		errs = append(errs, fmt.Errorf("log_format: %q is not one of text, json", c.LogFormat))
	}
	if c.DefaultPageSize < 1 {
		errs = append(errs, errors.New("default_page_size must be at least 1"))
	}
	if c.MaxPageSize < c.DefaultPageSize {
		errs = append(errs, errors.New("max_page_size must be at least default_page_size"))
	}
	if c.SessionTTL <= 0 {
		errs = append(errs, errors.New("session_ttl must be positive"))
	}
	for _, name := range c.DemoCommands {
		switch {
		case name == DemoCommandNotify:
			errs = append(errs, errors.New(
				`demo_commands must not contain "notify": a custom notification mails and pages `+
					`the real contacts of the monitored objects, and the demo account is public`))
		case !isDemoCommand(name):
			errs = append(errs, fmt.Errorf("demo_commands: %q is not one of %s",
				name, strings.Join(DemoCommandNames(), ", ")))
		}
	}
	if c.DemoMaxTargets < 1 {
		errs = append(errs, errors.New("demo_max_targets must be at least 1"))
	}
	if c.AuditRetentionDays < 0 {
		errs = append(errs, errors.New("audit_retention_days must be zero or more"))
	}
	if c.MaxEventClients < 1 {
		errs = append(errs, errors.New("max_event_clients must be at least 1"))
	}
	if c.QueryTimeout <= 0 {
		errs = append(errs, errors.New("query_timeout must be positive"))
	}
	return errors.Join(errs...)
}

// CommandsEnabled reports whether external commands can be submitted at all.
// Without a key the worker leaves /commands unserved, so the UI should say
// so up front instead of offering buttons that always fail.
func (c *Config) CommandsEnabled() bool {
	return c.WorkerCommandKey != "" && c.WorkerCommandURL != ""
}

// EventsEnabled reports whether the live event stream can be consumed.
func (c *Config) EventsEnabled() bool {
	return c.WorkerEventsKey != "" && c.WorkerEventsURL != ""
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// The command actions an operator may hand to the demo account. They are
// named here rather than taken from the command package so that adding a
// command cannot silently widen what a public demo can do: a new action
// has to be added to this list on purpose.
const (
	// Each name is one right, not one route: acknowledging and removing
	// an acknowledgement are the same right, as are scheduling a
	// downtime and cancelling one.
	DemoCommandAcknowledge  = "acknowledge"
	DemoCommandDowntime     = "downtime"
	DemoCommandReschedule   = "reschedule"
	DemoCommandSubmitResult = "submit-result"
	DemoCommandToggle       = "toggle"

	// DemoCommandNotify is listed only so it can be refused by name.
	DemoCommandNotify = "notify"
)

// DemoCommandNames returns every action that may appear in
// demo_commands.
func DemoCommandNames() []string {
	return []string{
		DemoCommandAcknowledge, DemoCommandDowntime, DemoCommandReschedule,
		DemoCommandSubmitResult, DemoCommandToggle,
	}
}

func isDemoCommand(name string) bool {
	for _, known := range DemoCommandNames() {
		if known == name {
			return true
		}
	}
	return false
}

// AllowsDemoCommand reports whether the demo account may submit one.
func (c Config) AllowsDemoCommand(name string) bool {
	if !c.DemoMode {
		return false
	}
	for _, allowed := range c.DemoCommands {
		if allowed == name {
			return true
		}
	}
	return false
}
