import { createApp } from '@backstage/frontend-defaults';
import catalogPlugin from '@backstage/plugin-catalog/alpha';
import { navModule } from './modules/nav';
import { openchoreoCardsModule } from './modules/openchoreo-cards/index.tsx';

export default createApp({
  features: [catalogPlugin, navModule, openchoreoCardsModule],
});
