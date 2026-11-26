"""Tests for desop.island module."""

from __future__ import annotations

import pytest
import time

from desop.models import IslandEvent
from desop.island import IslandController


class TestIslandController:
    def test_default_state(self):
        controller = IslandController()
        snap = controller.snapshot()
        assert snap.expanded is False
        assert snap.active is False

    def test_push_event(self):
        controller = IslandController()
        event = IslandEvent(kind="media", title="Test Song", progress=50)
        controller.push(event)
        
        snap = controller.snapshot()
        assert snap.active is True
        assert snap.title == "Test Song"
        assert snap.progress == 50

    def test_push_validation_empty_title(self):
        controller = IslandController()
        event = IslandEvent(kind="media", title="", progress=50)
        
        with pytest.raises(ValueError, match="title must not be empty"):
            controller.push(event)

    def test_push_validation_invalid_kind(self):
        controller = IslandController()
        event = IslandEvent(kind="invalid", title="Test", progress=50)
        
        with pytest.raises(ValueError, match="unsupported event kind"):
            controller.push(event)

    def test_push_validation_progress_bounds(self):
        controller = IslandController()
        
        event1 = IslandEvent(kind="media", title="Test", progress=-1)
        with pytest.raises(ValueError, match="progress must be 0-100"):
            controller.push(event1)
        
        event2 = IslandEvent(kind="media", title="Test", progress=150)
        with pytest.raises(ValueError, match="progress must be 0-100"):
            controller.push(event2)

    def test_dismiss(self):
        controller = IslandController()
        event = IslandEvent(kind="media", title="Test", progress=50)
        controller.push(event)
        
        snap = controller.snapshot()
        assert snap.active is True
        
        controller.dismiss()
        snap2 = controller.snapshot()
        assert snap2.active is False

    def test_set_expanded(self):
        controller = IslandController()
        
        snap1 = controller.snapshot()
        assert snap1.expanded is False
        
        controller.set_expanded(True)
        snap2 = controller.snapshot()
        assert snap2.expanded is True

    def test_event_expiry(self):
        controller = IslandController(ttl_seconds=1)
        event = IslandEvent(
            kind="media",
            title="Test",
            progress=50,
            expires_at=time.time() - 1
        )
        controller.push(event)
        
        snap = controller.snapshot()
        assert snap.active is False

    def test_snapshot_reset_after_dismiss(self):
        controller = IslandController()
        event = IslandEvent(kind="media", title="Test", progress=50)
        controller.push(event)
        
        controller.dismiss()
        
        snap = controller.snapshot()
        assert snap.active is False
        assert snap.title == ""
        assert snap.progress == 0

    def test_multiple_push(self):
        controller = IslandController()
        
        event1 = IslandEvent(kind="media", title="First", progress=30)
        controller.push(event1)
        
        snap1 = controller.snapshot()
        assert snap1.title == "First"
        
        event2 = IslandEvent(kind="download", title="Second", progress=70)
        controller.push(event2)
        
        snap2 = controller.snapshot()
        assert snap2.title == "Second"
        assert snap2.kind == "download"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
