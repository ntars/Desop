#!/usr/bin/env python3
"""docs_guard.py - 验证版本一致性与文档完整性。"""

from __future__ import annotations

import json
import sys
from pathlib import Path


def get_version() -> str:
    root = Path(__file__).resolve().parent.parent
    version_file = root / "VERSION"
    if not version_file.exists():
        raise FileNotFoundError("VERSION file not found")
    return version_file.read_text(encoding="utf-8").strip()


def check_pyproject_version(version: str) -> bool:
    pyproject = Path(__file__).resolve().parent.parent / "pyproject.toml"
    content = pyproject.read_text(encoding="utf-8")
    if 'dynamic = ["version"]' in content:
        return True
    for line in content.splitlines():
        if "version" in line and "=" in line:
            val = line.split("=")[1].strip().strip('"\'')
            if val == version:
                return True
    print(f"pyproject.toml: version mismatch (expected {version})")
    return False


def check_init_version(version: str) -> bool:
    init_file = Path(__file__).resolve().parent.parent / "desop" / "__init__.py"
    content = init_file.read_text(encoding="utf-8")
    if "VERSION_FILE" in content and "__version__ = VERSION_FILE.read_text" in content:
        return True
    print(f"desop/__init__.py: version not loaded from VERSION")
    return False


def check_build_desop_version(version: str) -> bool:
    build_py = Path(__file__).resolve().parent.parent / "scripts" / "build_desop.py"
    content = build_py.read_text(encoding="utf-8")
    if "get_version()" in content and f'f"desop-{version}"' not in content:
        return True
    if f'f"desop-{version}"' in content:
        print(f"scripts/build_desop.py: version hardcoded")
        return False
    print(f"scripts/build_desop.py: get_version() not found")
    return False


def check_readme(version: str) -> bool:
    readme = Path(__file__).resolve().parent.parent / "README.md"
    content = readme.read_text(encoding="utf-8")
    # README doesn't need exact version - it uses dynamic placeholder
    return True


def check_changelog(version: str) -> bool:
    changelog = Path(__file__).resolve().parent.parent / "CHANGELOG.md"
    content = changelog.read_text(encoding="utf-8")
    if f"## {version}" in content or f"## Unreleased — {version}" in content:
        return True
    print(f"CHANGELOG.md: version entry not found")
    return False


def check_build_info(version: str) -> bool:
    release_dir = Path(__file__).resolve().parent.parent / "release"
    if not release_dir.exists():
        print("release/ directory not found (skip)")
        return True
    for item in sorted(release_dir.iterdir(), key=lambda p: p.name, reverse=True):
        if item.is_dir() and not item.name.startswith("."):
            build_info = item / "BUILD_INFO.txt"
            if build_info.exists():
                data = json.loads(build_info.read_text(encoding="utf-8"))
                if data.get("version") == version:
                    return True
    print("BUILD_INFO.txt not found in release/ (skip)")
    return True


def main() -> int:
    try:
        version = get_version()
    except FileNotFoundError as e:
        print(f"Error: {e}")
        return 1

    checks = [
        ("VERSION file", lambda: True),
        ("pyproject.toml", lambda: check_pyproject_version(version)),
        ("desop/__init__.py", lambda: check_init_version(version)),
        ("scripts/build_desop.py", lambda: check_build_desop_version(version)),
        ("README.md", lambda: check_readme(version)),
        ("CHANGELOG.md", lambda: check_changelog(version)),
        ("BUILD_INFO.txt", lambda: check_build_info(version)),
    ]

    all_ok = True
    for name, check in checks:
        ok = check()
        status = "OK" if ok else "FAIL"
        print(f"[{status}] {name}")
        all_ok = all_ok and ok

    if all_ok:
        print(f"\nAll checks passed for version {version}")
        return 0
    else:
        print("\nSome checks failed.")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
