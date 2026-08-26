#!/usr/bin/env bash
# scripts/backup-openbao.sh
#
# Raft snapshot backup (BAO-STORAGE-DES-001 section 7; OQ-2 disposition:
# snapshot-on-script, weekly cadence by operator convention, no CronJob).
# Run weekly and before any cluster surgery. Snapshots land in
# ~/.rational-reserve/backups/openbao/ (host disk; off-host copy is a
# recorded out-of-scope follow-up).
set -euo pipefail

OPENBAO_POD=${OPENBAO_POD:-openbao-0}
OPENBAO_NS=${OPENBAO_NS:-openbao}
CUSTODY_DIR=${OPENBAO_CUSTODY_DIR:-"$HOME/.rational-reserve/openbao"}
BACKUP_DIR=${OPENBAO_BACKUP_DIR:-"$HOME/.rational-reserve/backups/openbao"}
KEEP=${OPENBAO_BACKUP_KEEP:-8}

info() { printf "\033[1;36m[openbao-backup]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[openbao-backup ERROR]\033[0m %s\n" "$*" >&2; exit 1; }

[ -r "$CUSTODY_DIR/root-token" ] || fail "missing $CUSTODY_DIR/root-token (run bootstrap first)"
ROOT_TOKEN=$(cat "$CUSTODY_DIR/root-token")

umask 077
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

TS=$(date -u +%Y%m%d-%H%M%S)
SNAP="$BACKUP_DIR/openbao-$TS.snap"
POD_TMP=/tmp/openbao-snapshot.snap

kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- env VAULT_TOKEN="$ROOT_TOKEN" \
    bao operator raft snapshot save "$POD_TMP" >/dev/null || fail "raft snapshot save failed"
kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- cat "$POD_TMP" > "$SNAP" || fail "snapshot stream failed"
kubectl -n "$OPENBAO_NS" exec "$OPENBAO_POD" -- rm -f "$POD_TMP" >/dev/null
[ -s "$SNAP" ] || fail "snapshot $SNAP is empty"
chmod 600 "$SNAP"

# Retention: keep the newest KEEP snapshots.
ls -1t "$BACKUP_DIR"/openbao-*.snap 2>/dev/null | tail -n "+$((KEEP + 1))" | while read -r old; do
    rm -f "$old"
done

info "snapshot written: $SNAP"
