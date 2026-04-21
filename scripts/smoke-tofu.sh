# scripts/smoke-tofu.sh
#!/usr/bin/env bash
set -e
tofu version >/dev/null
cd "$(dirname "$0")/../iac"
tofu init -reconfigure -input=false >/dev/null
set +e
tofu plan -detailed-exitcode >/dev/null
code=$?
set -e
case $code in
  0) echo "PASS: no diff" ;;
  2) echo "PASS: diff present" ;;
  *) exit $code ;;
esac
