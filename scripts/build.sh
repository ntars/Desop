#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

# Python build wrapper (calls build_desop.py)
echo "Building Desop for Windows..."
./scripts/build_desop.py
