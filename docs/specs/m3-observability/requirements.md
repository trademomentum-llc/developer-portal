# M3 Observability -- Requirements Document

> **Milestone:** M3 -- Observability
> **Version:** 0.1
> **Date:** 2026-05-23
> **Status:** Kickoff draft
> **Companion docs:** [design-specification.md](./design-specification.md), [technical-specification.md](./technical-specification.md)

---

## 1. Purpose

M3 adds an observability surface to the local IDP without changing the M2 delivery contract. A pushed application should still flow through Gitea Actions, Score, platform-config, Flux, and OpenChoreo. M3 makes that path inspectable through OpenTelemetry, SigNoz, and post-deploy cost visibility.

The first M3 increment is a spec and preflight increment. No cluster state should be mutated until the SigNoz footprint, storage class behavior, and existing OpenChoreo observability-plane resources are inventoried.

## 2. Context

M1 installed the substrate and explicitly deferred observability instrumentation. M2 proved the delivery loop end to end and used Infracost only as a pre-deploy PR gate. M3 now owns:

- OpenTelemetry collection for platform and demo workload signals.
- SigNoz as the local UI and backend for traces, metrics, and logs.
- Post-deploy visibility for the latest accepted Infracost outputs.
- Backstage instrumentation and possible Backstage containerization if host-only dev mode blocks stable telemetry.

Official upstream constraints matter for this milestone. SigNoz self-hosted Kubernetes docs list non-trivial minimum local resources, and SigNoz K8s Infra is the recommended default Kubernetes collection agent. OpenTelemetry's collector Helm chart supports deployment, daemonset, and statefulset modes, with Kubernetes presets that have RBAC and duplication trade-offs.

References:
- SigNoz local Kubernetes install: https://signoz.io/docs/install/kubernetes/local/
- SigNoz Kubernetes collection agents: https://signoz.io/docs/opentelemetry-collection-agents/k8s/get-started/
- OpenTelemetry Collector Helm chart: https://opentelemetry.io/docs/platforms/kubernetes/helm/collector/

## 3. Stakeholders

| Stakeholder | Role | Interest in M3 |
|---|---|---|
| User (platform owner) | Sole decision maker | A demonstrable observability plane that explains delivery, runtime health, and cost posture |
| Developers | IDP consumers | Component pages that link to traces, metrics, logs, and cost evidence |
| AI coding agents | First-class consumers | Machine-readable health and cost signals for later automated triage |
| Finance / budget owner | Future consumer | Post-deploy cost estimates tied to deployed components |
| Security / operations | Future consumer | Signals that can feed M6 runtime policies and incident workflows |

## 4. Functional Requirements

### 4.1 Preflight and inventory

- **FR-1** M3 SHALL inventory current cluster resources before installing anything: namespaces, storage classes, existing OpenChoreo observability-plane services, ports, and resource headroom.
- **FR-2** M3 SHALL verify that the local k3d/Colima environment can satisfy SigNoz's minimum resource and storage requirements, or explicitly document the gap before implementation.
- **FR-3** M3 SHALL not mutate cluster state from ad-hoc commands. Install and teardown must flow through repo scripts and/or OpenTofu modules, consistent with M2.

### 4.2 SigNoz

- **FR-4** SigNoz SHALL be the primary local observability UI for M3 unless preflight proves it conflicts with the existing OpenChoreo observability plane.
- **FR-5** SigNoz SHALL be installed into a dedicated namespace and exposed locally through a deterministic port-forward script.
- **FR-6** SigNoz data retention SHALL be bounded for local development. Default retention must be documented in the technical spec or values file.
- **FR-7** SigNoz installation SHALL use pinned chart and image versions selected during the M3 version-inventory task, not floating `latest` values.

### 4.3 OpenTelemetry collection

- **FR-8** M3 SHALL collect application traces over OTLP from at least `hello-m2`.
- **FR-9** M3 SHALL collect cluster or workload metrics sufficient to show pod availability, restart count, CPU, and memory signals for `hello-m2`.
- **FR-10** M3 SHOULD collect Kubernetes events and logs if the footprint is acceptable on the local cluster. If omitted, the omission must be documented.
- **FR-11** M3 SHALL add Kubernetes metadata such as namespace, pod, workload, and service name to telemetry before export to SigNoz.
- **FR-12** The OpenTelemetry collector topology SHALL avoid duplicate metrics/logs on a single-node k3d cluster.

### 4.4 Backstage observability surfacing

- **FR-13** Backstage SHALL expose links from catalog entities to the relevant SigNoz views, using catalog annotations or links before custom plugin code.
- **FR-14** Backstage backend instrumentation SHALL emit OTLP traces when M3 telemetry is enabled.
- **FR-15** Backstage telemetry SHALL be disabled or no-op by default when the M3 stack is not running.
- **FR-16** If host `yarn dev` cannot provide stable telemetry endpoints, M3 SHALL containerize Backstage as the minimum viable runtime required for tracing.

### 4.5 Infracost post-deploy visibility

- **FR-17** M3 SHALL persist the latest accepted Infracost JSON output for each deployed component in a deterministic, non-secret location.
- **FR-18** M3 SHALL expose a post-deploy cost dashboard or catalog link showing the latest monthly estimate and delta for `hello-m2`.
- **FR-19** M3 SHALL clearly label Infracost values as estimates, not live spend. Runtime cost allocation is deferred to M4 OpenCost.

### 4.6 Smoke checks and operations

- **FR-20** M3 SHALL provide `scripts/smoke-m3.sh` as the top-level smoke check.
- **FR-21** Per-tool smokes SHALL include SigNoz health, OTLP ingest, collector health, Backstage trace emission, `hello-m2` telemetry presence, and Infracost artifact visibility.
- **FR-22** M3 SHALL provide `scripts/install-m3.sh` and `scripts/teardown-m3.sh` if cluster resources are added.
- **FR-23** M3 teardown SHALL leave M1 and M2 resources healthy.

## 5. Non-Functional Requirements

- **NFR-1** All committed files SHALL remain ASCII-only.
- **NFR-2** No SigNoz credentials, ingestion keys, API keys, tokens, or generated dashboards with secrets SHALL be committed.
- **NFR-3** Local resource usage SHALL be bounded and documented before install.
- **NFR-4** Telemetry shall not require external SaaS connectivity.
- **NFR-5** Configuration must be deterministic and pinned.
- **NFR-6** Smoke tests must fail with actionable messages.

## 6. Out of Scope

- OpenCost runtime cost allocation. That remains M4.
- Cilium, Envoy Gateway, RabbitMQ/Kafka, OpenResty, and M6 security tools.
- Full Backstage custom observability plugin development unless links and annotations are insufficient.
- Production auth and SSO for SigNoz.

## 7. Acceptance Criteria

- [ ] M3 preflight reports cluster headroom, storage class, and existing observability resources.
- [ ] SigNoz is reachable locally and reports healthy.
- [ ] OpenTelemetry collector receives OTLP from `hello-m2`.
- [ ] SigNoz shows at least one trace or metric for `hello-m2`.
- [ ] Backstage catalog entity links point to the M3 observability surface.
- [ ] Latest Infracost post-deploy artifact is visible through a dashboard or catalog link.
- [ ] `scripts/smoke-m3.sh` exits 0 on a healthy M3 install.
- [ ] `scripts/teardown-m3.sh` removes M3 resources without breaking M2.
