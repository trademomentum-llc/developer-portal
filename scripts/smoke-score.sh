#!/usr/bin/env bash
# scripts/smoke-score.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
out=$("$ROOT/tools/score2openchoreo/bin/score2openchoreo" \
    --input "$ROOT/tools/score2openchoreo/fixtures/minimal.score.yaml" \
    --environment dev)
echo "$out" | yq eval 'select(document_index == 0) | .kind' - | grep -q Component
echo "$out" | yq eval 'select(document_index == 1) | .kind' - | grep -q Workload
echo "PASS"
