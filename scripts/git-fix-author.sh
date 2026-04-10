#!/usr/bin/env bash
set -euo pipefail

cd /Users/nnos/Projects/developer-portal

git config user.name "Jason M Jarmacz"
git config user.email "jason@trademomentum.net"

git commit --amend --author="Jason M Jarmacz <jason@trademomentum.net>" --no-edit
git push --force origin main
