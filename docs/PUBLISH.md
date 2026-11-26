# Desop 发布指南

## M5 正式发布检查单

### 1. 签名 + 分发

| 检查项 | 状态 |
|-------|------|
| Windows 11 x64 exe 构建成功 | ⏳ |
| 签名证书可用 | ⏳ |
| SHA256SUMS 生成 | ⏳ |
| CHANGELOG.md 版本条目完整 | ✅ |

### 2. 干净 Win11 虚拟机安装

| 检查项 | 状态 |
|-------|------|
| Windows 11 ARM64/AMD64 VM 创建 | ⏳ |
| 安装 Desop | ⏳ |
| 启动托盘图标显示 | ⏳ |
| Dock/顶部栏/小组件/灵动岛显示 | ⏳ |

### 3. 卸载验证

| 检查项 | 状态 |
|-------|------|
| 卸载脚本执行成功 | ⏳ |
| HKCU Run 项删除 | ⏳ |
| 安装目录清空 | ⏳ |
| %APPDATA%\Desop 清空 | ⏳ |

### 4. 文档完整性

| 检查项 | 状态 |
|-------|------|
| README.md 更新 | ✅ |
| docs/BUILD.md | ✅ |
| docs/INSTALL.md 完整 | ⏳ |
| docs/SUPPORT.md 完整 | ✅ |
| docs/KNOWN-ISSUES.md 完整 | ✅ |
| CHANGELOG.md 首版完整 | ✅ |

## 实机验收清单

### 环境准备

- [ ] Windows 11 ARM64/AMD64 虚拟机
- [ ] Python 3.10+ 可用
- [ ] 网络连接

### 基础功能测试

- [ ] 双击 desop.exe 启动
- [ ] 托盘图标显示
- [ ] Dock 栏显示（底部）
- [ ] 顶部栏显示
- [ ] 小组件显示（时钟/系统状态）
- [ ] 灵动岛显示

### 交互测试

- [ ] 鼠标悬停 Dock 图标放大
- [ ] 点击图标启动应用
- [ ] 设置窗口打开/保存
- [ ] 组件开关生效

### 配置验证

- [ ] config.json 生成于 %APPDATA%\Desop
- [ ] 开机自启添加 HKCU Run 键值

### 卸载验证

- [ ] uninstall.ps1 执行成功
- [ ] 进程结束
- [ ] 目录清空
- [ ] 注册表项删除

## 发布命令

```bash
# 构建
./scripts/build.sh

# 校验
python tools/docs_guard.py
pytest tests/ -q

# 安装到 VM 共享目录
cp -r release/desop-* /path/to/vm/shared/

# VM 内执行
cd \\Mac\DesopBuild
.\desop.exe
```
