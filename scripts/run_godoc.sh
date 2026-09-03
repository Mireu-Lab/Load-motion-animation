#!/usr/bin/env bash
set -euo pipefail

PORT="${1:-6060}"
HOST="127.0.0.1"

export PATH="$(go env GOPATH)/bin:${PATH}"
if ! command -v godoc &> /dev/null; then
    echo "godoc not found in PATH. Installing golang.org/x/tools/cmd/godoc@latest..."
    go install golang.org/x/tools/cmd/godoc@latest
fi

MODULE_PATH="$(go list -m 2>/dev/null || echo "gitlab.com/mireu-lab/Load-motion-animation")"

echo "========================================================"
echo " Starting local Godoc server for Load-motion-animation"
echo " Open your browser at:"
echo "   http://${HOST}:${PORT}/pkg/${MODULE_PATH}/"
echo " Press Ctrl+C to stop the server."
echo "========================================================"

godoc -http="${HOST}:${PORT}"
