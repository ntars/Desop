#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

ACTION="${ACTION:-uninstall}"
INSTALL_DIR="${HOME}/.local/bin"
REMOVE_CONFIG="${REMOVE_CONFIG:-false}"

case "$ACTION" in
    uninstall)
        if [[ -f "$INSTALL_DIR/desop" ]]; then
            rm "$INSTALL_DIR/desop"
        fi
        if [[ "$REMOVE_CONFIG" == "true" ]]; then
            rm -rf ~/Library/Application\ Support/Desop
        fi
        echo "Desop removed"
        ;;
    *)
        echo "Unknown action: $ACTION" >&2
        exit 1
        ;;
esac
