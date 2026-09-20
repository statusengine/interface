package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/statusengine/interface/internal/auth"
)

// A forwarded-for header is only believable when the peer we are actually
// talking to is a proxy we could plausibly have put there.
func TestClientIPTrustsForwardedOnlyFromPrivatePeers(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{"no header", "203.0.113.9:1234", "", "203.0.113.9"},
		{"public peer forges header", "203.0.113.9:1234", "10.0.0.1", "203.0.113.9"},
		{"loopback proxy is trusted", "127.0.0.1:1234", "198.51.100.7", "198.51.100.7"},
		{"private proxy is trusted", "10.1.2.3:1234", "198.51.100.7", "198.51.100.7"},
		{"first hop wins", "127.0.0.1:1234", "198.51.100.7, 10.0.0.1", "198.51.100.7"},
		{"empty header ignored", "127.0.0.1:1234", "", "127.0.0.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tc.remoteAddr
			if tc.forwarded != "" {
				r.Header.Set("X-Forwarded-For", tc.forwarded)
			}
			if got := clientIP(r); got != tc.want {
				t.Errorf("clientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

func withIdentity(r *http.Request, perms []string) *http.Request {
	ident := auth.Identity{
		User:        auth.User{Username: "tester"},
		Permissions: auth.NewPermissionSet(perms),
	}
	return r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity, ident))
}

func TestRequirePermission(t *testing.T) {
	reached := false
	h := requirePermission(auth.PermCmdAcknowledge)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reached = true
			w.WriteHeader(http.StatusOK)
		}))

	t.Run("granted", func(t *testing.T) {
		reached = false
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withIdentity(httptest.NewRequest(http.MethodPost, "/", nil),
			[]string{auth.PermCmdAcknowledge}))
		if w.Code != http.StatusOK || !reached {
			t.Errorf("status %d, handler reached %t; want 200 and true", w.Code, reached)
		}
	})

	t.Run("read-only role is refused", func(t *testing.T) {
		reached = false
		w := httptest.NewRecorder()
		h.ServeHTTP(w, withIdentity(httptest.NewRequest(http.MethodPost, "/", nil),
			auth.ReadPermissions()))
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
		if reached {
			t.Error("the handler ran despite the missing permission")
		}
		var body errorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("response is not the standard error shape: %v", err)
		}
		if body.Error.Code != CodeForbidden {
			t.Errorf("code = %q, want %q", body.Error.Code, CodeForbidden)
		}
	})

	t.Run("anonymous is refused", func(t *testing.T) {
		reached = false
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
		if reached {
			t.Error("the handler ran without an identity")
		}
	})
}

func TestRecoverMiddlewareTurnsPanicIntoJSON(t *testing.T) {
	log := discardLogger(t)
	h := recoverMiddleware(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("something went sideways")
	}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not the standard error shape: %v", err)
	}
	// The panic message must not reach the client.
	if body.Error.Message != "internal server error" {
		t.Errorf("message = %q, want a generic one", body.Error.Message)
	}
}
