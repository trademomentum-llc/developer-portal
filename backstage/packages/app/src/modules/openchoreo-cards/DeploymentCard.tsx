import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { predictRuntimeNamespace } from './namespace-predictor';

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
 */
export const DeploymentCard = () => {
  const { entity } = useEntity();

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const env = annotations['openchoreo.dev/environment'] || 'development';
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;
  const template = annotations['openchoreo.dev/runtime-namespace-template'] || '(computed)';

  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  const openchoreoBase = annotations['openchoreo.dev/api-base'] || 'http://localhost:9090';

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

        <Box mt={2} display="flex" flexDirection="column" gridGap={4}>
          <Link to={`${openchoreoBase}/projects/${project}/components/${component}/releases`}>
            OpenChoreo ReleaseBindings
          </Link>
          <Link to={`${openchoreoBase}/namespaces/${predictedNs}`}>
            Data-plane resources (predicted NS)
          </Link>
          <Link to={`/iac/environments/${env}/kustomization.yaml`}>
            Flux Kustomization (platform-config)
          </Link>
          <Link to={`http://localhost:8080/dashboards?namespace=${predictedNs}`}>
            Pods / Workloads in SigNoz (filtered)
          </Link>
        </Box>

        <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
          The namespace above is produced by the deterministic predictor (Option C).
          When OpenChoreo publishes actual status back into the catalog entity, the
          "observed" value will be shown side-by-side for drift detection.
        </Typography>
      </Box>
    </InfoCard>
  );
};