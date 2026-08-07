<template>
  <el-dialog 
    :title="dialogTitle" 
    :visible.sync="dialogVisible" 
    width="520px"
    @close="handleClose"
    class-check-trigger-dialog
    v-show="dialogVisible"
    custom-class="check-trigger-modal"
  >
    <!-- 状态图标 -->
    <div class="status-indicator" :class="indicatorClass">
      <div class="indicator-icon">
        <i v-if="taskInfo.status === 'pending'" class="el-icon-time"></i>
        <i v-else-if="taskInfo.status === 'sent'" class="el-icon-postcard"></i>
        <i v-else-if="taskInfo.status === 'executing'" class="el-icon-loading"></i>
        <i v-else-if="taskInfo.status === 'completed'" class="el-icon-success"></i>
        <i v-else-if="taskInfo.status === 'failed'" class="el-icon-error"></i>
        <i v-else-if="taskInfo.status === 'timeout'" class="el-icon-warning"></i>
        <i v-else class="el-icon-info"></i>
      </div>
      <div class="indicator-text">{{ indicatorText }}</div>
    </div>
    
    <!-- 信息卡片 -->
    <div class="info-card">
      <div class="info-row">
        <div class="info-label">
          <i class="el-icon-s-platform"></i>
          <span>目标主机</span>
        </div>
        <div class="info-value">{{ client.hostname || '-' }}</div>
      </div>
      <div class="info-separator"></div>
      <div class="info-row">
        <div class="info-label">
          <i class="el-icon-share"></i>
          <span>IP 地址</span>
        </div>
        <div class="info-value">{{ client.ip || '-' }}</div>
      </div>
      <div class="info-separator"></div>
      <div class="info-row">
        <div class="info-label">
          <i class="el-icon-coin"></i>
          <span>任务 ID</span>
        </div>
        <div class="info-value task-id">{{ taskInfo.taskId || taskId || '-' }}</div>
      </div>
    </div>
    
    <!-- 进度条区域 -->
    <transition name="fade-slide">
      <div class="progress-section" v-if="taskInfo.taskId && hasFetchedStatus">
        <el-progress 
          v-if="taskInfo.status"
          :percentage="statusPercentage" 
          :status="progressStatus"
          :stroke-width="24"
          :text-inside="true"
          :color="progressColor"
        />
        <p v-if="taskInfo.status" class="status-description">{{ statusText }}</p>
      </div>
    </transition>
    
    <!-- 错误信息 -->
    <transition name="fade-slide">
      <el-alert 
        v-if="taskInfo.errorMessage" 
        :title="taskInfo.errorMessage" 
        type="error" 
        show-icon 
        closable
        class="error-alert"
      />
    </transition>
    
    <!-- 操作按钮 -->
    <div slot="footer">
      <el-button @click="closeDialog">关闭</el-button>
      <el-button
        v-if="canRetry"
        type="warning"
        :loading="retrying"
        @click="handleRetry"
      >
        <i v-if="!retrying" class="el-icon-refresh-left"></i>
        {{ retrying ? '重试中...' : '重试' }}
      </el-button>
      <el-button 
        type="primary" 
        :loading="isExecuting"
        :disabled="taskInfo.status === 'completed'"
        @click="refreshStatus"
      >
        {{ isExecuting ? '执行中' : '刷新状态' }}
      </el-button>
    </div>
  </el-dialog>
</template>

<script>
import { getTaskStatus, deleteTask, triggerCheck } from '@/api/task-check'

export default {
  name: 'CheckTriggerDialog',
  props: {
    visible: Boolean,
    client: Object,
    taskId: String // 从父组件传入的 task_id
  },
  data() {
    return {
      taskInfo: {},
      pollTimer: null,
      pollInterval: 3000, // 3 秒轮询一次
      hasFetchedStatus: false,  // 标记是否已经获取过状态
      retrying: false  // 重试操作进行中
    }
  },
  computed: {
    // 使用 computed 属性避免直接修改 prop
    dialogVisible: {
      get() {
        return this.visible
      },
      set(value) {
        this.$emit('update:visible', value)
      }
    },
    dialogTitle() {
      switch (this.taskInfo.status) {
        case 'pending': return '等待任务下发...'
        case 'sent': return '任务已下发'
        case 'executing': return '正在执行加固检查...'
        case 'completed': return '检查完成'
        case 'failed': return '检查失败'
        case 'timeout': return '检查超时'
        default: return '立即检查'
      }
    },
    statusPercentage() {
      const statusMap = {
        'pending': 10,
        'sent': 30,
        'executing': 75,
        'completed': 100,
        'failed': 80,
        'timeout': 80
      }
      return statusMap[this.taskInfo.status] || 0
    },
    progressColor() {
      if (!this.taskInfo.status) return undefined
      const colorMap = {
        'pending': '#10B981',    // 绿色
        'sent': '#34D399',       // 浅绿
        'executing': '#6366F1',  // 靛蓝
        'completed': '#10B981',  // 绿色
        'failed': '#EF4444',     // 红色
        'timeout': '#F59E0B'     // 橙色
      }
      return colorMap[this.taskInfo.status] || '#10B981'
    },
    progressStatus() {
      // 用于 el-progress 组件的 status 属性
      if (this.taskInfo.status === 'completed') {
        return 'success'
      }
      if (['failed', 'timeout'].includes(this.taskInfo.status)) {
        return 'exception'
      }
      // pending, sent, executing 不设置 status
      return undefined
    },
    indicatorClass() {
      // 用于状态指示器图标的背景样式类
      if (!this.taskInfo.status) return ''
      if (this.taskInfo.status === 'completed') {
        return 'success'
      }
      if (['failed', 'timeout'].includes(this.taskInfo.status)) {
        return 'exception'
      }
      return ''
    },
    statusText() {
      const messages = {
        'pending': '任务已创建，等待客户端轮询...',
        'sent': '指令已下发，等待客户端响应...',
        'executing': '客户端正在执行检查，请稍候...',
        'completed': '加固检查已完成',
        'failed': '执行失败：' + (this.taskInfo.error_message || '未知原因'),
        'timeout': '执行超时，请重试'
      }
      return messages[this.taskInfo.status] || ''
    },
    indicatorText() {
      if (!this.taskInfo.status) return '检查中'
      const textMap = {
        'pending': '等待中',
        'sent': '已下发',
        'executing': '执行中',
        'completed': '已完成',
        'failed': '失败',
        'timeout': '超时'
      }
      return textMap[this.taskInfo.status] || '检查中'
    },
    isExecuting() {
      return ['pending', 'sent', 'executing'].includes(this.taskInfo.status)
    },
    // 重试按钮：存在任务且非已完成状态时显示（卡死/失败/超时均可重试）
    canRetry() {
      const taskId = this.taskInfo.taskId || this.taskId
      return !!taskId && this.taskInfo.status && this.taskInfo.status !== 'completed'
    }
  },
  watch: {
    visible(newVal) {
      if (newVal && this.client) {
        // 如果传入了 taskId，直接使用
        if (this.taskId) {
          this.taskInfo.taskId = this.taskId
        }
        this.startPolling()
      } else {
        this.stopPolling()
      }
    }
  },
  methods: {
    startPolling() {
      // 如果已经在轮询，先停止旧的
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
      }
      
      this.pollTimer = setInterval(async () => {
        await this.fetchTaskStatus()
      }, this.pollInterval)
      
      // 立即获取一次状态
      this.fetchTaskStatus()
    },
    
    stopPolling() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer)
        this.pollTimer = null
      }
    },
    
    async fetchTaskStatus() {
      // 如果没有 taskInfo.taskId，尝试从 props 获取
      let taskId = this.taskInfo.taskId || this.taskId
      
      if (!taskId) {
        console.warn('❌ [CheckTriggerDialog] 没有可用的任务 ID')
        return
      }
      
      try {
        const res = await getTaskStatus(taskId)
        
        this.taskInfo = {
          ...res,
          taskId: res.task_id,
          error_message: res.error_message
        }
        this.hasFetchedStatus = true
        
        // 任务完成、失败或超时时停止轮询
        if (['completed', 'failed', 'timeout'].includes(res.status)) {
          this.stopPolling()
          // ✅ 不自动关闭对话框，让用户查看结果后手动关闭
        }
      } catch (error) {
        console.error('❌ [CheckTriggerDialog] 获取任务状态失败:', error)
        // 不显示错误提示，静默处理
      }
    },
    
    refreshStatus() {
      this.fetchTaskStatus()
    },
    
    // 重试：删除当前卡死/失败的任务，重新创建新任务并继续轮询
    async handleRetry() {
      const oldTaskId = this.taskInfo.taskId || this.taskId
      if (!oldTaskId || !this.client || !this.client.client_uuid) {
        this.$message.error('缺少任务或客户端信息，无法重试')
        return
      }
      
      try {
        await this.$confirm(
          '将删除当前任务并重新创建新任务，是否继续？',
          '重试确认',
          { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
        )
      } catch (e) {
        return // 用户取消
      }
      
      this.retrying = true
      try {
        // 1. 删除当前任务（失败不阻断，可能是 404 已不存在）
        try {
          await deleteTask(oldTaskId)
        } catch (e) {
          console.warn('删除旧任务失败（可能已不存在）:', e)
        }
        
        // 2. 创建新任务
        const res = await triggerCheck(this.client.client_uuid)
        
        // 3. 切换到新任务并同步父组件的 taskId
        this.taskInfo = {
          taskId: res.task_id,
          status: res.status || 'pending'
        }
        this.hasFetchedStatus = true
        this.$emit('update:taskId', res.task_id)
        
        this.$message.success('新任务已创建')
        this.startPolling()
      } catch (error) {
        const msg = (error.response && error.response.data && error.response.data.error) || '重试失败'
        this.$message.error(msg)
      } finally {
        this.retrying = false
      }
    },
    
    closeDialog() {
      this.stopPolling()
      this.dialogVisible = false
    },
    
    handleClose() {
      this.stopPolling()
      this.dialogVisible = false
    }
  },
  beforeDestroy() {
    this.stopPolling()
  }
}
</script>

<style scoped lang="scss">
/* 🎨 弹窗容器 */
.check-trigger-modal {
  border-radius: var(--radius-xl);
  overflow: hidden;
  
  .el-dialog__header {
    padding: 24px 24px 16px;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
    border-bottom: none;
    
    .el-dialog__title {
      color: white !important;
      font-size: 18px;
      font-weight: 600;
    }
    
    .el-dialog__headerbtn .el-dialog__close {
      color: white;
      opacity: 0.9;
      font-size: 24px;
      
      &:hover {
        opacity: 1;
        transform: scale(1.1);
      }
    }
  }
  
  .el-dialog__body {
    padding: 24px;
  }
  
  .el-dialog__footer {
    padding: 16px 24px 24px;
    border-top: 1px solid var(--color-border-light);
    
    .el-button {
      min-width: 88px;
    }
  }
}

/* 🌟 状态指示器 */
.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 24px 16px;
  margin-bottom: 24px;
  border-radius: var(--radius-lg);
  transition: all var(--transition-base);
}

.indicator-icon {
  position: relative;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.indicator-icon i {
  font-size: 36px;
  z-index: 2;
}

/* 成功状态 - 绿色背景环 */
.status-indicator.success {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(16, 185, 129, 0.2) 100%);
}

.status-indicator.success .indicator-icon i {
  color: var(--color-primary);
  position: relative;
}

/* 成功状态的脉冲光环动画 */
.status-indicator.success .indicator-icon::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 3px solid var(--color-primary);
  animation: pulse-ring-success 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  z-index: 1;
}

/* 失败状态 - 红色背景环 */
.status-indicator.exception {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1) 0%, rgba(239, 68, 68, 0.2) 100%);
}

.status-indicator.exception .indicator-icon i {
  color: var(--color-danger);
}

/* 加载动画 */
.status-indicator.executing .indicator-icon i.el-icon-loading {
  animation: rotate 1.5s linear infinite;
  color: var(--color-info);
}

.indicator-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* ℹ️ 信息卡片 */
.info-card {
  background: white;
  border-radius: var(--radius-lg);
  padding: 20px;
  box-shadow: var(--shadow-sm);
  margin-bottom: 24px;
  border: 1px solid var(--color-border-light);
}

.info-row {
  display: grid;
  grid-template-columns: 88px 1fr;
  gap: 12px;
  align-items: center;
}

.info-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.info-label i {
  font-size: 16px;
  color: var(--color-primary);
}

.info-value {
  font-size: 14px;
  color: var(--color-text-primary);
  word-break: break-all;
  text-align: right;
}

.task-id {
  font-family: var(--font-mono);
  font-size: 13px;
}

.info-separator {
  height: 1px;
  background: var(--color-border-light);
  margin: 8px 0;
}

/* 📊 进度条区域 */
.progress-section {
  margin-bottom: 24px;
}

.status-description {
  margin-top: 12px;
  font-size: 13px;
  color: var(--color-text-secondary);
  text-align: center;
  line-height: 1.5;
}

/* ⚠️ 错误提示 */
.error-alert {
  margin-bottom: 16px;
  background: var(--color-danger-alpha-10);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-danger);
}

/* 🎭 过渡动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

@keyframes rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 成功状态的脉冲光环动画 */
@keyframes pulse-ring-success {
  0% {
    transform: translate(-50%, -50%) scale(0.9);
    opacity: 1;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.15);
    opacity: 0.6;
  }
  100% {
    transform: translate(-50%, -50%) scale(0.9);
    opacity: 1;
  }
}
</style>
