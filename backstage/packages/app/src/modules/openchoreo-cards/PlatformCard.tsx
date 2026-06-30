import { InfoCard, Link } from '@backstage/core-components';
import { useEntity } from '@backstage/plugin-catalog-react';
import { Box, Typography } from '@material-ui/core';

// Platform health / dependency links for a Component entity.
// Surfaces the IDP substrate that the workload depends on (delivery, registry,
// cluster add-ons, OpenChoreo planes) so the Platform angle is visible alongside
// Deployment, Runtime, Cost, and Policy.
export const PlatformCard = () => {
  const { entity } = useEntity();
  const name = entity.metadata.name;
  const annotations = entity.metadata.annotations ?? {};
  const repoUrl = annotations['gitea.io/project']
    ? `http://localhost:3333/${annotations['gitea.io/project']}`
    : 'http://localhost:3333/openchoreo/hello-m2';

  return (
    <InfoCard title="Platform Dependencies" variant="gridItem">
      <Box>
        <Typography variant="body2" gutterBottom>
          Component: <strong>{name}</strong>
        </Typography>

        <Box display="flex" flexDirection="column" gridGap={4} mt={1}>
          <Link to={repoUrl}>Gitea repo</Link>
          <Link to={`${repoUrl}/actions`}>Gitea Actions runs</Link>
          <Link to="http://localhost:3333/openchoreo/platform-config">
            Flux-watched platform-config repo
          </Link>
          <Link to="http://localhost:8080/services">
            SigNoz platform health
          </Link>
          <Typography variant="caption" style={{ marginTop: 8, opacity: 0.7 }}>
            Local-registry, Gitea runner, and OpenChoreo plane health are
            surfaced here; live status integration is deferred to M4 platform
            hardening.
          </Typography>
        </Box>
      </Box>
    </InfoCard>
  );
}
