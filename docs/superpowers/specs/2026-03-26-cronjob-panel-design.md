# CronJob Panel - 設計文件

## 概述

輕量 Web UI，用來檢視和管理 Kubernetes 上的 CronJob，支援多叢集（GKE、EKS、AKS、自建 K8s）。

## 架構

```
瀏覽器 (Nuxt SPA) → Go API (/api/*) → client-go → K8s Cluster A
                                                  → K8s Cluster B
                                    → SQLite (叢集設定、session)
```

- **Monorepo** 結構：`frontend/`（Nuxt 3）+ `backend/`（Go + Gin）
- Go binary 內嵌前端靜態檔案（`embed.FS`），打成單一 Docker image
- CI/CD 透過 GitHub Actions 打包

## 前端

- **框架**：Nuxt 3 + TypeScript
- **元件庫**：Nuxt UI + Nuxt Icon
- **輸出**：`nuxt generate` 產生靜態檔案，嵌入 Go binary

### 頁面

| 頁面 | 用途 |
|------|------|
| 登入頁 | Discord OAuth 登入 |
| 叢集管理 | 新增 / 移除叢集連線 |
| CronJob 列表 | 列出所選叢集的所有 CronJob（名稱、namespace、排程、上次執行、狀態） |
| CronJob 詳情 | 最近執行的 Job 列表 |
| Pod Log | 檢視 Job 的 Pod log |

### 操作

- 手動觸發執行（從 CronJob 建立一次性 Job）
- 暫停 / 恢復 CronJob

## 後端

- **框架**：Go + Gin
- **K8s 互動**：client-go（官方維護）
- **資料庫**：SQLite（WAL 模式），儲存叢集認證資訊與 session

### API 路由

| 方法 | 路徑 | 用途 |
|------|------|------|
| GET | `/api/auth/discord` | Discord OAuth 登入 |
| GET | `/api/auth/discord/callback` | OAuth callback |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/clusters` | 列出所有叢集 |
| POST | `/api/clusters` | 新增叢集 |
| DELETE | `/api/clusters/:id` | 移除叢集 |
| GET | `/api/clusters/:id/cronjobs` | 列出叢集的 CronJob |
| GET | `/api/clusters/:id/namespaces/:ns/cronjobs/:name` | CronJob 詳情 |
| GET | `/api/clusters/:id/namespaces/:ns/cronjobs/:name/jobs` | CronJob 的 Job 歷史 |
| GET | `/api/clusters/:id/namespaces/:ns/pods/:pod/log` | Pod log |
| POST | `/api/clusters/:id/namespaces/:ns/cronjobs/:name/trigger` | 手動觸發 Job |
| PATCH | `/api/clusters/:id/namespaces/:ns/cronjobs/:name/suspend` | 暫停 / 恢復 |

### SQLite Schema

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    discord_id TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    avatar_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE clusters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    auth_type TEXT NOT NULL,       -- 'token' 或 'kubeconfig'
    auth_data BLOB NOT NULL,       -- 加密儲存
    created_by TEXT NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 認證

- **登入**：Discord OAuth 2.0（先做），之後擴充 Google、GitHub、Apple、無密碼 Email
- **Session**：存在 SQLite，用 HTTP-only cookie 傳遞 session ID
- **叢集認證**：K8s 憑證加密存在 SQLite，不暴露給前端

## 部署

- 單一 Docker image，Go binary 內嵌前端靜態檔案
- SQLite 檔案需掛載 persistent volume 以避免容器重啟遺失資料
- GitHub Actions 負責：build 前端 → 嵌入 Go → build Docker image

## 不做的事（初期）

- 建立 / 刪除 / 編輯 CronJob YAML 定義
- 告警通知
- 排程編輯器
- Discord 以外的 OAuth provider
