// backend/auth/discord_test.go
package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDiscordLogin_RedirectsToDiscord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := DiscordConfig{
		ClientID:    "test-client-id",
		RedirectURI: "http://localhost:8080/api/auth/discord/callback",
	}

	r := gin.New()
	r.GET("/api/auth/discord", DiscordLoginHandler(cfg))

	req := httptest.NewRequest("GET", "/api/auth/discord", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("got %d, want 302", w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected Location header")
	}
}

func TestDiscordCallback_ExchangesCodeAndCreatesSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database := setupTestDB(t)

	// Mock Discord token endpoint
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-token",
			"token_type":   "Bearer",
		})
	}))
	defer tokenServer.Close()

	// Mock Discord user endpoint
	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "123456",
			"username": "testuser",
			"avatar":   "abc123",
		})
	}))
	defer userServer.Close()

	cfg := DiscordConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:8080/api/auth/discord/callback",
		TokenURL:     tokenServer.URL,
		UserURL:      userServer.URL,
	}

	r := gin.New()
	r.GET("/api/auth/discord/callback", DiscordCallbackHandler(cfg, database))

	req := httptest.NewRequest("GET", "/api/auth/discord/callback?code=mock-code", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("got %d, want 302 redirect, body: %s", w.Code, w.Body.String())
	}

	// Check session cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie to be set")
	}
}
