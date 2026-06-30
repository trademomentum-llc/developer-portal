# Design Specification: Custom Self-Hosted Developer Portal
## Document Control
- **Version**: 1.0
- **Date**: June 6, 2026
- **Author**: Grok Platform Engineering Team

## 1. Overview
High-level design for the Backstage-based developer portal that maps directly to the DEVELOPER CONTROL PLANE PORTAL in platform.pdf pages 3–8.

## 2. User Experience Design
- UI: Backstage default theme customized with organizational branding (logo, colors).
- Key Screens: Home dashboard, Catalog (search/filter), Entity pages (overview, relations, docs), Scaffolder template picker, TechDocs viewer.
- Navigation: Sidebar with Catalog, Docs, Create, APIs, Infrastructure.

## 3. Data Model
- Entities: Component, API, Resource, System, Group, User (Backstage catalog model).
- Storage: catalog-info.yaml files in Git + PostgreSQL for runtime state.
- TechDocs: Markdown files rendered from Git via MinIO backend.

## 4. Component Diagram (Text Representation)
Developer UI (React) <-> Backstage Backend (Node.js)

 Catalog Processor
* Scaffolder Engine
* TechDocs Renderer
* Plugin Registry

 Kubernetes Plugin (local k3s)
* Git Plugin (Forgejo/GitHub free)
* ArgoCD Plugin
* Prometheus/Grafana Plugin
* Keycloak Auth Plugin

## 5. Integration Design
- All integrations use official Backstage plugins to maintain loose coupling as per the reference architecture.

## 6. Security Design
- Auth flow: OAuth2 → Keycloak → Backstage RBAC.
- Data in transit: TLS everywhere.