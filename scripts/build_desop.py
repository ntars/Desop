#!/usr/bin/env python3
"""Build Desop for macOS."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path


def run(argv: list[str]) -> int:
    """Run a command and return its exit code."""
    result = subprocess.run(argv, check=False)
    return result.returncode


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    os.chdir(root)

    version = "0.1.0"
    out_dir = root / "release" / f"desop-macos-{version}"
    out_dir.mkdir(parents=True, exist_ok=True)

    print(f"Building Desop {version} for macOS")

    # Clean previous build
    dist_dir = root / "dist"
    if dist_dir.exists():
        shutil.rmtree(dist_dir)

    # Build with PyInstaller
    spec = root / "desop.spec"
    if not spec.exists():
        # Generate default spec
        spec.write_text(
            """# -*- mode: python ; coding: utf-8 -*-

block_cipher = None

a = Analysis(
    ['desop/__main__.py'],
    pathex=[],
    binaries=[],
    datas=[],
    hiddenimports=['PySide6', 'PySide6.QtWidgets', 'PySide6.QtGui'],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
)

pyz = PYZ(a.pyc, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='desop',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=False,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
""",
            encoding="utf-8",
        )

    print("Running PyInstaller...")
    if run([sys.executable, "-m", "PyInstaller", "--onefile", "--windowed", "--name", "desop", "desop/__main__.py"]):
        return 1

    # Copy binary
    dist_bin = Path("dist") / "desop"
    if not dist_bin.exists():
        print(f"Binary not found at {dist_bin}", file=sys.stderr)
        return 1

    dest_bin = out_dir / "desop.app"
    if dest_bin.exists():
        shutil.rmtree(dest_bin)

    # Create .app structure
    (out_dir / "desop.app" / "Contents" / "MacOS").mkdir(parents=True, exist_ok=True)
    (out_dir / "desop.app" / "Contents" / "Resources").mkdir(parents=True, exist_ok=True)
    (out_dir / "desop.app" / "Contents" / "Frameworks").mkdir(parents=True, exist_ok=True)

    # Info.plist
    plist_content = f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>Desop</string>
    <key>CFBundleDisplayName</key>
    <string>Desop</string>
    <key>CFBundleIdentifier</key>
    <string>cn.ntars.desop</string>
    <key>CFBundleVersion</key>
    <string>{version}</string>
    <key>CFBundleShortVersionString</key>
    <string>{version}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
</dict>
</plist>
"""
    (out_dir / "desop.app" / "Contents" / "Info.plist").write_text(plist_content, encoding="utf-8")

    # Copy binary to MacOS folder
    shutil.copy(dist_bin, out_dir / "desop.app" / "Contents" / "MacOS" / "desop")

    # Copy documentation
    for doc in ["README.md", "LICENSE", "LICENSE-MIT", "NOTICE", "LICENSING.md"]:
        src = root / doc
        if src.exists():
            shutil.copy(src, out_dir)

    # Copy scripts
    for script in ["install.sh", "uninstall.sh"]:
        src = root / "scripts" / script
        if src.exists():
            shutil.copy(src, out_dir)

    # Copy docs
    for doc in ["BUILD.md", "INSTALL.md", "ACCEPTANCE.md"]:
        src = root / "docs" / doc
        if src.exists():
            shutil.copy(src, out_dir)

    # Create distribution zip
    archive = root / "release" / f"desop-macos-{version}.zip"
    if archive.exists():
        archive.unlink()

    print("Creating distribution package...")
    subprocess.run(
        ["zip", "-r", str(archive), "desop-macos-0.1.0"],
        cwd=root / "release",
        check=True,
    )

    print(f"Output: {out_dir}")
    print(f"Archive: {archive}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
