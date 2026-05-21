# local-registry module

Creates the M2 in-cluster OCI registry used by Gitea Actions and OpenChoreo
workloads.

The registry is intentionally simple:

- namespace: `local-registry`
- Deployment image: `registry:2.8`
- Service DNS: `registry.local-registry.svc.cluster.local:5000`
- Service type: `NodePort`
- NodePort: `30082`

The NodePort is paired with `scripts/install-m1.sh` task 2a, which configures
k3s containerd to mirror pulls for
`registry.local-registry.svc.cluster.local:5000` to
`http://127.0.0.1:30082`. This avoids the k3d node trying to resolve cluster
DNS from outside the pod network and avoids HTTPS against the HTTP-only dev
registry.
