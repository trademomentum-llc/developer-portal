#!/usr/bin/env bash
# scripts/smoke-score.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin score

out=$("$ROOT/tools/score2openchoreo/bin/score2openchoreo" \
    --input "$ROOT/tools/score2openchoreo/fixtures/minimal.score.yaml" \
    --environment dev)
if echo "$out" | yq eval 'select(document_index == 0) | .kind' - | grep -q Component; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
if echo "$out" | yq eval 'select(document_index == 1) | .kind' - | grep -q Workload; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
echo "PASS"
