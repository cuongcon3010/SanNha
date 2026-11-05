package testutil

import (
	"crypto/rand"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"sannha/be/internal/platform/database"
)

// Postgres gives each test a fresh schema, never truncating application tables.
func Postgres(t *testing.T) (*sql.DB, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("integration test requires TEST_DATABASE_URL pointing to a dedicated PostgreSQL test database")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL")
	}
	admin, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + strings.ToLower(rand.Text())
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec("CREATE SCHEMA " + identifier); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	var db *sql.DB
	t.Cleanup(func() {
		if db != nil {
			db.Close()
		}
		if _, err := admin.Exec("DROP SCHEMA " + identifier + " CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
		}
		admin.Close()
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	scopedURL := u.String()
	db, err = database.Open(scopedURL)
	if err != nil {
		t.Fatal(err)
	}
	return db, scopedURL
}
