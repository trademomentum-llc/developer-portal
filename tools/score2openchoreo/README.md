# score2openchoreo

Converts a [Score](https://score.dev) YAML document to an OpenChoreo
`core.choreo.dev/v1alpha1/Component` CRD. Used by the M2 CI pipeline after
schema validation, before committing the rendered Component to
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

## Determinism

Convert walks containers and variables in sorted key order so that two
runs over the same Score document produce byte-identical Component output.
This matters because CI commits the output into a git repo.
