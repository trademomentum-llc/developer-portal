import {
  ApiRef,
  BackstageIdentityApi,
  OAuthApi,
  ProfileInfoApi,
  SessionApi,
  createApiRef,
} from '@backstage/core-plugin-api';
import { OAuth2 } from '@backstage/core-app-api';
import {
  ApiBlueprint,
  configApiRef,
  createFrontendModule,
  discoveryApiRef,
  oauthRequestApiRef,
} from '@backstage/frontend-plugin-api';
import { SignInPageBlueprint } from '@backstage/plugin-app-react';
import { SignInPage } from '@backstage/core-components';

export const giteaAuthApiRef: ApiRef<
  OAuthApi & ProfileInfoApi & BackstageIdentityApi & SessionApi
> = createApiRef({
  id: 'internal.auth.gitea',
});

const giteaAuthApiFactory = ApiBlueprint.make({
  params: defineParams =>
    defineParams({
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
          configApi,
          environment:
            configApi.getOptionalString('auth.environment') ?? 'development',
          provider: {
            id: 'gitea',
            title: 'Gitea',
            icon: () => null,
          },
          defaultScopes: ['read:user'],
        }),
    }),
});

const giteaSignInPage = SignInPageBlueprint.make({
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

export const giteaSignInModule = createFrontendModule({
  // This extension intentionally uses the 'app' plugin id so it replaces the
  // default Backstage sign-in page. The app/sign-in-page config key disables
  // the default page without disabling this override.
  pluginId: 'app',
  extensions: [giteaAuthApiFactory, giteaSignInPage],
});
