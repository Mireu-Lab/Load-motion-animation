#!/usr/bin/env bash
# ==============================================================================
# Post-Creation Command for DevContainers & Codespaces
# Author/PM: limmireu1214@gmail.com (@Mireu-Lab)
# ==============================================================================
set -euo pipefail

echo "=== Initializing Limmireu Antigravity (AGY) & Multi-Server Sync Environment ==="
curl -fsSL https://raw.githubusercontent.com/Mireu-Lab/codespace-agy-env/main/install.sh | bash

echo "=== Codespace AGY Environment Setup Completed! ==="
