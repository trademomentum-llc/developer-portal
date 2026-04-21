# scripts/smoke-score.sh
#!/usr/bin/env bash
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
out=$("$ROOT/tools/score2openchoreo/bin/score2openchoreo" \
    --input "$ROOT/tools/score2openchoreo/fixtures/minimal.score.yaml" \
    --environment dev --namespace openchoreo-data-plane --project openchoreo)
echo "$out" | yq eval '.kind' - | grep -q Component
echo "PASS"
