# System Hardening Platform Backend

基于 Go + Gin + GORM 的系统加固管理平台后端 API。

## 项目结构

```
backend/
├── cmd/                 # 主程序入口
│   └── main.go
├── configs/            # 配置文件
│   └── config.go
├── database/           # 数据库连接
│   └── db.go
├── models/             # 数据模型
│   └── user.go
├── routes/             # 路由配置
│   └── router.go
├── middleware/         # 中间件
│   └── logger.go
├── handlers/           # 业务处理器
├── controllers/        # 控制器
├── migrations/         # 数据库迁移
├── .env               # 环境变量配置
├── go.mod
└── README.md
```

## 环境要求

- Go 1.21+
- MariaDB/MySQL 5.7+

## 安装步骤

1. 克隆项目
```bash
git clone <repository_url>
cd backend
```

2. 安装依赖
```bash
go mod tidy
```

3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件配置数据库连接等信息
```

4. 运行开发服务器
```bash
go run cmd/main.go
```

## API 接口

### 健康检查
- `GET /api/health` - 检查服务状态

### 测试接口
- `GET /api/test` - 测试 API 连通性

## 开发说明

- 使用 Gin 框架处理 HTTP 请求
- 使用 GORM 进行数据库操作
- CORS 已配置，支持跨域请求
- 日志中间件记录请求信息
- 错误处理中间件统一错误响应
