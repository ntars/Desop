"""Tests for desop.layout module."""

from __future__ import annotations

import pytest

from desop.layout import (
    clamp,
    smoothstep,
    dock_icon_scales,
    dock_positions,
    snap,
    fit_rect,
    island_rect,
    human_bytes,
    lerp,
    ease_out_cubic,
)
from desop.models import Rect


class TestClamp:
    def test_clamp_within_range(self):
        assert clamp(5, 0, 10) == 5

    def test_clamp_below_range(self):
        assert clamp(-1, 0, 10) == 0

    def test_clamp_above_range(self):
        assert clamp(15, 0, 10) == 10

    def test_clamp_reversed_range(self):
        assert clamp(5, 10, 0) == 10


class TestSmoothstep:
    def test_smoothstep_0(self):
        assert smoothstep(0.0) == 0.0

    def test_smoothstep_1(self):
        assert smoothstep(1.0) == 1.0

    def test_smoothstep_middle(self):
        assert smoothstep(0.5) == 0.5

    def test_smoothstep_clamped(self):
        assert smoothstep(-1.0) == 0.0
        assert smoothstep(2.0) == 1.0


class TestDockIconScales:
    def test_no_magnification(self):
        scales = dock_icon_scales(5, None, 48, 0, False)
        assert len(scales) == 5
        assert all(s == 1.0 for s in scales)

    def test_with_magnification(self):
        # With magnification enabled, pointer at center should give max scale
        scales = dock_icon_scales(5, 240, 48, 0, True)
        assert len(scales) == 5
        # At least center icon should have scale > 1.0 when pointer is close
        assert any(s > 1.0 for s in scales)

    def test_magnification_scale_limit(self):
        scales = dock_icon_scales(5, 240, 48, 0, True, max_scale=1.5)
        assert scales[2] <= 1.5


class TestDockPositions:
    def test_single_icon(self):
        result, bounds = dock_positions(1, 1920, 1080, 48, 52, 8, None, False, 1.5)
        assert len(result) == 1
        assert result[0].width == 52

    def test_multiple_icons(self):
        result, bounds = dock_positions(5, 1920, 1080, 48, 52, 8, None, False, 1.5)
        assert len(result) == 5
        for rect in result:
            assert rect.width > 0


class TestSnap:
    def test_snap_to_grid(self):
        assert snap(25, 8) == 24
        assert snap(30, 8) == 32

    def test_snap_no_grid(self):
        assert snap(25, 1) == 25
        assert snap(25, 0) == 25


class TestFitRect:
    def test_fit_inside(self):
        rect = Rect(0, 0, 100, 100)
        bounds = Rect(0, 0, 200, 200)
        result = fit_rect(rect, bounds)
        assert result.width == 100
        assert result.height == 100

    def test_fit_exceeds_bounds(self):
        rect = Rect(0, 0, 300, 300)
        bounds = Rect(0, 0, 200, 200)
        result = fit_rect(rect, bounds)
        assert result.width <= 200
        assert result.height <= 200


class TestIslandRect:
    def test_default_size(self):
        rect = island_rect(1920, 0, False, False)
        assert rect.width == 96
        assert rect.height == 30

    def test_expanded_size(self):
        rect = island_rect(1920, 0, True, False)
        assert rect.width == 420

    def test_expanded_true(self):
        rect = island_rect(1920, 0, False, True)
        assert rect.width == 420


class TestHumanBytes:
    def test_bytes(self):
        assert human_bytes(1024) == "1.0 KB"

    def test_kb(self):
        assert human_bytes(1024 * 1024) == "1.0 MB"

    def test_mb(self):
        assert human_bytes(1024 * 1024 * 1024) == "1.0 GB"


class TestLerp:
    def test_lerp_start(self):
        assert lerp(0, 10, 0.0) == 0

    def test_lerp_end(self):
        assert lerp(0, 10, 1.0) == 10

    def test_lerp_middle(self):
        assert lerp(0, 10, 0.5) == 5

    def test_lerp_clamped(self):
        assert lerp(0, 10, -1.0) == 0
        assert lerp(0, 10, 2.0) == 10


class TestEaseOutCubic:
    def test_ease_start(self):
        assert ease_out_cubic(0.0) == 0.0

    def test_ease_end(self):
        assert ease_out_cubic(1.0) == 1.0

    def test_ease_middle(self):
        assert ease_out_cubic(0.5) == 0.875


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
