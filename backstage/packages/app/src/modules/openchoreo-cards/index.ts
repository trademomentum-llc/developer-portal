import { createFrontendModule } from '@backstage/frontend-plugin-api';
import { convertLegacyEntityCardExtension } from '@backstage/plugin-catalog-react/alpha';
import { OpenChoreoOverviewCard } from './OpenChoreoOverviewCard';
import { ObservabilityLinksCard } from './ObservabilityLinksCard';
import { CostCard } from './CostCard';
import { PolicyCard } from './PolicyCard';
import { DeploymentCard } from './DeploymentCard';

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

export const openchoreoCardsModule = createFrontendModule({
  pluginId: 'catalog',
  extensions: [
    // These contribute to Component entity pages via the catalog plugin.
    // For production entity page layouts a custom EntityLayout override is
    // recommended so the five cards can be placed in deliberate grid sections.
    convertLegacyEntityCardExtension(OpenChoreoOverviewCard, {
      name: 'openchoreo-overview',
      filter: componentCardFilter,
    }),
    convertLegacyEntityCardExtension(ObservabilityLinksCard, {
      name: 'openchoreo-observability',
      filter: componentCardFilter,
    }),
    convertLegacyEntityCardExtension(CostCard, {
      name: 'openchoreo-cost',
      filter: componentCardFilter,
    }),
    convertLegacyEntityCardExtension(PolicyCard, {
      name: 'openchoreo-policy',
      filter: componentCardFilter,
    }),
    convertLegacyEntityCardExtension(DeploymentCard, {
      name: 'openchoreo-deployment',
      filter: componentCardFilter,
    }),
  ],
});