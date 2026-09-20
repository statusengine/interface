package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/config"
)

func limitedServer(t *testing.T, perClient, demo int) *Server {
	t.Helper()
	cfg := config.Default()
	cfg.DemoMode = true
	cfg.DemoUser = "guest"
	return &Server{cfg: cfg, log: discardLogger(t), commandLimits: newCommandLimiter(perClient, demo)}
}

func send(t *testing.T, s *Server, username, ip string) int {
	t.Helper()
	handler := s.commandRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/commands/reschedule", nil)
	r.RemoteAddr = ip + ":40000"
	ident := auth.Identity{User: auth.User{Username: username}}
	r = r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity, ident))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w.Code
}

func TestCommandsAreLimitedPerCaller(t *testing.T) {
	s := limitedServer(t, 3, 10)

	for i := 1; i <= 3; i++ {
		if code := send(t, s, "ops", "10.0.0.1"); code != http.StatusAccepted {
			t.Fatalf("command %d got %d, want 202", i, code)
		}
	}
	if code := send(t, s, "ops", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Errorf("the fourth command got %d, want 429", code)
	}
	// One caller's budget is their own. Two operators are two callers,
	// and so are two addresses.
	if code := send(t, s, "other", "10.0.0.1"); code != http.StatusAccepted {
		t.Error("a different account should have its own budget")
	}
	if code := send(t, s, "ops", "10.0.0.2"); code != http.StatusAccepted {
		t.Error("a different address should have its own budget")
	}
}

// The demo account is shared, so it gets the smaller per-visitor budget.
func TestTheDemoAccountIsHeldToItsOwnBudget(t *testing.T) {
	s := limitedServer(t, 60, 2)

	if send(t, s, "guest", "10.0.0.1") != http.StatusAccepted ||
		send(t, s, "guest", "10.0.0.1") != http.StatusAccepted {
		t.Fatal("the first two demo commands should be accepted")
	}
	if code := send(t, s, "guest", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Errorf("the third demo command from one visitor got %d, want 429", code)
	}
	// A second visitor is not blocked by the first.
	if code := send(t, s, "guest", "10.0.0.2"); code != http.StatusAccepted {
		t.Errorf("a second visitor got %d, want 202", code)
	}
}

// A crowd, or one person with many addresses, must not sum up to the
// flood the per-visitor limit exists to prevent.
func TestTheWholeDemoCrowdHasACeiling(t *testing.T) {
	s := limitedServer(t, 60, 1)

	total := 0
	for i := 0; i < demoTotalFactor+3; i++ {
		ip := "10.0.0." + string(rune('1'+i))
		if send(t, s, "guest", ip) == http.StatusAccepted {
			total++
		}
	}
	if total != demoTotalFactor {
		t.Errorf("the demo accepted %d commands across visitors, want the ceiling of %d",
			total, demoTotalFactor)
	}
}

func TestARefusalSaysWhenToComeBack(t *testing.T) {
	s := limitedServer(t, 1, 1)
	send(t, s, "ops", "10.0.0.1")

	handler := s.commandRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	r := httptest.NewRequest(http.MethodPost, "/api/v1/commands/reschedule", nil)
	r.RemoteAddr = "10.0.0.1:40000"
	r = r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity,
		auth.Identity{User: auth.User{Username: "ops"}}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("a 429 without Retry-After leaves the client guessing")
	}
}

func TestNoLimiterWhenEveryLimitIsOff(t *testing.T) {
	if l := newCommandLimiter(-1, -1); l != nil {
		t.Error("switching both limits off should leave no limiter at all")
	}
	s := limitedServer(t, -1, -1)
	for i := 0; i < 20; i++ {
		if code := send(t, s, "ops", "10.0.0.1"); code != http.StatusAccepted {
			t.Fatalf("command %d got %d with limiting off", i, code)
		}
	}
}
