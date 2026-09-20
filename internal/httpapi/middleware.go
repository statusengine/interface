package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/statusengine/interface/internal/auth"
)

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyIdentity
	ctxKeyLogger
)

// middleware is the shape every wrapper below has.
type middleware func(http.Handler) http.Handler

// chain applies wrappers so the first listed is the outermost, which is
// how they read on the page.
func chain(h http.Handler, mw ...middleware) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// statusRecorder captures the status code so the access log can report it.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Flush forwards to the wrapped writer so Server-Sent Events still stream
// through the log wrapper.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err == nil {
				id = hex.EncodeToString(b)
			}
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func loggingMiddleware(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			reqLog := log.With("request_id", requestID(r.Context()))
			r = r.WithContext(context.WithValue(r.Context(), ctxKeyLogger, reqLog))

			next.ServeHTTP(rec, r)

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			level := slog.LevelInfo
			switch {
			case rec.status >= 500:
				level = slog.LevelError
			case rec.status >= 400:
				level = slog.LevelWarn
			}
			reqLog.Log(r.Context(), level, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote", clientIP(r),
			)
		})
	}
}

func recoverMiddleware(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					// The client went away mid-response; that is not a bug.
					panic(rec)
				}
				log.Error("panic serving request",
					"request_id", requestID(r.Context()),
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// securityHeadersMiddleware sets the headers that cost nothing and close
// off whole classes of problem. The CSP is strict because this app ships
// its own bundle and loads nothing from anywhere else.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; "+
				"script-src 'self'; connect-src 'self'; font-src 'self' data:; "+
				"frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

// authMiddleware resolves the session cookie into an identity. It rejects
// anonymous requests outright: every route it wraps needs a caller.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ident, ok := s.identityFromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, CodeUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyIdentity, ident)))
	})
}

func (s *Server) identityFromRequest(r *http.Request) (auth.Identity, bool) {
	c, err := r.Cookie(s.cfg.SessionCookie)
	if err != nil || c.Value == "" {
		return auth.Identity{}, false
	}
	ident, err := s.auth.Authenticate(r.Context(), c.Value)
	if err != nil {
		if !errors.Is(err, auth.ErrNotFound) {
			loggerFrom(r.Context()).Error("resolving session", "error", err)
		}
		return auth.Identity{}, false
	}
	return ident, true
}

// requirePermission guards a route. It runs after authMiddleware, so an
// identity is always present.
func requirePermission(perm string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ident, ok := identityFrom(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, CodeUnauthorized, "authentication required")
				return
			}
			if !ident.Can(perm) {
				writeError(w, http.StatusForbidden, CodeForbidden,
					"your role does not grant "+perm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func identityFrom(ctx context.Context) (auth.Identity, bool) {
	ident, ok := ctx.Value(ctxKeyIdentity).(auth.Identity)
	return ident, ok
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyRequestID).(string)
	return id
}

func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// clientIP returns the address to attribute a request to. X-Forwarded-For
// is honoured only when the direct peer is a loopback or private address,
// because a header from the open internet is whatever the client typed.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil || !(ip.IsLoopback() || ip.IsPrivate()) {
		return host
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		if first = strings.TrimSpace(first); first != "" {
			return first
		}
	}
	return host
}
