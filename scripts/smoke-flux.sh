# scripts/smoke-flux.sh
#!/usr/bin/env bash
set -e
flux reconcile source git platform-addons >/dev/null
flux get kustomizations platform-addons | grep -q True
echo "PASS"
