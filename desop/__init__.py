"""Desop - macOS-style desktop layer for Windows written in Python."""

import os
from pathlib import Path

# Load version from VERSION file (single source of truth)
VERSION_FILE = Path(__file__).resolve().parent.parent / "VERSION"
__version__ = VERSION_FILE.read_text(encoding="utf-8").strip() if VERSION_FILE.exists() else "0.0.0"

APP_NAME = "Desop"
BUNDLE_ID = "cn.ntars.desop"
