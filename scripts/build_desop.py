#!/usr/bin/env python3
"""Build Desop for Windows."""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path


def run(argv: list[str], cwd=None) -> int:
    """Run a command and return its exit code."""
    result = subprocess.run(argv, cwd=cwd, check=False)
    return result.returncode


def get_version() -> str:
    """Read version from VERSION file."""
    root = Path(__file__).resolve().parent.parent
    version_file = root / "VERSION"
    return version_file.read_text(encoding="utf-8").strip()


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    os.chdir(root)

    version = get_version()
    out_dir = root / "release" / f"desop-{version}"
    out_dir.mkdir(parents=True, exist_ok=True)

    print(f"Building Desop {version} for Windows (amd64)")

    # Clean previous build
    dist_dir = root / "dist"
    if dist_dir.exists():
        shutil.rmtree(dist_dir)

    build_dir = root / "build"
    if build_dir.exists():
        shutil.rmtree(build_dir)

    # Build with PyInstaller
    print("Running PyInstaller...")
    if run([sys.executable, "-m", "PyInstaller", "--onefile", "--windowed",
            "--name", "desop", "--target-arch", "x86_64",
            "--add-data", f"desop{os.pathsep}desop",
            "desop/__main__.py"]):
        print("PyInstaller failed", file=sys.stderr)
        return 1

    # Copy binary
    dist_bin = Path("dist") / "desop.exe"
    if not dist_bin.exists():
        print(f"Binary not found at {dist_bin}", file=sys.stderr)
        return 1

    dest_bin = out_dir / "desop.exe"
    shutil.copy(dist_bin, dest_bin)
    print(f"Copied binary to {dest_bin}")

    # Create BUILD_INFO.txt
    build_info = {
        "version": version,
        "build_time": datetime.now(timezone.utc).isoformat(),
        "python": sys.version,
        "pyinstaller_version": subprocess.run(
            [sys.executable, "-m", "PyInstaller", "--version"],
            capture_output=True, text=True
        ).stdout.strip(),
        "commit": subprocess.run(
            ["git", "rev-parse", "HEAD"],
            capture_output=True, text=True
        ).stdout.strip()
    }
    (out_dir / "BUILD_INFO.txt").write_text(
        json.dumps(build_info, indent=2) + "\n",
        encoding="utf-8"
    )

    # Copy documentation
    for doc in ["README.md", "README_EN.md", "LICENSE", "LICENSE-MIT", 
                "NOTICE", "LICENSING.md", "CHANGELOG.md"]:
        src = root / doc
        if src.exists():
            shutil.copy(src, out_dir)

    # Copy scripts
    for script in ["install.ps1", "uninstall.ps1"]:
        src = root / "scripts" / script
        if src.exists():
            shutil.copy(src, out_dir)

    # Copy docs
    for doc in ["BUILD.md", "INSTALL.md", "ACCEPTANCE.md"]:
        src = root / "docs" / doc
        if src.exists():
            shutil.copy(src, out_dir)

    # Generate SHA256SUMS
    import hashlib
    sha256sums = []
    for f in out_dir.iterdir():
        if f.is_file():
            sha256 = hashlib.sha256(f.read_bytes()).hexdigest()
            sha256sums.append(f"{sha256}  {f.name}")
    (out_dir / "SHA256SUMS").write_text("\n".join(sha256sums) + "\n", encoding="utf-8")

    # Create distribution zip
    archive = root / "release" / f"desop-{version}.zip"
    if archive.exists():
        archive.unlink()

    print("Creating distribution package...")
    subprocess.run(
        ["zip", "-r", str(archive), f"desop-{version}"],
        cwd=root / "release",
        check=True,
    )

    print(f"Output: {out_dir}")
    print(f"Archive: {archive}")
    print("SHA256SUMS:")
    print((out_dir / "SHA256SUMS").read_text(encoding="utf-8"))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
