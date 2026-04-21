#!/usr/bin/env bash
# scripts/smoke-m2.sh
set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for check in tofu actions flux score infracost gatekeeper openbao; do
    printf "[%s] " "$check"
    "$ROOT/scripts/smoke-$check.sh" || { echo "FAIL"; exit 1; }
done
echo "M2 smoke: all pass"
