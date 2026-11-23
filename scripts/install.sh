#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

ACTION="${ACTION:-install}"
INSTALL_DIR="${HOME}/.local/bin"
KEEP_CONFIG="${KEEP_CONFIG:-false}"

case "$ACTION" in
    install)
        if [[ ! -d "$INSTALL_DIR" ]]; then
            mkdir -p "$INSTALL_DIR"
        fi
        if [[ -f "dist/desop" ]]; then
            cp dist/desop "$INSTALL_DIR/desop"
            chmod +x "$INSTALL_DIR/desop"
        elif [[ -f "release/desop-macos-*/desop.app/Contents/MacOS/desop" ]]; then
            mkdir -p "$INSTALL_DIR"
            cp release/desop-macos-*/desop.app/Contents/MacOS/desop "$INSTALL_DIR/desop"
            chmod +x "$INSTALL_DIR/desop"
        else
            echo "desop binary not found. Run build_desop.py first." >&2
            exit 1
        fi
        echo "Desop installed to $INSTALL_DIR/desop"
        ;;
    uninstall)
        if [[ -f "$INSTALL_DIR/desop" ]]; then
            rm "$INSTALL_DIR/desop"
        fi
        echo "Desop removed from $INSTALL_DIR"
        ;;
    *)
        echo "Unknown action: $ACTION" >&2
        exit 1
        ;;
esac
