# 构建指南 / Build Guide

## 前置条件

- Python 3.10+ 或更新版本。
- Windows 10/11 目标运行时；构建可在 macOS、Linux 或 Windows 上完成。
- 无需管理员权限，无第三方运行时依赖（PyInstaller 打包时会自动包含）。

## 本地构建

```bash
# macOS / Linux / Windows
./scripts/build.sh
```

输出目录：`release/desop-<version>/`。

## 全量自检

```bash
./scripts/check.sh
```

检查项：`black`、`flake8`、`mypy`、单元测试、交叉编译。

## 运行形态

```text
desop.exe                      启动已启用组件
desop.exe status               查看状态和配置路径
desop.exe enable dock,widgets  启用组件并持久化
desop.exe disable island       禁用组件并持久化
desop.exe settings             打开图形设置
desop.exe api-event            查看灵动岛本地 API 示例
```

托盘图标右键可开关组件、切换开机自启、退出。设置保存在 `%APPDATA%\Desop\config.json`。

## 签名与校验

每次构建会生成 `SHA256SUMS` 文件：

```bash
# 校验制品完整性
cd release/desop-<version>
sha256sum -c SHA256SUMS
```

## 构建信息

`BUILD_INFO.txt` 记录：

- 版本号
- 构建时间（UTC）
- Python 版本
- PyInstaller 版本
- Git commit hash

用于问题追踪与版本回溯。
