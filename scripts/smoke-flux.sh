#!/usr/bin/env bash
# scripts/smoke-flux.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin flux

if flux reconcile source git platform-addons >/dev/null; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
if flux get kustomizations platform-addons | grep -q True; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
echo "PASS"
