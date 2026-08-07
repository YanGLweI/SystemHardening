# System Hardening Platform Backend

基于 Go + Gin + GORM 的系统加固管理平台后端 API，提供 Linux / Windows 客户端注册、Token 认证、检查数据接收、标准配置管理与合规比对、安装包管理与自动更新分发、看板统计、邮件通知与每日合规报告推送，支持 LDAP 域控登录。

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
│   ├── user.go          # 用户
│   ├── client.go        # 客户端（uuid/token/区域/版本）
│   ├── region.go        # 区域
│   ├── linux_check.go   # Linux 加固检查数据
│   ├── linux_standard.go# Linux 标准配置
│   ├── linux_group.go   # Linux 标准分组
│   ├── windows_check.go # Windows 加固检查数据
│   ├── compliance.go    # 合规比对结果
│   ├── mail_config.go   # SMTP 邮件配置
│   └── report_schedule.go # 报告推送计划
├── routes/              # 路由注册
├── middleware/          # JWT 认证 / 请求日志
├── handlers/            # 认证（LDAP）、客户端业务
├── controllers/         # 检查/标准/区域/客户端/看板/邮件控制器
├── services/            # LDAP / 邮件（SMTP）/ 定时调度服务
├── packages/            # 客户端安装包存储（linux / windows，不入库）
├── scripts/             # 字段初始化工具（go run）
├── certificate/         # LDAPS CA 证书（ca.crt）
├── config.yml           # 主配置文件
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
cp .env.example .env        # 环境变量（数据库连接等，可留空使用 config.yml）
# 编辑 config.yml 修改 database / ldap / jwt 配置

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

主配置 `config.yml`：

```yaml
server:
  port: "8080"

database:
  host: "127.0.0.1"
  port: 3306
  user: "it"
  password: "your-password"
  dbname: "system_hardening"

ldap:
  server: "ldaps://dc.example.local:636"   # 域控 LDAPS 地址
  base_dn: "dc=example,dc=local"
  user_filter: "(sAMAccountName=%s)"

jwt:
  secret_key: "your-secret-key"
  expiry_hour: 8

packages:
  linux_package_dir: "./packages/linux"      # Linux 安装包存储目录
  windows_package_dir: "./packages/windows"  # Windows 安装包存储目录
  server_url: "http://后端IP:8080"           # 客户端下载更新包时使用的地址
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

## 数据模型

| 表 | 说明 |
|----|------|
| `users` | 用户（LDAP 认证） |
| `clients` | 客户端注册信息（UUID、Token、区域、最后心跳） |
| `regions` | 区域 |
| `systemcheck` | Linux 加固检查数据（GORM 动态字段） |
| `linux_standards` | Linux 标准配置（含分组） |
| `systemcheck_windows` | Windows 加固检查数据 |
| `windows_standards` | Windows 标准配置 |
| `package_meta` | 安装包元信息（版本/文件名/哈希） |
| `mail_configs` | SMTP 邮件配置 |
| `report_schedules` | 报告推送计划 |

## 开发说明

- Gin 框架处理 HTTP 请求，路由集中在 `routes/router.go`
- GORM AutoMigrate 自动建表，新字段直接加在模型上即可生效
- JWT 中间件保护管理接口，客户端接口使用 Token 认证
- 安装包上传后写入 `packages/` 目录并同步 `package_meta` 版本记录，客户端通过 `check-update` 轮询实现自动更新
- `services/scheduler.go` 负责报告计划调度，按配置时间生成合规报告并通过 SMTP 推送
- CORS 已配置，允许跨域请求
- 日志中间件记录请求信息
