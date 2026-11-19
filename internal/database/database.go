// Package database provides PostgreSQL connection pooling, health checks,
// and migration management for AegisOpsAI. It wraps pgxpool for connection
// management and goose for schema migrations.
package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	// Required for goose to use pgx as the database driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPool creates a new PostgreSQL connection pool with the given DSN and
// maximum connection count. It validates connectivity before returning.
func NewPool(ctx context.Context, dsn string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 1 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// Validate connectivity.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	slog.Info("database connection pool established",
		slog.Int("max_conns", int(maxConns)),
	)

	return pool, nil
}

// HealthCheck verifies that the database is reachable and responsive.
// It returns nil if the database responds within the context deadline.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result int
	err := pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// RunMigrations executes pending database migrations from the embedded
// filesystem. It uses goose with the "pgx" driver and applies all
// pending up migrations.
func RunMigrations(ctx context.Context, dsn string, migrationsFS embed.FS, dir string) error {
	goose.SetBaseFS(migrationsFS)

	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return fmt.Errorf("opening migration database connection: %w", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			slog.Error("closing migration database connection", slog.String("error", cerr.Error()))
		}
	}()

	if err := goose.RunContext(ctx, "up", db, dir); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("database migrations applied successfully")
	return nil
}

// QueryRow is a context-aware helper that executes a query expected to return
// at most one row. It wraps pgxpool.Pool.QueryRow for convenience.
func QueryRow(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) pgx.Row {
	return pool.QueryRow(ctx, sql, args...)
}

// Exec is a context-aware helper that executes a query that doesn't return rows.
// It returns the number of rows affected and any error encountered.
func Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error) {
	tag, err := pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("executing query: %w", err)
	}
	return tag.RowsAffected(), nil
}
