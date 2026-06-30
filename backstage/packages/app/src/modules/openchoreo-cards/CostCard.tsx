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
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  // Post-deploy cost artifact committed by CI to platform-config.
  const giteaBase = 'http://localhost:3333';
  const costArtifactUrl = `${giteaBase}/openchoreo/platform-config/raw/branch/main/cost-artifacts/${component}/${env}/latest.json`;
  const openCostUrl = `http://localhost:29003/?namespace=${predictedNs}`;

  return (
    <InfoCard title="Cost (Infracost + OpenCost)" variant="gridItem">
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
          <Link to={costArtifactUrl}>
            Post-deploy Infracost artifact (platform-config)
          </Link>
          <Link to={openCostUrl}>
            Live OpenCost allocation for {predictedNs}
          </Link>
          <Link to={`${giteaBase}/openchoreo/hello-m2/actions`}>
            CI cost breakdown runs
          </Link>
          <Link to="/policies/C3-infracost-delta.rego">
            C3 Cost-Delta Policy (Rego)
          </Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          The Infracost artifact is generated on every push by the hello-m2 CI workflow
          and committed to platform-config. OpenCost provides live runtime allocation
          for the predicted namespace. Namespace is computed deterministically (see
          Option C).
        </Typography>
      </Box>
    </InfoCard>
  );
};