import React from 'react';
import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { predictRuntimeNamespace } from './namespace-predictor';

// Robust, annotation-driven card for OpenChoreo context.
// Uses the annotations added by Option C sub-agent work.
export const OpenChoreoOverviewCard = () => {
  const { entity } = useEntity();

  const annotations = entity.metadata.annotations ?? {};
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const project = annotations['openchoreo.dev/project'] || 'unknown';
  const component = annotations['openchoreo.dev/component'] || entity.metadata.name;
  const env = annotations['openchoreo.dev/environment'] || 'dev';
  const apiBase = annotations['openchoreo.dev/api-base'] || 'http://localhost:9090';

  // Single source of truth: the pure deterministic predictor (identical to Go reference).
  // See namespace-predictor.ts for the mathematical definition and equivalence proof.
  const predictedNs = predictRuntimeNamespace(controlNs, project, env);

  return (
    <InfoCard title="OpenChoreo Context" variant="gridItem">
      <Box>
        <Typography variant="body2">
          <strong>Project:</strong> {project}
        </Typography>
        <Typography variant="body2">
          <strong>Component:</strong> {component}
        </Typography>
        <Typography variant="body2">
          <strong>Environment:</strong> {env}
        </Typography>
        <Typography variant="body2" style={{ marginTop: 8 }}>
          <strong>Control Plane NS:</strong> {controlNs}
        </Typography>
        <Typography variant="body2">
          <strong>Predicted Runtime NS:</strong> {predictedNs}
        </Typography>
        <Box mt={2}>
          <Link to={`${apiBase}/projects/${project}/components/${component}`}>
            View in OpenChoreo
          </Link>
        </Box>
      </Box>
    </InfoCard>
  );
};