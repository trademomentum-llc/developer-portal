#!/usr/bin/env bash
#
# repair-k3d-node-ip.sh
#
# The k3d server container's docker IP is reassigned whenever the Colima
# VM restarts (serverlb/tools race server-0 for the low addresses), while
# the kine datastore's Node record (/registry/minions/k3d-openchoreo-server-0)
# keeps the previous IP. On mismatch k3s dies with:
#   "Failed to start networking: ... error getting node subnet:
#    failed to find interface with specified node ip"
# (incidents: 2026-08-19, and twice on 2026-08-21).
#
# Lessons baked in from the 2026-08-21 recoveries:
#   - state.db is WAL-mode: always copy the trio (state.db, state.db-wal,
#     state.db-shm). Copying state.db alone can read as "malformed".
#   - docker reuses a stopped container's endpoint IP on plain
#     `docker start`, so: start server-0 alone first (it claims the
#     lowest free address), sample that IP, stop it, patch the DB to it,
#     then start it again plus serverlb/tools.
#   - kubectl access goes through serverlb -> host port 6550; if Colima's
#     host-side forwarder wedges, use an SSH tunnel instead:
#     ssh -F ~/.colima/_lima/colima/ssh.config -L 6550:127.0.0.1:6550 -N lima-colima
#
# Usage: scripts/repair-k3d-node-ip.sh [--dry-run]

set -euo pipefail

CLUSTER="openchoreo"
SERVER="k3d-${CLUSTER}-server-0"
DB_DIR="/var/lib/rancher/k3s/server/db"
WORK="/tmp/k3d-state-trio.$$"
export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"
DRY=0
[ "${1:-}" = "--dry-run" ] && DRY=1

sample_ip() {
    docker network inspect "k3d-${CLUSTER}" --format '{{json .Containers}}' \
        | python3 -c "import json,sys; cs=json.load(sys.stdin); print(next((v['IPv4Address'].split('/')[0] for v in cs.values() if v['Name']=='${SERVER}'), ''))"
}

echo "stopping cluster ${CLUSTER}..."
k3d cluster stop "${CLUSTER}" >/dev/null

echo "starting ${SERVER} alone to pin the lowest free IP..."
docker start "${SERVER}" >/dev/null
sleep 6
CURRENT_IP=$(sample_ip)
docker stop "${SERVER}" >/dev/null

if [ -z "${CURRENT_IP}" ]; then
    echo "ERROR: could not sample ${SERVER} IP" >&2
    exit 1
fi
echo "server IP for this boot: ${CURRENT_IP}"

mkdir -p "${WORK}"
docker cp "${SERVER}:${DB_DIR}/state.db" "${WORK}/state.db"
docker cp "${SERVER}:${DB_DIR}/state.db-wal" "${WORK}/state.db-wal" 2>/dev/null || true
docker cp "${SERVER}:${DB_DIR}/state.db-shm" "${WORK}/state.db-shm" 2>/dev/null || true
BACKUP="${HOME}/.rational-reserve/backups/state.db.bak-pre-nodeip-fix-$(date +%Y-%m-%d-%H%M)"
cp "${WORK}/state.db" "${BACKUP}"
echo "backup: ${BACKUP}"

python3 - "${WORK}/state.db" "${CURRENT_IP}" "${DRY}" <<'PYEOF'
import re, sqlite3, sys

path, new_ip, dry = sys.argv[1], sys.argv[2], sys.argv[3] == "1"
db = sqlite3.connect(path)
ic = db.execute("PRAGMA integrity_check").fetchone()[0]
print("integrity_check:", ic)
if ic != "ok":
    sys.exit(1)

rowid, name, value = db.execute(
    "SELECT id, name, value FROM kine WHERE name LIKE '/registry/minions/%' "
    "ORDER BY id DESC LIMIT 1").fetchone()
ips = sorted(set(re.findall(rb'\b172\.20\.0\.\d+\b', bytes(value))))
print(f"latest node row: id={rowid}; node-IPs present: {[i.decode() for i in ips]}")

new = new_ip.encode()
stale = [i for i in ips if i != new]
if not stale:
    print("node record already matches -- nothing to patch")
elif len(stale) != 1:
    print(f"ERROR: expected exactly one stale node IP, got {stale}; refusing", file=sys.stderr)
    sys.exit(1)
else:
    old = stale[0]
    assert len(old) == len(new), "same-length constraint violated"
    v = bytearray(value)
    count, start = 0, 0
    while True:
        i = bytes(v).find(old, start)
        if i < 0:
            break
        v[i:i+len(old)] = new
        count += 1
        start = i + len(new)
    print(f"patching {old.decode()} -> {new_ip} ({count} occurrence(s) in row {rowid})")
    if not dry:
        db.execute("UPDATE kine SET value=? WHERE id=?", (bytes(v), rowid))
        db.commit()
        print("checkpoint:", db.execute("PRAGMA wal_checkpoint(TRUNCATE)").fetchone())
    else:
        print("dry-run: not writing")
db.close()
PYEOF

if [ "${DRY}" = "0" ]; then
    docker cp "${WORK}/state.db" "${SERVER}:${DB_DIR}/state.db"
    # Fold the patched state: replace wal/shm with the checkpointed ones.
    docker cp "${WORK}/state.db-wal" "${SERVER}:${DB_DIR}/state.db-wal" 2>/dev/null || true
    docker cp "${WORK}/state.db-shm" "${SERVER}:${DB_DIR}/state.db-shm" 2>/dev/null || true
    echo "starting ${SERVER} (same endpoint, same IP)..."
    docker start "${SERVER}" >/dev/null
    echo "waiting for the apiserver to listen inside the container..."
    for i in $(seq 1 30); do
        sleep 10
        LISTENERS=$(docker exec "${SERVER}" sh -c 'awk "NR>1 {print \$2}" /proc/net/tcp /proc/net/tcp6 2>/dev/null | grep -c 192B' 2>/dev/null || echo 0)
        [ "${LISTENERS}" -gt 0 ] && break
    done
    echo "apiserver 6443 listeners: ${LISTENERS}"
    docker start "k3d-${CLUSTER}-serverlb" "k3d-${CLUSTER}-tools" >/dev/null
    echo "done. If localhost:6550 stays refused, Colima's forwarder is wedged:"
    echo "  ssh -F ~/.colima/_lima/colima/ssh.config -L 6550:127.0.0.1:6550 -N lima-colima"
else
    k3d cluster start "${CLUSTER}" >/dev/null
fi
rm -rf "${WORK}"
