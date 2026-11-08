# Runbook: Promote a component from dev to staging

> FR-20 / OQ-14: promotion stays a manual commit ("human approval is a
> feature, not a bug", docs/specs/m2-iac-cd/design-specification.md). There
> is no in-portal promote button; this runbook is the procedure.
>
> Scope: the local k3d-openchoreo platform. `<app>` below is the component
> name as it appears in `platform-config/environments/dev/<app>.yaml`
> (example: `hello-m2`). All commands are read-safe except the single
> commit in step 3 and the rollback in step 6.

## How the mechanism works

1. CI renders the OpenChoreo resources (Component + SecretReference +
   Workload) from `score.yaml` with `score2openchoreo --environment
   development` and commits them to `platform-config` as
   `environments/dev/<app>.yaml` (scripts/ci/commit-to-platform-config.sh).
2. Two Flux kustomizations in `flux-system` watch the same repo at different
   paths, both with `prune: true`:
   - `platform-config-dev`    -> `./environments/dev`
   - `platform-config-staging` -> `./environments/staging`
3. Promotion = making the rendered resources exist under
   `environments/staging/` on `main`. Flux `platform-config-staging`
   applies them; OpenChoreo reconciles a ReleaseBinding for the staging
   environment and schedules pods into the predicted staging namespace.

## 1. Preconditions

- `kubectl --context k3d-openchoreo` works.
- Local Gitea reachable: `curl -s http://localhost:3333/api/v1/version`
  (port-forward is managed by scripts/start-backstage.sh).
- The dev deployment is healthy:

  ```
  kubectl --context k3d-openchoreo -n default get releasebinding
  kubectl --context k3d-openchoreo get kustomization -n flux-system platform-config-staging
  ```

  Expected today: `hello-m2-development` with `Ready=True`, and the
  staging kustomization `READY=True`.

## 2. Identify what is deployed in dev

Read the image tag and git sha from the committed dev file -- promote the
artifact that CI actually built, never a rebuilt one:

```
curl -s "http://localhost:3333/openchoreo/platform-config/raw/branch/main/environments/dev/<app>.yaml"
```

Note `spec.container.image` (e.g.
`registry.local-registry.svc.cluster.local:5000/hello-m2:<sha>`) and the
`GIT_SHA` env value.

## 3. Create the staging file and commit

Preferred, environment-accurate path: re-render the SAME image for the
staging environment from the app repo (score.yaml at the deployed sha),
then commit the render as `environments/staging/<app>.yaml`:

```
cd "<app-checkout>" && git checkout "<deployed-sha>"
EXPECTED_NS=$(go run /path/to/developer-portal/tools/namespace-predictor/main.go default default staging)
/path/to/score2openchoreo \
  --input score.yaml \
  --environment staging \
  --image "registry.local-registry.svc.cluster.local:5000/<app>:<deployed-sha-short>" \
  --extra-env OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector-opentelemetry-collector.otel-system.svc.cluster.local:4318 \
  --extra-env OPENCHOREO_PROJECT=default \
  --extra-env OPENCHOREO_COMPONENT="<app>" \
  --extra-env OPENCHOREO_ENVIRONMENT=staging \
  --extra-env OPENCHOREO_RUNTIME_NAMESPACE="${EXPECTED_NS}" \
  --extra-env GIT_SHA="<deployed-sha-short>" \
  > /tmp/component-staging.yaml
```

Commit it via a local clone of platform-config (or the Gitea contents API,
same shape as scripts/ci/commit-to-platform-config.sh with `staging` as the
environment argument):

```
git -C platform-config cp environments/dev/<app>.yaml environments/staging/<app>.yaml   # only for the verbatim-copy fallback below
# preferred: cp /tmp/component-staging.yaml platform-config/environments/staging/<app>.yaml
git -C platform-config add environments/staging/<app>.yaml
git -C platform-config commit -m "promote: <app> <sha> dev -> staging"
git -C platform-config push origin main
```

Verbatim-copy fallback (acceptable for a demo, with caveats): copy
`environments/dev/<app>.yaml` to `environments/staging/<app>.yaml`
unchanged. Caveats you accept by doing so:

- The Workload env values (`ENVIRONMENT`, `OPENCHOREO_ENVIRONMENT`,
  `OPENCHOREO_RUNTIME_NAMESPACE`) still say `development`; re-render
  instead if those values must be accurate.
- The SecretReference remoteRef still points at
  `apps/<app>/development/...` in OpenBao; staging secrets must exist
  under that key or be re-rendered.
- The two kustomizations manage overlapping objects while the files are
  identical; do not delete the dev file without removing the staging file
  first (both kustomizations run `prune: true`).

## 4. Verify the promotion

```
kubectl --context k3d-openchoreo -n flux-system get kustomization platform-config-staging
kubectl --context k3d-openchoreo -n flux-system reconcile kustomization platform-config-staging   # optional: skip the poll interval
kubectl --context k3d-openchoreo -n default get releasebinding
go run tools/namespace-predictor/main.go default default staging
kubectl --context k3d-openchoreo get pods -n "<predicted-staging-ns>"
```

Expected: the staging kustomization reports the new revision applied; a
ReleaseBinding `<app>-staging` appears and reaches `Ready=True`; pods land
Running in the predicted staging namespace (for hello-m2 that namespace
predicts to `dp-default-default-staging-1362f732`).

Portal cross-check: the component's Deployment tab (Backstage) shows the
predicted namespace and -- once the OpenChoreo API proxy carries a token --
the observed ReleaseBinding block; the kubernetes workload section lists
the staging pods by their `openchoreo.dev/component=<app>` label.

## 5. If staging does not reconcile

- `kubectl --context k3d-openchoreo -n flux-system describe kustomization platform-config-staging`
  -- look for validation errors in the rendered YAML.
- `kubectl --context k3d-openchoreo -n default describe releasebinding "<app>-staging"`
  -- OpenChoreo condition reasons.
- OpenBao: confirm the secret keys referenced by the SecretReference exist
  for the staging environment.

## 6. Rollback

Rollback is a git operation; Flux prunes the staging resources on the next
sync:

```
git -C platform-config revert --no-edit "<promotion-commit>"
git -C platform-config push origin main
kubectl --context k3d-openchoreo -n flux-system reconcile kustomization platform-config-staging
kubectl --context k3d-openchoreo get pods -n "<predicted-staging-ns>"   # expect: terminating / gone
```

## Verification status of this runbook (2026-08-21)

Per the Phase-2/3 closure spec, the demo workload `hello-m2` stays
dev-only. This runbook was verified by dry-reading every command against
the live cluster: the staging kustomization exists and is Ready, the dev
file and its image tag resolve, the namespace predictor output above is
the real algorithm output, and the rollback path is the same
commit-and-reconcile loop exercised by CI. The copy+commit itself is
deliberately not executed here.
