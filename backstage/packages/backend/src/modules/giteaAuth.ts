import { createBackendModule } from '@backstage/backend-plugin-api';
import {
  authProvidersExtensionPoint,
  createOAuthAuthenticator,
  createOAuthProviderFactory,
  commonSignInResolvers,
} from '@backstage/plugin-auth-node';

/**
 * Gitea user profile as returned by GET /api/v1/user.
 *
 * Only the fields used for identity mapping are declared; the rest are
 * intentionally ignored so the provider stays robust against additions.
 */
type GiteaProfile = {
  id: number;
  login: string;
  email?: string;
  full_name?: string;
  avatar_url?: string;
};

type GiteaContext = {
  baseUrl: string;
  clientId: string;
  clientSecret: string;
  callbackUrl: string;
};

const giteaAuthenticator = createOAuthAuthenticator<GiteaContext, GiteaProfile>({
  defaultProfileTransform: async result => ({
    profile: {
      email: result.fullProfile.email,
      displayName: result.fullProfile.full_name || result.fullProfile.login,
      picture: result.fullProfile.avatar_url,
    },
  }),
  scopes: {
    persist: true,
  },
  initialize({ callbackUrl, config }) {
    const hostname = config.getString('hostname');
    const clientId = config.getString('clientId');
    const clientSecret = config.getString('clientSecret');
    const secure = config.getOptionalBoolean('secure') ?? true;
    const baseUrl = `${secure ? 'https' : 'http'}://${hostname}`;
    return { baseUrl, clientId, clientSecret, callbackUrl };
  },
  async start({ scope, state }, ctx) {
    const url = new URL(`${ctx.baseUrl}/login/oauth/authorize`);
    url.searchParams.set('client_id', ctx.clientId);
    url.searchParams.set('redirect_uri', ctx.callbackUrl);
    url.searchParams.set('response_type', 'code');
    url.searchParams.set('state', state);
    url.searchParams.set('scope', scope);
    return { url: url.toString() };
  },
  async authenticate({ req }, ctx) {
    const code = (req.query as Record<string, string | undefined>).code;
    if (!code) {
      throw new Error('Gitea authentication failed: missing authorization code');
    }

    const tokenResponse = await fetch(`${ctx.baseUrl}/login/oauth/access_token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        client_id: ctx.clientId,
        client_secret: ctx.clientSecret,
        code,
        redirect_uri: ctx.callbackUrl,
        grant_type: 'authorization_code',
      }),
    });
    if (!tokenResponse.ok) {
      throw new Error(
        `Gitea token exchange failed: ${tokenResponse.status} ${tokenResponse.statusText}`,
      );
    }
    const tokenData = (await tokenResponse.json()) as {
      access_token?: string;
      token_type?: string;
      scope?: string;
    };
    const accessToken = tokenData.access_token;
    if (!accessToken) {
      throw new Error('Gitea token exchange did not return an access_token');
    }

    const profileResponse = await fetch(`${ctx.baseUrl}/api/v1/user`, {
      headers: {
        Authorization: `token ${accessToken}`,
        Accept: 'application/json',
      },
    });
    if (!profileResponse.ok) {
      throw new Error(
        `Gitea profile fetch failed: ${profileResponse.status} ${profileResponse.statusText}`,
      );
    }
    const profile = (await profileResponse.json()) as GiteaProfile;
    if (!profile.login) {
      throw new Error('Gitea profile did not contain a login field');
    }

    return {
      fullProfile: profile,
      session: {
        accessToken,
        tokenType: tokenData.token_type || 'bearer',
        scope: tokenData.scope || '',
      },
    };
  },
  async refresh({ refreshToken }, ctx) {
    if (!refreshToken) {
      throw new Error('Gitea refresh failed: no refresh token available');
    }

    const tokenResponse = await fetch(`${ctx.baseUrl}/login/oauth/access_token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        client_id: ctx.clientId,
        client_secret: ctx.clientSecret,
        refresh_token: refreshToken,
        grant_type: 'refresh_token',
      }),
    });
    if (!tokenResponse.ok) {
      throw new Error(
        `Gitea token refresh failed: ${tokenResponse.status} ${tokenResponse.statusText}`,
      );
    }
    const tokenData = (await tokenResponse.json()) as {
      access_token?: string;
      refresh_token?: string;
      token_type?: string;
      scope?: string;
    };
    const accessToken = tokenData.access_token;
    if (!accessToken) {
      throw new Error('Gitea token refresh did not return an access_token');
    }

    const profileResponse = await fetch(`${ctx.baseUrl}/api/v1/user`, {
      headers: {
        Authorization: `token ${accessToken}`,
        Accept: 'application/json',
      },
    });
    if (!profileResponse.ok) {
      throw new Error(
        `Gitea profile fetch during refresh failed: ${profileResponse.status} ${profileResponse.statusText}`,
      );
    }
    const profile = (await profileResponse.json()) as GiteaProfile;
    if (!profile.login) {
      throw new Error('Gitea profile did not contain a login field');
    }

    return {
      fullProfile: profile,
      session: {
        accessToken,
        tokenType: tokenData.token_type || 'bearer',
        scope: tokenData.scope || '',
      },
    };
  },
});

const giteaSignInResolver = async (
  info: {
    profile: { displayName?: string; email?: string; picture?: string };
    result: { fullProfile: GiteaProfile };
  },
  ctx: { issueToken: (params: { claims: { sub: string; ent?: string[] } }) => Promise<unknown> },
) => {
  const login = info.result.fullProfile.login;
  const userEntityRef = `user:default/${login}`;
  return ctx.issueToken({
    claims: {
      sub: userEntityRef,
      ent: [userEntityRef, 'group:default/openchoreo'],
    },
  }) as Promise<{ token: string }>;
};

const giteaAuthProviderModule = createBackendModule({
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
            signInResolver: giteaSignInResolver,
            signInResolverFactories: {
              ...commonSignInResolvers,
            },
          }),
        });
      },
    });
  },
});

export default giteaAuthProviderModule;
