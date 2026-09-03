#!/usr/bin/env bash
set -euo pipefail

# Output directory for static site (default: public)
OUT_DIR="${1:-public}"
PORT="${GODOC_PORT:-6065}"
HOST="127.0.0.1"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${PROJECT_ROOT}"

echo "=== Generating Static Godoc Documentation into '${OUT_DIR}' ==="

# 1. Ensure godoc is installed
export PATH="$(go env GOPATH)/bin:${PATH}"
if ! command -v godoc &> /dev/null; then
    echo "Installing godoc..."
    go install golang.org/x/tools/cmd/godoc@latest
fi

# Clean destination
rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

# 2. Start godoc server in background
godoc -http="${HOST}:${PORT}" &
GODOC_PID=$!

cleanup() {
    kill "${GODOC_PID}" 2>/dev/null || true
    wait "${GODOC_PID}" 2>/dev/null || true
}
trap cleanup EXIT

MODULE_PATH="$(go list -m 2>/dev/null || echo "gitlab.com/mireu-lab/Load-motion-animation")"

# 3. Wait for godoc to be ready
echo "Waiting for godoc server to be ready on ${HOST}:${PORT}..."
READY=0
for i in {1..30}; do
    if curl -s "http://${HOST}:${PORT}/pkg/${MODULE_PATH}/" > /dev/null 2>&1; then
        READY=1
        break
    fi
    sleep 0.5
done

if [ "$READY" -ne 1 ]; then
    echo "Error: godoc server failed to start within timeout" >&2
    exit 1
fi

# 4. Scrape documentation using wget
TEMP_SCRAPE=$(mktemp -d)
pushd "${TEMP_SCRAPE}" > /dev/null

echo "Mirroring documentation pages and assets..."
wget -q -m -k -E -np \
    -e robots=off \
    "http://${HOST}:${PORT}/pkg/${MODULE_PATH}/" \
    "http://${HOST}:${PORT}/lib/godoc/" || true

popd > /dev/null

if [ -d "${TEMP_SCRAPE}/${HOST}:${PORT}" ]; then
    cp -r "${TEMP_SCRAPE}/${HOST}:${PORT}"/* "${OUT_DIR}/"
fi
rm -rf "${TEMP_SCRAPE}"

# 5. Create root redirection to module package docs
cat << HTML > "${OUT_DIR}/index.html"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="refresh" content="0; url=pkg/${MODULE_PATH}/index.html">
    <title>Load-motion-animation Godoc</title>
    <script>
        window.location.href = "pkg/${MODULE_PATH}/index.html";
    </script>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding-top: 60px;">
    <h2>Load-motion-animation Documentation</h2>
    <p>Redirecting to <a href="pkg/${MODULE_PATH}/index.html">Package Documentation</a>...</p>
</body>
</html>
HTML

echo "=== Godoc HTML Generation Complete! ==="
ls -lh "${OUT_DIR}"
