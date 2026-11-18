# Provenance Recognition Certificate

**Certificate ID:** PRC-developer-portal-2026-11-18-r14

**Supersedes:** PRC-developer-portal-2026-11-18-r13 (dependency
re-issue for PR #53, which adds the CodeQL advanced setup workflow
`.github/workflows/codeql.yml`). Group 6 was regenerated: row 1
(actions/checkout) adds the v7.0.1 pin
(3d3c42e5aac5ba805825da76410c181273ba90b1) used by that workflow alongside
the existing v4.4.0 pin in the Gitea CI templates, and row 9
(github/codeql-action) is restored from REMOVED to active at v4.39.3
(peeled commit 33d7b85da1a04171e211befe2c739303b9a4c066; init, autobuild
and analyze sub-actions). Both pins were re-resolved with `git ls-remote`
and both upstream LICENSE files were read at the pinned commits (MIT) on
2026-11-18. Group 6 stays at 9 entries and the component total at 216;
groups 1-5 and 7-8 are unchanged from r13. Chain history: r1-r4 were
issued and superseded before the provenance package's first commit and
exist in no retained history. r5 entered git history at commit d85e568,
r6 at 82783ee and r7 at f20cff8; r13 entered with PR #52 (6a06f65). The
retention-in-git-history rule in the Revocation and update rule section
has been in force since r5, so r5 through r13 and this r14 are all
retained.

---

## Issuer

- **Issuing party:** developer-portal (self-hosted Internal Developer
  Platform), a project of the Sovereign portfolio.
- **Repository:** /Users/nnos/Projects/Sovereign/developer-portal
- **Issuance:** by the project maintainers. This certificate is
  project-issued (self-attested); see the Attestation block.

## Subject

The third-party works recognized by this certificate are the 216
components listed in `provenance/PROVENANCE.md`, grouped as follows:

| Group | Entries |
|---|---|
| 1. Platform and Infrastructure | 15 |
| 2. Helm Charts and Container Images | 16 |
| 3. IaC Providers | 5 |
| 4. Go Tooling and Modules | 22 |
| 5. Backstage and Node.js | 146 |
| 6. CI/CD Actions and Images | 9 |
| 7. Schemas and Specifications | 1 |
| 8. Documentation Tooling | 2 |
| **Total** | **216** |

The total includes one row retained with a REMOVED marker
(@backstage/plugin-scaffolder-backend-module-github, group 5 #68,
REMOVED 2026-08-19): a removal is recorded, not erased, per the
listing's convention. github/codeql-action (group 6 #9), REMOVED
2026-08-21, is restored to active as of 2026-11-18 (r14, PR #53); its
removal history is kept in the row.

In addition, 3 UNVERIFIED items and honest gaps are recorded in the
UNVERIFIED subsection of the listing (original U-numbers retained: U7,
U19, U25), one residual caveat carried forward from the closed U17 (the
ephemeral catthehacker CI job image), and one documented accepted
residual risk (react-router/react-router-dom 6.30.4, three moderate
advisories, no fixed 6.x, v7 rejected as unsupported against all
Backstage peer ranges; accepted 2026-08-18). The listing records that
this risk is superseded: as measured 2026-11-18 the installed version is
7.18.4, outside the ^6.30.2 Backstage peer range (yarn YN0060). Runtime
impact is UNVERIFIED. cert-manager and Argo
Workflows are confirmed live as sibling-openchoreo-managed components
(v1.19.4 and v3.6.2, both Apache-2.0) but remain outside this repo's
component scope and are not listed as components.

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
| 4. Go Tooling and Modules | 22 | BSD-3-Clause, Apache-2.0, MIT, LGPL-2.1, MIT AND Apache-2.0 (x/net v0.56.0, x/sys v0.46.0, x/text v0.39.0 as of 2026-08-18; govulncheck firmed to v1.7.0, host-installed via go install) | go.dev; github.com/go-yaml/yaml; github.com/open-telemetry/opentelemetry-go |
| 5. Backstage and Node.js | 146 | Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause, ISC, BlueOak-1.0.0, (MPL-2.0 OR Apache-2.0), (BSD-3-Clause OR GPL-2.0) (node-forge, first listed 2026-11-18); 65 resolution override rows (74 resolution keys over 69 packages; 2026-11-18 PR #52 pins for proxy-addr, node-forge, compression, source-map-js, @azure/identity, isomorphic-git, techdocs-node 1.15.6, zod 4.6.7, catalog-model, cli-module-auth/new, @graphql-tools/utils 12.1.0, catalog-backend 3.9.2); vendored Yarn 4.18.0 with npmMinimalAgeGate "7d"; plugin-permission-backend-module-allow-all-policy marked REMOVED (de1ac07, 2026-08-18) | github.com/backstage/backstage; github.com/facebook/react; github.com/yarnpkg/berry |
| 6. CI/CD Actions and Images | 9 | MIT, MPL-2.0, BSD-3-Clause, GPL-2.0-only, Apache-2.0, MPL-2.0 AND MIT, MIT AND BSD-2-Clause AND GPL-2.0-or-later, Zlib; actions SHA-pinned 2026-08-18 (checkout v4.4.0, setup-go v5.6.0, setup-opentofu v1.0.8) and 2026-11-18 (checkout v7.0.1 3d3c42e5, codeql-action v4.39.3 33d7b85d, both in .github/workflows/codeql.yml); codeql-action restored from REMOVED to active 2026-11-18 (CodeQL advanced setup, PR #53); Trivy 0.74.0 and OSV-Scanner v2.5.1 digest-pinned (registry-manifest verified); base images floating-tag snapshots | github.com/actions/checkout; github.com/github/codeql-action; hub.docker.com/_/golang; hub.docker.com/_/alpine; hub.docker.com/_/node |
| 7. Schemas and Specifications | 1 | Apache-2.0 (vendored schema byte-identical to score-spec/spec commit 3ecb17d430c2bbf46d2dfc161fabc7d432d6d1f5) | github.com/score-spec/spec |
| 8. Documentation Tooling | 2 | BSD-2-Clause, Apache-2.0 | github.com/mkdocs/mkdocs; github.com/backstage/mkdocs-techdocs-core |

## Integrity

SHA-256 digests computed with `sha256sum` (GNU coreutils; same SHA-256
output as `shasum -a 256`) after the final content of both files was
written (2026-11-18, r14):

- `THIRD-PARTY-LICENSES.md`:
  `4b3f527f8243afd4359977834268ab6b4e8530b01ccb8db9fa88a7ee0050e72e`
- `provenance/PROVENANCE.md`:
  `eed5daacdf50a1a316492ab0477a0da06aea0694f490080478cdfeabb49a16ec`

Any change to either file invalidates the corresponding digest and
requires this certificate to be regenerated. Verify with:

```
shasum -a 256 THIRD-PARTY-LICENSES.md provenance/PROVENANCE.md
# or: sha256sum THIRD-PARTY-LICENSES.md provenance/PROVENANCE.md
```

and compare against the values above.

## Attestation

- **Date of issuance:** 2026-11-18 (America/New_York)
- **Issuer attestation:** Issued by the maintainers of developer-portal
  (Sovereign portfolio) on the date above, from the verified dependency
  inventory of the repository at that date. The r14 inventory was
  measured on a Linux build host from a fresh clone at PR #53's branch
  (the `Repository` path above is the maintainers' canonical checkout and
  was not used for this issue).
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
