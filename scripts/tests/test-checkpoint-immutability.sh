#!/usr/bin/env bash
#
# test-checkpoint-immutability.sh
# Integration tests for scripts/checkpoint-immutability.sh
# (RECORD-IMMUTABILITY-TECH-001 section 10.3).
#
# Usage:
#   ./scripts/tests/test-checkpoint-immutability.sh
#
# Every case runs in a throwaway scratch repo (mktemp -d) with its own
# SSH signing key and bare origin/github remotes. The real repository, its
# tags, and the user's git config are never touched; scratch is removed on
# exit, so the suite is re-runnable.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CHECKPOINT="${ROOT_DIR}/scripts/checkpoint-immutability.sh"

BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
RESET="\033[0m"

log() { echo -e "$*"; }
pass() { echo -e "${GREEN} PASS${RESET} $*"; }
fail() { echo -e "${RED} FAIL${RESET} $*"; exit 1; }
info() { echo -e "${BOLD}INFO${RESET} $*"; }

SCRATCH="$(mktemp -d)"
trap 'rm -rf "${SCRATCH}"' EXIT

BASE="checkpoint-$(date -u +%Y-%m)"

info "Testing ${CHECKPOINT}"
info "Scratch world: ${SCRATCH} (removed on exit)"

# make_repo DIR -- scratch repo with one commit and no signing config.
make_repo() {
    local dir="$1"
    git init -q -b main "${dir}"
    git -C "${dir}" config user.email "ckpt-test@example.com"
    git -C "${dir}" config user.name "Ckpt Test"
    git -C "${dir}" config commit.gpgsign false
    echo x > "${dir}/f"
    git -C "${dir}" add f
    git -C "${dir}" commit -q -m "feat: x" -m "why"
}

# add_signing DIR KEYDIR -- throwaway SSH signing, repo-local config only.
# user.signingKey points at the private key (no ssh-agent needed); the
# allowedSigners principal is the tagger email (verified on git 2.55.0).
add_signing() {
    local dir="$1" keydir="$2"
    mkdir -p "${keydir}"
    ssh-keygen -q -t ed25519 -N "" -C "scratch" -f "${keydir}/key"
    git -C "${dir}" config gpg.format ssh
    git -C "${dir}" config user.signingKey "${keydir}/key"
    awk '{print "ckpt-test@example.com " $1 " " $2}' "${keydir}/key.pub" > "${keydir}/allowed"
    git -C "${dir}" config gpg.ssh.allowedSignersFile "${keydir}/allowed"
}

# add_remotes DIR ORIGIN_BARE GITHUB_BARE -- two bare push targets.
add_remotes() {
    local dir="$1" origin_bare="$2" github_bare="$3"
    git init -q --bare "${origin_bare}"
    git init -q --bare "${github_bare}"
    git -C "${dir}" remote add origin "${origin_bare}"
    git -C "${dir}" remote add github "${github_bare}"
}

# remote_has DIR REMOTE PATTERN -- true if REMOTE advertises a matching ref.
remote_has() {
    git -C "$1" ls-remote --exit-code "$2" "$3" >/dev/null 2>&1
}

# ---- Case A: refusal without signing config (spec 10.3) -------------------

info "Case A: refusal path (no user.signingKey)"
A="${SCRATCH}/a"
make_repo "${A}"
code=0
out="$(cd "${A}" && "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 1 ] || fail "expected exit 1, got ${code}: ${out}"
echo "${out}" | grep -q "user.signingKey is not set (OQ-01 unresolved)" \
    || fail "missing signing-config message: ${out}"
echo "${out}" | grep -q "refusing to create an unsigned tag (NFR-006)" \
    || fail "missing refusal message: ${out}"
[ -z "$(git -C "${A}" tag -l 'checkpoint-*')" ] || fail "refusal created a tag"
pass "refuses unsigned tag: exit 1, spec'd message, no tag created"

# ---- Case H: refusal when a push remote is missing (MINOR-1 preflight) ----

info "Case H: refusal path (github remote missing)"
H="${SCRATCH}/h"
make_repo "${H}"
add_signing "${H}" "${SCRATCH}/h-keys"
git init -q --bare "${SCRATCH}/h-origin.git"
git -C "${H}" remote add origin "${SCRATCH}/h-origin.git"
code=0
out="$(cd "${H}" && "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 1 ] || fail "expected exit 1, got ${code}: ${out}"
echo "${out}" | grep -qF "missing git remote(s): github" \
    || fail "message must name the missing remote: ${out}"
echo "${out}" | grep -qF "no tag was created" \
    || fail "missing no-tag-created statement: ${out}"
[ -z "$(git -C "${H}" tag -l 'checkpoint-*')" ] || fail "preflight created a tag"
pass "missing github remote: exit 1, message names github, no tag created"

# ---- Case B+E: happy path, signed tag, push to BOTH remotes ----------------

info "Case B+E: happy path with throwaway SSH signing, dual remotes"
B="${SCRATCH}/b"
make_repo "${B}"
add_signing "${B}" "${SCRATCH}/b-keys"
add_remotes "${B}" "${SCRATCH}/b-origin.git" "${SCRATCH}/b-github.git"
code=0
out="$(cd "${B}" && "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "happy path failed (exit ${code}): ${out}"
[ "$(git -C "${B}" cat-file -t "${BASE}" 2>/dev/null)" = "tag" ] \
    || fail "${BASE} missing or not an annotated tag"
code=0
vout="$(git -C "${B}" tag -v "${BASE}" 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "git tag -v failed: ${vout}"
echo "${vout}" | grep -q 'Good "git" signature' || fail "unexpected tag -v output: ${vout}"
msg="$(git -C "${B}" tag -l --format='%(contents)' "${BASE}")"
head_sha="$(git -C "${B}" rev-parse HEAD)"
tree_sha="$(git -C "${B}" rev-parse 'HEAD^{tree}')"
echo "${msg}" | grep -q "head:     ${head_sha}" || fail "message missing head field"
echo "${msg}" | grep -q "tree:     ${tree_sha}" || fail "message missing tree field"
echo "${msg}" | grep -q "date-utc: " || fail "message missing date-utc field"
echo "${msg}" | grep -q "prev:     none" || fail "first checkpoint must chain prev: none"
remote_has "${B}" origin "refs/tags/${BASE}" || fail "origin does not have ${BASE}"
remote_has "${B}" github "refs/tags/${BASE}" || fail "github does not have ${BASE}"
echo "${out}" | grep -q "signed, verified, and pushed to origin and github" \
    || fail "missing final summary line: ${out}"
pass "annotated signed tag ${BASE}; git tag -v good; pushed to origin AND github"

# ---- Case C: same-month rerun takes -r2, chained to base -------------------

info "Case C: same-month rerun produces -r2"
code=0
out="$(cd "${B}" && "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "rerun failed (exit ${code}): ${out}"
[ "$(git -C "${B}" cat-file -t "${BASE}-r2" 2>/dev/null)" = "tag" ] \
    || fail "${BASE}-r2 not created"
git -C "${B}" tag -v "${BASE}-r2" >/dev/null 2>&1 || fail "${BASE}-r2 not signed"
msg="$(git -C "${B}" tag -l --format='%(contents)' "${BASE}-r2")"
echo "${msg}" | grep -q "prev:     ${BASE}" || fail "-r2 must chain prev to ${BASE}: ${msg}"
remote_has "${B}" origin "refs/tags/${BASE}-r2" || fail "origin does not have ${BASE}-r2"
remote_has "${B}" github "refs/tags/${BASE}-r2" || fail "github does not have ${BASE}-r2"
pass "rerun creates ${BASE}-r2, signed, chained prev: ${BASE}, pushed to both remotes"

# ---- Case D: PREV ordering with base, -r2, -r10 (M6 adjudication) ----------

info "Case D: PREV chains to -r10 when base, -r2, -r10 all exist"
D="${SCRATCH}/d"
make_repo "${D}"
add_signing "${D}" "${SCRATCH}/d-keys"
add_remotes "${D}" "${SCRATCH}/d-origin.git" "${SCRATCH}/d-github.git"
git -C "${D}" tag "${BASE}"
git -C "${D}" tag "${BASE}-r2"
git -C "${D}" tag "${BASE}-r10"
code=0
out="$(cd "${D}" && "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "checkpoint with pre-existing tags failed (exit ${code}): ${out}"
# Next free suffix after base and -r2 is -r3; its prev must be the latest
# checkpoint in version order, which is -r10.
new_tag="$(echo "${out}" | sed -n 's/^checkpoint: \(checkpoint-[^ ]*\) signed, verified.*/\1/p')"
[ -n "${new_tag}" ] || fail "could not parse new tag name from: ${out}"
msg="$(git -C "${D}" tag -l --format='%(contents)' "${new_tag}")"
echo "${msg}" | grep -q "prev:     ${BASE}-r10" \
    || fail "new checkpoint must chain prev to ${BASE}-r10: ${msg}"
git -C "${D}" tag -v "${new_tag}" >/dev/null 2>&1 || fail "${new_tag} not signed"
remote_has "${D}" origin "refs/tags/${new_tag}" || fail "origin does not have ${new_tag}"
remote_has "${D}" github "refs/tags/${new_tag}" || fail "github does not have ${new_tag}"
pass "with base/-r2/-r10 present, ${new_tag} chains prev: ${BASE}-r10"

# ---- Case F: dry-run purity (spec 10.3) ------------------------------------

info "Case F: dry run creates no tag and pushes nothing"
F="${SCRATCH}/f"
make_repo "${F}"
add_signing "${F}" "${SCRATCH}/f-keys"
add_remotes "${F}" "${SCRATCH}/f-origin.git" "${SCRATCH}/f-github.git"
code=0
out="$(cd "${F}" && RECORD_CHECKPOINT_DRY_RUN=1 "${CHECKPOINT}" 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "env-var dry run failed (exit ${code}): ${out}"
echo "${out}" | grep -q "DRY RUN -- would create signed tag ${BASE}" \
    || fail "missing dry-run header: ${out}"
echo "${out}" | grep -q "head:     " || fail "dry run did not print the tag message"
[ -z "$(git -C "${F}" tag -l 'checkpoint-*')" ] || fail "dry run created a tag"
if remote_has "${F}" origin 'refs/tags/checkpoint-*'; then
    fail "dry run pushed to origin"
fi
if remote_has "${F}" github 'refs/tags/checkpoint-*'; then
    fail "dry run pushed to github"
fi
pass "RECORD_CHECKPOINT_DRY_RUN=1 prints the message, creates no tag, pushes nothing"

code=0
out="$(cd "${F}" && "${CHECKPOINT}" --dry-run 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "--dry-run flag failed (exit ${code}): ${out}"
echo "${out}" | grep -q "DRY RUN" || fail "--dry-run flag not honored: ${out}"
[ -z "$(git -C "${F}" tag -l 'checkpoint-*')" ] || fail "--dry-run created a tag"
pass "--dry-run flag is a side-effect-free alias"

# spec 10.3 suffix logic: with the base tag present, dry run selects -r2.
git -C "${F}" tag "${BASE}"
code=0
out="$(cd "${F}" && "${CHECKPOINT}" --dry-run 2>&1)" || code=$?
[ "${code}" -eq 0 ] || fail "suffix dry run failed (exit ${code}): ${out}"
echo "${out}" | grep -q "would create signed tag ${BASE}-r2" \
    || fail "with base present, dry run should select ${BASE}-r2: ${out}"
[ -z "$(git -C "${F}" tag -l "${BASE}-r2")" ] || fail "dry run created ${BASE}-r2"
pass "with ${BASE} present, dry run selects ${BASE}-r2 (and still creates nothing)"

# ---- Case G: static checks on the checkpoint script ------------------------

info "Case G: bash -n and shellcheck"
bash -n "${CHECKPOINT}" || fail "bash -n failed on ${CHECKPOINT}"
pass "bash -n clean"
if command -v shellcheck >/dev/null 2>&1; then
    shellcheck "${CHECKPOINT}" || fail "shellcheck findings in ${CHECKPOINT}"
    pass "shellcheck clean ($(shellcheck --version | awk '/^version:/{print $2}'))"
else
    info "shellcheck not installed -- skipping (no install attempted)"
fi

log ""
pass "all checkpoint-immutability tests passed"
