[简体中文](README.md) | **English**

# Desop

**A desktop beautification suite for Windows: recreate the macOS desktop feel on Windows with a dock, a menu bar, desktop widgets, a dynamic island and more.**

> **Status: Buildable MVP.** A Windows 10/11 user-mode implementation is available (pure Go standard library, zero third-party runtime dependencies). See [`docs/BUILD.md`](docs/BUILD.md), [`docs/INSTALL.md`](docs/INSTALL.md), and the [Windows acceptance checklist](docs/ACCEPTANCE.md).

![Status](https://img.shields.io/badge/status-buildable%20MVP-2ea44f?style=flat-square)
[![GitHub license](https://img.shields.io/github/license/ntars/Desop?style=flat-square)](./LICENSE)
[![GitHub Star](https://img.shields.io/github/stars/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/stargazers)
[![GitHub Fork](https://img.shields.io/github/forks/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/network/members)
![GitHub Repo size](https://img.shields.io/github/repo-size/ntars/Desop?style=flat-square&color=3cb371)

[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d4.svg?style=flat-square&logo=windows&logoColor=white)](#)
[![Runtime](https://img.shields.io/badge/runtime-user%20mode-2ea44f.svg?style=flat-square)](#)
[![System Changes](https://img.shields.io/badge/system%20changes-none-8957e5.svg?style=flat-square)](#)

---

## 1. Name and Positioning

> **仓库描述 / Description**：`Desop —— Windows 端桌面美化套装：把 macOS 桌面观感复刻到 Windows，提供 dock 栏、顶部栏、桌面小组件与灵动岛 | A Windows desktop beautification suite: macOS-style dock, menu bar, widgets and dynamic island. Buildable MVP.`
>
> **话题标签 / Topics**（20 个，仅英文）：`beautification`, `desktop-customization`, `desktop-widgets`, `dock`, `dynamic-island`, `macos-style`, `menubar`, `mvp`, `theme`, `ui`, `user-mode`, `windows`, `windows-10`, `windows-11`, `desktop-environment`, `taskbar`, `wallpaper`, `theming`, `windows-desktop`, `concept`


**Name**: Desop. "Des" from Desktop, "op" for operation layer — the desktop operation layer.

**One-line positioning**: rebuild a macOS-style desktop interaction layer on Windows, so that switching systems no longer means switching habits.

**Three boundaries**:

1. **It is a desktop layer, not a system layer.** Everything runs in user mode: no system files touched, no shell replaced, no system updates disabled. Uninstalling restores the original state.
2. **It pursues look and feel, not pixel-perfect imitation.** Rounded corners, frosted glass and easing curves can be learned from macOS, but Windows window management and taskbar behaviour are system contracts and are not forced.
3. **It is a set of separable components, not one bundle.** Dock, menu bar, widgets and dynamic island are independent; a user may enable only one or two.

**Relationship to macOS**: this project references macOS desktop interaction design. It reuses no macOS code, assets or fonts; all visual elements are independently implemented.

---

## 2. Overview

### 2.1 Background

People moving from macOS to Windows are usually bothered less by missing features than by **a break in habit**: no global menu bar, no dock with magnification and bounce, no widgets at hand, and notifications confined to a bubble in the bottom-right corner.

Existing options sit at two extremes: heavyweight beautification packs that edit the registry, replace the shell and modify system files — hard to recover if something breaks; or single-purpose utilities that handle only the dock or only widgets, each with its own style, which together look messier than before.

### 2.2 Problems addressed

1. **Visual fragmentation.** When the four elements come from different tools, corner radii, shadows, animation durations and fonts all disagree — it looks like four programs stacked together.
2. **Uncontrolled risk.** Solutions that modify system files are expensive to recover from, so ordinary users will not try them.
3. **Performance cost.** Poorly implemented frosted glass and live blur drop the whole machine's frame rate, turning beautification into a burden.

### 2.3 Core approach

One set of design tokens shared by all four components:

| Token | Purpose |
|---|---|
| Radius and spacing | One set of values for every component, not one copy each |
| Shadow and material | Unified blur radius, opacity and edge light for frosted glass |
| Easing curves | Unified durations and easing, so nothing "bounces here and slides there" |
| Font and size | Follow the system but unify the scale |

The four components run as independent processes or modules, share the tokens, and do not depend on one another — disabling one leaves the others untouched.

### 2.4 Use cases

- Users who moved from macOS to Windows and want to keep their habits;
- Users who like a tidy desktop with a dock and widgets but do not want a heavyweight pack;
- Users who need long-term stability and cannot accept "beautification broke the system".

### 2.5 Non-goals

- No system files modified, no `explorer.exe` replaced, no protected settings changed;
- No theme store, no wallpaper distribution;
- No deep data collection for performance or hardware widgets (read-only public interfaces only);
- No promise of behaviour identical to macOS — where system contracts differ, Windows wins.

---

## 3. Core Features

**Status: implemented MVP features are marked ✅; items pending Windows acceptance or limited by Windows APIs are marked ⚠.**

### 3.1 Dock

| Capability | Description | Status |
|---|---|---|
| Icon magnification | Neighbouring icons scale by distance as the pointer approaches | ✅ Implemented with bounded smoothstep |
| Running indicator | A dot under running applications | ✅ Implemented via read-only process snapshot |
| Bounce feedback | Bounce while launching, switchable off | ✅ Damped launch animation implemented |
| Auto-hide | Hide at the edge with an adjustable trigger zone | ✅ Basic logic wired; Windows acceptance pending |
| Position | Bottom / left / right | ✅ Implemented |
| Multiple monitors | Independent dock per monitor or main screen only | ⚠ Currently primary screen only |

### 3.2 Menu bar

| Capability | Description | Status |
|---|---|---|
| Global menu | Show the foreground window title as a menu entry | ✅ Implemented; limited by exposed window titles, disabled by default |
| Status area | Time, date, battery, network, volume | ✅ Time/date/network implemented; battery/volume degrade honestly |
| Tray consolidation | Collect system tray icons into the bar | ⚠ Disabled by default, not implemented |
| Auto-hide | Hide during fullscreen applications | ✅ Conservative detection wired; acceptance pending |

### 3.3 Desktop widgets

| Capability | Description | Status |
|---|---|---|
| Clock and calendar | Large clock, month view, schedule summary | ✅ Clock and month view implemented; schedule not implemented |
| System status | CPU / memory / disk / network throughput | ✅ Implemented via read-only public Win32 APIs |
| Notes | Local plain-text notes, no network | ✅ Implemented and persisted |
| Weather | User-supplied data source only; offline by default | ✅ Offline fallback; no network calls |
| Layout | Free dragging, snap alignment, toggleable grid snapping | ✅ Drag and grid snap implemented; acceptance pending |

### 3.4 Dynamic island

| Capability | Description | Status |
|---|---|---|
| Resting form | A capsule centred at the top of the screen, tiny when collapsed | ✅ Implemented |
| Expand animation | Expands on click or event | ✅ Instant expansion implemented; spring curve pending |
| Event sources | Music, download progress, timers, notification summaries | ✅ Local API supports media/download/timer/custom; system media session pending |
| Third-party access | Local interface allowing other programs to push status | ✅ Loopback only (127.0.0.1) |

### 3.5 Explicitly out of scope

- No wallpaper engine or animated wallpapers;
- No forced window-effect injection;
- No system accent-colour replacement;
- No boot animation changes.

---

## 4. Core Characteristics

**Status: implementation is in place; evidence is source code and automated tests. Performance and GUI evidence still require Windows acceptance.**

### Highlight 1: Zero system changes, uninstall restores everything

- **Implementation**: no system directories, no shell replacement, no explorer injection. Autostart writes only one HKCU Run value, removable by the uninstall script.
- **Evidence**: `internal/win32/app_windows.go`, `scripts/uninstall.ps1`, `docs/ACCEPTANCE.md`.

### Highlight 2: One token set, four consistent components

- **Implementation**: radii, spacing, blur, opacity, motion and fonts are defined centrally; components only read.
- **Evidence**: `internal/tokens/tokens.go` and `internal/tokens/tokens_test.go`.

### Highlight 3: Performance budget as a release gate

- **Implementation**: components are togglable, timers are throttled, and Dock repaints only on pointer changes.
- **Evidence**: logic tests and cross-build via `scripts/check.sh`; CPU/memory/frame rate must be measured on Windows per `docs/ACCEPTANCE.md`.

### Highlight 4: Selectable components, no bundling

- **Implementation**: four independent component switches in config, CLI, tray and settings window.
- **Evidence**: `internal/config/config_test.go`, `internal/app/settings_test.go`, and the Windows acceptance checklist.

### Highlight 5: Reuse system capabilities first

- **Implementation**: request DWM system acrylic first; degrade to translucent GDI drawing when unsupported.
- **Evidence**: `DwmSetSystemBackdrop` and `SetLayeredWindowAttributes` in `internal/win32/win32_windows.go`.

---

## 5. Design Details and Engineering Trade-offs

**This chapter records design commitments to be honoured during implementation.**

### 5.1 Event sources for the dynamic island must be honest

The dynamic island is not a system capability, so every event source must be stated:

| Event | Source | Limitation |
|---|---|---|
| Music playback | System media session (SMTC) | Some players do not report |
| Download progress | Pushed by the application (local interface) | Requires application cooperation |
| Notification summary | System notification listening | Requires user authorisation |
| Timer | Internal to this program | This program only |

**What is not done**: pretending to see every application's state. If it cannot be read, it is not shown — rather than invented.

### 5.2 Yielding to fullscreen applications

- Menu bar and dock hide automatically while a fullscreen application runs;
- Detection prefers the system-provided fullscreen state over guessing from window size;
- When detection fails, prefer staying visible — accidental hiding is more annoying than extra display.

### 5.3 DPI and multi-monitor boundaries

- Each monitor computes its own scaling factor; coordinates are never shared;
- Dragging across screens recomputes against the target screen's DPI with no interpolation — interpolation makes icons deform at the crossing moment;
- Mixed DPI (4K primary plus 1080p secondary) is a mandatory test scenario.

### 5.4 Frosted glass fallback chain

```
System acrylic (Win11) → System blur (Win10) → Translucent solid → Opaque
```

Each level must degrade automatically and remain visually valid afterwards. A black rectangle is not an acceptable fallback.

### 5.5 Performance boundaries of icon magnification

- Only icons in the visible region are computed; off-screen ones are skipped;
- Magnification range is capped (three icons either side of the pointer by default);
- Uses compositor transforms, never triggering layout.

### 5.6 Completeness of uninstallation

Uninstalling must restore:

1. Startup entries;
2. Registry run keys;
3. Desktop shortcuts;
4. The user data directory (asking whether to keep the configuration).

**This is an acceptance criterion, not an optional nicety.**

---

## 6. Technology Stack and Dependencies

**Selected stack: Go 1.22+ standard library + public Win32 APIs, zero third-party runtime dependencies.**

| Constraint | Content |
|---|---|
| Runtime | Windows 10 / 11, user mode, no administrator rights |
| System changes | None. No system directories, no shell replacement |
| Rendering | Native windows + GDI + DWM system backdrop |
| System interfaces | Public Win32 APIs; read-only status |
| Local interface | Loopback HTTP on 127.0.0.1, configurable port |
| Licence | Apache-2.0 (core & API) / MIT (shells) |

Dependencies: zero third-party runtime dependencies; Go toolchain (BSD-3-Clause) is build-only. See [NOTICE](NOTICE).

---

## 7. Installation and Deployment

A Windows portable build directory and PowerShell install/uninstall scripts are provided.

| Platform | Form | Status |
|---|---|---|
| Windows 11 | Portable `desop.exe` directory | ✅ Buildable; Windows acceptance pending |
| Windows 10 | Same | ✅ Buildable; Windows acceptance pending |

Build: `./scripts/build.sh`. See [`docs/INSTALL.md`](docs/INSTALL.md) and [`docs/ACCEPTANCE.md`](docs/ACCEPTANCE.md).

---

## 8. Quick Start

```text
./scripts/build.sh
cd release/desop-<version>
desop.exe
desop.exe status
desop.exe settings
desop.exe enable dock,widgets
desop.exe disable island
desop.exe api-event
desop.exe selftest
```

Right-click the tray icon to toggle components, switch autostart, restart or exit. Full steps: [`docs/INSTALL.md`](docs/INSTALL.md).

---

## 9. Configuration

Config file: `%APPDATA%\Desop\config.json`; created on first launch. Component toggles are available in the settings window and CLI.

| Category | Content | Status |
|---|---|---|
| Component toggles | Enable/disable each of the four components | ✅ |
| Appearance | Shared tokens; theme/blur level UI | ⚠ Tokens centralised, UI levels pending |
| Behaviour | Auto-hide, trigger width, autostart | ✅ Partially implemented |
| Advanced | Debug logs, performance diagnostics | ⚠ Not implemented |

---

## 10. API and Module Reference

| Module | Responsibility | Status |
|---|---|---|
| `internal/tokens` | Single definition of radii, shadows and motion | ✅ |
| `internal/dock` | Icon layout, magnification, running indicators | ✅ |
| `internal/topbar` | Status aggregation, layout, auto-hide policy | ✅ Partial |
| `internal/widgets` | Clock, system status, notes, weather fallback | ✅ |
| `internal/island` | Capsule and local event API | ✅ |
| `internal/win32` | Native windows, tray, read-only system interfaces | ✅ |
| `cmd/desop` | Entry point and CLI | ✅ |

Local API: `POST http://127.0.0.1:17890/event`, `POST /dismiss`, `GET /health`. Example: `desop.exe api-event`.

---

## 11. Repository Layout

```text
Desop/
├── cmd/desop/          # entry point and CLI
├── internal/           # tokens, config, four component cores, Win32 layer
├── docs/               # BUILD / INSTALL / ACCEPTANCE
├── scripts/            # build/check/install/uninstall
├── release/            # build output (gitignore)
├── README.md / README_EN.md
├── LICENSE / LICENSE-MIT / NOTICE / LICENSING.md
└── CHANGELOG.md
```

---

## 12. Performance

**Automated results and Windows measurements are recorded separately.** Formatting, tests and cross-build are covered by `./scripts/check.sh`; the following targets still require Windows acceptance per `docs/ACCEPTANCE.md`:

| Metric | Target | Status |
|---|---|---|
| Idle CPU | < 1% (all four components enabled) | Windows acceptance pending |
| Idle memory | < 150 MB | Windows acceptance pending |
| Dock magnification frame rate | ≥ 55 fps (1080p, 20 icons) | Windows acceptance pending |
| Cold start | ≤ 2 s | Windows acceptance pending |

A stable release is blocked until these targets are met.

---

## 13. Security and Privacy

| Dimension | Current implementation |
|---|---|
| System changes | None. User mode, no writes to system directories |
| Privileges | No administrator rights required |
| Network | Loopback only by default; weather never fetches by itself |
| Data collection | No telemetry, no analytics, no crash reporting |
| Local storage | Configuration and note text only |
| Notification listening | Not implemented; no permission requested |

**Explicitly not done**: no uploads, no reading user file contents, no silent background networking.

---

## 14. Compatibility

| Item | Status |
|---|---|
| Windows 11 | Buildable; Windows acceptance pending |
| Windows 10 (1809+) | Buildable; some blur effects degrade |
| Multiple monitors | Primary-screen MVP; mixed DPI acceptance pending |
| High DPI (125% / 150% / 200%) | Work-area coordinates are used; real-device verification needed |
| Windows 7 / 8 | Unsupported |

---

## 15. Known Issues and Limitations

### 15.1 Current facts

1. **Buildable code exists**: Windows user-mode implementation using the Go standard library.
2. **Build and deployment scripts exist**: `scripts/build.sh`, `scripts/check.sh`, PowerShell install/uninstall.
3. **Windows acceptance is pending**: this repository was validated on macOS with logical tests and cross-compilation; GUI/DPI/performance must be proven on Windows.

### 15.2 Identified design limitations

| Limitation | Reason | Solvable |
|---|---|---|
| Limited global menu coverage | Windows applications need not expose menu structure | Not solvable; best effort only |
| The dynamic island is simulated in software | Windows has no equivalent system mechanism | Not solvable; event sources are limited |
| Some blur effects depend on the OS version | Windows 10 and 11 differ in capability | Mitigated by the fallback chain |
| May be flagged by security software | Involves system-level hooks | Must be verified early; drop the feature if necessary |

### 15.3 Questions now answered by the MVP or pending Windows acceptance


| Question | Why it matters |
|---|---|
| Blur availability across Windows 10 / 11 | Determines whether a custom-drawn fallback is needed |
| Global menu coverage across common applications | Determines whether the feature is worth building |
| Multi-monitor and DPI scaling edge cases | Determines layout computation complexity |
| Whether anti-cheat or security software flags it | Any system-level hook must be cleared beforehand |
| Whether uninstall fully restores the system | Determines whether users dare install it |

---

## 16. Changelog

See [CHANGELOG.md](CHANGELOG.md) for the first buildable version. Releases: <https://github.com/ntars/Desop/releases>.

---

## 17. Roadmap

The current version is an MVP; next steps follow Windows acceptance:

1. Fix blocking issues found by the acceptance checklist;
2. Refine Dock motion curves and spring expansion;
3. Decide on global menu and tray consolidation based on coverage;
4. Add more widgets and multi-monitor policies.

---

## 18. Contributing

Scenario, compatibility and feel feedback are welcome. Code contributors should read `docs/ACCEPTANCE.md` and keep changes verifiable.

Contribution rules:

1. Token changes must update the token layer, never hard-code in components;
2. Register new dependencies in NOTICE before merging;
3. Do not introduce unknown or incompatible code licences.

---

## 19. Copyright and License

Copyright (c) 2024–2026 Ntarstudio

This project is licensed per directory, using **only Apache License 2.0 and MIT** — both permissive. Full texts live in [LICENSE](LICENSE) (Apache-2.0) and [LICENSE-MIT](LICENSE-MIT) (MIT); the layer map is in [LICENSING.md](LICENSING.md).

| Path | License |
|---|---|
| Desktop-layer core, local API and protocol | Apache-2.0 |
| Component shells, installer, build scripts | MIT |
| Brand name, logo, interface design | All rights reserved |

Both licenses allow commercial use, modification, closed-source distribution and sublicensing — **no copyleft**. Your derivative work does not have to be open-sourced. Just keep the copyright notice, the license texts and [NOTICE](NOTICE) when redistributing. Apache-2.0 adds an express patent license (Section 3).

**Why not AGPL**: this project exposes a local HTTP interface for third-party programs to push dynamic-island status, but that is not a "network service" — the endpoint runs on the user's own machine and is never exposed to the public internet. AGPL would only scare off plugin authors and enterprise users, which is exactly the ecosystem this project needs.

**Names and trademarks.** The names and marks 「Ntarstudio」, 「星玖工作室」 and 「Desop」 belong to Ntarstudio. The code licenses grant copyright permissions only, not trademark rights.

**Third-party names.** macOS, Windows, Dock, Dynamic Island, explorer.exe and other names appearing in this document belong to their respective owners and are used only to describe design references and compatibility targets. No affiliation or endorsement is implied.

**No warranty.** This project is at the planning stage and makes no availability commitment. The program (when it exists) is provided "as is", without any express or implied warranty.

---

## 20. NOTICE and Open-Source Compliance

**There are no third-party dependencies at present; NOTICE exists as a placeholder.** The full statement is in [NOTICE](./NOTICE), covering: the current zero-dependency fact, reference declarations, registration requirements for third-party components (seven fields), four rules, **extra requirements for icons and fonts** (beautification projects most often fail on assets), and a licence compatibility table.

**Referenced but not used as code**:

| Object | Rights holder | Relationship |
|---|---|---|
| macOS desktop interaction design | Apple Inc. | **Design reference only; no code, assets, fonts or icons used** |
| Windows system APIs | Microsoft Corporation | Called through public interfaces only; no system components modified |

**Registration requirements when third-party components are introduced** (for the future):

Any new library, framework, font, icon set or algorithm source must first be registered here: name, version, copyright holder, license type, official address, purpose in this project, and whether it was modified with a description of the changes. **Where a licence cannot be confirmed, mark it "unconfirmed" and hold the merge rather than assuming or inventing.**

Indirect and transitive dependencies must be registered as well; "not a direct reference" is not an acceptable reason to omit them.

---

Ntarstudio · <https://www.ntars.cn/>