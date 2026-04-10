#!/usr/bin/env bash
set -euo pipefail
cd /Users/nnos/Projects/developer-portal
git remote set-url origin http://localhost:3002/trademomentum.net/developer-portal.git
echo "Origin URL updated -- no credentials embedded"
git remote -v
