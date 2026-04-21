#!/usr/bin/env bash
# confirm.sh -- interactive prompt helpers. No-ops if stdin is not a TTY.

confirm() {
    local prompt=${1:-"Continue?"}
    if [ ! -t 0 ]; then
        return 0
    fi
    local answer
    printf "%s [y/N] " "$prompt"
    read -r answer
    case "$answer" in
        y|Y|yes|YES) return 0 ;;
        *) return 1 ;;
    esac
}

export -f confirm
