# Linux 加固客户端 - RHEL 9

## 概述

Linux 加固客户端是部署在 RHEL 9 服务器上的自动化安全加固工具。它以 systemd 服务方式运行，定期执行安全配置检查与加固，并将结果上报到后端服务器进行集中管理。

## 架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│ RHEL 9      │         │ Go 客户端     │         │ 后端服务     │
│ 目标服务器   │ ←────→  │ (守护进程)    │ ←────→  │             │
│             │ Shell   │              │ HTTP    │ • Gin API   │
│ System_     │ 脚本    │ • TokenMgr   │ API     │ • MariaDB   │
│ Check-1.2.sh│         │ • Scheduler  │         │ • Web 前端   │
│             │         │ • Uploader   │         │             │
└─────────────┘         └──────────────┘         └─────────────┘
```

## 功能特性

### 核心功能

1. **自动加固检查**
   - 启动后立即执行一次安全检查
   - 之后每 24 小时自动执行一次
   - 自动加固不合规的配置项

2. **Token 认证管理**
   - 首次启动自动注册（临时 Token → 正式 Token）
   - Token 存储在本地 JSON 文件（权限 0600）
   - 到期前 24 小时自动刷新

3. **数据上报**
   - 每次检查完成后自动上传结果
   - 后端按 client_uuid 去重（UPDATE 而非重复 INSERT）

4. **服务管理**
   - systemd 服务集成
   - 开机自启、异常自动重启

5. **自动更新**
   - 每 5 分钟向后端 `/api/client/check-update` 轮询版本（携带 `X-Client-Version` 头）
   - 发现新版本后自动下载 zip 安装包并校验
   - 备份本地 `config.yaml` 后解压安装，保留配置与 Token
   - 重启 systemd 服务完成升级，无需人工干预

6. **立即检查任务**
   - 轮询后端 `/api/client/tasks/pending` 拉取管理端下发的立即检查任务
   - 执行加固脚本并实时上报执行状态与结果（executing / completed / failed）

### 加固检查项

| 类别 | 检查项 |
|------|--------|
| 软件包安全 | dnf.conf gpgcheck、redhat.repo gpgcheck |
| 密码策略 | PASS_MAX_DAYS=30、PASS_MIN_DAYS=1、PASS_MIN_LEN=14、PASS_WARN_AGE=7、INACTIVE=30 |
| 用户环境 | TMOUT=180、根账户 GID |
| 计划任务 | cron 服务状态、cron/at 文件权限 |
| SSH 配置 | LogLevel、X11Forwarding=no、MaxAuthTries=4、PermitRootLogin=no、LoginGraceTime=60、ClientAliveInterval=60 等 |
| PAM 配置 | minlen=14、minclass=4、dcredit/ucredit/lcredit/ocredit=-1、password_remember=24 |
| 文件权限 | passwd=644、shadow=000、group=644、gshadow=000 |
| 加密策略 | crypto-policies（禁用弱算法） |
| 时间同步 | NTP 服务器配置 |

## 源码结构

```
linux-client/
├── main.go              # 客户端入口（注册、调度、信号处理）
├── api_client.go        # 后端 API 通信
├── token_manager.go     # Token 管理（JSON 文件读写、自动刷新）
├── script_executor.go   # 执行加固脚本并解析输出
├── checkupdate.go       # 版本更新检查（5 分钟轮询）
├── downloader.go        # 更新包下载（临时文件 + 校验）
├── updater.go           # 更新安装（备份配置、解压、重启服务）
├── task_fetch.go        # 立即检查任务拉取与执行
├── config.go            # YAML 配置加载
└── uninstall_server.sh  # 卸载脚本
```

## 编译

```bash
# 方式一：使用项目根目录构建脚本（交叉编译 + 打包 zip，推荐）
bash scripts/build-linux-client.sh 2.0.8

# 方式二：手动交叉编译（版本号通过 ldflags 注入）
cd linux-client
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=2.0.8" -o ../dist/linux-hardening-client .
```

> 版本号通过 `-ldflags "-X main.version=<版本>"` 编译时注入，自动更新依赖该版本与后端比对。

## 安装部署

安装请使用 `dist/` 目录下的安装包和 `install_client_interactive.sh` 安装脚本，详见 [dist/README.md](../dist/README.md)。

> 最新安装包（`linux-hardening-client_v<版本>.zip`）请从 GitHub Releases 下载：
> https://github.com/YanGLweI/SystemHardening/releases

### 快速安装

```bash
# 1. 上传安装包到目标服务器
scp dist/linux-hardening-client_XXX.zip root@目标服务器:/tmp/

# 2. 解压
ssh root@目标服务器 "cd /tmp && unzip linux-hardening-client_XXX.zip"

# 3. 运行安装脚本
ssh root@目标服务器 "bash /tmp/install_client_interactive.sh http://后端IP:8080"
```

### 卸载

```bash
# 在目标服务器上执行
bash uninstall_server.sh
```

或手动卸载：
```bash
systemctl stop linux-hardening-client
rm -rf /opt/linux-hardening-client
rm -f /etc/systemd/system/linux-hardening-client.service
systemctl daemon-reload
```

## 配置文件

安装后位于 `/opt/linux-hardening-client/config.yaml`：

```yaml
server_url: http://后端IP:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: 主机名（自动检测）
ip_address: IP地址（自动检测）
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

## Token 生命周期

```
客户端启动 → 检查本地 tokens.json
  ├── 无 Token → 请求临时 Token (5分钟) → 注册 → 获取正式 Token (14天)
  └── 有 Token → 检查是否即将过期
        ├── 未过期 → 直接使用
        └── 即将过期 → 自动刷新
```

## 维护命令

```bash
# 查看服务状态
systemctl status linux-hardening-client

# 重启服务（会立即触发一次检查）
systemctl restart linux-hardening-client

# 查看实时日志
journalctl -u linux-hardening-client -f

# 手动执行加固脚本
/opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

## 故障排查

### 1. 注册失败

```
注册失败：register failed
```

**解决**：确认后端服务正常运行，检查网络连通性：
```bash
curl http://后端IP:8080/api/health
```

### 2. 脚本执行失败

```
[ERROR] Script execution failed
```

**解决**：检查脚本权限：
```bash
chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

### 3. 数据上传失败（401）

```
[ERROR] Upload failed: Token 无效或已过期
```

**解决**：删除本地 Token 文件让客户端重新注册：
```bash
rm -f /opt/linux-hardening-client/data/tokens.json
systemctl restart linux-hardening-client
```

### 4. 加固未生效

**注意**：systemd 服务文件中不要添加 `ProtectSystem=strict` 等安全限制，否则脚本无法修改系统配置文件。

### 5. 自动更新失败

**解决**：检查日志中的更新记录，确认后端 `packages` 配置中 `server_url` 为客户端可访问的后端地址：
```bash
journalctl -u linux-hardening-client | grep -i "UPDATER\|UPDATE"
```

---

**版本**: 2.0.8  
**最后更新**: 2026-08-08
