# Linux Hardening Client - RHEL 9

## 概述

Linux Hardening Client 是一个用于 RHEL 9 系统的自动化安全加固工具。它定期执行安全配置检查，并将结果回传到中央服务器进行管理。

## 架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────┐
│ RHEL 9      │         │ Go Client    │         │ Server  │
│ Server      │ ←────→  │ (守护进程)   │ ←────→  │ MySQL   │
│             │ Shell   │              │ HTTP    │         │
│ • System_   │ Script  │ • TokenMgr   │ API     │ Web UI  │
│   Check.sh  │         │ • Scheduler  │         │ LDAP    │
│ • mysql-    │         │ • Uploader   │         │         │
│   insert.sh │         │ • Logger     │         │         │
└─────────────┘         └──────────────┘         └─────────┘
```

## 功能特性

### ✅ 核心功能

1. **自动加固检查**
   - 执行系统安全配置检查脚本
   - 收集所有加固检查结果
   - 验证配置合规性

2. **Token 认证管理**
   - 临时 Token 申请（5 分钟有效期）
   - 客户端注册与 Token 刷新
   - JSON 文件本地存储（权限 0600）
   - 自动续期（提前 24 小时预警）

3. **数据上报**
   - 每天自动上传一次
   - 失败重试机制
   - 完整日志记录

4. **Service 管理**
   - systemd 服务集成
   - 开机自启
   - 自动重启

### 🔧 支持的检查项

- 系统信息（主机名、内核、OS 版本）
- 软件包 GPG 签名验证（dnf.conf, redhat.repo）
- 密码策略（PASS_MAX_DAYS, PASS_MIN_LEN, 复杂度要求）
- 用户账户管理（INACTIVE, GID）
- 计划任务配置（cron, at）
- SSH 安全配置（LogLevel, MaxAuthTries, PermitRootLogin）
- PAM 密码质量设置（minlen, minclass, dcredit）
- 文件权限（passwd, shadow, group 等）
- 加密策略（NO-SHA1, NO-WEAKMAC）
- NTP 时间同步

## 安装部署

### 前置条件

1. **RHEL 9 服务器**
2. **依赖包**：
   ```bash
   yum install curl jq
   ```

3. **后端服务运行中**
   - 地址：`http://10.60.254.127:8080`（请修改为实际地址）

### 安装步骤

#### Step 1: 构建客户端（可选）

如果还没有编译好的二进制文件：

```bash
cd /path/to/system_hardening/client
GOOS=linux GOARCH=amd64 go build -o ../bin/linux-hardening-client .
```

#### Step 2: 复制安装包到目标服务器

```bash
# 方式 1: scp
scp dist/linux-hardening-client_XXX.zip root@your-rhel-server:/tmp/

# 方式 2: 直接复制到当前目录
cp dist/linux-hardening-client_XXX.zip /tmp/
cd /tmp && unzip linux-hardening-client_XXX.zip
```

#### Step 3: 运行安装脚本

```bash
sudo bash /tmp/install_client_interactive.sh http://YOUR_SERVER_IP:8080
```

**参数说明**:
- `SERVER_URL`: 后端服务地址（不提供则交互式输入）

#### Step 4: 验证安装

```bash
# 查看服务状态
systemctl status linux-hardening-client

# 查看日志
journalctl -u linux-hardening-client -f

# 测试手动执行（可选）
/opt/linux-hardening-client/bin/linux-hardening-client
```

## 配置文件

位置：`/opt/linux-hardening-client/config.yaml`

```yaml
server_url: http://YOUR_SERVER_IP:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: your-hostname
ip_address: 192.168.1.100
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

## API 接口

### 1. 请求临时 Token

**POST** `/api/client/request-temp-token`

**Request**:
```json
{
  "device_name": "your-hostname",
  "ip_address": "192.168.1.100"
}
```

**Response**:
```json
{
  "temp_token": "abc123...",
  "expires_in": 300,
  "expires_at": "2026-08-01T12:00:00Z"
}
```

### 2. 注册客户端

**POST** `/api/client/register`

**Request**:
```json
{
  "temp_token": "abc123...",
  "device_name": "your-hostname",
  "ip_address": "192.168.1.100",
  "os_version": "Red Hat Enterprise Linux 9.4"
}
```

**Response**:
```json
{
  "client_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "short_token": "def456...",
  "refresh_token": "ghi789...",
  "expires_at": "2026-08-15T00:00:00Z"
}
```

### 3. 刷新 Token

**POST** `/api/client/refresh-token`

**Request**:
```json
{
  "refresh_token": "ghi789..."
}
```

**Response**:
```json
{
  "short_token": "jkl012...",
  "expires_at": "2026-08-15T00:00:00Z"
}
```

### 4. 上传加固数据

**POST** `/api/client/upload-data`

**Headers**:
```
Content-Type: application/json
X-Client-Token: jkl012...
```

**Request**:
```json
{
  "data": {
    "date": "2026/08/01_00:00:00",
    "hostname": "your-hostname",
    "operasystem": "Red Hat Enterprise Linux release 9.4",
    "kernel": "5.14.0-xxx.el9.x86_64",
    "ip": "192.168.1.100",
    // ... 其他字段
  }
}
```

## 维护操作

### 启动/停止/重启

```bash
sudo systemctl start linux-hardening-client
sudo systemctl stop linux-hardening-client
sudo systemctl restart linux-hardening-client
```

### 启用/禁用开机自启

```bash
sudo systemctl enable linux-hardening-client
sudo systemctl disable linux-hardening-client
```

### 查看日志

```bash
# 实时日志
journalctl -u linux-hardening-client -f

# 最近日志
journalctl -u linux-hardening-client -n 100

# 错误日志
journalctl -u linux-hardening-client -p err
```

### 卸载

```bash
# 停止并移除服务
sudo systemctl stop linux-hardening-client
sudo systemctl disable linux-hardening-client
sudo rm -rf /opt/linux-hardening-client
sudo rm /etc/systemd/system/linux-hardening-client.service
sudo systemctl daemon-reload
```

## Token 生命周期

```
Installation → Temp Token (5 min) → Register → Short Token (14 days)
                                              ↓
                                   Auto Refresh (24h before expiry)
                                              ↓
                                   If still failed → Reinstall required
```

### Token 刷新策略

1. **正常情况**: 自动在到期前 24 小时刷新
2. **失败处理**: 
   - 第 1 次失败：重试
   - 连续失败：记录错误日志
   - 永久失效：需要重新运行 install_client_interactive.sh

## 故障排查

### 常见问题

#### 1. Token 过期或无效

**症状**:
```
TOKEN Token expired or expiring, attempting refresh...
ERROR Token refresh failed
```

**解决方案**:
```bash
# 重新安装
sudo bash /tmp/install_client_interactive.sh http://YOUR_SERVER_IP:8080
```

#### 2. 脚本执行失败

**症状**:
```
ERROR Script execution failed
```

**解决方案**:
```bash
# 检查脚本是否可执行
ls -l /opt/linux-hardening-client/scripts/System_Check-1.2.sh
chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh

# 手动测试脚本
bash /opt/linux-hardening-client/scripts/System_Check-1.2.sh 2>&1 | tail -20
```

#### 3. 网络问题无法连接服务器

**症状**:
```
ERROR HTTP request failed
```

**解决方案**:
```bash
# 测试网络连接
curl -v http://10.60.254.127:8080/api/client/request-temp-token

# 检查防火墙规则
firewall-cmd --list-all
# 或暂时关闭防火墙测试
systemctl stop firewalld
```

#### 4. 服务未启动

**症状**:
```
Unit linux-hardening-client.service could not be found.
```

**解决方案**:
```bash
# 重新安装
sudo bash /tmp/install_client_interactive.sh http://YOUR_SERVER_IP:8080

# 或手动启动服务
sudo systemctl daemon-reload
sudo systemctl start linux-hardening-client
```

### 调试模式

```bash
# 前台运行（便于调试）
sudo /opt/linux-hardening-client/bin/linux-hardening-client &

# 查看详细日志
sudo journalctl -u linux-hardening-client -f --no-pager
```

## 性能影响

- **CPU 使用率**: < 5%（仅在每日运行时）
- **内存占用**: ~20MB
- **磁盘空间**: ~5MB（程序 + 脚本 + 日志）
- **网络流量**: ~2KB（每次数据上报）

## 安全说明

1. **Token 安全性**
   - 临时 Token 仅使用一次，立即失效
   - 短期 Token 采用加密存储
   - 敏感数据不记录在日志中

2. **通信安全**
   - 建议在生产环境使用 HTTPS
   - Token 通过 HTTPS 传输防止中间人攻击

3. **数据完整性**
   - 所有配置更改都有备份文件（.bak）
   - 支持审计追踪

## 升级

### 新二进制文件

```bash
# 停止服务
sudo systemctl stop linux-hardening-client

# 替换二进制
sudo cp new-binary /opt/linux-hardening-client/bin/
sudo chmod +x /opt/linux-hardening-client/bin/linux-hardening-client

# 重新启动
sudo systemctl start linux-hardening-client
```

## 技术支持

如有问题，请提供以下信息：

1. 服务日志：`journalctl -u linux-hardening-client -n 100 --no-pager`
2. 配置信息：`/opt/linux-hardening-client/config.yaml`
3. 系统版本：`cat /etc/redhat-release`
4. 脚本输出（最近一次运行）：`/opt/linux-hardening-client/logs/client.log`

---

**版本**: 1.0.0  
**最后更新**: 2026-08-01
