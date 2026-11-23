from __future__ import annotations

import argparse
import json
import os
import signal
import subprocess
import sys
import time
import urllib.error
from pathlib import Path

from . import __version__
from . import config as config_module
from .services import request_json, system_summary


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="desop",
        description="Desop - macOS desktop layer written in Python",
    )
    parser.add_argument("--version", action="version", version=__version__)
    parser.add_argument(
        "action",
        nargs="?",
        default="start",
        choices=(
            "start",
            "run",
            "stop",
            "status",
            "enable",
            "disable",
            "settings",
            "selftest",
            "api-event",
            "doctor",
            "reset-config",
            "package",
        ),
    )
    parser.add_argument("component", nargs="?")
    parser.add_argument("--config", type=Path)
    parser.add_argument("--seconds", type=int, default=10)
    parser.add_argument("--json", action="store_true")
    return parser


def _print_config(cfg: config_module.Config, path: Path) -> None:
    print(f"config: {path}")
    print(f"enabled: {', '.join(cfg.enabled_components())}")
    print(f"version: {__version__}")


def run_commands(args: argparse.Namespace) -> int:
    path = args.config or config_module.config_path()
    if args.action == "api-event":
        print(
            "Desop island API listens on 127.0.0.1:17890 by default.\n"
            "\n"
            "POST /event\n"
            '  {"kind":"download","title":"Update","detail":"Downloading","progress":42}\n'
            "\n"
            "POST /dismiss\n"
            "GET  /health\n"
            "GET  /status\n"
        )
        return 0
    if args.action == "doctor":
        payload = system_summary()
        try:
            import PySide6.QtWidgets  # noqa: F401

            payload["pyside6"] = True
        except ImportError:
            payload["pyside6"] = False
        payload["config"] = str(path)
        payload["config_exists"] = path.exists()
        try:
            payload["island_api"] = request_json(
                f"http://127.0.0.1:17890/health", timeout=1.0
            )
        except (OSError, ValueError, urllib.error.URLError):
            payload["island_api"] = None
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 0
    if args.action == "reset-config":
        cfg = config_module.Config()
        config_module.save(cfg, path)
        _print_config(cfg, path)
        return 0
    if args.action in {"enable", "disable"}:
        if not args.component:
            raise SystemExit(f"{args.action} requires a component")
        cfg = config_module.load(path)
        for name in args.component.split(","):
            cfg.set_component(name.strip().lower(), args.action == "enable")
        config_module.save(cfg, path)
        _print_config(cfg, path)
        return 0
    if args.action == "status":
        cfg = config_module.load(path)
        _print_config(cfg, path)
        return 0
    if args.action == "settings":
        cfg = config_module.load(path)
        from .ui import DesktopApplication

        app = DesktopApplication(cfg, path, settings_only=True)
        app.run()
        return 0
    if args.action == "stop":
        pid_file = config_module.config_dir() / "desop.pid"
        try:
            pid = int(pid_file.read_text(encoding="utf-8").strip())
            os.kill(pid, signal.SIGTERM)
            print(f"stop signal sent to pid {pid}")
            return 0
        except (OSError, ValueError, FileNotFoundError):
            try:
                request_json("http://127.0.0.1:17890/health", timeout=0.5)
                print("Desop is running but has no stop endpoint; use Activity Monitor.")
            except (OSError, ValueError, urllib.error.URLError):
                print("Desop is not running.")
            return 0
    if args.action == "selftest":
        return selftest(args.seconds)
    if args.action == "package":
        return package()
    return 0


def selftest(seconds: int) -> int:
    start = time.monotonic()
    deadline = start + max(seconds, 1)
    result: dict[str, object] = {
        "started": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "seconds": seconds,
        "checks": {},
    }
    cfg = config_module.load()
    checks: dict[str, bool] = {
        "config_valid": True,
        "pyside6_import": False,
        "island_api": False,
    }
    try:
        import PySide6.QtWidgets  # noqa: F401

        checks["pyside6_import"] = True
    except ImportError:
        pass
    try:
        health = request_json("http://127.0.0.1:17890/health", timeout=0.5)
        checks["island_api"] = health.get("ok") is True
    except (OSError, ValueError, urllib.error.URLError):
        pass
    while time.monotonic() < deadline:
        time.sleep(0.05)
    result["checks"] = checks
    result["elapsed"] = round(time.monotonic() - start, 3)
    result["components"] = cfg.enabled_components()
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if all(checks.values()) else 1


def package() -> int:
    script = Path(__file__).resolve().parent.parent / "scripts" / "build_desop.py"
    if not script.exists():
        print(f"packaging script not found: {script}", file=sys.stderr)
        return 1
    result = subprocess.run([sys.executable, str(script)], check=False)
    return result.returncode


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    if args.action in {"start", "run"}:
        cfg = config_module.load(args.config)
        from .ui import DesktopApplication

        app = DesktopApplication(cfg, args.config or config_module.config_path())
        app.run()
        return 0
    return run_commands(args)
