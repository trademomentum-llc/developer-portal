import { InfoCard, Link } from '@backstage/core-components';
import { useApi, fetchApiRef } from '@backstage/core-plugin-api';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { useEffect, useState } from 'react';
import { predictRuntimeNamespace } from './namespace-predictor';
import { fetchConstraint } from './gatekeeper';

// Violation entry shape from Gatekeeper constraint .status.violations.
// Field names verified against gatekeeper v3.17.1 pkg/audit/manager.go
// StatusViolation (json: kind, namespace, name, message); `msg` is kept as a
// fallback for older Gatekeeper versions.
interface ConstraintViolation {
  kind?: string;
  namespace?: string;
  name?: string;
  message?: string;
  msg?: string;
}

interface ConstraintState {
  kind: string;
  name: string;
  totalViolations: number;
  violations: ConstraintViolation[];
}

// The Wave 0 constraint set (SEC-PLANE-WAVE0-TECH-001 section 5.4).
// `resource` is the CRD plural used in the API path (verified live via
// `kubectl get crd | grep constraints.gatekeeper.sh`).
const CONSTRAINTS = [
  { kind: 'C1PlatformAddonsMainProtected', resource: 'c1platformaddonsmainprotected', name: 'c1-enforce' },
  { kind: 'C2ScoreSchemaValid', resource: 'c2scoreschemavalid', name: 'c2-enforce' },
  { kind: 'C3InfracostDelta', resource: 'c3infracostdelta', name: 'c3-enforce' },
];

const MAX_RENDERED_VIOLATIONS = 5;

/**
 * PolicyCard
 *
 * Surfaces the Policy & Compliance angle.
 * References the full IDP Policy Guard Layer triad and the active Rego/Go constraints
 * (C1, C2, C3) that protect the platform.
 *
 * Live constraint violation state is read through the Backstage kubernetes proxy
 * (cluster `k3d-openchoreo-local`, host-side `kubectl proxy` on :8001 managed by
 * scripts/start-backstage.sh). The predicted runtime namespace scopes the policy
 * context shown beside it.
 */
export const PolicyCard = () => {
  const { entity } = useEntity();
  const { fetch } = useApi(fetchApiRef);

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  const [constraints, setConstraints] = useState<ConstraintState[] | null>(null);
  const [violationsError, setViolationsError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    Promise.all(
      CONSTRAINTS.map(async c => {
        const res = await fetchConstraint(fetch, c.resource, c.name);
        if (!res.ok) {
          throw new Error(`kubernetes proxy returned ${res.status}`);
        }
        const body = await res.json();
        return {
          kind: c.kind,
          name: c.name,
          totalViolations: body?.status?.totalViolations ?? 0,
          violations: (body?.status?.violations ?? []) as ConstraintViolation[],
        };
      }),
    )
      .then(states => {
        if (!cancelled) {
          setConstraints(states);
        }
      })
      .catch(err => {
        if (!cancelled) {
          setViolationsError(err.message);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [fetch]);

  return (
    <InfoCard title="Policy & Compliance" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Component:</strong> {component} | <strong>Env:</strong> {env}
        </Typography>
        <Typography variant="body2" style={{ marginTop: 8 }}>
          <strong>Policy scope (predicted NS):</strong> <code>{predictedNs}</code>
        </Typography>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to="/policies/">Active Policy Bundle (C1/C2/C3)</Link>
          <Link to="/plugins/rr-policy-guards/">RR Policy Guard Binaries</Link>
          <Link to="https://github.com/openchoreo/openchoreo/tree/main/config/policies">
            Gatekeeper ConstraintTemplates (upstream)
          </Link>
          <Link to={`/policies/C2-score-schema-valid.rego`}>
            C2 Score Schema Enforcement
          </Link>
          <Link to={`/policies/C3-infracost-delta.rego`}>
            C3 Cost Delta Guard
          </Link>
        </Box>

        <Box mt={2}>
          <Typography variant="body2">
            <strong>Live Gatekeeper violations:</strong>
          </Typography>
          {violationsError ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              Gatekeeper violations: not wired (kubernetes proxy unavailable:{' '}
              {violationsError})
            </Typography>
          ) : constraints === null ? (
            <Typography variant="body2" style={{ marginTop: 4 }}>
              loading...
            </Typography>
          ) : (
            constraints.map(c => {
              const shown = c.violations.slice(0, MAX_RENDERED_VIOLATIONS);
              return (
                <Box key={c.name} mt={1}>
                  <Typography variant="body2">
                    <code>{c.name}</code> ({c.kind}): {c.totalViolations}{' '}
                    violation{c.totalViolations === 1 ? '' : 's'}
                  </Typography>
                  {shown.map(v => (
                    <Typography
                      variant="caption"
                      component="div"
                      key={`${v.kind}-${v.namespace ?? ''}-${v.name}`}
                    >
                      {v.kind} {v.namespace ? `${v.namespace}/` : ''}
                      {v.name}: {v.message ?? v.msg ?? ''}
                    </Typography>
                  ))}
                  {shown.length !== c.totalViolations && (
                    <Typography variant="caption" component="div">
                      showing first {shown.length} of {c.totalViolations}
                    </Typography>
                  )}
                </Box>
              );
            })
          )}
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Progressive enforcement (audit, warn, enforce) is defined in the IDP Policy Guard
          Layer Design Specification. Violations above are read live from each constraint's
          .status through the kubernetes proxy (kubectl proxy on localhost:8001).
        </Typography>
      </Box>
    </InfoCard>
  );
};
