# System Hardening Platform - 系统加固管理平台

基于 Vue2 + Go 构建的前后端分离系统加固管理平台，支持 RHEL 9 Linux 服务器的自动化安全加固与合规检查。

## 技术栈

### 前端
- Vue 2.6.14
- Element UI 2.15.13
- Vue Router 3.5.4
- Vuex 3.6.2
- Axios 1.6.2
- Sass

### 后端
- Go 1.25+
- Gin 1.9.1
- GORM 1.25.5
- MariaDB/MySQL

### 客户端
- Go 1.25+（纯 Go 实现，无 CGO 依赖）
- systemd 服务管理
- Shell 脚本（安全加固检查）

## 项目结构

```
system_hardening/
├── backend/             # 后端服务
│   ├── cmd/            # 主程序入口
│   ├── configs/        # 配置文件（YAML）
│   ├── database/       # 数据库连接（GORM AutoMigrate）
│   ├── models/         # 数据模型
│   ├── routes/         # 路由配置
│   ├── middleware/     # 中间件（JWT 认证）
│   ├── handlers/       # 业务处理
│   ├── controllers/    # 控制器
│   ├── services/       # LDAP 服务
│   └── config.yml      # 后端配置文件
│
├── client/              # Linux 加固客户端源码
│   ├── main.go         # 客户端入口
│   ├── api_client.go   # API 通信
│   ├── token_manager.go# Token 管理（JSON 文件存储）
│   ├── script_executor.go # 脚本执行与解析
│   ├── config.go       # 配置加载
│   └── uninstall_server.sh # 卸载脚本
│
├── dist/                # 客户端发布包
│   ├── linux-hardening-client        # 客户端二进制（Linux amd64）
│   ├── System_Check-1.2.sh           # 安全加固检查脚本
│   ├── install_client_interactive.sh # 交互式安装脚本（含 systemd 服务生成）
│   ├── config.example.yaml           # 配置示例
│   └── README.md                     # 安装指南
│
└── frontend/           # 前端应用
    ├── public/         # 静态资源
    └── src/
        ├── api/        # API 请求
        ├── assets/     # 资源文件
        ├── components/ # 公共组件
        ├── router/     # 路由配置
        ├── store/      # Vuex 状态管理
        ├── views/      # 页面组件
        ├── App.vue
        └── main.js
```

## 系统架构

```
┌────────────────┐       ┌────────────────┐       ┌────────────────┐
│  RHEL 9 服务器  │       │   Go 后端服务    │       │   Vue 前端      │
│                │       │                │       │                │
│ linux-hardening│ HTTP  │ • 客户端注册     │ HTTP  │ • 登录（LDAP）  │
│ -client        │──────→│ • Token 管理    │←──────│ • 合规检查展示  │
│                │ API   │ • 数据接收/存储  │ API   │ • 标准配置管理  │
│ System_Check   │       │ • LDAP 认证     │       │ • 客户端管理    │
│ -1.2.sh        │       │                │       │                │
└────────────────┘       └───────┬────────┘       └────────────────┘
                                 │
                          ┌──────┴──────┐
                          │  MariaDB    │
                          │ system_     │
                          │ hardening   │
                          └─────────────┘
```

## 快速开始

### 环境要求
- Node.js >= 16.x
- Go >= 1.21
- MariaDB/MySQL 5.7+

### 1. 后端启动

```bash
cd backend

# 安装依赖
go mod tidy

# 编辑 config.yml 修改数据库和 LDAP 配置

# 构建
go build -o bin/server cmd/main.go

# 运行
./bin/server
```

后端服务将在 http://localhost:8080 启动

### 2. 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

前端应用将在 http://localhost:8081 启动

### 3. 客户端构建

```bash
cd client

# 交叉编译 Linux amd64 版本
GOOS=linux GOARCH=amd64 go build -o ../dist/linux-hardening-client .
```

## API 接口

### 认证接口
- `POST /api/auth/login` - LDAP 登录
- `GET /api/profile` - 获取用户信息

### 客户端接口（无需认证）
- `POST /api/client/request-temp-token` - 请求临时安装 Token（5 分钟有效）
- `POST /api/client/register` - 客户端注册
- `POST /api/client/refresh-token` - 刷新 Token
- `POST /api/client/upload-data` - 上传系统检查数据

### 管理接口（需 JWT 认证）
- `GET /api/health` - 健康检查
- `GET /api/linux-checks` - 获取加固检查数据
- `POST /api/linux-standards` - 创建标准配置
- `GET /api/linux-standards` - 获取标准配置列表
- `PUT /api/linux-standards/:id` - 更新标准配置
- `DELETE /api/linux-standards/:id` - 删除标准配置

### 跨域配置

后端已配置 CORS，允许所有来源的请求。生产环境建议限制来源。

前端已配置代理，开发环境下 `/api` 请求会转发到后端服务 `http://localhost:8080`

## 数据库配置

后端配置文件 `backend/config.yml`：
```yaml
server:
  port: "8080"

database:
  host: "10.60.254.127"
  port: 3306
  user: "it"
  password: "your-password"
  dbname: "system_hardening"

ldap:
  server: "ldaps://your-dc:636"
  base_dn: "dc=example,dc=local"
  ...

jwt:
  secret_key: "your-secret-key"
  expiry_hour: 8
```

数据库使用 GORM AutoMigrate 自动建表和同步字段结构。

## Linux 加固客户端

### 概述

客户端部署在 RHEL 9 目标服务器上，以 systemd 服务方式运行。启动后自动执行安全加固检查脚本，并将结果上报到后端服务器。

### 工作流程

```
客户端启动 → 加载/注册 Token → 执行 System_Check-1.2.sh → 解析输出 → 上传数据到后端
     ↓
每 24 小时自动重复检查
```

### 安装包内容（dist/）

| 文件 | 说明 |
|------|------|
| `linux-hardening-client` | 客户端二进制（Linux amd64） |
| `System_Check-1.2.sh` | 安全加固检查脚本 |
| `install_client_interactive.sh` | 交互式安装脚本（含 systemd 服务生成） |
| `uninstall.sh` | 卸载脚本（自动检测安装路径） |
| `config.example.yaml` | 配置文件示例 |
| `README.md` | 安装指南 |

### 安装方式

```bash
# 1. 将安装包上传到目标服务器
scp linux-hardening-client_XXX.zip root@target-server:/tmp/

# 2. 解压
cd /tmp && unzip linux-hardening-client_XXX.zip

# 3. 运行交互式安装（自动检测主机名和 IP）
bash install_client_interactive.sh http://BACKEND_IP:8080

# 4. 验证
systemctl status linux-hardening-client
journalctl -u linux-hardening-client -f
```

### 客户端配置

安装后配置文件位于 `/opt/linux-hardening-client/config.yaml`：

```yaml
server_url: http://BACKEND_IP:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: auto-detected-hostname
ip_address: auto-detected-ip
script_path: /opt/linux-hardening-client/scripts/System_Check-1.2.sh
```

### Token 认证机制

```
安装 → 请求临时 Token (5分钟) → 注册 → 获取 Short Token (14天) + Refresh Token
                                                    ↓
                                         到期前 24 小时自动刷新
```

- Token 存储在本地 JSON 文件（`tokens.json`），权限 0600
- 客户端重启后自动加载已有 Token，无需重新注册

### 加固检查项

脚本 `System_Check-1.2.sh` 会自动检查并加固以下配置：

| 类别 | 检查项 |
|------|--------|
| 软件包安全 | dnf.conf gpgcheck、redhat.repo gpgcheck |
| 密码策略 | PASS_MAX_DAYS、PASS_MIN_DAYS、PASS_MIN_LEN、PASS_WARN_AGE、INACTIVE |
| 用户环境 | TMOUT、根账户 GID |
| 计划任务 | cron 服务状态、cron/at 文件权限 |
| SSH 配置 | LogLevel、X11Forwarding、MaxAuthTries、PermitRootLogin、LoginGraceTime、ClientAliveInterval 等 |
| PAM 配置 | minlen、minclass、dcredit、ucredit、lcredit、ocredit、password_remember |
| 文件权限 | passwd、shadow、group、gshadow 等 |
| 加密策略 | crypto-policies |
| 时间同步 | NTP 服务器配置 |

### 客户端维护命令

```bash
# 查看服务状态
systemctl status linux-hardening-client

# 重启服务（会立即触发一次检查）
systemctl restart linux-hardening-client

# 查看日志
journalctl -u linux-hardening-client -f

# 手动执行检查脚本
/opt/linux-hardening-client/scripts/System_Check-1.2.sh

# 卸载
systemctl stop linux-hardening-client
rm -rf /opt/linux-hardening-client
rm -f /etc/systemd/system/linux-hardening-client.service
systemctl daemon-reload
```

## 开发规范

### 代码风格
- Go: 遵循官方 Go Code Review Comments
- JavaScript/Vue: ESLint + Prettier 配置
- CSS: BEM 命名规范

### 提交规范
建议使用 Git Commit Linter 确保提交信息规范

## 注意事项

⚠️ **重要**：
1. 确保客户端能访问后端服务地址
2. 生产环境请勿使用默认密码
3. systemd 服务文件不要添加 `ProtectSystem=strict` 等安全限制，否则加固脚本无法修改系统配置文件
4. 客户端以 root 用户运行（需要修改系统配置）

## 后续开发计划

- [x] 用户认证（LDAP 域控集成）
- [x] 客户端自动注册与 Token 管理
- [x] 系统加固检查与数据上报
- [x] 合规性标准配置管理
- [ ] 权限管理系统
- [ ] 批量部署工具
- [ ] 合规性报告导出
- [ ] 审计日志

## 故障排除

### 后端启动失败
1. 检查 Go 版本是否 >= 1.21
2. 验证数据库连接配置
3. 检查端口 8080 是否被占用

### 前端启动失败
1. 清理 node_modules 重新安装：`rm -rf node_modules && npm install`
2. 检查 Node.js 版本 >= 16
3. 验证端口 8081 是否可用

### 无法连接后端 API
1. 确认后端服务已启动 (http://localhost:8080)
2. 检查前端代理配置是否正确
3. 查看浏览器控制台网络请求错误

## License

MIT License
