"""Tests for desop.config module."""

from __future__ import annotations

import json
import tempfile
from pathlib import Path

import pytest

from desop.config import Config, DockConfig, TopBarConfig, WidgetsConfig, IslandConfig


class TestDockConfig:
    def test_default_values(self):
        cfg = DockConfig()
        assert cfg.enabled is True
        assert cfg.magnification is True
        assert cfg.icon_size == 52

    def test_serialization(self):
        cfg = DockConfig(enabled=False, icon_size=64)
        data = cfg.to_dict()
        assert data["enabled"] is False
        assert data["icon_size"] == 64


class TestTopBarConfig:
    def test_default_values(self):
        cfg = TopBarConfig()
        assert cfg.enabled is True
        assert cfg.show_clock is True

    def test_serialization(self):
        cfg = TopBarConfig(enabled=False, show_volume=False)
        data = cfg.to_dict()
        assert data["enabled"] is False
        assert data["show_volume"] is False


class TestWidgetsConfig:
    def test_default_items(self):
        cfg = WidgetsConfig()
        assert len(cfg.items) == 4
        assert cfg.items[0].kind == "clock"
        assert cfg.snap_to_grid is True


class TestIslandConfig:
    def test_default_values(self):
        cfg = IslandConfig()
        assert cfg.enabled is True
        assert cfg.local_api is True
        assert cfg.port == 17890


class TestConfig:
    def test_default_config(self):
        cfg = Config()
        assert cfg.schema_version == 1
        assert len(cfg.enabled_components()) > 0

    def test_enabled_components(self):
        cfg = Config(
            dock=DockConfig(enabled=False),
            topbar=TopBarConfig(enabled=False),
        )
        comps = cfg.enabled_components()
        assert "dock" not in comps
        assert "topbar" not in comps
        assert "widgets" in comps

    def test_set_component(self):
        cfg = Config()
        cfg.set_component("dock", False)
        assert cfg.dock.enabled is False
        with pytest.raises(ValueError, match="unknown component"):
            cfg.set_component("invalid", True)

    def test_validate_fails_on_invalid_icon_size(self):
        cfg = Config()
        cfg.dock.icon_size = 20
        with pytest.raises(ValueError, match="dock.icon_size"):
            cfg.validate()

    def test_validate_fails_on_invalid_glass_opacity(self):
        cfg = Config()
        cfg.appearance.glass_opacity = 0.2
        with pytest.raises(ValueError, match="appearance.glass_opacity"):
            cfg.validate()

    def test_validate_fails_on_duplicate_widget_id(self):
        cfg = Config()
        widget1 = cfg.widgets.items[0]
        widget2 = cfg.widgets.items[1]
        widget1.id = "same"
        widget2.id = "same"
        with pytest.raises(ValueError, match="duplicate or empty widget id"):
            cfg.validate()

    def test_validate_fails_on_invalid_wallpaper_fit(self):
        cfg = Config()
        cfg.wallpaper.fit = "invalid"
        with pytest.raises(ValueError, match="wallpaper.fit must be cover"):
            cfg.validate()


class TestConfigPersistence:
    def test_save_and_load(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            from desop.config import save, load
            path = Path(tmpdir) / "config.json"
            cfg = Config()
            save(cfg, path)
            
            loaded = load(path)
            assert loaded.schema_version == 1

    def test_load_missing_file(self):
        from desop.config import load
        path = Path("/tmp/nonexistent_config.json")
        loaded = load(path)
        assert loaded.schema_version == 1


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
