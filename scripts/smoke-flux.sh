#!/usr/bin/env bash
# scripts/smoke-flux.sh
set -e
flux reconcile source git platform-addons >/dev/null
flux get kustomizations platform-addons | grep -q True
echo "PASS"
