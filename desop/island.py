from __future__ import annotations

import threading
import time

from .models import IslandEvent, IslandSnapshot


class IslandController:
    def __init__(self, expanded: bool = False, ttl_seconds: int = 30) -> None:
        self._expanded = expanded
        self._ttl_seconds = ttl_seconds
        self._event: IslandEvent | None = None
        self._lock = threading.Lock()

    def push(self, event: IslandEvent) -> None:
        if not event.title.strip():
            raise ValueError("title must not be empty")
        if event.kind not in {"media", "timer", "download", "custom"}:
            raise ValueError("unsupported event kind")
        if not 0 <= event.progress <= 100:
            raise ValueError("progress must be 0-100")
        if event.expires_at <= 0:
            event.expires_at = time.time() + self._ttl_seconds
        with self._lock:
            self._event = event
            self._expanded = True

    def dismiss(self) -> None:
        with self._lock:
            self._event = None
            self._expanded = False

    def set_expanded(self, expanded: bool) -> None:
        with self._lock:
            self._expanded = expanded

    def snapshot(self) -> IslandSnapshot:
        with self._lock:
            if self._event and self._event.expires_at <= time.time():
                self._event = None
            event = self._event
            return IslandSnapshot(
                active=event is not None,
                expanded=self._expanded,
                kind=event.kind if event else "",
                title=event.title if event else "",
                detail=event.detail if event else "",
                progress=event.progress if event else 0,
            )
