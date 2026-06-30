#!/usr/bin/env python3
"""
Test end-to-end Backstage sign-in via Gitea.
Expects Backstage dev server on localhost:3001 and Gitea on localhost:3333.
"""
import os
import sys
from playwright.sync_api import sync_playwright

BACKSTAGE_URL = os.environ.get("BACKSTAGE_URL", "http://localhost:3001")
GITEA_USER = os.environ.get("GITEA_USER", "gitea_admin")
GITEA_PASSWORD_FILE = os.path.expanduser("~/.rational-reserve/m1-gitea-admin-password")

if not os.path.exists(GITEA_PASSWORD_FILE):
    print("Gitea admin password file not found", file=sys.stderr)
    sys.exit(1)

GITEA_PASSWORD = open(GITEA_PASSWORD_FILE).read().strip()


def main():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        print(f"Opening {BACKSTAGE_URL}")
        page.goto(BACKSTAGE_URL)
        page.wait_for_load_state("networkidle")

        # Wait for the custom sign-in page to render.
        page.wait_for_selector("text=Gitea", timeout=10000)
        print("Clicking Gitea sign-in")

        # Intercept the auth popup.
        with page.expect_popup() as popup_info:
            page.click("text=Gitea >> .. >> button")
        popup = popup_info.value
        popup.wait_for_load_state("networkidle")

        print(f"Popup URL: {popup.url}")

        # Gitea login page.
        if "login" in popup.url:
            popup.fill("input[name='user_name']", GITEA_USER)
            popup.fill("input[name='password']", GITEA_PASSWORD)
            popup.click("button[type='submit']")
            popup.wait_for_load_state("networkidle")
            print(f"After login URL: {popup.url}")

        # OAuth authorize page.
        if "/oauth/authorize" in popup.url or "/login/oauth/authorize" in popup.url:
            # Authorize button text may be "Authorize Application" or similar.
            popup.click("button:has-text('Authorize')")
            popup.wait_for_load_state("networkidle")
            print(f"After authorize URL: {popup.url}")

        # Wait for the popup to close or redirect back to Backstage.
        popup.wait_for_event("close", timeout=20000)
        print("Popup closed")

        # Back on Backstage, wait for the app to load and capture identity.
        page.wait_for_load_state("networkidle")
        page.wait_for_timeout(2000)

        identity = page.evaluate(
            """
            () => {
                try {
                    const item = localStorage.getItem('@backstage/core-plugin-api_GiteaAuth');
                    return item ? JSON.parse(item) : null;
                } catch (e) {
                    return null;
                }
            }
            """
        )
        print(f"Gitea auth session in localStorage: {identity}")

        # Try to read the user profile from the profile API.
        profile = page.evaluate(
            """
            async () => {
                try {
                    const resp = await fetch('/api/auth/gitea/session');
                    return await resp.json();
                } catch (e) {
                    return { error: String(e) };
                }
            }
            """
        )
        print(f"Profile via /api/auth/gitea/session: {profile}")

        # Screenshot for debugging.
        page.screenshot(path="/tmp/backstage-after-gitea-signin.png", full_page=True)

        browser.close()


if __name__ == "__main__":
    main()
