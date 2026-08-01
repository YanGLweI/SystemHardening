# System Hardening Platform Frontend

基于 Vue2 + ElementUI 的系统加固管理平台前端界面。

## 项目结构

```
frontend/
├── public/                 # 静态资源
│   └── index.html
├── src/
│   ├── api/               # API 请求
│   │   └── request.js
│   ├── assets/            # 资源文件
│   │   └── styles/
│   ├── components/        # 公共组件
│   ├── config/            # 配置文件
│   ├── router/            # 路由配置
│   │   └── index.js
│   ├── store/             # Vuex 状态管理
│   │   └── index.js
│   ├── utils/             # 工具函数
│   ├── views/             # 页面组件
│   │   ├── Home.vue
│   │   └── About.vue
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── babel.config.js
├── package.json
├── vue.config.js
└── README.md
```

## 技术栈

- Vue 2.6.14
- Vue Router 3.5.4
- Vuex 3.6.2
- Element UI 2.15.13
- Axios 1.6.2
- Sass

## 安装步骤

1. 进入项目目录
```bash
cd frontend
```

2. 安装依赖
```bash
npm install
```

3. 启动开发服务器
```bash
npm run dev
```

访问 http://localhost:8081 查看应用

## 构建生产版本

```bash
npm run build
```

## 功能特性

- [x] Vue Router 路由配置
- [x] Vuex 状态管理
- [x] Axios HTTP 客户端
- [x] CORS 跨域支持
- [x] Element UI 组件库
- [ ] 用户认证
- [ ] 路由守卫
- [ ] 错误处理

## 开发说明

### 代理配置

前端开发环境已配置代理，所有 `/api` 请求会被转发到后端服务 `http://localhost:8080`

### API 调用示例

```javascript
import request from '@/api/request'

// GET 请求
export const getUserList = () => {
  return request({
    url: '/users',
    method: 'get'
  })
}

// POST 请求
export const createUser = (data) => {
  return request({
    url: '/users',
    method: 'post',
    data: data
  })
}
```
