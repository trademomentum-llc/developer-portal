# Design Specification: Backstage Gitea Authentication Provider

**Document ID:** BACKSTAGE-GITEA-AUTH-DES-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-06-30-Backstage-Gitea-Auth-Provider-Requirements.md

---

## 1. Overview

This document describes the design of a custom Backstage authentication provider that uses Gitea's OAuth2 implementation as the identity source. The provider is split into a backend module (responsible for the OAuth2 handshake and token refresh) and a frontend module (responsible for rendering the sign-in option and exposing the auth API).

---

## 2. High-Level Flow

```
+-------------+     OAuth2 authorize      +----------------+
|   Browser   | ------------------------> |  Gitea (local) |
+-------------+                           +----------------+
       |                                          |
       | authorization code                       |
       v                                          v
+------------------+                       +----------------+
| Backstage app    | <---- access token ---| Gitea token    |
| (SignInPage)     |                       | endpoint       |
+--------+---------+                       +----------------+
         |
         | code + state
         v
+--------+----------------------------------+
| Backstage auth backend (gitea provider)   |
| - exchanges code for token                |
| - calls GET /api/v1/user                  |
| - returns Backstage profile + token set   |
+----------------+--------------------------+
                 |
                 | Backstage identity
                 v
+---------------------------------------------+
| Catalog / Permission / Plugin APIs          |
+---------------------------------------------+
```

---

## 3. Components

### 3.1 Backend Module: `packages/backend/src/modules/giteaAuth.ts`

A new backend module registered against the `auth` plugin via `createBackendModule`. It uses `createOAuthProviderFactory` with a custom `OAuthAuthenticator` that talks to Gitea.

Responsibilities:
- Read `auth.providers.gitea.${environment}` configuration.
- Redirect the browser to Gitea's `/login/oauth/authorize` endpoint.
- Exchange the authorization code at `/login/oauth/access_token`.
- Call `/api/v1/user` to retrieve profile information.
- Return an `OAuthAuthenticatorResult` containing the Gitea profile and access token.
- Provide a sign-in resolver that maps the profile to a Backstage identity.

### 3.2 Frontend Module: `packages/app/src/modules/giteaSignIn.ts`

A frontend module that:
- Declares a new `ApiRef` named `giteaAuthApiRef` with type `OAuthApi & ProfileInfoApi & SessionApi`.
- Provides an API factory using the standard `OAuth2` implementation from `@backstage/core-app-api` (or a thin custom implementation) that targets the backend provider id `gitea`.
- Uses `SignInPageBlueprint` to add a Gitea sign-in option.

### 3.3 App Wiring

`packages/app/src/App.tsx` will import the frontend module and add it to the `features` array of `createApp`. The guest provider remains available in local dev via configuration.

`packages/backend/src/index.ts` will import and add the backend module.

### 3.4 Configuration

Local development (`app-config.local.yaml.example`):

```yaml
auth:
  environment: development
  providers:
    gitea:
      development:
        clientId: ${GITEA_OAUTH_CLIENT_ID}
        clientSecret: ${GITEA_OAUTH_CLIENT_SECRET}
        hostname: localhost:3333
        secure: false
```

Production (`app-config.production.yaml`):

```yaml
auth:
  environment: production
  providers:
    gitea:
      production:
        clientId: ${GITEA_OAUTH_CLIENT_ID}
        clientSecret: ${GITEA_OAUTH_CLIENT_SECRET}
        hostname: ${GITEA_HOSTNAME}
        secure: true
```

The `hostname` value is used to build the Gitea OAuth2 URLs. The callback URL passed to Gitea is the standard Backstage auth handler frame, e.g. `http://localhost:7008/api/auth/gitea/handler/frame`.

### 3.5 Identity Resolver

The sign-in resolver returns:

```ts
{
  token: string;          // Backstage user token
  identity: {
    type: 'user';
    userEntityRef: `user:default/${giteaLogin}`;
    ownershipEntityRefs: string[]; // groups the user belongs to
  };
}
```

Group membership is derived from Gitea organization and team list endpoints. In the initial implementation a single default ownership `group:default/openchoreo` may be used as a fallback, with full team mapping as a follow-up.

---

## 4. Security Considerations

- Client secrets are read from environment variables/files and never committed.
- The provider runs only over HTTPS (`secure: true`) in production configuration.
- State parameter validation is handled by Backstage's OAuth2 helper.
- Tokens are stored in the browser by Backstage's session management; the backend never persists them.

---

## 5. Dependencies

Backend:
- `@backstage/plugin-auth-node`
- `@backstage/backend-plugin-api`
- `@backstage/plugin-auth-backend`

Frontend:
- `@backstage/core-plugin-api`
- `@backstage/core-app-api`
- `@backstage/plugin-app-react`
- `@backstage/frontend-plugin-api`

---

## 6. Testing Strategy

- Unit test for the Gitea OAuth2 URL builder and profile mapping.
- Smoke test `scripts/smoke-auth.sh` that:
  1. Ensures `scripts/setup-gitea-oauth.sh` has run.
  2. Obtains a Gitea user token or uses an existing admin PAT to create one.
  3. Exercises the Backstage `/api/auth/gitea/refresh` endpoint.
  4. Asserts the returned profile matches the Gitea user.
- `scripts/smoke-all.sh` must continue to pass.

---

## 7. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Gitea OAuth2 scopes are minimal compared to OIDC | Use direct `/api/v1/user` call with the access token; do not rely on `userinfo`. |
| Backstage auth API changes between releases | Pin Backstage package versions and test with `yarn tsc` on every change. |
| Guest sign-in accidentally enabled in production | Keep guest provider out of `app-config.production.yaml`; use config validation. |
| Gitea hostname differs between port-forwards | Make `hostname` configurable and expose it through environment variables. |

---

## 8. Acceptance Criteria

- A Gitea user can sign into Backstage through a Gitea-branded button.
- The Backstage profile returned after sign-in contains the Gitea `login` as the user entity ref.
- `yarn tsc` and the smoke suite pass.
