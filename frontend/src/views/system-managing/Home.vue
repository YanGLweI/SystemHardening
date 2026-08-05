<template>
  <div class="home-container">
    <!-- 页面标题 -->
    <div class="page-header animate-item" :style="{ animationDelay: '0ms' }">
      <h1 class="page-title">系统看板</h1>
      <p class="page-subtitle">实时监控系统的运行状态和关键指标</p>
    </div>

    <!-- 数据卡片网格 -->
    <div class="stats-grid">
      <!-- 卡片 1: 客户端总数 -->
      <el-card class="stat-card client-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '80ms' }">
        <div class="card-header-row">
          <div class="card-icon-wrapper">
            <div class="icon-circle client-icon">
              <i class="el-icon-monitor"></i>
            </div>
          </div>
          <el-tag size="mini" type="success" effect="plain" v-if="stats.online_clients > 0">
            {{ stats.online_clients }} 在线
          </el-tag>
        </div>
        <div class="card-content">
          <div class="card-label">客户端总数</div>
          <div class="card-value">
            <span class="counter-value">{{ animatedTotalClients }}</span>
          </div>
          <div class="card-trend">
            <span class="trend-online">
              <i class="el-icon-success"></i> 在线：{{ stats.online_clients }}
            </span>
            <span class="trend-offline">
              <i class="el-icon-remove"></i> 离线：{{ stats.offline_clients }}
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" @click="$router.push('/client-management')">
            查看详情
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 2: Linux 加固 -->
      <el-card class="stat-card linux-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '160ms' }">
        <div class="card-header-row">
          <div class="card-icon-wrapper">
            <div class="icon-circle linux-icon">
              <i class="el-icon-s-platform"></i>
            </div>
          </div>
          <el-tag size="mini" :type="linuxComplianceRate >= 80 ? 'success' : 'warning'" effect="plain">
            {{ linuxComplianceRate }}%
          </el-tag>
        </div>
        <div class="card-content">
          <div class="card-label">Linux 加固主机</div>
          <div class="card-value">
            <span class="counter-value">{{ animatedLinuxHosts }}</span>
          </div>
          <div class="card-trend">
            <span class="trend-success">
              <i class="el-icon-check"></i> 合规：{{ stats.linux_compliant_count }}
            </span>
            <span class="trend-danger">
              <i class="el-icon-close"></i> 不合规：{{ stats.linux_non_compliant_count }}
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" @click="$router.push('/check/linux')">
            查看详情
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 3: Windows 加固 -->
      <el-card class="stat-card windows-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '240ms' }">
        <div class="card-header-row">
          <div class="card-icon-wrapper">
            <div class="icon-circle windows-icon">
              <i class="el-icon-monitor"></i>
            </div>
          </div>
          <el-tag size="mini" :type="windowsComplianceRate >= 80 ? 'success' : 'warning'" effect="plain">
            {{ windowsComplianceRate }}%
          </el-tag>
        </div>
        <div class="card-content">
          <div class="card-label">Windows 加固主机</div>
          <div class="card-value">
            <span class="counter-value">{{ animatedWindowsHosts }}</span>
          </div>
          <div class="card-trend">
            <span class="trend-success">
              <i class="el-icon-check"></i> 合规：{{ stats.windows_compliant_count }}
            </span>
            <span class="trend-danger">
              <i class="el-icon-close"></i> 不合规：{{ stats.windows_non_compliant_count }}
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" @click="$router.push('/check/windows')">
            查看详情
          </el-button>
        </div>
      </el-card>

      <!-- 卡片 4: 区域管理 -->
      <el-card class="stat-card region-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '320ms' }">
        <div class="card-header-row">
          <div class="card-icon-wrapper">
            <div class="icon-circle region-icon">
              <i class="el-icon-s-grid"></i>
            </div>
          </div>
        </div>
        <div class="card-content">
          <div class="card-label">管理区域</div>
          <div class="card-value">
            <span class="counter-value">{{ animatedRegions }}</span>
          </div>
          <div class="card-trend">
            <span class="trend-label">
              <i class="el-icon-location"></i> 已划分区域数
            </span>
          </div>
        </div>
        <div class="card-footer">
          <el-button type="primary" size="small" @click="$router.push('/region-management')">
            查看详情
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <!-- 客户端在线状态环形图 -->
      <el-card class="chart-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '400ms' }">
        <div slot="header" class="chart-card-header">
          <span class="chart-title">客户端在线状态</span>
        </div>
        <div class="chart-body">
          <div class="donut-chart-wrapper">
            <svg class="donut-chart" viewBox="0 0 120 120">
              <!-- 背景圆环 -->
              <circle class="donut-bg" cx="60" cy="60" r="50" fill="none" stroke="#F3F4F6" stroke-width="14" />
              
              <!-- 在线部分 -->
              <circle
                class="donut-segment online-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#10B981"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="onlineLength"
                stroke-dashoffset="0"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
              
              <!-- 离线部分 - 使用白色填充覆盖背景以消除重叠 -->
              <circle
                v-if="offlineLength > 0"
                class="donut-segment offline-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#E5E7EB"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="offlineLength"
                :stroke-dashoffset="-onlineLength"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
            </svg>
            <div class="donut-center-label">
              <span class="donut-value">{{ stats.total_clients }}</span>
              <span class="donut-label">总数</span>
            </div>
          </div>
          <div class="chart-legend">
            <div class="legend-item">
              <span class="legend-dot online-dot"></span>
              <span class="legend-text">在线</span>
              <span class="legend-value">{{ stats.online_clients }}</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot offline-dot"></span>
              <span class="legend-text">离线</span>
              <span class="legend-value">{{ stats.offline_clients }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- Linux 合规率环形图 -->
      <el-card class="chart-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '480ms' }">
        <div slot="header" class="chart-card-header">
          <span class="chart-title">Linux 合规率</span>
        </div>
        <div class="chart-body">
          <div class="donut-chart-wrapper">
            <svg class="donut-chart" viewBox="0 0 120 120">
              <!-- 背景圆环 -->
              <circle class="donut-bg" cx="60" cy="60" r="50" fill="none" stroke="#F3F4F6" stroke-width="14" />
              
              <!-- 合规部分 -->
              <circle
                class="donut-segment compliant-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#10B981"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="linuxCompliantLength"
                stroke-dashoffset="0"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
              
              <!-- 不合规部分 - 仅在有不合规数据时显示 -->
              <circle
                v-if="linuxNonCompliantLength > 0"
                class="donut-segment noncompliant-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#EF4444"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="linuxNonCompliantLength"
                :stroke-dashoffset="-linuxCompliantLength"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
            </svg>
            <div class="donut-center-label">
              <span class="donut-value">{{ linuxComplianceRate }}<small>%</small></span>
              <span class="donut-label">合规率</span>
            </div>
          </div>
          <div class="chart-legend">
            <div class="legend-item">
              <span class="legend-dot linux-compliant-dot"></span>
              <span class="legend-text">合规</span>
              <span class="legend-value">{{ stats.linux_compliant_count }}</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot linux-noncompliant-dot"></span>
              <span class="legend-text">不合规</span>
              <span class="legend-value">{{ stats.linux_non_compliant_count }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- Windows 合规率环形图 -->
      <el-card class="chart-card animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '560ms' }">
        <div slot="header" class="chart-card-header">
          <span class="chart-title">Windows 合规率</span>
        </div>
        <div class="chart-body">
          <div class="donut-chart-wrapper">
            <svg class="donut-chart" viewBox="0 0 120 120">
              <!-- 背景圆环 -->
              <circle class="donut-bg" cx="60" cy="60" r="50" fill="none" stroke="#F3F4F6" stroke-width="14" />
              
              <!-- 合规部分 -->
              <circle
                class="donut-segment win-compliant-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#3B82F6"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="windowsCompliantLength"
                stroke-dashoffset="0"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
              
              <!-- 不合规部分 - 仅在有不合规数据时显示 -->
              <circle
                v-if="windowsNonCompliantLength > 0"
                class="donut-segment win-noncompliant-segment"
                cx="60" cy="60" r="50"
                fill="none"
                stroke="#F59E0B"
                stroke-width="14"
                stroke-linecap="round"
                :stroke-dasharray="windowsNonCompliantLength"
                :stroke-dashoffset="-windowsCompliantLength"
                transform="rotate(-90 60 60)"
                :class="{ 'animate-in': loaded }"
              />
            </svg>
            <div class="donut-center-label">
              <span class="donut-value">{{ windowsComplianceRate }}<small>%</small></span>
              <span class="donut-label">合规率</span>
            </div>
          </div>
          <div class="chart-legend">
            <div class="legend-item">
              <span class="legend-dot win-compliant-dot"></span>
              <span class="legend-text">合规</span>
              <span class="legend-value">{{ stats.windows_compliant_count }}</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot win-noncompliant-dot"></span>
              <span class="legend-text">不合规</span>
              <span class="legend-value">{{ stats.windows_non_compliant_count }}</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 快捷操作 -->
    <div class="action-section animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '640ms' }">
      <div class="action-title">快捷操作</div>
      <div class="action-buttons">
        <el-button type="success" plain icon="el-icon-plus" size="medium" @click="$router.push('/client-management')">
          新增客户端
        </el-button>
        <el-button type="primary" icon="el-icon-refresh" size="medium" :loading="loading" @click="fetchDashboardData">
          刷新数据
        </el-button>
        <el-button type="info" plain icon="el-icon-s-check" size="medium" @click="$router.push('/check/linux')">
          Linux 加固检查
        </el-button>
        <el-button type="warning" plain icon="el-icon-s-check" size="medium" @click="$router.push('/check/windows')">
          Windows 加固检查
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { getDashboardStats } from '@/api/dashboard'

export default {
  name: 'Home',
  data() {
    return {
      loading: false,
      loaded: false,
      stats: {
        total_clients: 0,
        online_clients: 0,
        offline_clients: 0,
        linux_host_count: 0,
        linux_compliant_count: 0,
        linux_non_compliant_count: 0,
        windows_host_count: 0,
        windows_compliant_count: 0,
        windows_non_compliant_count: 0,
        region_count: 0
      },
      // 数字动画
      animatedTotalClients: 0,
      animatedLinuxHosts: 0,
      animatedWindowsHosts: 0,
      animatedRegions: 0
    }
  },
  computed: {
    // Linux 合规率
    linuxComplianceRate() {
      const total = this.stats.linux_host_count
      if (total === 0) return 0
      return Math.round((this.stats.linux_compliant_count / total) * 100)
    },
    // Windows 合规率
    windowsComplianceRate() {
      const total = this.stats.windows_host_count
      if (total === 0) return 0
      return Math.round((this.stats.windows_compliant_count / total) * 100)
    },
    // SVG 环形图参数（周长 = 2 * PI * 50 ≈ 314.16）
    circumference() {
      return 2 * Math.PI * 50
    },
    // 客户端在线状态环形图
    clientOnlinePercent() {
      if (this.stats.total_clients === 0) return 0
      return (this.stats.online_clients / this.stats.total_clients)
    },
    onlineLength() {
      return this.circumference * this.clientOnlinePercent
    },
    offlineLength() {
      return this.circumference - this.onlineLength
    },
    // Linux 合规环形图
    linuxCompliantPercent() {
      const total = this.stats.linux_host_count
      if (total === 0) return 0
      return this.stats.linux_compliant_count / total
    },
    linuxCompliantLength() {
      return this.circumference * this.linuxCompliantPercent
    },
    linuxNonCompliantLength() {
      return this.circumference - this.linuxCompliantLength
    },
    // Windows 合规环形图
    windowsCompliantPercent() {
      const total = this.stats.windows_host_count
      if (total === 0) return 0
      return this.stats.windows_compliant_count / total
    },
    windowsCompliantLength() {
      return this.circumference * this.windowsCompliantPercent
    },
    windowsNonCompliantLength() {
      return this.circumference - this.windowsCompliantLength
    },
  },
  mounted() {
    this.fetchDashboardData()
  },
  methods: {
    async fetchDashboardData() {
      this.loading = true
      try {
        const res = await getDashboardStats()
        if (res && res.data) {
          this.stats = res.data
          this.animateCounters()
        }
      } catch (err) {
        console.error('Failed to fetch dashboard stats:', err)
      } finally {
        this.loading = false
        // 触发入场动画
        this.$nextTick(() => {
          this.loaded = true
        })
      }
    },
    // 数字递增动画
    animateCounters() {
      this.animateNumber('animatedTotalClients', this.stats.total_clients, 600)
      this.animateNumber('animatedLinuxHosts', this.stats.linux_host_count, 700)
      this.animateNumber('animatedWindowsHosts', this.stats.windows_host_count, 800)
      this.animateNumber('animatedRegions', this.stats.region_count, 500)
    },
    animateNumber(key, target, duration) {
      const start = this[key]
      const diff = target - start
      if (diff === 0) return
      const startTime = performance.now()
      const step = (currentTime) => {
        const elapsed = currentTime - startTime
        const progress = Math.min(elapsed / duration, 1)
        // easeOutCubic 缓动
        const eased = 1 - Math.pow(1 - progress, 3)
        this[key] = Math.round(start + diff * eased)
        if (progress < 1) {
          requestAnimationFrame(step)
        }
      }
      requestAnimationFrame(step)
    }
  }
}
</script>

<style scoped>
.home-container {
  max-width: 1400px;
  margin: 0 auto;
}

/* 页面头部 */
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

/* 入场动画 */
.animate-item {
  opacity: 0;
  transform: translateY(24px);
  animation: fadeSlideUp 0.5s ease-out forwards;
}

@keyframes fadeSlideUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 数据卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
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

/* 卡片头部行 */
.card-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

/* 图标容器 */
.card-icon-wrapper {
  width: 52px;
  height: 52px;
}

.icon-circle {
  width: 100%;
  height: 100%;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.client-icon {
  background: rgba(16, 185, 129, 0.1);
}
.client-icon i { 
  color: var(--color-primary); 
}

.linux-icon {
  background: rgba(16, 185, 129, 0.1);
}
.linux-icon i { 
  color: var(--color-primary); 
}

.windows-icon {
  background: rgba(16, 185, 129, 0.1);
}
.windows-icon i { 
  color: var(--color-primary); 
}

.region-icon {
  background: rgba(245, 158, 11, 0.1);
}
.region-icon i { 
  color: var(--color-warning); 
}

/* 卡片内容 */
.card-content {
  margin-bottom: var(--spacing-4);
}

.card-label {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-2);
  font-weight: 500;
}

.card-value {
  font-size: 36px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
  margin-bottom: var(--spacing-3);
}

.counter-value {
  display: inline-block;
  font-variant-numeric: tabular-nums;
}

/* 趋势信息 */
.card-trend {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
  font-size: 13px;
}

.trend-online {
  color: var(--color-success);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-offline {
  color: var(--color-text-secondary);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-success {
  color: var(--color-success);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-danger {
  color: var(--color-danger);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
}

.trend-label {
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 卡片底部 */
.card-footer {
  border-top: 1px solid var(--color-border-light);
  padding-top: var(--spacing-4);
  margin-top: var(--spacing-4);
}

.card-footer :deep(.el-button) {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  margin: 0 !important;
}

/* 图表区域 */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.chart-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  background: var(--color-bg-card);
}

.chart-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chart-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.chart-body {
  display: flex;
  align-items: center;
  gap: var(--spacing-8);
  padding: var(--spacing-4) 0;
}

/* SVG 环形图 */
.donut-chart-wrapper {
  position: relative;
  width: 140px;
  height: 140px;
  flex-shrink: 0;
}

.donut-chart {
  width: 100%;
  height: 100%;
}

.donut-segment {
  transition: stroke-dasharray 0.8s ease-out;
}

.donut-center-label {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.donut-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
}

.donut-value small {
  font-size: 14px;
  font-weight: 600;
}

.donut-label {
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* 图例 */
.chart-legend {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  flex: 1;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: 13px;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.online-dot { background: #10B981; }
.offline-dot { background: #E5E7EB; }
.linux-compliant-dot { background: #10B981; }
.linux-noncompliant-dot { background: #EF4444; }
.win-compliant-dot { background: #3B82F6; }
.win-noncompliant-dot { background: #F59E0B; }

.legend-text {
  color: var(--color-text-secondary);
  flex: 1;
}

.legend-value {
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

/* 快捷操作 */
.action-section {
  text-align: center;
}

.action-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-4);
}

.action-buttons {
  display: flex;
  gap: var(--spacing-4);
  justify-content: center;
  flex-wrap: wrap;
}

/* 响应式设计 */
@media screen and (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-4);
  }

  .charts-grid {
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

  .chart-body {
    flex-direction: column;
    align-items: center;
  }
}
</style>
