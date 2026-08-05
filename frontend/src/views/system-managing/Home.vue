<template>
  <div class="home-container">
    <!-- 🎨 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">系统看板</h1>
      <p class="page-subtitle">实时监控系统的运行状态和关键指标</p>
    </div>

    <!-- 📊 数据卡片网格 -->
    <div class="stats-grid">
      <!-- 卡片 1: 客户端数量 -->
      <el-card class="stat-card client-card" shadow="md" :body-style="{ padding: '24px' }">
        <div class="card-icon-wrapper">
          <div class="icon-circle client-icon">
            <i class="el-icon-monitor"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-label">Linux 客户端总数</div>
          <div class="card-value">16</div>
          <div class="card-trend">
            <span class="trend-label">在线：14</span>
            <span class="trend-label">离线：2</span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" @click="$router.push('/linux-hardening')">
            查看详情
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 2: 加固任务统计 -->
      <el-card class="stat-card task-card" shadow="md" :body-style="{ padding: '24px' }">
        <div class="card-icon-wrapper">
          <div class="icon-circle task-icon">
            <i class="el-icon-tickets"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-label">今日加固任务</div>
          <div class="card-value">48</div>
          <div class="card-trend">
            <span class="trend-success">
              <i class="el-icon-check"></i> 已完成：45
            </span>
            <span class="trend-waiting">
              <i class="el-icon-loading"></i> 进行中：3
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small">
            查看任务列表
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 3: 系统状态 -->
      <el-card class="stat-card system-card" shadow="md" :body-style="{ padding: '24px' }">
        <div class="card-icon-wrapper">
          <div class="icon-circle system-icon">
            <i class="el-icon-monitor"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-label">系统运行状态</div>
          <div class="card-value status-active">正常</div>
          <div class="card-trend">
            <span class="uptime">
              <i class="el-icon-time"></i> 运行时长：<strong>99.9%</strong>
            </span>
            <span class="last-check">
              <i class="el-icon-date"></i> 最后检查：刚刚
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="success" size="small" plain>
            健康诊断
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 4: 快捷入口 -->
      <el-card class="stat-card quick-card" shadow="md" :body-style="{ padding: '24px' }">
        <div class="card-icon-wrapper">
          <div class="icon-circle quick-icon">
            <i class="el-icon-link"></i>
          </div>
        </div>
        <div class="card-content">
          <div class="card-label">快速导航</div>
          <div class="quick-links">
            <el-button size="small" plain class="quick-btn">
              <i class="el-icon-s-check"></i> Linux 加固
            </el-button>
            <el-button size="small" plain class="quick-btn">
              <i class="el-icon-document"></i> 标准配置
            </el-button>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" plain @click="$router.push('/linux-hardening')">
            查看全部模块
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 🔧 功能按钮组 -->
    <div class="action-buttons">
      <el-button type="success" plain icon="el-icon-plus" size="medium">
        新增客户端
      </el-button>
      <el-button type="primary" icon="el-icon-refresh" size="medium">
        刷新数据
      </el-button>
      <el-button type="warning" plain icon="el-icon-download" size="medium">
        导出报表
      </el-button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Home',
  data() {
    return {
      // 模拟数据（实际项目应从 API 获取）
      stats: {
        clientCount: 16,
        onlineClientCount: 14,
        offlineClientCount: 2,
        todayTasks: 48,
        completedTasks: 45,
        runningTasks: 3,
        systemUptime: '99.9%'
      }
    }
  },
  mounted() {
    // TODO: 从后端 API 获取真实数据
    // this.fetchStatsData()
  },
  methods: {
    async fetchStatsData() {
      // 这里可以调用后端 API 获取实时数据
      console.log('Fetching statistics data...')
    }
  }
}
</script>

<style scoped>
.home-container {
  max-width: 1400px;
  margin: 0 auto;
}

/* 🎨 页面头部 */
.page-header {
  margin-bottom: var(--spacing-8);
}

.page-title {
  font-size: 32px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 8px 0;
  letter-spacing: -0.5px;
}

.page-subtitle {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
}

/* 📊 数据卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.stat-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-base);
  background: var(--color-bg-card);
}

.stat-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

/* 🟢 图标容器 */
.card-icon-wrapper {
  position: relative;
  width: 60px;
  height: 60px;
  margin-bottom: var(--spacing-4);
}

.icon-circle {
  width: 100%;
  height: 100%;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}

.client-icon i {
  color: var(--color-primary);
  opacity: 0.9;
}

.task-icon i {
  color: #F59E0B;
  opacity: 0.9;
}

.system-icon i {
  color: #10B981;
  opacity: 0.9;
}

.quick-icon i {
  color: #3B82F6;
  opacity: 0.9;
}

/* 🟢 卡片内容 */
.card-content {
  margin-bottom: var(--spacing-4);
}

.card-label {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-3);
  font-weight: 500;
}

.card-value {
  font-size: 36px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
  margin-bottom: var(--spacing-3);
}

.card-value.status-active {
  color: var(--color-success);
  font-size: 28px;
}

/* 🟢 趋势信息 */
.card-trend {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  font-size: 13px;
}

.trend-label {
  color: var(--color-text-secondary);
}

.trend-success {
  color: var(--color-success);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-waiting {
  color: var(--color-warning);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.uptime,
.last-check {
  color: var(--color-text-secondary);
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.uptime strong {
  color: var(--color-text-primary);
  font-weight: 600;
}

/* 🟢 卡片底部 */
.card-footer {
  border-top: 1px solid var(--color-border-light);
  padding-top: var(--spacing-4);
  margin-top: var(--spacing-4);
}

/* 强制覆盖卡片底部按钮样式 */
.card-footer :deep(.el-button) {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  margin: 0 !important;
}

/* 🟢 快捷链接 */
.quick-links {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  width: 100%;
}

/* 强制覆盖 Element UI 按钮默认样式 */
.quick-links :deep(.el-button) {
  width: 100% !important;
  display: flex !important;
  align-items: center !important;
  justify-content: flex-start !important;
  padding: 10px 16px !important;
  margin: 0 !important;
  border: 1px solid var(--color-border) !important;
  border-radius: var(--radius-md) !important;
  text-align: left !important;
  min-height: 40px !important;
  background: white !important;
  color: var(--color-text-regular) !important;
  transition: all var(--transition-base) !important;
  box-sizing: border-box !important;
}

.quick-links :deep(.el-button:hover) {
  border-color: var(--color-primary) !important;
  background: var(--color-primary-alpha-10) !important;
  color: var(--color-primary) !important;
}

.quick-links :deep(.el-button i) {
  font-size: 16px !important;
  min-width: 16px !important;
  margin-right: 8px !important;
}

.quick-links :deep(.el-button span) {
  flex: 1 !important;
}

/* 🔧 操作按钮组 */
.action-buttons {
  display: flex;
  gap: var(--spacing-4);
  justify-content: center;
  flex-wrap: wrap;
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-4);
  }
  
  .card-value {
    font-size: 28px;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .action-buttons .el-button {
    width: 100%;
  }
}
</style>
