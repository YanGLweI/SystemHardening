<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <h3>系统加固平台</h3>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="el-menu-vertical"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        @select="handleMenuSelect"
      >
        <el-menu-item index="/home">
          <i class="el-icon-s-platform"></i>
          <span>系统看板</span>
        </el-menu-item>        <el-submenu index="security-hardening">
          <template #title>
            <i class="el-icon-lock"></i>
            <span>安全加固</span>
          </template>
          <el-menu-item index="/linux-hardening">Linux 加固</el-menu-item>
          <el-menu-item index="/linux-standard">Linux 标准配置</el-menu-item>
        </el-submenu>

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
      username: ''
    }
  },
  computed: {
    activeMenu() {
      // 根据当前路径匹配菜单项
      const path = this.$route.path
      if (path === '/home') return '/home'
      if (path === '/linux-hardening') return '/linux-hardening'
      if (path === '/linux-standard') return '/linux-standard'
      return path
    },
    breadcrumbText() {
      const currentRoute = this.$route.name
      const routes = {
        'Home': '首页 / 系统看板',
        'LinuxHardening': '首页 / Linux 加固',
        'LinuxStandard': '首页 / Linux 标准配置',
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
      if (['/home', '/linux-hardening', '/linux-standard'].includes(index)) {
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

.sidebar {
  background-color: #304156;
  overflow-y: auto;
  height: 100%;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4c;
  margin-bottom: 20px;
}

.logo h3 {
  color: #fff;
  font-size: 18px;
  margin: 0;
}

.el-menu-vertical {
  border-right: none;
  height: 100%;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
  line-height: 60px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.breadcrumb {
  font-size: 16px;
  color: #606266;
}

.user-info {
  cursor: pointer;
}

.user-name {
  color: #606266;
  font-size: 14px;
}

.user-name i.el-icon-arrow-down {
  margin-left: 5px;
  font-size: 12px;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  height: calc(100vh - 60px);
  overflow-y: auto;
}
</style>
