// backend/db/db_test.go
package db

import (
	"os"
	"testing"
)

func TestInitDB_CreatesTables(t *testing.T) {
	path := t.TempDir() + "/test.db"
	defer os.Remove(path)

	database, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.Close()

	tables := []string{"users", "sessions", "clusters"}
	for _, table := range tables {
		var name string
		err := database.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestInitDB_WALMode(t *testing.T) {
	path := t.TempDir() + "/test.db"
	defer os.Remove(path)

	database, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer database.Close()

	var mode string
	database.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if mode != "wal" {
		t.Errorf("expected WAL mode, got %q", mode)
	}
}
