# Linux 加固客户端安装包 - 安装指南

## 前置条件

- RHEL 9 或兼容发行版
- 依赖：`curl`、`jq`（如未安装：`yum install curl jq`）
- 后端服务已启动且网络可达

## 安装包内容

| 文件 | 说明 |
|------|------|
| `linux-hardening-client` | 客户端二进制文件（Linux amd64） |
| `System_Check-1.2.sh` | 安全加固检查脚本 |
| `install_client_interactive.sh` | 交互式安装脚本 |
| `config.example.yaml` | 配置文件示例 |
| `README.md` | 本文档 |

## 安装步骤

### 方式一：交互式安装（推荐）

```bash
# 1. 上传安装包到目标服务器
scp linux-hardening-client_XXX.zip root@目标服务器:/tmp/

# 2. 解压
cd /tmp && unzip linux-hardening-client_XXX.zip

# 3. 运行安装脚本（交互式输入后端地址）
bash install_client_interactive.sh

# 或直接指定后端地址（无需交互）
bash install_client_interactive.sh http://后端IP:8080
```

安装脚本会自动完成以下操作：
- 检测主机名和 IP 地址
- 创建安装目录 `/opt/linux-hardening-client/`
- 部署二进制文件和加固脚本
- 生成 systemd 服务文件并启动服务
- 生成配置文件 `config.yaml`

### 方式二：手动安装

```bash
# 1. 创建目录
mkdir -p /opt/linux-hardening-client/{bin,scripts,data,logs}

# 2. 部署文件
cp linux-hardening-client /opt/linux-hardening-client/bin/
chmod +x /opt/linux-hardening-client/bin/linux-hardening-client

cp System_Check-1.2.sh /opt/linux-hardening-client/scripts/
chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh

# 3. 生成配置文件（修改为实际值）
cat > /opt/linux-hardening-client/config.yaml << EOF
server_url: http://后端IP:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: $(hostname)
ip_address: $(hostname -I | awk '{print $1}')
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
EOF

# 4. 安装 systemd 服务（安装脚本会自动生成，手动安装时需自行创建）
cat > /etc/systemd/system/linux-hardening-client.service << 'EOF'
[Unit]
Description=Linux Hardening Client
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=/opt/linux-hardening-client/bin/linux-hardening-client
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
User=root
Group=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable linux-hardening-client
systemctl start linux-hardening-client
```

## 验证安装

```bash
# 查看服务状态
systemctl status linux-hardening-client

# 查看实时日志
journalctl -u linux-hardening-client -f

# 确认数据库中已有数据（在后端服务器执行）
mysql -u 用户名 -p 数据库名 -e "SELECT hostname, login_grace_time FROM systemcheck;"
```

正常日志输出示例：
```
=== Linux Hardening Client v1.0.0 ===
Server URL: http://后端IP:8080
Device: 主机名 (IP地址)
✅ 客户端注册成功! UUID: xxx
🚀 Performing initial security check...
[CHECK] Starting daily security check...
📊 UploadData - DnfConfGpgcheck: 1, RedhatRepoGpgcheck: 1
[API] ✅ Data uploaded successfully, status code: 200
[CHECK] ✅ Daily check completed successfully
```

## 配置文件说明

位置：`/opt/linux-hardening-client/config.yaml`

```yaml
server_url: http://后端IP:8080          # 后端服务地址
local_db_path: /opt/linux-hardening-client/data/tokens.json  # Token 存储路径
device_name: 主机名                      # 设备标识
ip_address: IP地址                       # 设备 IP
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh  # 加固脚本路径
```

## 卸载

```bash
systemctl stop linux-hardening-client
systemctl disable linux-hardening-client
rm -rf /opt/linux-hardening-client
rm -f /etc/systemd/system/linux-hardening-client.service
systemctl daemon-reload
```

## 故障排查

### 1. 服务启动失败

```bash
# 查看详细错误
journalctl -u linux-hardening-client -n 50 --no-pager

# 手动运行测试
/opt/linux-hardening-client/bin/linux-hardening-client
```

### 2. 无法连接后端

```bash
# 测试网络连通性
curl -s http://后端IP:8080/api/client/request-temp-token \
  -H "Content-Type: application/json" \
  -d '{"device_name":"test","ip_address":"1.2.3.4"}'
```

### 3. 加固未生效

- 确认 systemd 服务文件中**没有** `ProtectSystem=strict` 等安全限制
- 手动执行脚本验证：`/opt/linux-hardening-client/scripts/System_Check-1.2.sh`
- 检查脚本权限：`ls -la /opt/linux-hardening-client/scripts/`

### 4. Token 过期

```bash
# 删除本地 Token，重启后自动重新注册
rm -f /opt/linux-hardening-client/data/tokens.json
systemctl restart linux-hardening-client
```

## 注意事项

1. 客户端以 **root** 用户运行（需要修改系统配置文件）
2. 安装后客户端会**立即执行一次**加固检查并上传数据
3. 之后每 **24 小时**自动执行一次
4. 同一设备多次上传不会产生重复记录（后端按 client_uuid 更新）
