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
)

// Bootstrap creates the three built-in roles if they are missing, and the
// demo account when demo mode is on. It is idempotent, so it can run on
// every start.
//
// It deliberately does not create an administrator with a fixed password.
// The first admin is made with `seid user create`, so a default credential
// never exists to be forgotten.
func Bootstrap(ctx context.Context, store *Store, log *slog.Logger, demoMode bool, demoUser string) error {
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
			Permissions: append(ReadPermissions(), CommandPermissions()...),
			IsSystem:    true,
		},
		{
			Name:        RoleGuest,
			Description: "Read-only; every external command is refused",
			Permissions: ReadPermissions(),
			IsSystem:    true,
		},
	}

	for _, r := range builtins {
		_, err := store.RoleByName(ctx, r.Name)
		if err == nil {
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

	if !demoMode || demoUser == "" {
		return nil
	}
	return ensureDemoUser(ctx, store, log, demoUser)
}

func ensureDemoUser(ctx context.Context, store *Store, log *slog.Logger, username string) error {
	if _, err := store.UserByUsername(ctx, username); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	guest, err := store.RoleByName(ctx, RoleGuest)
	if err != nil {
		return fmt.Errorf("auth: demo account needs the %q role: %w", RoleGuest, err)
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
		RoleID:       guest.ID,
		IsActive:     true,
	}); err != nil && !errors.Is(err, ErrDuplicate) {
		return err
	}
	log.Info("created demo account", "username", username, "role", RoleGuest)
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
	if ident.Role.Name != RoleGuest {
		return "", Identity{}, fmt.Errorf("auth: demo account %q is on role %q, not %q - refusing passwordless login",
			user.Username, ident.Role.Name, RoleGuest)
	}
	if ident.Permissions.HasAnyCommand() {
		return "", Identity{}, fmt.Errorf("auth: demo account %q would be granted commands - refusing passwordless login", user.Username)
	}

	return s.openSession(ctx, user, ident, userAgent, ip)
}
