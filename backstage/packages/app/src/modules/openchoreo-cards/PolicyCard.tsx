import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { predictRuntimeNamespace } from './namespace-predictor';

/**
 * PolicyCard
 *
 * Surfaces the Policy & Compliance angle.
 * References the full IDP Policy Guard Layer triad and the active Rego/Go constraints
 * (C1, C2, C3) that protect the platform.
 *
 * The predicted runtime namespace scopes any future violation queries.
 */
export const PolicyCard = () => {
  const { entity } = useEntity();

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

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

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Progressive enforcement (audit → warn → enforce) is defined in the IDP Policy Guard
          Layer Design Specification. Violations for this namespace will appear here once the
          M3 policy collector is wired.
        </Typography>
      </Box>
    </InfoCard>
  );
};