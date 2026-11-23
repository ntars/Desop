from __future__ import annotations

import json
import os
import platform
import shutil
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
from datetime import datetime
from pathlib import Path
from typing import Any

from .models import ForegroundApplication, SystemStats
from . import __version__


def _run(argv: list[str], timeout: float = 2.0) -> str:
    try:
        result = subprocess.run(
            argv,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )
        return result.stdout.strip() if result.returncode == 0 else ""
    except (OSError, subprocess.SubprocessError):
        return ""


def _psutil():
    try:
        import psutil

        return psutil
    except ImportError:
        return None


class SystemMonitor:
    def __init__(self) -> None:
        self._psutil = _psutil()
        self._previous_net: tuple[float, int, int] | None = None
        self._lock = threading.Lock()
        self._stats = SystemStats()
        self._foreground = ForegroundApplication()
        self._started = False

    def start(self) -> None:
        if self._started:
            return
        self._started = True
        self.refresh()
        self._thread = threading.Thread(target=self._loop, name="desop-monitor", daemon=True)
        self._thread.start()

    def _loop(self) -> None:
        while True:
            self.refresh()
            time.sleep(1.0)

    def refresh(self) -> None:
        stats = SystemStats()
        if self._psutil:
            try:
                stats.cpu_percent = self._psutil.cpu_percent(interval=None)
                stats.memory_percent = self._psutil.virtual_memory().percent
                usage = shutil.disk_usage(Path.home())
                stats.disk_percent = usage.used / usage.total * 100 if usage.total else 0
                net = self._psutil.net_io_counters()
                now = time.monotonic()
                if net and self._previous_net:
                    elapsed = max(now - self._previous_net[0], 0.1)
                    stats.network_down = max(
                        int((net.bytes_recv - self._previous_net[1]) / elapsed), 0
                    )
                    stats.network_up = max(
                        int((net.bytes_sent - self._previous_net[2]) / elapsed), 0
                    )
                if net:
                    self._previous_net = (now, net.bytes_recv, net.bytes_sent)
                battery = self._psutil.sensors_battery()
                if battery:
                    stats.battery_percent = int(battery.percent)
                    stats.battery_charging = bool(battery.power_plugged)
                stats.network_online = bool(self._psutil.net_if_stats())
            except Exception:
                pass
        if not stats.network_online:
            result = subprocess.run(
                ["/sbin/route", "-n", "get", "default"],
                capture_output=True,
                text=True,
                timeout=1.5,
                check=False,
            )
            stats.network_online = result.returncode == 0 and "gateway:" in result.stdout
        stats.volume = self._volume()
        with self._lock:
            self._stats = stats
            self._foreground = self._frontmost()

    def _volume(self) -> int:
        output = _run(
            ["osascript", "-e", "output volume of (get volume settings)"], 1.0
        )
        try:
            return max(0, min(100, int(output)))
        except ValueError:
            return -1

    def _frontmost(self) -> ForegroundApplication:
        script = (
            'tell application "System Events"\n'
            "set frontProcess to first application process whose frontmost is true\n"
            "set processName to name of frontProcess\n"
            'set bundleIdentifier to bundle identifier of frontProcess\n'
            "return processName & linefeed & bundleIdentifier\n"
            "end tell"
        )
        output = _run(["osascript", "-e", script], 1.5)
        if not output:
            return ForegroundApplication()
        parts = output.splitlines()
        return ForegroundApplication(
            name=parts[0] if parts else "",
            bundle_id=parts[1] if len(parts) > 1 else "",
        )

    @property
    def stats(self) -> SystemStats:
        with self._lock:
            return SystemStats(**vars(self._stats))

    @property
    def foreground(self) -> ForegroundApplication:
        with self._lock:
            return ForegroundApplication(**vars(self._foreground))


class ProcessRegistry:
    def __init__(self) -> None:
        self._psutil = _psutil()

    def running_names(self) -> set[str]:
        if not self._psutil:
            output = _run(["pgrep", "-xl"], 1.0)
            names: set[str] = set()
            for line in output.splitlines():
                parts = line.split(maxsplit=1)
                if len(parts) == 2:
                    names.add(parts[1].lower())
            return names
        names = set()
        for process in self._psutil.process_iter(["name"]):
            try:
                if process.info["name"]:
                    names.add(process.info["name"].lower())
            except Exception:
                continue
        return names


class ApplicationResolver:
    ROOTS = (
        Path("/Applications"),
        Path.home() / "Applications",
        Path("/System/Applications"),
        Path("/System/Applications/Utilities"),
    )

    def __init__(self) -> None:
        self._cache: dict[str, str] = {}

    def locate(self, name: str) -> str:
        if name in self._cache:
            return self._cache[name]
        if name == "Finder":
            self._cache[name] = "/System/Library/CoreServices/Finder.app"
            return self._cache[name]
        for root in self.ROOTS:
            direct = root / f"{name}.app"
            if direct.exists():
                self._cache[name] = str(direct)
                return str(direct)
        for root in self.ROOTS:
            if not root.exists():
                continue
            for candidate in root.glob("*.app"):
                if candidate.stem.casefold() == name.casefold():
                    self._cache[name] = str(candidate)
                    return str(candidate)
        self._cache[name] = ""
        return ""

    def launch(self, name: str) -> bool:
        path = self.locate(name)
        if path:
            result = subprocess.run(
                ["/usr/bin/open", "-a", path],
                capture_output=True,
                text=True,
                timeout=5,
                check=False,
            )
            return result.returncode == 0
        return False


class PersistenceService:
    def __init__(self, config_path: Path) -> None:
        self.config_path = config_path

    def install_launch_agent(self) -> Path:
        agents = Path.home() / "Library" / "LaunchAgents"
        agents.mkdir(parents=True, exist_ok=True)
        plist = agents / "cn.ntars.desop.plist"
        executable = Path(sys.executable)
        if getattr(sys, "frozen", False):
            command = [str(executable), "--run"]
        else:
            command = [
                str(Path.home() / ".local" / "bin" / "desop"),
            ]
        args = "".join(f"    <string>{part}</string>\n" for part in command)
        plist.write_text(
            '<?xml version="1.0" encoding="UTF-8"?>\n'
            '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" '
            '"http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n'
            '<plist version="1.0">\n'
            "<dict>\n"
            "  <key>Label</key>\n"
            "  <string>cn.ntars.desop</string>\n"
            "  <key>ProgramArguments</key>\n"
            "  <array>\n"
            f"{args}"
            "  </array>\n"
            "  <key>RunAtLoad</key>\n"
            "  <true/>\n"
            "  <key>KeepAlive</key>\n"
            "  <false/>\n"
            "<key>ProcessType</key>\n"
            "  <string>Interactive</string>\n"
            "</dict>\n"
            "</plist>\n",
            encoding="utf-8",
        )
        return plist

    def remove_launch_agent(self) -> None:
        plist = Path.home() / "Library" / "LaunchAgents" / "cn.ntars.desop.plist"
        try:
            plist.unlink()
        except FileNotFoundError:
            pass


class LocalAPIServer:
    def __init__(self, host: str, port: int, on_event, on_dismiss, on_expand) -> None:
        from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
        import json as json_module

        parent = self

        class Handler(BaseHTTPRequestHandler):
            server_version = "Desop/" + __version__

            def log_message(self, format: str, *args: Any) -> None:
                return

            def _json(self, status: int, payload: dict[str, Any]) -> None:
                body = json_module.dumps(payload, ensure_ascii=False).encode("utf-8")
                self.send_response(status)
                self.send_header("Content-Type", "application/json; charset=utf-8")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)

            def do_GET(self) -> None:
                if self.path == "/health":
                    self._json(200, {"ok": True, "name": "Desop", "version": __version__})
                elif self.path == "/status":
                    self._json(200, {"ok": True, "components": parent.components})
                else:
                    self._json(404, {"ok": False, "error": "not found"})

            def do_POST(self) -> None:
                if self.path == "/dismiss":
                    on_dismiss()
                    self._json(200, {"ok": True})
                    return
                if self.path == "/event/collapse":
                    on_expand(False)
                    self._json(200, {"ok": True})
                    return
                if self.path != "/event":
                    self._json(404, {"ok": False, "error": "not found"})
                    return
                try:
                    length = int(self.headers.get("Content-Length", "0"))
                    if length <= 0 or length > 16384:
                        raise ValueError("request body is empty or too large")
                    raw = json_module.loads(self.rfile.read(length))
                    if not isinstance(raw, dict):
                        raise ValueError("event must be an object")
                    kind = str(raw.get("kind", "custom")).strip().lower()
                    if kind not in {"media", "timer", "download", "custom"}:
                        raise ValueError("unsupported event kind")
                    title = str(raw.get("title", "")).strip()
                    if not title:
                        raise ValueError("event title must not be empty")
                    progress = int(raw.get("progress", 0))
                    if not 0 <= progress <= 100:
                        raise ValueError("progress must be 0-100")
                    from .models import IslandEvent

                    event = IslandEvent(
                        kind=kind,
                        title=title[:120],
                        detail=str(raw.get("detail", ""))[:240],
                        progress=progress,
                        expires_at=float(raw.get("expires_at", 0)),
                    )
                    parent.on_event(event)
                    self._json(202, {"ok": True})
                except (ValueError, TypeError, json_module.JSONDecodeError) as exc:
                    self._json(400, {"ok": False, "error": str(exc)})

        self.components: list[str] = []
        self._server = ThreadingHTTPServer((host, port), Handler)
        self._thread = threading.Thread(target=self._server.serve_forever, daemon=True)
        self.on_event = on_event

    @property
    def address(self) -> tuple[str, int]:
        host, port = self._server.server_address[:2]
        return str(host), int(port)

    def start(self) -> None:
        self._thread.start()

    def stop(self) -> None:
        self._server.shutdown()
        self._server.server_close()


def request_json(url: str, timeout: float = 2.0) -> dict[str, Any]:
    request = urllib.request.Request(url, headers={"Accept": "application/json"})
    with urllib.request.urlopen(request, timeout=timeout) as response:
        payload = json.loads(response.read().decode("utf-8"))
        if not isinstance(payload, dict):
            raise ValueError("response is not an object")
        return payload


def system_summary() -> dict[str, Any]:
    return {
        "python": platform.python_version(),
        "platform": platform.platform(),
        "machine": platform.machine(),
        "version": __version__,
        "pid": os.getpid(),
    }
