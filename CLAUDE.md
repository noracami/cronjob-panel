# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CronJob Panel is a lightweight Web UI for viewing and managing Kubernetes CronJobs across multiple clusters (GKE, EKS, AKS, self-hosted). It replaces the need for kubectl or Cloud Console for CronJob operations.

## Architecture

- **Frontend**: SPA (single-page application)
- **Backend**: Thin proxy that handles authentication and permission filtering, exposing only CronJob-related K8s APIs
- **Auth**: K8s credentials are managed server-side and never exposed to the frontend

```
Browser (SPA) → Backend Proxy → Cluster A (GKE)
                               → Cluster B (EKS)
                               → Cluster C (self-hosted K8s)
```

## Scope

The project supports: listing CronJobs, viewing execution history (Jobs), viewing Pod logs, manually triggering Jobs, and suspending/resuming CronJobs. It does **not** handle creating/deleting/editing CronJob YAML, alerts, or schedule editing.

## Key Reference

- `cronjob-panel-spec.md` — full requirements specification with K8s API mappings
