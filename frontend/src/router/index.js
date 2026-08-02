import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import About from '../views/About.vue'
import LinuxHardening from '../views/LinuxHardening.vue'
import LinuxStandard from '../views/LinuxStandard.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { title: '登录 - 系统加固平台' }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('../layouts/Index.vue'),
    redirect: '/home',
    children: [
      {
        path: '/home',
        name: 'Home',
        component: () => import('../views/Home.vue'),
        meta: { requiresAuth: true, title: '首页 - 系统加固平台' }
      },
      {
        path: '/about',
        name: 'About',
        component: About,
        meta: { requiresAuth: true, title: '关于 - 系统加固平台' }
      },
      {
        path: '/linux-hardening',
        name: 'LinuxHardening',
        component: LinuxHardening,
        meta: { requiresAuth: true, title: 'Linux 加固 - 系统加固平台' }
      },
      {
        path: '/linux-standard',
        name: 'LinuxStandard',
        component: LinuxStandard,
        meta: { requiresAuth: true, title: 'Linux 标准配置 - 系统加固平台' }
      }
    ]
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  // 设置页面标题
  document.title = to.meta.title || '系统加固平台'
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !token) {
    // 重定向到登录页，并带上原始路径作为回调
    next({
      path: '/login',
      query: { redirect: to.fullPath }
    })
  } else if (to.path === '/login' && token) {
    // 如果已登录，访问登录页则跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router
