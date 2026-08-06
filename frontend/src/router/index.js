import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/system-managing/Home.vue'
import Login from '../views/auth/Login.vue'
import About from '../views/system-managing/About.vue'
import ClientManagement from '../views/system-managing/ClientManagement.vue'
import RegionManagement from '../views/system-managing/RegionManagement.vue'
import StandardManagement from '../views/system-managing/StandardManagement.vue'
import CheckManagement from '../views/system-managing/CheckManagement.vue'
import MailNotification from '../views/system-managing/MailNotification.vue'
import NotFound from '../views/auth/NotFound.vue'
import sessionState, { showSessionExpiredDialog } from '../utils/session'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { title: '登录 - 系统加固平台' }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: NotFound,
    meta: { title: '页面未找到 - 系统加固平台' }
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
        component: Home,
        meta: { requiresAuth: true, title: '首页 - 系统加固平台' }
      },
      {
        path: '/about',
        name: 'About',
        component: About,
        meta: { requiresAuth: true, title: '关于 - 系统加固平台' }
      },
      {
        path: '/check',
        name: 'CheckManagement',
        component: CheckManagement,
        redirect: '/check/linux',
        children: [
          {
            path: '/check/linux',
            name: 'LinuxHardeningContent',
            component: () => import('../views/content/hardening/LinuxHardeningContent.vue'),
            meta: { requiresAuth: true, title: 'Linux 加固检查 - 系统加固平台' }
          },
          {
            path: '/check/windows',
            name: 'WindowsHardeningContent',
            component: () => import('../views/content/hardening/WindowsHardeningContent.vue'),
            meta: { requiresAuth: true, title: 'Windows 加固检查 - 系统加固平台' }
          }
        ]
      },
      {
        path: '/standard',
        name: 'StandardManagement',
        component: StandardManagement,
        redirect: '/standard/linux',
        children: [
          {
            path: '/standard/linux',
            name: 'LinuxStandardContent',
            component: () => import('../views/content/standards/LinuxStandardContent.vue'),
            meta: { requiresAuth: true, title: 'Linux 标准配置 - 系统加固平台' }
          },
          {
            path: '/standard/windows',
            name: 'WindowsStandardContent',
            component: () => import('../views/content/standards/WindowsStandardContent.vue'),
            meta: { requiresAuth: true, title: 'Windows 标准配置 - 系统加固平台' }
          }
        ]
      },
      {
        path: '/client-management',
        name: 'ClientManagement',
        component: ClientManagement,
        meta: { requiresAuth: true, title: '客户端管理 - 系统加固平台' }
      },
      {
        path: '/region-management',
        name: 'RegionManagement',
        component: RegionManagement,
        meta: { requiresAuth: true, title: '区域管理 - 系统加固平台' }
      },
      {
        path: '/mail-notification',
        name: 'MailNotification',
        component: MailNotification,
        meta: { requiresAuth: true, title: '邮件通知 - 系统加固平台' }
      }
    ]
  },
  {
    // 通配符路由：所有不存在的路径都重定向到 404 页面
    path: '*',
    redirect: '/404'
  }
]

const router = new VueRouter({
  mode: 'hash',  // 改为 hash 模式以避免后端路由问题
  base: process.env.BASE_URL,
  routes
})

// 全局处理 Vue Router 内部错误
// 防止路由重定向等导航失败的未捕获错误
if (typeof window !== 'undefined') {
  window.addEventListener('unhandledrejection', (event) => {
    const err = event.reason
    if (err && err._isRouter === true) {
      // Vue Router 内部导航失败（重定向、中止、取消、重复），非致命错误
      // preventDefault 阻止浏览器默认行为，stopImmediatePropagation 阻止其他监听器（如 webpack overlay）接收
      event.preventDefault()
      event.stopImmediatePropagation()
    }
  })
}

// Vue Router 导航失败错误类型（redirected/aborted/cancelled/duplicated）
const NAVIGATION_FAILURE_TYPES = [2, 4, 8, 16]

// 吞掉导航失败错误，其他错误继续抛出
const swallowNavigationFailure = (err) => {
  if (err && err._isRouter === true && NAVIGATION_FAILURE_TYPES.includes(err.type)) {
    return
  }
  return Promise.reject(err)
}

// 包装 router.push/replace：捕获导航失败产生的 Promise 拒绝
// 防止未处理的 rejection 被 webpack-dev-server overlay 显示为运行时错误
const originalPush = router.push.bind(router)
router.push = function (location, onComplete, onAbort) {
  const result = originalPush(location, onComplete, onAbort)
  if (result && typeof result.catch === 'function') {
    return result.catch(swallowNavigationFailure)
  }
  return result
}

const originalReplace = router.replace.bind(router)
router.replace = function (location, onComplete, onAbort) {
  const result = originalReplace(location, onComplete, onAbort)
  if (result && typeof result.catch === 'function') {
    return result.catch(swallowNavigationFailure)
  }
  return result
}

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  // 设置页面标题
  document.title = to.meta.title || '系统加固平台'
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !token) {
    if (sessionState.isExpired) {
      // 会话已过期：阻止导航，弹出对话框
      showSessionExpiredDialog()
      next(false)
    } else {
      // 未登录（首次访问等）：正常重定向到登录页，并带上原始路径作为回调
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    }
  } else if (to.path === '/login' && token) {
    // 如果已登录，访问登录页则跳转到首页
    next('/')
  } else {
    next()
  }
})

export default router
