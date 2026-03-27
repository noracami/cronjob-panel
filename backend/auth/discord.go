// backend/auth/discord.go
package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	discordAuthURL     = "https://discord.com/api/oauth2/authorize"
	discordDefaultToken = "https://discord.com/api/oauth2/token"
	discordDefaultUser  = "https://discord.com/api/users/@me"
)

// DiscordConfig holds OAuth configuration. TokenURL and UserURL can be
// overridden in tests to point at mock HTTP servers.
type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	FrontendURL  string // where to redirect after login (e.g. http://localhost:3000)
	TokenURL     string // override for testing; defaults to Discord token endpoint
	UserURL      string // override for testing; defaults to Discord user endpoint
}

func (c DiscordConfig) tokenURL() string {
	if c.TokenURL != "" {
		return c.TokenURL
	}
	return discordDefaultToken
}

func (c DiscordConfig) userURL() string {
	if c.UserURL != "" {
		return c.UserURL
	}
	return discordDefaultUser
}

// DiscordLoginHandler redirects the browser to Discord's OAuth2 authorization page.
func DiscordLoginHandler(cfg DiscordConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := url.Values{}
		params.Set("client_id", cfg.ClientID)
		params.Set("redirect_uri", cfg.RedirectURI)
		params.Set("response_type", "code")
		params.Set("scope", "identify")

		target := discordAuthURL + "?" + params.Encode()
		c.Redirect(http.StatusFound, target)
	}
}

// DiscordCallbackHandler handles the OAuth callback from Discord.
// It exchanges the code for a token, fetches the Discord user, upserts the
// user in the database, creates a session, and redirects to /.
func DiscordCallbackHandler(cfg DiscordConfig, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		accessToken, err := exchangeCode(cfg, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
			return
		}

		discordUser, err := getDiscordUser(cfg, accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get discord user"})
			return
		}

		userID, err := upsertUser(db, discordUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user"})
			return
		}

		sessionID, err := CreateSession(db, userID, 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}

		c.SetCookie("session", sessionID, int((7 * 24 * time.Hour).Seconds()), "/", "", false, true)
		redirectTo := "/"
		if cfg.FrontendURL != "" {
			redirectTo = cfg.FrontendURL
		}
		c.Redirect(http.StatusFound, redirectTo)
	}
}

type discordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func exchangeCode(cfg DiscordConfig, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", cfg.RedirectURI)

	resp, err := http.PostForm(cfg.tokenURL(), data)
	if err != nil {
		return "", fmt.Errorf("post token: %w", err)
	}
	defer resp.Body.Close()

	var tok discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}
	return tok.AccessToken, nil
}

type discordUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func getDiscordUser(cfg DiscordConfig, accessToken string) (discordUserResponse, error) {
	req, err := http.NewRequest("GET", cfg.userURL(), nil)
	if err != nil {
		return discordUserResponse{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return discordUserResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var user discordUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return discordUserResponse{}, fmt.Errorf("decode user response: %w", err)
	}
	return user, nil
}

func upsertUser(db *sql.DB, u discordUserResponse) (string, error) {
	// Check if user with this discord_id already exists.
	var existingID string
	err := db.QueryRow("SELECT id FROM users WHERE discord_id = ?", u.ID).Scan(&existingID)
	if err == nil {
		// User exists — update username and avatar_url.
		_, err = db.Exec(
			"UPDATE users SET username = ?, avatar_url = ? WHERE id = ?",
			u.Username, avatarURL(u), existingID,
		)
		return existingID, err
	}

	// New user — insert.
	newID := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO users (id, discord_id, username, avatar_url) VALUES (?, ?, ?, ?)",
		newID, u.ID, u.Username, avatarURL(u),
	)
	return newID, err
}

func avatarURL(u discordUserResponse) string {
	if u.Avatar == "" {
		return ""
	}
	ext := "png"
	if strings.HasPrefix(u.Avatar, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", u.ID, u.Avatar, ext)
}
