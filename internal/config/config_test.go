package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "seid.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPrecedenceFlagBeatsEnvBeatsFile(t *testing.T) {
	path := writeConfig(t, "listen_addr: 1.1.1.1:1\nmysql_dsn: from-file\n")

	t.Setenv("SEI_LISTEN_ADDR", "2.2.2.2:2")

	cfg, err := Load([]string{"-config", path, "-listen-addr", "3.3.3.3:3"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ListenAddr != "3.3.3.3:3" {
		t.Errorf("ListenAddr = %q, want the flag to win", cfg.ListenAddr)
	}
	if cfg.MySQLDSN != "from-file" {
		t.Errorf("MySQLDSN = %q, want the file value to survive", cfg.MySQLDSN)
	}
}

func TestEnvBeatsFile(t *testing.T) {
	path := writeConfig(t, "mysql_dsn: from-file\n")
	t.Setenv("SEI_MYSQL_DSN", "from-env")

	cfg, err := Load([]string{"-config", path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MySQLDSN != "from-env" {
		t.Errorf("MySQLDSN = %q, want from-env", cfg.MySQLDSN)
	}
}

func TestDefaultsSurviveAPartialFile(t *testing.T) {
	path := writeConfig(t, "mysql_dsn: dsn\n")

	cfg, err := Load([]string{"-config", path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SessionTTL != 12*time.Hour {
		t.Errorf("SessionTTL = %v, want the default", cfg.SessionTTL)
	}
	if cfg.ListenAddr != Default().ListenAddr {
		t.Errorf("ListenAddr = %q, want the default", cfg.ListenAddr)
	}
}

// A misspelled key that silently does nothing is the worst outcome for a
// config file, so it has to be an error.
func TestUnknownKeyIsAnError(t *testing.T) {
	path := writeConfig(t, "mysql_dsn: dsn\nlisten_addres: 1.2.3.4:5\n")

	_, err := Load([]string{"-config", path})
	if err == nil {
		t.Fatal("want an error for an unknown key, got nil")
	}
	if !strings.Contains(err.Error(), "listen_addres") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		mutate func(*Config)
		want   string
	}{
		"missing dsn":          {func(c *Config) { c.MySQLDSN = "" }, "mysql_dsn"},
		"unknown metrics":      {func(c *Config) { c.MetricsProvider = "influx" }, "metrics_provider"},
		"graphite without url": {func(c *Config) { c.MetricsProvider = "graphite" }, "graphite_url"},
		"bad log format":       {func(c *Config) { c.LogFormat = "xml" }, "log_format"},
		"page size inverted":   {func(c *Config) { c.DefaultPageSize = 100; c.MaxPageSize = 10 }, "max_page_size"},
		"zero session ttl":     {func(c *Config) { c.SessionTTL = 0 }, "session_ttl"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := Default()
			cfg.MySQLDSN = "dsn"
			tc.mutate(&cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatal("want an error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should mention %q", err, tc.want)
			}
		})
	}
}

func TestFeatureTogglesFollowTheKeys(t *testing.T) {
	cfg := Default()
	if cfg.CommandsEnabled() {
		t.Error("commands must be off until a key is configured")
	}
	if cfg.EventsEnabled() {
		t.Error("events must be off until a key is configured")
	}

	cfg.WorkerCommandKey = "k"
	cfg.WorkerEventsKey = "k"
	if !cfg.CommandsEnabled() || !cfg.EventsEnabled() {
		t.Error("both should be on once their keys are set")
	}
}

func TestListenAddrDefaultsToLoopback(t *testing.T) {
	// This process holds a key that can drive the monitoring core.
	// Reaching the network should be a decision, not a default.
	if got := Default().ListenAddr; !strings.HasPrefix(got, "127.0.0.1:") {
		t.Errorf("default ListenAddr = %q, want a loopback address", got)
	}
}

// The demo account is the one account a stranger gets, so what it may
// send is an allowlist and the validation is part of the guard.
func TestDemoCommandsMustNameKnownActions(t *testing.T) {
	c := Default()
	c.MySQLDSN = "user:pass@tcp(127.0.0.1:3306)/statusengine"
	c.DemoCommands = []string{"acknowledge", "make-coffee"}

	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "make-coffee") {
		t.Errorf("Validate() = %v, want it to name the unknown action", err)
	}
}

// A custom notification mails and pages the real contacts of whatever is
// being monitored. It is refused by name rather than left to judgement.
func TestDemoCommandsNeverIncludeNotify(t *testing.T) {
	c := Default()
	c.MySQLDSN = "user:pass@tcp(127.0.0.1:3306)/statusengine"
	c.DemoCommands = []string{"notify"}

	err := c.Validate()
	if err == nil || !strings.Contains(err.Error(), "notify") {
		t.Errorf("Validate() = %v, want it to refuse notify", err)
	}
}

func TestAllowsDemoCommand(t *testing.T) {
	c := Default()
	c.DemoMode = true
	c.DemoCommands = []string{"acknowledge"}

	if !c.AllowsDemoCommand("acknowledge") {
		t.Error("an allowlisted command should be allowed")
	}
	if c.AllowsDemoCommand("submit-result") {
		t.Error("a command nobody listed should not be allowed")
	}

	// Demo mode off means there is no demo account to allow anything to.
	c.DemoMode = false
	if c.AllowsDemoCommand("acknowledge") {
		t.Error("with demo mode off, nothing is allowed")
	}
}
