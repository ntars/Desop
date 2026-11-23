from __future__ import annotations

import math

from .models import Rect


def clamp(value: float, low: float, high: float) -> float:
    if high < low:
        return low
    return min(max(value, low), high)


def smoothstep(value: float) -> float:
    value = clamp(value, 0.0, 1.0)
    return value * value * (3.0 - 2.0 * value)


def dock_icon_scales(
    count: int,
    pointer_x: float | None,
    item_width: float,
    start_x: float,
    magnification: bool,
    max_scale: float = 1.55,
    influence: float = 2.2,
) -> list[float]:
    if not magnification or pointer_x is None:
        return [1.0] * count
    scales: list[float] = []
    for index in range(count):
        center = start_x + index * item_width + item_width / 2
        distance = abs(pointer_x - center) / max(item_width * influence, 1.0)
        lift = 1.0 - smoothstep(distance)
        scales.append(1.0 + (max_scale - 1.0) * lift)
    return scales


def dock_positions(
    count: int,
    screen_width: float,
    screen_height: float,
    bottom_margin: float,
    icon_size: float,
    gap: float,
    pointer_x: float | None,
    magnification: bool,
    max_scale: float,
) -> tuple[list[Rect], Rect]:
    if count <= 0:
        return [], Rect(screen_width / 2, screen_height - bottom_margin, 0, icon_size)
    slot = icon_size + gap
    total = count * icon_size + (count - 1) * gap
    start = (screen_width - total) / 2
    scales = dock_icon_scales(
        count, pointer_x, slot, start, magnification, max_scale
    )
    widths = [icon_size * scale for scale in scales]
    total_scaled = sum(widths) + (count - 1) * gap
    cursor = (screen_width - total_scaled) / 2
    rectangles: list[Rect] = []
    for width in widths:
        top = screen_height - bottom_margin - width
        rectangles.append(Rect(cursor, top, width, width))
        cursor += width + gap
    bounds = Rect(
        (screen_width - total_scaled) / 2 - 18,
        screen_height - bottom_margin - icon_size * max_scale - 24,
        total_scaled + 36,
        icon_size * max_scale + 42,
    )
    return rectangles, bounds


def snap(value: float, grid: int) -> float:
    if grid <= 1:
        return value
    return round(value / grid) * grid


def fit_rect(rect: Rect, available: Rect) -> Rect:
    width = min(rect.width, available.width)
    height = min(rect.height, available.height)
    x = clamp(rect.x, available.x, available.right - width)
    y = clamp(rect.y, available.y, available.bottom - height)
    return Rect(x, y, width, height)


def island_rect(
    screen_width: float, top: float, active: bool, expanded: bool
) -> Rect:
    width, height = (420.0, 104.0) if active or expanded else (96.0, 30.0)
    return Rect((screen_width - width) / 2, top + 8, width, height)


def human_bytes(value: float) -> str:
    units = ("B", "KB", "MB", "GB", "TB")
    number = float(max(value, 0))
    for unit in units:
        if number < 1024 or unit == units[-1]:
            return f"{number:.1f} {unit}"
        number /= 1024
    return f"{number:.1f} TB"


def lerp(start: float, end: float, factor: float) -> float:
    return start + (end - start) * clamp(factor, 0.0, 1.0)


def ease_out_cubic(value: float) -> float:
    value = clamp(value, 0.0, 1.0)
    return 1.0 - math.pow(1.0 - value, 3)
