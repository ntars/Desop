**简体中文** | [English](README_EN.md)

# Desop

**Windows 端桌面美化套装：把 macOS 的桌面观感复刻到 Windows，提供 dock 栏、顶部栏、桌面小组件、灵动岛等元素。**

> **当前状态：首个可构建版本 / Buildable MVP。** Windows 10/11 用户态实现已落地（纯 Go 标准库，零第三方运行时依赖）。构建与运行见 [`docs/BUILD.md`](docs/BUILD.md)、[`docs/INSTALL.md`](docs/INSTALL.md)；Windows 实机验收见 [`docs/ACCEPTANCE.md`](docs/ACCEPTANCE.md)。

![Status](https://img.shields.io/badge/status-buildable%20MVP-2ea44f?style=flat-square)
[![GitHub license](https://img.shields.io/github/license/ntars/Desop?style=flat-square)](./LICENSE)
[![GitHub Star](https://img.shields.io/github/stars/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/stargazers)
[![GitHub Fork](https://img.shields.io/github/forks/ntars/Desop?style=flat-square)](https://github.com/ntars/Desop/network/members)
![GitHub Repo size](https://img.shields.io/github/repo-size/ntars/Desop?style=flat-square&color=3cb371)

[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d4.svg?style=flat-square&logo=windows&logoColor=white)](#)
[![Runtime](https://img.shields.io/badge/runtime-user%20mode-2ea44f.svg?style=flat-square)](#)
[![System Changes](https://img.shields.io/badge/system%20changes-none-8957e5.svg?style=flat-square)](#)

---

## 一、项目名称与定位

> **仓库描述 / Description**：`Desop —— Windows 端桌面美化套装：把 macOS 桌面观感复刻到 Windows，提供 dock 栏、顶部栏、桌面小组件与灵动岛 | A Windows desktop beautification suite: macOS-style dock, menu bar, widgets and dynamic island. Buildable MVP.`
>
> **话题标签 / Topics**（20 个，仅英文）：`beautification`, `desktop-customization`, `desktop-widgets`, `dock`, `dynamic-island`, `macos-style`, `menubar`, `mvp`, `theme`, `ui`, `user-mode`, `windows`, `windows-10`, `windows-11`, `desktop-environment`, `taskbar`, `wallpaper`, `theming`, `windows-desktop`, `concept`


**名称**：Desop。取 Desktop 的 Des 与 op（operation / 操作层）——桌面操作层。

**一句话定位**：在 Windows 上重建一套 macOS 式的桌面交互层，让「换系统」不再等于「换习惯」。

**定位的三条边界**：

1. **它是桌面层，不是系统层**。所有元素都跑在用户态，不改系统文件、不替换 shell、不需要禁用系统更新。卸载即还原。
2. **它追求观感与手感，不追求像素级复刻**。macOS 的圆角、毛玻璃、动效曲线可以学，但 Windows 的窗口管理、任务栏行为是系统契约，不硬改。
3. **它是一套可拆的组件，不是一个整包**。dock、顶部栏、小组件、灵动岛各自独立，用户可以只开其中一两个。

**与 macOS 的关系**：本项目参照 macOS 的桌面交互设计，不复用其任何代码、素材或字体。所有视觉元素为独立实现。

---

## 二、项目简介

### 2.1 背景

从 macOS 转到 Windows 的人，最先不适应的往往不是功能缺失，而是**交互习惯的断裂**：没有顶部全局菜单、没有 dock 的图标放大与弹跳、没有随手的桌面小组件、通知只有右下角一个气泡。

现有方案通常两个极端：要么是重度美化包，改注册表、换 shell、动系统文件，出问题难恢复；要么是单个小工具，只管 dock 或只管小组件，风格各自为政，凑在一起反而更乱。

### 2.2 要解决的问题

1. **观感割裂**。四个元素分别来自不同工具时，圆角半径、阴影、动效时长、字体全不一致，看起来像四个软件拼在一起。
2. **风险不可控**。改系统文件的方案一旦出问题，恢复成本高，普通用户不敢尝试。
3. **性能负担**。毛玻璃与实时模糊如果实现粗糙，会让整机掉帧，美化变成负担。

### 2.3 核心思路

统一一套设计令牌，四个组件共享：

| 令牌 | 作用 |
|---|---|
| 圆角与间距 | 所有组件用同一套数值，不各写一份 |
| 阴影与材质 | 毛玻璃的模糊半径、透明度、边界光统一 |
| 动效曲线 | 时长与缓动函数统一，避免「有的弹有的滑」 |
| 字体与字号 | 跟随系统但统一档位 |

四个组件各自独立进程或模块，共享令牌，互不依赖——关掉一个不影响其余。

### 2.4 适用场景

- 从 macOS 转到 Windows、想保留原有操作习惯的用户；
- 喜欢整洁桌面、希望有 dock 与小组件但不想装重型美化包的用户；
- 需要长期稳定运行、不能接受「美化导致系统异常」的用户。

### 2.5 非目标

- 不改系统文件、不替换 `explorer.exe`、不修改受保护的系统设置；
- 不做主题商店、不做壁纸分发；
- 不做性能监控、硬件信息类小组件的深度数据采集（只读系统公开接口）；
- 不承诺与 macOS 行为完全一致——系统契约不同的地方以 Windows 为准。

---

## 三、核心功能

**当前状态：已实现 MVP 功能以 ✅ 标注；待实机验收或受 Windows 接口限制的能力以 ⚠ 标注。**

### 3.1 Dock 栏

| 能力 | 说明 | 状态 |
|---|---|---|
| 图标放大 | 指针靠近时相邻图标按距离衰减放大 | ✅ 已实现（平滑衰减，范围受限） |
| 运行指示 | 已运行应用下方显示指示点 | ✅ 已实现（只读进程快照） |
| 弹跳反馈 | 启动中弹跳，可关闭 | ✅ 已实现启动阻尼动画 |
| 自动隐藏 | 贴边隐藏，触发区宽度可调 | ✅ 已接入基础逻辑，待实机验收 |
| 位置 | 底部 / 左侧 / 右侧 | ✅ 已实现 |
| 多显示器 | 每个显示器独立 dock 或仅主屏 | ⚠ 当前使用主屏，副屏策略待实机确认 |

### 3.2 顶部栏

| 能力 | 说明 | 状态 |
|---|---|---|
| 全局菜单 | 显示当前前台窗口的标题作为菜单入口 | ✅ 已实现；受窗口标题可用性限制，默认关闭 |
| 状态区 | 时间、日期、电量、网络、音量 | ✅ 时间/日期/网络已实现；电量、音量诚实降级 |
| 托盘收纳 | 把系统托盘图标收进顶部栏 | ⚠ 默认关闭，未实现 |
| 自动隐藏 | 全屏应用时自动隐藏 | ✅ 已接入保守检测，待实机验收 |

### 3.3 桌面小组件

| 能力 | 说明 | 状态 |
|---|---|---|
| 时钟日历 | 大号时钟、月历、日程摘要 | ✅ 时钟与月历已实现；日程未实现 |
| 系统状态 | CPU / 内存 / 磁盘 / 网络速率 | ✅ 已实现（公开 Win32 只读接口） |
| 便签 | 本地纯文本便签，不联网 | ✅ 已实现并持久化 |
| 天气 | 需用户自行填入数据源，默认不联网 | ✅ 离线降级；不主动联网 |
| 布局 | 自由拖拽、吸附对齐、网格吸附可关 | ✅ 已实现拖拽与网格吸附，待实机验收 |

### 3.4 灵动岛

| 能力 | 说明 | 状态 |
|---|---|---|
| 常驻形态 | 屏幕顶部居中的胶囊，收起时极小 | ✅ 已实现 |
| 展开动效 | 点击或事件触发时展开 | ✅ 已实现即时展开；弹性曲线待优化 |
| 事件接入 | 音乐播放、下载进度、通知摘要、计时器 | ✅ 本地 API 支持 media/download/timer/custom；系统媒体会话未接入 |
| 第三方接入 | 提供本地接口，允许其他程序推送状态 | ✅ 仅监听 127.0.0.1 |

### 3.5 明确不做

- 不做壁纸引擎与动态壁纸；
- 不做窗口特效强制注入；
- 不做系统主题色替换；
- 不做开机动画修改。

---

## 四、核心特性

**当前状态：实现已落地，证据以源码与自动化测试为主；性能与 GUI 证据待 Windows 实机补齐。**

### 亮点一：零系统改动，卸载即还原

- **实现方式**：不写系统目录、不替换 shell、不注入 explorer；开机自启仅写 HKCU Run 一个值，可由卸载脚本删除。
- **证据**：`internal/win32/app_windows.go`、`scripts/uninstall.ps1`、`docs/ACCEPTANCE.md`（卸载验收项）。

### 亮点二：一套设计令牌，四个组件风格一致

- **实现方式**：圆角、间距、模糊、透明度、动效与字体集中定义，组件只读。
- **证据**：`internal/tokens/tokens.go` 与 `internal/tokens/tokens_test.go`。

### 亮点三：性能预算作为发布门槛

- **实现方式**：组件按需开关、定时器节流、Dock 指针变化才重绘。
- **证据**：逻辑测试与交叉编译见 `scripts/check.sh`；CPU/内存/帧率需按 `docs/ACCEPTANCE.md` 实机测量。

### 亮点四：组件可选，不做捆绑

- **实现方式**：四个组件配置独立，CLI、托盘与设置窗口可分别开关。
- **证据**：`internal/config/config_test.go`、`internal/app/settings_test.go` 与 Windows 实机验收清单。

### 亮点五：复用系统能力优先，不硬造

- **实现方式**：优先请求 DWM 系统亚克力；不支持时退化为半透明 GDI 绘制。
- **证据**：`internal/win32/win32_windows.go` 的 `DwmSetSystemBackdrop` 与 `SetLayeredWindowAttributes`。

---

## 五、设计细节与工程取舍

**本章为设计约定，实现时须遵守。**

### 5.1 灵动岛的事件来源要诚实

灵动岛不是系统能力，因此必须明确每个事件的来源：

| 事件 | 来源 | 局限 |
|---|---|---|
| 音乐播放 | 系统媒体会话（SMTC） | 部分播放器不上报 |
| 下载进度 | 应用主动推送（本地接口） | 需应用配合 |
| 通知摘要 | 系统通知监听 | 需用户授权 |
| 计时器 | 本程序内部 | 仅本程序 |

**不做的事**：不假装能拿到所有应用的状态。拿不到就不显示，而不是编一个。

### 5.2 全屏应用的避让

- 全屏应用运行时，顶部栏与 dock 自动隐藏；
- 检测方式优先用系统提供的全屏状态，而非窗口尺寸猜测；
- 检测不到时宁可保留，不做激进隐藏——误隐藏比多显示更烦人。

### 5.3 DPI 与多显示器的边界

- 每个显示器独立计算缩放比例，不共用一套坐标；
- 跨屏拖动时按目标屏的 DPI 重算，不做插值动画——插值会让图标在跨屏瞬间变形；
- 混合 DPI（如 4K 主屏 + 1080p 副屏）是必须测试的场景。

### 5.4 毛玻璃的降级链

```
系统亚克力（Win11） → 系统模糊（Win10） → 半透明纯色 → 不透明
```

每一级都要能自动降级，且降级后视觉仍成立。不允许出现「系统不支持就变成一块黑」。

### 5.5 图标放大的性能边界

- 只对可见区域的图标计算放大，屏幕外的跳过；
- 放大范围有上限（默认指针左右各 3 个图标）；
- 走合成层变换，不触发布局重排。

### 5.6 卸载的完整性

卸载后必须还原：

1. 开机自启项；
2. 注册表启动项；
3. 桌面快捷方式；
4. 用户数据目录（询问是否保留配置）。

**这一条是验收项，不是可选项。**

---

## 六、技术栈与依赖清单

**技术选型已确定：Go 1.22+ 标准库 + 公开 Win32 API，零第三方运行时依赖。**

| 约束 | 内容 |
|---|---|
| 运行环境 | Windows 10 / 11，用户态，不需管理员权限 |
| 系统改动 | 零。不写系统目录、不替换 shell |
| 界面渲染 | 原生窗口 + GDI + DWM 系统背景 |
| 系统接口 | 公开 Win32 API；只读系统状态 |
| 本地接口 | 回环 HTTP（127.0.0.1），可被配置端口 |
| 许可 | Apache-2.0（内核与接口）/ MIT（外壳） |

依赖清单：运行时零第三方依赖；构建期使用 Go 工具链（BSD-3-Clause）。详见 [NOTICE](NOTICE)。

---

## 七、安装与部署

已提供 Windows 便携版目录构建脚本与 PowerShell 安装/卸载脚本。

| 平台 | 形态 | 状态 |
|---|---|---|
| Windows 11 | 便携版 `desop.exe` 目录 | ✅ 可构建，待实机验收 |
| Windows 10 | 同上 | ✅ 可构建，待实机验收 |

构建：`./scripts/build.sh`。安装与卸载见 [`docs/INSTALL.md`](docs/INSTALL.md)，实机验收见 [`docs/ACCEPTANCE.md`](docs/ACCEPTANCE.md)。

---

## 八、快速开始

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

托盘图标右键可开关组件、切换开机自启、重启或退出。完整步骤见 [`docs/INSTALL.md`](docs/INSTALL.md)。

---

## 九、配置说明

配置文件：`%APPDATA%\Desop\config.json`；首次启动自动生成。图形设置窗口与 CLI 均可修改组件开关。

| 类别 | 内容 | 状态 |
|---|---|---|
| 组件开关 | 四个组件各自的启用/禁用 | ✅ |
| 外观 | 统一令牌；主题色/模糊强度档位 | ⚠ 令牌集中定义，UI 档位待后续 |
| 行为 | 自动隐藏、触发区宽度、开机自启 | ✅ 部分已实现 |
| 高级 | 调试日志、性能诊断 | ⚠ 未实现 |

---

## 十、API 与模块说明

| 模块 | 职责 | 状态 |
|---|---|---|
| `internal/tokens` | 圆角、阴影、动效曲线唯一定义 | ✅ |
| `internal/dock` | 图标布局、放大、运行指示 | ✅ |
| `internal/topbar` | 状态聚合、布局、自动隐藏策略 | ✅ 部分 |
| `internal/widgets` | 时钟、系统状态、便签、天气降级 | ✅ |
| `internal/island` | 胶囊、本地事件接口 | ✅ |
| `internal/win32` | 原生窗口、托盘、系统只读接口 | ✅ |
| `cmd/desop` | 启动入口与 CLI | ✅ |

本地 API：`POST http://127.0.0.1:17890/event`、`POST /dismiss`、`GET /health`。示例：`desop.exe api-event`。

---

## 十一、目录结构

```text
Desop/
├── cmd/desop/          # 程序入口与 CLI
├── internal/           # tokens、config、四组件逻辑、Win32 层
├── docs/               # BUILD / INSTALL / ACCEPTANCE
├── scripts/            # build/check/install/uninstall
├── release/            # 构建产物（gitignore）
├── README.md / README_EN.md
├── LICENSE / LICENSE-MIT / NOTICE / LICENSING.md
└── CHANGELOG.md
```

---

## 十二、性能表现

**自动化结果与 Windows 实机数据分开记录。** 逻辑测试、格式检查与交叉编译见 `./scripts/check.sh`；以下性能目标仍需在 Windows 实机按 `docs/ACCEPTANCE.md` 测量：

| 指标 | 目标 | 状态 |
|---|---|---|
| 空闲 CPU | < 1%（四组件全开） | 待 Windows 实机 |
| 空闲内存 | < 150 MB | 待 Windows 实机 |
| Dock 放大帧率 | ≥ 55 fps（1080p，20 图标） | 待 Windows 实机 |
| 冷启动 | ≤ 2 s | 待 Windows 实机 |

未达标不发布稳定版；不会用“先发出去再优化”代替验收。

---

## 十三、安全与隐私

| 维度 | 当前实现 |
|---|---|
| 系统改动 | 零。用户态运行，不写系统目录 |
| 权限 | 不需要管理员权限 |
| 网络 | 默认只监听本机回环；天气不主动联网 |
| 数据收集 | 无遥测、无统计、无崩溃上报 |
| 本地存储 | 仅存配置与便签文本 |
| 通知监听 | 未实现；不请求授权 |

**明确不做**：不上传任何数据，不读取用户文件内容，不在后台静默联网。

---

## 十四、兼容性说明

| 项 | 当前状态 |
|---|---|
| Windows 11 | 可构建，待实机验收 |
| Windows 10（1809+） | 可构建，部分模糊效果降级 |
| 多显示器 | 主屏 MVP；混合 DPI 待实机验收 |
| 高 DPI（125% / 150% / 200%） | 代码使用工作区坐标，需实机验证 |
| Windows 7 / 8 | 不支持 |

---

## 十五、已知问题与限制

### 15.1 当前阶段的事实

1. **已有可构建代码**：Go 标准库 Windows 用户态实现。
2. **已有构建与部署脚本**：`scripts/build.sh`、`scripts/check.sh`、PowerShell 安装/卸载。
3. **Windows 实机验收待执行**：本仓库在 macOS 环境中完成逻辑测试与交叉编译，但 GUI/DPI/性能必须由 Windows 实机证明。

### 15.2 已识别的设计限制

| 限制 | 原因 | 是否可解 |
|---|---|---|
| 全局菜单覆盖面有限 | Windows 应用不强制暴露菜单结构 | 不可解，只能尽力 |
| 灵动岛是软件模拟 | Windows 无对应系统机制 | 不可解，事件来源受限 |
| 部分模糊效果依赖系统版本 | Win10 与 Win11 能力不同 | 通过降级链缓解 |
| 可能被安全软件误判 | 涉及系统级钩子的功能 | 需提前验证，必要时放弃该功能 |

### 15.3 排期阶段必须先回答的问题

答不上就不进入开发：

| 问题 | 为什么关键 |
|---|---|
| 各 Windows 版本的模糊效果可用性 | 决定毛玻璃是否需要自绘兜底 |
| 全局菜单在常见应用上的覆盖率 | 决定这个功能是否值得做 |
| 多显示器与 DPI 缩放的边界情况 | 决定布局计算复杂度 |
| 是否会被反作弊 / 安全软件误判 | 涉及系统级钩子的功能必须提前确认 |
| 卸载后能否完全还原 | 决定用户是否敢装 |

---

## 十六、更新日志

首个可构建版本的变更记录见 [CHANGELOG.md](CHANGELOG.md)。发行版记录见 <https://github.com/ntars/Desop/releases>。

---

## 十七、后续路线图

当前版本为 MVP；按 Windows 实机验收结果推进：

1. 修复验收清单中发现的阻断问题；
2. 优化 Dock 动效曲线与弹性展开；
3. 视覆盖率决定是否实现全局菜单与托盘收纳；
4. 补充更多小组件与多显示器策略。

---

## 十八、参与贡献

欢迎提交场景、兼容性与手感反馈。代码贡献请先阅读 `docs/ACCEPTANCE.md`，确保改动可验证。

贡献纪律：

1. 设计令牌的改动必须同步令牌层定义，不允许组件内硬编码；
2. 新增依赖必须先在 NOTICE 登记，再合入代码；
3. 不得引入来源不明或许可证不兼容的代码。

---

## 十九、版权与许可

Copyright (c) 2024–2026 Ntarstudio（星玖工作室）

本项目按目录分层授权，**只使用 Apache License 2.0 与 MIT 两种宽松许可证**，全文见 [LICENSE](LICENSE)（Apache-2.0）与 [LICENSE-MIT](LICENSE-MIT)（MIT），分层说明见 [LICENSING.md](LICENSING.md)。

| 路径 | 许可 |
|---|---|
| 桌面层内核、本地接口与协议（代码落地后的模块） | Apache-2.0 |
| 组件外壳、安装器、构建脚本（代码落地后的模块） | MIT |
| `README.md` `README_EN.md` `LICENSING.md` `NOTICE` | MIT |
| 品牌名称、Logo、界面设计 | 保留所有权利 |

**默认归属**：上表未列出的本项目自研文件，一律适用根 `LICENSE`（Apache-2.0）。目录与路径分层见 [LICENSING.md](LICENSING.md)。

两种许可都可商用、可修改、可闭源、可再许可，**没有传染性**：你的衍生作品无需开源。再分发时保留版权声明、许可全文与 [NOTICE](NOTICE) 即可。Apache-2.0 另外提供明示专利授权（第 3 条）。

**为什么不用 AGPL**：本项目提供本地 HTTP 接口供第三方程序推送灵动岛状态，但这不构成「网络服务」——接口跑在用户自己的机器上，不对公网提供。AGPL 在这里唯一的效果是劝退插件作者与企业用户；而本项目最需要的恰恰是插件生态。

**名称与商标**。「Ntarstudio」「星玖工作室」「Desop」等名称与标识归星玖工作室所有。代码许可只给代码权利，不给商标权。

**第三方名称**。文档中出现的 macOS、Windows、Dock、灵动岛（Dynamic Island）、explorer.exe 等名称归各自权利人，仅用于说明设计参照对象与兼容目标，不表示任何隶属或背书关系。

**无担保**。本项目处于排期阶段，不提供任何形式的可用性承诺。程序（未来）按「现状」提供，不附带任何明示或暗示的担保。

---

## 二十、NOTICE 开源合规声明

**当前无第三方依赖，NOTICE 已建为占位文件。** 完整声明见 [NOTICE](./NOTICE)，其中包含：当前零依赖的事实、参照对象声明、第三方组件登记要求（七字段）、四条纪律、**图标与字体的额外要求**（美化类项目最易在素材上出问题）、许可证兼容性对照表。

**已声明但未使用其代码的参照对象**：

| 对象 | 权利人 | 与本项目的关系 |
|---|---|---|
| macOS 桌面交互设计 | Apple Inc. | **仅作设计参照，未使用任何代码、素材、字体或图标** |
| Windows 系统 API | Microsoft Corporation | 仅通过公开接口调用，不修改系统组件 |

**引入第三方组件时的登记要求**（未来适用）：

任何新增的第三方库、框架、字体、图标、算法来源，必须先在此登记：名称、版本、版权方、许可证类型、官方地址、在本项目中的用途、是否修改及修改说明。**无法确认许可证归属的，标注「待确认」并暂不合入，不得默认或编造。**

间接依赖与传递依赖同样必须登记，不得以「不是直接引用」为由省略。

---

Ntarstudio（星玖工作室）· <https://www.ntars.cn/>