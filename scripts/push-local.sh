#!/usr/bin/env bash
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")

cd /Users/nnos/Projects/developer-portal
git add -A
git commit -m "$1

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"

git push "http://gitea_admin:${GITEA_ADMIN_PASSWORD}@localhost:3002/trademomentum.net/developer-portal.git" main
echo "Pushed to local Gitea"
