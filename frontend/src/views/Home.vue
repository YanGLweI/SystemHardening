<template>
  <div class="home-container">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <h3>系统加固平台</h3>
        </div>
        <el-menu
          default-active="1"
          class="el-menu-vertical"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
          @select="handleMenuSelect"
        >
          <el-menu-item index="1">
            <i class="el-icon-s-platform"></i>
            <span>系统看板</span>
          </el-menu-item>
          <el-submenu index="2">
            <template #title>
              <i class="el-icon-user"></i>
              <span>用户管理</span>
            </template>
            <el-menu-item index="2-1">用户列表</el-menu-item>
            <el-menu-item index="2-2">权限配置</el-menu-item>
          </el-submenu>
          <el-submenu index="3">
            <template #title>
              <i class="el-icon-lock"></i>
              <span>安全加固</span>
            </template>
            <el-menu-item index="3-1">
              <i class="el-icon-document-copy"></i>
              <span>策略管理</span>
            </el-menu-item>
            <el-menu-item index="3-2">
              <i class="el-icon-warning"></i>
              <span>漏洞扫描</span>
            </el-menu-item>
            <el-menu-item index="3-3">
              <i class="el-icon-coin"></i>
              <span>合规检测</span>
            </el-menu-item>
            <el-menu-item index="3-4">
              <i class="el-icon-server"></i>
              <span>Linux 加固</span>
            </el-menu-item>
          </el-submenu>
          <el-submenu index="4">
            <template #title>
              <i class="el-icon-document"></i>
              <span>报表中心</span>
            </template>
            <el-menu-item index="4-1">
              <i class="el-icon-pie-chart"></i>
              <span>审计报告</span>
            </el-menu-item>
            <el-menu-item index="4-2">
              <i class="el-icon-data-line"></i>
              <span>统计图表</span>
            </el-menu-item>
          </el-submenu>
          <el-menu-item index="5">
            <i class="el-icon-setting"></i>
            <span>系统设置</span>
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
          <!-- 根据当前激活菜单显示不同内容 -->
          <component :is="currentComponent" v-if="currentComponent" />
          
          <!-- Linux 加固页面（作为子组件嵌入） -->
          <LinuxHardening v-show="activeView === 'linux'" @refresh="fetchData" />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import LinuxHardening from './LinuxHardening.vue'

export default {
  name: 'Home',
  components: {
    LinuxHardening
  },
  data() {
    return {
      currentMenu: '1',
      activeView: 'home', // 'home' | 'linux'
      username: ''
    }
  },
  computed: {
    currentComponent() {
      if (this.activeView === 'home') {
        return 'home-component'
      }
      return null
    },
    breadcrumbText() {
      return this.activeView === 'linux' ? '首页 / Linux 加固' : '首页 / 系统看板'
    }
  },
  created() {
    // 从 localStorage 获取用户名
    this.username = localStorage.getItem('username') || ''
  },
  methods: {
    handleMenuSelect(index, indexPath) {
      console.log('菜单选择:', index, indexPath)
      if (index === '3-4') {
        this.activeView = 'linux'
      } else if (index === '1') {
        this.activeView = 'home'
      }
    },
    fetchData() {
      // 数据刷新回调
      console.log('刷新数据...')
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
.home-container {
  height: 100vh;
  overflow: hidden;
}

.layout-container {
  height: 100%;
}

/* 侧边栏样式 */
.sidebar {
  background-color: #304156;
  overflow-y: auto;
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
}

/* 头部样式 */
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

/* 主内容区样式 */
.main-content {
  background-color: #f0f2f5;
  padding: 20px;
}

.development-card {
  min-height: calc(100vh - 140px);
}

.card-header {
  display: flex;
  align-items: center;
  color: #303133;
}

.card-header i {
  margin-right: 8px;
  font-size: 18px;
}

.card-content {
  padding: 60px 20px;
  text-align: center;
  color: #606266;
}

.card-content .el-icon-large {
  font-size: 80px;
  color: #e6a23c;
  margin-bottom: 20px;
}

.card-content h3 {
  font-size: 24px;
  color: #303133;
  margin-bottom: 20px;
  font-weight: 500;
}

.card-content p {
  font-size: 16px;
  margin: 10px 0;
  color: #909399;
}
</style>
