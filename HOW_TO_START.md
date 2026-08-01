# 系统加固管理平台 - 快速启动指南

## ✅ 项目已搭建完成

前后端基础框架已成功搭建，并且都能够在开发环境中正常启动。

## 🚀 启动服务

### 启动后端（终端 1）
```bash
cd backend
go run cmd/main.go
```
访问：http://localhost:8080
- 健康检查：http://localhost:8080/api/health
- API 测试：http://localhost:8080/api/test

**✅ 后端状态**: 运行中 ✓  
**数据库连接**: 已成功连接到 10.60.254.127:3306/system_hardening

### 启动前端（终端 2）
```bash
cd frontend
npm run dev
```
访问：http://localhost:8081

**✅ 前端状态**: 运行中 ✓

## 📋 项目结构

```
system_hardening/
├── backend/                  # 后端 Go 服务
│   ├── cmd/main.go          # 入口文件
│   ├── configs/             # 配置管理
│   ├── database/            # 数据库连接
│   ├── models/              # 数据模型（含 User 模型）
│   ├── routes/              # 路由配置
│   ├── middleware/          # 中间件（日志、错误处理）
│   ├── handlers/            # 业务处理器（待实现）
│   ├── controllers/         # 控制器（待实现）
│   ├── migrations/          # 数据库迁移（待实现）
│   ├── .env                 # 环境变量配置
│   └── README.md
│
└── frontend/                # 前端 Vue 应用
    ├── public/index.html    # HTML 模板
    └── src/
        ├── main.js          # 入口文件
        ├── App.vue          # 根组件
        ├── router/          # 路由配置
        ├── store/           # Vuex 状态管理
        ├── api/             # API 请求封装
        ├── views/           # 页面组件（Home, About）
        └── assets/          # 静态资源
    └── README.md
```

## 🔧 当前已实现功能

### 后端
- ✅ Go Gin 框架基础配置
- ✅ GORM 数据库连接（MySQL/MariaDB）
- ✅ CORS 跨域支持
- ✅ 日志中间件
- ✅ 错误处理中间件
- ✅ 用户数据模型（User）
- ✅ 健康检查接口 `/api/health`
- ✅ 测试接口 `/api/test`

### 前端
- ✅ Vue2 + ElementUI 项目初始化
- ✅ Vue Router 路由配置
- ✅ Vuex 状态管理框架
- ✅ Axios HTTP 客户端封装
- ✅ API 代理配置（dev 环境下转发到后端）
- ✅ Home 页面
- ✅ About 页面

## 🗄️ 数据库配置

数据库信息：
- **地址**: 10.60.254.127
- **端口**: 3306
- **用户**: it
- **密码**: a*999999
- **数据库**: system_hardening

配置文件位于 `backend/.env`

## 📝 下一步开发建议

1. **创建数据库表**
   ```sql
   CREATE DATABASE IF NOT EXISTS system_hardening DEFAULT CHARACTER SET utf8mb4;
   
   -- 在 application 中会自动通过 GORM 自动迁移创建表
   ```

2. **用户管理模块**
   - 用户注册
   - 用户登录
   - 用户列表
   - 用户详情

3. **权限管理**
   - 角色管理
   - 权限分配

4. **系统加固功能**
   - 安全策略管理
   - 漏洞扫描
   - 合规性检测

## ⚠️ 注意事项

1. **开发环境**: 确保能访问数据库服务器 10.60.254.127
2. **生产部署**: 请修改 `.env` 中的敏感信息
3. **ESLint**: 前端有一些 ESLint 警告，但不影响运行
4. **Vue 版本**: 当前使用 Vue2（已 EOL），建议后续升级到 Vue3

## 🛠️ 常用命令

### 后端
```bash
# 安装依赖
go mod tidy

# 运行开发服务器
go run cmd/main.go

# 构建生产版本
go build -o system-hardening-backend cmd/main.go
```

### 前端
```bash
# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 构建生产版本
npm run build

# 执行代码检查
npm run lint
```

## 📞 技术支持

如有问题请查看：
- 后端日志：`cat /tmp/backend.log`
- 前端日志：`cat /tmp/frontend.log`

祝开发顺利！🎉
