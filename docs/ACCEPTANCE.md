# Windows 实机验收清单 / Windows Acceptance Checklist

本清单必须在 Windows 10/11 实机逐项勾选。`./scripts/check.sh` 覆盖可自动化的逻辑测试与交叉编译；`powershell -File ./scripts/acceptance.ps1 -Exe <path-to-desop.exe> -GenerateConfig` 可在干净机器上生成默认配置并输出基础 CLI/配置/API 的 PASS/FAIL 证据，不能替代 GUI、DPI、性能与安全验证。

## 环境与准备

- [ ] Windows 11 与 Windows 10 22H2 各一台（或虚拟机）。
- [ ] 单屏 1920×1080 100% 缩放。
- [ ] 双屏：主屏 3840×2160 200%，副屏 1920×1080 100%。
- [ ] 构建产物来自 `./scripts/build.sh`，`VERSION` 与 `release/VERSION` 一致。

## 主流程闭环

- [ ] 双击 `desop.exe`，5 秒内托盘出现图标，四个组件均显示。
- [ ] `desop.exe status` 输出配置路径与四个已启用组件。
- [ ] Dock：鼠标靠近时相邻图标平滑放大，移开后恢复；不再重叠。
- [ ] Dock：点击图标可通过 ShellExecute 启动对应应用，已运行应用显示圆点。
- [ ] Dock：设置 `position` 为 `left` / `right` 后重启，Dock 位于对应边缘且不出屏。
- [ ] 顶部栏：显示日期时间；任务栏托盘保持原有行为，无崩溃。
- [ ] 小组件：时钟、系统状态、便签可见；拖动便签后位置被保存，重启后恢复。
- [ ] 灵动岛：默认小胶囊；本地 API 推送事件后展开显示标题与进度，`dismiss` 后收起。
- [ ] 设置窗口：取消勾选任一组件 → 保存 → 重启，仅该组件不出现，其余组件正常。

## 配置与本地接口

- [ ] `%APPDATA%\Desop\config.json` 首次启动自动生成，JSON 可解析。
- [ ] `desop.exe enable island` / `disable island` 改配置后重启生效。
- [ ] `desop.exe stop` 能停止正在运行的实例。
- [ ] `http://127.0.0.1:17890/health` 返回 `{"ok":true}`。
- [ ] 从另一台机器访问 `http://<本机IP>:17890/health` 失败（只监听 loopback）。
- [ ] 推送非法事件（空标题、进度 101）返回 400，不崩溃。

## 全屏、DPI 与多显示器

- [ ] 全屏游戏/视频：顶部栏按配置自动隐藏；退出全屏后恢复。
- [ ] 200% 主屏 + 100% 副屏：Dock 图标大小符合各屏 DPI，无跨屏拉伸。
- [ ] 窗口跨屏拖动后，小组件仍完整位于目标屏内。

## 性能预算

- [ ] 运行 `desop.exe selftest`，10 秒空闲自检中 CPU 与工作集均输出 PASS。
- [ ] 四个组件全开空闲 2 分钟，任务管理器中 `desop.exe` CPU 平均 < 1%。
- [ ] 四个组件全开空闲 2 分钟，工作集 < 150 MB。
- [ ] 1080p、20 个图标下 Dock 放大无明显掉帧（目标 ≥ 55 fps；需录屏或性能计数器佐证）。
- [ ] 冷启动到托盘图标出现 ≤ 2 秒。

## 安全与卸载

- [ ] 进程以普通用户权限运行，未触发 UAC 提示。
- [ ] `%WINDIR%`、`Program Files` 无 Desop 写入文件。
- [ ] 开启自启后，`HKCU\Software\Microsoft\Windows\CurrentVersion\Run` 仅有 `Desop` 一个新增值。
- [ ] 执行 `uninstall.ps1 -Action uninstall`，选择不保留配置：进程结束、安装目录与 `%APPDATA%\Desop` 删除、Run 值删除。
- [ ] 非交互卸载：`uninstall.ps1 -Action uninstall -NonInteractive -RemoveConfig` 不等待输入即完成上述还原。
- [ ] 卸载后重启 Windows，无 Desop 残留进程或启动项。

## 已知限制（验收时按设计接受，不算缺陷）

- 全局菜单默认关闭：当前显示前台窗口标题作为菜单入口，不读取应用内部菜单结构；托盘收纳未实现。
- 电量状态显示 `unknown`：MVP 不引入 COM 依赖，不编造数据。
- 天气只显示“未配置数据源 / 离线”，不主动联网。
- 毛玻璃优先请求系统亚克力；若系统不支持，窗口退化为半透明/纯色，不出现黑块。

验收结果记录：将勾选结果、Windows 版本、DPI 配置与性能截图写入 GitHub Release 或 `docs/ACCEPTANCE-RESULTS.md`。

