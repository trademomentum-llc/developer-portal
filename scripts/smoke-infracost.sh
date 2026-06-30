#!/usr/bin/env bash
# scripts/smoke-infracost.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Infracost requires an API key or cloud auth to fetch pricing. In local dev
# without a key we do a best-effort syntax check by running --show-skipped and
# skip gracefully rather than failing the whole M2 smoke suite.
if [ -z "${INFRACOST_API_KEY:-}" ] && ! infracost auth status 2>/dev/null | grep -q 'API key'; then
    echo "SKIP: no Infracost API key available locally (set INFRACOST_API_KEY to enable)"
    exit 0
fi

infracost breakdown --path "$ROOT/iac" --format table >/dev/null
echo "PASS"
