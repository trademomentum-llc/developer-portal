import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';
import { predictRuntimeNamespace } from './namespace-predictor';

// Robust card for multi-angle observability links.
// Uses SigNoz + the runtime namespace from annotations (or predictor).
export const ObservabilityLinksCard = () => {
  const { entity } = useEntity();

  const annotations = entity.metadata.annotations ?? {};
  const serviceName = entity.metadata.name;
  const env = annotations['openchoreo.dev/environment'] || 'development';

  // Single source of truth: the pure deterministic predictor (identical to Go reference).
  // The template annotation is retained only as a hint; the computed value is authoritative.
  const projectForNs = annotations['openchoreo.dev/project'] || 'unknown';
  const controlNs = annotations['openchoreo.dev/control-plane-namespace'] || 'default';
  const runtimeNs = annotations['openchoreo.dev/runtime-namespace-template'] || predictRuntimeNamespace(controlNs, projectForNs, env);

  const signozBase = 'http://localhost:8080'; // Configurable via app-config in future

  return (
    <InfoCard title="Observability (Multi-Angle)" variant="gridItem">
      <Box>
        <Typography variant="body2" gutterBottom>
          Service: <strong>{serviceName}</strong> | Env: <strong>{env}</strong>
        </Typography>

        <Typography variant="body2" style={{ marginBottom: 4 }}>
          Runtime namespace pattern: <code>{runtimeNs}</code>
        </Typography>

        <Box display="flex" flexDirection="column" gridGap={4} mt={1}>
          <Link to={`${signozBase}/traces?service=${serviceName}&env=${env}`}>
            Traces in SigNoz
          </Link>
          <Link to={`${signozBase}/metrics?service=${serviceName}`}>
            Metrics in SigNoz
          </Link>
          <Link to={`${signozBase}/logs?service=${serviceName}`}>
            Logs in SigNoz
          </Link>
          <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
            (Links will be enriched with exact runtime namespace once M3 collector + predictor integration lands)
          </Typography>
        </Box>
      </Box>
    </InfoCard>
  );
};