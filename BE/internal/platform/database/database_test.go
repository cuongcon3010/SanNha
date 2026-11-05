package database_test

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"sannha/be/internal/platform/database"
	"sannha/be/internal/testutil"
)

func TestUniqueViolationChecksSQLStateAndConstraint(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505", ConstraintName: "booking_active_slot"})
	if !database.IsUniqueViolation(err, "booking_active_slot") {
		t.Fatal("unique violation not recognized")
	}
	if database.IsUniqueViolation(err, "identity_users_email_key") || database.IsUniqueViolation(nil, "booking_active_slot") || database.IsUniqueViolation(&pgconn.PgError{Code: "23503", ConstraintName: "booking_active_slot"}, "booking_active_slot") {
		t.Fatal("unrelated error treated as a slot conflict")
	}
}

func TestOpenRejectsEmptyAndRedactsMalformedDSN(t *testing.T) {
	for _, dsn := range []string{"", " ", "postgres://user:TEST_PLACEHOLDER@%invalid"} {
		db, err := database.Open(dsn)
		if db != nil || err == nil || strings.Contains(err.Error(), "TEST_PLACEHOLDER") {
			t.Fatal("invalid DSN accepted or credentials leaked")
		}
	}
}

func TestPostgresMigrationsAreConcurrentAndTransactional(t *testing.T) {
	db, dsn := testutil.Postgres(t)
	other, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- database.Migrate(other, "concurrent_pg_v1", `CREATE TABLE migration_probe (id TEXT PRIMARY KEY)`)
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name=$1`, "concurrent_pg_v1").Scan(&count); err != nil || count != 1 {
		t.Fatalf("migration count=%d error=%v", count, err)
	}
	if err = database.Migrate(db, "broken_pg_v1", `CREATE TABLE rollback_probe (id TEXT); SELECT nonexistent_column`); err == nil {
		t.Fatal("invalid migration accepted")
	}
	var relation sql.NullString
	if err = db.QueryRow(`SELECT to_regclass('rollback_probe')::text`).Scan(&relation); err != nil || relation.Valid {
		t.Fatal("failed migration left a table behind", err)
	}
	if err = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name=$1`, "broken_pg_v1").Scan(&count); err != nil || count != 0 {
		t.Fatal("failed migration was recorded", err)
	}
}
