#!/usr/bin/env python3
"""
Test end-to-end Backstage sign-in via Gitea.
Expects Backstage dev server on localhost:3001 and Gitea on localhost:3333.
"""
import os
import stat
import sys
import tempfile
from pathlib import Path

from playwright.sync_api import sync_playwright

BACKSTAGE_URL = os.environ.get("BACKSTAGE_URL", "http://localhost:3001")
GITEA_USER = os.environ.get("GITEA_USER", "gitea_admin")
GITEA_PASSWORD_FILE = os.path.expanduser("~/.rational-reserve/m1-gitea-admin-password")

if not os.path.exists(GITEA_PASSWORD_FILE):
    print("Gitea admin password file not found", file=sys.stderr)
    sys.exit(1)

with open(GITEA_PASSWORD_FILE, encoding="utf-8") as handle:
    GITEA_PASSWORD = handle.read().strip()


def artifact_dir() -> Path:
    configured = os.environ.get("ARTIFACT_DIR")
    if configured:
        path = Path(configured)
        path.mkdir(parents=True, exist_ok=True)
    else:
        path = Path(tempfile.mkdtemp(prefix="gitea-signin-"))
    try:
        path.chmod(stat.S_IRWXU)
    except OSError:
        pass
    return path


def main() -> None:
    out_dir = artifact_dir()
    print(f"Artifact directory: {out_dir}")

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        print(f"Opening {BACKSTAGE_URL}")
        page.goto(BACKSTAGE_URL)
        page.wait_for_load_state("networkidle")

        page.wait_for_selector("text=Gitea", timeout=10000)
        print("Clicking Gitea sign-in")

        with page.expect_popup() as popup_info:
            page.click("text=Gitea >> .. >> button")
        popup = popup_info.value
        popup.wait_for_load_state("networkidle")

        print(f"Popup URL: {popup.url}")

        if "login" in popup.url:
            popup.fill("input[name='user_name']", GITEA_USER)
            popup.fill("input[name='password']", GITEA_PASSWORD)
            popup.click("button[type='submit']")
            popup.wait_for_load_state("networkidle")
            print(f"After login URL: {popup.url}")

        if "/oauth/authorize" in popup.url or "/login/oauth/authorize" in popup.url:
            popup.click("button:has-text('Authorize')")
            popup.wait_for_load_state("networkidle")
            print(f"After authorize URL: {popup.url}")

        popup.wait_for_event("close", timeout=20000)
        print("Popup closed")

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

        screenshot = out_dir / "backstage-after-gitea-signin.png"
        page.screenshot(path=str(screenshot), full_page=True)
        print(f"Screenshot written to {screenshot}")

        browser.close()


if __name__ == "__main__":
    main()
