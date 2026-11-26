from __future__ import annotations

import copy
import json
import os
import tempfile
from dataclasses import asdict, dataclass, field
from pathlib import Path
from typing import Any


@dataclass
class GeneralConfig:
    launch_at_login: bool = False
    show_menu_icon: bool = True
    language: str = "zh-CN"
    reduced_motion: bool = False

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class DockConfig:
    enabled: bool = True
    magnification: bool = True
    magnification_scale: float = 1.55
    icon_size: int = 52
    auto_hide: bool = False
    show_running: bool = True
    items: list[dict[str, str]] = field(
        default_factory=lambda: [
            {"kind": "finder", "name": "Finder"},
            {"kind": "app", "name": "Safari"},
            {"kind": "app", "name": "Notes"},
            {"kind": "app", "name": "Terminal"},
            {"kind": "app", "name": "Settings"},
        ]
    )

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class TopBarConfig:
    enabled: bool = True
    show_clock: bool = True
    show_battery: bool = True
    show_network: bool = True
    show_volume: bool = True
    show_frontmost_app: bool = True

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class WidgetConfig:
    id: str
    kind: str
    x: float
    y: float
    width: float = 300.0
    height: float = 180.0
    enabled: bool = True
    options: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class WidgetsConfig:
    enabled: bool = True
    snap_to_grid: bool = True
    grid_size: int = 8
    items: list[WidgetConfig] = field(
        default_factory=lambda: [
            WidgetConfig("clock", "clock", 40, 90, 320, 190),
            WidgetConfig("system", "system", 40, 300, 320, 210),
            WidgetConfig("calendar", "calendar", 380, 90, 340, 330),
            WidgetConfig("notes", "notes", 740, 90, 320, 260),
        ]
    )

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class IslandConfig:
    enabled: bool = True
    expanded: bool = False
    local_api: bool = True
    port: int = 17890
    event_ttl_seconds: int = 30

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class WallpaperConfig:
    enabled: bool = False
    path: str = ""
    fit: str = "cover"
    muted: bool = True
    loop: bool = True
    playback_rate: float = 1.0

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class AppearanceConfig:
    accent: str = "#FFB45A"
    tint: str = "#101419"
    glass_opacity: float = 0.76
    corner_radius: float = 24.0
    shadow: bool = True
    font_scale: float = 1.0

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


@dataclass
class Config:
    schema_version: int = 1
    general: GeneralConfig = field(default_factory=GeneralConfig)
    appearance: AppearanceConfig = field(default_factory=AppearanceConfig)
    wallpaper: WallpaperConfig = field(default_factory=WallpaperConfig)
    dock: DockConfig = field(default_factory=DockConfig)
    topbar: TopBarConfig = field(default_factory=TopBarConfig)
    widgets: WidgetsConfig = field(default_factory=WidgetsConfig)
    island: IslandConfig = field(default_factory=IslandConfig)

    def enabled_components(self) -> list[str]:
        result: list[str] = []
        for name, enabled in (
            ("wallpaper", self.wallpaper.enabled),
            ("dock", self.dock.enabled),
            ("topbar", self.topbar.enabled),
            ("widgets", self.widgets.enabled),
            ("island", self.island.enabled),
        ):
            if enabled:
                result.append(name)
        return result

    def set_component(self, name: str, enabled: bool) -> None:
        if name not in {"wallpaper", "dock", "topbar", "widgets", "island"}:
            raise ValueError(f"unknown component: {name}")
        getattr(self, name).enabled = enabled

    def validate(self) -> None:
        if self.schema_version != 1:
            raise ValueError(f"unsupported schema version {self.schema_version}")
        if self.dock.icon_size < 32 or self.dock.icon_size > 96:
            raise ValueError("dock.icon_size must be between 32 and 96")
        if not 1.0 <= self.dock.magnification_scale <= 2.2:
            raise ValueError("dock.magnification_scale must be between 1.0 and 2.2")
        if not 0.35 <= self.appearance.glass_opacity <= 1.0:
            raise ValueError("appearance.glass_opacity must be between 0.35 and 1.0")
        if not 4 <= self.widgets.grid_size <= 64:
            raise ValueError("widgets.grid_size must be between 4 and 64")
        if not 1024 <= self.island.port <= 65535:
            raise ValueError("island.port must be between 1024 and 65535")
        seen: set[str] = set()
        for widget in self.widgets.items:
            if not widget.id or widget.id in seen:
                raise ValueError(f"duplicate or empty widget id: {widget.id}")
            seen.add(widget.id)
            if widget.kind not in {"clock", "system", "calendar", "notes", "weather"}:
                raise ValueError(f"unsupported widget kind: {widget.kind}")
        if self.wallpaper.fit not in {"cover", "contain", "fill", "stretch"}:
            raise ValueError("wallpaper.fit must be cover, contain, fill or stretch")

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> "Config":
        cfg = cls()
        for section_name in (
            "general",
            "appearance",
            "wallpaper",
            "dock",
            "topbar",
            "island",
        ):
            section = raw.get(section_name, {})
            target = getattr(cfg, section_name)
            if isinstance(section, dict):
                for key, value in section.items():
                    if hasattr(target, key):
                        setattr(target, key, value)
        widgets = raw.get("widgets", {})
        if isinstance(widgets, dict):
            cfg.widgets.enabled = bool(widgets.get("enabled", cfg.widgets.enabled))
            cfg.widgets.snap_to_grid = bool(
                widgets.get("snap_to_grid", cfg.widgets.snap_to_grid)
            )
            cfg.widgets.grid_size = int(widgets.get("grid_size", cfg.widgets.grid_size))
            items = widgets.get("items")
            if isinstance(items, list):
                parsed: list[WidgetConfig] = []
                for item in items:
                    if not isinstance(item, dict) or "id" not in item or "kind" not in item:
                        continue
                    parsed.append(
                        WidgetConfig(
                            id=str(item["id"]),
                            kind=str(item["kind"]),
                            x=float(item.get("x", 40)),
                            y=float(item.get("y", 90)),
                            width=float(item.get("width", 300)),
                            height=float(item.get("height", 180)),
                            enabled=bool(item.get("enabled", True)),
                            options=copy.deepcopy(item.get("options", {})),
                        )
                    )
                if parsed:
                    cfg.widgets.items = parsed
        cfg.validate()
        return cfg


def config_dir() -> Path:
    override = os.environ.get("DESOP_CONFIG_DIR")
    if override:
        return Path(override).expanduser()
    return Path.home() / "Library" / "Application Support" / "Desop"


def config_path() -> Path:
    return config_dir() / "config.json"


def load(path: Path | None = None) -> Config:
    target = Path(path) if path else config_path()
    if not target.exists():
        return Config()
    try:
        raw = json.loads(target.read_text(encoding="utf-8"))
        if not isinstance(raw, dict):
            raise ValueError("root must be an object")
        return Config.from_dict(raw)
    except (OSError, ValueError, TypeError, json.JSONDecodeError):
        backup = target.with_suffix(".invalid.json")
        try:
            backup.write_text(target.read_text(encoding="utf-8"), encoding="utf-8")
        except OSError:
            pass
        cfg = Config()
        save(cfg, target)
        return cfg


def save(cfg: Config, path: Path | None = None) -> Path:
    cfg.validate()
    target = Path(path) if path else config_path()
    target.parent.mkdir(parents=True, exist_ok=True)
    payload = json.dumps(cfg.to_dict(), ensure_ascii=False, indent=2) + "\n"
    fd, temporary = tempfile.mkstemp(prefix=".config-", suffix=".json", dir=target.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write(payload)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temporary, target)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)
    return target
