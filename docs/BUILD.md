# 构建指南 / Build Guide

## 前置条件

- Go 1.22 或更新版本。
- Windows 10/11 目标运行时；构建可在 macOS、Linux 或 Windows 上完成。
- 无需管理员权限，无需 CGO，无第三方运行时依赖。

## 本地构建

```bash
# macOS / Linux
./scripts/build.sh

# Windows PowerShell
$env:GOCACHE="$PWD\.cache\go-build"; $env:GOPATH="$PWD\.cache\gopath"
$env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
go build -trimpath -o release\desop.exe .\cmd\desop
```

输出目录：`release/desop-<version>/`。`VERSION` 环境变量可覆盖版本号。

## 全量自检

```bash
./scripts/check.sh
```

检查项：`gofmt`、`go vet`、全平台单元测试、Windows 交叉编译。

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

