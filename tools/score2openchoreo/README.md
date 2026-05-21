# score2openchoreo

Converts a [Score](https://score.dev) YAML document to OpenChoreo
`openchoreo.dev/v1alpha1` resources. The renderer emits a `Component` and a
matching `Workload` as a multi-document YAML stream. Used by the M2 CI pipeline
after schema validation, before committing the rendered resources to
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

...produces `SecretKeyRef{name: "db-secret", key: <field>}`. Override by
setting `resources.db.metadata.name`:

```yaml
resources:
  db:
    type: secret
    metadata:
      name: my-db-credentials   # -> SecretKeyRef{name: "my-db-credentials"}
```

This convention pairs with the platform's External-Secrets sync, which
creates per-app Kubernetes Secrets named `<resource-key>-secret` from the
OpenBao `kv/<env>/<resource-key>/` path. Override when the upstream secret
already has a different Kubernetes name.

### 3. Supported resource types

Only `secret` and `environment` resource types are handled. Any other
`type:` value fails Convert with an explicit error.

### 4. Component type inference

By default, a Score file with `service.ports` renders an OpenChoreo
`deployment/service` component. A Score file without service ports renders
`deployment/worker`.

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
