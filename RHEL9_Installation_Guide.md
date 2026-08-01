# RHEL 9 客户端安装与测试指南

## 📋 环境信息

**目标服务器**: `10.60.254.127`  
**SSH 账号**: `root`  
**SSH 密码**: `!Qw2!Qw2`  

**服务端地址**: `http://localhost:8080` (运行在 macOS)  
**数据库**: MySQL @ `test-it` 连接

---

## 🚀 快速安装步骤

### Step 1: 在本地准备文件

```bash
cd /Users/yeung/Projects/system_hardening
```

确认以下文件已存在：
- ✅ `/Users/yeung/Projects/system_hardening/bin/linux-hardening-client` - 客户端二进制文件
- ✅ `/Users/yeung/Projects/system_hardening/client/install.sh` - 安装脚本
- ✅ `/Users/yeung/Projects/未命名文件夹/RHEL/System_Check-1.2.sh` - 加固脚本

### Step 2: 传输文件到 RHEL 9 服务器

```bash
# 方式 A: scp 传输所有文件
scp -o StrictHostKeyChecking=no \
    /Users/yeung/Projects/system_hardening/bin/linux-hardening-client \
    root@10.60.254.127:/tmp/

scp -o StrictHostKeyChecking=no \
    /Users/yeung/Projects/system_hardening/client/install.sh \
    root@10.60.254.127:/tmp/

scp -o StrictHostKeyChecking=no \
    /Users/yeung/Projects/未命名文件夹/RHEL/System_Check-1.2.sh \
    root@10.60.254.127:/tmp/
```

### Step 3: 在 RHEL 9 上安装

通过 SSH 登录服务器：
```bash
ssh root@10.60.254.127
# 输入密码：!Qw2!Qw2
```

执行安装：
```bash
# 进入/tmp目录
cd /tmp

# 运行安装脚本（指定服务端 URL）
bash install.sh http://localhost:8080
```

或者手动安装：
```bash
INSTALL_DIR="/opt/linux-hardening-client"

# 创建目录
mkdir -p $INSTALL_DIR/{bin,scripts,data,logs}

# 移动文件
mv linux-hardening-client $INSTALL_DIR/bin/
chmod +x $INSTALL_DIR/bin/linux-hardening-client

mv System_Check-1.2.sh $INSTALL_DIR/scripts/
chmod +x $INSTALL_DIR/scripts/System_Check-1.2.sh

# 生成配置文件
cat > $INSTALL_DIR/config.yaml << 'EOF'
server_url: http://localhost:8080
local_db_path: /opt/linux-hardening-client/data/tokens.db
device_name: $(hostname)
ip_address: $(ip a | grep global | awk '{print $2}' | head -1)
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
EOF

echo "✅ Installation complete!"
```

### Step 4: 请求临时 Token 并注册

由于服务在 macOS 上，需要修改配置中的 server_url。

编辑配置文件：
```bash
# 查看服务器实际 IP
ip addr show | grep global

# 编辑配置
vim /opt/linux-hardening-client/config.yaml
```

将 `server_url` 改为 macOS 的局域网 IP（例如 `192.168.1.100`）：
```yaml
server_url: http://192.168.1.100:8080
```

获取临时 Token：
```bash
TEMP_TOKEN=$(curl -s -X POST "http://192.168.1.100:8080/api/client/request-temp-token" \
  -H "Content-Type: application/json" \
  -d '{
    "device_name": "'$(hostname)'",
    "ip_address": "'$(ip addr show | grep global | awk '{print $2}' | head -1)'"
  }' | jq -r '.temp_token')

echo "临时 Token: $TEMP_TOKEN"
```

注册客户端：
```bash
REGISTER_RESP=$(curl -s -X POST "http://192.168.1.100:8080/api/client/register" \
  -H "Content-Type: application/json" \
  -d '{
    "temp_token": "'"${TEMP_TOKEN}"'",
    "device_name": "'"$(hostname)"'",
    "ip_address": "'"$(ip addr show | grep global | awk '{print $2}' | head -1)"'",
    "os_version": "'"$(cat /etc/redhat-release)"'"
  }')

echo "$REGISTER_RESP" | jq .
```

保存响应中的 `short_token`、`refresh_token` 和 `expires_at`。

### Step 5: 保存 Token 到 SQLite

```bash
sqlite3 /opt/linux-hardening-client/data/tokens.db << EOF
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY,
    short_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
INSERT OR REPLACE INTO tokens (id, short_token, refresh_token, expires_at)
VALUES (1, 
    '$(echo $REGISTER_RESP | jq -r '.short_token')',
    '$(echo $REGISTER_RESP | jq -r '.refresh_token')',
    '$(echo $REGISTER_RESP | jq -r '.expires_at')'
);
EOF

echo "✅ Tokens saved successfully"
```

### Step 6: 启动客户端

```bash
# 前台测试运行（推荐先测试）
/opt/linux-hardening-client/bin/linux-hardening-client &

# 或后台运行
nohup /opt/linux-hardening-client/bin/linux-hardening-client > /opt/linux-hardening-client/logs/client.log 2>&1 &

# 查看进程
ps aux | grep linux-hardening-client
```

### Step 7: 验证功能

#### 测试 API 连通性
```bash
curl -X POST "http://192.168.1.100:8080/api/client/upload-data" \
  -H "Content-Type: application/json" \
  -H "X-Client-Token: $(sqlite3 /opt/linux-hardening-client/data/tokens.db 'SELECT short_token FROM tokens WHERE id=1;')" \
  -d '{
    "data": {
      "date": "'$(date '+%Y/%m/%d_%H:%M:%S')'",
      "hostname": "'$(hostname)'",
      "operasystem": "'$(cat /etc/redhat-release)'",
      "kernel": "'$(uname -r)'",
      "ip": "'$(ip a | grep global | awk '{print $2}' | head -1)'",
      "dnf_conf_gpgcheck": "gpgcheck=1",
      "redhat_repo_gpgcheck": "gpgcheck = 1",
      "pass_max_days": "30",
      "pass_min_days": "1",
      "pass_min_len": "14",
      "pass_warn_age": "7",
      "inactive": "30",
      "gid": "0",
      "tmout": "180",
      "crypto_policies": "DEFAULT:NO-SHA1",
      "ntp_server": "server ntp.example.com iburst"
    }
  }'
```

预期响应：
```json
{
  "status": "success",
  "record_id": 1,
  "message": "Data uploaded successfully"
}
```

---

## 🔍 查看数据库数据

使用 MCP dbx 工具：

```bash
# Mac Terminal 中执行
open /Users/yeung/Library/Application\ Support/Qoder/SharedClientCache/projects/-Users-yeung-Projects-system_hardening/mcps/dbx
```

或通过 Qoder IDE 的 MCP 面板查看 `test-it` 连接的 `system_hardening` 数据库。

---

## 📊 验证数据是否正确上传

### 方法 1: 通过 MCP dbx 查看

```sql
-- 查看客户端列表
SELECT * FROM clients;

-- 查看加固检查结果（最新的 10 条）
SELECT * FROM systemcheck ORDER BY id DESC LIMIT 10;

-- 查看 Token 信息（客户端本地存储）
-- 注意：这是客户端本地 SQLite 中的数据，不是服务器端
```

### 方法 2: 通过 MySQL 命令行

在 RHEL 9 服务器上：
```bash
mysql -u root -p!Qw2!Qw2!Qw2!Qw2 -h localhost system_hardening -e "
SELECT '=== Clients ===';
SELECT * FROM clients;

SELECT '';
SELECT '=== Latest Check Records ===';
SELECT id, hostname, ip, date, client_uuid 
FROM systemcheck 
ORDER BY id DESC 
LIMIT 5;
"
```

### 方法 3: 通过服务器日志

查看后端日志（在 macOS 终端）：
```bash
# 如果后台运行
tail -f /Users/yeung/Projects/system_hardening/backend/backend.log

# 或使用之前保存的日志文件
cat /Users/yeung/Projects/system_hardening/backend/backend_output.log | tail -50
```

期望看到类似输出：
```
[INFO] | 200 POST | data uploaded successfully | ...
```

---

## 🐛 故障排查

### 问题 1: 无法连接服务器

**症状**: `curl: Connection refused`

**解决**:
```bash
# 检查 macOS 防火墙
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate off

# 或添加端口例外
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --allow /path/to/backend/bin/server
```

### 问题 2: Token 过期

**症状**: `{"error":"Token 无效或已过期"}`

**解决**:
```bash
# 重新请求临时 Token
TEMP_TOKEN=$(curl -s -X POST "http://SERVER_IP:8080/api/client/request-temp-token" ...)

# 重新注册
curl -s -X POST "http://SERVER_IP:8080/api/client/register" ...

# 更新本地数据库
sqlite3 /opt/linux-hardening-client/data/tokens.db "...UPDATE..."
```

### 问题 3: 脚本执行失败

**症状**: `ERROR Script execution failed`

**解决**:
```bash
# 手动执行脚本测试
bash /opt/linux-hardening-client/scripts/System_Check-1.2.sh 2>&1 | tail -100

# 检查权限
ls -l /opt/linux-hardening-client/scripts/System_Check-1.2.sh
chmod +x /opt/linux-hardening-client/scripts/System_Check-1.2.sh

# 查看错误日志
cat /opt/linux-hardening-client/logs/client.log | grep ERROR
```

---

## 📝 完整测试流程清单

- [ ] 1. 复制二进制文件和脚本到 RHEL 9
- [ ] 2. 安装到 `/opt/linux-hardening-client/`
- [ ] 3. 配置正确的 server_url
- [ ] 4. 请求临时 Token 并注册
- [ ] 5. 保存 Token 到 SQLite
- [ ] 6. 启动客户端
- [ ] 7. 手动触发一次检查上传
- [ ] 8. 检查数据库是否有新记录
- [ ] 9. 验证数据完整性
- [ ] 10. 查看后端日志确认成功接收

---

## 🎯 预期结果

成功的测试应该显示：

1. **数据库中新增客户端记录**:
   ```sql
   SELECT * FROM clients WHERE device_name = 'RHEL9-Server';
   ```

2. **数据库中新增加固检查记录**:
   ```sql
   SELECT * FROM systemcheck WHERE hostname = 'RHEL9-Server' ORDER BY id DESC LIMIT 1;
   ```

3. **后端日志包含上传成功消息**:
   ```
   [INFO] upload success | record_id: 1
   ```

4. **客户端日志正常**:
   ```
   [CHECK] Starting daily security check...
   [API] Uploading data for device: RHEL9-Server
   [CHECK] ✅ Daily check completed successfully
   ```

---

## 📞 需要帮助？

如有问题，请提供以下信息：

1. **客户端日志**:
   ```bash
   cat /opt/linux-hardening-client/logs/client.log
   ```

2. **数据库连接状态**:
   ```bash
   curl http://localhost:8080/api/health
   ```

3. **表结构信息**:
   ```bash
   mysql -u root -p!Qw2!Qw2!Qw2!Qw2 -e "DESC systemcheck;" system_hardening
   ```

---

**祝安装顺利！🚀**
