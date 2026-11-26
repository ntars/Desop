"""Tests for desop.models module."""

from __future__ import annotations

import pytest

from desop.models import Rect, IslandEvent, IslandSnapshot, SystemStats


class TestRect:
    def test_creation(self):
        r = Rect(10, 20, 300, 200)
        assert r.x == 10
        assert r.y == 20
        assert r.width == 300
        assert r.height == 200

    def test_right(self):
        r = Rect(10, 20, 300, 200)
        assert r.right == 310

    def test_bottom(self):
        r = Rect(10, 20, 300, 200)
        assert r.bottom == 220

    def test_contains(self):
        r = Rect(0, 0, 100, 100)
        assert r.contains(50, 50) is True
        assert r.contains(-1, 50) is False
        assert r.contains(100, 50) is False
        assert r.contains(50, -1) is False
        assert r.contains(50, 100) is False

    def test_intersects(self):
        a = Rect(0, 0, 100, 100)
        b = Rect(50, 50, 100, 100)
        c = Rect(200, 200, 100, 100)
        
        assert a.intersects(b) is True
        assert a.intersects(c) is False


class TestIslandEvent:
    def test_creation(self):
        event = IslandEvent(kind="media", title="Test", progress=50)
        assert event.kind == "media"
        assert event.title == "Test"
        assert event.progress == 50

    def test_validate_empty_title(self):
        event = IslandEvent(kind="media", title="", progress=50)
        with pytest.raises(ValueError, match="title must not be empty"):
            event.validate()

    def test_validate_invalid_kind(self):
        event = IslandEvent(kind="invalid", title="Test", progress=50)
        with pytest.raises(ValueError, match="unsupported event kind"):
            event.validate()

    def test_validate_progress_bounds(self):
        event = IslandEvent(kind="media", title="Test", progress=-1)
        with pytest.raises(ValueError, match="progress must be 0-100"):
            event.validate()
        
        event2 = IslandEvent(kind="media", title="Test", progress=101)
        with pytest.raises(ValueError, match="progress must be 0-100"):
            event2.validate()


class TestIslandSnapshot:
    def test_default_values(self):
        snap = IslandSnapshot()
        assert snap.active is False
        assert snap.expanded is False

    def test_with_data(self):
        snap = IslandSnapshot(
            active=True,
            expanded=True,
            kind="media",
            title="Test Song",
            detail="Artist Name",
            progress=75,
        )
        assert snap.active is True
        assert snap.expanded is True
        assert snap.title == "Test Song"
        assert snap.progress == 75


class TestSystemStats:
    def test_default_values(self):
        stats = SystemStats()
        assert stats.cpu_percent == 0.0
        assert stats.memory_percent == 0.0
        assert stats.battery_percent == -1


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
