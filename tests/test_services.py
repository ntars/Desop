"""Tests for desop.services module."""

from __future__ import annotations

import pytest

from desop.models import IslandEvent


class TestIslandController:
    def test_default_expanded_false(self):
        from desop.island import IslandController
        controller = IslandController()
        snap = controller.snapshot()
        assert snap.expanded is False

    def test_push_event(self):
        from desop.island import IslandController
        controller = IslandController()
        event = IslandEvent(kind="media", title="Test", progress=50)
        controller.push(event)
        snap = controller.snapshot()
        assert snap.active is True
        assert snap.title == "Test"

    def test_push_invalid_event(self):
        from desop.island import IslandController
        controller = IslandController()
        
        with pytest.raises(ValueError, match="title must not be empty"):
            controller.push(IslandEvent(kind="media", title="", progress=50))

        with pytest.raises(ValueError, match="unsupported event kind"):
            controller.push(IslandEvent(kind="invalid", title="Test", progress=50))

        with pytest.raises(ValueError, match="progress must be 0-100"):
            controller.push(IslandEvent(kind="media", title="Test", progress=150))

    def test_dismiss(self):
        from desop.island import IslandController
        controller = IslandController()
        event = IslandEvent(kind="media", title="Test", progress=50)
        controller.push(event)
        controller.dismiss()
        snap = controller.snapshot()
        assert snap.active is False

    def test_set_expanded(self):
        from desop.island import IslandController
        controller = IslandController()
        controller.set_expanded(True)
        snap = controller.snapshot()
        assert snap.expanded is True

    def test_snapshot_expires(self):
        from desop.island import IslandController
        import time
        controller = IslandController(ttl_seconds=1)
        event = IslandEvent(kind="media", title="Test", progress=50, expires_at=time.time() - 1)
        controller.push(event)
        snap = controller.snapshot()
        assert snap.active is False


if __name__ == "__main__":
    pytest.main([__file__, "-v"])


class TestSystemMonitor:
    def test_init(self):
        from desop.services import SystemMonitor
        monitor = SystemMonitor()
        assert hasattr(monitor, "_stats")
        assert hasattr(monitor, "_foreground")


class TestProcessRegistry:
    def test_init(self):
        from desop.services import ProcessRegistry
        registry = ProcessRegistry()
        assert hasattr(registry, "_psutil")


class TestApplicationResolver:
    def test_init(self):
        from desop.services import ApplicationResolver
        resolver = ApplicationResolver()
        assert hasattr(resolver, "_cache")

    def test_locate_finder(self):
        from desop.services import ApplicationResolver
        resolver = ApplicationResolver()
        path = resolver.locate("Finder")
        assert "Finder.app" in path
