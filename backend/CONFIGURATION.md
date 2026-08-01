# 配置文件说明

## 📋 概述

系统现在使用 `config.yml` 作为主要配置文件，替代了之前硬编码在代码中的配置。

## 🔧 配置结构

```yaml
# Server Configuration - 服务器配置
server:
  port: "8080"          # 服务端口

# Database Configuration - 数据库配置
database:
  host: "10.60.254.127"    # 数据库主机地址
  port: 3306              # 数据库端口
  user: "it"             # 数据库用户名
  password: "a*999999"   # 数据库密码
  dbname: "system_hardening"  # 数据库名称

# LDAP Configuration - LDAP 域控配置
ldap:
  server: "10.60.254.252:636"     # LDAP 服务器地址和端口
  base_dn: "dc=hot,dc=local"      # LDAP Base DN
  domain_suffix: "hot.local"      # 域名后缀
  use_tls: true                    # 是否使用 TLS
  insecure: true                   # 是否跳过证书验证（生产环境建议设为 false）
  cert_path: "./certificate/ca.crt"  # CA 证书路径
  admin_username: "CN=IT 服务账号，OU=服务账号，OU=HOT,DC=hot,DC=local"  # 管理员账号
  admin_password: "YourPasswordHere!"  # 管理员密码
  user_filter: "(sAMAccountName=%s)"  # 用户过滤器
  security_group_dn: "CN=IT 部，OU=IT 部，OU=HOT,DC=hot,DC=local"  # 安全组 DN

# JWT Configuration - JWT Token 配置
jwt:
  secret_key: "your-super-secret-key-min-32-chars"  # JWT 密钥（至少 32 字符）
  expiry_hour: 1                                    # Token 有效期（小时）
```

## 🚀 使用方法

### 1. 修改配置

直接编辑 `config.yml` 文件：

```bash
# 打开配置文件
vim config.yml
# 或
nano config.yml
```

### 2. 环境变量覆盖（可选）

系统支持通过环境变量覆盖 `config.yml` 中的配置，适合 CI/CD 场景：

```bash
# 设置环境变量覆盖配置
export SERVER_PORT=9090
export DB_HOST="new-host-address"
export LDAP_PASSWORD="updated-password"

# 启动应用，环境变量会覆盖 YAML 配置
./backend_app
```

优先级：环境变量 > config.yml > 默认值

### 3. 启动应用

```bash
cd backend
go run cmd/main.go
```

## 📝 注意事项

### LDAP 管理员账号

LDAP 管理员账号必须是完整的 DN 格式：
- ✅ 正确：`CN=IT 服务账号，OU=服务账号，OU=HOT,DC=hot,DC=local`
- ❌ 错误：`IT 服务账号` 或 `ylw@hot.local`

### LDAP 密码

请确保填写正确的 LDAP 管理员密码。如果密码包含特殊字符，需要用双引号包裹：

```yaml
admin_password: "My$ecureP@ssw0rd!"
```

### 证书验证

- `insecure: true` - 跳过证书验证（开发环境推荐）
- `insecure: false` - 严格验证证书（生产环境推荐）
- 如果启用证书验证，确保 `cert_path` 指向有效的 CA 证书

### JWT 密钥安全

- 生产环境必须修改默认的 `secret_key`
- 建议使用随机生成的长字符串
- 可以使用以下命令生成：
  ```bash
  openssl rand -base64 32
  ```

## 🔒 安全性建议

1. **不要将 config.yml 提交到版本控制**
   - 在 `.gitignore` 中添加：`config.yml`
   - 或使用敏感信息加密工具

2. **环境变量覆盖**
   - 敏感信息（如密码）建议使用环境变量
   - 生产环境避免在文件中明文存储密码

3. **CA 证书管理**
   - 将证书放在安全目录
   - 限制文件访问权限：`chmod 600 certificate/ca.crt`

## 📞 问题排查

如果启动时看到以下警告：

```
Warning: config.yml not found, falling back to defaults
```

说明 `config.yml` 文件不存在或路径不正确。确保在应用运行时目录中有该文件。

## 💡 示例

一个完整的、可用的 `config.yml` 示例请参考根目录下的 `config.yml` 文件。
