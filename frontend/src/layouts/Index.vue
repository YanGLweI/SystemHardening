<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="sidebarWidth + 'px'" class="sidebar">
      <div class="logo">
        <div class="logo-icon">
          <img src="favicon.ico" alt="Logo" />
        </div>
        <h3 class="logo-title">系统加固平台</h3>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="el-menu-vertical"
        background-color="#059669"
        text-color="rgba(255, 255, 255, 0.75)"
        active-text-color="#FFFFFF"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/home">
          <i class="el-icon-s-platform"></i>
          <span>系统看板</span>
        </el-menu-item>
        <el-submenu index="security-hardening">
          <template #title>
            <i class="el-icon-lock"></i>
            <span>安全加固</span>
          </template>
          <el-menu-item index="/linux-hardening">Linux 加固</el-menu-item>
          <el-menu-item index="/linux-standard">Linux 标准</el-menu-item>
        </el-submenu>
        <el-menu-item index="/client-management">
          <i class="el-icon-monitor"></i>
          <span>客户端管理</span>
        </el-menu-item>

      </el-menu>
    </el-aside>

    <!-- 右侧主容器 -->
    <el-container>
      <!-- 头部 -->
      <el-header class="header">
        <div class="header-content">
          <div class="breadcrumb">
            {{ breadcrumbText }}
          </div>
          <div class="user-info">
            <el-dropdown>
              <span class="user-name">
                <i class="el-icon-user-solid"></i>
                {{ username || '未登录' }}
                <i class="el-icon-arrow-down"></i>
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item>
                  <i class="el-icon-setting"></i>
                  个人设置
                </el-dropdown-item>
                <el-dropdown-item divided @click.native="logout">
                  <i class="el-icon-switch-button"></i>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
export default {
  name: 'Layout',
  data() {
    return {
      username: '',
      sidebarWidth: 240
    }
  },
  computed: {
    activeMenu() {
      // 根据当前路径匹配菜单项
      const path = this.$route.path
      if (path === '/home') return '/home'
      if (path === '/linux-hardening') return '/linux-hardening'
      if (path === '/linux-standard') return '/linux-standard'
      if (path === '/client-management') return '/client-management'
    },
    breadcrumbText() {
      const currentRoute = this.$route.name
      const routes = {
        'Home': '首页 / 系统看板',
        'LinuxHardening': '首页 / Linux 加固',
        'LinuxStandard': '首页 / Linux 标准配置',
        'ClientManagement': '首页 / 客户端管理',
        'About': '关于'
      }
      return routes[currentRoute] || '首页'
    }
  },
  created() {
    this.username = localStorage.getItem('username') || ''
  },
  methods: {
    handleMenuSelect(index) {
      // 如果是有效路径且不是当前路由，才进行跳转
      if (['/home', '/linux-hardening', '/linux-standard', '/client-management'].includes(index)) {
        if (this.$route.path !== index) {
          this.$router.push(index)
        }
      }
    },
    logout() {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      window.location.href = '/login'
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

/* 🟢 侧边栏 - 薄荷绿背景 */
.sidebar {
  background: linear-gradient(180deg, #047857 0%, #059669 50%, #10B981 100%);
  overflow: hidden; /* 移除多余滚动条 */
  height: 100%;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  position: relative;
}

.logo {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 0 16px;
}

.logo-icon {
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-icon img {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.logo-title {
  color: var(--color-white);
  font-size: 18px;
  font-weight: 600;
  margin: 0;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.el-menu-vertical {
  border-right: none;
  height: 100%;
  background: transparent !important;
}

/* 🟢 菜单项样式优化 - 强制覆盖 Element UI 内联样式 */
:deep(.el-menu-item) {
  background-color: transparent !important;
  color: rgba(255, 255, 255, 0.75) !important;
  transition: all 0.2s ease !important;
  border-left: 3px solid transparent;
  margin-bottom: 4px;
}

:deep(.el-menu-item i) {
  color: rgba(255, 255, 255, 0.75) !important;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
}

:deep(.el-menu-item:hover i) {
  color: #fff !important;
}

:deep(.el-menu-item:focus) {
  background-color: rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
}

:deep(.el-menu-item:focus i) {
  color: #fff !important;
}

:deep(.el-menu-item.is-active) {
  background: rgba(255, 255, 255, 0.15) !important;
  color: #fff !important;
  border-left: 3px solid #34D399;
  font-weight: 500;
}

:deep(.el-menu-item.is-active i) {
  color: #fff !important;
}

/*  子菜单样式 */
:deep(.el-submenu__title) {
  color: rgba(255, 255, 255, 0.75) !important;
  transition: all 0.2s ease !important;
  border-left: 3px solid transparent;
  background-color: transparent !important;
}

:deep(.el-submenu__title i) {
  color: rgba(255, 255, 255, 0.75) !important;
}

:deep(.el-submenu__title:hover) {
  background-color: rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
}

:deep(.el-submenu__title:hover i) {
  color: #fff !important;
}

:deep(.el-submenu__title .el-icon-arrow-down) {
  transition: transform var(--transition-base);
}

:deep(.el-submenu.is-opened > .el-submenu__title .el-icon-arrow-down) {
  transform: rotate(180deg);
}

/*  子菜单项样式 */
:deep(.el-submenu .el-menu-item) {
  background-color: transparent !important;
  color: rgba(255, 255, 255, 0.75) !important;
  transition: all 0.2s ease !important;
  border-left: 3px solid transparent;
  margin-bottom: 2px;
}

:deep(.el-submenu .el-menu-item i) {
  color: rgba(255, 255, 255, 0.75) !important;
}

:deep(.el-submenu .el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
}

:deep(.el-submenu .el-menu-item:hover i) {
  color: #fff !important;
}

:deep(.el-submenu .el-menu-item.is-active) {
  background: rgba(255, 255, 255, 0.15) !important;
  color: #fff !important;
  border-left: 3px solid #34D399;
  font-weight: 500;
}

:deep(.el-submenu .el-menu-item.is-active i) {
  color: #fff !important;
}

/* 🎨 顶部导航栏 */
.header {
  background: var(--color-bg-header);
  border-bottom: 1px solid var(--color-border-light);
  padding: 0 var(--spacing-6);
  line-height: var(--header-height);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.breadcrumb {
  font-size: 16px;
  color: var(--color-text-primary);
  font-weight: 500;
}

.user-info {
  cursor: pointer;
  padding: 6px 12px;
  border-radius: var(--radius-full);
  transition: background var(--transition-fast);
}

.user-info:hover {
  background: var(--color-bg-hover);
}

.user-name {
  color: var(--color-text-regular);
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.user-name i {
  color: var(--color-secondary);
}

.user-name i.el-icon-user-solid {
  color: var(--color-primary);
  font-size: 18px;
}

.user-name i.el-icon-arrow-down {
  margin-left: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* 🎨 主内容区 */
.main-content {
  background: var(--color-bg-page);
  padding: var(--spacing-6);
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
  position: relative;
}

.main-content::-webkit-scrollbar {
  width: 6px;
}

.main-content::-webkit-scrollbar-track {
  background: var(--color-gray-100);
  border-radius: var(--radius-md);
}

.main-content::-webkit-scrollbar-thumb {
  background: var(--color-gray-300);
  border-radius: var(--radius-md);
  transition: background var(--transition-base);
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: var(--color-primary);
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  :deep(.breadcrumb) {
    font-size: 14px;
  }
  
  .user-name span {
    display: none;
  }
  
  .main-content {
    padding: var(--spacing-4);
  }
}
</style>
