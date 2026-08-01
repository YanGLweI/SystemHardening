# System Hardening Platform - 系统加固管理平台

基于 Vue2 + Go 构建的前后端分离系统加固管理平台。

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

## 项目结构

```
system_hardening/
├── backend/             # 后端服务
│   ├── cmd/            # 主程序入口
│   ├── configs/        # 配置文件
│   ├── database/       # 数据库连接
│   ├── models/         # 数据模型
│   ├── routes/         # 路由配置
│   ├── middleware/     # 中间件
│   ├── handlers/       # 业务处理
│   ├── controllers/    # 控制器
│   ├── migrations/     # 数据库迁移
│   └── .env           # 环境变量配置
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

# 配置环境变量（已在 .env 中配置默认值）
cp .env .env.local
# 编辑 .env.local 修改数据库配置

# 运行开发服务器
go run cmd/main.go
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

## API 接口

### 健康检查
- `GET /api/health` - 检查服务状态
- `GET /api/test` - 测试 API 连通性

### 跨域配置

后端已配置 CORS，允许所有来源的请求。生产环境建议限制来源。

前端已配置代理，开发环境下 `/api` 请求会转发到后端服务 `http://localhost:8080`

## 数据库配置

默认数据库配置（`.env`）：
```
DB_HOST=10.60.254.127
DB_PORT=3306
DB_USER=it
DB_PASSWORD=a*999999
DB_NAME=system_hardening
```

如需修改，请编辑相应配置文件或设置环境变量。

## 开发规范

### 代码风格
- Go: 遵循官方 Go Code Review Comments
- JavaScript/Vue: ESLint + Prettier 配置
- CSS: BEM 命名规范

### 提交规范
建议使用 Git Commit Linter 确保提交信息规范

## 注意事项

⚠️ **重要**：
1. 确保能访问数据库地址 10.60.254.127
2. 生产环境请勿使用 `.env` 中的默认密码
3. Vue 2 已达到 EOL，新项目建议使用 Vue 3

## 后续开发计划

- [ ] 用户认证与授权
- [ ] 权限管理系统
- [ ] 系统加固任务管理
- [ ] 安全漏洞扫描
- [ ] 合规性报告
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
