# Technical Specification: M4 Actual Cluster Cost Visibility

**Document ID:** M4-COST-VISIBILITY-TECH-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-M4-Cost-Visibility-Design-Specification.md

---

## 1. Module Layout

```
iac/modules/cost/
  main.tf
  variables.tf
  README.md

observability/cost/
  prometheus-values.local.yaml
  opencost-values.local.yaml

scripts/
  install-m4.sh
  teardown-m4.sh
  smoke-m4.sh
```

---

## 2. OpenTofu Resources

### 2.1 `iac/modules/cost/main.tf`

- `helm_release.prometheus` from `https://prometheus-community.github.io/helm-charts`, chart `prometheus`, version pinned.
- `helm_release.opencost` from `https://opencost.github.io/opencost-helm-chart`, chart `opencost`, version pinned.
- Values files loaded from `observability/cost/` relative to repo root.

### 2.2 Sizing for Local Cluster

`prometheus-values.local.yaml`:

```yaml
server:
  persistentVolume:
    enabled: false
  retention: "2d"
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
alertmanager:
  enabled: false
pushgateway:
  enabled: false
```

`opencost-values.local.yaml`:

```yaml
opencost:
  prometheus:
    internal:
      enabled: true
      namespaceName: opencost
      serviceName: prometheus-server
      port: 80
  ui:
    enabled: true
  resources:
    requests:
      cpu: 50m
      memory: 128Mi
```

---

## 3. Scripts

### `scripts/install-m4.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../iac"
tofu init -reconfigure
tofu apply -target=module.cost -auto-approve
```

### `scripts/teardown-m4.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../iac"
tofu destroy -target=module.cost -auto-approve
```

### `scripts/smoke-m4.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
kubectl --context k3d-openchoreo -n opencost port-forward svc/opencost 29003:9090 &
PF=$!
sleep 3
curl -fsS "http://localhost:29003/model/allocation?window=today&aggregate=namespace" >/tmp/smoke-m4-cost.json
kill $PF
```

---

## 4. Backstage CostCard Update

Add a helper in the CostCard that builds the OpenCost UI URL:

```ts
const openCostUrl = `http://localhost:29003/opencost/ui/?namespace=${predictedNs}`;
```

Render it as a link button below the existing Infracost and artifact links.

---

## 5. Version Pins

| Component | Version | Source |
|---|---|---|
| Prometheus chart | TBD (latest stable at implementation time) | prometheus-community |
| OpenCost chart | TBD (latest stable at implementation time) | opencost |

Pins will be recorded in `iac/modules/cost/variables.tf` and `observability/cost/values.local.yaml`.

---

## 6. Acceptance

- `tofu plan -target=module.cost` is clean.
- `scripts/install-m4.sh` completes without errors.
- `scripts/smoke-m4.sh` returns valid allocation JSON.
- `scripts/smoke-m3.sh` still passes 22/22 after install.

---

**End of Technical Specification**
