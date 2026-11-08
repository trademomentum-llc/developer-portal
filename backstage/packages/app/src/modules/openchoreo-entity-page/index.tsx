import { createFrontendModule } from '@backstage/frontend-plugin-api';
import { EntityContentBlueprint } from '@backstage/plugin-catalog-react/alpha';
import { EntityKubernetesContent } from '@backstage/plugin-kubernetes';
import Grid from '@material-ui/core/Grid';
import CloudIcon from '@material-ui/icons/Cloud';
import AssessmentIcon from '@material-ui/icons/Assessment';
import AttachMoneyIcon from '@material-ui/icons/AttachMoney';
import SecurityIcon from '@material-ui/icons/Security';
import SettingsIcon from '@material-ui/icons/Settings';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';

// Reuse the existing OpenChoreo card components in deliberate tab layouts.
// The Overview tab already renders all five cards via EntityCardBlueprint;
// this module adds dedicated angle tabs without duplicating content.
import { DeploymentCard } from '../openchoreo-cards/DeploymentCard';
import { ObservabilityLinksCard } from '../openchoreo-cards/ObservabilityLinksCard';
import { AlertsCard } from '../openchoreo-cards/AlertsCard';
import { CostCard } from '../openchoreo-cards/CostCard';
import { PolicyCard } from '../openchoreo-cards/PolicyCard';
import { PlatformCard } from '../openchoreo-cards/PlatformCard';
import { SecurityCard } from '../openchoreo-cards/SecurityCard';
import { CiRunsCard } from '../openchoreo-cards/CiRunsCard';
import { TestResultsCard } from '../openchoreo-cards/TestResultsCard';

const componentFilter = 'kind:component';

const deploymentContent = EntityContentBlueprint.make({
  name: 'openchoreo-deployment',
  params: {
    path: '/openchoreo-deployment',
    title: 'Deployment',
    group: 'deployment',
    icon: <CloudIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <DeploymentCard />
        </Grid>
        {/* FR-19: Backstage kubernetes plugin workload view, a section below
            the custom card (the custom card stays primary). Reaches the
            k3d-openchoreo cluster through the localKubectlProxy locator
            (kubectl proxy on :8001); discovers workloads via the entity's
            backstage.io/kubernetes-label-selector annotation
            (openchoreo.dev/component=<name>). Until that annotation is on the
            ingested entity the plugin renders its own explicit
            missing-annotation empty state (NFR-04). The plugin's standalone
            per-entity tab is disabled in app-config.yaml to avoid a duplicate
            surface. */}
        <Grid item xs={12}>
          <EntityKubernetesContent />
        </Grid>
      </Grid>
    ),
  },
});

const observabilityContent = EntityContentBlueprint.make({
  name: 'openchoreo-observability',
  params: {
    path: '/openchoreo-observability',
    title: 'Observability',
    group: 'observability',
    icon: <AssessmentIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <ObservabilityLinksCard />
        </Grid>
        <Grid item xs={12} md={6}>
          <AlertsCard />
        </Grid>
      </Grid>
    ),
  },
});

const costContent = EntityContentBlueprint.make({
  name: 'openchoreo-cost',
  params: {
    path: '/openchoreo-cost',
    title: 'Cost',
    group: 'cost',
    icon: <AttachMoneyIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <CostCard />
        </Grid>
      </Grid>
    ),
  },
});

const policyContent = EntityContentBlueprint.make({
  name: 'openchoreo-policy',
  params: {
    path: '/openchoreo-policy',
    title: 'Policy',
    group: 'operation',
    icon: <SecurityIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <PolicyCard />
        </Grid>
      </Grid>
    ),
  },
});

const platformContent = EntityContentBlueprint.make({
  name: 'openchoreo-platform',
  params: {
    path: '/openchoreo-platform',
    title: 'Platform',
    group: 'platform',
    icon: <SettingsIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <PlatformCard />
        </Grid>
      </Grid>
    ),
  },
});

const securityContent = EntityContentBlueprint.make({
  name: 'security',
  params: {
    path: '/security',
    title: 'Security',
    group: 'security',
    icon: <SecurityIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <SecurityCard />
        </Grid>
      </Grid>
    ),
  },
});

const engagementContent = EntityContentBlueprint.make({
  name: 'engagement',
  params: {
    path: '/engagement',
    title: 'Engagement',
    group: 'engagement',
    icon: <PlayCircleOutlineIcon />,
    filter: componentFilter,
    loader: async () => (
      <Grid container spacing={2}>
        <Grid item xs={12} md={6}>
          <CiRunsCard />
        </Grid>
        <Grid item xs={12} md={6}>
          <TestResultsCard />
        </Grid>
      </Grid>
    ),
  },
});

export const openchoreoEntityPageModule = createFrontendModule({
  pluginId: 'catalog',
  extensions: [
    deploymentContent,
    observabilityContent,
    costContent,
    policyContent,
    platformContent,
    securityContent,
    engagementContent,
  ],
});
