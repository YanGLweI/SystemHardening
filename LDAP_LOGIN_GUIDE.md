# LDAP 登录系统集成指南

## 📋 概述

本系统实现了基于 LDAP + JWT 的完整认证机制，包含前端登录界面和后端认证服务。

### 核心特性

- ✅ **LDAP 域控认证**：通过 LDAPS 加密协议连接域控服务器
- ✅ **JWT Token 认证**：1 小时有效期的 Bearer Token
- ✅ **安全组权限控制**：只有指定安全组成员可登录
- ✅ **专业登录页面**：带入场动画和流畅的用户体验
- ✅ **路由守卫**：自动拦截未授权访问

---

## 🚀 快速开始

### 1. 后端配置

#### 1.1 安装依赖
```bash
cd backend
go mod tidy
```

#### 1.2 配置文件
复制环境变量示例文件：
```bash
cp .env.example .env
```

修改 `.env` 中的敏感信息（密码、JWT Secret 等）。

#### 1.3 启动服务
```bash
go run cmd/main.go
```

服务将在 `http://localhost:8080` 启动。

### 2. 前端配置

#### 2.1 安装依赖
```bash
cd frontend
npm install
```

#### 2.2 启动开发服务器
```bash
npm run dev
```

前端应用将在 `http://localhost:8081` 启动。

---

## 🔧 API 接口

### 1. 登录接口
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应示例**：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user_info": {
    "username": "ylw",
    "email": "ylw@hot.local",
    "details": {
      "cn": "ylw",
      "mail": "ylw@hot.local"
    }
  }
}
```

### 2. 获取用户信息
```http
GET /api/profile
Authorization: Bearer <token>
```

### 3. 健康检查（无需认证）
```http
GET /api/health
```

---

## 🔐 安全配置

### 1. JWT Secret Key
在 `backend/.env` 中设置强密钥：
```env
JWT_SECRET=your-super-secret-key-min-32-changes-please-change-in-production
```
⚠️ **生产环境必须替换为安全的随机密钥！**

### 2. LDAP 证书验证
默认使用 `InsecureSkipVerify=true`，生产环境建议：
```env
LDAP_INSECURE=false
LDAP_CERT_PATH=./certificate/ca.crt
```

### 3. HTTPS 强制启用
生产环境应配置 HTTPS 证书。

---

## 📝 技术架构

### 后端 (Go + Gin)
- **LDAP Service**: `backend/services/ldap_service.go`
- **JWT Utils**: `backend/utils/jwt_utils.go`
- **Auth Middleware**: `backend/middleware/auth.go`
- **Controllers**: `backend/handlers/auth_handler.go`
- **Router**: `backend/routes/router.go`

### 前端 (Vue.js + Element UI)
- **Login Page**: `frontend/src/views/Login.vue`
- **API Client**: `frontend/src/api/request.js`
- **Router Guard**: `frontend/src/router/index.js`

---

## 🎨 登录页面特性

### 动画效果
- 卡片入场：从下方淡入上移 (`translateY 30px`, `opacity`)
- 背景光晕：三个浮动渐变球体，营造科技感
- 加载状态：绿色安全徽章呼吸动画

### 表单验证
- 用户名格式校验
- 必填项检查
- 错误提示友好

### 用户体验
- 自动聚焦输入框
- Enter 键快捷提交
- 记住我功能
- 响应式设计（移动端适配）

---

## 🧪 测试脚本

### 测试登录
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test_user","password":"test_password"}'
```

### 测试受保护接口
```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer <your_token>"
```

---

## 📂 项目结构

```
backend/
├── configs/           # 配置文件
├── services/          # LDAP 服务层
├── utils/             # JWT 工具函数
├── middleware/        # JWT 认证中间件
├── handlers/          # 登录处理器
├── routes/            # 路由配置
├── certificate/       # CA 证书
└── .env.example       # 环境变量示例

frontend/
├── src/
│   ├── views/         # 登录页面
│   ├── api/           # API 客户端
│   └── router/        # 路由守卫
├── .env.development   # 开发环境配置
└── .env.production    # 生产环境配置
```

---

## ⚠️ 注意事项

1. **CA 证书路径**：确保 `backend/certificate/ca.crt` 存在且可读
2. **域控制器网络**：确保能访问 `10.60.254.252:636`
3. **安全组权限**：只有 IT 部安全组成员可以登录
4. **Token 有效期**：默认为 1 小时，过期需重新登录
5. **日志审计**：生产环境建议开启详细的日志记录

---

## 🐛 故障排除

### 问题 1: LDAP 连接失败
```
Error: failed to connect to LDAP server
```
**解决方法**：
- 检查网络连接
- 确认域控地址和端口正确
- 验证证书路径是否正确

### 问题 2: 认证失败
```
Error: invalid username or password
```
**解决方法**：
- 检查用户名和密码是否正确
- 确认用户在指定的安全组中
- 查看后端日志获取更多详情

### 问题 3: Token 无效
```
Error: token 无效或已过期
```
**解决方法**：
- 重新登录获取新 token
- 检查 JWT_SECRET 是否一致
- 确认系统时间准确

---

## 📞 技术支持

如有问题请联系 IT 部。
