#!/usr/bin/env bash
# scripts/smoke-gatekeeper.sh
set -e
kubectl get constrainttemplates | awk '{print tolower($1)}' | grep -q c1platformaddonsmainprotected
kubectl get constrainttemplates | awk '{print tolower($1)}' | grep -q c2scoreschemavalid
kubectl get constrainttemplates | awk '{print tolower($1)}' | grep -q c3infracostdelta
echo "PASS"
