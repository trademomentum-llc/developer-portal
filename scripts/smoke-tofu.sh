#!/usr/bin/env bash
# scripts/smoke-tofu.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
source "${ROOT}/scripts/lib/smoke-json.sh"
smoke_json_parse_args "$@"
smoke_json_begin tofu

if tofu version >/dev/null; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
cd "$(dirname "$0")/../iac"
if tofu init -reconfigure -input=false >/dev/null; then
    smoke_json_count pass
else
    smoke_json_count fail
    exit 1
fi
set +e
tofu plan -detailed-exitcode >/dev/null
code=$?
set -e
case $code in
  0) echo "PASS: no diff"; smoke_json_count pass ;;
  2) echo "PASS: diff present"; smoke_json_count pass ;;
  *) smoke_json_count fail; exit "$code" ;;
esac
