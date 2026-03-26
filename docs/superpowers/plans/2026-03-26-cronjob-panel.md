# CronJob Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a web UI for viewing and managing Kubernetes CronJobs across multiple clusters, with Discord OAuth login.

**Architecture:** Go + Gin backend serving a Nuxt 3 SPA. SQLite stores cluster credentials and sessions. client-go connects to K8s clusters. Single Docker image with embedded frontend static files.

**Tech Stack:** Go, Gin, client-go, SQLite, Nuxt 3, Nuxt UI, Nuxt Icon, TypeScript, Docker, GitHub Actions

---

## File Structure

```
cronjob-panel/
├── backend/
│   ├── main.go                      # Entry point, server startup, embed frontend
│   ├── go.mod
│   ├── go.sum
│   ├── db/
│   │   ├── db.go                    # SQLite init, WAL mode, migrations
│   │   └── db_test.go
│   ├── auth/
│   │   ├── discord.go               # Discord OAuth handlers
│   │   ├── session.go               # Session middleware + CRUD
│   │   ├── discord_test.go
│   │   └── session_test.go
│   ├── cluster/
│   │   ├── store.go                 # Cluster CRUD (SQLite)
│   │   ├── client.go                # client-go factory, cached clients
│   │   ├── store_test.go
│   │   └── client_test.go
│   ├── cronjob/
│   │   ├── handler.go               # CronJob list, detail, trigger, suspend handlers
│   │   ├── handler_test.go
│   │   └── service.go               # K8s CronJob/Job/Pod operations via client-go
│   ├── router/
│   │   └── router.go                # Gin router setup, mount all routes
│   └── static/                      # Nuxt build output gets copied here (gitignored)
├── frontend/
│   ├── nuxt.config.ts
│   ├── package.json
│   ├── app.vue
│   ├── pages/
│   │   ├── login.vue                # Discord OAuth login page
│   │   ├── index.vue                # Redirect to clusters or cronjobs
│   │   ├── clusters.vue             # Cluster management
│   │   └── clusters/
│   │       └── [id]/
│   │           ├── cronjobs.vue     # CronJob list for cluster
│   │           └── [ns]/
│   │               └── [name].vue   # CronJob detail + job history + pod log
│   ├── composables/
│   │   ├── useAuth.ts               # Auth state, login/logout
│   │   └── useApi.ts                # Fetch wrapper with auth
│   ├── layouts/
│   │   └── default.vue              # App shell: sidebar with cluster switcher
│   └── components/
│       ├── CronJobTable.vue         # CronJob list table
│       ├── JobHistoryTable.vue      # Job history table
│       ├── PodLogViewer.vue         # Pod log display
│       └── ClusterForm.vue          # Add cluster form
├── Dockerfile
└── .github/
    └── workflows/
        └── build.yml
```

---

### Task 1: 初始化 Go 後端專案

**Files:**
- Create: `backend/go.mod`
- Create: `backend/main.go`

- [ ] **Step 1: 初始化 Go module**

```bash
cd backend
go mod init github.com/kerke/cronjob-panel
```

- [ ] **Step 2: 建立 main.go，啟動 Gin server**

```go
// backend/main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: 安裝 Gin 並確認編譯**

```bash
cd backend
go get github.com/gin-gonic/gin
go build -o cronjob-panel .
```

Expected: binary `cronjob-panel` 產生，無錯誤。

- [ ] **Step 4: 執行並測試 health endpoint**

```bash
./cronjob-panel &
curl http://localhost:8080/api/health
kill %1
```

Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
git add backend/go.mod backend/go.sum backend/main.go
git commit -m "feat: init Go backend with Gin health endpoint"
```

---

### Task 2: SQLite 資料庫初始化與 Migration

**Files:**
- Create: `backend/db/db.go`
- Create: `backend/db/db_test.go`

- [ ] **Step 1: 安裝 SQLite driver**

```bash
cd backend
go get github.com/mattn/go-sqlite3
```

- [ ] **Step 2: 寫 db_test.go 測試 migration**

```go
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
```

- [ ] **Step 3: 執行測試確認失敗**

```bash
cd backend
go test ./db/ -v
```

Expected: FAIL — `InitDB` 不存在。

- [ ] **Step 4: 實作 db.go**

```go
// backend/db/db.go
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		discord_id TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		avatar_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id),
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS clusters (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		endpoint TEXT NOT NULL,
		auth_type TEXT NOT NULL,
		auth_data BLOB NOT NULL,
		created_by TEXT NOT NULL REFERENCES users(id),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(schema)
	return err
}
```

- [ ] **Step 5: 執行測試確認通過**

```bash
cd backend
go test ./db/ -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/db/
git commit -m "feat: add SQLite database init with WAL mode and schema migration"
```

---

### Task 3: Session 管理

**Files:**
- Create: `backend/auth/session.go`
- Create: `backend/auth/session_test.go`

- [ ] **Step 1: 寫 session_test.go**

```go
// backend/auth/session_test.go
package auth

import (
	"os"
	"testing"
	"time"

	"github.com/kerke/cronjob-panel/db"
)

func setupTestDB(t *testing.T) *db.DB_wrapper {
	t.Helper()
	path := t.TempDir() + "/test.db"
	database, err := db.InitDB(path)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
		os.Remove(path)
	})
	return database
}

func createTestUser(t *testing.T, database *sql.DB) string {
	t.Helper()
	id := "user-test-123"
	_, err := database.Exec(
		"INSERT INTO users (id, discord_id, username) VALUES (?, ?, ?)",
		id, "discord-456", "testuser",
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return id
}

func TestCreateSession(t *testing.T) {
	database := setupTestDB(t)
	userID := createTestUser(t, database)

	sessionID, err := CreateSession(database, userID, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session ID")
	}
}

func TestGetSession_Valid(t *testing.T) {
	database := setupTestDB(t)
	userID := createTestUser(t, database)

	sessionID, _ := CreateSession(database, userID, 24*time.Hour)

	gotUserID, err := GetSessionUserID(database, sessionID)
	if err != nil {
		t.Fatalf("GetSessionUserID: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("got %q, want %q", gotUserID, userID)
	}
}

func TestGetSession_Expired(t *testing.T) {
	database := setupTestDB(t)
	userID := createTestUser(t, database)

	sessionID, _ := CreateSession(database, userID, -1*time.Hour)

	_, err := GetSessionUserID(database, sessionID)
	if err == nil {
		t.Fatal("expected error for expired session")
	}
}

func TestDeleteSession(t *testing.T) {
	database := setupTestDB(t)
	userID := createTestUser(t, database)

	sessionID, _ := CreateSession(database, userID, 24*time.Hour)
	DeleteSession(database, sessionID)

	_, err := GetSessionUserID(database, sessionID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./auth/ -v
```

Expected: FAIL — functions 不存在。

- [ ] **Step 3: 實作 session.go**

```go
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
```

- [ ] **Step 4: 安裝 uuid 套件並執行測試**

```bash
cd backend
go get github.com/google/uuid
go test ./auth/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/auth/session.go backend/auth/session_test.go
git commit -m "feat: add session create/get/delete with expiry"
```

---

### Task 4: Auth Middleware

**Files:**
- Create: `backend/auth/middleware.go`
- Create: `backend/auth/middleware_test.go`

- [ ] **Step 1: 寫 middleware_test.go**

```go
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
	userID := createTestUser(t, database)
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
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./auth/ -v -run TestAuthMiddleware
```

Expected: FAIL

- [ ] **Step 3: 實作 middleware.go**

```go
// backend/auth/middleware.go
package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		userID, err := GetSessionUserID(db, cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session invalid or expired"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
cd backend
go test ./auth/ -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add backend/auth/middleware.go backend/auth/middleware_test.go
git commit -m "feat: add auth middleware with session cookie validation"
```

---

### Task 5: Discord OAuth

**Files:**
- Create: `backend/auth/discord.go`
- Create: `backend/auth/discord_test.go`

- [ ] **Step 1: 寫 discord_test.go（用 mock HTTP server 模擬 Discord API）**

```go
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
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./auth/ -v -run TestDiscord
```

Expected: FAIL

- [ ] **Step 3: 實作 discord.go**

```go
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
	defaultDiscordAuthURL  = "https://discord.com/api/oauth2/authorize"
	defaultDiscordTokenURL = "https://discord.com/api/oauth2/token"
	defaultDiscordUserURL  = "https://discord.com/api/users/@me"
)

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURL     string // override for testing
	UserURL      string // override for testing
}

func (c DiscordConfig) tokenURL() string {
	if c.TokenURL != "" {
		return c.TokenURL
	}
	return defaultDiscordTokenURL
}

func (c DiscordConfig) userURL() string {
	if c.UserURL != "" {
		return c.UserURL
	}
	return defaultDiscordUserURL
}

type discordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type discordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func DiscordLoginHandler(cfg DiscordConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := url.Values{
			"client_id":     {cfg.ClientID},
			"redirect_uri":  {cfg.RedirectURI},
			"response_type": {"code"},
			"scope":         {"identify"},
		}
		c.Redirect(http.StatusFound, defaultDiscordAuthURL+"?"+params.Encode())
	}
}

func DiscordCallbackHandler(cfg DiscordConfig, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}

		// Exchange code for token
		token, err := exchangeCode(cfg, code)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "token exchange failed"})
			return
		}

		// Get Discord user
		user, err := getDiscordUser(cfg, token.AccessToken)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to get user"})
			return
		}

		// Upsert user in DB
		userID, err := upsertUser(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
			return
		}

		// Create session
		sessionID, err := CreateSession(db, userID, 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}

		c.SetCookie("session", sessionID, 7*24*60*60, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	}
}

func exchangeCode(cfg DiscordConfig, code string) (*discordTokenResponse, error) {
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURI},
	}

	resp, err := http.Post(cfg.tokenURL(), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func getDiscordUser(cfg DiscordConfig, accessToken string) (*discordUser, error) {
	req, _ := http.NewRequest("GET", cfg.userURL(), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user discordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func upsertUser(db *sql.DB, user *discordUser) (string, error) {
	var existingID string
	err := db.QueryRow("SELECT id FROM users WHERE discord_id = ?", user.ID).Scan(&existingID)
	if err == nil {
		db.Exec("UPDATE users SET username = ?, avatar_url = ? WHERE id = ?",
			user.Username, avatarURL(user), existingID)
		return existingID, nil
	}

	id := uuid.New().String()
	_, err = db.Exec(
		"INSERT INTO users (id, discord_id, username, avatar_url) VALUES (?, ?, ?, ?)",
		id, user.ID, user.Username, avatarURL(user),
	)
	return id, err
}

func avatarURL(user *discordUser) string {
	if user.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", user.ID, user.Avatar)
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
cd backend
go test ./auth/ -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add backend/auth/discord.go backend/auth/discord_test.go
git commit -m "feat: add Discord OAuth login and callback handlers"
```

---

### Task 6: Cluster CRUD

**Files:**
- Create: `backend/cluster/store.go`
- Create: `backend/cluster/store_test.go`

- [ ] **Step 1: 寫 store_test.go**

```go
// backend/cluster/store_test.go
package cluster

import (
	"os"
	"testing"

	"github.com/kerke/cronjob-panel/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	database, err := db.InitDB(path)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
		os.Remove(path)
	})
	// Insert a test user for FK
	database.Exec("INSERT INTO users (id, discord_id, username) VALUES ('user1', 'd1', 'test')")
	return database
}

func TestCreateAndListClusters(t *testing.T) {
	database := setupTestDB(t)

	err := CreateCluster(database, Cluster{
		Name:      "prod-gke",
		Endpoint:  "https://gke.example.com",
		AuthType:  "token",
		AuthData:  []byte("secret-token"),
		CreatedBy: "user1",
	})
	if err != nil {
		t.Fatalf("CreateCluster: %v", err)
	}

	clusters, err := ListClusters(database)
	if err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("got %d clusters, want 1", len(clusters))
	}
	if clusters[0].Name != "prod-gke" {
		t.Errorf("got name %q, want %q", clusters[0].Name, "prod-gke")
	}
}

func TestDeleteCluster(t *testing.T) {
	database := setupTestDB(t)

	CreateCluster(database, Cluster{
		Name: "to-delete", Endpoint: "https://x.com",
		AuthType: "token", AuthData: []byte("x"), CreatedBy: "user1",
	})

	clusters, _ := ListClusters(database)
	err := DeleteCluster(database, clusters[0].ID)
	if err != nil {
		t.Fatalf("DeleteCluster: %v", err)
	}

	clusters, _ = ListClusters(database)
	if len(clusters) != 0 {
		t.Fatalf("got %d clusters, want 0", len(clusters))
	}
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./cluster/ -v
```

Expected: FAIL

- [ ] **Step 3: 實作 store.go**

```go
// backend/cluster/store.go
package cluster

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint"`
	AuthType  string    `json:"auth_type"`
	AuthData  []byte    `json:"-"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateCluster(db *sql.DB, c Cluster) error {
	c.ID = uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO clusters (id, name, endpoint, auth_type, auth_data, created_by) VALUES (?, ?, ?, ?, ?, ?)",
		c.ID, c.Name, c.Endpoint, c.AuthType, c.AuthData, c.CreatedBy,
	)
	return err
}

func ListClusters(db *sql.DB) ([]Cluster, error) {
	rows, err := db.Query("SELECT id, name, endpoint, auth_type, created_by, created_at FROM clusters")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Endpoint, &c.AuthType, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

func GetCluster(db *sql.DB, id string) (*Cluster, error) {
	var c Cluster
	err := db.QueryRow(
		"SELECT id, name, endpoint, auth_type, auth_data, created_by, created_at FROM clusters WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Endpoint, &c.AuthType, &c.AuthData, &c.CreatedBy, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func DeleteCluster(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM clusters WHERE id = ?", id)
	return err
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
cd backend
go test ./cluster/ -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add backend/cluster/
git commit -m "feat: add cluster CRUD with SQLite store"
```

---

### Task 7: client-go Cluster Client Factory

**Files:**
- Create: `backend/cluster/client.go`
- Create: `backend/cluster/client_test.go`

- [ ] **Step 1: 安裝 client-go**

```bash
cd backend
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
```

- [ ] **Step 2: 寫 client_test.go**

```go
// backend/cluster/client_test.go
package cluster

import (
	"testing"
)

func TestNewK8sClient_InvalidToken(t *testing.T) {
	// Should not panic, just return a client (connection fails lazily)
	client, err := NewK8sClient("https://127.0.0.1:99999", "token", []byte("fake-token"))
	if err != nil {
		t.Fatalf("NewK8sClient: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
```

- [ ] **Step 3: 執行測試確認失敗**

```bash
cd backend
go test ./cluster/ -v -run TestNewK8sClient
```

Expected: FAIL

- [ ] **Step 4: 實作 client.go**

```go
// backend/cluster/client.go
package cluster

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClient(endpoint, authType string, authData []byte) (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	switch authType {
	case "token":
		config = &rest.Config{
			Host:            endpoint,
			BearerToken:     string(authData),
			TLSClientConfig: rest.TLSClientConfig{Insecure: true}, // TODO: support CA certs
		}
	case "kubeconfig":
		config, err = clientcmd.RESTConfigFromKubeConfig(authData)
		if err != nil {
			return nil, fmt.Errorf("parse kubeconfig: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown auth type: %s", authType)
	}

	return kubernetes.NewForConfig(config)
}
```

- [ ] **Step 5: 執行測試確認通過**

```bash
cd backend
go test ./cluster/ -v
```

Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add backend/cluster/client.go backend/cluster/client_test.go
git commit -m "feat: add K8s client factory for token and kubeconfig auth"
```

---

### Task 8: CronJob Service（K8s 操作）

**Files:**
- Create: `backend/cronjob/service.go`
- Create: `backend/cronjob/service_test.go`

- [ ] **Step 1: 寫 service_test.go（使用 client-go fake client）**

```go
// backend/cronjob/service_test.go
package cronjob

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListCronJobs(t *testing.T) {
	client := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "*/5 * * * *",
		},
	})

	svc := NewService(client)
	cronjobs, err := svc.ListCronJobs(context.Background(), "")
	if err != nil {
		t.Fatalf("ListCronJobs: %v", err)
	}
	if len(cronjobs) != 1 {
		t.Fatalf("got %d, want 1", len(cronjobs))
	}
	if cronjobs[0].Name != "my-cronjob" {
		t.Errorf("got %q, want %q", cronjobs[0].Name, "my-cronjob")
	}
}

func TestSuspendCronJob(t *testing.T) {
	client := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
	})

	svc := NewService(client)
	err := svc.SetSuspend(context.Background(), "default", "my-cronjob", true)
	if err != nil {
		t.Fatalf("SetSuspend: %v", err)
	}

	cj, _ := client.BatchV1().CronJobs("default").Get(context.Background(), "my-cronjob", metav1.GetOptions{})
	if cj.Spec.Suspend == nil || !*cj.Spec.Suspend {
		t.Error("expected cronjob to be suspended")
	}
}

func TestTriggerJob(t *testing.T) {
	client := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cronjob",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers:    []corev1.Container{{Name: "worker", Image: "busybox"}},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	})

	svc := NewService(client)
	job, err := svc.TriggerJob(context.Background(), "default", "my-cronjob")
	if err != nil {
		t.Fatalf("TriggerJob: %v", err)
	}
	if job == nil {
		t.Fatal("expected non-nil job")
	}
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./cronjob/ -v
```

Expected: FAIL

- [ ] **Step 3: 實作 service.go**

```go
// backend/cronjob/service.go
package cronjob

import (
	"context"
	"fmt"
	"io"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

type Service struct {
	client kubernetes.Interface
}

func NewService(client kubernetes.Interface) *Service {
	return &Service{client: client}
}

func (s *Service) ListCronJobs(ctx context.Context, namespace string) ([]batchv1.CronJob, error) {
	list, err := s.client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (s *Service) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	return s.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (s *Service) ListJobs(ctx context.Context, namespace, cronjobName string) ([]batchv1.Job, error) {
	list, err := s.client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Filter jobs owned by this cronjob
	var jobs []batchv1.Job
	for _, job := range list.Items {
		for _, ref := range job.OwnerReferences {
			if ref.Kind == "CronJob" && ref.Name == cronjobName {
				jobs = append(jobs, job)
				break
			}
		}
	}
	return jobs, nil
}

func (s *Service) GetPodLog(ctx context.Context, namespace, podName string) (string, error) {
	req := s.client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: ptr.To(int64(500)),
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	bytes, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Service) TriggerJob(ctx context.Context, namespace, cronjobName string) (*batchv1.Job, error) {
	cj, err := s.GetCronJob(ctx, namespace, cronjobName)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-manual-%d", cronjobName, time.Now().Unix()),
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cj.Name,
					UID:        cj.UID,
				},
			},
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}

	return s.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
}

func (s *Service) SetSuspend(ctx context.Context, namespace, name string, suspend bool) error {
	cj, err := s.GetCronJob(ctx, namespace, name)
	if err != nil {
		return err
	}
	cj.Spec.Suspend = ptr.To(suspend)
	_, err = s.client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	return err
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
cd backend
go test ./cronjob/ -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add backend/cronjob/
git commit -m "feat: add CronJob service with list, trigger, suspend, and pod log"
```

---

### Task 9: CronJob HTTP Handlers

**Files:**
- Create: `backend/cronjob/handler.go`
- Create: `backend/cronjob/handler_test.go`

- [ ] **Step 1: 寫 handler_test.go**

```go
// backend/cronjob/handler_test.go
package cronjob

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/gin-gonic/gin"
)

func TestHandleListCronJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := fake.NewSimpleClientset(&batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: "cj1", Namespace: "default"},
		Spec:       batchv1.CronJobSpec{Schedule: "*/5 * * * *"},
	})

	r := gin.New()
	h := NewHandler(func(clusterID string) (*Service, error) {
		return NewService(client), nil
	})
	r.GET("/api/clusters/:id/cronjobs", h.ListCronJobs)

	req := httptest.NewRequest("GET", "/api/clusters/c1/cronjobs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}

	var result []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 1 {
		t.Fatalf("got %d items, want 1", len(result))
	}
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
cd backend
go test ./cronjob/ -v -run TestHandle
```

Expected: FAIL

- [ ] **Step 3: 實作 handler.go**

```go
// backend/cronjob/handler.go
package cronjob

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceFactory func(clusterID string) (*Service, error)

type Handler struct {
	getService ServiceFactory
}

func NewHandler(factory ServiceFactory) *Handler {
	return &Handler{getService: factory}
}

func (h *Handler) ListCronJobs(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cronjobs, err := svc.ListCronJobs(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cronjobs)
}

func (h *Handler) GetCronJob(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cj, err := svc.GetCronJob(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cj)
}

func (h *Handler) ListJobs(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobs, err := svc.ListJobs(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) GetPodLog(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log, err := svc.GetPodLog(c.Request.Context(), c.Param("ns"), c.Param("pod"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": log})
}

func (h *Handler) TriggerJob(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := svc.TriggerJob(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *Handler) SetSuspend(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body struct {
		Suspend bool `json:"suspend"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if err := svc.SetSuspend(c.Request.Context(), c.Param("ns"), c.Param("name"), body.Suspend); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
cd backend
go test ./cronjob/ -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add backend/cronjob/handler.go backend/cronjob/handler_test.go
git commit -m "feat: add CronJob HTTP handlers"
```

---

### Task 10: Router 組裝與 Cluster API Handlers

**Files:**
- Create: `backend/router/router.go`
- Modify: `backend/main.go`

- [ ] **Step 1: 實作 router.go**

```go
// backend/router/router.go
package router

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/cluster"
	"github.com/kerke/cronjob-panel/cronjob"
)

func Setup(r *gin.Engine, db *sql.DB, discordCfg auth.DiscordConfig) {
	// Auth routes (no middleware)
	r.GET("/api/auth/discord", auth.DiscordLoginHandler(discordCfg))
	r.GET("/api/auth/discord/callback", auth.DiscordCallbackHandler(discordCfg, db))

	api := r.Group("/api")
	api.Use(auth.RequireAuth(db))

	// Logout
	api.POST("/auth/logout", func(c *gin.Context) {
		cookie, _ := c.Cookie("session")
		auth.DeleteSession(db, cookie)
		c.SetCookie("session", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Cluster CRUD
	api.GET("/clusters", func(c *gin.Context) {
		clusters, err := cluster.ListClusters(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, clusters)
	})

	api.POST("/clusters", func(c *gin.Context) {
		var body struct {
			Name     string `json:"name"`
			Endpoint string `json:"endpoint"`
			AuthType string `json:"auth_type"`
			AuthData string `json:"auth_data"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		userID := c.GetString("user_id")
		err := cluster.CreateCluster(db, cluster.Cluster{
			Name:      body.Name,
			Endpoint:  body.Endpoint,
			AuthType:  body.AuthType,
			AuthData:  []byte(body.AuthData),
			CreatedBy: userID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	api.DELETE("/clusters/:id", func(c *gin.Context) {
		if err := cluster.DeleteCluster(db, c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// CronJob routes
	serviceFactory := func(clusterID string) (*cronjob.Service, error) {
		cl, err := cluster.GetCluster(db, clusterID)
		if err != nil {
			return nil, err
		}
		k8sClient, err := cluster.NewK8sClient(cl.Endpoint, cl.AuthType, cl.AuthData)
		if err != nil {
			return nil, err
		}
		return cronjob.NewService(k8sClient), nil
	}

	h := cronjob.NewHandler(serviceFactory)
	api.GET("/clusters/:id/cronjobs", h.ListCronJobs)
	api.GET("/clusters/:id/namespaces/:ns/cronjobs/:name", h.GetCronJob)
	api.GET("/clusters/:id/namespaces/:ns/cronjobs/:name/jobs", h.ListJobs)
	api.GET("/clusters/:id/namespaces/:ns/pods/:pod/log", h.GetPodLog)
	api.POST("/clusters/:id/namespaces/:ns/cronjobs/:name/trigger", h.TriggerJob)
	api.PATCH("/clusters/:id/namespaces/:ns/cronjobs/:name/suspend", h.SetSuspend)
}
```

- [ ] **Step 2: 更新 main.go**

```go
// backend/main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/db"
	"github.com/kerke/cronjob-panel/router"
)

func main() {
	database, err := db.InitDB("data/cronjob-panel.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	discordCfg := auth.DiscordConfig{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
	}

	r := gin.Default()
	router.Setup(r, database, discordCfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: 確認編譯通過**

```bash
cd backend
go build ./...
```

Expected: 無錯誤

- [ ] **Step 4: Commit**

```bash
git add backend/router/router.go backend/main.go
git commit -m "feat: wire up router with all API routes"
```

---

### Task 11: 初始化 Nuxt 3 前端專案

**Files:**
- Create: `frontend/` (Nuxt 3 project)

- [ ] **Step 1: 建立 Nuxt 3 專案**

```bash
cd /path/to/cronjob-panel
npx nuxi@latest init frontend
```

- [ ] **Step 2: 安裝 Nuxt UI 和 Nuxt Icon**

```bash
cd frontend
npx nuxi module add ui
npx nuxi module add icon
npm install
```

- [ ] **Step 3: 設定 nuxt.config.ts 的 API proxy**

```ts
// frontend/nuxt.config.ts
export default defineNuxtConfig({
  modules: ['@nuxt/ui', '@nuxt/icon'],
  devtools: { enabled: true },
  devServer: {
    port: 3000,
  },
  nitro: {
    devProxy: {
      '/api': {
        target: 'http://localhost:8080/api',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 4: 確認 dev server 啟動**

```bash
cd frontend
npm run dev
```

Expected: Nuxt dev server 在 localhost:3000 啟動，無錯誤。

- [ ] **Step 5: Commit**

```bash
git add frontend/
git commit -m "feat: init Nuxt 3 frontend with Nuxt UI and Icon"
```

---

### Task 12: 前端 Auth Composable 與登入頁

**Files:**
- Create: `frontend/composables/useAuth.ts`
- Create: `frontend/composables/useApi.ts`
- Create: `frontend/pages/login.vue`
- Modify: `frontend/app.vue`

- [ ] **Step 1: 建立 useApi.ts**

```ts
// frontend/composables/useApi.ts
export function useApi() {
  async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`/api${path}`, {
      credentials: 'include',
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    })

    if (res.status === 401) {
      navigateTo('/login')
      throw new Error('Unauthorized')
    }

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body.error || `API error: ${res.status}`)
    }

    return res.json()
  }

  return { apiFetch }
}
```

- [ ] **Step 2: 建立 useAuth.ts**

```ts
// frontend/composables/useAuth.ts
export function useAuth() {
  const { apiFetch } = useApi()

  function login() {
    window.location.href = '/api/auth/discord'
  }

  async function logout() {
    await apiFetch('/auth/logout', { method: 'POST' })
    navigateTo('/login')
  }

  return { login, logout }
}
```

- [ ] **Step 3: 建立登入頁**

```vue
<!-- frontend/pages/login.vue -->
<template>
  <div class="flex items-center justify-center min-h-screen">
    <UCard class="w-80">
      <template #header>
        <h2 class="text-lg font-semibold text-center">CronJob Panel</h2>
      </template>
      <UButton
        icon="i-simple-icons-discord"
        label="使用 Discord 登入"
        color="primary"
        block
        size="lg"
        @click="login"
      />
    </UCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })
const { login } = useAuth()
</script>
```

- [ ] **Step 4: 更新 app.vue**

```vue
<!-- frontend/app.vue -->
<template>
  <NuxtPage />
</template>
```

- [ ] **Step 5: 確認登入頁正常顯示**

```bash
cd frontend
npm run dev
```

瀏覽 http://localhost:3000/login，應看到 Discord 登入按鈕。

- [ ] **Step 6: Commit**

```bash
git add frontend/composables/ frontend/pages/login.vue frontend/app.vue
git commit -m "feat: add auth composables and Discord login page"
```

---

### Task 13: 前端 Layout 與叢集管理頁

**Files:**
- Create: `frontend/layouts/default.vue`
- Create: `frontend/pages/clusters.vue`
- Create: `frontend/components/ClusterForm.vue`
- Create: `frontend/pages/index.vue`

- [ ] **Step 1: 建立 default layout**

```vue
<!-- frontend/layouts/default.vue -->
<template>
  <div class="min-h-screen flex">
    <aside class="w-64 border-r p-4 flex flex-col">
      <h1 class="text-lg font-bold mb-4">CronJob Panel</h1>
      <nav class="flex flex-col gap-1">
        <UButton
          to="/clusters"
          variant="ghost"
          icon="i-heroicons-server-stack"
          label="叢集管理"
        />
      </nav>
      <div class="mt-auto">
        <UButton
          variant="ghost"
          icon="i-heroicons-arrow-right-on-rectangle"
          label="登出"
          @click="logout"
        />
      </div>
    </aside>
    <main class="flex-1 p-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const { logout } = useAuth()
</script>
```

- [ ] **Step 2: 建立 index.vue（重導到 clusters）**

```vue
<!-- frontend/pages/index.vue -->
<script setup lang="ts">
navigateTo('/clusters')
</script>
```

- [ ] **Step 3: 建立 ClusterForm.vue**

```vue
<!-- frontend/components/ClusterForm.vue -->
<template>
  <UModal v-model:open="open">
    <template #header>
      <h3 class="text-lg font-semibold">新增叢集</h3>
    </template>
    <UForm :state="form" @submit="submit" class="p-4 space-y-4">
      <UFormField label="名稱" required>
        <UInput v-model="form.name" placeholder="prod-gke" />
      </UFormField>
      <UFormField label="Endpoint" required>
        <UInput v-model="form.endpoint" placeholder="https://k8s.example.com" />
      </UFormField>
      <UFormField label="認證方式" required>
        <USelect v-model="form.auth_type" :items="['token', 'kubeconfig']" />
      </UFormField>
      <UFormField label="認證資料" required>
        <UTextarea v-model="form.auth_data" rows="4" placeholder="Bearer token 或 kubeconfig 內容" />
      </UFormField>
      <UButton type="submit" label="新增" :loading="loading" />
    </UForm>
  </UModal>
</template>

<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ created: [] }>()
const { apiFetch } = useApi()
const loading = ref(false)

const form = reactive({
  name: '',
  endpoint: '',
  auth_type: 'token',
  auth_data: '',
})

async function submit() {
  loading.value = true
  try {
    await apiFetch('/clusters', {
      method: 'POST',
      body: JSON.stringify(form),
    })
    emit('created')
    open.value = false
    Object.assign(form, { name: '', endpoint: '', auth_type: 'token', auth_data: '' })
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 4: 建立叢集管理頁**

```vue
<!-- frontend/pages/clusters.vue -->
<template>
  <div>
    <div class="flex justify-between items-center mb-4">
      <h2 class="text-xl font-semibold">叢集管理</h2>
      <UButton icon="i-heroicons-plus" label="新增叢集" @click="showForm = true" />
    </div>

    <UTable :columns="columns" :rows="clusters">
      <template #actions-data="{ row }">
        <div class="flex gap-2">
          <UButton
            :to="`/clusters/${row.id}/cronjobs`"
            size="xs"
            variant="soft"
            label="CronJobs"
          />
          <UButton
            size="xs"
            color="red"
            variant="soft"
            icon="i-heroicons-trash"
            @click="remove(row.id)"
          />
        </div>
      </template>
    </UTable>

    <ClusterForm v-model:open="showForm" @created="refresh" />
  </div>
</template>

<script setup lang="ts">
const { apiFetch } = useApi()
const showForm = ref(false)
const clusters = ref<any[]>([])

const columns = [
  { key: 'name', label: '名稱' },
  { key: 'endpoint', label: 'Endpoint' },
  { key: 'auth_type', label: '認證方式' },
  { key: 'actions', label: '操作' },
]

async function refresh() {
  clusters.value = await apiFetch('/clusters')
}

async function remove(id: string) {
  await apiFetch(`/clusters/${id}`, { method: 'DELETE' })
  refresh()
}

onMounted(refresh)
</script>
```

- [ ] **Step 5: 確認頁面正常顯示**

```bash
cd frontend
npm run dev
```

瀏覽 http://localhost:3000/clusters，應看到叢集列表（空）和新增按鈕。

- [ ] **Step 6: Commit**

```bash
git add frontend/layouts/ frontend/pages/ frontend/components/ClusterForm.vue
git commit -m "feat: add layout, cluster management page with CRUD"
```

---

### Task 14: CronJob 列表頁

**Files:**
- Create: `frontend/components/CronJobTable.vue`
- Create: `frontend/pages/clusters/[id]/cronjobs.vue`

- [ ] **Step 1: 建立 CronJobTable.vue**

```vue
<!-- frontend/components/CronJobTable.vue -->
<template>
  <UTable :columns="columns" :rows="cronjobs">
    <template #status-data="{ row }">
      <UBadge :color="row.spec?.suspend ? 'amber' : 'emerald'">
        {{ row.spec?.suspend ? '已暫停' : '執行中' }}
      </UBadge>
    </template>
    <template #actions-data="{ row }">
      <div class="flex gap-2">
        <UButton
          :to="`/clusters/${clusterId}/${row.metadata.namespace}/${row.metadata.name}`"
          size="xs"
          variant="soft"
          label="詳情"
        />
        <UButton
          size="xs"
          variant="soft"
          :icon="row.spec?.suspend ? 'i-heroicons-play' : 'i-heroicons-pause'"
          @click="$emit('toggleSuspend', row)"
        />
        <UButton
          size="xs"
          variant="soft"
          icon="i-heroicons-play-circle"
          label="觸發"
          @click="$emit('trigger', row)"
        />
      </div>
    </template>
  </UTable>
</template>

<script setup lang="ts">
defineProps<{
  cronjobs: any[]
  clusterId: string
}>()

defineEmits<{
  toggleSuspend: [row: any]
  trigger: [row: any]
}>()

const columns = [
  { key: 'metadata.name', label: '名稱' },
  { key: 'metadata.namespace', label: 'Namespace' },
  { key: 'spec.schedule', label: '排程' },
  { key: 'status.lastScheduleTime', label: '上次執行' },
  { key: 'status', label: '狀態' },
  { key: 'actions', label: '操作' },
]
</script>
```

- [ ] **Step 2: 建立 CronJob 列表頁**

```vue
<!-- frontend/pages/clusters/[id]/cronjobs.vue -->
<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <UButton to="/clusters" variant="ghost" icon="i-heroicons-arrow-left" />
      <h2 class="text-xl font-semibold">CronJobs</h2>
    </div>

    <CronJobTable
      :cronjobs="cronjobs"
      :cluster-id="clusterId"
      @toggle-suspend="toggleSuspend"
      @trigger="trigger"
    />
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { apiFetch } = useApi()

const clusterId = route.params.id as string
const cronjobs = ref<any[]>([])

async function refresh() {
  cronjobs.value = await apiFetch(`/clusters/${clusterId}/cronjobs`)
}

async function toggleSuspend(row: any) {
  const ns = row.metadata.namespace
  const name = row.metadata.name
  await apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/suspend`, {
    method: 'PATCH',
    body: JSON.stringify({ suspend: !row.spec?.suspend }),
  })
  refresh()
}

async function trigger(row: any) {
  const ns = row.metadata.namespace
  const name = row.metadata.name
  await apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/trigger`, {
    method: 'POST',
  })
  refresh()
}

onMounted(refresh)
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/CronJobTable.vue frontend/pages/clusters/
git commit -m "feat: add CronJob list page with suspend and trigger actions"
```

---

### Task 15: CronJob 詳情頁（Job 歷史 + Pod Log）

**Files:**
- Create: `frontend/components/JobHistoryTable.vue`
- Create: `frontend/components/PodLogViewer.vue`
- Create: `frontend/pages/clusters/[id]/[ns]/[name].vue`

- [ ] **Step 1: 建立 JobHistoryTable.vue**

```vue
<!-- frontend/components/JobHistoryTable.vue -->
<template>
  <UTable :columns="columns" :rows="jobs">
    <template #status-data="{ row }">
      <UBadge :color="jobStatusColor(row)">
        {{ jobStatusText(row) }}
      </UBadge>
    </template>
    <template #actions-data="{ row }">
      <UButton
        size="xs"
        variant="soft"
        label="查看 Log"
        @click="$emit('viewLog', row)"
      />
    </template>
  </UTable>
</template>

<script setup lang="ts">
defineProps<{ jobs: any[] }>()
defineEmits<{ viewLog: [row: any] }>()

const columns = [
  { key: 'metadata.name', label: 'Job 名稱' },
  { key: 'status.startTime', label: '開始時間' },
  { key: 'status.completionTime', label: '完成時間' },
  { key: 'status', label: '狀態' },
  { key: 'actions', label: '操作' },
]

function jobStatusText(job: any): string {
  if (job.status?.succeeded) return '成功'
  if (job.status?.failed) return '失敗'
  if (job.status?.active) return '執行中'
  return '未知'
}

function jobStatusColor(job: any): string {
  if (job.status?.succeeded) return 'emerald'
  if (job.status?.failed) return 'red'
  if (job.status?.active) return 'sky'
  return 'gray'
}
</script>
```

- [ ] **Step 2: 建立 PodLogViewer.vue**

```vue
<!-- frontend/components/PodLogViewer.vue -->
<template>
  <UModal v-model:open="open">
    <template #header>
      <h3 class="text-lg font-semibold">Pod Log</h3>
    </template>
    <div class="p-4">
      <pre class="bg-gray-900 text-green-400 p-4 rounded text-sm overflow-auto max-h-96 font-mono">{{ log || '載入中...' }}</pre>
    </div>
  </UModal>
</template>

<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })
defineProps<{ log: string }>()
</script>
```

- [ ] **Step 3: 建立 CronJob 詳情頁**

```vue
<!-- frontend/pages/clusters/[id]/[ns]/[name].vue -->
<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <UButton
        :to="`/clusters/${clusterId}/cronjobs`"
        variant="ghost"
        icon="i-heroicons-arrow-left"
      />
      <h2 class="text-xl font-semibold">{{ ns }} / {{ name }}</h2>
    </div>

    <UCard class="mb-4" v-if="cronjob">
      <div class="grid grid-cols-2 gap-2 text-sm">
        <div><strong>排程：</strong>{{ cronjob.spec?.schedule }}</div>
        <div>
          <strong>狀態：</strong>
          <UBadge :color="cronjob.spec?.suspend ? 'amber' : 'emerald'">
            {{ cronjob.spec?.suspend ? '已暫停' : '執行中' }}
          </UBadge>
        </div>
        <div><strong>上次執行：</strong>{{ cronjob.status?.lastScheduleTime || '無' }}</div>
      </div>
    </UCard>

    <h3 class="text-lg font-semibold mb-2">執行歷史</h3>
    <JobHistoryTable :jobs="jobs" @view-log="viewLog" />

    <PodLogViewer v-model:open="showLog" :log="podLog" />
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { apiFetch } = useApi()

const clusterId = route.params.id as string
const ns = route.params.ns as string
const name = route.params.name as string

const cronjob = ref<any>(null)
const jobs = ref<any[]>([])
const showLog = ref(false)
const podLog = ref('')

async function refresh() {
  const [cj, jobList] = await Promise.all([
    apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}`),
    apiFetch(`/clusters/${clusterId}/namespaces/${ns}/cronjobs/${name}/jobs`),
  ])
  cronjob.value = cj
  jobs.value = jobList || []
}

async function viewLog(job: any) {
  showLog.value = true
  podLog.value = ''
  // Get first pod name from job (simplified: assume pod name = job name + suffix)
  // In production, you'd list pods by job label selector
  const podName = job.metadata.name
  try {
    const result = await apiFetch<{ log: string }>(
      `/clusters/${clusterId}/namespaces/${ns}/pods/${podName}/log`
    )
    podLog.value = result.log
  } catch {
    podLog.value = '無法取得 log'
  }
}

onMounted(refresh)
</script>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/components/JobHistoryTable.vue frontend/components/PodLogViewer.vue frontend/pages/clusters/
git commit -m "feat: add CronJob detail page with job history and pod log viewer"
```

---

### Task 16: Go 內嵌前端靜態檔案

**Files:**
- Modify: `backend/main.go`
- Create: `backend/static/.gitkeep`

- [ ] **Step 1: 建立 static 目錄和 .gitkeep**

```bash
mkdir -p backend/static
touch backend/static/.gitkeep
```

- [ ] **Step 2: 更新 main.go 加入 embed**

```go
// backend/main.go
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/db"
	"github.com/kerke/cronjob-panel/router"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	database, err := db.InitDB("data/cronjob-panel.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	discordCfg := auth.DiscordConfig{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
	}

	r := gin.Default()
	router.Setup(r, database, discordCfg)

	// Serve frontend static files
	frontendFS, _ := fs.Sub(staticFS, "static")
	r.NoRoute(gin.WrapH(http.FileServer(http.FS(frontendFS))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: 確認編譯通過**

```bash
cd backend
go build ./...
```

Expected: 無錯誤

- [ ] **Step 4: Commit**

```bash
git add backend/main.go backend/static/.gitkeep
git commit -m "feat: embed frontend static files in Go binary"
```

---

### Task 17: Dockerfile

**Files:**
- Create: `Dockerfile`

- [ ] **Step 1: 建立 Dockerfile**

```dockerfile
# Dockerfile
FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npx nuxt generate

FROM golang:1.22-alpine AS backend-build
RUN apk add --no-cache gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/frontend/.output/public ./static/
RUN CGO_ENABLED=1 go build -o cronjob-panel .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-build /app/backend/cronjob-panel .
RUN mkdir -p data
EXPOSE 8080
CMD ["./cronjob-panel"]
```

- [ ] **Step 2: 確認 Dockerfile 語法正確**

```bash
docker build --check .
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile
git commit -m "feat: add multi-stage Dockerfile for single-image build"
```

---

### Task 18: GitHub Actions 工作流

**Files:**
- Create: `.github/workflows/build.yml`

- [ ] **Step 1: 建立 build.yml**

```yaml
# .github/workflows/build.yml
name: Build and Push

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: cronjob-panel:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

- [ ] **Step 2: Commit**

```bash
mkdir -p .github/workflows
git add .github/workflows/build.yml
git commit -m "feat: add GitHub Actions build workflow"
```

---

## Summary

| Task | 主題 | 依賴 |
|------|------|------|
| 1 | Go 後端初始化 | — |
| 2 | SQLite 初始化 | 1 |
| 3 | Session 管理 | 2 |
| 4 | Auth Middleware | 3 |
| 5 | Discord OAuth | 3, 4 |
| 6 | Cluster CRUD | 2 |
| 7 | client-go Client Factory | 6 |
| 8 | CronJob Service | 7 |
| 9 | CronJob HTTP Handlers | 8 |
| 10 | Router 組裝 | 5, 6, 9 |
| 11 | Nuxt 3 前端初始化 | — |
| 12 | Auth Composable + 登入頁 | 11 |
| 13 | Layout + 叢集管理頁 | 12 |
| 14 | CronJob 列表頁 | 13 |
| 15 | CronJob 詳情頁 | 14 |
| 16 | Go 內嵌靜態檔案 | 10 |
| 17 | Dockerfile | 16 |
| 18 | GitHub Actions | 17 |
