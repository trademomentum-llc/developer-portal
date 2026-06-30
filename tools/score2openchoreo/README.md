# score2openchoreo

Converts a [Score](https://score.dev) YAML document to OpenChoreo
`openchoreo.dev/v1alpha1` resources. The renderer emits a `Component`, any
needed `SecretReference` resources, and a matching `Workload` as a
multi-document YAML stream. Used by the M2 CI pipeline after schema validation,
before committing the rendered resources to
`openchoreo/platform-config/environments/<env>/`.

## Full spec

The authoritative specs live in `docs/specs/m2-iac-cd/`. This README covers
only the conversion conventions that are not self-evident from the code.

## Build

```
cd tools/score2openchoreo
go test ./...
go build -o bin/score2openchoreo .
```

## Conventions a Score author needs to know

### 1. Resource references must be the entire value

A Score variable value is either:

- a plain literal (copied verbatim to `env[].value`), or
- the string `${resources.<key>.<field>}` exactly -- no prefix, no suffix,
  no other text.

Inline substitution is NOT supported. A value like
`prefix-${resources.db.password}` is rejected with an error naming the
offending variable. This is deliberate: silent passthrough of unresolved
references has shipped a literal `${resources.X.Y}` into production env
vars in other converters, and deterministic-first is a project rule. If
you need composition, build the full string in a pre-step or change the
resource shape.

### 2. Secret name defaults to `<resource-key>-secret`

For a resource of `type: secret`, the Kubernetes Secret name defaults to
the resource's own key suffixed with `-secret`. Example:

```yaml
resources:
  db:
    type: secret
```

...produces `SecretKeyRef{name: "db-secret", key: <field>}` and a
`SecretReference` named `db-secret`. Override by setting
`resources.db.metadata.name`:

```yaml
resources:
  db:
    type: secret
    metadata:
      name: my-db-credentials   # -> SecretKeyRef{name: "my-db-credentials"}
```

This convention pairs with OpenChoreo's `SecretReference` flow. The default
remote key is `apps/<component>/<environment>/<secret-name>` and the default
remote property is the referenced field. Override the remote key with
`resources.db.metadata.remoteRef.key`; override the remote property with
`resources.db.metadata.remoteRef.property`.

### 3. Supported resource types

Only `secret` and `environment` resource types are handled. Any other
`type:` value fails Convert with an explicit error.

### 4. Component type inference

By default, a Score file with `service.ports` renders an OpenChoreo
`deployment/service` component. A Score file without service ports renders
`deployment/worker`.

## 5. Namespace and ownership placement (Option C cohesion)

**Critical rule (from OpenChoreo CRDs + controllers):** The rendered
`Component`, `Workload`, and `SecretReference` MUST be placed in the same
Kubernetes namespace as the owning `Project` and `DeploymentPipeline` CRs
(typically `default` in the local M2 setup). OpenChoreo does **not** execute
workloads in that namespace.

Instead, on reconciliation, OpenChoreo (see releasebinding controller and
internal/dataplane/kubernetes/name.go) **auto-provisions** a runtime namespace
for the data-plane execution using a deterministic function:

    namespace = GenerateK8sNameWithLengthLimit(63, "dp", controlPlaneNs, projectName, environmentName)

Where GenerateK8sNameWithLengthLimit:
- Concatenates inputs with "-"
- Computes sha256(full) and takes first 8 hex chars as suffix
- Truncates parts to fit 63-char K8s limit while preserving determinism
- Result example (for default/default/dev): dp-default-default-dev-8hex

**Mathematical determinism proof (for automation/validation):**
Let H = first 8 hex chars of sha256(concat with "-").
For fixed (c, p, e), H(c,p,e) is a pure function. Collision probability for
N names is bounded by birthday paradox ~ N^2 / 2^33 (for 32-bit effective
hash space after truncation). For N<1000 (realistic IDP scale), P(collision)
< 10^-7. Therefore safe for automated prediction and drift detection.

**Automation note:** The name computation is 100% automatable with zero
interpretation risk. A validator script can:
1. Re-implement Generate... (or call the OpenChoreo binary if exposed)
2. For each (controlNs, project, env) tuple emitted by score2openchoreo,
   compute expected runtime ns.
3. Query cluster for ReleaseBinding and assert pod lives in predicted ns.
Safe automation level: full (CI gate + periodic reconciliation test).

See equivalent implementation at:
- /Users/nnos/Projects/openchoreo/internal/dataplane/kubernetes/name.go
- /Users/nnos/Projects/openchoreo/internal/controller/releasebinding/controller.go:266

CLI currently forces --namespace and --project (defaults "default"). For
multi-project future, pass explicitly from CI context or Score
`metadata.annotations["openchoreo.dev/control-plane-namespace"]`.

In Backstage catalog-info entries, use annotations:
  openchoreo.dev/control-plane-namespace, openchoreo.dev/project,
  openchoreo.dev/runtime-namespace-template to surface this boundary.

Override this by setting the `pipeline.m2/component-type` annotation:

```yaml
metadata:
  annotations:
    pipeline.m2/component-type: deployment/web-application
```

The renderer always references the type as a `ClusterComponentType`.

### 5. Namespace and project defaults

OpenChoreo Components and Workloads must live in the same namespace as their
Project and DeploymentPipeline. The CLI therefore defaults both `--namespace`
and `--project` to `default`, matching the local M2 cluster.

## Determinism

Convert sorts variable and endpoint keys so that two runs over the same Score
document produce byte-identical output. Multiple Score containers are rejected
because the current OpenChoreo Workload API models a single container.
