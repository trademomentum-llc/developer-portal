#!/usr/bin/env bash
# colors.sh -- terminal color helpers sourced by install/teardown scripts.
# No-ops if stdout is not a TTY.

if [ -t 1 ]; then
    COLOR_RESET='\033[0m'
    COLOR_BOLD='\033[1m'
    COLOR_RED='\033[1;31m'
    COLOR_GREEN='\033[1;32m'
    COLOR_YELLOW='\033[1;33m'
    COLOR_BLUE='\033[1;34m'
    COLOR_CYAN='\033[1;36m'
else
    COLOR_RESET=''
    COLOR_BOLD=''
    COLOR_RED=''
    COLOR_GREEN=''
    COLOR_YELLOW=''
    COLOR_BLUE=''
    COLOR_CYAN=''
fi

export COLOR_RESET COLOR_BOLD COLOR_RED COLOR_GREEN COLOR_YELLOW COLOR_BLUE COLOR_CYAN
