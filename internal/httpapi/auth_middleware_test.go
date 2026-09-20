package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/config"
)

// stubAuthenticator lets a test decide exactly how authentication fails.
type stubAuthenticator struct {
	identity auth.Identity
	err      error
}

func (s stubAuthenticator) Login(context.Context, string, string, string, string) (string, auth.Identity, error) {
	return "", s.identity, s.err
}

func (s stubAuthenticator) LoginDemo(context.Context, string, string) (string, auth.Identity, error) {
	return "", s.identity, s.err
}

func (s stubAuthenticator) Authenticate(context.Context, string) (auth.Identity, error) {
	return s.identity, s.err
}

func (s stubAuthenticator) Logout(context.Context, string) error { return s.err }

func (s stubAuthenticator) SessionTTL() time.Duration { return time.Hour }

func serverWith(t *testing.T, a Authenticator) *Server {
	t.Helper()
	cfg := config.Default()
	cfg.MySQLDSN = "unused"
	return &Server{cfg: cfg, log: discardLogger(t), auth: a}
}

func requestWithCookie(cfg config.Config) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
	r.AddCookie(&http.Cookie{Name: cfg.SessionCookie, Value: "a-token"})
	return r
}

func decodeError(t *testing.T, body []byte) apiError {
	t.Helper()
	var parsed errorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("response is not the standard error shape: %v", err)
	}
	return parsed.Error
}

// The regression this exists for: a database outage used to answer 401
// on every request, which logged everyone out of a working installation
// and sent them to a login page that could not work either.
func TestAuthMiddlewareDistinguishesAbsentFromUncheckable(t *testing.T) {
	handlerRan := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerRan = true
		w.WriteHeader(http.StatusOK)
	})

	cases := map[string]struct {
		authErr    error
		withCookie bool
		wantStatus int
		wantCode   string
	}{
		"no cookie at all": {
			authErr: nil, withCookie: false,
			wantStatus: http.StatusUnauthorized, wantCode: CodeUnauthorized,
		},
		"a token the store does not know": {
			authErr: auth.ErrNotFound, withCookie: true,
			wantStatus: http.StatusUnauthorized, wantCode: CodeUnauthorized,
		},
		"the store could not be reached": {
			authErr: errors.New("dial tcp 127.0.0.1:3306: connect: connection refused"), withCookie: true,
			wantStatus: http.StatusServiceUnavailable, wantCode: CodeUnavailable,
		},
		"the query hit its deadline": {
			authErr: context.DeadlineExceeded, withCookie: true,
			wantStatus: http.StatusServiceUnavailable, wantCode: CodeUnavailable,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			handlerRan = false
			s := serverWith(t, stubAuthenticator{err: tc.authErr})

			r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
			if tc.withCookie {
				r = requestWithCookie(s.cfg)
			}

			w := httptest.NewRecorder()
			s.authMiddleware(next).ServeHTTP(w, r)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if got := decodeError(t, w.Body.Bytes()).Code; got != tc.wantCode {
				t.Errorf("code = %q, want %q", got, tc.wantCode)
			}
			if handlerRan {
				t.Error("the guarded handler ran anyway")
			}
		})
	}
}

func TestAuthMiddlewarePassesAValidSessionThrough(t *testing.T) {
	want := auth.Identity{
		User:        auth.User{ID: 7, Username: "ops"},
		Role:        auth.Role{Name: "operator"},
		Permissions: auth.NewPermissionSet(auth.ReadPermissions()),
	}
	s := serverWith(t, stubAuthenticator{identity: want})

	var got auth.Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = identityFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	s.authMiddleware(next).ServeHTTP(w, requestWithCookie(s.cfg))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got.User.Username != "ops" {
		t.Errorf("the handler saw %q", got.User.Username)
	}
	if !got.Can(auth.PermHostsRead) {
		t.Error("the identity lost its permissions on the way through")
	}
}

// A timeout is not an internal error: the query was valid and the
// database was reachable, it just could not answer in time, and the
// useful reply says what to do differently.
func TestInternalErrorSeparatesDeadlineFromFailure(t *testing.T) {
	s := serverWith(t, stubAuthenticator{})
	s.cfg.QueryTimeout = 20 * time.Second

	t.Run("deadline", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/history/checks", nil)
		s.internalError(w, r, "the check history", context.DeadlineExceeded)

		if w.Code != http.StatusGatewayTimeout {
			t.Errorf("status = %d, want 504", w.Code)
		}
		body := decodeError(t, w.Body.Bytes())
		if body.Code != CodeTimeout {
			t.Errorf("code = %q, want %q", body.Code, CodeTimeout)
		}
		if !contains(body.Message, "Narrow the time window") {
			t.Errorf("the message should say what to do differently: %q", body.Message)
		}
	})

	// The browser navigated away. Nobody is listening and nothing is
	// wrong.
	t.Run("cancelled", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
		s.internalError(w, r, "hosts", context.Canceled)

		if w.Body.Len() != 0 {
			t.Errorf("wrote a body for a cancelled request: %q", w.Body.String())
		}
	})

	t.Run("a real failure stays a 500 and says nothing about the schema", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts", nil)
		s.internalError(w, r, "hosts", errors.New("Error 1054: Unknown column 'secret' in 'field list'"))

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", w.Code)
		}
		if contains(w.Body.String(), "Unknown column") {
			t.Error("the SQL error reached the client")
		}
	})
}
