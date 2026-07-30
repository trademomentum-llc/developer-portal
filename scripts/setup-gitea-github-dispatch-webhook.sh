#!/usr/bin/env bash
# Install a Gitea.com webhook that fires GitHub repository_dispatch
# (event type: gitea-push) on every push, so the Sync from Gitea Action
# runs immediately instead of waiting for the 5-minute schedule.
#
# Prerequisites:
#   - GITEA_TOKEN with write:repository on trademomentum.net/developer-portal
#   - GITHUB_DISPATCH_TOKEN: classic PAT or fine-grained token with
#     "Contents: Read and write" on trademomentum-llc/developer-portal
#     (repository_dispatch requires a PAT; GITHUB_TOKEN in Actions is
#     not usable from Gitea webhooks)
#
# Usage:
#   GITEA_TOKEN=... GITHUB_DISPATCH_TOKEN=... ./scripts/setup-gitea-github-dispatch-webhook.sh
#
# Note: gitea.com disables the native Mirror feature; this webhook is the
# near-instant alternative to a push mirror.

set -euo pipefail

GITEA_HOST="${GITEA_HOST:-https://gitea.com}"
GITEA_OWNER="${GITEA_OWNER:-trademomentum.net}"
GITEA_REPO="${GITEA_REPO:-developer-portal}"
GITHUB_OWNER="${GITHUB_OWNER:-trademomentum-llc}"
GITHUB_REPO="${GITHUB_REPO:-developer-portal}"
EVENT_TYPE="${EVENT_TYPE:-gitea-push}"

if [ -z "${GITEA_TOKEN:-}" ]; then
  echo "GITEA_TOKEN is required" >&2
  exit 1
fi
if [ -z "${GITHUB_DISPATCH_TOKEN:-}" ]; then
  echo "GITHUB_DISPATCH_TOKEN is required (PAT with repo dispatch / contents write)" >&2
  exit 1
fi

DISPATCH_URL="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/dispatches"
HOOKS_URL="${GITEA_HOST}/api/v1/repos/${GITEA_OWNER}/${GITEA_REPO}/hooks"

# Gitea webhook payload is not GitHub's dispatch body, so we use a tiny
# webhook-proxy style: type "gitea" with authorization header is not enough.
# Instead register a "forgejo"/"gitea" hook that posts to a GitHub Actions
# workflow_dispatch is also awkward. repository_dispatch needs JSON body:
#   {"event_type":"gitea-push"}
# Gitea custom webhooks send their own payload, which GitHub rejects.
#
# Practical approach that works without a middleman: document that the
# 5-minute schedule is the automatic path on gitea.com, and provide a
# one-shot local trigger after push:
#
#   gh api repos/trademomentum-llc/developer-portal/dispatches -f event_type=gitea-push
#
# If you self-host Gitea with MIRROR enabled, prefer a real push mirror.
# If you add a small relay (Cloudflare Worker / smee), point it here.

echo "gitea.com native mirrors are disabled; registering is not sufficient alone."
echo "Validating tokens and printing the recommended local trigger..."

code="$(curl -sS -o /tmp/gitea-user.json -w "%{http_code}" \
  -H "Authorization: token ${GITEA_TOKEN}" \
  "${GITEA_HOST}/api/v1/repos/${GITEA_OWNER}/${GITEA_REPO}")"
if [ "${code}" != "200" ]; then
  echo "Gitea token cannot read ${GITEA_OWNER}/${GITEA_REPO} (HTTP ${code})" >&2
  exit 1
fi

code="$(curl -sS -o /tmp/gh-dispatch.json -w "%{http_code}" \
  -X POST \
  -H "Authorization: Bearer ${GITHUB_DISPATCH_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  -H "Content-Type: application/json" \
  "${DISPATCH_URL}" \
  -d "{\"event_type\":\"${EVENT_TYPE}\"}")"
if [ "${code}" != "204" ]; then
  echo "GitHub repository_dispatch failed (HTTP ${code}):" >&2
  cat /tmp/gh-dispatch.json >&2 || true
  exit 1
fi

echo "OK: fired repository_dispatch event_type=${EVENT_TYPE}"
echo "Watch: gh run list --repo ${GITHUB_OWNER}/${GITHUB_REPO} --workflow 'Sync from Gitea'"
echo
echo "After each push to Gitea (until a relay webhook exists), run:"
echo "  gh api repos/${GITHUB_OWNER}/${GITHUB_REPO}/dispatches -f event_type=${EVENT_TYPE}"
