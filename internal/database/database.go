// Package database opens and configures the MySQL connection pool shared by
// every repository.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Options are the pool settings that matter for this workload.
type Options struct {
	DSN          string
	MaxOpenConns int
	ConnMaxLife  time.Duration
}

// Open dials MySQL and verifies the connection before returning, so a bad
// DSN or an unreachable server is a startup failure rather than a surprise
// on the first request.
func Open(ctx context.Context, opt Options) (*sql.DB, error) {
	db, err := sql.Open("mysql", opt.DSN)
	if err != nil {
		return nil, fmt.Errorf("database: opening: %w", err)
	}

	db.SetMaxOpenConns(opt.MaxOpenConns)
	// Idle matches open so the pool doesn't tear down and redial between
	// bursts of list requests, which arrive in clusters as an operator
	// pages through a table.
	db.SetMaxIdleConns(opt.MaxOpenConns)
	db.SetConnMaxLifetime(opt.ConnMaxLife)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: connecting: %w", err)
	}
	return db, nil
}
