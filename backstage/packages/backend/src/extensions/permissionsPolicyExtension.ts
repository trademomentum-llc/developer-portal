import {
  PermissionPolicy,
  PolicyQuery,
  PolicyQueryUser,
} from '@backstage/plugin-permission-node';
import {
  AuthorizeResult,
  isPermission,
  PolicyDecision,
} from '@backstage/plugin-permission-common';
import {
  catalogEntityCreatePermission,
  catalogEntityReadPermission,
  catalogEntityRefreshPermission,
} from '@backstage/plugin-catalog-common/alpha';
import {
  catalogConditions,
  createCatalogConditionalDecision,
} from '@backstage/plugin-catalog-backend/alpha';
import {
  kubernetesClustersReadPermission,
  kubernetesProxyPermission,
  kubernetesResourcesReadPermission,
} from '@backstage/plugin-kubernetes-common';
import { Config } from '@backstage/config';

// Role model (SEC-PLANE-WAVE0-TECH-001 section 8):
//   admin     = user ref in permission.policy.adminUsers, or group:default/gitea_admin
//   developer = ownershipEntityRefs contains group:default/openchoreo
//   viewer    = any authenticated identity
// Anything not listed below is admin-only.
//
// Two corrections against the spec sketch were forced by the installed tree:
//   - createConditionalDecision is not exported by
//     @backstage/plugin-permission-node@0.10.11; createCatalogConditionalDecision
//     from @backstage/plugin-catalog-backend/alpha is the catalog-resource
//     equivalent used below for the owner-scoped refresh decision.
//   - catalogConditions.isEntityOwner is a condition factory and must be
//     invoked with the user's claims; the spec sketch passed it un-called.
export class SecurityRbacPolicy implements PermissionPolicy {
  constructor(private readonly config: Config) {}

  private rolesFor(ownershipEntityRefs: string[]): {
    isAdmin: boolean;
    isDeveloper: boolean;
  } {
    const admins =
      this.config.getOptionalStringArray('permission.policy.adminUsers') ??
      ['gitea_admin'];
    const adminRefs = new Set([
      ...admins.map(u => `user:default/${u}`),
      'group:default/gitea_admin',
    ]);
    return {
      isAdmin: ownershipEntityRefs.some(r => adminRefs.has(r)),
      isDeveloper: ownershipEntityRefs.includes('group:default/openchoreo'),
    };
  }

  async handle(
    request: PolicyQuery,
    user?: PolicyQueryUser,
  ): Promise<PolicyDecision> {
    const ownershipEntityRefs = user?.identity.ownershipEntityRefs ?? [];
    const { isAdmin, isDeveloper } = this.rolesFor(ownershipEntityRefs);
    if (isAdmin) {
      return { result: AuthorizeResult.ALLOW };
    }
    if (request.permission.name === catalogEntityReadPermission.name) {
      return { result: AuthorizeResult.ALLOW }; // viewer and up
    }
    if (
      request.permission.name === catalogEntityCreatePermission.name ||
      request.permission.name === kubernetesProxyPermission.name ||
      request.permission.name === kubernetesResourcesReadPermission.name ||
      request.permission.name === kubernetesClustersReadPermission.name
    ) {
      return isDeveloper
        ? { result: AuthorizeResult.ALLOW }
        : { result: AuthorizeResult.DENY };
    }
    if (isPermission(request.permission, catalogEntityRefreshPermission)) {
      // Developers refresh entities they own; admins are handled above.
      return isDeveloper
        ? createCatalogConditionalDecision(
            request.permission,
            catalogConditions.isEntityOwner({ claims: ownershipEntityRefs }),
          )
        : { result: AuthorizeResult.DENY };
    }
    // catalogEntityDeletePermission and everything unlisted: admin-only.
    return { result: AuthorizeResult.DENY };
  }
}
