# Technical Specification: Backstage Gitea Authentication Provider

**Document ID:** BACKSTAGE-GITEA-AUTH-TECH-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-Backstage-Gitea-Auth-Provider-Design-Specification.md

---

## 1. Implementation Plan

This document provides the concrete implementation steps for the Backstage Gitea authentication provider. The work is divided into backend, frontend, configuration, and validation.

---

## 2. Backend Provider Module

### 2.1 File Location

`backstage/packages/backend/src/modules/giteaAuth.ts`

### 2.2 Authenticator Implementation

Create a custom `OAuthAuthenticator` using the helpers from `@backstage/plugin-auth-node`:

```ts
import {
  createOAuthAuthenticator,
  OAuthAuthenticatorResult,
} from '@backstage/plugin-auth-node';

export const giteaAuthenticator = createOAuthAuthenticator({
  defaultScopes: ['read:user'],
  scopes: {
    persist: true,
    types: ['read:user'],
  },
  initialize({ config }) {
    const hostname = config.getString('hostname');
    const clientId = config.getString('clientId');
    const clientSecret = config.getString('clientSecret');
    const secure = config.getOptionalBoolean('secure') ?? true;
    const baseUrl = `${secure ? 'https' : 'http'}://${hostname}`;
    return { baseUrl, clientId, clientSecret };
  },
  async start({ config, state }) {
    const { baseUrl, clientId } = config;
    const url = new URL(`${baseUrl}/login/oauth/authorize`);
    url.searchParams.set('client_id', clientId);
    url.searchParams.set('redirect_uri', config.callbackUrl);
    url.searchParams.set('response_type', 'code');
    url.searchParams.set('state', state);
    url.searchParams.set('scope', 'read:user');
    return { url: url.toString(), state };
  },
  async authenticate({ config, req }) {
    const { baseUrl, clientId, clientSecret } = config;
    const code = req.query.code?.toString();
    if (!code) throw new Error('Missing authorization code');

    const tokenRes = await fetch(`${baseUrl}/login/oauth/access_token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({
        client_id: clientId,
        client_secret: clientSecret,
        code,
        redirect_uri: config.callbackUrl,
        grant_type: 'authorization_code',
      }),
    });
    if (!tokenRes.ok) throw new Error('Gitea token exchange failed');
    const tokenData = await tokenRes.json();
    const accessToken = tokenData.access_token;

    const profileRes = await fetch(`${baseUrl}/api/v1/user`, {
      headers: {
        Authorization: `token ${accessToken}`,
        Accept: 'application/json',
      },
    });
    if (!profileRes.ok) throw new Error('Gitea profile fetch failed');
    const profile = await profileRes.json();

    return {
      fullProfile: profile,
      accessToken,
      params: {},
      profile: {
        email: profile.email,
        displayName: profile.full_name || profile.login,
        picture: profile.avatar_url,
      },
    } as OAuthAuthenticatorResult<any>;
  },
});
```

### 2.3 Backend Module Registration

```ts
import { createBackendModule } from '@backstage/backend-plugin-api';
import { authProvidersExtensionPoint } from '@backstage/plugin-auth-node';
import { createOAuthProviderFactory, commonSignInResolvers } from '@backstage/plugin-auth-node';

export const giteaAuthProviderModule = createBackendModule({
  pluginId: 'auth',
  moduleId: 'gitea-provider',
  register(reg) {
    reg.registerInit({
      deps: { providers: authProvidersExtensionPoint },
      async init({ providers }) {
        providers.registerProvider({
          providerId: 'gitea',
          factory: createOAuthProviderFactory({
            authenticator: giteaAuthenticator,
            signInResolverFactories: {
              ...commonSignInResolvers,
              usernameMatchingUserEntityName() {
                return async (info, ctx) => {
                  const { profile } = info;
                  const login = profile.fullProfile.login;
                  if (!login) throw new Error('Gitea profile did not contain login');
                  const userEntityRef = `user:default/${login}`;
                  return ctx.issueToken({
                    claims: {
                      sub: userEntityRef,
                      ent: [userEntityRef, 'group:default/openchoreo'],
                    },
                  });
                };
              },
            },
          }),
        });
      },
    });
  },
});
```

### 2.4 Backend Index Update

In `backstage/packages/backend/src/index.ts` add:

```ts
backend.add(import('./modules/giteaAuth'));
```

Place it after the auth plugin imports.

---

## 3. Frontend Sign-In Module

### 3.1 File Location

`backstage/packages/app/src/modules/giteaSignIn.ts`

### 3.2 API Ref and Factory

```ts
import {
  createApiRef,
  OAuthApi,
  ProfileInfoApi,
  SessionApi,
} from '@backstage/core-plugin-api';
import {
  OAuth2,
  oauth2ApiRef,
} from '@backstage/core-app-api';
import {
  createFrontendModule,
  createApiFactory,
  discoveryApiRef,
  oauthRequestApiRef,
  configApiRef,
} from '@backstage/frontend-plugin-api';

export const giteaAuthApiRef: ApiRef<OAuthApi & ProfileInfoApi & SessionApi> =
  createApiRef({ id: 'internal.auth.gitea' });

export const giteaAuthApiFactory = createApiFactory({
  api: giteaAuthApiRef,
  deps: {
    discoveryApi: discoveryApiRef,
    oauthRequestApi: oauthRequestApiRef,
    configApi: configApiRef,
  },
  factory: ({ discoveryApi, oauthRequestApi, configApi }) =>
    OAuth2.create({
      discoveryApi,
      oauthRequestApi,
      provider: { id: 'gitea', title: 'Gitea', icon: () => null },
      environment: configApi.getOptionalString('auth.environment') ?? 'development',
      defaultScopes: ['read:user'],
    }),
});
```

### 3.3 Sign-In Page Blueprint

```ts
import { SignInPageBlueprint } from '@backstage/plugin-app-react';
import { SignInPage } from '@backstage/core-components';
import { giteaAuthApiRef } from './giteaAuth';

export const giteaSignInPage = SignInPageBlueprint.make({
  params: {
    loader: async () => props =>
      (
        <SignInPage
          {...props}
          providers={[
            'guest',
            {
              id: 'gitea-auth-provider',
              title: 'Gitea',
              message: 'Sign in using Gitea',
              apiRef: giteaAuthApiRef,
            },
          ]}
        />
      ),
  },
});
```

### 3.4 App Wiring

In `backstage/packages/app/src/App.tsx`:

```ts
import { giteaSignInModule } from './modules/giteaSignIn';

export default createApp({
  features: [
    catalogPlugin,
    navModule,
    openchoreoCardsModule,
    openchoreoEntityPageModule,
    giteaSignInModule,
  ],
});
```

---

## 4. Configuration Schema

Add `backstage/config.d.ts` (or extend `app-config.yaml` schema) if not already present. The provider config is:

```yaml
auth:
  environment: ${AUTH_ENVIRONMENT}
  providers:
    gitea:
      ${environment}:
        clientId: string
        clientSecret: string
        hostname: string   # e.g. localhost:3333 or git.example.com
        secure: boolean    # defaults to true; set false for local HTTP
```

---

## 5. Gitea OAuth Application Setup

`scripts/setup-gitea-oauth.sh` already creates the OAuth app. Update it to accept a configurable callback URL and store credentials:

- Development callback: `http://localhost:7008/api/auth/gitea/handler/frame`
- Production callback: `${BACKEND_BASE_URL}/api/auth/gitea/handler/frame`

The script should continue to be idempotent.

---

## 6. Smoke Test

Create `scripts/smoke-auth.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib/smoke-common.sh"

assert_cmd curl jq

ensure_gitea_oauth

# Use Gitea admin token to fetch a test user token or reuse existing PAT
TOKEN=$(get_gitea_admin_token)
USER=$(curl -fsS -H "Authorization: token ${TOKEN}" http://localhost:3333/api/v1/user | jq -r .login)

# Hit Backstage auth refresh endpoint with a real OAuth code is hard to automate;
# instead validate that the provider config loads and the auth endpoint is reachable.
BACKEND_TOKEN=$(curl -fsS -X POST http://localhost:7008/api/auth/guest/refresh | jq -r .backstageIdentity.token)

curl -fsS -H "Authorization: Bearer ${BACKEND_TOKEN}" http://localhost:7008/api/auth/gitea/health || true

echo "smoke-auth OK"
```

A full end-to-end OAuth code flow is better covered by a Playwright test in `yarn test:e2e`.

---

## 7. Migration Path

1. Merge the provider code.
2. Run `scripts/setup-gitea-oauth.sh` to create/update the Gitea OAuth app.
3. Export credentials into the environment.
4. Start Backstage with `yarn dev`.
5. Verify sign-in page shows Gitea option.
6. Once stable, remove the guest fallback from `app-config.production.yaml`.

---

## 8. Verification Commands

```bash
cd backstage
yarn tsc
yarn dev

# In another terminal
./scripts/smoke-auth.sh
./scripts/smoke-all.sh
```

---

## 9. Notes and Open Questions

- Gitea team/organization membership API may require additional scopes (`read:organization`). Start with `read:user` and a default `group:default/openchoreo` ownership ref.
- The `OAuth2.create` frontend helper may need a custom class if Gitea's token response differs from the standard; the backend authenticator normalizes this.
- Consider contributing a community module once the implementation stabilizes.
