// Fetches one Gatekeeper constraint object through the kubernetes proxy.
// Cluster name must match the clusterLocatorMethods entry in app-config.yaml (section 5.3).
export async function fetchConstraint(
  fetchFn: typeof fetch,
  kind: string,
  name: string,
): Promise<Response> {
  return fetchFn(
    `/api/kubernetes/proxy/apis/constraints.gatekeeper.sh/v1beta1/${kind}/${name}`,
    { headers: { 'Backstage-Kubernetes-Cluster': 'k3d-openchoreo-local' } },
  );
}
