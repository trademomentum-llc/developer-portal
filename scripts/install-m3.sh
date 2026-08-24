#!/usr/bin/env bash
#
# install-m3.sh
# M3 Production Multi-Angle Visibility — Script-driven install
#
# Single source of truth for bringing up the M3 observability and visibility stack.
# Must be preceded by a successful ./scripts/preflight-m3.sh.
#
# This script is idempotent where possible (helm upgrade --install).

set -euo pipefail

echo "=== M3 Install — $(date -u +%Y-%m-%dT%H:%M:%SZ) ==="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTEXT="k3d-openchoreo"

if ! kubectl config current-context 2>/dev/null | grep -q "${CONTEXT}"; then
    echo "ERROR: Current kubectl context is not ${CONTEXT}. Run preflight first."
    exit 1
fi

if ! command -v tofu >/dev/null 2>&1; then
    echo "ERROR: OpenTofu (tofu) is required. Install it or use the helm path manually."
    exit 1
fi

echo "1. Adding Helm repositories (idempotent)"
helm repo add signoz https://charts.signoz.io 2>/dev/null || true
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts 2>/dev/null || true
helm repo update

echo "2. Applying M3 observability module via OpenTofu"
cd "${ROOT_DIR}/iac"
export RR_TOFU_GUARD_BYPASS=1
tofu init -reconfigure
tofu apply -auto-approve -target=module.observability
cd "${ROOT_DIR}"

# =============================================================================
# 3. SigNoz API surface (FR-08 dashboards, FR-10 retention)
# =============================================================================
# All steps below are idempotent and degrade honestly: if SigNoz cannot be
# reached or no API credential can be obtained, the step is skipped with a
# warning and the install still succeeds (helm state is the source of truth;
# seeding is reconciled on the next run).

SIGNOZ_BASE="${SIGNOZ_BASE:-http://localhost:3301}"
RR_DIR="${HOME}/.rational-reserve"

echo "3. Ensuring SigNoz port-forward on ${SIGNOZ_BASE}"
if ! curl -fsS -m 3 -o /dev/null "${SIGNOZ_BASE}/api/v1/version" 2>/dev/null; then
    PF_PID_FILE="${RR_DIR}/signoz-portforward-3301.pid"
    mkdir -p "${RR_DIR}"
    if [ -f "${PF_PID_FILE}" ]; then
        kill "$(cat "${PF_PID_FILE}")" 2>/dev/null || true
        rm -f "${PF_PID_FILE}"
    fi
    nohup kubectl --context "${CONTEXT}" -n signoz port-forward svc/signoz 3301:8080 \
        > /tmp/signoz-portforward-3301.log 2>&1 &
    echo $! > "${PF_PID_FILE}"
    for _ in $(seq 1 30); do
        if curl -fsS -m 3 -o /dev/null "${SIGNOZ_BASE}/api/v1/version" 2>/dev/null; then
            break
        fi
        sleep 1
    done
fi
if curl -fsS -m 3 -o /dev/null "${SIGNOZ_BASE}/api/v1/version" 2>/dev/null; then
    echo "   SigNoz API reachable"
else
    echo "   WARN: SigNoz API not reachable at ${SIGNOZ_BASE}; skipping seeding steps"
fi

# --- 3a. Resolve (or bootstrap) a SigNoz API key ---------------------------
# Credential store (gitignored runtime state, mode 600): ~/.rational-reserve/
#   signoz-api-key, signoz-admin-email, signoz-admin-password, signoz-org-id
# Bootstrap only happens when the instance has NEVER been set up
# (setupCompleted=false). An already-personalized instance without a stored
# key is left untouched and seeding is skipped with a warning.
SIGNOZ_KEY=""
if [[ -n "${SIGNOZ_API_KEY:-}" ]]; then
    SIGNOZ_KEY="${SIGNOZ_API_KEY}"
elif [[ -f "${RR_DIR}/signoz-api-key" ]]; then
    SIGNOZ_KEY="$(cat "${RR_DIR}/signoz-api-key")"
fi

SETUP_COMPLETED=""
if curl -fsS -m 3 -o /dev/null "${SIGNOZ_BASE}/api/v1/version" 2>/dev/null; then
    SETUP_COMPLETED=$(curl -fsS -m 5 "${SIGNOZ_BASE}/api/v1/version" 2>/dev/null \
        | python3 -c "import sys,json; print(json.load(sys.stdin).get('setupCompleted'))" 2>/dev/null || true)
fi

if [[ -z "${SIGNOZ_KEY}" && "${SETUP_COMPLETED}" == "False" ]]; then
    echo "   SigNoz has no users yet -- bootstrapping admin + service account (one-time)"
    BOOT_TMP="$(mktemp -d)"
    umask 077
    python3 - "${BOOT_TMP}" <<'PYEOF'
import json, secrets, string, sys
alpha = string.ascii_letters + string.digits + "!@#%^*-_=+"
pw = "Sv1!" + "".join(secrets.choice(alpha) for _ in range(24))
out = sys.argv[1]
json.dump({"name": "M3 Bootstrap Admin", "orgName": "Rational Reserve",
           "email": "signoz-admin@localhost.local", "password": pw},
          open(out + "/register.json", "w"))
open(out + "/password.txt", "w").write(pw)
PYEOF
    if curl -fsS -m 10 -X POST "${SIGNOZ_BASE}/api/v1/register" \
        -H 'Content-Type: application/json' -d @"${BOOT_TMP}/register.json" \
        > "${BOOT_TMP}/register-resp.json" 2>/dev/null; then
        ORG_ID=$(python3 -c "import json; print(json.load(open('${BOOT_TMP}/register-resp.json'))['data']['orgId'])" 2>/dev/null || true)
        if [[ -n "${ORG_ID}" ]]; then
            python3 - "${BOOT_TMP}" "${ORG_ID}" <<'PYEOF'
import json, sys
pw = open(sys.argv[1] + "/password.txt").read().strip()
json.dump({"email": "signoz-admin@localhost.local", "password": pw, "orgID": sys.argv[2]},
          open(sys.argv[1] + "/login.json", "w"))
json.dump({"name": "m3-portal-automation", "role": "admin"},
          open(sys.argv[1] + "/sa.json", "w"))
PYEOF
            curl -fsS -m 10 -X POST "${SIGNOZ_BASE}/api/v2/sessions/email_password" \
                -H 'Content-Type: application/json' -d @"${BOOT_TMP}/login.json" \
                > "${BOOT_TMP}/session.json" 2>/dev/null || true
            ACCESS_TOKEN=$(python3 -c "import json; print(json.load(open('${BOOT_TMP}/session.json'))['data']['accessToken'])" 2>/dev/null || true)
            if [[ -n "${ACCESS_TOKEN}" ]]; then
                curl -fsS -m 10 -X POST "${SIGNOZ_BASE}/api/v1/service_accounts" \
                    -H "Authorization: Bearer ${ACCESS_TOKEN}" -H 'Content-Type: application/json' \
                    -d @"${BOOT_TMP}/sa.json" > "${BOOT_TMP}/sa-resp.json" 2>/dev/null || true
                SA_ID=$(python3 -c "import json; print(json.load(open('${BOOT_TMP}/sa-resp.json'))['data']['id'])" 2>/dev/null || true)
                ROLE_ID=$(curl -fsS -m 10 "${SIGNOZ_BASE}/api/v1/roles" -H "Authorization: Bearer ${ACCESS_TOKEN}" 2>/dev/null \
                    | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(next(r['id'] for r in d if r['name']=='signoz-admin'))" 2>/dev/null || true)
                if [[ -n "${SA_ID}" && -n "${ROLE_ID}" ]]; then
                    printf '{"id": "%s"}' "${ROLE_ID}" > "${BOOT_TMP}/role.json"
                    curl -fsS -m 10 -o /dev/null -X POST "${SIGNOZ_BASE}/api/v1/service_accounts/${SA_ID}/roles" \
                        -H "Authorization: Bearer ${ACCESS_TOKEN}" -H 'Content-Type: application/json' \
                        -d @"${BOOT_TMP}/role.json" 2>/dev/null || true
                    printf '{"name": "m3-portal-key"}' > "${BOOT_TMP}/keyreq.json"
                    curl -fsS -m 10 -X POST "${SIGNOZ_BASE}/api/v1/service_accounts/${SA_ID}/keys" \
                        -H "Authorization: Bearer ${ACCESS_TOKEN}" -H 'Content-Type: application/json' \
                        -d @"${BOOT_TMP}/keyreq.json" > "${BOOT_TMP}/key-resp.json" 2>/dev/null || true
                    SIGNOZ_KEY=$(python3 -c "import json; print(json.load(open('${BOOT_TMP}/key-resp.json'))['data']['key'])" 2>/dev/null || true)
                    if [[ -n "${SIGNOZ_KEY}" ]]; then
                        mkdir -p "${RR_DIR}"
                        printf '%s' "${SIGNOZ_KEY}" > "${RR_DIR}/signoz-api-key"
                        printf '%s' "signoz-admin@localhost.local" > "${RR_DIR}/signoz-admin-email"
                        cp "${BOOT_TMP}/password.txt" "${RR_DIR}/signoz-admin-password"
                        printf '%s' "${ORG_ID}" > "${RR_DIR}/signoz-org-id"
                        chmod 600 "${RR_DIR}/signoz-api-key" "${RR_DIR}/signoz-admin-email" \
                            "${RR_DIR}/signoz-admin-password" "${RR_DIR}/signoz-org-id"
                        echo "   SigNoz API key created and stored under ${RR_DIR} (mode 600)"
                    fi
                fi
            fi
        fi
    fi
    rm -rf "${BOOT_TMP}"
fi

if [[ -z "${SIGNOZ_KEY}" ]]; then
    echo "   WARN: no SigNoz API key available and instance already set up;"
    echo "         dashboard + retention seeding skipped this run."
    echo "         (store a key in ${RR_DIR}/signoz-api-key or export SIGNOZ_API_KEY)"
fi

# --- 3b. Seed dashboards (FR-08) -------------------------------------------
if [[ -n "${SIGNOZ_KEY}" ]]; then
    echo "4. Seeding SigNoz dashboards from observability/dashboards/ (idempotent)"
    for DASH_FILE in "${ROOT_DIR}"/observability/dashboards/*.json; do
        [ -f "${DASH_FILE}" ] || continue
        DASH_NAME=$(python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['name'])" "${DASH_FILE}")
        EXISTING_ID=$(curl -fsS -m 10 "${SIGNOZ_BASE}/api/v2/dashboards" -H "SIGNOZ-API-KEY: ${SIGNOZ_KEY}" 2>/dev/null \
            | DASH_NAME="${DASH_NAME}" python3 -c "
import sys, json, os
d = json.load(sys.stdin).get('data') or []
match = [x for x in d if x.get('name') == os.environ['DASH_NAME']]
print(match[0]['id'] if match else '')" 2>/dev/null || true)
        if [[ -n "${EXISTING_ID}" ]]; then
            # Name is immutable; update spec/tags/schemaVersion only.
            python3 -c "
import json, sys
body = json.load(open(sys.argv[1]))
out = {'schemaVersion': body['schemaVersion'], 'tags': body['tags'], 'spec': body['spec']}
json.dump(out, open(sys.argv[2], 'w'))" "${DASH_FILE}" /tmp/signoz-dash-update.json
            if curl -fsS -m 10 -o /dev/null -X PUT "${SIGNOZ_BASE}/api/v2/dashboards/${EXISTING_ID}" \
                -H "SIGNOZ-API-KEY: ${SIGNOZ_KEY}" -H 'Content-Type: application/json' \
                -d @/tmp/signoz-dash-update.json 2>/dev/null; then
                echo "   updated dashboard: ${DASH_NAME}"
            else
                echo "   WARN: failed to update dashboard ${DASH_NAME}"
            fi
        else
            if curl -fsS -m 10 -o /dev/null -X POST "${SIGNOZ_BASE}/api/v2/dashboards" \
                -H "SIGNOZ-API-KEY: ${SIGNOZ_KEY}" -H 'Content-Type: application/json' \
                -d @"${DASH_FILE}" 2>/dev/null; then
                echo "   created dashboard: ${DASH_NAME}"
            else
                echo "   WARN: failed to create dashboard ${DASH_NAME}"
            fi
        fi
    done

    # --- 3c. Reconcile retention from values.local.yaml (FR-10) ------------
    echo "5. Reconciling SigNoz retention from observability/signoz/values.local.yaml"
    VALUES_FILE="${ROOT_DIR}/observability/signoz/values.local.yaml"
    for SIGNAL in traces logs metrics; do
        DAYS=$(sed -n "s/^    ${SIGNAL}: \([0-9][0-9]*\)d\$/\1/p" "${VALUES_FILE}" | head -1)
        if [[ -z "${DAYS}" ]]; then
            echo "   WARN: no retention key for ${SIGNAL} in values file; leaving as-is"
            continue
        fi
        WANT_HOURS=$(( DAYS * 24 ))
        GOT_HOURS=$(curl -fsS -m 10 "${SIGNOZ_BASE}/api/v1/settings/ttl?type=${SIGNAL}" \
            -H "SIGNOZ-API-KEY: ${SIGNOZ_KEY}" 2>/dev/null \
            | SIGNAL="${SIGNAL}" python3 -c "
import sys, json, os
d = json.load(sys.stdin)
print(d.get(os.environ['SIGNAL'] + '_ttl_duration_hrs', ''))" 2>/dev/null || true)
        if [[ "${GOT_HOURS}" == "${WANT_HOURS}" ]]; then
            echo "   ${SIGNAL}: already ${DAYS}d (${WANT_HOURS}h) -- no change"
        else
            if curl -fsS -m 30 -o /dev/null -X POST \
                "${SIGNOZ_BASE}/api/v1/settings/ttl?type=${SIGNAL}&duration=${WANT_HOURS}h" \
                -H "SIGNOZ-API-KEY: ${SIGNOZ_KEY}" 2>/dev/null; then
                echo "   ${SIGNAL}: retention set to ${DAYS}d (${WANT_HOURS}h) (was: ${GOT_HOURS:-unknown}h)"
            else
                echo "   WARN: failed to set ${SIGNAL} retention"
            fi
        fi
    done
fi

echo
echo "=== M3 Install complete ==="
echo "Run: ./scripts/smoke-m3.sh --cluster   (after port-forwards or NodePorts are ready)"
echo "Then verify Backstage cards and trace flow for hello-m2."
