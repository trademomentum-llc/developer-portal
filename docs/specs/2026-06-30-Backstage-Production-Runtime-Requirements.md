# Requirements Specification: Backstage Production Runtime

**Document ID:** BACKSTAGE-PROD-RUNTIME-REQ-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-Backstage-Gitea-Auth-Provider-Requirements.md, 2026-06-30-M4-Networking-Requirements.md

---

## 1. Purpose

This document defines the requirements for running Backstage in a production-like configuration against a PostgreSQL database. Today Backstage runs in local development mode with a SQLite file and the guest provider. This increment validates that the production config template (`app-config.production.yaml`) actually works by deploying a PostgreSQL instance and starting the Backstage backend with `NODE_ENV=production`.

---

## 2. Vision

A single command starts Backstage using the tracked production configuration, backed by a PostgreSQL database running in the k3d-openchoreo cluster. The guest provider is disabled, Gitea authentication is required, and the permission framework is enabled. This gives confidence that the developer portal can be deployed to a real production environment without dev-only bypasses.

---

## 3. Scope

### In Scope

- A repeatable OpenTofu module that deploys PostgreSQL into the cluster.
- A Kubernetes Service that exposes PostgreSQL to the host (NodePort).
- A helper script that initializes the Backstage database and user.
- A helper script that starts Backstage locally in production mode using `app-config.production.yaml`.
- A smoke test that verifies Backstage starts and serves requests with the production config.

### Out of Scope

- Containerizing and deploying Backstage inside Kubernetes (future milestone).
- Real TLS certificates and public DNS (the `.local` Envoy Gateway names remain sufficient).
- High availability for PostgreSQL (single replica is acceptable for the local production model).

---

## 4. Functional Requirements

### 4.1 PostgreSQL Deployment

- FR-PROD-1: PostgreSQL runs in a dedicated namespace (e.g., `backstage`).
- FR-PROD-2: A `backstage` database and `backstage` user are created automatically.
- FR-PROD-3: The password is generated and stored in a Kubernetes Secret, not committed.
- FR-PROD-4: PostgreSQL is reachable from the macOS host via a NodePort service.

### 4.2 Production Backstage Startup

- FR-PROD-5: `scripts/start-backstage-production.sh` reads credentials from the cluster Secret.
- FR-PROD-6: The script sets `NODE_ENV=production` and loads `app-config.production.yaml`.
- FR-PROD-7: Backstage migrations run automatically before the server starts.
- FR-PROD-8: The guest provider is disabled and the Gitea provider is enabled.

### 4.3 Validation

- FR-PROD-9: `scripts/smoke-backstage-production.sh` confirms the backend is reachable.
- FR-PROD-10: `scripts/smoke-all.sh` continues to pass after the production runtime is added.

---

## 5. Non-Functional Requirements

- All configuration follows the existing repo-driven, version-pinned, ASCII-only conventions.
- PostgreSQL must fit within the existing Colima resource envelope.
- No secrets are committed.

---

## 6. Success Criteria

- `scripts/install-backstage-production.sh` deploys PostgreSQL and initializes the database.
- `scripts/start-backstage-production.sh` starts Backstage with the production config.
- `scripts/smoke-backstage-production.sh` passes.
- `scripts/smoke-all.sh` still passes.

---

## 7. References

- Backstage production config: `backstage/app-config.production.yaml`
- Gitea OAuth setup: `scripts/setup-gitea-oauth.sh`
