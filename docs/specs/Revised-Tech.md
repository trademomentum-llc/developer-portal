# Technical Specification: Custom Self-Hosted Developer Portal
## Document Control
- **Version**: 1.0
- **Date**: June 6, 2026

## 1. Technology Stack
- Core: Backstage (latest stable via @backstage/create-app, Node.js 20+)
- Runtime: k3s v1.30+ (or Docker Compose 2.x)
- Database: PostgreSQL 16 (Docker image)
- Object Storage: MinIO (S3-compatible)
- Auth: Keycloak 24+
- Git: Forgejo (Gitea fork) or GitHub free tier
- CD: Argo CD 2.12+
- Observability: Prometheus + Grafana
- IaC: Crossplane + Terraform (OSS)

## 2. Hardware/Environment Requirements
- Local server/VM: Linux (Ubuntu 24.04 LTS), 8+ vCPU, 16+ GB RAM, 100+ GB SSD.
- Network: Internal DNS, Ingress-NGINX.

## 3. Configuration
- Backstage app-config.yaml: catalog, auth, kubernetes, techdocs (MinIO).
- Helm values for production deployment.
- Environment variables for secrets (Vault injection).

## 4. Dependencies and Versions
All versions pinned for reproducibility; full list maintained in git repository.