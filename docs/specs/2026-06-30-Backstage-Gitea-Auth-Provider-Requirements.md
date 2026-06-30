# Requirements Specification: Backstage Gitea Authentication Provider

**Document ID:** BACKSTAGE-GITEA-AUTH-REQ-001  
**Version:** 0.1  
**Date:** 2026-06-30  
**Predecessors:** 2026-05-28-M3-Production-Multi-Angle-Visibility-Requirements.md, 2026-06-30-M4-Cost-Visibility-Requirements.md

---

## 1. Purpose

This document defines the requirements for replacing the Backstage `guest` authentication provider with a real identity provider backed by the local Gitea instance. Today the developer portal relies on the guest provider in local development and leaves `auth.providers` empty in the production config template. This work closes that gap so that every Backstage user has a distinct identity mapped from Gitea, enabling audit trails, permissions, and catalog ownership that reflect the actual platform users.

---

## 2. Vision

A developer or operator opening the developer portal is prompted to sign in with Gitea. After OAuth approval, Backstage creates an identity derived from the Gitea user profile and group/organization membership. The same identity resolves to catalog `User` and `Group` entities, so ownership, permissions, and entity visibility align with the Gitea organization model.

---

## 3. Scope

### In Scope

- A custom Backstage backend authentication provider that implements OAuth2 against the local Gitea instance.
- A frontend sign-in surface (API ref + SignInPage blueprint) that uses the new provider.
- Configuration schema and app-config examples for local development and production.
- A sign-in resolver that maps the Gitea profile to a Backstage user entity ref and group ownership.
- A smoke test that proves a Gitea user can obtain a Backstage token through the provider.

### Out of Scope

- Gitea OIDC discovery (`/.well-known/openid-configuration`). Gitea's OAuth2 implementation is used directly because its OIDC userinfo endpoint is not compatible with Backstage's built-in OIDC authenticator without custom session handling.
- SAML, LDAP, or external identity providers.
- Automatic provisioning of Backstage `User`/`Group` entities from Gitea (the catalog provider path remains manual or provider-based; this provider only emits identity refs).

---

## 4. Functional Requirements

### 4.1 Backend Provider

- FR-AUTH-1: Register a provider under the id `gitea` in the Backstage auth backend.
- FR-AUTH-2: Implement the OAuth2 authorization-code flow with Gitea token endpoints.
- FR-AUTH-3: Fetch the Gitea user profile from `GET /api/v1/user` using the access token.
- FR-AUTH-4: Expose the provider configuration under `auth.providers.gitea.${environment}`.
- FR-AUTH-5: Support a configurable `hostname` so the same provider works for both local port-forwards and a production Gitea host.

### 4.2 Frontend Sign-In

- FR-AUTH-6: Add a Gitea sign-in option to the Backstage `SignInPage`.
- FR-AUTH-7: Use a dedicated `giteaAuthApiRef` so other plugins can request Gitea-scoped sessions.
- FR-AUTH-8: Keep guest sign-in available in local development while making Gitea the default/production path.

### 4.3 Identity Mapping

- FR-AUTH-9: Map the Gitea `login` field to a Backstage user entity ref `user:default/<login>`.
- FR-AUTH-10: Map Gitea organization/team membership to Backstage group entity refs under `group:default/<org-or-team>`.
- FR-AUTH-11: Provide a fallback email/name display from the Gitea profile when available.

---

## 5. Non-Functional Requirements

- The provider must be implemented as a Backstage backend module using the new backend system (`createBackendModule`).
- No emojis or non-ASCII characters in any committed file.
- Client secrets must be read from environment variables or files, never committed.
- The implementation must pass `yarn tsc` and the existing smoke suite.

---

## 6. Success Criteria

- `scripts/setup-gitea-oauth.sh` creates or reuses a Gitea OAuth application and stores credentials securely.
- Backstage dev server can be started with `auth.providers.gitea.development` configured and the Gitea sign-in button appears.
- A user can sign in via Gitea and the `/api/auth/gitea/refresh` endpoint returns a valid Backstage profile.
- `scripts/smoke-auth.sh` (new) passes end-to-end.
- `scripts/smoke-all.sh` continues to pass after the provider is wired.

---

## 7. References

- Backstage docs: https://backstage.io/docs/auth/
- Gitea OAuth2: https://docs.gitea.com/development/oauth2-provider
- Existing helper: `scripts/setup-gitea-oauth.sh`
