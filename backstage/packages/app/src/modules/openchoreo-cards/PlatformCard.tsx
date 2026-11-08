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
          {/* Dev SigNoz via managed :3301 forward; ingress alternative: https://signoz.local */}
          <Link to="http://localhost:3301/services">
            SigNoz platform health
          </Link>
          {/* FR-20: manual dev->staging promotion runbook (OQ-14: promotion
              stays a manual commit). TechDocs of the developer-portal
              component (techdocs-ref dir:., docs/runbooks/promotion.md);
              builds on demand with the local generator. */}
          <Link to="http://localhost:3001/docs/default/component/developer-portal/runbooks/promotion/">
            Promotion runbook (dev to staging)
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
