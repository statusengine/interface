package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
)

// Built-in role names.
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleGuest    = "guest"

	// RoleDemo exists only when an operator has named commands the
	// public demo account may submit. It is guest plus exactly those,
	// rebuilt from the configuration on every start, so the way to
	// change what a stranger can do is the config file and nothing else.
	RoleDemo = "demo"
)

// BootstrapOptions carry what the built-in roles and the demo account
// depend on.
type BootstrapOptions struct {
	DemoMode bool
	DemoUser string
	// DemoCommands are the allowlist entries from the configuration,
	// already validated. Empty means the demo account stays read-only
	// and no demo role is created.
	DemoCommands []string
}

// DemoPermissions is what the demo role holds: everything readable,
// plus the commands named in the allowlist and nothing else.
func DemoPermissions(allowlist []string) []string {
	perms := ReadPermissions()
	for _, name := range allowlist {
		if perm, ok := DemoCommandPermission(name); ok {
			perms = append(perms, perm)
		}
	}
	return perms
}

// Bootstrap creates the three built-in roles if they are missing, and the
// demo account when demo mode is on. It is idempotent, so it can run on
// every start.
//
// It deliberately does not create an administrator with a fixed password.
// The first admin is made with `seid user create`, so a default credential
// never exists to be forgotten.
func Bootstrap(ctx context.Context, store *Store, log *slog.Logger, opt BootstrapOptions) error {
	builtins := []Role{
		{
			Name:        RoleAdmin,
			Description: "Full access, including user management",
			Permissions: []string{Wildcard},
			IsSystem:    true,
		},
		{
			Name:        RoleOperator,
			Description: "Read everything and submit external commands",
			// The command log too: the first person who wants to know
			// whether a command took is the one who sent it. Guests stay
			// out of it, because it names people and their addresses.
			Permissions: append(append(ReadPermissions(), CommandPermissions()...), PermAuditRead),
			IsSystem:    true,
		},
		{
			Name:        RoleGuest,
			Description: "Read-only; every external command is refused",
			Permissions: ReadPermissions(),
			IsSystem:    true,
		},
	}
	if len(opt.DemoCommands) > 0 {
		builtins = append(builtins, Role{
			Name:        RoleDemo,
			Description: "The public demo account: read-only plus the commands named in demo_commands",
			Permissions: DemoPermissions(opt.DemoCommands),
			IsSystem:    true,
		})
	}

	for _, r := range builtins {
		existing, err := store.RoleByName(ctx, r.Name)
		if err == nil {
			// The built-in roles are defined here, in code, so this is
			// where they are kept true. Without the reconcile a
			// permission added in a later version would only ever reach
			// new installations, and nothing in the product could grant
			// it afterwards.
			if samePermissions(existing.Permissions, r.Permissions) {
				continue
			}
			if err := store.SetRolePermissions(ctx, existing.ID, r.Permissions); err != nil {
				return err
			}
			log.Info("updated built-in role", "role", r.Name,
				"added", missing(r.Permissions, existing.Permissions),
				"removed", missing(existing.Permissions, r.Permissions))
			continue
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		if _, err := store.CreateRole(ctx, r); err != nil {
			// Another instance starting at the same moment may have won
			// the race; that is the outcome we wanted either way.
			if errors.Is(err, ErrDuplicate) {
				continue
			}
			return err
		}
		log.Info("created built-in role", "role", r.Name)
	}

	if !opt.DemoMode || opt.DemoUser == "" {
		return nil
	}
	return ensureDemoUser(ctx, store, log, opt)
}

func ensureDemoUser(ctx context.Context, store *Store, log *slog.Logger, opt BootstrapOptions) error {
	username := opt.DemoUser
	roleName := RoleGuest
	if len(opt.DemoCommands) > 0 {
		roleName = RoleDemo
	}
	role, err := store.RoleByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("auth: demo account needs the %q role: %w", roleName, err)
	}

	existing, err := store.UserByUsername(ctx, username)
	if err == nil {
		// Already there. What can change between starts is which role
		// it belongs on: adding a name to demo_commands has to reach an
		// account that already exists, and removing the last one has to
		// take the commands away again.
		if existing.RoleID == role.ID {
			return nil
		}
		if err := store.SetRole(ctx, existing.ID, role.ID); err != nil {
			return err
		}
		// Sessions carry the permissions they were opened with.
		if err := store.DeleteUserSessions(ctx, existing.ID); err != nil {
			return err
		}
		log.Info("moved the demo account to a different role",
			"username", username, "from", existing.RoleName, "to", roleName)
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// The demo account signs in through the normal form, so it needs a
	// real password. The login page posts it for the visitor when demo
	// mode is on; it is random here so the account is not reachable on a
	// deployment that later turns demo mode off without deleting it.
	password, err := randomPassword()
	if err != nil {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if _, err := store.CreateUser(ctx, User{
		Username:     username,
		PasswordHash: hash,
		DisplayName:  "Demo",
		RoleID:       role.ID,
		IsActive:     true,
	}); err != nil && !errors.Is(err, ErrDuplicate) {
		return err
	}
	log.Info("created demo account", "username", username, "role", roleName)
	return nil
}

// randomPassword produces a password nobody knows, for the demo account:
// it signs in through LoginDemo, never through the form.
func randomPassword() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generating demo password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// LoginDemo signs the demo account in without a password. It refuses any
// account that is not the configured demo user and not on the guest role,
// so this path can never become a way into a privileged account.
func (s *Service) LoginDemo(ctx context.Context, userAgent, ip string) (string, Identity, error) {
	if s.opt.DemoUser == "" {
		return "", Identity{}, ErrInvalidCredentials
	}
	// It needs no password, so nothing else costs an attacker anything
	// here - and every call writes a session row.
	if !s.limiter.Allow(ip) {
		return "", Identity{}, ErrRateLimited
	}
	user, err := s.store.UserByUsername(ctx, s.opt.DemoUser)
	if errors.Is(err, ErrNotFound) {
		return "", Identity{}, ErrInvalidCredentials
	}
	if err != nil {
		return "", Identity{}, err
	}
	if !user.IsActive {
		return "", Identity{}, ErrInvalidCredentials
	}

	ident, err := s.identityFor(ctx, user)
	if err != nil {
		return "", Identity{}, err
	}
	// The guard on the passwordless path: this account may hold nothing
	// beyond reading, plus exactly the commands the configuration names.
	// The role is rebuilt from that configuration on every start, so
	// this can only fire if the database was edited underneath us - and
	// then refusing is the right answer.
	if ident.Role.Name != RoleGuest && ident.Role.Name != RoleDemo {
		return "", Identity{}, fmt.Errorf(
			"auth: demo account %q is on role %q, not %q or %q - refusing passwordless login",
			user.Username, ident.Role.Name, RoleGuest, RoleDemo)
	}
	allowed := NewPermissionSet(DemoPermissions(s.opt.DemoCommands))
	for _, held := range ident.Permissions.List() {
		if !allowed.Has(held) {
			return "", Identity{}, fmt.Errorf(
				"auth: demo account %q holds %q, which demo_commands does not allow - refusing passwordless login",
				user.Username, held)
		}
	}

	return s.openSession(ctx, user, ident, userAgent, ip)
}

// samePermissions compares two permission sets as sets: the order a role
// happens to be stored in is not a difference worth an UPDATE.
func samePermissions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, p := range a {
		seen[p]++
	}
	for _, p := range b {
		seen[p]--
		if seen[p] < 0 {
			return false
		}
	}
	return true
}

// missing returns the entries of want that are not in have, so the log
// can say what actually changed instead of only that something did.
func missing(want, have []string) []string {
	set := make(map[string]struct{}, len(have))
	for _, p := range have {
		set[p] = struct{}{}
	}
	var out []string
	for _, p := range want {
		if _, ok := set[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}
