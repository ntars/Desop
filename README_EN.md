[简体中文](README.md) | **English**

# Desop

**A desktop beautification suite for Windows: recreate the macOS desktop feel on Windows with a dock, a menu bar, desktop widgets, a dynamic island and more.**

> **Status: Planned.** This repository currently holds project material only — no runnable code, no build output, no releases. Everything below describes **positioning and intent**, not implemented features. Chapters that are still planned are explicitly marked "Planned".

![Status](https://img.shields.io/badge/status-planned-fb8500?style=flat-square)
[![GitHub license](https://img.shields.io/github/license/ntars/Desop?style=flat-square)](./LICENSE)
[![GitHub Star](https://img.shields.io/github/stars/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/stargazers)
[![GitHub Fork](https://img.shields.io/github/forks/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/network/members)
![GitHub Repo size](https://img.shields.io/github/repo-size/ntars/Desop?style=flat-square&color=3cb371)

[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d4.svg?style=flat-square&logo=windows&logoColor=white)](#)
[![Runtime](https://img.shields.io/badge/runtime-user%20mode-2ea44f.svg?style=flat-square)](#)
[![System Changes](https://img.shields.io/badge/system%20changes-none-8957e5.svg?style=flat-square)](#)

---

## 1. Name and Positioning

> **仓库描述 / Description**：`Desop —— Windows 端桌面美化套装：把 macOS 桌面观感复刻到 Windows，提供 dock 栏、顶部栏、桌面小组件与灵动岛（规划中） | A Windows desktop beautification suite: macOS-style dock, menu bar, widgets and dynamic island. Planned, no code yet.`
>
> **话题标签 / Topics**（20 个，仅英文）：`beautification`, `desktop-customization`, `desktop-widgets`, `dock`, `dynamic-island`, `macos-style`, `menubar`, `planned`, `theme`, `ui`, `user-mode`, `windows`, `windows-10`, `windows-11`, `desktop-environment`, `taskbar`, `wallpaper`, `theming`, `windows-desktop`, `concept`


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

**Everything in this chapter is planned; none of the four components is implemented.**

### 3.1 Dock

| Capability | Description | Status |
|---|---|---|
| Icon magnification | Neighbouring icons scale by distance as the pointer approaches, imitating macOS's magnetic effect | Planned |
| Running indicator | A dot under running applications | Planned |
| Bounce feedback | Bounce while launching, switchable off | Planned |
| Auto-hide | Hide at the edge, trigger-zone width adjustable | Planned |
| Position | Bottom / left / right | Planned |
| Multiple monitors | Independent dock per monitor or main screen only | Planned |

### 3.2 Menu bar

| Capability | Description | Status |
|---|---|---|
| Global menu | Show the menu items of the foreground window (limited by how far Windows applications cooperate) | Planned |
| Status area | Time, date, battery, network, volume | Planned |
| Tray consolidation | Collect system tray icons into the bar, reducing clutter in the bottom-right corner | Planned |
| Auto-hide | Hide during fullscreen applications | Planned |

**Known limitation**: Windows applications are not obliged to expose their menu structure, so the global menu only works for applications supporting the standard menu API; others show empty. This is stated at the planning stage — no promise that "it works for every app".

### 3.3 Desktop widgets

| Capability | Description | Status |
|---|---|---|
| Clock and calendar | Large clock, month view, schedule summary | Planned |
| System status | CPU / memory / disk / network throughput (read-only public interfaces) | Planned |
| Notes | Local plain-text notes, no network | Planned |
| Weather | The user supplies a data source; offline by default | Planned |
| Layout | Free dragging, snap alignment, toggleable grid snapping | Planned |

### 3.4 Dynamic island

| Capability | Description | Status |
|---|---|---|
| Resting form | A capsule centred at the top of the screen, tiny when collapsed | Planned |
| Expand animation | Expands on click or event, with a spring curve | Planned |
| Event sources | Music playback, download progress, notification summaries, timers | Planned |
| Third-party access | A local interface allowing other programs to push status | Planned |

**Note**: on macOS the dynamic island is a product of hardware and system together; Windows has no equivalent mechanism, so this project's implementation is **pure software simulation** — driven by reading public information such as system media sessions and download progress. This must be stated plainly rather than letting anyone assume a system-level capability.

### 3.5 Explicitly out of scope

- No wallpaper engine or animated wallpapers;
- No forced injection of window effects (rounded corners, transparency);
- No replacement of system accent colours;
- No modification of boot animations.

---

## 4. Core Characteristics

**This chapter describes design intent. Nothing is implemented, so no reproducible evidence is offered.** Each item will gain its "implementation" and "evidence" once a first runnable build exists.

### Highlight 1: Zero system changes, uninstall restores everything

- **Name**: Pure user-mode implementation.
- **Value**: Solves the fear that "a beautification pack breaks the system and it will not boot" — the primary reason ordinary users avoid such tools.
- **Implementation (planned)**: No writes to system directories, no shell replacement, no injection into explorer. All elements are windows of independent processes interacting with the system through public APIs.
- **Differentiation**: Most heavyweight packs demand disabling system updates and replacing system files, with extremely high recovery costs on failure.
- **Evidence**: To be provided after implementation. What can be checked today is that "zero system changes" is already written into the security principles in Chapter 13 and the compatibility boundaries in Chapter 14.

### Highlight 2: One token set, four consistent components

- **Name**: Unified token layer.
- **Value**: Solves "four tools stacked together with four different radii, shadows and easings".
- **Implementation (planned)**: Radius, spacing, shadow, blur radius, animation duration and easing are defined centrally; components only read.
- **Differentiation**: Most alternatives are built by different authors and are inconsistent by nature.
- **Evidence**: To be provided after implementation.

### Highlight 3: Performance budget as a release gate

- **Name**: Three hard targets.
- **Value**: Solves "installing a beautification tool made everything slow" — the second most common reason such tools get uninstalled.
- **Implementation (planned)**: Idle CPU < 1%, idle memory < 150 MB, dock magnification ≥ 55 fps. If unmet, it does not ship.
- **Differentiation**: Most tools treat performance as "later optimisation"; this project treats it as an entry requirement.
- **Evidence**: To be provided after implementation.

### Highlight 4: Selectable components, no bundling

- **Name**: Enable what you need.
- **Value**: Solves "I only wanted the dock and got four things forced on me".
- **Implementation (planned)**: Four independent components, installable and disableable individually; a disabled component leaves no process behind.
- **Differentiation**: All-in-one packs usually bind every component together.
- **Evidence**: To be provided after implementation.

### Highlight 5: Reuse system capabilities first, do not reinvent

- **Name**: System blur before custom blur.
- **Value**: Solves "custom-drawn frosted glass causing battery drain and dropped frames".
- **Implementation (planned)**: Prefer DWM / acrylic and other system blur; custom drawing only as a fallback where unsupported.
- **Differentiation**: Custom blur is expensive on low-end machines and competes with the system compositor.
- **Evidence**: To be provided after implementation.

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

**Not decided.** The planning stage avoids naming a framework rather than writing a selection table that will be overturned later.

### 6.1 Fixed constraints

| Constraint | Content |
|---|---|
| Runtime | Windows 10 / 11, user mode, runs without administrator privileges |
| System changes | None. No writes to system directories, no shell replacement |
| Dependency discipline | As few runtime dependencies as possible; prefer system capabilities |
| License | Apache-2.0 (core & API) / MIT (shells) |

### 6.2 Directional thinking (not a settled choice)

| Layer | Direction | Status |
|---|---|---|
| UI rendering | Native windows plus system composition; no self-built rendering engine | To be assessed |
| Blur | Prefer DWM / acrylic | To be verified |
| System interfaces | Public Win32 APIs and Windows App SDK | To be assessed |
| Component communication | In-process modules or local IPC | To be assessed |

### 6.3 Dependency list

**Zero at present.** The repository contains no code and therefore no direct or indirect dependencies. Any future dependency must be registered here with name, version, copyright holder, license, address, purpose and modification status.

---

## 7. Installation and Deployment

**Nothing is installable yet.** The repository holds documentation only — no code, no build scripts, no releases.

Planned distribution (unimplemented):

| Platform | Form | Status |
|---|---|---|
| Windows 11 | `.exe` installer / portable build | Planned |
| Windows 10 | Same as above | Planned |

**The only viable route today**: `git clone https://github.com/ntars/Desop.git`, then read the documentation.

---

## 8. Quick Start

**There is nothing to start yet** — nothing runs.

To understand the project, read in this order:

1. Chapter 2, Overview — what it wants to solve;
2. Chapter 3, Core Features — what each of the four components does;
3. Section 15.3, Questions to validate — what is still unanswered.

---

## 9. Configuration

**No configuration exists.** The project is unimplemented; there is no config file, environment variable or command-line argument.

Planned configuration (unimplemented):

| Category | Content | Status |
|---|---|---|
| Component switches | Enable/disable each of the four components | Planned |
| Appearance | Accent colour, blur strength, radius preset | Planned |
| Behaviour | Auto-hide, trigger-zone width, launch at startup | Planned |
| Advanced | Debug logging, performance diagnostics | Planned |

---

## 10. API and Module Reference

**No API, no modules.** The repository contains only `README.md`, `README_EN.md` and `LICENSE`.

Planned module split (unimplemented):

| Module | Responsibility | Status |
|---|---|---|
| Token layer | The single definition of radius, shadow and easing | Planned |
| Dock | Icon magnification, running indicators, auto-hide | Planned |
| Menu bar | Global menu, status area, tray consolidation | Planned |
| Widgets | Clock, system status, notes, weather | Planned |
| Dynamic island | Capsule form, expand animation, event intake | Planned |
| Local interface | Lets third parties push dynamic-island status | Planned |

---

## 11. Repository Layout

**Actual structure today**:

```
Desop/
├── README.md          # Chinese documentation
├── README_EN.md       # English documentation (this file)
├── LICENSE            # Apache-2.0 full text (GitHub detects the license from this)
├── LICENSE-MIT        # MIT full text
└── LICENSING.md       # Layer routing table (human-readable)
```

Planned structure (not created):

```
Desop/
├── tokens/            # Design token layer
├── components/
│   ├── dock/          # Dock
│   ├── topbar/        # Menu bar
│   ├── widgets/       # Desktop widgets
│   └── island/        # Dynamic island
├── system/            # System interface wrappers (windows, tray, media session, performance counters)
├── docs/              # Design and validation documents
├── tests/             # Assertions
└── tools/             # Build and check scripts
```

---

## 12. Performance

**No measured data.** The project is unimplemented; any performance figure would be a guess, and this project does not write guesses.

Planned targets, serving as release gates rather than current results:

| Metric | Target | Status |
|---|---|---|
| Idle CPU | < 1% (all four components enabled) | Unverified |
| Idle memory | < 150 MB | Unverified |
| Dock magnification frame rate | ≥ 55 fps (1080p, 20 icons) | Unverified |
| Cold start | ≤ 2 s | Unverified |

If the targets are not met, it does not ship — rather than "ship first, optimise later".

---

## 13. Security and Privacy

**This chapter records design principles to be honoured during implementation.**

| Dimension | Planned approach | Status |
|---|---|---|
| System changes | None. Runs in user mode, no writes to system directories | Planned |
| Privileges | No administrator rights required | Planned |
| Network | Zero network requests by default (weather only when the user configures it) | Planned |
| Data collection | No telemetry, no analytics, no crash reporting | Planned |
| Local storage | Configuration only, never user content | Planned |
| Notification listening | Requires explicit authorisation, revocable at any time | Planned |

**Explicitly not done**: no data uploads, no reading of user file contents, no silent background networking.

---

## 14. Compatibility

**There is nothing to be compatible with yet** — no runnable artefact.

Planned range (unverified):

| Item | Plan | Status |
|---|---|---|
| Windows 11 | Full support | Planned |
| Windows 10 (1809+) | Supported, some blur effects degrade | Planned |
| Multiple monitors | Supported, including mixed DPI | Planned |
| High DPI (125% / 150% / 200%) | Supported | Planned |
| Windows 7 / 8 | **Unsupported** | Explicitly excluded |

---

## 15. Known Issues and Limitations

### 15.1 Facts about the current stage

1. **There is no code.** The repository holds documentation and a licence only.
2. **There is no schedule.** The planning stage makes no release-date commitment.
3. **The technical route is undecided.** The thinking in Chapter 6 may be overturned by validation results.

### 15.2 Identified design limitations

| Limitation | Reason | Solvable |
|---|---|---|
| Limited global menu coverage | Windows applications need not expose menu structure | Not solvable; best effort only |
| The dynamic island is simulated in software | Windows has no equivalent system mechanism | Not solvable; event sources are limited |
| Some blur effects depend on the OS version | Windows 10 and 11 differ in capability | Mitigated by the fallback chain |
| May be flagged by security software | Involves system-level hooks | Must be verified early; drop the feature if necessary |

### 15.3 Questions that must be answered before development

No answer, no code:

| Question | Why it matters |
|---|---|
| Blur availability across Windows 10 / 11 | Determines whether a custom-drawn fallback is needed |
| Global menu coverage across common applications | Determines whether the feature is worth building |
| Multi-monitor and DPI scaling edge cases | Determines layout computation complexity |
| Whether anti-cheat or security software flags it | Any system-level hook must be cleared beforehand |
| Whether uninstall fully restores the system | Determines whether users dare install it |

---

## 16. Changelog

**No versions yet.** The project is at the planning stage and has released nothing.

The changelog will live on the **GitHub Releases** page: <https://github.com/ntars/Desop/releases>

The README describes the project's current state; releases record what changed each time. The two are never mixed.

---

## 17. Roadmap

Ordered by dependency; each item is a prerequisite for the next. **None has started.**

1. **Technical validation**: answer the five questions in 15.3, especially "does uninstall fully restore" and "does security software flag it".
2. **Design token layer**: settle radius, shadow and easing before any component — otherwise four components grow four systems.
3. **Dock**: build it alone and verify the performance budget is reachable.
4. **Desktop widgets**: verify the feel of dragging and snapping.
5. **Menu bar**: status area and tray consolidation first; the global menu depends on coverage.
6. **Dynamic island**: last, since it relies on the animation foundation the earlier components establish.

---

## 18. Contributing

**Code contributions are not accepted at this stage** — there is no code to change.

What is useful now:

- **Describe your scenario**: which component do you want most? What do you use instead today, and what is wrong with it?
- **Report compatibility**: if you have hit problems with beautification tools on a specific Windows version, DPI or multi-monitor setup, describe them.
- **Suggest feel**: which macOS animation details do existing Windows tools get wrong?

Issues should state the Windows version, display configuration and scaling factor.

**Contribution discipline** (for the future):

1. Changes to design tokens must update the token layer definition; hard-coding inside components is not allowed;
2. A new dependency must be registered in NOTICE before the code is merged;
3. Code of unknown origin or with an incompatible licence must not be introduced.

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