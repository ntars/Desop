# 安装、运行与卸载 / Install, Run and Uninstall

## 安装（用户态，便携）

1. 下载并解压 `desop-<version>.zip`，或直接运行构建脚本生成的 `release/desop-<version>/` 目录。
2. 双击 `desop.exe`。首次启动会在 `%APPDATA%\Desop\config.json` 创建默认配置。
3. 托盘出现 `Desop` 图标；右键可开关组件、打开设置、设置开机自启或退出。

可选：将程序固定到任务栏。Desop 不修改系统目录、不替换 `explorer.exe`、不写系统级注册表。

## 命令行安装（可选）

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1 -Action install
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1 -Action autostart-on
```

安装目录默认为 `%LOCALAPPDATA%\Desop`，开机自启只写 `HKCU\...\Run` 一个值。

## 配置

设置窗口可开关四个组件与开机自启；Dock、小组件位置、灵动岛端口等高级项在 `%APPDATA%\Desop\config.json` 中编辑。修改配置后重启 Desop 生效。也可运行 `desop.exe stop` 停止当前实例。

灵动岛第三方接口示例：

```powershell
Invoke-RestMethod -Method Post -Uri http://127.0.0.1:17890/event `
  -ContentType application/json `
  -Body '{"kind":"download","title":"Update","progress":42}'
```

该接口只监听 `127.0.0.1`，不对外网开放；若端口被占用，会在后续 10 个端口中自动择一，不影响桌面组件启动。

## 卸载（完整性验收项）

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\uninstall.ps1 -Action uninstall
powershell -ExecutionPolicy Bypass -File .\scripts\uninstall.ps1 -Action uninstall -NonInteractive -RemoveConfig
```

卸载脚本会：结束 Desop 进程、删除 `HKCU\...\Run` 的 `Desop` 值、删除安装目录，并询问是否保留 `%APPDATA%\Desop` 配置。它不会触碰系统目录、系统服务或全局注册表。

手动退出：托盘右键 → Exit Desop。也可在任务管理器结束 `desop.exe`。

