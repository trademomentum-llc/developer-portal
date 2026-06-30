import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { predictRuntimeNamespace } from './namespace-predictor';

/**
 * CostCard
 *
 * Surfaces the Cost angle of the M3 multi-angle visibility model.
 * Annotation-driven and fully deterministic via the shared namespace predictor.
 *
 * In a full M3 deployment this card would also display live post-deploy
 * cost attribution from an Infracost collector + OpenChoreo cost signals.
 */
export const CostCard = () => {
  const { entity } = useEntity();

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const costCenter = annotations['openchoreo.dev/cost-center'] || project;

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  // Future: these would be real collector endpoints or static artifact URLs
  const infracostBase = 'http://localhost:8088'; // placeholder for M3 collector

  return (
    <InfoCard title="Cost (Infracost + Budget)" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Cost Center:</strong> {costCenter}
        </Typography>
        <Typography variant="body2">
          <strong>Environment:</strong> {env}
        </Typography>
        <Typography variant="body2" style={{ marginTop: 8 }}>
          <strong>Predicted Runtime NS (cost scope):</strong> <code>{predictedNs}</code>
        </Typography>

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={`${infracostBase}/reports?ns=${predictedNs}`}>
            Pre-deploy Infracost diff (PR gate)
          </Link>
          <Link to={`${infracostBase}/attribution?namespace=${predictedNs}`}>
            Post-deploy cost attribution
          </Link>
          <Link to="/policies/C3-infracost-delta.rego">
            C3 Cost-Delta Policy (Rego)
          </Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          Cost signals will be enriched once the M3 Infracost collector + OpenChoreo cost
          annotations are active. Namespace is computed deterministically (see Option C).
        </Typography>
      </Box>
    </InfoCard>
  );
};