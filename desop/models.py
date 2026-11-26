from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Callable


@dataclass(frozen=True)
class Rect:
    x: float
    y: float
    width: float
    height: float

    @property
    def right(self) -> float:
        return self.x + self.width

    @property
    def bottom(self) -> float:
        return self.y + self.height

    def contains(self, x: float, y: float) -> bool:
        return self.x <= x < self.right and self.y <= y < self.bottom

    def intersects(self, other: "Rect") -> bool:
        return (self.x < other.right and self.right > other.x and
                self.y < other.bottom and self.bottom > other.y)


@dataclass
class IslandEvent:
    kind: str = ""
    title: str = ""
    detail: str = ""
    progress: int = 0
    expires_at: float = 0.0

    def validate(self) -> None:
        if not self.title.strip():
            raise ValueError("title must not be empty")
        if self.kind not in {"media", "timer", "download", "custom"}:
            raise ValueError("unsupported event kind")
        if not 0 <= self.progress <= 100:
            raise ValueError("progress must be 0-100")


@dataclass
class IslandSnapshot:
    active: bool = False
    expanded: bool = False
    kind: str = ""
    title: str = ""
    detail: str = ""
    progress: int = 0


@dataclass
class SystemStats:
    cpu_percent: float = 0.0
    memory_percent: float = 0.0
    disk_percent: float = 0.0
    network_down: int = 0
    network_up: int = 0
    battery_percent: int = -1
    battery_charging: bool = False
    network_online: bool = False
    volume: int = -1


@dataclass
class ForegroundApplication:
    name: str = ""
    bundle_id: str = ""


@dataclass
class DockItem:
    name: str
    kind: str
    path: str = ""
    running: bool = False
    icon: object | None = None


EventCallback = Callable[[IslandEvent], None]
