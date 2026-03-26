// backend/auth/session.go
package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrSessionNotFound = errors.New("session not found or expired")

func CreateSession(db *sql.DB, userID string, duration time.Duration) (string, error) {
	id := uuid.New().String()
	expiresAt := time.Now().Add(duration)
	_, err := db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		id, userID, expiresAt,
	)
	return id, err
}

func GetSessionUserID(db *sql.DB, sessionID string) (string, error) {
	var userID string
	err := db.QueryRow(
		"SELECT user_id FROM sessions WHERE id = ? AND expires_at > ?",
		sessionID, time.Now(),
	).Scan(&userID)
	if err != nil {
		return "", ErrSessionNotFound
	}
	return userID, nil
}

func DeleteSession(db *sql.DB, sessionID string) {
	db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
}
