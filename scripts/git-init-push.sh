#!/usr/bin/env bash
# scripts/git-init-push.sh -- Initialize git and push to local Gitea.
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")

cd /Users/nnos/Projects/developer-portal

git init -b main
git config user.email "admin@trademomentum.net"
git config user.name "Jimmy Paras"
git add -A
git commit -m "M1 Substrate -- initial commit

- Three M1 specs (Requirements, Design, Technical)
- rr-policy-guards plugin (emoji-guard, bash-guard, brew-guard)
- Backstage scaffold wired to Gitea (trademomentum.net org)
- Install/teardown scripts
- Third-party license provenance (THIRD-PARTY-LICENSES.md)
- Gitea values, helper scripts, README

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"

git remote add origin "http://gitea_admin:${GITEA_ADMIN_PASSWORD}@localhost:3002/trademomentum.net/developer-portal.git"
git push -u origin main

echo ""
echo "Pushed to http://localhost:3002/trademomentum.net/developer-portal"
