#!/usr/bin/env bash
# scripts/smoke-gatekeeper.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin gatekeeper

for tpl in c1platformaddonsmainprotected c2scoreschemavalid c3infracostdelta; do
    if kubectl get constrainttemplates | awk '{print tolower($1)}' | grep -q "${tpl}"; then
        smoke_json_count pass
    else
        smoke_json_count fail
        exit 1
    fi
done
echo "PASS"
