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

export const openchoreoCardsModule = createFrontendModule({
  pluginId: 'catalog',
  extensions: [
    // Only the overview card stays on the default Overview grid.
    // The other four cards are rendered inside dedicated entity-page tabs
    // contributed by openchoreo-entity-page, so content is not duplicated.
    openchoreoOverviewCard,
  ],
});