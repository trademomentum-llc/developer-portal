# Architectural Specification: Custom Developer Portal within IDP
## Document Control
- **Version**: 1.0
- **Date**: June 6, 2026

## 1. High-Level Architecture
Adheres exactly to platform.pdf pages 7–8 blank template:
- DEVELOPER CONTROL PLANE → PORTAL = Backstage (self-hosted)
- Integrates left-to-right with IDE/CODE and COPILOTS
- Integrates right-to-left with INTEGRATION AND DELIVERY PLANE and RESOURCE PLANE
- Bottom integration with SECURITY PLANE

## 2. Deployment View
- Single k3s cluster (local) or hybrid (Backstage pod + free-tier cloud plugins).
- All components containerized; GitOps managed.

## 3. Data Flow
1. Developer → Backstage UI → Catalog query (PostgreSQL + Git)
2. Scaffolder → Template → PR to Git → Argo CD applies to cluster
3. TechDocs → Git → MinIO → Rendered in portal