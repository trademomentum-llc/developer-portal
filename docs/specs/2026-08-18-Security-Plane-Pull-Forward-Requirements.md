# Requirements Specification: Security Plane Pull-Forward

**Document ID:** SEC-PLANE-PULLFORWARD-REQ-001
**Version:** 0.1
**Date:** 2026-08-18
**Status:** Approved by directive -- the 2026-08-18 Tier-3 directive (section 2) is the explicit approval basis for the Wave 0 and Wave 1 stack additions (decision D5); Wave 2 items remain PROPOSED
**Relationship:** Implements the Security plane register (SEC-G1..SEC-G15) of `docs/specs/2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md` and resolves its security open questions OQ-19 through OQ-26, with one honest exception: OQ-23 is dispositioned to the control plane rather than resolved here (section 6.1). Per-wave implementation triads (Requirements -> Design Specification -> Technical Specification, per `~/Projects/Sovereign/Structure/POLICIES.md`) follow this document before any cluster state changes.
**Predecessors:** 2026-08-18-Five-Plane-Collaborative-Portal-Requirements.md; 2026-05-28-IDP-Policy-Guard-Layer triad; M1-M4 milestone triads

---

## 1. Purpose

The five-plane umbrella document verified that the security plane's namesake capability -- threat intelligence -- has zero implementation, and that the plane's detective and surface layers are largely absent (SEC-G1..SEC-G15). It carried those items as gated proposals pending open questions OQ-19..OQ-26. This document ends that deferral: it decides the open questions with documented reasoning, scopes the pulled-forward security plane into three capacity-honest waves, and states the functional and non-functional requirements each wave must satisfy. Nothing here is an open question; prerequisites are stated as steps.

## 2. Goal

User directive, 2026-08-18 (Tier-3), quoted verbatim:

> The system requires a security plane be more than just functional. Pull it forward.

"More than just functional" is defined operationally in section 12: the plane must **detect**, **inform**, and **gate** -- every component has a named consumer and an enforcement or visibility path. A scanner that never blocks, a violation no one can see, or an alert with no channel is excluded by construction.

## 3. Scope per wave

### 3.1 Wave 0 -- now, on the measured host envelope (2 vCPU / 3.9 GB)

Everything in Wave 0 is portal-side code/config, CI-burst work, or reuse of already-deployed components; it adds zero new long-running cluster workloads.

- Trivy CLI (fs + image) and OSV-Scanner in the Gitea Actions pipeline, pinned by digest/checksum, with an enforced fail threshold.
- Security scan artifacts committed to `platform-config` (SARIF/JSON), following the cost-artifact precedent.
- Custom Security tab on Component entity pages, rendering live scan and violation data.
- Gatekeeper violation visibility at zero new footprint: live constraint `.status.violations` on PolicyCard, `gatekeeper_violations` into the existing M4 Prometheus, audit-pod JSON into the existing OTEL collector -> SigNoz.
- Custom Backstage permission policy module (admin/developer/viewer) replacing allow-all.
- TLS on the three `.local` gateways via cert-manager Certificate resources + HTTPS listeners on the existing Gateway (local-CA issuer; macOS trust is a documented user step).
- In-repo `.github/dependabot.yml` + code scanning workflow for the GitHub mirror.
- Guard audit-log hash chaining (phase-2 record-immutability item, homed in the security plane).

### 3.2 Wave 1 -- after the documented one-time Colima resize

Prerequisite step (stated, not questioned): resize the Colima VM to >= 6 CPU / 12 GB and re-measure before any Wave 1 install step runs.

- Falco (modern_ebpf) + Falcosidekick OTLP -> existing OTEL collector -> SigNoz runtime detection, with codified SigNoz log-based alert rules.
- Trivy Operator for continuous in-cluster scanning (scanJobsConcurrentLimit 1-2, `trivy.slow`), VulnerabilityReports readable by the Security tab.
- MISP via misp-docker slim with the CIRCL OSINT feed enabled, as the threat-intelligence platform of record, with at least one working indicator-egress path.

### 3.3 Wave 2 -- scale-out documentation only (PROPOSED)

- Documentation of dedicated-host scale-out options: Wazuh (SIEM), OpenCTI, Velociraptor, Cloud Custodian -- capacity envelopes, trigger conditions, disqualifications. No installs. Wave 2 items remain PROPOSED pending their own explicit approvals.

### 3.4 Out of scope

- Implementation itself (this is a requirements document; per-wave triads follow).
- TheHive 5 / Cortex at any wave -- disqualified on license (section 10).
- Any claim of SIEM capability from the SigNoz pipeline (section 6, D3).
- The record-immutability mechanism itself (companion workstream, TODO.md goal slice 2); this document only honors its touchpoints.
- Guard multi-user distribution (SEC-G13/TRV-B11): a control-plane collaboration item (five-plane FR-30), not re-scoped here.
- OpenBao dev-mode storage backend (SEC-G14): M2 backlog item m2i-6 (`TODO.md:156`), not re-scoped here; no Wave 0/1 item depends on it.

## 4. Current state summary (from the SEC register)

Verified 2026-08-18 in the five-plane document, section 4.4; evidence paths cited there.

- **Live (preventive, pipeline-scoped):** six rr-policy-guards; Gatekeeper C1-C3 pipeline constraints with Rego tests; publication SCA gate (Semgrep, Gitleaks, yarn/npm audit, govulncheck); local append-only guard audit JSONL; SECURITY.md disclosure policy; dependency remediation posture; Gitea OAuth portal authentication; production auth hardening.
- **Partial:** Backstage permission framework enabled in production but wired to the allow-all policy (SEC-G12); OpenBao live but dev-mode inmem (SEC-G14); GitHub-side scanning referenced but no in-repo config (SEC-G11).
- **Scaffold:** PolicyCard renders static C1/C2/C3 links and states violations "will appear here once the M3 policy collector is wired" (`PolicyCard.tsx:50-53`) -- it is not wired (SEC-G8, TRV-B8).
- **Absent (the pull-forward targets):** threat intelligence platform (SEC-G1), image vulnerability scanning (SEC-G2), runtime detection (SEC-G3), security event aggregation (SEC-G4), incident response (SEC-G5), endpoint DFIR (SEC-G6), cloud governance (SEC-G7), TLS on `.local` routes (SEC-G9), any Security surface in the portal (SEC-G10, TRV-B6), portable guard distribution (SEC-G13), M6 spec package (SEC-G15 -- closed by this document).

## 5. Capacity analysis (honest numbers)

Measured on the target host 2026-08-18; this measurement is the environmental fact of record (NFR-02).

- Host: Colima VM, **2 vCPU / 3.9 GB RAM**, ~84% memory already in use (**~655 MB free**); kernel 6.8.0-100-generic, aarch64, **BTF present**; k3d single node, k3s v1.32.9+k3s1.
- Older repo documents describing the host as 6 CPU / 10 GB are **stale** as of this measurement and are flagged as such wherever they appear.
- **Wave 0 fits the measured envelope:** Trivy CLI and OSV-Scanner run as ephemeral act-runner CI job pods (burst-only); Gatekeeper visibility reuses the deployed audit pod, the existing M4 Prometheus, and the existing OTEL collector (zero new footprint); Security tab, RBAC, and TLS are portal code/config; dependabot and code scanning run GitHub-side.
- **Wave 1 post-resize budget (>= 6 CPU / 12 GB):** Falco DaemonSet requests 100m/512Mi per node; Falcosidekick is a lightweight forwarder; Trivy Operator scan jobs request 100m/100Mi, limit 500m/500Mi, with scanJobsConcurrentLimit 1-2 and `trivy.slow`; MISP realistic ~3-4 GB with slim images and reduced workers (official whole-host sizing 2+ cores / 8-16 GB).
- **Does not fit even after resize:** Wazuh all-in-one (4c/8GB dedicated minimum), OpenCTI (~16 GB with ES+Redis+RabbitMQ+MinIO), TheHive 5 (9c/12GB plus Cortex 4.x at 8c/16GB with its own Elasticsearch). These are Wave 2 documentation-only (section 10).

### 5.1 Environmental flags (recorded, owned elsewhere)

- cert-manager v1.19.4 is sibling-managed; the 1.19 line is EOL 2026-07-08 (upstream 1.21.x). The sibling-owned upgrade is flagged; Wave 0 TLS works on the deployed version.
- Envoy Gateway is pinned at chart 1.3.1 against upstream v1.9.0; drift flagged for the networking module owner. The Wave 0 TLS pattern (section 8, FR-09) needs no controller change.
- SigNoz runs demo-grade retention (emptyDir, 3d logs). Wave 0/1 security events land on that pipeline knowingly (D3); durability uplift is an observation-plane item (five-plane FR-10/OQ-06), not silently assumed here.

## 6. Decisions (decided, with reasoning and evidence)

| ID | Decision | Reasoning | Evidence |
|---|---|---|---|
| D1 | **Wave structure by capacity reality.** Wave 0 (now, 2 vCPU / 3.9 GB): Trivy CLI + OSV-Scanner in Gitea CI pinned by digest/checksum; Gatekeeper violation visibility (constraint `.status.violations` + `gatekeeper_violations` metric into existing M4 Prometheus + audit JSON into existing OTEL/SigNoz); custom Security tab; custom RBAC permission policy module; TLS via Certificate resources + HTTPS listeners (local-CA issuer; macOS trust = documented user step); in-repo dependabot.yml + code scanning for the GitHub mirror; guard audit-log hash chaining. Wave 1 (after the documented one-time Colima resize to >= 6 CPU / 12 GB -- a stated prerequisite step, not a question): Falco + Falcosidekick OTLP -> SigNoz; Trivy Operator (scanJobsConcurrentLimit 1-2, `trivy.slow`); MISP via misp-docker slim with CIRCL OSINT feed. Wave 2: scale-out documentation only. | The measured envelope (~655 MB free) admits zero new long-running cluster workloads; CI-burst and portal-side work fits today, and the heavy-but-valuable items fit after one documented resize. Splitting this way delivers detection, information, and gating now instead of blocking the whole plane on a host upgrade. | Host measurement 2026-08-18 (section 5); per-component footprints (section 5); five-plane OQ-19/OQ-21 resolved here. |
| D2 | **MISP is the threat-intelligence platform of record.** Full REST API (+ PyMISP) is the indicator-egress mechanism. | It is the only open-source TIP that fits the capacity envelope (~3-4 GB slim with reduced workers). OpenCTI (~16 GB dependency stack) fails capacity; TheHive 5 fails license (proprietary source-available). AGPL-3.0 satisfies the open-source-only rule; compliance note in section 7. | Research digest 2026-08-18: MISP 2.5.44 AGPL-3.0 core, misp-docker GPL-3.0; OpenCTI 7.x sizing; TheHive 5.7.5 license drift. Resolves OQ-22. |
| D3 | **The SigNoz pipeline is the security-event sink at this scale** -- Falcosidekick native OTLP Logs output + Gatekeeper audit JSON + SigNoz log-based alert rules. It is labeled honestly, in docs and UI, as **security observability, not a SIEM** (no correlation engine, no MITRE packs, no case management). | Reuses the M3 pipeline at zero new backend footprint on a host with ~655 MB free; Wazuh (4c/8GB dedicated) does not fit and is Wave 2 documentation. Honest labeling is required by NFR-03 (no fabricated status). | Falcosidekick 2.34.1 native OTLP Logs output; SigNoz log-based alert rules; Wazuh v4.14.7 sizing. Resolves OQ-26. |
| D4 | **Component admission requirements.** Every new component requires: (a) license verified against the register in section 7; (b) provenance listing regeneration + recognition-certificate re-issue at implementation time (attribution triple; attribution recorded, never claimed); (c) pin-by-digest discipline -- exact version + SHA256 checksum or image digest, never mutable tags. | The March 2026 Trivy supply-chain compromise (CVE-2026-33634: malicious `trivy-action` tags force-pushed) demonstrates that mutable tags are an attack vector into CI itself. The repo already SHA-pins its CI actions as of 2026-08-18; the same discipline extends to every security tool this document admits. | CVE-2026-33634; repo CI pin posture 2026-08-18; portfolio attribution-triple convention (AGENTS.md). |
| D5 | **Approval basis.** The 2026-08-18 Tier-3 directive ("The system requires a security plane be more than just functional. Pull it forward") is recorded as the explicit user approval for the Wave 0 and Wave 1 stack additions (Trivy, OSV-Scanner, Falco, Falcosidekick, Trivy Operator, MISP, plus the portal-side RBAC/TLS/tab work) -- the same discipline as the 2026-04-20 Flux approval and the M2 Gatekeeper pull-forward. Wave 2 items remain PROPOSED. | NFR-02 of the five-plane document requires explicit approval for anything outside the locked-in stack; the directive is that approval for this plane's Wave 0/1 scope, and this record is the audit trail. Wave 2 gets no implied approval. | User directive 2026-08-18; AGENTS.md locked-in stack note (Flux approval 2026-04-20). Resolves OQ-19 and OQ-20. |
| D6 | **Binding constraint set.** The NFRs in section 9 are adopted as decisions, not aspirations: open-source-only (license register of record); capacity honesty (the 2c/4GB measurement is the environmental fact; stale 6c/10GB doc claims flagged); no fabricated status (live data or explicit "not wired"); pin-by-digest; provenance regeneration; smoke coverage for every wave item (extend the smoke-all pattern); WCAG 2.1 AA on new UI surfaces; record-immutability touchpoints (security events are records; security decisions journaled/ADR'd). | "More than just functional" fails silently without enforceable constraints; writing them as decisions makes each one checkable in review and smoke. | Five-plane NFR-01..NFR-10 practice; M3/M4 smoke-all precedent; repo provenance convention. |

### 6.1 Resolved and dispositioned open questions

Seven of the eight security open questions from the five-plane document are decided here. The eighth, OQ-23, is dispositioned to the control-plane workstream rather than resolved in this document; it is recorded here so the register stays complete and the disposition is explicit rather than silent.

| OQ (five-plane doc) | Resolution |
|---|---|
| OQ-19 (slice vs M6 ownership) | Pulled forward now, by directive (D5). This document is the security plane requirements package. |
| OQ-20 (stack approvals) | Wave 0/1 additions approved by the 2026-08-18 directive (D5); recorded here, Flux/Gatekeeper-style. |
| OQ-21 (capacity) | Wave structure by measured envelope + documented resize step (D1, section 5). |
| OQ-22 (TIP vs feed-driven awareness) | Both, sequenced: feed-driven scanner awareness in Wave 0 (FR-01..FR-04); MISP as TIP of record in Wave 1 (D2, FR-14). |
| OQ-23 (guard portability) | Dispositioned to the control plane: portable guard distribution is the control-plane collaboration item five-plane FR-30 (TRV-B11), not re-scoped here. The documentation half -- the guard README's phantom in-repo `hooks/hooks.json` reference -- was already fixed by the 2026-08-18 anomaly pass. The security plane's own guard-layer deliverable is trail integrity (FR-11). |
| OQ-24 (TLS posture) | cert-manager Certificate-resource pattern on the existing Gateway (FR-09). |
| OQ-25 (authorization model) | Custom RBAC permission policy module, three roles (FR-08). |
| OQ-26 (audit centralization) | SigNoz pipeline as the honestly-labeled security-event sink (D3, FR-07, FR-12). |

## 7. License register (of record)

| Component | Version | License | Role | Notes |
|---|---|---|---|---|
| Trivy CLI | v0.74.0 | Apache-2.0 | CI fs/image/config scanning (Wave 0) | Pin exact version + SHA256 checksum or image digest; never mutable action tags (CVE-2026-33634) |
| OSV-Scanner | v2.5.1 | Apache-2.0 | Dependency advisory scanning in CI (Wave 0) | Offline mode capable; complements Trivy (OS/distro/IaC/secrets) |
| Falco | 0.44.1 | Apache-2.0 | Runtime detection (Wave 1) | Drivers are kmod or modern_ebpf only (userspace path removed); modern_ebpf verified on kernel 6.8.0-100-generic (BTF present); least-privileged mode exists |
| Falcosidekick | 2.34.1 | Apache-2.0 | Falco event forwarding (Wave 1) | Apache-2.0 since 2.29.0; the README MIT badge is stale |
| Trivy Operator | chart 0.35.0 / app 0.33.0 | Apache-2.0 | Continuous in-cluster scanning (Wave 1) | Emits VulnerabilityReport CRDs + Prometheus metrics; scanJobsConcurrentLimit tunable |
| MISP | 2.5.44 | AGPL-3.0 (core); misp-docker packaging GPL-3.0 | Threat-intel platform of record (Wave 1) | Slim variants; default feeds ship disabled incl. CIRCL OSINT; full REST API + PyMISP |
| Gatekeeper | 3.17 | Apache-2.0 | Admission policy (deployed) | No change; audit consumed by FR-05..FR-07 |
| cert-manager | v1.19.4 | Apache-2.0 | Certificate issuance (sibling-managed) | 1.19 line EOL 2026-07-08; sibling-owned upgrade flagged (section 5.1) |
| @backstage-community/plugin-rbac-backend | n/a | Apache-2.0 | Evaluated RBAC alternative, not selected | Custom policy module chosen: no new runtime dependency, roles are few and stable, config-as-code in repo; the plugin's management UI is unneeded at this scale |

**Copyleft compliance note (AGPL-3.0 / GPL-3.0 / GPL-2.0):** MISP (AGPL-3.0), misp-docker (GPL-3.0), and Wave 2's Wazuh (GPL-2.0) run self-hosted for internal platform-engineering use. The platform is not offered to third parties as a network service and nothing is resold, so the AGPL network-use clause is not triggered beyond what ordinary internal operation requires; upstream source remains available and any modifications will be published per the licenses. Internal self-hosted use is compliant, and these copyleft licenses satisfy the portfolio's open-source-only rule.

## 8. Functional requirements

Every FR traces to a verified SEC-G gap or TRV-B breakdown from the five-plane document and/or to a decision in section 6. No invented features.

### 8.1 Wave 0 (measured envelope: 2 vCPU / 3.9 GB)

- **FR-01:** The canonical app pipeline (`seed-repos/hello-m2/.gitea/workflows/ci.yaml` and `iac/templates/ci.yaml`) runs Trivy `fs` scanning on the repository and Trivy `image` scanning on the built image before push to the registry. The job fails when findings at or above the enforced severity threshold (default: HIGH and CRITICAL) are present, with a reviewed, dated suppressions list as the only bypass. Trivy is pinned by exact version + SHA256 checksum or image digest. **Traces to:** SEC-G2; D1, D4.
- **FR-02:** OSV-Scanner runs in the same pipeline against dependency manifests/lockfiles, emitting advisory-precise JSON/SARIF; failures follow the FR-01 threshold policy. Pinned per D4. **Traces to:** SEC-G2; D1, D4.
- **FR-03:** Scan outputs (SARIF + JSON) are committed per push into `platform-config` under `security-artifacts/<app>/<env>/`, following the contents-API precedent of `scripts/ci/commit-cost-artifact.sh`, producing an append-only, digest-friendly record consumable by the portal. **Traces to:** SEC-G10, SEC-G2; D1; honors five-plane FR-40/NFR-05.
- **FR-04:** A Security tab is added to Component entity pages via the repo's EntityContentBlueprint pattern (`backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx`), rendering the latest FR-03 scan artifacts and violation state for the entity. Any data source not yet wired renders an explicit "not wired" state, never a placeholder posing as live. **Traces to:** SEC-G10, TRV-B6; D1, D6.
- **FR-05:** PolicyCard is rewired from static links to live Gatekeeper constraint `.status.violations` state (Kubernetes API via a backend route, scoped by the predicted namespace), replacing the "once the M3 policy collector is wired" scaffold text (`PolicyCard.tsx:50-53`). **Traces to:** SEC-G8, TRV-B8; D1.
- **FR-06:** The existing M4 Prometheus scrapes `gatekeeper_violations`; the metric is queryable and available to the Security tab and to future alert rules. **Traces to:** SEC-G8; D1.
- **FR-07:** Gatekeeper audit-pod JSON stdout is collected into the existing standalone OTEL collector and forwarded to SigNoz, making violation events queryable in the security-event sink. **Traces to:** SEC-G4; D1, D3.
- **FR-08:** The allow-all permission policy is replaced by a custom PermissionPolicy backend module (new-backend `policyExtensionPoint` pattern) defining **admin / developer / viewer** roles resolved from Gitea org group claims carried by the existing sign-in resolver (`ownershipEntityRefs` group claims, `catalogConditions.isEntityOwner` for ownership-scoped rules). **Traces to:** SEC-G12; D1. Resolves OQ-25.
- **FR-09:** `gitea.local`, `signoz.local`, and `opencost.local` serve HTTPS: cert-manager Certificate resources created by this repo (SelfSigned ClusterIssuer for smoke; local-CA ClusterIssuer for real use) referenced by HTTPS listeners with `certificateRefs` added to the existing Gateway (`iac/modules/networking/gateway/main.tf`, HTTP-only today). The gateway-shim annotation pattern is rejected (it requires `enableGatewayAPI` on the sibling-owned controller). macOS keychain trust of the local CA is a documented user step. **Traces to:** SEC-G9; D1. Resolves OQ-24.
- **FR-10:** `.github/dependabot.yml` and a code scanning workflow for the GitHub mirror are version-controlled in this repo, so scanner posture is config-as-code and no longer depends on GitHub-side settings surviving the mirror relationship. **Traces to:** SEC-G11; D1.
- **FR-11:** Guard audit JSONL entries carry a SHA-256 hash chain (each entry commits to its predecessor), verifiable by a guard subcommand; the chain format is compatible with the separately-specified record-immutability mechanism (phase-2 item, homed in the security plane). Cross-document note: the record-immutability triad's OQ-04 asks both whether chaining is in the phase-2 batch at all and which guards it covers; this document resolves the "at all" half under the Tier-3 approval basis (D5) -- Wave 0 ships chaining -- while the per-guard scope half remains OQ-04's to decide. The chain format stays compatible with RECORD-IMMUTABILITY-TECH-001 section 9 (phase-2 guard audit-log hash chaining sketch). **Traces to:** SEC-G4; CTL-G6 touchpoint; D1.

### 8.2 Wave 1 (prerequisite step: Colima resized to >= 6 CPU / 12 GB and re-measured)

- **FR-12:** Falco 0.44.1 runs with the modern_ebpf driver (verified on kernel 6.8.0-100-generic, BTF present) in least-privileged mode; Falcosidekick forwards events via its native OTLP Logs output through the existing OTEL collector into SigNoz; codified SigNoz log-based alert rules fire on Falco priority thresholds. **Traces to:** SEC-G3, SEC-G4; D1, D3.
- **FR-13:** Trivy Operator (chart 0.35.0 / app 0.33.0) continuously scans running workloads with scanJobsConcurrentLimit 1-2 and `trivy.slow`; its VulnerabilityReport CRDs are readable by the Security tab, so per-component posture reflects the running image, not only the last CI run. **Traces to:** SEC-G2, SEC-G10; D1, D4.
- **FR-14:** MISP 2.5.44 is deployed via misp-docker slim variants with reduced workers; feeds ship disabled by default and are enabled deliberately, starting with the CIRCL OSINT feed; at least one indicator-egress path works end to end (MISP attributes `restSearch` consumed by a policy or blocklist artifact in this platform). **Traces to:** SEC-G1; D1, D2.

### 8.3 Wave 2 (documentation only; PROPOSED)

- **FR-15:** A scale-out document describes dedicated-host deployment of Wazuh (SIEM) and the deferral conditions of OpenCTI, Velociraptor, and Cloud Custodian -- capacity envelopes, trigger conditions, and the TheHive disqualification record -- without installing anything. **Traces to:** SEC-G4 (SIEM proper), SEC-G5, SEC-G6, SEC-G7; D1.

## 9. Non-functional requirements

- **NFR-01 Open-source only.** Every added component is open-source and appears in the section 7 license register; AGPL-3.0/GPL-3.0/GPL-2.0 are acceptable copyleft for self-hosted internal use (compliance note, section 7); proprietary source-available components are disqualified (TheHive 5, section 10).
- **NFR-02 Capacity honesty.** The measured envelope (2 vCPU / 3.9 GB, ~655 MB free, 2026-08-18) is the environmental fact of record; older 6c/10GB claims in repo docs are stale and flagged; Wave 1 may not begin before the resize step is executed and re-measured.
- **NFR-03 No fabricated status.** Every security surface shows live data or an explicit "not wired" state; the security-event sink is labeled security observability, never "SIEM".
- **NFR-04 Pin-by-digest.** Every admitted tool is pinned by exact version + SHA256 checksum or image digest; mutable action tags are prohibited (CVE-2026-33634).
- **NFR-05 Provenance regeneration.** Each wave's implementation regenerates `THIRD-PARTY-LICENSES.md` and `provenance/PROVENANCE.md` and re-issues `provenance/PROVENANCE-RECOGNITION-CERTIFICATE.md`; superseded certificates stay in git history. Attribution is recorded, never claimed.
- **NFR-06 Smoke coverage.** Every Wave 0/1 item ships with smoke checks wired into the `smoke-all.sh` pattern (a `smoke-security.sh` suite per wave); `smoke-all.sh` remains green after each wave and the M2 delivery contract is preserved.
- **NFR-07 Accessibility.** New UI surfaces (Security tab, rewired PolicyCard) meet WCAG 2.1 AA essentials: keyboard navigation, sufficient color contrast, labeled controls, screen-reader-meaningful structure.
- **NFR-08 Record-immutability touchpoints.** Security events are records: scan artifacts are append-only and digest-friendly (FR-03), the guard audit trail is hash-chained (FR-11), and security decisions (this document, the resize step, feed enablement, suppressions) are journaled/ADR'd per repo convention.
- **NFR-09 Approval discipline.** Wave 0/1 are approved by the 2026-08-18 Tier-3 directive (D5); Wave 2 items remain PROPOSED; anything beyond this document requires its own explicit approval.
- **NFR-10 Governance.** Per-wave implementation proceeds via its own Requirements/Design/Technical triad per POLICIES.md before any cluster state changes.

## 10. Wave 2 dispositions (disqualifications and deferrals, with reasons)

| Candidate | Version | Disposition | Reason |
|---|---|---|---|
| TheHive 5 + Cortex | 5.7.5 / Cortex 4.x | **DISQUALIFIED** | License drift: TheHive 5 is proprietary source-available (StrangeBee); TheHive 3/4 were AGPL but their upstream support ended 2021-12-31 (v3) and 2022-12-31 (v4), and 2025-08-08 was the end-of-public-availability announcement (repositories pulled), not a support-EOL date. Fails NFR-01 outright. Capacity independently disqualifying: 9c/12GB plus Cortex 8c/16GB with its own Elasticsearch. Incident-response case management remains an open capability gap with no fitting open-source candidate on this host; recorded here, not silently dropped. |
| Wazuh (SIEM) | v4.14.7 | Deferred to Wave 2 documentation | GPL-2.0 acceptable, but all-in-one needs 4c/8GB dedicated minimum -- does not fit even the resized envelope as a co-tenant. D3 covers the interim honestly. |
| OpenCTI | 7.x calver | Deferred | Apache-2.0 CE acceptable, but ES+Redis+RabbitMQ+MinIO+platform ~16 GB does not fit. MISP is the TIP of record (D2). |
| Velociraptor | 0.77.2 | Out of scope (lab-only) | AGPL-3.0 and fits, but there is no endpoint fleet on a single dev host; nothing to defend. |
| Cloud Custodian | 0.9.51.0 | Deferred until cloud scope exists | Apache-2.0, but with no cloud accounts it duplicates Gatekeeper; revisit when the platform gains cloud scope. |

## 11. SEC-G coverage map

Every SEC register entry has a requirement or a recorded disposition. No gap is left hanging.

| Gap | Disposition |
|---|---|
| SEC-G1 (no TIP) | FR-14 (Wave 1 MISP, D2); feed-driven awareness precursor FR-01..FR-04 |
| SEC-G2 (no image scanning) | FR-01, FR-02 (CI), FR-03 (artifacts), FR-13 (continuous, Wave 1) |
| SEC-G3 (no runtime detection) | FR-12 (Wave 1 Falco) |
| SEC-G4 (no security event aggregation) | FR-07 (audit logs), FR-11 (hash-chained guard trail), FR-12 (Falco events); SIEM proper deferred to Wave 2 documentation (D3, FR-15) |
| SEC-G5 (no incident response) | TheHive DISQUALIFIED on license (section 10); gap recorded as open with no fitting OSS candidate on this host |
| SEC-G6 (no endpoint DFIR) | Velociraptor lab-only, out of scope on a single dev host (section 10) |
| SEC-G7 (no cloud governance) | Cloud Custodian deferred until cloud scope exists (section 10) |
| SEC-G8 (violation visibility absent) | FR-05 (PolicyCard live), FR-06 (Prometheus metric) |
| SEC-G9 (no TLS on portal routes) | FR-09 |
| SEC-G10 (no security surface) | FR-04 (Security tab), fed by FR-03 and FR-13 |
| SEC-G11 (no in-repo scanner config) | FR-10 |
| SEC-G12 (allow-all authorization) | FR-08 |
| SEC-G13 (guard registration not portable) | Control-plane collaboration item (five-plane FR-30 / TRV-B11); not re-scoped here. The security plane's guard-layer deliverable is trail integrity (FR-11). |
| SEC-G14 (OpenBao dev-mode) | M2 backlog m2i-6 (`TODO.md:156`); no Wave 0/1 item depends on it |
| SEC-G15 (no M6 spec package) | Closed by this document; per-wave triads follow (NFR-10) |

## 12. What "more than just functional" means here

A functional plane exists. More than functional means it **detects**, **informs**, and **gates**: every Wave 0/1 component has a named consumer and a live enforcement or visibility path. This table is the acceptance test for that standard.

| Component | Consumer | Enforcement / visibility path |
|---|---|---|
| Trivy in CI (FR-01) | CI pipeline; Security tab | Gates: job fails at/above the severity threshold; artifact visible per component |
| OSV-Scanner (FR-02) | CI pipeline; Security tab | Gates merges on dependency advisories; results published |
| Scan artifacts (FR-03) | Security tab; auditors | Append-only, digest-friendly record in platform-config |
| Security tab (FR-04) | Team member on the entity page | The plane's portal home; live data or explicit "not wired" |
| PolicyCard (FR-05) | Team member; reviewers | Live violation state per predicted namespace |
| `gatekeeper_violations` (FR-06) | Prometheus; future alert rules | Time-series of admission denials |
| Gatekeeper audit logs (FR-07) | SigNoz queries + alert rules | Violation events queryable in the sink |
| RBAC policy (FR-08) | Every authenticated portal request | Enforces admin/developer/viewer at the API |
| TLS on `.local` routes (FR-09) | Every route user | Protects credentials and content in transit |
| dependabot + code scanning (FR-10) | GitHub mirror; maintainers | Automated update PRs and scanning on the public mirror, config-as-code |
| Guard hash chain (FR-11) | verify-guard; auditors | Tamper-evident guard audit trail |
| Falco + Falcosidekick (FR-12) | SigNoz alert rules; operator | Runtime detection -> notification |
| Trivy Operator (FR-13) | Security tab | Continuous posture of running images |
| MISP (FR-14) | Policies/blocklists via restSearch | Threat intelligence -> enforcement artifact |

Conversely, this document refuses three failure modes by construction: a scanner that never blocks (FR-01/FR-02 enforce thresholds), a violation no one can see (FR-05..FR-07 surface it three ways), and an alert with no channel or a backend wearing a SIEM label it has not earned (D3, NFR-03).

## 13. Success criteria

- **Wave 0 exit:** a CI run fails on a seeded HIGH/CRITICAL finding; scan artifacts are committed per push; the Security tab renders them live; PolicyCard shows live constraint violations; `gatekeeper_violations` is queryable in Prometheus; audit events are queryable in SigNoz; the three roles are enforced by the permission backend; the three `.local` routes serve HTTPS; dependabot and code scanning are in-repo; the guard hash chain verifies; `smoke-all.sh` (extended with the security suite) is green.
- **Wave 1 exit (post-resize):** a triggered Falco rule produces a SigNoz alert; a VulnerabilityReport for a running image is visible on the Security tab; the CIRCL OSINT feed is enabled in MISP and one indicator-egress path is demonstrated end to end.
- **Wave 2 exit:** the scale-out documentation (FR-15) is published; nothing is installed.

---

**End of Requirements Specification**

This document was created per the governing persona rules and the POLICIES.md triad rule, implementing the Security plane register of the 2026-08-18 five-plane requirements document under the 2026-08-18 Tier-3 directive. Per-wave Design Specifications and Technical Specifications follow before implementation.
