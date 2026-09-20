package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/statusengine/interface/internal/auth"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// identityResponse is what the client learns about itself. Permissions are
// expanded, so the UI never has to know that the admin role stores "*".
type identityResponse struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	IsDemo      bool     `json:"is_demo"`
	ReadOnly    bool     `json:"read_only"`
	ExpiresAt   int64    `json:"expires_at"`
}

func (s *Server) identityResponse(ident auth.Identity) identityResponse {
	return identityResponse{
		Username:    ident.User.Username,
		DisplayName: ident.User.DisplayName,
		Email:       ident.User.Email,
		Role:        ident.Role.Name,
		Permissions: ident.Permissions.List(),
		IsDemo:      ident.IsDemo(s.cfg.DemoUser),
		ReadOnly:    !ident.Permissions.HasAnyCommand(),
		ExpiresAt:   time.Now().Add(s.auth.SessionTTL()).Unix(),
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		writeFieldError(w, http.StatusBadRequest, CodeBadRequest, "username and password are required", "username")
		return
	}

	token, ident, err := s.auth.Login(r.Context(), req.Username, req.Password,
		r.UserAgent(), clientIP(r))
	switch {
	case errors.Is(err, auth.ErrRateLimited):
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, CodeRateLimited,
			"too many login attempts; try again shortly")
		return
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, CodeUnauthorized, "invalid username or password")
		return
	case err != nil:
		loggerFrom(r.Context()).Error("login failed", "error", err)
		writeError(w, http.StatusInternalServerError, CodeInternal, "could not complete the login")
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, s.identityResponse(ident))
}

// handleLoginDemo signs in the read-only demo account without a password.
// It exists only when demo mode is on.
func (s *Server) handleLoginDemo(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.DemoMode {
		writeError(w, http.StatusNotFound, CodeNotFound, "demo mode is not enabled")
		return
	}

	token, ident, err := s.auth.LoginDemo(r.Context(), r.UserAgent(), clientIP(r))
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, CodeUnauthorized, "the demo account is not available")
		return
	case err != nil:
		loggerFrom(r.Context()).Error("demo login failed", "error", err)
		writeError(w, http.StatusInternalServerError, CodeInternal, "could not complete the login")
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, s.identityResponse(ident))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.cfg.SessionCookie); err == nil {
		if err := s.auth.Logout(r.Context(), c.Value); err != nil {
			loggerFrom(r.Context()).Error("logout failed", "error", err)
		}
	}
	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	ident, ok := identityFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, CodeUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, s.identityResponse(ident))
}

// setSessionCookie stores the token in a cookie the page's JavaScript
// cannot read. SameSite=Lax is enough because every state-changing route
// is a POST, which Lax does not send cross-site.
func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.auth.SessionTTL().Seconds()),
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
