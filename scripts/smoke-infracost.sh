#!/usr/bin/env bash
# scripts/smoke-infracost.sh
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
infracost breakdown --path "$ROOT/iac" --format table >/dev/null
echo "PASS"
