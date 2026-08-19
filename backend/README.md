# System Hardening Platform Backend

基于 Go + Gin + GORM 的系统加固管理平台后端 API，提供 Linux / Windows 客户端注册、Token 认证、检查数据接收、标准配置管理与合规比对、安装包管理与自动更新分发、立即检查任务调度、看板统计、邮件通知与每日合规报告推送，支持 LDAP 域控登录。

## 技术栈

- Go 1.25+
- Gin 1.9.1
- GORM 1.25.5（AutoMigrate 自动建表）
- MariaDB/MySQL 5.7+
- LDAP（域控认证）+ JWT（`golang-jwt`）

## 项目结构

```
backend/
├── cmd/                 # 主程序入口
├── configs/             # 配置加载（config.yml）
├── database/            # 数据库连接（GORM AutoMigrate）
├── models/              # 数据模型
│   ├── client.go        # 客户端（uuid/token/区域/版本）
│   ├── region.go        # 区域
│   ├── linux_check.go   # Linux 加固检查数据
│   ├── linux_standard.go# Linux 标准配置
│   ├── linux_group.go   # Linux 标准分组
│   ├── windows_check.go # Windows 加固检查数据
│   ├── compliance.go    # 合规比对结果
│   ├── check_task.go    # 立即检查任务（状态机：pending→executing→completed/failed）
│   ├── standard_exemption.go # 标准字段例外（字段 × 客户端）
│   ├── types.go         # 通用类型（JSONMap 等）
│   ├── mail_config.go   # SMTP 邮件配置
│   └── report_schedule.go # 报告推送计划
├── routes/              # 路由注册
├── middleware/          # JWT 认证 / 请求日志
├── handlers/            # 认证（LDAP）、客户端业务
├── controllers/         # 检查/标准/区域/客户端/看板/邮件/检查任务控制器
├── services/            # LDAP / 邮件（SMTP）/ 定时调度服务
├── packages/            # 客户端安装包存储（linux / windows，不入库）
├── scripts/             # 字段初始化工具（go run）
├── certificate/         # LDAPS CA 证书（ca.crt）
├── config.yml.example   # 配置示例文件（必须复制为 config.yml）
├── .env                 # 环境变量（本地开发，不入库）
└── .env.example         # 环境变量示例
```

## 环境要求

- Go 1.21+
- MariaDB/MySQL 5.7+

## 安装与运行

```bash
# 1. 安装依赖
cd backend
go mod tidy

# 2. 配置
# 复制配置示例并编辑实际参数
cp config.yml.example config.yml
vim config.yml  # 修改数据库/ldap/jwt等配置
# 或者使用环境变量覆盖（见下方配置说明）

# 3. 开发运行
go run cmd/main.go

# 4. 生产构建
go build -o bin/server cmd/main.go
./bin/server

# 或使用启动脚本
./start_backend.sh
```

启动后服务监听 http://localhost:8080，数据库表结构由 GORM AutoMigrate 自动创建/同步。

## 配置说明

### 配置文件

复制 `config.yml.example` 为 `config.yml` 并编辑实际参数：

```bash
cp config.yml.example config.yml
```

**重要**: `config.yml`包含敏感信息（数据库密码、LDAP 凭证等），已加入 `.gitignore`，不应提交到代码仓库。

### 配置结构

主配置文件 `config.yml` 的结构如下：

```yaml
server:
  port: "8080"                  # 服务端口

database:
  host: "localhost"              # 数据库主机
  port: 3306                     # 数据库端口
  user: "your_db_user"
  password: "your_db_password"
  dbname: "system_hardening"

ldap:
  server: "ldaps://dc.example.local:636"   # 域控服务器地址
  base_dn: "dc=example,dc=local"           # 基础 DN
  domain_suffix: "example.local"           # 域名后缀
  use_tls: true                          # 是否使用 TLS
  insecure: true                         # 测试环境可设为 true
  cert_path: "./certificate/ca.crt"      # CA 证书路径
  admin_username: "admin@domain.local"
  admin_password: "your_admin_password"
  security_group_dn: "CN=Group,OU=...,DC=..."

jwt:
  secret_key: "your-secret-key-min-32-chars"  # JWT 签名密钥（至少 32 字符）
  expiry_hour: 8                              # Token 有效期（小时）

packages:
  linux_package_dir: "./packages/linux"       # Linux 安装包目录
  windows_package_dir: "./packages/windows"   # Windows 安装包目录
  server_url: "http://your-server-ip:8080"    # 客户端访问的后端地址
```

### 环境变量覆盖

所有配置项支持通过环境变量覆盖（优先级高于 YAML 文件）：

```bash
# 在 .env 文件或 export 命令中设置
export SERVER_PORT=8080
export DB_HOST=10.60.254.127
export DB_USER=it
export DB_PASSWORD=your_password
export LDAP_SERVER=ldaps://dc.example.local:636
export LDAP_ADMIN_USERNAME=admin@domain.local
export LDAP_ADMIN_PASSWORD=your_password
export JWT_SECRET=your-super-secret-key-min-32-changes
export SERVER_URL=http://your-server-ip:8080
```

完整环境变量名参见 `backend/.env.example`
```

## API 接口

### 认证接口
- `POST /api/auth/login` - LDAP 登录
- `GET /api/profile` - 获取用户信息（JWT）

### 客户端接口（无需认证）
- `POST /api/client/request-temp-token` - 请求临时安装 Token（5 分钟有效）
- `POST /api/client/register` - 客户端注册（返回 Short Token + Refresh Token）
- `POST /api/client/refresh-token` - 刷新 Token
- `POST /api/client/heartbeat` - 客户端心跳
- `POST /api/client/upload-data` - 上传 Linux 系统检查数据（按 client_uuid 去重更新）
- `POST /api/client/upload-data-windows` - 上传 Windows 加固检查数据
- `GET /api/client/check-update` - 检查新版本（客户端每 5 分钟轮询，携带 `X-Client-Version` 头）
- `GET /api/client/tasks/pending` - 客户端拉取待执行的立即检查任务
- `PUT /api/client/tasks/:id/result` - 客户端上报任务执行结果

### 安装包接口
- `GET /api/packages/:type/download` - 下载安装包（公开，客户端自动更新使用）
- `POST /api/packages/upload` - 上传安装包（需 JWT，上传后自动同步版本号）
- `GET /api/packages/:type/info` - 获取安装包版本信息（需 JWT）

### 管理接口（需 JWT 认证）
- `GET /api/health` - 健康检查
- `GET /api/test` - 连通性测试
- `GET /api/linux-checks` - Linux 加固检查列表
- `GET /api/linux-checks/:id` - Linux 加固检查详情
- `POST /api/linux-standards` - 创建 Linux 标准配置
- `GET /api/linux-standards` - Linux 标准配置列表
- `GET /api/linux-standards/fields` - 可用 Linux 字段列表
- `PUT /api/linux-standards/:id` - 更新 Linux 标准配置
- `DELETE /api/linux-standards/:id` - 删除 Linux 标准配置
- `GET /api/windows-checks` - Windows 加固检查列表
- `GET /api/windows-checks/:id` - Windows 加固检查详情
- `POST /api/windows-standards` - 创建 Windows 标准配置
- `GET /api/windows-standards` - Windows 标准配置列表
- `GET /api/windows-standards/fields` - 可用 Windows 字段列表
- `PUT /api/windows-standards/:id` - 更新 Windows 标准配置
- `DELETE /api/windows-standards/:id` - 删除 Windows 标准配置
- `POST /api/regions` - 创建区域
- `GET /api/regions` - 区域列表
- `PUT /api/regions/:id/clients` - 配置区域关联客户端
- `DELETE /api/regions/:id` - 删除区域
- `GET /api/clients` - 客户端列表
- `DELETE /api/clients/:id` - 删除客户端（硬删除，联动清理检查数据）
- `GET /api/dashboard/stats` - 看板统计（在线状态、区域分布、合规率）
- `GET /api/mail-config` - 获取 SMTP 邮件配置
- `PUT /api/mail-config` - 保存邮件配置
- `POST /api/mail/test` - 发送测试邮件
- `GET /api/report-schedules` - 报告计划列表
- `POST /api/report-schedules` - 创建报告计划
- `PUT /api/report-schedules/:id` - 更新报告计划
- `DELETE /api/report-schedules/:id` - 删除报告计划
- `POST /api/report-schedules/:id/send` - 立即发送报告
- `POST /api/tasks/trigger` - 触发指定客户端立即检查（同一客户端并发限制 1）
- `GET /api/tasks/:id` - 查询任务状态
- `GET /api/tasks/client/:client_uuid` - 获取客户端最新任务
- `DELETE /api/tasks/:id` - 删除任务（卡死任务重试）

## 数据模型

| 表 | 说明 |
|----|------|
| `clients` | 客户端注册信息（UUID、Token、区域、最后心跳） |
| `regions` | 区域 |
| `systemcheck` | Linux 加固检查数据（GORM 动态字段） |
| `linux_standards` | Linux 标准配置（含分组） |
| `systemcheck_windows` | Windows 加固检查数据 |
| `windows_standards` | Windows 标准配置 |
| `package_meta` | 安装包元信息（版本/文件名/哈希） |
| `check_tasks` | 立即检查任务（状态机、重试计数） |
| `standard_exemptions` | 标准字段例外（字段 × 客户端，软删除） |
| `mail_configs` | SMTP 邮件配置 |
| `report_schedules` | 报告推送计划 |

## 开发说明

- Gin 框架处理 HTTP 请求，路由集中在 `routes/router.go`
- 登录采用纯 LDAP 域控认证（不存储本地用户表），签发 JWT 供管理接口使用
- GORM AutoMigrate 自动建表，新字段直接加在模型上即可生效
- JWT 中间件保护管理接口，客户端接口使用 Token 认证
- 安装包上传后写入 `packages/` 目录并同步 `package_meta` 版本记录，客户端通过 `check-update` 轮询实现自动更新
- `services/scheduler.go` 负责报告计划调度，按配置时间生成合规报告并通过 SMTP 推送
- CORS 已配置，允许跨域请求
- 日志中间件记录请求信息
