#!/usr/bin/env bash
# wait-for.sh -- wait for a kubernetes resource to reach Ready.

wait_for_pod_ready() {
    local ns=$1
    local label=$2
    local timeout=${3:-300s}
    kubectl wait --for=condition=ready pod -l "$label" -n "$ns" --timeout="$timeout"
}

wait_for_port() {
    local host=$1
    local port=$2
    local timeout=${3:-60}
    local elapsed=0
    while [ "$elapsed" -lt "$timeout" ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            return 0
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    return 1
}

export -f wait_for_pod_ready wait_for_port
