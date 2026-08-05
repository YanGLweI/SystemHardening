<template>
  <div id="app">
    <router-view />
    <!-- 会话过期对话框 -->
    <el-dialog
      :visible.sync="sessionExpiredDialogVisible"
      title="会话已过期"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      width="420px"
      top="30vh"
    >
      <div class="session-expired-body" style="text-align: center; padding: 20px 0;">
        <i class="el-icon-warning-outline" style="color: #e6a23c; font-size: 48px;"></i>
        <p style="margin-top: 16px; font-size: 15px; color: #606266;">
          您的登录会话已过期，请重新登录以继续使用系统。
        </p>
      </div>
      <span slot="footer" class="dialog-footer" style="text-align: center;">
        <el-button @click="stayOnPage">停留本页</el-button>
        <el-button type="primary" @click="goToLogin">登录</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import sessionState, { hideSessionExpiredDialog, clearSessionExpired } from './utils/session'

export default {
  name: 'App',
  computed: {
    sessionExpiredDialogVisible: {
      get() {
        return sessionState.dialogVisible
      },
      set(val) {
        if (!val) {
          hideSessionExpiredDialog()
        }
      }
    }
  },
  methods: {
    stayOnPage() {
      hideSessionExpiredDialog()
    },
    goToLogin() {
      clearSessionExpired()
      this.$router.push('/login')
    }
  }
}
</script>

<style>
#app {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100%;
}
</style>
