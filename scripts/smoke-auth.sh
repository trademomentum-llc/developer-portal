#!/usr/bin/env bash
#
# smoke-auth.sh
# Validates that the Backstage Gitea authentication provider is wired and the
# OAuth2 start endpoint redirects to the local Gitea authorize URL.
#
# Usage:
#   ./scripts/smoke-auth.sh
#
# Expects Backstage to be running (scripts/start-backstage.sh) and Gitea to be
# reachable at localhost:3333.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKSTAGE_BACKEND="${BACKSTAGE_BACKEND:-http://localhost:7008}"
GITEA_URL="${GITEA_URL:-http://localhost:3333}"

BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

log() { echo -e "$*"; }
pass() { echo -e "${GREEN} PASS${RESET} $*"; }
fail() { echo -e "${RED} FAIL${RESET} $*"; exit 1; }
info() { echo -e "${BOLD}INFO${RESET} $*"; }

info "Validating Backstage Gitea auth provider"

if ! command -v curl >/dev/null 2>&1; then
    fail "curl is required"
fi

info "Checking Backstage backend at ${BACKSTAGE_BACKEND}"
if ! curl -fsS "${BACKSTAGE_BACKEND}/" >/dev/null 2>&1; then
    fail "Backstage backend is not reachable at ${BACKSTAGE_BACKEND}"
fi
pass "Backstage backend is reachable"

info "Ensuring Gitea OAuth application exists"
"${ROOT_DIR}/scripts/setup-gitea-oauth.sh" >/dev/null 2>&1 || true

info "Checking Gitea auth start endpoint"
START_URL="${BACKSTAGE_BACKEND}/api/auth/gitea/start?env=development"
LOCATION=$(curl -fsS -o /dev/null -w "%{redirect_url}" "${START_URL}" 2>/dev/null || true)

if [[ -z "${LOCATION}" ]]; then
    # Some curl versions report the final URL after following redirects; retry
    # with --head and grep the Location header to be safe.
    LOCATION=$(curl -IsS "${START_URL}" 2>/dev/null | awk '/^[Ll]ocation:/ {print $2}' | tr -d '\r' || true)
fi

if [[ -z "${LOCATION}" ]]; then
    fail "Auth start endpoint did not return a redirect"
fi

EXPECTED_PREFIX="${GITEA_URL}/login/oauth/authorize"
if [[ "${LOCATION}" != "${EXPECTED_PREFIX}"* ]]; then
    fail "Unexpected redirect location: ${LOCATION} (expected ${EXPECTED_PREFIX}*)"
fi
pass "Auth start redirects to Gitea OAuth authorize URL"

info "Checking Gitea is reachable"
if ! curl -fsS "${GITEA_URL}/api/v1/version" >/dev/null 2>&1; then
    fail "Gitea is not reachable at ${GITEA_URL}"
fi
pass "Gitea is reachable"

log ""
log "All Gitea auth smoke checks passed."
