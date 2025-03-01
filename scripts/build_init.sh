#!/bin/sh
# ===========================================================================
# File: build_init.sh
# Description: common variables & functions for the build scripts.
# ===========================================================================

set -e

# Global variables
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Invoke from project root
VERSION=$(cat ./scripts/VERSION)

# Version function used for version string comparison
version() { echo "$@" | awk -F. '{ printf("%d%03d%03d%03d\n", $1,$2,$3,$4); }'; }

# Ensure output directory existed
mkdir_output() {
    if [ -z "$1" ]; then
        mkdir -p devsecdb-build
        OUTPUT_DIR=$(cd devsecdb-build > /dev/null && pwd)
    else
        OUTPUT_DIR="$1"
    fi
    echo "$OUTPUT_DIR"
}