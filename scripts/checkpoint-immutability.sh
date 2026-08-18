#!/usr/bin/env bash
# checkpoint-immutability.sh -- monthly signed checkpoint tag (FR-005 / D-03).
#
# Creates an annotated, signed tag checkpoint-YYYY-MM binding the signer to
# the exact head commit and tree of the current HEAD, chaining to the
# previous checkpoint, and pushes it to origin AND github.
#
# The script REFUSES to create an unsigned tag (NFR-006): an unsigned
# checkpoint claims an anchor it does not provide. Tag signing requires
# OQ-01 resolved (signing key generated and git configured, section 5 of
# RECORD-IMMUTABILITY-TECH-001); until then this script exits 1 without
# creating anything. It likewise exits 1 before creating anything when the
# origin or github push remote is missing, so a failed run never leaves a
# partially published checkpoint.
#
# Usage:  scripts/checkpoint-immutability.sh [--dry-run]
# Dry run (no tag, no push; testability without mutation):
#         RECORD_CHECKPOINT_DRY_RUN=1 scripts/checkpoint-immutability.sh
#         (--dry-run is an alias for the env var.)

set -euo pipefail

case "${1:-}" in
  --dry-run) RECORD_CHECKPOINT_DRY_RUN=1 ;;
  "") ;;
  *) echo "usage: checkpoint-immutability.sh [--dry-run]" >&2; exit 1 ;;
esac

PREFIX="checkpoint"
BASE="${PREFIX}-$(date -u +%Y-%m)"

HEAD="$(git rev-parse HEAD)"
TREE="$(git rev-parse 'HEAD^{tree}')"
DATE_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
# Previous checkpoint for the chain. Sort key verified empirically on git
# 2.55.0: -version:refname compares digit runs numerically, so with base,
# -r2, and -r10 tags present it orders r10 > r2 > base > earlier zero-padded
# YYYY-MM -- head -n 1 is always the latest checkpoint. (Plain -refname
# would misrank r10 below r2 lexicographically; -creatordate ties when the
# tags share one commit.)
PREV="$(git tag -l "${PREFIX}-*" --sort=-version:refname | head -n 1 || true)"

# Append-only naming: never move an existing tag. A rerun inside the same
# month takes the next free -rN suffix (DES-001 D-03).
TAG="${BASE}"
N=2
while git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null 2>&1; do
  TAG="${BASE}-r${N}"
  N=$((N + 1))
done

# Signing is mandatory. git tag -s would also fail without a usable key,
# but fail here with an explicit reason instead of a gpg/ssh error.
if ! git config --get user.signingKey >/dev/null 2>&1; then
  echo "checkpoint: user.signingKey is not set (OQ-01 unresolved)." >&2
  echo "checkpoint: refusing to create an unsigned tag (NFR-006)." >&2
  exit 1
fi

# Both push targets must exist before anything is created: pushing to
# origin and only then discovering github is gone would leave partial
# publication, and a rerun would mint -r2 instead of retrying.
MISSING=""
for REMOTE in origin github; do
  if ! git remote get-url "${REMOTE}" >/dev/null 2>&1; then
    MISSING="${MISSING} ${REMOTE}"
  fi
done
if [ -n "${MISSING}" ]; then
  echo "checkpoint: missing git remote(s):${MISSING}." >&2
  echo "checkpoint: refusing to create a tag that cannot be pushed to both origin and github; no tag was created." >&2
  exit 1
fi

MSG_FILE="$(mktemp "${TMPDIR:-/tmp}/checkpoint-msg.XXXXXX")"
trap 'rm -f "${MSG_FILE}"' EXIT
{
  echo "record checkpoint ${TAG}"
  echo
  echo "head:     ${HEAD}"
  echo "tree:     ${TREE}"
  echo "date-utc: ${DATE_UTC}"
  echo "prev:     ${PREV:-none}"
  echo "spec:     RECORD-IMMUTABILITY-TECH-001 (FR-005 / D-03)"
} > "${MSG_FILE}"

if [ "${RECORD_CHECKPOINT_DRY_RUN:-0}" = "1" ]; then
  echo "checkpoint: DRY RUN -- would create signed tag ${TAG} with message:"
  cat "${MSG_FILE}"
  echo "checkpoint: DRY RUN -- would run: git tag -s -F <msg> ${TAG} ${HEAD}"
  echo "checkpoint: DRY RUN -- would run: git tag -v ${TAG}"
  echo "checkpoint: DRY RUN -- would push refs/tags/${TAG} to origin and github"
  exit 0
fi

git tag -s -F "${MSG_FILE}" "${TAG}" "${HEAD}"

# Verify before publishing: an unverifiable tag is not an anchor.
# SSH signatures require gpg.ssh.allowedSignersFile, otherwise
# git verify-tag fails with trust level undefined (man git-config).
git tag -v "${TAG}"

git push origin "refs/tags/${TAG}"
git push github "refs/tags/${TAG}"

echo "checkpoint: ${TAG} signed, verified, and pushed to origin and github."
