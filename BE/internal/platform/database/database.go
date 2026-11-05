package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

func Open(dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("DATABASE_URL must not be empty")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		// Driver parse errors can contain the DSN, including its password.
		return nil, fmt.Errorf("invalid DATABASE_URL")
	}
	cfg.ConnectTimeout = 5 * time.Second
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot connect to PostgreSQL; check DATABASE_URL, TLS and server availability")
	}
	return db, nil
}

func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}

func Migrate(db *sql.DB, name, schema string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize migrations across processes before creating the tracking table.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(739177312)`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name=$1`, name).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err = tx.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(name) VALUES ($1)`, name); err != nil {
			return err
		}
	}
	return tx.Commit()
}
