from __future__ import annotations

import json
import os
import sys
import threading
import time
from pathlib import Path
from typing import Any

from PySide6.QtCore import Qt, QTimer, QThread
from PySide6.QtGui import QColor, QCursor, QFont, QPainter, QPixmap
from PySide6.QtWidgets import (
    QApplication,
    QDialog,
    QLabel,
    QLineEdit,
    QMainWindow,
    QPushButton,
    QSystemTrayIcon,
    QVBoxLayout,
    QWidget,
)

from . import __version__
from .config import Config, load, save
from .island import IslandController
from .layout import dock_positions, island_rect
from .models import DockItem, SystemStats
from .services import (
    ApplicationResolver,
    LocalAPIServer,
    ProcessRegistry,
    SystemMonitor,
)


class DockWindow(QMainWindow):
    def __init__(self, config: Config, parent=None) -> None:
        super().__init__(parent)
        self.config = config
        self.setWindowFlags(
            Qt.WindowType.FramelessWindowHint
            | Qt.WindowType.StaysOnTopHint
            | Qt.WindowType.Tool
        )
        self.setAttribute(Qt.WidgetAttribute.WA_TranslucentBackground)
        self.setWindowTitle("Desop Dock")
        self.setFixedSize(800, 72)

        self._timer = QTimer(self)
        self._timer.timeout.connect(self._on_timeout)
        self._timer.start(16)

        self._items: list[DockItem] = []
        self._running_names: set[str] = set()
        self._magnified_index: int | None = None
        self._pointer_pos: tuple[float, float] = (0.0, 0.0)

    def set_items(self, items: list[DockItem]) -> None:
        self._items = items

    def set_running_names(self, names: set[str]) -> None:
        self._running_names = names

    def mouseMoveEvent(self, event) -> None:
        pos = event.globalPosition().toPoint()
        self._pointer_pos = (float(pos.x()), float(pos.y()))
        super().mouseMoveEvent(event)

    def _on_timeout(self) -> None:
        self.update()

    def paintEvent(self, event) -> None:
        painter = QPainter(self)
        painter.setRenderHint(QPainter.RenderHint.Antialiasing)

        screen_width = self.width()
        bottom_margin = 48.0
        icon_size = float(self.config.dock.icon_size)
        gap = 8.0

        count = len(self._items) or 5
        positions, bounds = dock_positions(
            count,
            screen_width,
            float(self.height()),
            bottom_margin,
            icon_size,
            gap,
            self._pointer_pos[0] if self.config.dock.magnification else None,
            self.config.dock.magnification,
            self.config.dock.magnification_scale,
        )

        for index, item in enumerate(self._items):
            is_running = item.name.lower() in self._running_names
            rect = positions[index] if index < len(positions) else bounds

            painter.setBrush(QColor(self.config.appearance.tint))
            painter.setPen(Qt.PenStyle.NoPen)
            radius = self.config.appearance.corner_radius
            painter.drawRoundedRect(rect.x, rect.y, rect.width, rect.height, radius, radius)

            if is_running and self.config.dock.show_running:
                dot_rect = QRectF(
                    rect.x + rect.width / 2 - 6,
                    rect.y + rect.height + 4,
                    12,
                    12,
                )
                painter.setBrush(QColor("#34C759"))
                painter.drawEllipse(dot_rect)

        painter.end()

    def sizeHint(self):
        return self.minimumSizeHint()


class TopBarWindow(QMainWindow):
    def __init__(self, config: Config, parent=None) -> None:
        super().__init__(parent)
        self.config = config
        self.setWindowFlags(
            Qt.WindowType.FramelessWindowHint
            | Qt.WindowType.StaysOnTopHint
            | Qt.WindowType.Tool
        )
        self.setAttribute(Qt.WidgetAttribute.WA_TranslucentBackground)
        self.setWindowTitle("Desop TopBar")
        self.setFixedHeight(48)

        self._clock_label = QLabel("12:00")
        self._date_label = QLabel("2024-01-01")
        layout = QVBoxLayout()
        layout.addWidget(self._clock_label)
        layout.addWidget(self._date_label)
        self.setLayout(layout)

        self._timer = QTimer(self)
        self._timer.timeout.connect(self._update_time)
        self._timer.start(1000)
        self._update_time()

    def _update_time(self) -> None:
        now = time.localtime()
        self._clock_label.setText(time.strftime("%H:%M", now))
        self._date_label.setText(time.strftime("%Y-%m-%d %A", now))


class IslandWindow(QMainWindow):
    def __init__(self, config: Config, parent=None) -> None:
        super().__init__(parent)
        self.config = config
        self.controller = IslandController(expanded=config.island.expanded)
        self.setWindowFlags(
            Qt.WindowType.FramelessWindowHint
            | Qt.WindowType.StaysOnTopHint
            | Qt.WindowType.Tool
        )
        self.setAttribute(Qt.WidgetAttribute.WA_TranslucentBackground)
        self.setWindowTitle("Desop Island")

        self._label = QLabel("Desop")
        layout = QVBoxLayout()
        layout.addWidget(self._label)
        self.setLayout(layout)

        if config.island.local_api:
            self._api_server = LocalAPIServer(
                "127.0.0.1",
                config.island.port,
                on_event=self.controller.push,
                on_dismiss=self.controller.dismiss,
                on_expand=self.controller.set_expanded,
            )
            self._api_server.start()
        else:
            self._api_server = None

    def closeEvent(self, event) -> None:
        if self._api_server:
            self._api_server.stop()
        super().closeEvent(event)


class SettingsDialog(QDialog):
    def __init__(self, config: Config, path: Path, parent=None) -> None:
        super().__init__(parent)
        self.config = config
        self.path = path

        self.setWindowTitle("Desop Settings")
        self.setMinimumWidth(500)

        layout = QVBoxLayout()

        layout.addWidget(QLabel("Components"))
        self._dock_check = QPushButton(
            f"Dock: {'ON' if config.dock.enabled else 'OFF'}"
        )
        self._dock_check.clicked.connect(self._toggle_dock)
        layout.addWidget(self._dock_check)

        self._topbar_check = QPushButton(
            f"Top Bar: {'ON' if config.topbar.enabled else 'OFF'}"
        )
        self._topbar_check.clicked.connect(self._toggle_topbar)
        layout.addWidget(self._topbar_check)

        self._widgets_check = QPushButton(
            f"Widgets: {'ON' if config.widgets.enabled else 'OFF'}"
        )
        self._widgets_check.clicked.connect(self._toggle_widgets)
        layout.addWidget(self._widgets_check)

        self._island_check = QPushButton(
            f"Island: {'ON' if config.island.enabled else 'OFF'}"
        )
        self._island_check.clicked.connect(self._toggle_island)
        layout.addWidget(self._island_check)

        self._wallpaper_check = QPushButton(
            f"Wallpaper: {'ON' if config.wallpaper.enabled else 'OFF'}"
        )
        self._wallpaper_check.clicked.connect(self._toggle_wallpaper)
        layout.addWidget(self._wallpaper_check)

        save_button = QPushButton("Save & Close")
        save_button.clicked.connect(self._save_and_close)
        layout.addWidget(save_button)

        self.setLayout(layout)

    def _toggle_dock(self) -> None:
        self.config.dock.enabled = not self.config.dock.enabled
        self._dock_check.setText(f"Dock: {'ON' if self.config.dock.enabled else 'OFF'}")

    def _toggle_topbar(self) -> None:
        self.config.topbar.enabled = not self.config.topbar.enabled
        self._topbar_check.setText(
            f"Top Bar: {'ON' if self.config.topbar.enabled else 'OFF'}"
        )

    def _toggle_widgets(self) -> None:
        self.config.widgets.enabled = not self.config.widgets.enabled
        self._widgets_check.setText(
            f"Widgets: {'ON' if self.config.widgets.enabled else 'OFF'}"
        )

    def _toggle_island(self) -> None:
        self.config.island.enabled = not self.config.island.enabled
        self._island_check.setText(
            f"Island: {'ON' if self.config.island.enabled else 'OFF'}"
        )

    def _toggle_wallpaper(self) -> None:
        self.config.wallpaper.enabled = not self.config.wallpaper.enabled
        self._wallpaper_check.setText(
            f"Wallpaper: {'ON' if self.config.wallpaper.enabled else 'OFF'}"
        )

    def _save_and_close(self) -> None:
        save(self.config, self.path)
        self.accept()


class SystemTrayIcon(QSystemTrayIcon):
    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setIcon(self._make_icon())

    def _make_icon(self) -> QPixmap:
        pixmap = QPixmap(64, 64)
        pixmap.fill(QColor("#101419"))
        painter = QPainter(pixmap)
        painter.setPen(QColor("#FFB45A"))
        painter.drawRoundedRect(12, 12, 40, 40, 12, 12)
        painter.end()
        return pixmap


class DesktopApplication:
    def __init__(
        self, config: Config, config_path: Path, settings_only: bool = False
    ) -> None:
        self.config = config
        self.config_path = config_path
        self.settings_only = settings_only

        self.app = QApplication.instance()
        if self.app is None:
            self.app = QApplication(sys.argv)
        self.app.setApplicationName("Desop")
        self.app.setApplicationVersion(__version__)

        self._system_monitor = SystemMonitor()
        self._process_registry = ProcessRegistry()
        self._resolver = ApplicationResolver()

        self._dock_window: DockWindow | None = None
        self._topbar_window: TopBarWindow | None = None
        self._island_window: IslandWindow | None = None
        self._tray_icon: SystemTrayIcon | None = None

    def _create_windows(self) -> None:
        if self.config.dock.enabled and not self._dock_window:
            self._dock_window = DockWindow(self.config)
        if self.config.topbar.enabled and not self._topbar_window:
            self._topbar_window = TopBarWindow(self.config)
        if self.config.island.enabled and not self._island_window:
            self._island_window = IslandWindow(self.config)

        if not self.settings_only and not self._tray_icon:
            self._tray_icon = SystemTrayIcon()
            menu = self._tray_icon.contextMenu()
            action = menu.addAction("Settings")
            action.triggered.connect(self._show_settings)
            menu.addAction("Exit").triggered.connect(self.app.quit)
            self._tray_icon.show()

    def _show_settings(self) -> None:
        dialog = SettingsDialog(self.config, self.config_path)
        dialog.exec()

    def run(self) -> None:
        if not self.settings_only:
            print("Desop starting...")

        self._create_windows()

        if not self.settings_only:
            print(f"Enabled: {', '.join(self.config.enabled_components())}")
            if self._tray_icon:
                print("Tray icon: right-click for settings or exit.")
            if self._island_window and self._island_window._api_server:
                host, port = self._island_window._api_server.address
                print(f"Island API: http://{host}:{port}")

        if self.settings_only:
            dialog = SettingsDialog(self.config, self.config_path)
            dialog.exec()
        else:
            self._system_monitor.start()

            pid_file = config_module.config_dir() / "desop.pid"
            try:
                pid_file.write_text(str(os.getpid()), encoding="utf-8")
            except OSError:
                pass

            try:
                self.app.exec()
            finally:
                try:
                    pid_file.unlink(missing_ok=True)
                except OSError:
                    pass


if __name__ == "__main__":
    cfg = load()
    app = DesktopApplication(cfg, config_path())
    app.run()
