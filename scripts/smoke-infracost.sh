# scripts/smoke-infracost.sh
#!/usr/bin/env bash
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
infracost breakdown --path "$ROOT/iac" --format table >/dev/null
echo "PASS"
