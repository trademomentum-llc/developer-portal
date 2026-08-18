#!/bin/sh
# install-git-hooks.sh -- drop rr-commit-guard hooks into a target repo.
#
# Usage:
#   plugins/rr-policy-guards/scripts/install-git-hooks.sh [REPO_PATH]
#
# REPO_PATH defaults to the current working directory. The installer is
# idempotent: re-running replaces the hook files with the latest versions.
#
# After installation:
#   .git/hooks/pre-commit   -> calls rr-commit-guard --scan-staged
#   .git/hooks/commit-msg   -> calls rr-commit-guard --validate-msg "$1"
#   .git/hooks/pre-push     -> calls rr-commit-guard --pre-push "$1" "$2"
#
# All three hooks expect rr-commit-guard on PATH, or RR_COMMIT_GUARD_BIN to
# point at the binary.

set -eu

REPO="${1:-$(pwd)}"

if ! git -C "${REPO}" rev-parse --git-dir >/dev/null 2>&1; then
  echo "install-git-hooks: ${REPO} is not a git repository" >&2
  exit 1
fi

HOOKS_DIR="$(git -C "${REPO}" rev-parse --git-path hooks)"
case "${HOOKS_DIR}" in
  /*) ;;                      # absolute -- use as is
  *)  HOOKS_DIR="${REPO}/${HOOKS_DIR}" ;;
esac

mkdir -p "${HOOKS_DIR}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC="${SCRIPT_DIR}/../git-hooks"

if [ ! -d "${SRC}" ]; then
  echo "install-git-hooks: source hooks directory not found at ${SRC}" >&2
  exit 1
fi

for hook in pre-commit commit-msg pre-push; do
  src="${SRC}/${hook}"
  dst="${HOOKS_DIR}/${hook}"
  if [ -e "${dst}" ] && ! grep -q "rr-commit-guard" "${dst}" 2>/dev/null; then
    backup="${dst}.bak.$(date +%Y%m%d%H%M%S)"
    echo "install-git-hooks: existing ${hook} (non-rr-commit-guard) -- backing up to ${backup}"
    mv "${dst}" "${backup}"
  fi
  install -m 0755 "${src}" "${dst}"
  echo "install-git-hooks: installed ${dst}"
done

echo "install-git-hooks: done. Set RR_COMMIT_GUARD_BIN if rr-commit-guard is not on PATH."
