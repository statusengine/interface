// Package migrate applies the interface's own schema. The migrations are
// embedded in the binary so a deployment is still one file, and they only
// ever create sei_* objects - the worker owns statusengine_*.
package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"
)

//go:embed sql/*.sql
var files embed.FS

// lockName is held for the duration of a migration run so two daemons
// starting at once do not both try to create the same table. Without it the
// loser fails on a duplicate-key error at a random point, leaving the schema
// half-applied and the operator reading a confusing message.
const lockName = "sei_migrate"

// Run applies every migration that has not been recorded yet, in filename
// order, each in its own transaction.
func Run(ctx context.Context, db *sql.DB, log *slog.Logger) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquiring connection: %w", err)
	}
	defer conn.Close()

	var got sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", lockName).Scan(&got); err != nil {
		return fmt.Errorf("migrate: acquiring lock: %w", err)
	}
	if !got.Valid || got.Int64 != 1 {
		return fmt.Errorf("migrate: another process is holding the %q lock", lockName)
	}
	defer func() {
		// Best effort: the lock is released anyway when the connection
		// closes, so a failure here is not worth aborting a good startup.
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), "SELECT RELEASE_LOCK(?)", lockName)
	}()

	if err := ensureTable(ctx, conn); err != nil {
		return err
	}
	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		body, err := files.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("migrate: reading %s: %w", name, err)
		}
		if err := apply(ctx, conn, version, string(body)); err != nil {
			return err
		}
		log.Info("applied migration", "version", version)
	}
	return nil
}

func ensureTable(ctx context.Context, conn *sql.Conn) error {
	const ddl = `CREATE TABLE IF NOT EXISTS sei_schema_migrations (
		version    VARCHAR(190) NOT NULL,
		applied_at BIGINT       NOT NULL,
		PRIMARY KEY (version)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	if _, err := conn.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("migrate: creating sei_schema_migrations: %w", err)
	}
	return nil
}

func appliedVersions(ctx context.Context, conn *sql.Conn) (map[string]bool, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version FROM sei_schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("migrate: reading applied versions: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("migrate: reading applied versions: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func migrationNames() ([]string, error) {
	entries, err := fs.ReadDir(files, "sql")
	if err != nil {
		return nil, fmt.Errorf("migrate: listing migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func apply(ctx context.Context, conn *sql.Conn, version, body string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate: %s: begin: %w", version, err)
	}
	defer tx.Rollback()

	for i, stmt := range Statements(body) {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %s: statement %d: %w", version, i+1, err)
		}
	}

	const insert = "INSERT INTO sei_schema_migrations (version, applied_at) VALUES (?, ?)"
	if _, err := tx.ExecContext(ctx, insert, version, time.Now().Unix()); err != nil {
		return fmt.Errorf("migrate: %s: recording: %w", version, err)
	}
	return tx.Commit()
}

// Statements splits a migration file into individual statements. It strips
// -- line comments and splits on semicolons, which is enough because these
// files are ours: plain DDL and seed inserts, no stored programs and no
// semicolons inside literals. It is exported so a test can assert that.
func Statements(body string) []string {
	var out []string
	var cur strings.Builder

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			s := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(cur.String()), ";"))
			if s != "" {
				out = append(out, s)
			}
			cur.Reset()
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		out = append(out, s)
	}
	return out
}
