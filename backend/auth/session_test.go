// backend/auth/session_test.go
package auth

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/kerke/cronjob-panel/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	t.Cleanup(func() { os.Remove(path) })

	database, err := db.InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func insertTestUser(t *testing.T, database *sql.DB) string {
	t.Helper()
	id := "test-user-001"
	_, err := database.Exec(
		"INSERT INTO users (id, discord_id, username) VALUES (?, ?, ?)",
		id, "discord-001", "testuser",
	)
	if err != nil {
		t.Fatalf("insert test user failed: %v", err)
	}
	return id
}

func TestCreateSession(t *testing.T) {
	database := setupTestDB(t)
	userID := insertTestUser(t, database)

	sessionID, err := CreateSession(database, userID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sessionID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestGetSession_Valid(t *testing.T) {
	database := setupTestDB(t)
	userID := insertTestUser(t, database)

	sessionID, err := CreateSession(database, userID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	gotUserID, err := GetSessionUserID(database, sessionID)
	if err != nil {
		t.Fatalf("GetSessionUserID failed: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected user_id %q, got %q", userID, gotUserID)
	}
}

func TestGetSession_Expired(t *testing.T) {
	database := setupTestDB(t)
	userID := insertTestUser(t, database)

	_, err := CreateSession(database, userID, -time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Use a fresh session ID query to find the just-inserted expired session
	var sessionID string
	err = database.QueryRow("SELECT id FROM sessions WHERE user_id = ?", userID).Scan(&sessionID)
	if err != nil {
		t.Fatalf("could not find expired session: %v", err)
	}

	_, err = GetSessionUserID(database, sessionID)
	if err == nil {
		t.Error("expected error for expired session, got nil")
	}
}

func TestDeleteSession(t *testing.T) {
	database := setupTestDB(t)
	userID := insertTestUser(t, database)

	sessionID, err := CreateSession(database, userID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	DeleteSession(database, sessionID)

	_, err = GetSessionUserID(database, sessionID)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}
