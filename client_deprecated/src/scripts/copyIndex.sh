#!/bin/bash
set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
mkdir -p "$SCRIPT_DIR/../dist"
cp "$SCRIPT_DIR/../src/index.html" "$SCRIPT_DIR/../dist/index.html"
cp "$SCRIPT_DIR/../src/favicon.ico" "$SCRIPT_DIR/../dist/favicon.ico"
echo "Copied index.html to dist/"
