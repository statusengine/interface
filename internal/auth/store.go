package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// ErrNotFound is returned when a user, role or session does not exist.
var ErrNotFound = errors.New("auth: not found")

// ErrDuplicate is returned when a username or role name is already taken.
var ErrDuplicate = errors.New("auth: already exists")

// Role is a named set of permissions.
type Role struct {
	ID          uint32   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsSystem    bool     `json:"is_system"`
}

// User is an account. PasswordHash never leaves this package's callers in
// a response body; the JSON tag is "-" so it cannot do so by accident.
type User struct {
	ID           uint32 `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	RoleID       uint32 `json:"role_id"`
	RoleName     string `json:"role"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    int64  `json:"created_at"`
	LastLoginAt  *int64 `json:"last_login_at"`
}

// Session is a logged-in browser. ID is the hash of the token handed to the
// client, not the token itself.
type Session struct {
	ID        string
	UserID    uint32
	CreatedAt int64
	ExpiresAt int64
}

// Store is the persistence for accounts, roles and sessions.
type Store struct {
	db *sql.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

const userColumns = `u.id, u.username, u.password_hash, u.display_name, u.email,
	u.role_id, r.name, u.is_active, u.created_at, u.last_login_at`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Email,
		&u.RoleID, &u.RoleName, &u.IsActive, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

// UserByUsername looks an account up for login.
func (s *Store) UserByUsername(ctx context.Context, username string) (User, error) {
	const q = `SELECT ` + userColumns + `
		FROM sei_users u JOIN sei_roles r ON r.id = u.role_id
		WHERE u.username = ?`
	return scanUser(s.db.QueryRowContext(ctx, q, username))
}

// UserByID looks an account up by primary key.
func (s *Store) UserByID(ctx context.Context, id uint32) (User, error) {
	const q = `SELECT ` + userColumns + `
		FROM sei_users u JOIN sei_roles r ON r.id = u.role_id
		WHERE u.id = ?`
	return scanUser(s.db.QueryRowContext(ctx, q, id))
}

// ListUsers returns every account, ordered by username.
func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	const q = `SELECT ` + userColumns + `
		FROM sei_users u JOIN sei_roles r ON r.id = u.role_id
		ORDER BY u.username`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("auth: listing users: %w", err)
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("auth: listing users: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// CreateUser inserts an account and returns its new ID.
func (s *Store) CreateUser(ctx context.Context, u User) (uint32, error) {
	now := time.Now().Unix()
	const q = `INSERT INTO sei_users
		(username, password_hash, display_name, email, role_id, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := s.db.ExecContext(ctx, q, u.Username, u.PasswordHash, u.DisplayName,
		u.Email, u.RoleID, u.IsActive, now, now)
	if err != nil {
		if isDuplicateKey(err) {
			return 0, fmt.Errorf("%w: user %q", ErrDuplicate, u.Username)
		}
		return 0, fmt.Errorf("auth: creating user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("auth: creating user: %w", err)
	}
	return uint32(id), nil
}

// SetPassword replaces an account's password hash.
func (s *Store) SetPassword(ctx context.Context, userID uint32, hash string) error {
	const q = `UPDATE sei_users SET password_hash = ?, updated_at = ? WHERE id = ?`
	return s.affectOne(ctx, "setting password", q, hash, time.Now().Unix(), userID)
}

// SetRole moves an account to a different role.
func (s *Store) SetRole(ctx context.Context, userID, roleID uint32) error {
	const q = `UPDATE sei_users SET role_id = ?, updated_at = ? WHERE id = ?`
	return s.affectOne(ctx, "setting role", q, roleID, time.Now().Unix(), userID)
}

// SetActive enables or disables an account. Disabling also drops its
// sessions, otherwise a disabled user keeps working until their cookie
// expires.
func (s *Store) SetActive(ctx context.Context, userID uint32, active bool) error {
	const q = `UPDATE sei_users SET is_active = ?, updated_at = ? WHERE id = ?`
	if err := s.affectOne(ctx, "setting active", q, active, time.Now().Unix(), userID); err != nil {
		return err
	}
	if !active {
		return s.DeleteUserSessions(ctx, userID)
	}
	return nil
}

// TouchLogin records a successful login.
func (s *Store) TouchLogin(ctx context.Context, userID uint32) error {
	const q = `UPDATE sei_users SET last_login_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, time.Now().Unix(), userID)
	if err != nil {
		return fmt.Errorf("auth: recording login: %w", err)
	}
	return nil
}

func (s *Store) affectOne(ctx context.Context, what, query string, args ...any) error {
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("auth: %s: %w", what, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: %s: %w", what, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// RoleByName looks a role up by its unique name.
func (s *Store) RoleByName(ctx context.Context, name string) (Role, error) {
	const q = `SELECT id, name, description, permissions, is_system FROM sei_roles WHERE name = ?`
	return scanRole(s.db.QueryRowContext(ctx, q, name))
}

// ListRoles returns every role, ordered by name.
func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	const q = `SELECT id, name, description, permissions, is_system FROM sei_roles ORDER BY name`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("auth: listing roles: %w", err)
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("auth: listing roles: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CreateRole inserts a role and returns its new ID.
func (s *Store) CreateRole(ctx context.Context, r Role) (uint32, error) {
	for _, p := range r.Permissions {
		if !IsKnownPermission(p) {
			return 0, fmt.Errorf("auth: %q is not a known permission", p)
		}
	}
	perms, err := json.Marshal(r.Permissions)
	if err != nil {
		return 0, fmt.Errorf("auth: encoding permissions: %w", err)
	}
	const q = `INSERT INTO sei_roles (name, description, permissions, is_system, created_at)
		VALUES (?, ?, ?, ?, ?)`
	res, err := s.db.ExecContext(ctx, q, r.Name, r.Description, perms, r.IsSystem, time.Now().Unix())
	if err != nil {
		if isDuplicateKey(err) {
			return 0, fmt.Errorf("%w: role %q", ErrDuplicate, r.Name)
		}
		return 0, fmt.Errorf("auth: creating role: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("auth: creating role: %w", err)
	}
	return uint32(id), nil
}

// SetRolePermissions replaces a role's permission set.
//
// Only Bootstrap calls this, to keep the built-in roles matching their
// definition in code. Nothing in the product edits a role, so this can
// never overwrite a choice somebody made in the interface.
func (s *Store) SetRolePermissions(ctx context.Context, roleID uint32, perms []string) error {
	for _, p := range perms {
		if !IsKnownPermission(p) {
			return fmt.Errorf("auth: %q is not a known permission", p)
		}
	}
	encoded, err := json.Marshal(perms)
	if err != nil {
		return fmt.Errorf("auth: encoding permissions: %w", err)
	}
	const q = `UPDATE sei_roles SET permissions = ? WHERE id = ?`
	return s.affectOne(ctx, "updating role permissions", q, encoded, roleID)
}

func scanRole(row interface{ Scan(...any) error }) (Role, error) {
	var r Role
	var perms []byte
	err := row.Scan(&r.ID, &r.Name, &r.Description, &perms, &r.IsSystem)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(perms, &r.Permissions); err != nil {
		return r, fmt.Errorf("auth: role %q has unreadable permissions: %w", r.Name, err)
	}
	return r, nil
}

// CreateSession stores a session keyed by the token's hash.
func (s *Store) CreateSession(ctx context.Context, sess Session, userAgent, ip string) error {
	const q = `INSERT INTO sei_sessions (id, user_id, created_at, expires_at, user_agent, ip)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q, sess.ID, sess.UserID, sess.CreatedAt, sess.ExpiresAt,
		truncate(userAgent, 255), truncate(ip, 45))
	if err != nil {
		return fmt.Errorf("auth: creating session: %w", err)
	}
	return nil
}

// SessionByID returns a session that has not expired yet.
func (s *Store) SessionByID(ctx context.Context, id string) (Session, error) {
	const q = `SELECT id, user_id, created_at, expires_at FROM sei_sessions
		WHERE id = ? AND expires_at > ?`
	var sess Session
	err := s.db.QueryRowContext(ctx, q, id, time.Now().Unix()).
		Scan(&sess.ID, &sess.UserID, &sess.CreatedAt, &sess.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return sess, ErrNotFound
	}
	if err != nil {
		return sess, fmt.Errorf("auth: reading session: %w", err)
	}
	return sess, nil
}

// ExtendSession pushes a session's expiry out, so an operator working
// through a shift is not logged out mid-acknowledgement.
func (s *Store) ExtendSession(ctx context.Context, id string, expiresAt int64) error {
	const q = `UPDATE sei_sessions SET expires_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, expiresAt, id)
	if err != nil {
		return fmt.Errorf("auth: extending session: %w", err)
	}
	return nil
}

// DeleteSession ends one session.
func (s *Store) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sei_sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("auth: deleting session: %w", err)
	}
	return nil
}

// DeleteUserSessions ends every session of one account.
func (s *Store) DeleteUserSessions(ctx context.Context, userID uint32) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sei_sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("auth: deleting user sessions: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes rows that can no longer authenticate
// anyone. Called periodically; the expiry check on read does the real work.
func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM sei_sessions WHERE expires_at <= ?`, time.Now().Unix())
	if err != nil {
		return 0, fmt.Errorf("auth: pruning sessions: %w", err)
	}
	return res.RowsAffected()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// isDuplicateKey recognises MySQL's duplicate-entry error, which is how a
// taken username surfaces: the UNIQUE index is the check, not a preceding
// SELECT that another request could slip past.
func isDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}
