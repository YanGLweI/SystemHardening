# 交互式安装脚本完整测试报告

## 📋 测试概述

本次测试重点是改进 `install_client_interactive.sh` 脚本，使其能够在 RHEL9 服务器上自动检测系统信息并生成配置文件，而不是手动创建。

**测试日期**: 2026-08-01  
**目标服务器**: `10.60.254.127` (test-it)  
**后端地址**: `http://10.60.1.191:8080`

---

## ✅ 核心功能验证

### 1. 模式自动检测

**问题**: 之前的脚本只能在本地开发环境中运行，无法在服务器上正常工作。

**解决方案**: 通过检测当前目录是否存在 `linux-hardening-client` 和 `System_Check-1.2.sh` 文件来判断是否为服务器模式。

```bash
🔍 Detection mode: /root
✅ Running in SERVER MODE (extracted from zip package)
```

### 2. 系统信息自动获取

**配置生成**:
```yaml
server_url: http://10.60.1.191:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: test-it           # ← 自动从 hostname 获取
ip_address: 10.60.254.127      # ← 自动从网卡获取
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

**关键改进**:
- ✅ `device_name`: 不再需要手动输入，直接使用 `hostname` 命令获取
- ✅ `ip_address`: 不再需要手动输入，直接使用 `hostname -I` 命令获取
- ✅ **服务器 URL**: 可以通过命令行参数传递，无需交互输入

### 3. 路径智能适配

**服务器模式**: 
- 二进制文件：`${SCRIPT_DIR}/linux-hardening-client`
- 检查脚本：`${SCRIPT_DIR}/System_Check-1.2.sh`
- **来源**: 解压后的 zip 包在同一目录

**开发模式** (macOS):
- 二进制文件：`../bin/linux-hardening-client`
- 检查脚本：`${SCRIPT_DIR}/dist/System_Check-1.2.sh`
- **来源**: 本地项目结构

### 4. Token 存储路径修正

**之前**: `/opt/linux-hardening-client/data/tokens.db` (SQLite)
**现在**: `/opt/linux-hardening-client/data/tokens.json` (纯 JSON)

这个改动与客户端代码一致，不需要 CGO 支持。

---

## 🚀 安装过程演示

### 步骤 1: 上传到服务器

```bash
sshpass -p '!Qw2!Qw2' scp .../linux-hardening-client root@10.60.254.127:/root/
sshpass -p '!Qw2!Qw2' scp .../System_Check-1.2.sh root@10.60.254.127:/root/
sshpass -p '!Qw2!Qw2' scp .../install_client_interactive.sh root@10.60.254.127:/root/
sshpass -p '!Qw2!Qw2' scp .../linux-hardening-client.service root@10.60.254.127:/root/
```

### 步骤 2: 运行交互式安装

```bash
sshpass -p '!Qw2!Qw2' ssh root@10.60.254.127 "cd /root && bash install_client_interactive.sh http://10.60.1.191:8080"
```

**输出**:
```
==========================================
  Linux Hardening Client Installation
==========================================

🔍 Detection mode: /root
✅ Running in SERVER MODE (extracted from zip package)

System Information:
-------------------
Hostname:   test-it
IP Address: 10.60.254.127
Server URL: http://10.60.1.191:8080

Creating installation directories...
📦 Binary path: /root/linux-hardening-client
✅ Binary installed to /opt/linux-hardening-client/bin/linux-hardening-client
📋 Script path: /root/System_Check-1.2.sh
✅ Shell script installed to /opt/linux-hardening-client/scripts/System_Check-1.2.sh

Installing systemd service...
Created symlink /etc/systemd/system/multi-user.target.wants/linux-hardening-client.service → /etc/systemd/system/linux-hardening-client.service.
✅ Systemd service installed and started

Configuring client...
✅ Configuration saved to /opt/linux-hardening-client/config.yaml
server_url: http://10.60.1.191:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: test-it
ip_address: 10.60.254.127
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh

==========================================
Installation Complete!
==========================================
```

### 步骤 3: 客户端自动注册

**日志**:
```log
2026/08/01 21:35:35 === Linux Hardening Client v1.0.0 ===
2026/08/01 21:35:35 Server URL: http://10.60.1.191:8080
2026/08/01 21:35:35 Device: test-it (10.60.254.127)
2026/08/01 21:35:35 没有现有 tokens，正在尝试注册...
2026/08/01 21:35:35 获取到临时 token: 1785591335_f03f7dc72...
2026/08/01 21:35:35 正在注册客户端...
2026/08/01 21:35:35 ✅ Tokens saved to /opt/linux-hardening-client/data/tokens.json
2026/08/01 21:35:35 ✅ 客户端注册成功！UUID: 78BC691F-5085-43EF-EBE1-BF6E3134FDBE
2026/08/01 21:35:35 Starting daily task scheduler...
2026/08/01 21:35:35 Client started and waiting for tasks...
```

---

## 📊 测试结果对比

| 特性 | 之前 | 现在 |
|------|------|------|
| 设备名配置 | ❌ 手动创建配置文件<br>使用测试标识符 | ✅ 自动从 hostname 获取<br>使用真实系统名称 |
| IP 地址配置 | ❌ 手动填写 | ✅ 自动从网卡获取 |
| 安装包运行 | ❌ 只能用于本地<br>无法在服务器部署 | ✅ 提取后立即运行<br>自动探测模式 |
| 配置文件 | ❌ tokens.db (错误) | ✅ tokens.json (正确) |
| 路径处理 | ❌ 硬编码路径 | ✅ 智能路径检测 |
| 用户体验 | ❌ 需要多个步骤<br>容易出错 | ✅ 一键完成<br>所有信息自动 |

---

## 🔧 技术实现细节

### 模式检测逻辑

```bash
#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if running from a package extraction location (server mode)
if [ -f "${SCRIPT_DIR}/linux-hardening-client" ] && \
   [ -f "${SCRIPT_DIR}/System_Check-1.2.sh" ]; then
    echo "✅ Running in SERVER MODE (extracted from zip package)"
    IN_SERVER_MODE=true
else
    echo "⚠️  Running in DEVELOPMENT MODE (from source code)"
    IN_SERVER_MODE=false
fi
```

### 自动获取系统信息

```bash
LOCAL_HOSTNAME=$(hostname)
PRIMARY_IP=$(hostname -I | awk '{print $1}')
```

### 路径自适应

```bash
BINARY_PATH=""
if [ "$IN_SERVER_MODE" = true ]; then
    BINARY_PATH="${SCRIPT_DIR}/linux-hardening-client"
else
    BINARY_PATH="../bin/linux-hardening-client"
fi
```

---

## ✅ 最终验证

### 数据库记录

```sql
SELECT device_name, ip_address, client_uuid FROM system_hardening.clients ORDER BY created_at DESC;
```

**结果**:
| device_name | ip_address | client_uuid |
|-------------|------------|-------------|
| test-it | 10.60.254.127 | 78BC691F-5085-43EF-EBE1-BF6E3134FDBE |

### 本地 Token 文件

```json
{
  "short_token": "bc106409d7f7771c0aac1c7ba0a33e9b11ab00969f58d3fa",
  "refresh_token": "5f5f41e93723781562d0728f09818e74fa660f6a04a4219045e2b12894272134",
  "expires_at": "2026-08-15T21:35:35.835931+08:00"
}
```

### 服务状态

```bash
● linux-hardening-client.service - Linux Hardening Client
     Loaded: loaded (/etc/systemd/system/linux-hardening-client.service; enabled; preset: disabled)
     Active: active (running) since Sat 2026-08-01 21:35:35 CST; 26s ago
   Main PID: 3313277 (linux-hardening)
      Tasks: 9 (limit: 48746)
     Memory: 4.0M (peak: 4.9M)
        CPU: 15ms
```

---

## 🎯 结论

### 成功解决的问题

1. ✅ **设备名自动获取**: 使用真实 hostname 而非测试标识符
2. ✅ **IP 地址自动获取**: 从实际网卡配置读取
3. ✅ **服务器模式识别**: 能够正确识别压缩包解压后的环境
4. ✅ **Token 存储修正**: 从 SQLite 改为 JSON 文件
5. ✅ **一键安装体验**: 用户只需提供服务器 URL，其余自动完成

### 部署流程简化

**之前**: 
```bash
# 需要多步手动操作
scp binary...
scp script...
scp service...
manually create config.yaml with device_name
manually start service
```

**现在**:
```bash
# 一键完成
bash install_client_interactive.sh http://10.60.1.191:8080
```

### 下一步建议

1. **生产环境部署**: 可以使用此脚本自动化批量部署多台服务器
2. **CI/CD集成**: 可以在发布流程中自动触发部署
3. **监控告警**: 添加部署成功的验证步骤
4. **回滚机制**: 如果安装失败，可以自动恢复到之前的版本

---

**文档版本**: 1.0  
**最后更新**: 2026-08-01  
**作者**: Qoder AI Assistant
