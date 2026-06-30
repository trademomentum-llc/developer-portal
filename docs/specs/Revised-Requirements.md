# Requirements Document: Custom Self-Hosted Developer Portal (IDP PORTAL Component)
## Document Control
- **Version**: 1.0
- **Date**: June 6, 2026
- **Author**: Grok Platform Engineering Team (Coordinator: Grok)
- **Status**: Approved for Design Phase
- **Based on**: platform.pdf (Humanitec IDP Reference Architecture Template, pages 3–13)

## 1. Purpose
Define functional, non-functional, and technical requirements for a custom developer portal that fulfills the DEVELOPER CONTROL PLANE – PORTAL role in the IDP reference architecture. The portal must be built using open-source software, deployed on local infrastructure (or hybrid cloud free tier), and integrate with the other planes as shown in the AWS/GCP/Azure/Multicloud examples and blank templates (pages 7–8).

## 2. Scope
- In Scope: Backstage core portal with service catalog, self-service scaffolding, TechDocs, plugin-based integrations to Git, CI/CD, Kubernetes resources, observability, and security planes.
- Out of Scope: Commercial SaaS portals (e.g., Port, Cortex paid tiers), full multi-region geo-redundancy unless later specified, paid cloud services beyond free tier.

## 3. Stakeholders
- Developers (end users)
- Platform Engineering Team (operators)
- Security & Compliance Team
- Jason (Sponsor, New York City)

## 4. Functional Requirements
FR-01: Provide a centralized Software Catalog using Backstage entity model (catalog-info.yaml files stored in Git).
FR-02: Enable self-service project scaffolding via Backstage Scaffolder with predefined templates for new services, libraries, and infrastructure.
FR-03: Integrate TechDocs for in-portal Markdown documentation sourced directly from Git repositories.
FR-04: Support discoverability and search across catalog entities, docs, and plugins.
FR-05: Provide plugin-based integrations with Version Control (Gitea/Forgejo or GitHub free), CI/CD (Argo CD/Flux/GitHub Actions), Registry (Harbor), Resource Plane (local k3s Kubernetes), Observability (Prometheus/Grafana), and Security (Keycloak/OPA/Vault).
FR-06: Support user authentication, role-based access, and audit logging.
FR-07: Allow extension via custom plugins while maintaining the exact plane structure from platform.pdf.

## 5. Non-Functional Requirements
NFR-01: Open-source only – Backstage (CNCF) as core; all supporting components OSS.
NFR-02: Deployment: Local-first on k3s/KIND/Minikube or Docker Compose; hybrid option using Oracle Cloud Always Free / AWS/GCP/Azure free tier for supplemental resources only.
NFR-03: Scalability: Support 10–100 concurrent developers with sub-2-second page loads; horizontal scaling via Kubernetes.
NFR-04: Security: OAuth2/Keycloak auth, secrets management (Vault), policy-as-code (OPA), container scanning (Trivy).
NFR-05: Availability: 99.5% uptime for production build; automated backups via Velero.
NFR-06: Cost: Zero licensing; only free-tier or local hardware.
NFR-07: Maintainability: GitOps-managed deployment (Argo CD); IaC with Terraform/Crossplane.

## 6. Constraints and Assumptions
- Local infrastructure has sufficient resources (minimum 8 vCPU / 16 GB RAM / 100 GB SSD).
- Git repositories and Kubernetes clusters are available or will be provisioned per Resource/Integration planes.
- Alignment to platform.pdf blank templates required.

## 7. Acceptance Criteria
- Successful local deployment of Backstage with catalog populated from sample entities.
- End-to-end test of scaffolding a new service that appears in the catalog.
- Integration verification with at least Git, Kubernetes, and Observability tools.