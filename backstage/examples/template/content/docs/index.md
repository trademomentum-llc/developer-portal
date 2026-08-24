# ${{ values.name }}

Service scaffolded from the developer-portal example-nodejs-template.

## What this is

A minimal Node.js HTTP service listening on port 3000:

- `GET /` returns a greeting.
- `GET /healthz` returns `{"status":"ok"}` for probes.

## Lifecycle on this platform

1. Push to `main` runs `.gitea/workflows/ci.yaml`: unit tests, Trivy and
   OSV security gates, then build and push of the container image to the
   local OCI registry.
2. On push, the deploy stage renders `score.yaml` into OpenChoreo resources
   with `score2openchoreo` and commits them to
   `openchoreo/platform-config/environments/dev/`.
3. Flux applies the change and OpenChoreo reconciles the component into a
   running pod in the development data-plane namespace.
