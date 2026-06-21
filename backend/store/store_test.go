package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestURL(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://keep:keep@localhost:5432/keep?sslmode=disable"
	}
	return dsn
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := getTestURL(t)
	if err := RunMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := Connect(context.Background())
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	truncateTables(t, pool)
	return pool
}

func truncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE note_labels, list_items, permissions, notes, labels CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
