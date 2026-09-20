package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/statusengine/interface/internal/ratelimit"
)

// ErrInvalidCredentials covers a wrong password, an unknown user and a
// disabled account. They are one error on purpose: telling them apart
// turns the login form into a way to find out which accounts exist.
var ErrInvalidCredentials = errors.New("auth: invalid username or password")

// ErrRateLimited is returned when one client has tried too many logins.
var ErrRateLimited = errors.New("auth: too many login attempts")

// Identity is the authenticated caller, carried on the request context.
type Identity struct {
	User        User
	Role        Role
	Permissions PermissionSet
	SessionID   string
}

// Can reports whether the caller holds perm.
func (i Identity) Can(perm string) bool { return i.Permissions.Has(perm) }

// IsDemo reports whether this is the read-only demo account, which the UI
// uses to explain why actions are unavailable rather than just hiding them.
func (i Identity) IsDemo(demoUser string) bool {
	return demoUser != "" && i.User.Username == demoUser
}

// Options configure the Service.
type Options struct {
	SessionTTL      time.Duration
	LoginRateLimit  int
	LoginRateWindow time.Duration
	DemoUser        string

	// DemoCommands is the allowlist from the configuration. The
	// passwordless login checks the demo account against it: an account
	// that somehow holds more than this does not get a session that
	// way, whatever the database says.
	DemoCommands []string
}

// Service turns credentials into sessions and sessions back into
// identities.
type Service struct {
	store *Store
	log   *slog.Logger
	opt   Options

	limiter *ratelimit.Window
}

// NewService wires a Service to its store.
func NewService(store *Store, log *slog.Logger, opt Options) *Service {
	if opt.SessionTTL <= 0 {
		opt.SessionTTL = 12 * time.Hour
	}
	return &Service{
		store:   store,
		log:     log,
		opt:     opt,
		limiter: ratelimit.New(opt.LoginRateLimit, opt.LoginRateWindow),
	}
}

// DemoUser returns the configured demo account name, empty when demo mode
// is off.
func (s *Service) DemoUser() string { return s.opt.DemoUser }

// SessionTTL returns how long a new session lasts.
func (s *Service) SessionTTL() time.Duration { return s.opt.SessionTTL }

// Login verifies credentials and opens a session. It returns the raw token
// to be handed to the client; only its hash is stored.
func (s *Service) Login(ctx context.Context, username, password, userAgent, ip string) (token string, id Identity, err error) {
	username = strings.TrimSpace(username)
	if !s.limiter.Allow(ip) {
		return "", Identity{}, ErrRateLimited
	}

	user, err := s.store.UserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) {
		// Spend the same work on an unknown user as on a real one, so the
		// response time does not reveal which usernames exist.
		_, _ = VerifyPassword(password, dummyHash)
		return "", Identity{}, ErrInvalidCredentials
	}
	if err != nil {
		return "", Identity{}, err
	}

	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil {
		// A hash we cannot parse is an operational problem, not a wrong
		// password - log it, but tell the client the same thing.
		s.log.Error("cannot verify stored password hash", "username", user.Username, "error", err)
		return "", Identity{}, ErrInvalidCredentials
	}
	if !ok || !user.IsActive {
		return "", Identity{}, ErrInvalidCredentials
	}

	s.limiter.Reset(ip)

	ident, err := s.identityFor(ctx, user)
	if err != nil {
		return "", Identity{}, err
	}
	return s.openSession(ctx, user, ident, userAgent, ip)
}

// openSession issues a token and records the session. Both login paths end
// here so a session looks the same however it was obtained.
func (s *Service) openSession(ctx context.Context, user User, ident Identity, userAgent, ip string) (string, Identity, error) {
	token, hash, err := newToken()
	if err != nil {
		return "", Identity{}, err
	}
	now := time.Now()
	sess := Session{
		ID:        hash,
		UserID:    user.ID,
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(s.opt.SessionTTL).Unix(),
	}
	if err := s.store.CreateSession(ctx, sess, userAgent, ip); err != nil {
		return "", Identity{}, err
	}
	if err := s.store.TouchLogin(ctx, user.ID); err != nil {
		// The login itself succeeded; a missing timestamp is not worth
		// failing it.
		s.log.Warn("could not record last login", "username", user.Username, "error", err)
	}
	ident.SessionID = hash
	return token, ident, nil
}

// Authenticate resolves a session token into an identity, sliding the
// expiry forward when the session is past its halfway point.
func (s *Service) Authenticate(ctx context.Context, token string) (Identity, error) {
	if token == "" {
		return Identity{}, ErrNotFound
	}
	hash := hashToken(token)

	sess, err := s.store.SessionByID(ctx, hash)
	if err != nil {
		return Identity{}, err
	}
	user, err := s.store.UserByID(ctx, sess.UserID)
	if err != nil {
		return Identity{}, err
	}
	if !user.IsActive {
		// Disabling an account takes effect now, not when the cookie runs
		// out.
		_ = s.store.DeleteSession(ctx, hash)
		return Identity{}, ErrNotFound
	}

	ident, err := s.identityFor(ctx, user)
	if err != nil {
		return Identity{}, err
	}
	ident.SessionID = hash

	if halfway := sess.CreatedAt + int64(s.opt.SessionTTL.Seconds())/2; time.Now().Unix() > halfway {
		newExpiry := time.Now().Add(s.opt.SessionTTL).Unix()
		if err := s.store.ExtendSession(ctx, hash, newExpiry); err != nil {
			s.log.Warn("could not extend session", "error", err)
		}
	}
	return ident, nil
}

// Logout ends the session behind token. An unknown token is not an error:
// the caller wanted to be logged out and they are.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, hashToken(token))
}

func (s *Service) identityFor(ctx context.Context, user User) (Identity, error) {
	roles, err := s.store.ListRoles(ctx)
	if err != nil {
		return Identity{}, err
	}
	for _, r := range roles {
		if r.ID == user.RoleID {
			return Identity{
				User:        user,
				Role:        r,
				Permissions: NewPermissionSet(r.Permissions),
			}, nil
		}
	}
	return Identity{}, fmt.Errorf("auth: user %q references role %d, which does not exist", user.Username, user.RoleID)
}

// PruneSessions deletes expired rows until ctx is done. Run it in a
// goroutine at startup.
func (s *Service) PruneSessions(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := s.store.DeleteExpiredSessions(ctx)
			if err != nil {
				s.log.Warn("could not prune sessions", "error", err)
				continue
			}
			if n > 0 {
				s.log.Debug("pruned expired sessions", "count", n)
			}
		}
	}
}

// newToken returns a fresh session token and the hash to store for it.
func newToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("auth: generating session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

// hashToken is a plain SHA-256: the token is 256 bits of randomness, so
// there is nothing for a slow hash to defend against here - unlike a
// password, it cannot be guessed from a dictionary.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
