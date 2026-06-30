/*
 * Local frontend config schema for the Developer Portal app.
 *
 * The new frontend system schema in @backstage/frontend-app-api does not
 * declare app.title as visible, but @backstage/core-components SignInPage
 * reads it at runtime. This file makes the value available to the browser.
 */
export interface Config {
  app?: {
    /**
     * The title of the app, as shown in the Backstage web interface.
     * @visibility frontend
     */
    title?: string;
  };
}
