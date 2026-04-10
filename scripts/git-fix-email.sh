#!/usr/bin/env bash
set -euo pipefail

GITEA_ADMIN_PASSWORD=$(cat "$HOME/.rational-reserve/m1-gitea-admin-password")

cd /Users/nnos/Projects/developer-portal

git config user.email "jason@trademomentumllc.com"
git commit --amend --author="Jason M Jarmacz <jason@trademomentumllc.com>" --no-edit
git push --force origin main
