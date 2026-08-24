#!/usr/bin/env bash
# scripts/smoke-security.sh
# Security plane smoke suite (Wave 0). Harness skeleton per
# docs/specs/2026-08-18-Security-Plane-Wave0-Technical-Specification.md
# section 14.2: Lane A creates this file; each later lane APPENDS its own
# check block (append-only; no lane edits another lane's block).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"

source "${REPO_ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin security

PASSED=0
FAILED=0
SKIPPED=0

info() { echo "[smoke-security] $*"; }
pass() { echo "PASS: $*"; PASSED=$((PASSED + 1)); smoke_json_count pass; }
fail() { echo "FAIL: $*" >&2; FAILED=$((FAILED + 1)); smoke_json_count fail; }
skip() { echo "SKIP: $*"; SKIPPED=$((SKIPPED + 1)); smoke_json_count skip; }

# yaml_ok <file>: 0 = parses, 1 = invalid, 2 = no YAML parser available.
yaml_ok() {
    local f=$1
    if command -v yq >/dev/null 2>&1; then
        yq e '.' "$f" >/dev/null 2>&1
    elif python3 -c 'import yaml' >/dev/null 2>&1; then
        python3 -c 'import sys, yaml; yaml.safe_load(open(sys.argv[1]))' "$f" >/dev/null 2>&1
    else
        return 2
    fi
}

# =============================================================================
# Lane A -- CI scanning (FR-01 Trivy, FR-02 OSV-Scanner, FR-03 artifacts)
# =============================================================================
info "Lane A: CI scanning (FR-01, FR-02, FR-03)"

SEED_WORKFLOW="${REPO_ROOT}/seed-repos/hello-m2/.gitea/workflows/ci.yaml"
TEMPLATE_WORKFLOW="${REPO_ROOT}/iac/templates/ci.yaml"
TRIVY_DIGEST="sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969"
OSV_DIGEST="sha256:8108ae94eadea5a02c9bec6e646909d5b790b44bd62d7f5b7f0b1d6d0ffc7734"

LANE_A_STEPS=(
    "Trivy filesystem scan (gate)"
    "Trivy filesystem report (SARIF artifact)"
    "OSV dependency scan (gate)"
    "Trivy image scan (gate)"
    "Trivy image report (SARIF artifact)"
    "Assemble security scan artifact"
    "Commit security artifact to platform-config"
)

for wf in "${SEED_WORKFLOW}" "${TEMPLATE_WORKFLOW}"; do
    rel=${wf#"${REPO_ROOT}"/}

    # A1: workflow file parses as YAML.
    yaml_ok "${wf}" && rc=0 || rc=$?
    if [ "${rc}" -eq 0 ]; then
        pass "${rel} is valid YAML"
    elif [ "${rc}" -eq 2 ]; then
        skip "${rel} YAML validity (no yq or python3+yaml on this host)"
    else
        fail "${rel} does not parse as YAML"
    fi

    # A2: all Lane A scan/artifact steps are present.
    missing=()
    for step in "${LANE_A_STEPS[@]}"; do
        grep -qF "name: ${step}" "${wf}" || missing+=("${step}")
    done
    if [ "${#missing[@]}" -eq 0 ]; then
        pass "${rel} has all Lane A scan/artifact steps"
    else
        fail "${rel} missing steps: ${missing[*]}"
    fi

    # A3: both scanner images are referenced by their spec digests.
    if grep -qF "aquasec/trivy@${TRIVY_DIGEST}" "${wf}" \
        && grep -qF "ghcr.io/google/osv-scanner@${OSV_DIGEST}" "${wf}"; then
        pass "${rel} references the pinned Trivy and OSV-Scanner digests"
    else
        fail "${rel} is missing a spec-pinned scanner digest"
    fi

    # A4: no mutable scanner tags and no version-tagged actions.
    if grep -nE '(aquasec/trivy|google/osv-scanner):' "${wf}" >/dev/null 2>&1; then
        fail "${rel} has a non-digest (tag) scanner image reference"
    elif grep -nE 'uses: +[^ ]+@v[0-9]' "${wf}" >/dev/null 2>&1; then
        fail "${rel} has an action pinned by version tag instead of commit SHA"
    else
        pass "${rel} has no mutable scanner tags or version-tagged actions"
    fi
done

# A5: CI helper script and suppression seeds exist; the script is executable.
if [ -x "${REPO_ROOT}/scripts/ci/commit-security-artifacts.sh" ]; then
    pass "scripts/ci/commit-security-artifacts.sh exists and is executable"
else
    fail "scripts/ci/commit-security-artifacts.sh missing or not executable"
fi
for f in seed-repos/hello-m2/.trivyignore seed-repos/hello-m2/osv-scanner.toml; do
    if [ -f "${REPO_ROOT}/${f}" ]; then
        pass "${f} exists"
    else
        fail "${f} missing"
    fi
done

# A6: FR-03 artifact readable via the Gitea Contents API (spec 3.5).
# Degrades to SKIP with a printed reason when the token, the port-forward,
# or the first post-Lane-A push is not there yet.
TOKEN_FILE="${HOME}/.rational-reserve/m1-gitea-token"
ARTIFACT_API="http://localhost:3333/api/v1/repos/openchoreo/platform-config/contents/security-artifacts/hello-m2/development/latest.json"
if [ ! -f "${TOKEN_FILE}" ]; then
    skip "FR-03 artifact read (${TOKEN_FILE} not present)"
else
    GITEA_TOKEN="$(tr -d '[:space:]' < "${TOKEN_FILE}")"
    body=$(mktemp)
    http_code=$(curl -sS --max-time 5 -o "${body}" -w '%{http_code}' \
        -u "gitea_admin:${GITEA_TOKEN}" "${ARTIFACT_API}" 2>/dev/null || true)
    case "${http_code}" in
        200)
            if jq -er '.content // empty' "${body}" | base64 -d 2>/dev/null \
                | jq -e '.artifact_type == "security-scan"' >/dev/null 2>&1; then
                pass "security-artifacts/hello-m2/development/latest.json readable, artifact_type=security-scan"
            else
                fail "latest.json exists but does not decode to an artifact_type=security-scan document"
            fi
            ;;
        404)
            skip "FR-03 artifact read (no security artifact committed yet; first push after Lane A pending)"
            ;;
        000|"")
            skip "FR-03 artifact read (Gitea port-forward on localhost:3333 unreachable)"
            ;;
        *)
            fail "FR-03 artifact read returned HTTP ${http_code}"
            ;;
    esac
    rm -f "${body}"
fi

# =============================================================================
# Lane B -- Gatekeeper visibility (FR-05, FR-06, FR-07)
# =============================================================================
info "Lane B: Gatekeeper visibility (FR-05, FR-06, FR-07)"

# FR-05: app-config.yaml carries the kubernetes cluster locator.
if grep -q "authProvider: localKubectlProxy" "${REPO_ROOT}/backstage/app-config.yaml" \
  && grep -q "name: k3d-openchoreo-local" "${REPO_ROOT}/backstage/app-config.yaml" \
  && grep -q "url: http://localhost:8001" "${REPO_ROOT}/backstage/app-config.yaml"; then
    pass "app-config.yaml kubernetes block (localKubectlProxy, k3d-openchoreo-local)"
else
    fail "app-config.yaml kubernetes block missing or changed"
fi

# FR-05: start-backstage.sh manages the kubectl proxy on :8001; stop script reaps it.
if grep -q "ensure_kubectl_proxy" "${REPO_ROOT}/scripts/start-backstage.sh" \
  && grep -q "kubectl --context k3d-openchoreo proxy --port=8001" "${REPO_ROOT}/scripts/start-backstage.sh" \
  && grep -q "kubectl-proxy-8001.pid" "${REPO_ROOT}/scripts/stop-backstage.sh"; then
    pass "kubectl proxy manager wired in start/stop scripts"
else
    fail "kubectl proxy manager missing in start/stop scripts"
fi

# FR-05: PolicyCard renders live constraint state via the shared gatekeeper helper.
if [ -f "${REPO_ROOT}/backstage/packages/app/src/modules/openchoreo-cards/gatekeeper.ts" ] \
  && grep -q "fetchConstraint" "${REPO_ROOT}/backstage/packages/app/src/modules/openchoreo-cards/PolicyCard.tsx" \
  && grep -q "Backstage-Kubernetes-Cluster" "${REPO_ROOT}/backstage/packages/app/src/modules/openchoreo-cards/gatekeeper.ts" \
  && grep -q "not wired (kubernetes proxy unavailable" "${REPO_ROOT}/backstage/packages/app/src/modules/openchoreo-cards/PolicyCard.tsx"; then
    pass "PolicyCard live Gatekeeper wiring present"
else
    fail "PolicyCard still static or gatekeeper.ts missing"
fi

# FR-06: Prometheus values carry the gatekeeper pod-role scrape job.
if grep -q "job_name: gatekeeper" "${REPO_ROOT}/observability/cost/prometheus-values.local.yaml" \
  && grep -q "role: pod" "${REPO_ROOT}/observability/cost/prometheus-values.local.yaml" \
  && grep -q 'regex: "8888"' "${REPO_ROOT}/observability/cost/prometheus-values.local.yaml"; then
    pass "Prometheus extraScrapeConfigs gatekeeper job present"
else
    fail "Prometheus extraScrapeConfigs gatekeeper job missing"
fi

# FR-07: collector values carry the filelog receiver, pipeline entry, and hostPath mount.
# The logs pipeline receivers list may include additional receivers (e.g.
# filelog/k8s-pods); membership of filelog/gatekeeper-audit is what matters,
# not exact list equality.
if grep -q "filelog/gatekeeper-audit" "${REPO_ROOT}/observability/otel/collector-values.local.yaml" \
  && grep "receivers: \[otlp" "${REPO_ROOT}/observability/otel/collector-values.local.yaml" | grep -q "filelog/gatekeeper-audit" \
  && grep -q "path: /var/log/pods" "${REPO_ROOT}/observability/otel/collector-values.local.yaml"; then
    pass "OTEL collector filelog gatekeeper-audit receiver present"
else
    fail "OTEL collector filelog gatekeeper-audit receiver missing"
fi

# FR-05 (cluster mode): the kubernetes API group is reachable through the proxy.
if curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://localhost:8001/apis/constraints.gatekeeper.sh/v1beta1" 2>/dev/null | grep -q '^200$'; then
    pass "kubectl proxy serves constraints.gatekeeper.sh/v1beta1 on :8001"
else
    skip "kubectl proxy on :8001 unreachable (start Backstage via scripts/start-backstage.sh)"
fi

# FR-06 (cluster mode): Prometheus has the gatekeeper job with 4 up targets.
# Live-state dependent: prefers the prometheus-server service (the API the
# check actually verifies) and degrades to SKIP when no Prometheus server is
# reachable, per the suite's SKIP semantics for cluster-dependent checks.
PROM_SVCS=$(kubectl --context "${CONTEXT}" -n opencost get svc -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | tr ' ' '\n' || true)
PROM_SVC=$(echo "${PROM_SVCS}" | grep -m1 -x 'prometheus-server' || true)
if [ -z "${PROM_SVC}" ]; then
    PROM_SVC=$(echo "${PROM_SVCS}" | grep -m1 '^prometheus' || true)
fi
if [ -z "${PROM_SVC}" ]; then
    skip "no Prometheus service in opencost namespace (run scripts/install-m4.sh first)"
else
    kubectl --context "${CONTEXT}" -n opencost port-forward "svc/${PROM_SVC}" "${LOCAL_PORT_PROM:-39090}:80" >/tmp/smoke-security-prom-pf.log 2>&1 &
    PROM_PF_PID=$!
    PROM_READY=0
    for i in $(seq 1 30); do
        if curl -fsS -o /dev/null "http://localhost:${LOCAL_PORT_PROM:-39090}/-/ready" 2>/dev/null; then
            PROM_READY=1
            break
        fi
        sleep 1
    done
    if [ "${PROM_READY}" -eq 0 ]; then
        kill "${PROM_PF_PID}" 2>/dev/null || true
        skip "Prometheus server unreachable via svc/${PROM_SVC} (pod pending or port-forward failed)"
    else
        UP_COUNT=$(curl -fsS "http://localhost:${LOCAL_PORT_PROM:-39090}/api/v1/targets?state=active" 2>/dev/null \
            | python3 -c "import json,sys; d=json.load(sys.stdin); print(sum(1 for t in d['data']['activeTargets'] if t['labels'].get('job')=='gatekeeper' and t['health']=='up'))" 2>/dev/null || echo 0)
        kill "${PROM_PF_PID}" 2>/dev/null || true
        if [ "${UP_COUNT}" = "4" ]; then
            pass "Prometheus gatekeeper job has 4 up targets"
        else
            fail "Prometheus gatekeeper job has ${UP_COUNT} up targets (expected 4)"
        fi
    fi
fi

# =============================================================================
# Lane D -- Infra/config (FR-09 TLS on .local gateways, FR-10 dependabot/CodeQL)
# =============================================================================
info "Lane D: infra/config (FR-09, FR-10)"

GATEWAY_MAIN="${REPO_ROOT}/iac/modules/networking/gateway/main.tf"
GATEWAY_TLS="${REPO_ROOT}/iac/modules/networking/gateway/tls.tf"

# D1: the three HTTPS listeners are declared on the Gateway (spec 9.3) and the
# port-80 HTTP listener stays (no redirect in Wave 0).
for listener in https-gitea https-signoz https-opencost; do
    if grep -qF "name: ${listener}" "${GATEWAY_MAIN}" 2>/dev/null; then
        pass "gateway/main.tf declares listener ${listener}"
    else
        fail "gateway/main.tf missing listener ${listener}"
    fi
done
if grep -qF "name: http" "${GATEWAY_MAIN}" 2>/dev/null; then
    pass "gateway/main.tf keeps the port-80 HTTP listener (no redirect, spec 9.3)"
else
    fail "gateway/main.tf lost the port-80 HTTP listener"
fi

# D2: tls.tf declares the issuer chain and the per-route Certificate loop
# (spec 9.3): selfsigned bootstrap -> local CA keypair -> local-ca issuer.
for res in selfsigned_bootstrap_issuer local_ca_certificate local_ca_issuer route_certificates; do
    if grep -qF "kubectl_manifest\" \"${res}\"" "${GATEWAY_TLS}" 2>/dev/null; then
        pass "gateway/tls.tf declares ${res}"
    else
        fail "gateway/tls.tf missing kubectl_manifest.${res}"
    fi
done

# D3: route Certificates are Ready on the cluster (spec 9.5). Degrades to
# SKIP when the cluster is unreachable or the module apply is still pending.
CONTEXT="${KUBECTL_CONTEXT:-k3d-openchoreo}"
if ! kubectl --context "${CONTEXT}" get ns envoy-gateway >/dev/null 2>&1; then
    skip "FR-09 certificate readiness (cluster ${CONTEXT} unreachable)"
elif ! kubectl --context "${CONTEXT}" -n envoy-gateway get certificate gitea-tls >/dev/null 2>&1; then
    skip "FR-09 certificate readiness (no Certificates in envoy-gateway yet; install-m4-networking.sh apply pending)"
elif kubectl --context "${CONTEXT}" wait --for=condition=Ready certificate -n envoy-gateway --all --timeout=120s >/dev/null 2>&1; then
    pass "all Certificates in envoy-gateway are Ready"
else
    fail "Certificates in envoy-gateway not Ready within 120s"
fi

# D4: FR-10 config files exist and parse. dependabot.yml is required;
# code-scanning.yml must NOT exist: GitHub-side CodeQL is provided by the
# org "GitHub recommended" security configuration (default setup), which
# rejects SARIF from an in-repo advanced workflow -- the two cannot
# coexist, so the file was removed 2026-08-21 (see spec section 10
# amendment). The dynamic CodeQL workflow is green on every push.
f="${REPO_ROOT}/.github/dependabot.yml"
if [ ! -f "${f}" ]; then
    fail ".github/dependabot.yml missing"
else
    yaml_ok "${f}" && rc=0 || rc=$?
    if [ "${rc}" -eq 0 ]; then
        pass ".github/dependabot.yml exists and is valid YAML"
    elif [ "${rc}" -eq 2 ]; then
        skip ".github/dependabot.yml YAML validity (no yq or python3+yaml on this host)"
    else
        fail ".github/dependabot.yml does not parse as YAML"
    fi
fi
if [ -f "${REPO_ROOT}/.github/workflows/code-scanning.yml" ]; then
    fail "code-scanning.yml present -- conflicts with the org default CodeQL setup (remove it; see spec section 10 amendment)"
else
    pass "no in-repo advanced CodeQL workflow (org default setup owns GitHub-side scanning)"
fi

# D5: dependabot.yml covers every go.mod root (enumeration per spec 10.2;
# a missing root fails the suite, spec 14.1 negative test).
missing=()
while IFS= read -r gomod; do
    dir="/$(dirname "${gomod}")"
    grep -qF "directory: ${dir}" "${REPO_ROOT}/.github/dependabot.yml" 2>/dev/null || missing+=("${dir}")
done < <(cd "${REPO_ROOT}" && find tools seed-repos plugins -name go.mod -type f | sort)
if [ "${#missing[@]}" -eq 0 ]; then
    pass "dependabot.yml covers all go.mod roots"
else
    fail "dependabot.yml missing go.mod roots: ${missing[*]}"
fi

# D6: every action reference in the mirror workflows is pinned by a full
# 40-char commit SHA (spec D4; version tags and short SHAs are rejected).
unpinned=$(grep -rhoE 'uses:[[:space:]]+[^[:space:]]+' "${REPO_ROOT}/.github/workflows/" \
    | sed -E 's/^uses:[[:space:]]+//' \
    | grep -vE '@[0-9a-f]{40}$' || true)
if [ -z "${unpinned}" ]; then
    pass "all .github/workflows action references are pinned by full commit SHA"
else
    fail "action references not full-SHA pinned: $(echo "${unpinned}" | tr '\n' ' ')"
fi

# =============================================================================
# Lane C -- Portal surfaces (FR-04 Security tab, FR-08 RBAC policy)
# =============================================================================
info "Lane C: portal surfaces (FR-04, FR-08)"

ENTITY_PAGE="${REPO_ROOT}/backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx"
BACKEND_INDEX="${REPO_ROOT}/backstage/packages/backend/src/index.ts"

# C1: Security tab registered on the entity page (spec 4.5).
if grep -qF "path: '/security'" "${ENTITY_PAGE}" \
    && grep -qF "securityContent," "${ENTITY_PAGE}"; then
    pass "Security tab registered in openchoreo-entity-page/index.tsx"
else
    fail "Security tab missing from backstage/packages/app/src/modules/openchoreo-entity-page/index.tsx"
fi

# C2: FR-04 frontend files exist (gatekeeper.ts is Lane B's, consumed by the tab).
for f in \
    backstage/packages/app/src/modules/openchoreo-cards/SecurityCard.tsx \
    backstage/packages/app/src/modules/openchoreo-cards/gatekeeper.ts; do
    if [ -f "${REPO_ROOT}/${f}" ]; then
        pass "${f} exists"
    else
        fail "${f} missing"
    fi
done

# C3: FR-08 RBAC files exist (spec 8.7).
for f in \
    backstage/packages/backend/src/extensions/permissionsPolicyExtension.ts \
    backstage/packages/backend/src/modules/permissionsPolicy.ts; do
    if [ -f "${REPO_ROOT}/${f}" ]; then
        pass "${f} exists"
    else
        fail "${f} missing"
    fi
done

# C4: allow-all policy fully removed from the backend (spec 8.7).
if grep -qF "plugin-permission-backend-module-allow-all-policy" \
    "${BACKEND_INDEX}" \
    "${REPO_ROOT}/backstage/packages/backend/package.json" 2>/dev/null; then
    fail "allow-all policy still referenced in backend index.ts or package.json"
else
    pass "allow-all policy removed from backend index.ts and package.json"
fi

# C5: RBAC module registered on the backend.
if grep -qF "import('./modules/permissionsPolicy')" "${BACKEND_INDEX}"; then
    pass "permissionsPolicy module registered in backend index.ts"
else
    fail "permissionsPolicy module not registered in backend index.ts"
fi

# C6: adminUsers configured in base and production config (spec 8.3/8.4).
if grep -qF 'adminUsers: ["${BACKSTAGE_ADMIN_USERS:-gitea_admin}"]' \
    "${REPO_ROOT}/backstage/app-config.yaml" \
    && grep -qF 'adminUsers: ["${BACKSTAGE_ADMIN_USERS:-gitea_admin}"]' \
    "${REPO_ROOT}/backstage/app-config.production.yaml"; then
    pass "permission.policy.adminUsers configured in app-config.yaml and app-config.production.yaml"
else
    fail "permission.policy.adminUsers missing from app-config.yaml or app-config.production.yaml"
fi

# C7: local dev keeps permissions disabled so the guest flow works (spec 8.5).
if grep -qF 'enabled: false' "${REPO_ROOT}/backstage/app-config.local.yaml.example"; then
    pass "app-config.local.yaml.example keeps permission.enabled: false"
else
    fail "app-config.local.yaml.example no longer carries permission.enabled: false"
fi

# =============================================================================
# Lane E -- Guard audit-log hash chaining (FR-11)
# =============================================================================
info "Lane E: guard audit-log hash chaining (FR-11)"

AUDIT_CHAIN_BIN="${REPO_ROOT}/plugins/rr-policy-guards/bin/rr-audit-chain"
GUARD_LOGS=(bash-guard brew-guard commit-guard emoji-guard tofu-guard verify-guard)

# E1: the verifier binary exists and is executable.
if [ -x "${AUDIT_CHAIN_BIN}" ]; then
    pass "rr-audit-chain binary exists and is executable"
else
    fail "rr-audit-chain missing (build: cd plugins/rr-policy-guards/tools/audit-chain && go build -o ../../bin/rr-audit-chain .)"
fi

# E2: all six audit writers emit prev_hash (commit-guard's struct tag lives
# in types.go; the other five in audit.go).
E2_MISSING=()
for g in emoji bash brew tofu verify; do
    grep -q 'json:"prev_hash"' "${REPO_ROOT}/plugins/rr-policy-guards/tools/${g}-guard/audit.go" \
        || E2_MISSING+=("${g}-guard/audit.go")
done
grep -q 'json:"prev_hash"' "${REPO_ROOT}/plugins/rr-policy-guards/tools/commit-guard/types.go" \
    || E2_MISSING+=("commit-guard/types.go")
if [ "${#E2_MISSING[@]}" -eq 0 ]; then
    pass "all six guards write the prev_hash chain field"
else
    fail "prev_hash write-side missing in: ${E2_MISSING[*]}"
fi

# E3: verify the live guard logs present on this host (spec 11.9). A guard
# with no log yet is SKIP, not FAIL. Pre-chaining legacy logs FAIL honestly
# ("missing prev_hash" at line 1) until archived aside.
if [ -x "${AUDIT_CHAIN_BIN}" ]; then
    for name in "${GUARD_LOGS[@]}"; do
        log="${HOME}/.rational-reserve/logs/${name}.jsonl"
        if [ ! -f "${log}" ]; then
            skip "${name}: no audit log yet (not a chain failure)"
            continue
        fi
        if "${AUDIT_CHAIN_BIN}" verify "${log}" >/dev/null 2>&1; then
            pass "${name}.jsonl hash chain verifies"
        else
            fail "${name}.jsonl: $("${AUDIT_CHAIN_BIN}" verify "${log}" 2>&1 | head -1)"
        fi
    done
fi

# =============================================================================
# Plane agent clientCA freshness (operational guard; 2026-08-20 incident)
# =============================================================================
# The Cluster*Plane CRs pin each agent's self-signed cert as clientCA; every
# cert-manager renewal silently breaks the gateway channel (websocket bad
# handshake, observed 2026-06-30..08-20). Fail loudly on drift so a stale pin
# is caught before it takes the planes down again.
info "Plane agent clientCA freshness"

if kubectl --context k3d-openchoreo get clusterdataplane default >/dev/null 2>&1; then
    if "${REPO_ROOT}/scripts/repin-plane-agent-ca.sh" --check >/dev/null 2>&1; then
        pass "all plane agent clientCA pins match live certs"
    else
        fail "plane agent clientCA drift (repair: scripts/repin-plane-agent-ca.sh)"
    fi
else
    skip "cluster not reachable; clientCA freshness not checked"
fi

# =============================================================================
# Scaffolder-inherited project CI (FR-38/OQ-31)
# =============================================================================
# New projects get test + security stages from the portal template. The
# live e2e (scaffold-e2e-20260821 run 1) showed two wiring bugs: Trivy ran
# before the lockfile existed (num=0), and osv-scanner v2.5.1 exits 128 on
# a valid empty lockfile. These checks pin the repaired contract.
info "Scaffolder template CI inheritance (FR-38/OQ-31)"

SCAFFOLD_WF="${REPO_ROOT}/backstage/examples/template/content/.gitea/workflows/ci.yaml"
if [ ! -f "${SCAFFOLD_WF}" ]; then
    fail "backstage/examples/template/content/.gitea/workflows/ci.yaml missing"
else
    if grep -qF "name: Unit tests (node --test)" "${SCAFFOLD_WF}" \
        && grep -qF "name: Generate lockfile for the scanner" "${SCAFFOLD_WF}" \
        && grep -qF "name: Trivy filesystem scan (gate)" "${SCAFFOLD_WF}" \
        && grep -qF "name: OSV dependency scan (gate)" "${SCAFFOLD_WF}"; then
        pass "scaffolder template CI has test + lockfile + Trivy + OSV stages"
    else
        fail "scaffolder template CI is missing a required inherited stage"
    fi

    gen_line="$(grep -n "name: Generate lockfile for the scanner" "${SCAFFOLD_WF}" | head -1 | cut -d: -f1)"
    trivy_line="$(grep -n "name: Trivy filesystem scan (gate)" "${SCAFFOLD_WF}" | head -1 | cut -d: -f1)"
    osv_line="$(grep -n "name: OSV dependency scan (gate)" "${SCAFFOLD_WF}" | head -1 | cut -d: -f1)"
    if [ -n "${gen_line}" ] && [ -n "${trivy_line}" ] && [ -n "${osv_line}" ] \
        && [ "${gen_line}" -lt "${trivy_line}" ] && [ "${trivy_line}" -lt "${osv_line}" ]; then
        pass "scaffolder template CI generates the lockfile before Trivy and OSV"
    else
        fail "scaffolder template CI step order is lockfile -> Trivy -> OSV (got gen=${gen_line} trivy=${trivy_line} osv=${osv_line})"
    fi

    if grep -qF "aquasec/trivy@${TRIVY_DIGEST}" "${SCAFFOLD_WF}" \
        && grep -qF "ghcr.io/google/osv-scanner@${OSV_DIGEST}" "${SCAFFOLD_WF}"; then
        pass "scaffolder template CI references the pinned Trivy and OSV-Scanner digests"
    else
        fail "scaffolder template CI is missing a spec-pinned scanner digest"
    fi

    if grep -qF 'rc=$?' "${SCAFFOLD_WF}" \
        && grep -qF '[ "${rc}" -eq 128 ]' "${SCAFFOLD_WF}" \
        && grep -qF '[ "${DECLARED}" = "0" ]' "${SCAFFOLD_WF}" \
        && grep -qF '[ "${LOCKED}" = "0" ]' "${SCAFFOLD_WF}"; then
        pass "scaffolder template CI accepts OSV 128 only on a verified empty tree"
    else
        fail "scaffolder template CI is missing fail-closed OSV empty-tree handling"
    fi
fi

# =============================================================================
# Loopback-only host listeners (LAN exposure)
# =============================================================================
# Backstage, Gitea port-forwards, and the other host-side portal surfaces must
# not bind 0.0.0.0 / * / [::]. Wildcard binds were observed on :7008/:7009
# during the 2026-08-21 Colima SSH tightening review.
info "Host listener scope (loopback only)"

while IFS='|' read -r kind msg; do
    case "${kind}" in
        pass) pass "${msg}" ;;
        fail) fail "${msg}" ;;
        skip) skip "${msg}" ;;
    esac
done < <(python3 <<'PY'
import re
import subprocess

ports = (3001, 7008, 7009, 3333, 3002, 29003, 3301, 9090)

try:
    out = subprocess.check_output(
        ["lsof", "-nP", "-iTCP", "-sTCP:LISTEN"],
        text=True,
        errors="replace",
    )
except (subprocess.CalledProcessError, FileNotFoundError) as exc:
    print(f"skip|lsof not available ({exc})")
    raise SystemExit(0)

wildcard_re = re.compile(r"^(?:\*|0\.0\.0\.0|\[::\])$")
loopback_re = re.compile(r"^(?:127\.0\.0\.1|\[::1\])$")
name_re = re.compile(r"TCP\s+(\S+):(\d+)\s+\(LISTEN\)")

by_port = {p: [] for p in ports}
for line in out.splitlines():
    m = name_re.search(line)
    if not m:
        continue
    addr, port_s = m.group(1), m.group(2)
    port = int(port_s)
    if port in by_port:
        by_port[port].append(addr)

for port in ports:
    addrs = by_port[port]
    if not addrs:
        print(f"skip|port {port} has no host listener")
        continue
    wild = [a for a in addrs if wildcard_re.match(a)]
    if wild:
        print(f"fail|port {port} has a wildcard listener ({', '.join(wild)})")
        continue
    bad = [a for a in addrs if not loopback_re.match(a)]
    if bad:
        print(f"fail|port {port} listens on non-loopback address ({', '.join(bad)})")
        continue
    uniq = ", ".join(sorted(set(addrs)))
    print(f"pass|port {port} is loopback-only ({uniq})")
PY
)

# =============================================================================
# Summary
# =============================================================================
echo
info "Summary: ${PASSED} passed, ${FAILED} failed, ${SKIPPED} skipped"
if [ "${FAILED}" -ne 0 ]; then
    exit 1
fi
info "Security smoke suite passed (${SKIPPED} skipped)."
