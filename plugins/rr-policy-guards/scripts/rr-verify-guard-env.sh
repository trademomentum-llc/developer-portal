#!/bin/sh
# rr-verify-guard-env.sh -- environment wrapper for rr-verify-guard.
#
# The verify-guard reads forge credentials from environment variables
# (RR_VERIFY_GUARD_GITEA_URL / RR_VERIFY_GUARD_GITEA_TOKEN). Hook hosts
# (PreToolUse) do not source interactive shell profiles, so the guard's
# forge checks would otherwise run credential-less and degrade. This
# wrapper sources the custody env file (mode 600, untracked) if present
# and execs the guard binary. Tracked; contains no secrets itself.

ENV_FILE="${RR_VERIFY_GUARD_ENV:-$HOME/.rational-reserve/guard/gitea-local.env}"

if [ -r "$ENV_FILE" ]; then
    # shellcheck disable=SC1090
    . "$ENV_FILE"
    export RR_VERIFY_GUARD_GITEA_URL RR_VERIFY_GUARD_GITEA_TOKEN
fi

exec "$(dirname "$0")/../bin/rr-verify-guard.bin" "$@"
