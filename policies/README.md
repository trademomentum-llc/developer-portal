# M2 Gatekeeper Policies

Three pipeline-scoped constraints for M2. Broader runtime policies stay M6.

| Constraint | Purpose |
|---|---|
| C-1 | Flux GitRepository for platform-addons must reference main |
| C-2 | OpenChoreo Components must carry pipeline.m2/score-valid=true |
| C-3 | OpenChoreo Components must have cost-delta annotation under threshold |

## Testing

```
opa test policies/
```
