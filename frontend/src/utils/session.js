import Vue from 'vue'

/**
 * 会话过期响应式状态
 * 用于控制"会话已过期"对话框的显示与路由行为
 */
const state = Vue.observable({
  isExpired: false,    // 会话是否已过期
  dialogVisible: false // 对话框是否显示
})

/**
 * 显示会话过期对话框
 * 同时设置过期状态，以便路由守卫拦截后续导航
 */
export function showSessionExpiredDialog() {
  state.isExpired = true
  state.dialogVisible = true
}

/**
 * 隐藏会话过期对话框（用户点击"停留本页"）
 * isExpired 保持 true，以便后续路由点击或请求时再次弹出
 */
export function hideSessionExpiredDialog() {
  state.dialogVisible = false
}

/**
 * 清除会话过期状态（用户点击"登录"）
 * 清除后路由守卫将允许跳转到登录页
 */
export function clearSessionExpired() {
  state.isExpired = false
  state.dialogVisible = false
}

export default state