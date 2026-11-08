import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';
import { predictRuntimeNamespace } from './namespace-predictor';

// Shapes of the OpenChoreo API's
// GET /api/v1/namespaces/{ns}/releasebindings?component=<name> response
// (openapi/openchoreo-api.yaml: ReleaseBindingList -> ReleaseBinding ->
// ReleaseBindingSpec/Status). All fields optional: the card degrades
// honestly on partial data (NFR-04).
interface ReleaseBindingCondition {
  type?: string;
  status?: string;
  reason?: string;
  message?: string;
  lastTransitionTime?: string;
}

interface ReleaseBinding {
  metadata?: { name?: string };
  spec?: {
    releaseName?: string;
    environment?: string;
    state?: string;
    owner?: { projectName?: string; componentName?: string };
  };
  status?: {
    observedGeneration?: number;
    conditions?: ReleaseBindingCondition[];
  };
}

// The observed-state list is capped so a pathological environment cannot
// stretch the card; the API returns most resources unpaginated at this
// scale (pagination.cursor is ignored).
const MAX_BINDINGS = 5;

// One-line summary of a condition: "Ready=True (Reason)".
function conditionSummary(c: ReleaseBindingCondition): string {
  const base = `${c.type ?? 'unknown'}=${c.status ?? 'Unknown'}`;
  return c.reason ? `${base} (${c.reason})` : base;
}

/**
 * DeploymentCard
 *
 * Primary surface for the Deployment & Reconciliation angle of the M3 model.
 * This card is the most direct consumer of the deterministic namespace predictor.
 *
 * It shows the exact predicted runtime namespace that OpenChoreo will (or did)
 * create for the (control-plane, project, environment) triple, allowing a developer
 * to correlate git state, OpenChoreo ReleaseBinding, Flux resources, and actual pods
 * with a single canonical string.
 *
 * FR-21: the observed-state block queries the OpenChoreo control-plane API
 * for the component's ReleaseBindings through the authenticated Backstage
 * proxy (/api/proxy/openchoreo -> localhost:9090) and renders
 * releaseName/state/conditions next to the predicted values. Every failure
 * mode (proxy down, API 401 because no Thunder-issued JWT is attached yet,
 * zero bindings) is an explicit, labeled not-wired/empty state -- never a
 * placeholder styled as live data (NFR-04).
 */
export const DeploymentCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;
  const template = annotations['openchoreo.dev/runtime-namespace-template'] || '(computed)';

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  const openchoreoBase = annotations['openchoreo.dev/api-base'] || 'http://localhost:9090';

  // GitOps source of truth: the local Gitea platform-config repo (same idiom as CostCard).
  const giteaBase = 'http://localhost:3333';

  const [bindings, setBindings] = useState<ReleaseBinding[] | null>(null);
  const [bindingsState, setBindingsState] = useState<'loading' | 'absent' | 'error'>(
    'loading',
  );
  const [bindingsError, setBindingsError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(
      `/api/proxy/openchoreo/api/v1/namespaces/${controlNs}/releasebindings?component=${encodeURIComponent(component)}`,
    )
      .then(async res => {
        if (!res.ok) {
          if (res.status === 401 || res.status === 403) {
            throw new Error(
              `OpenChoreo API returned ${res.status}: it enforces Thunder-issued JWT bearer tokens and the openchoreo proxy has no token source wired yet`,
            );
          }
          throw new Error(`openchoreo proxy returned ${res.status}`);
        }
        const body = await res.json();
        const list = (body?.items ?? []) as ReleaseBinding[];
        if (!cancelled) {
          if (list.length === 0) {
            setBindingsState('absent');
          } else {
            setBindings(list.slice(0, MAX_BINDINGS));
          }
        }
      })
      .catch(err => {
        if (!cancelled) {
          setBindingsError(err.message);
          setBindingsState('error');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch, controlNs, component]);

  return (
    <InfoCard title="Deployment & Reconciliation" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Component:</strong> {component} | <strong>Env:</strong> {env}
        </Typography>

        <Box mt={1}>
          <Typography variant="body2">
            <strong>Predicted Runtime Namespace:</strong>
          </Typography>
          <Typography variant="body2" style={{ fontFamily: 'monospace', fontSize: '0.85em' }}>
            {predictedNs}
          </Typography>
        </Box>

        <Typography variant="body2" style={{ marginTop: 4 }}>
          <strong>Template / Observed:</strong> <code>{template}</code>
        </Typography>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Observed ReleaseBindings (OpenChoreo API):</strong>
          </Typography>
          {bindingsState === 'absent' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              No ReleaseBindings recorded for this component in namespace{' '}
              {controlNs}
            </Typography>
          ) : bindingsState === 'error' ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              <span style={{ color: 'orange' }}>
                Observed state: not wired ({bindingsError})
              </span>
            </Typography>
          ) : bindings === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            bindings.map(rb => (
              <Box key={rb.metadata?.name ?? rb.spec?.releaseName} mt={1}>
                <Typography variant="body2">
                  <strong>{rb.metadata?.name ?? 'unknown'}</strong> release{' '}
                  <code>{rb.spec?.releaseName ?? 'unknown'}</code> state{' '}
                  <strong>{rb.spec?.state ?? 'unknown'}</strong>
                  {rb.spec?.environment ? ` env ${rb.spec.environment}` : ''}
                </Typography>
                <Typography variant="caption" style={{ opacity: 0.8 }}>
                  {(rb.status?.conditions ?? [])
                    .map(conditionSummary)
                    .join(' | ') || 'no conditions reported'}
                </Typography>
              </Box>
            ))
          )}
        </Box>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={`${openchoreoBase}/projects/${project}/components/${component}/releases`}>
            OpenChoreo ReleaseBindings
          </Link>
          <Link to={`${openchoreoBase}/namespaces/${predictedNs}`}>
            Data-plane resources (predicted NS)
          </Link>
          {/* Repo layout is environments/{dev,staging}; the env annotation ("development")
              does not map onto those dir names, so link the environments/ root. */}
          <Link to={`${giteaBase}/openchoreo/platform-config/src/branch/main/environments`}>
            Flux Kustomization (platform-config)
          </Link>
          {/* Dev SigNoz via managed :3301 forward; ingress alternative: https://signoz.local */}
          <Link to={`http://localhost:3301/dashboards?namespace=${predictedNs}`}>
            Pods / Workloads in SigNoz (filtered)
          </Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          The namespace above is produced by the deterministic predictor (Option C);
          the observed block is live data from the OpenChoreo API
          (namespaces/{controlNs}/releasebindings?component={component}) through
          the openchoreo proxy. A 401 there means the API is enforcing its
          Thunder-issued JWT and no token source is attached to the proxy yet --
          the block says so instead of faking data. Predicted vs observed sit
          side-by-side for drift detection.
        </Typography>
      </Box>
    </InfoCard>
  );
};
