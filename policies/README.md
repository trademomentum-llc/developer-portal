# M2 Gatekeeper Policies

Three pipeline-scoped constraints for M2. Broader runtime policies stay M6.

| Constraint | Purpose |
|---|---|
| C-1 | Flux GitRepository for platform-addons must reference main |
| C-2 | OpenChoreo Components must carry pipeline.m2/score-valid=true |
| C-3 | OpenChoreo Components must have cost-delta annotation under threshold |

## Testing

Run from the repository root:

```
opa test --v0-compatible policies/*.rego -v
```

The `--v0-compatible` flag is required for OPA 1.x because these policies
use Rego v0 syntax (the OPA-native dialect Gatekeeper 3.17 still expects).
Scope the glob to `*.rego` only -- the constraint YAMLs are Kubernetes
manifests and fail to load when OPA tries to treat them as data bundles.

Expected output: 6/6 PASS (two test cases per constraint: allow and deny).
