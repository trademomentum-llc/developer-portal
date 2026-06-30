# Engineering Specification: Custom Developer Portal Implementation
## Document Control
- **Version**: 1.0
- **Date**: June 6, 2026

## 1. Development Process
- GitOps workflow (Argo CD for the portal itself).
- IaC: Helm + Terraform/Crossplane for all infrastructure.
- CI/CD: GitHub Actions free tier or self-hosted Jenkins for portal customizations.
- Testing: Unit (Jest), integration (Playwright), end-to-end.

## 2. Deployment Pipeline
1. Local dev → Docker Compose
2. Staging → k3s cluster
3. Production → HA k3s with backups

## 3. Monitoring and Operations
- Self-monitoring via Prometheus/Grafana dashboards.
- Logging: Loki (OSS).
- Alerting: Alertmanager.