# Desop Change Log

## Unreleased — 首个可构建版本 (Python + PySide6)

### Added

- 纯 Python + PySide6 Windows 用户态实现：Dock、顶部栏、桌面小组件、灵动岛、托盘与设置窗口。
- 统一设计令牌层、配置持久化、CLI 组件开关与只读系统状态采样。
- 灵动岛 loopback 本地事件接口，支持第三方推送与离线默认配置。
- `scripts/build_desop.py`、`scripts/check.sh`、`scripts/install.sh`、`scripts/uninstall.sh`。
- `docs/BUILD.md`、`docs/INSTALL.md`、`docs/ACCEPTANCE.md` 与 Windows 实机验收清单。
- `pyproject.toml`、`setup.py` 用于 Python 打包。

### Fixed during implementation review

- Optional island API port conflicts no longer block the desktop shell from starting.
- Corrected tray callback decoding (low word of `lParam`), isolated tray menu command handling from control IDs, and made `stop` re-armable.
- Fixed settings dialog message scoping, stale stop events after shutdown, and a Win32 monitor-callback pointer safety violation.
- Dock icon sizing now follows the host monitor DPI instead of assuming 96 DPI everywhere.
- Uninstall scripts support non-interactive config removal for automated acceptance.
- Tray icon now carries a stock application icon, dock icons draw with explicit GDI brushes, and the island API exposes a collapse endpoint.
- Windows acceptance helper can generate a first-run config on a clean machine with `-GenerateConfig`.
- Diagnostic commands now create the default config on a clean machine; top bar placement, widget indexing, and config restoration after acceptance checks were corrected.
- Trigger first paint for all enabled components at startup instead of waiting for the first throttled refresh.
- Lock the Win32 shell and self-test to their creating OS thread so thread-scoped message queues and timers remain reliable.
- Stop repainting unchanged desktop components: dirty windows are invalidated only when content or geometry changes, visibility toggles only on state change, and dock repaint follows pointer movement over the dock.
- Make portable install/uninstall create the HKCU Run key when absent and register the installed executable for autostart instead of the source bundle.
- Added deterministic idle repaint-frame counting to the self-test and acceptance helper so window flicker fails verification even when CPU sampling is noisy.

### Added (earlier Go implementation)

- `desop.exe selftest` measures a 3-second idle interval, working set and idle paint frames on Windows, then reports CPU/memory/flicker PASS/FAIL.
- First-run config creation, a real `desop.exe stop` signal, loopback API port fallback, Dock launch bounce, calendar widget and foreground-title global menu.
- `scripts/acceptance.ps1` to generate Windows CLI/config/loopback-API PASS/FAIL evidence.

### Fixed during implementation review (earlier Go implementation)

- Optional island API port conflicts no longer block the desktop shell from starting.
- Corrected MIB_IFROW field offsets and CPU utilisation math using documented Win32 semantics.
- Fixed tray icon `NOTIFYICONDATAW` size/version and thread-scoped message pumping so the settings dialog no longer steals host messages.
- Fixed component window messages being filtered out and removed a recursive window-procedure owner lookup that could crash the GUI.
- Added unit tests for network counter parsing, CPU percentages and rate calculations.
- Registered all component window classes before creating windows; use a single host window procedure with owner lookup.
- Corrected Dock window geometry so magnified icons remain inside the window and clicks map to the right item.
- Fixed settings-window close/cancel and tray-command handling; exit now cleans up timers, tray icon and windows.
- Added conservative fullscreen detection for top bar and Dock auto-hide.
- Verified loopback island API integration with `DESOP_NETWORK_TESTS=1`.

### Known limitations

- 全局菜单、托盘收纳与电量数值依赖额外 Windows API，本版本按设计降级并不编造状态。
- DPI、多显示器、开机自启持久性与 2 分钟性能预算仍需按 `docs/ACCEPTANCE.md` 在目标 Windows 实机确认。

### Added (Python port)

- 纯 Python 实现：`desop/` 目录，包含 config、ui、services、island、layout、cli 等模块。
- PySide6 渲染：原生窗口 + 圆角/阴影/动效，仿 macOS 视觉效果。
- PyInstaller 打包：生成 `desop.app`（macOS）或 `desop.exe`（Windows），零依赖分发。
- CLI：`desop start|stop|status|enable|disable|settings|doctor|selftest|package`。
- 配置持久化：JSON 存储于 `%APPDATA%\Desop\config.json` 或 `~/Library/Application Support/Desop/config.json`。

### Fixed during Python implementation review

- Configuration load/save with atomic writes and backup on corruption.
- Island event validation (title, progress bounds, kind whitelist).
- System monitor refresh loop with network rate calculation.
- Local API server with /health, /event, /dismiss and /event/collapse endpoints.

## Previous versions — Go implementation (see git history)
