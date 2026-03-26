# CronJob Panel - 需求規格

## 目標

一個輕量 Web UI，用來檢視和管理 Kubernetes 上的 CronJob，取代需要 kubectl 或 Cloud Console 的操作。支援多叢集（GKE、EKS、AKS、自建 K8s）。

## 架構

```
瀏覽器 (SPA) → 後端 Proxy → 叢集 A (GKE)
                            → 叢集 B (EKS)
                            → 叢集 C (自建 K8s)
```

- **前端**：單頁應用
- **後端**：薄 proxy，負責認證與權限過濾，只暴露 CronJob 相關 K8s API
- **認證**：K8s 憑證由後端管理，不暴露給前端

## 功能

### 叢集管理

- 新增 / 移除叢集連線
- 每個叢集存一組認證資訊（kubeconfig 或 token）
- 前端可切換叢集

### 檢視

- 列出所有 CronJob（名稱、namespace、排程、上次執行時間、狀態）
- 查看單一 CronJob 的最近執行歷史（Job 列表）
- 查看 Job 的 Pod log

### 操作

- 手動觸發執行（從 CronJob 建立一次性 Job）
- 暫停 / 恢復 CronJob

### 存取控制

- 後端管認證（不將 K8s 憑證暴露給前端）
- 簡單的登入機制或限制內網存取

## 對應 K8s API

| 功能 | K8s API |
|------|---------|
| 列出 CronJob | `GET /apis/batch/v1/cronjobs` |
| CronJob 詳情 | `GET /apis/batch/v1/namespaces/{ns}/cronjobs/{name}` |
| 最近執行的 Job | `GET /apis/batch/v1/jobs?labelSelector=...` |
| Job 的 Pod log | `GET /api/v1/namespaces/{ns}/pods/{pod}/log` |
| 手動觸發 | `POST /apis/batch/v1/namespaces/{ns}/jobs` |
| 暫停 / 恢復 | `PATCH /apis/batch/v1/namespaces/{ns}/cronjobs/{name}` |

## 不做的事（初期）

- 建立 / 刪除 / 編輯 CronJob YAML 定義（用 kubectl apply 管）
- 告警通知
- 排程編輯器
