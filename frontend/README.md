# System Hardening Platform Frontend

基于 Vue2 + Element UI 的系统加固管理平台前端，提供 LDAP 登录、Linux/Windows 加固检查展示、标准配置管理、客户端与区域管理等功能。

## 技术栈

- Vue 2.6.14
- Vue Router 3.5.4（含登录守卫）
- Vuex 3.6.2
- Element UI 2.15.13
- Axios 1.6.2（JWT 拦截器）
- Sass

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── api/               # API 请求模块
│   │   ├── request.js            # Axios 封装（Token 注入/错误处理）
│   │   ├── linux-checks.js       # Linux 加固检查与标准配置
│   │   ├── windows-checks.js     # Windows 加固检查与标准配置
│   │   ├── clients.js            # 客户端管理
│   │   └── regions.js            # 区域管理
│   ├── assets/            # 样式（main.css 全局变量）
│   ├── views/             # 页面组件
│   │   ├── Login.vue             # LDAP 登录
│   │   ├── Home.vue              # 首页
│   │   ├── LinuxHardening.vue    # Linux 加固检查列表/详情
│   │   ├── LinuxStandard.vue     # Linux 标准配置管理
│   │   ├── WindowsHardening.vue  # Windows 加固检查列表/详情
│   │   ├── WindowsStandard.vue   # Windows 标准配置管理
│   │   ├── ClientManagement.vue  # 客户端管理
│   │   ├── RegionManagement.vue  # 区域管理
│   │   └── About.vue             # 关于
│   ├── router/            # 路由配置（登录守卫）
│   ├── store/             # Vuex 状态管理
│   ├── App.vue            # 根组件
│   └── main.js            # 入口文件
├── babel.config.js
├── package.json
├── vue.config.js          # 开发代理配置
└── README.md
```

## 功能特性

- [x] LDAP 域控登录 + JWT 认证
- [x] 路由守卫（未登录跳转登录页）
- [x] Linux 加固检查数据列表与详情（模糊搜索、合规状态筛选）
- [x] Linux 标准配置管理（字段级标准值、合规比对）
- [x] Windows 加固检查数据列表与详情
- [x] Windows 标准配置管理
- [x] 客户端管理（查看/删除）
- [x] 区域管理与客户端关联
- [x] 薄荷绿主题（全局 CSS 变量）
- [x] Axios 拦截器（Token 注入、401 跳转登录）

## 安装与运行

```bash
# 1. 安装依赖
cd frontend
npm install

# 2. 开发运行
npm run dev
```

访问 http://localhost:8081 查看应用，使用 LDAP 账户登录。

## 构建生产版本

```bash
npm run build
# 产物输出到 dist/
```

生产环境接口地址通过 `.env.production` 配置：

```
VUE_APP_API_BASE_URL=http://后端IP:8080
```

## 开发说明

### 代理配置

开发环境所有 `/api` 请求由 `vue.config.js` 代理转发到后端 `http://localhost:8080`。

### API 调用示例

```javascript
import request from '@/api/request'

// GET 请求
export const getLinuxChecks = (params) => {
  return request({
    url: '/linux-checks',
    method: 'get',
    params
  })
}

// POST 请求（需 JWT，拦截器自动携带）
export const createStandard = (data) => {
  return request({
    url: '/linux-standards',
    method: 'post',
    data
  })
}
```

### 环境变量

| 文件 | 用途 |
|------|------|
| `.env.development` | 开发环境（代理到 localhost:8080） |
| `.env.production` | 生产环境（后端真实地址） |
