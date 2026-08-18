# Provenance Recognition Certificate

**Certificate ID:** PRC-developer-portal-2026-08-18-r5

**Supersedes:** PRC-developer-portal-2026-08-18-r4 (same-day cluster
resolution re-issue: the 7 cluster-blocked UNVERIFIED rows were resolved
with live read-only evidence and folded into the listed entries -- Gitea,
Envoy Gateway, act-runner, k3s, Gatekeeper deployed versions and images,
plus cert-manager and Argo Workflows confirmed as sibling-managed live
context; no components added or removed). r1, r2, r3, and r4 were all
issued and superseded before the provenance package's first commit, so
they exist in no retained history; the retention-in-git-history rule in
the Revocation and update rule section applies to superseding events from
the first commit onward.

---

## Issuer

- **Issuing party:** developer-portal (self-hosted Internal Developer
  Platform), a project of the Sovereign portfolio.
- **Repository:** /Users/nnos/Projects/Sovereign/developer-portal
- **Issuance:** by the project maintainers. This certificate is
  project-issued (self-attested); see the Attestation block.

## Subject

The third-party works recognized by this certificate are the 189
components listed in `provenance/PROVENANCE.md`, grouped as follows:

| Group | Entries |
|---|---|
| 1. Platform and Infrastructure | 15 |
| 2. Helm Charts and Container Images | 16 |
| 3. IaC Providers | 5 |
| 4. Go Tooling and Modules | 22 |
| 5. Backstage and Node.js | 122 |
| 6. CI/CD Actions and Images | 6 |
| 7. Schemas and Specifications | 1 |
| 8. Documentation Tooling | 2 |
| **Total** | **189** |

In addition, 3 UNVERIFIED items and honest gaps are recorded in the
UNVERIFIED subsection of the listing (original U-numbers retained: U7,
U19, U25). They are carried openly rather than guessed or hidden. A
residual caveat carried forward from the closed U17 is recorded in the
act-runner listed entry: the catthehacker/ubuntu:act-* CI job image
exists only while a CI job is in flight; its tag is observable only
during a CI run. cert-manager and Argo Workflows are confirmed live as
sibling-openchoreo-managed components (v1.19.4 and v3.6.2, both
Apache-2.0) but remain outside this repo's component scope -- no repo
references -- and are explicitly not listed as components.

## Recognition statement

The issuer formally declares that:

1. Each work listed in `provenance/PROVENANCE.md` remains the property
   and achievement of its original authors and copyright holders, as
   recorded per entry in the listing.
2. This project claims no authorship over any of those works. Where this
   repository contains original work, it is identified as first-party in
   the listing's closing section and is not covered by this certificate.
3. Each listed work retains its original license, as recorded per entry;
   nothing in this project relicenses, regrants, or diminishes those
   licenses.
4. This certificate exists to make that recognition explicit, durable,
   and auditable: it binds the recognition to a specific generation of
   the license file and the provenance listing through the integrity
   digests below.

## Credential fields

Per-group summary of the recognized works. The full per-component listing
is `provenance/PROVENANCE.md`; this table is a summary, not a duplicate.

| Group | Entries | Licenses observed (SPDX) | Primary upstream URLs |
|---|---|---|---|
| 1. Platform and Infrastructure | 15 | MPL-2.0, Apache-2.0, MIT, BSD-2-Clause (deployed k3s server v1.32.9+k3s1; k3d host binary v5.9.0) | github.com/opentofu/opentofu; github.com/openchoreo/openchoreo; github.com/kubernetes/kubernetes |
| 2. Helm Charts and Container Images | 16 | Apache-2.0, MIT, BSD-3-Clause, PostgreSQL (Cilium userspace Apache-2.0 with eBPF objects dual GPL-2.0-only OR BSD-2-Clause); deployed versions firmed 2026-08-18: Gitea chart gitea-12.5.0 / app 1.25.4, Gatekeeper v3.17.1 (matches pin), Envoy gateway v1.3.1 with envoy distroless-v1.33.0, act_runner 0.3.1 with dind 29.4.0-dind | github.com/fluxcd/flux2; gitea.com/gitea/helm-chart; github.com/SigNoz/signoz; github.com/envoyproxy/gateway |
| 3. IaC Providers | 5 | MPL-2.0 | github.com/hashicorp/terraform-provider-kubernetes; github.com/alekc/terraform-provider-kubectl |
| 4. Go Tooling and Modules | 22 | BSD-3-Clause, Apache-2.0, MIT, LGPL-2.1, MIT AND Apache-2.0 | go.dev; github.com/go-yaml/yaml; github.com/open-telemetry/opentelemetry-go |
| 5. Backstage and Node.js | 122 | Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause, ISC, BlueOak-1.0.0, (MPL-2.0 OR Apache-2.0) | github.com/backstage/backstage; github.com/facebook/react; github.com/yarnpkg/berry |
| 6. CI/CD Actions and Images | 6 | MIT, MPL-2.0, BSD-3-Clause, GPL-2.0-only, Apache-2.0, MPL-2.0 AND MIT, MIT AND BSD-2-Clause AND GPL-2.0-or-later, Zlib (alpine:3.20 fully enumerated 2026-08-18; floating-tag snapshot; node image Debian layer DFSG-free main-only, per-package list PARTIAL) | github.com/actions/checkout; hub.docker.com/_/golang; hub.docker.com/_/alpine; hub.docker.com/_/node |
| 7. Schemas and Specifications | 1 | Apache-2.0 (vendored schema byte-identical to score-spec/spec commit 3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5) | github.com/score-spec/spec |
| 8. Documentation Tooling | 2 | BSD-2-Clause, Apache-2.0 | github.com/mkdocs/mkdocs; github.com/backstage/mkdocs-techdocs-core |

## Integrity

SHA-256 digests computed with `shasum -a 256` at generation time
(2026-08-18):

- `THIRD-PARTY-LICENSES.md`:
  `7f0bf49b8a1b17878ba579ea384596420c9c3554786fef309e3f303f78cd84d8`
- `provenance/PROVENANCE.md`:
  `0c6ba547b1412fc83470ba57e204b6b6f98a3fb3cf3f1ef8fa039bc30022fc38`

Any change to either file invalidates the corresponding digest and
requires this certificate to be regenerated. Verify with:

```
shasum -a 256 THIRD-PARTY-LICENSES.md provenance/PROVENANCE.md
```

and compare against the values above.

## Attestation

- **Date of issuance:** 2026-08-18
- **Issuer attestation:** Issued by the maintainers of developer-portal
  (Sovereign portfolio) on the date above, from the verified dependency
  inventory of the repository at that date.
- **Nature of the credential:** This certificate is project-issued
  (self-attested). It contains no external cryptographic signatures and
  no third-party endorsements. Its evidentiary value rests on the
  repository's git history, the per-entry repo evidence paths in
  `provenance/PROVENANCE.md`, and the integrity digests above.

## Revocation and update rule

When dependencies change (dependency PRs, milestone installs, lockfile
updates, chart or provider bumps), the listing `provenance/PROVENANCE.md`
is regenerated together with `THIRD-PARTY-LICENSES.md`, and a new
certificate with a new certificate ID and fresh digests supersedes this
one. Superseded certificates are not edited or deleted; they remain in
git history so the chain of recognition is immutable and auditable.
