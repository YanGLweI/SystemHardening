# System Hardening Platform - 系统加固管理平台

基于 Vue2 + Go + Rust 的前后端分离系统加固管理平台，支持 **RHEL 9 Linux 服务器**与 **Windows 终端（AD 域环境）** 的自动化安全加固与合规检查，提供统一看板、标准配置管理、客户端自动更新与每日合规报告邮件推送。

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
- GORM 1.25.5（AutoMigrate 自动建表）
- MariaDB/MySQL 5.7+
- LDAP（域控认证）+ JWT
- SMTP 邮件服务（合规报告定时推送）

### 客户端
- **Linux 客户端**（v1.4.0）：Go 1.22+（纯 Go 实现，无 CGO），systemd 服务管理，Shell 加固脚本，支持自动更新
- **Windows 客户端**（v2.0.6）：Rust（`x86_64-pc-windows-gnu`），Windows 服务 + WMI + 注册表 + GPO 文件解析，NSIS 安装包，支持自动更新

## 项目结构

```
system_hardening/
├── backend/              # 后端服务（Go + Gin + GORM）
│   ├── cmd/             # 主程序入口
│   ├── configs/         # 配置加载
│   ├── database/        # 数据库连接（AutoMigrate）
│   ├── models/          # 数据模型（检查、标准、客户端、区域、邮件、报告计划）
│   ├── routes/          # 路由配置
│   ├── middleware/      # JWT 认证 / 日志中间件
│   ├── handlers/        # 认证、客户端业务处理
│   ├── controllers/     # Linux/Windows 检查与标准、客户端、区域、看板、邮件控制器
│   ├── services/        # LDAP / 邮件 / 定时调度服务
│   ├── scripts/         # 字段初始化工具
│   ├── packages/        # 客户端安装包存储目录（linux / windows）
│   ├── certificate/     # LDAPS 证书（ca.crt）
│   └── config.yml       # 后端配置文件
│
├── linux-client/              # Linux 加固客户端源码（Go）
│   ├── main.go          # 入口（注册、调度、信号处理）
│   ├── api_client.go    # 后端 API 通信
│   ├── token_manager.go # Token 管理（JSON 文件存储、自动刷新）
│   ├── script_executor.go # 加固脚本执行与解析
│   ├── checkupdate.go   # 版本更新检查
│   ├── downloader.go    # 更新包下载
│   ├── updater.go       # 更新安装（备份配置、重启服务）
│   ├── config.go        # YAML 配置加载
│   └── uninstall_server.sh
│
├── windows-client/       # Windows 加固客户端源码（Rust）
│   ├── src/             # 服务 / 采集器 / API / Token / 配置 / 自动更新
│   ├── installer/       # NSIS 安装包脚本
│   └── config.example.yaml
│
├── frontend/             # 前端应用（Vue2 + Element UI）
│   └── src/
│       ├── api/         # API 请求（检查/标准/客户端/区域/看板/邮件/安装包）
│       ├── views/       # 登录、加固检查、标准配置、客户端/区域管理、邮件通知、看板
│       ├── router/      # 路由配置（含登录守卫）
│       └── store/       # Vuex 状态管理
│
├── scripts/              # 客户端构建脚本（build-linux-client.sh / build-windows-client.sh）
├── dist/                 # Linux 客户端发布目录（安装包素材 + 安装脚本 + 加固脚本）
└── dist_win/             # Windows 客户端发布目录（NSIS 安装包 exe）
```

## 系统架构

```
┌────────────────────┐      ┌────────────────────┐
│  RHEL 9 服务器      │      │  Windows 终端（AD域）│
│                    │      │                    │
│ linux-hardening-   │      │ windows_hardening_ │
│ client (Go/systemd)│      │ client (Rust/服务)  │
│ System_Check-1.2.sh│      │ WMI+注册表+GPO解析   │
└─────────┬──────────┘      └─────────┬──────────┘
          │ HTTP API                  │ HTTP API
          │ Token 认证 / 数据上报      │ Token 认证 / 数据上报
          │ 心跳 / 自动更新检查        │ 心跳 / 自动更新检查
          ▼                            ▼
┌─────────────────────────────────────────────┐
│              Go 后端服务（Gin）               │
│  • 客户端注册 / Token 管理 / 心跳             │
│  • Linux/Windows 检查数据接收与存储           │
│  • 标准配置管理（合规比对）                   │
│  • 安装包管理 / 版本检查 / 自动更新分发        │
│  • 看板统计 / 每日合规报告邮件推送             │
│  • LDAP 认证 + JWT                           │
└────────────────────┬────────────────────────┘
          │                                  │
          ▼                                  ▼
┌──────────────────┐              ┌──────────────────┐
│   Vue2 前端       │              │   MariaDB         │
│  • LDAP 登录      │◄────────────►│  system_hardening │
│  • 合规检查展示   │              └──────────────────┘
│  • 标准配置管理   │
│  • 客户端/区域管理 │
│  • 看板/邮件通知  │
└──────────────────┘
```

## 核心功能

- **加固检查**：Linux 客户端执行加固脚本并自动修复不合规项；Windows 客户端只读采集 29 项加固状态（GPO 由域控下发）
- **合规比对**：Linux/Windows 标准配置管理，字段级标准值与检查结果自动比对，标记合规/不合规
- **客户端自动更新**：管理端上传新版安装包后，双端客户端每 5 分钟检查版本，自动下载并静默升级（保留本地配置）
- **系统看板**：客户端在线状态、区域分布、合规率统计
- **邮件通知**：SMTP 配置管理 + 每日合规报告定时推送（支持多收件人）
- **区域管理**：按区域组织客户端，区域维度统计

## 快速开始

### 环境要求
- Node.js >= 16.x
- Go >= 1.21
- Rust（构建 Windows 客户端时需要）
- MariaDB/MySQL 5.7+

### 1. 后端启动

```bash
cd backend

# 安装依赖
go mod tidy

# 编辑 config.yml 修改数据库、LDAP、JWT、安装包配置

# 构建
go build -o bin/server cmd/main.go

# 运行
./bin/server
```

后端服务将在 http://localhost:8080 启动，数据库表由 GORM AutoMigrate 自动创建。

### 2. 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

前端应用将在 http://localhost:8081 启动（`/api` 请求自动代理到后端 8080）。

### 3. Linux 客户端构建

```bash
# 一键交叉编译 + 打包 zip 安装包（参数为版本号，默认 1.1.0）
bash scripts/build-linux-client.sh 1.4.0
# 产物：项目根目录 linux-hardening-client_v1.4.0.zip
```

手动构建：

```bash
cd linux-client
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=1.4.0" -o ../dist/linux-hardening-client .
```

### 4. Windows 客户端构建

```bash
# macOS/Linux 交叉编译 + NSIS 打包（需安装 mingw-w64 与 makensis）
bash scripts/build-windows-client.sh 2.0.6
# 产物：dist_win/SystemHardening_WindowsClient_Setup_2.0.6.exe
```

脚本会自动同步 `Cargo.toml` 与 `windows_client.nsi` 中的版本号。

## API 接口

### 认证接口
- `POST /api/auth/login` - LDAP 登录
- `GET /api/profile` - 获取用户信息（JWT）

### 客户端接口（Token 认证）
- `POST /api/client/request-temp-token` - 请求临时安装 Token（5 分钟有效）
- `POST /api/client/register` - 客户端注册
- `POST /api/client/refresh-token` - 刷新 Token
- `POST /api/client/heartbeat` - 客户端心跳
- `POST /api/client/upload-data` - 上传 Linux 系统检查数据
- `POST /api/client/upload-data-windows` - 上传 Windows 加固检查数据
- `GET /api/client/check-update` - 检查客户端新版本（携带 `X-Client-Version` 头）

### 安装包接口
- `GET /api/packages/:type/download` - 下载安装包（公开，客户端自动更新使用）
- `POST /api/packages/upload` - 上传安装包（需 JWT，上传后自动同步版本号）
- `GET /api/packages/:type/info` - 获取安装包版本信息（需 JWT）

### 管理接口（需 JWT 认证）
- `GET /api/health` - 健康检查
- `GET /api/dashboard/stats` - 看板统计数据

**Linux 加固检查**
- `GET /api/linux-checks` - 获取 Linux 加固检查数据（列表）
- `GET /api/linux-checks/:id` - 获取检查详情

**Linux 标准配置**
- `POST /api/linux-standards` - 创建标准配置
- `GET /api/linux-standards` - 获取标准配置列表
- `GET /api/linux-standards/fields` - 获取可用字段
- `PUT /api/linux-standards/:id` - 更新标准配置
- `DELETE /api/linux-standards/:id` - 删除标准配置

**Windows 加固检查**
- `GET /api/windows-checks` - 获取 Windows 加固检查数据（列表）
- `GET /api/windows-checks/:id` - 获取检查详情

**Windows 标准配置**
- `POST /api/windows-standards` - 创建标准配置
- `GET /api/windows-standards` - 获取标准配置列表
- `GET /api/windows-standards/fields` - 获取可用字段
- `PUT /api/windows-standards/:id` - 更新标准配置
- `DELETE /api/windows-standards/:id` - 删除标准配置

**区域与客户端管理**
- `POST /api/regions` - 创建区域
- `GET /api/regions` - 获取区域列表
- `PUT /api/regions/:id/clients` - 配置区域关联客户端
- `DELETE /api/regions/:id` - 删除区域
- `GET /api/clients` - 获取客户端列表
- `DELETE /api/clients/:id` - 删除客户端（硬删除）

**邮件通知与报告计划**
- `GET /api/mail-config` - 获取 SMTP 邮件配置
- `PUT /api/mail-config` - 保存邮件配置
- `POST /api/mail/test` - 发送测试邮件
- `GET /api/report-schedules` - 报告计划列表
- `POST /api/report-schedules` - 创建报告计划
- `PUT /api/report-schedules/:id` - 更新报告计划
- `DELETE /api/report-schedules/:id` - 删除报告计划
- `POST /api/report-schedules/:id/send` - 立即发送报告

### 跨域配置

后端已配置 CORS，允许所有来源的请求。生产环境建议限制来源。

## 数据库配置

后端配置文件 `backend/config.yml`：

```yaml
server:
  port: "8080"

database:
  host: "your-db-host"
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

packages:
  linux_package_dir: "./packages/linux"
  windows_package_dir: "./packages/windows"
  server_url: "http://后端IP:8080"   # 客户端下载更新包时使用的地址
```

数据库使用 GORM AutoMigrate 自动建表和同步字段结构。

## Linux 加固客户端

- **语言**：Go，纯 Go 实现，无 CGO 依赖
- **运行方式**：systemd 服务（root 运行），启动即检查，之后每 24 小时自动检查
- **功能**：自动执行 `System_Check-1.2.sh` 安全加固脚本（软件包、密码策略、SSH、PAM、文件权限、时间同步等），将结果上报后端，支持 Token 自动注册/刷新
- **自动更新**：每 5 分钟向后端检查版本，发现新版本后自动下载 zip 包、备份配置、解压安装并重启服务
- **安装**：zip 安装包 + 交互式安装脚本，详见 [linux-client/README.md](linux-client/README.md) 与 [dist/README.md](dist/README.md)

## Windows 加固客户端

- **语言**：Rust，Windows 服务方式运行（`SystemHardeningWinClient`，开机自启 + 故障自动重启）
- **功能**：只读采集（不改动系统配置，加固由域控 GPO 完成）：
  - 系统信息（主机名、域、IP、OS、许可证状态）
  - 账户密码策略（15 项）、审计策略（9 项）
  - 移动存储设备控制、屏幕保护策略
  - 管理员/来宾账户状态
- **采集来源**：WMI + 注册表 + secedit + 域策略文件（GptTmpl.inf / registry.pol，支持 AD 域环境）
- **自动更新**：每 5 分钟检查版本，下载 NSIS 安装包后静默安装（`/S`），保留本地配置
- **安装**：NSIS 安装包，自动注册服务并启动，详见 [windows-client/README.md](windows-client/README.md)

## Release 发布

正式安装包通过 GitHub Releases 发布：

| 平台 | 安装包 | 说明 |
|------|--------|------|
| Linux | `linux-hardening-client_v<版本>.zip` | 二进制 + 加固脚本 + 安装/卸载脚本 |
| Windows | `SystemHardening_WindowsClient_Setup_<版本>.exe` | NSIS 安装包（服务注册、配置生成） |

下载地址：https://github.com/YanGLweI/SystemHardening/releases

当前最新版本：**Linux 客户端 v1.4.0** / **Windows 客户端 v2.0.6**

## 后续开发计划

- [x] 用户认证（LDAP 域控集成）
- [x] 客户端自动注册与 Token 管理
- [x] Linux 系统加固检查与数据上报
- [x] Windows 加固检查模块（独立数据表、API、前端页面）
- [x] Windows / Linux 标准配置管理模块
- [x] 合规性标准配置管理
- [x] 区域管理与客户端管理
- [x] 系统看板与统计模块
- [x] 邮件通知与每日合规报告推送
- [x] 客户端安装包管理与自动更新（双端）
- [x] GitHub Releases 发布
- [ ] 权限管理系统
- [ ] 批量部署工具
- [ ] 合规性报告导出
- [ ] 审计日志

## License

MIT License
