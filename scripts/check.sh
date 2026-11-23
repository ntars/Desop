#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo '== black =='
if command -v black >/dev/null 2>&1; then
    if ! black --check desop/ scripts/*.py; then
        echo "Run: black desop/ scripts/*.py"
        exit 1
    fi
fi

echo '== flake8 =='
if command -v flake8 >/dev/null 2>&1; then
    if ! flake8 desop/ scripts/*.py; then
        exit 1
    fi
fi

echo '== mypy =='
if command -v mypy >/dev/null 2>&1; then
    if ! mypy desop/; then
        exit 1
    fi
fi

echo '== tests =='
# Add tests later

echo "All checks passed."
