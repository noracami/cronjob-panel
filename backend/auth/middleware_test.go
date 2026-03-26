// backend/auth/middleware_test.go
package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_NoSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)

	r := gin.New()
	r.Use(RequireAuth(database))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_ValidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)
	userID := insertTestUser(t, database)
	sessionID, _ := CreateSession(database, userID, 24*time.Hour)

	r := gin.New()
	r.Use(RequireAuth(database))
	r.GET("/test", func(c *gin.Context) {
		uid := c.GetString("user_id")
		c.JSON(200, gin.H{"user_id": uid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}
