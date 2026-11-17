#!/usr/bin/env bash
# ==============================================================================
# Post-Creation Command for DevContainers & Codespaces
# Repository: Mireu-Lab/Load-motion-animation
# Project Manager / Owner: limmireu1214@gmail.com (@Mireu-Lab)
# ==============================================================================
set -euo pipefail

echo "=== Initializing Limmireu Antigravity (AGY) & Multi-Server Sync Environment ==="
curl -fsSL https://raw.githubusercontent.com/Mireu-Lab/codespace-agy-env/main/install.sh | bash

# Ensure sessions are pulled and environment is up to date
if [ -x "${HOME}/.local/bin/agy-sync" ]; then
    echo "=== Synchronizing Antigravity conversation sessions from central archive ==="
    "${HOME}/.local/bin/agy-sync" pull || true
fi

echo "=== Codespace AGY Environment Setup Completed! ==="
