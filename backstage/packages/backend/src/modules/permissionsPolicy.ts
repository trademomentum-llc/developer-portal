import {
  coreServices,
  createBackendModule,
} from '@backstage/backend-plugin-api';
import { policyExtensionPoint } from '@backstage/plugin-permission-node/alpha';
import { SecurityRbacPolicy } from '../extensions/permissionsPolicyExtension';

// Wave 0 (SEC-PLANE-WAVE0-TECH-001 section 8.4): installs the
// SecurityRbacPolicy on the permission backend, replacing the permissive
// default policy. Default export mirrors the giteaAuth.ts module pattern.
const securityRbacPolicyModule = createBackendModule({
  pluginId: 'permission',
  moduleId: 'security-rbac-policy',
  register(reg) {
    reg.registerInit({
      deps: { policy: policyExtensionPoint, config: coreServices.rootConfig },
      async init({ policy, config }) {
        policy.setPolicy(new SecurityRbacPolicy(config));
      },
    });
  },
});

export default securityRbacPolicyModule;
