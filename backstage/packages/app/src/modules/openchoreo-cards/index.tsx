import { createFrontendModule } from '@backstage/frontend-plugin-api';
import { EntityCardBlueprint } from '@backstage/plugin-catalog-react/alpha';

// This module contributes the complete set of resilient, annotation-driven
// OpenChoreo entity cards that realize the M3 Production Multi-Angle Visibility
// model inside Backstage.
//
// All cards (including the original two) now consume the single deterministic
// namespace predictor (namespace-predictor.ts) whose output is mathematically
// identical to tools/namespace-predictor/main.go.
//
// See docs/specs/2026-05-28-OpenChoreo-Entity-Cards-*.md for the full triad.

const componentCardFilter = 'kind:component';

const openchoreoOverviewCard = EntityCardBlueprint.make({
  name: 'overview',
  params: {
    filter: componentCardFilter,
    loader: async () => {
      const { OpenChoreoOverviewCard } = await import('./OpenChoreoOverviewCard');
      return <OpenChoreoOverviewCard />;
    },
  },
});

const openchoreoObservabilityCard = EntityCardBlueprint.make({
  name: 'observability',
  params: {
    filter: componentCardFilter,
    loader: async () => {
      const { ObservabilityLinksCard } = await import('./ObservabilityLinksCard');
      return <ObservabilityLinksCard />;
    },
  },
});

const openchoreoCostCard = EntityCardBlueprint.make({
  name: 'cost',
  params: {
    filter: componentCardFilter,
    loader: async () => {
      const { CostCard } = await import('./CostCard');
      return <CostCard />;
    },
  },
});

const openchoreoPolicyCard = EntityCardBlueprint.make({
  name: 'policy',
  params: {
    filter: componentCardFilter,
    loader: async () => {
      const { PolicyCard } = await import('./PolicyCard');
      return <PolicyCard />;
    },
  },
});

const openchoreoDeploymentCard = EntityCardBlueprint.make({
  name: 'deployment',
  params: {
    filter: componentCardFilter,
    loader: async () => {
      const { DeploymentCard } = await import('./DeploymentCard');
      return <DeploymentCard />;
    },
  },
});

export const openchoreoCardsModule = createFrontendModule({
  pluginId: 'catalog',
  extensions: [
    // These contribute to Component entity pages via the catalog plugin.
    // For production entity page layouts a custom EntityLayout override is
    // recommended so the five cards can be placed in deliberate grid sections.
    openchoreoOverviewCard,
    openchoreoObservabilityCard,
    openchoreoCostCard,
    openchoreoPolicyCard,
    openchoreoDeploymentCard,
  ],
});