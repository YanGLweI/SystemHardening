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
              <!-- Linux Terminal Icon -->
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
                <path d="M7 10L9.5 12.5L7 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M16 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
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
              <!-- Windows Window Icon -->
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M4 4H11V11H4V4Z" fill="currentColor"/>
                <path d="M13 4H20V11H13V4Z" fill="currentColor" opacity="0.9"/>
                <path d="M4 13H11V20H4V13Z" fill="currentColor" opacity="0.8"/>
                <path d="M13 13H20V20H13V13Z" fill="currentColor" opacity="0.7"/>
              </svg>
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
              <i class="el-icon-position"></i> 已划分区域数
            </span>
            <span class="trend-label">
              <i class="el-icon-location"></i> 当前位于IT区域
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
          <div ref="onlineChart" class="echart-box online-chart-box"></div>
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

      <!-- 各区域合规数条形图 -->
      <el-card class="chart-card chart-card-wide animate-item" :class="{ 'is-loaded': loaded }" :body-style="{ padding: '24px' }" :style="{ animationDelay: '480ms' }">
        <div slot="header" class="chart-card-header">
          <span class="chart-title">各区域合规数</span>
        </div>
        <div class="region-chart-wrapper">
          <div ref="regionChart" class="echart-box region-chart-box"></div>
          <div v-if="!hasRegionData" class="chart-empty">暂无区域数据</div>
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
import * as echarts from 'echarts/core'
import { PieChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([TitleComponent, TooltipComponent, LegendComponent, GridComponent, PieChart, BarChart, CanvasRenderer])

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
        region_count: 0,
        region_compliance: []
      },
      // ECharts 实例
      onlineChart: null,
      regionChart: null,
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
    // 是否存在区域合规数据
    hasRegionData() {
      return Array.isArray(this.stats.region_compliance) && this.stats.region_compliance.length > 0
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.initCharts()
    })
    this.fetchDashboardData()
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
    if (this.onlineChart) {
      this.onlineChart.dispose()
      this.onlineChart = null
    }
    if (this.regionChart) {
      this.regionChart.dispose()
      this.regionChart = null
    }
  },
  methods: {
    async fetchDashboardData() {
      this.loading = true
      try {
        const res = await getDashboardStats()
        if (res && res.data) {
          this.stats = res.data
          this.animateCounters()
          this.renderCharts()
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
    },
    // 初始化 ECharts 实例
    initCharts() {
      if (this.$refs.onlineChart) {
        this.onlineChart = echarts.init(this.$refs.onlineChart)
      }
      if (this.$refs.regionChart) {
        this.regionChart = echarts.init(this.$refs.regionChart)
      }
      window.addEventListener('resize', this.handleResize)
      this.renderCharts()
    },
    // 渲染全部图表
    renderCharts() {
      this.renderOnlineChart()
      this.renderRegionChart()
    },
    // 客户端在线状态环形图
    renderOnlineChart() {
      if (!this.onlineChart) return
      this.onlineChart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: '{b}: {c} ({d}%)'
        },
        title: {
          text: String(this.stats.total_clients),
          subtext: '总数',
          left: 'center',
          top: '34%',
          textStyle: { fontSize: 22, fontWeight: 700, color: '#111827' },
          subtextStyle: { fontSize: 12, color: '#6B7280' }
        },
        series: [{
          type: 'pie',
          radius: ['60%', '82%'],
          label: { show: false },
          labelLine: { show: false },
          data: [
            { name: '在线', value: this.stats.online_clients, itemStyle: { color: '#10B981' } },
            { name: '离线', value: this.stats.offline_clients, itemStyle: { color: '#E5E7EB' } }
          ]
        }]
      })
    },
    // 各区域合规数条形图
    renderRegionChart() {
      if (!this.regionChart) return
      const list = this.stats.region_compliance || []
      if (list.length === 0) return
      this.regionChart.setOption({
        tooltip: {
          trigger: 'axis',
          axisPointer: { type: 'shadow' }
        },
        legend: {
          data: ['合规', '不合规'],
          top: 0,
          textStyle: { color: '#6B7280' }
        },
        grid: { left: 40, right: 16, top: 36, bottom: 28 },
        xAxis: {
          type: 'category',
          data: list.map(item => item.region_name),
          axisTick: { alignWithLabel: true },
          axisLabel: {
            interval: 0,
            rotate: list.length > 6 ? 30 : 0,
            color: '#6B7280'
          }
        },
        yAxis: {
          type: 'value',
          minInterval: 1,
          axisLabel: { color: '#6B7280' },
          splitLine: { lineStyle: { color: '#F3F4F6' } }
        },
        series: [
          {
            name: '合规',
            type: 'bar',
            barGap: '10%',
            barMaxWidth: 28,
            itemStyle: { color: '#93C5FD', borderRadius: [3, 3, 0, 0] },
            data: list.map(item => item.compliant_count)
          },
          {
            name: '不合规',
            type: 'bar',
            barMaxWidth: 28,
            itemStyle: { color: '#FCA5A5', borderRadius: [3, 3, 0, 0] },
            data: list.map(item => item.non_compliant_count)
          }
        ]
      })
      this.regionChart.resize()
    },
    // 窗口尺寸变化时自适应
    handleResize() {
      if (this.onlineChart) this.onlineChart.resize()
      if (this.regionChart) this.regionChart.resize()
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
.linux-icon svg {
  color: var(--color-primary);
}

.windows-icon {
  background: rgba(16, 185, 129, 0.1);
}
.windows-icon svg {
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

/* 图表区域：4 等分网格与上排统计卡对齐，条形图卡横跨后 3 列（1:3） */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.chart-card-wide {
  grid-column: 2 / -1;
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
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-4);
  padding: var(--spacing-4) 0;
}

/* ECharts 图表容器 */
.echart-box {
  width: 100%;
}

.online-chart-box {
  width: 200px;
  height: 200px;
  flex-shrink: 0;
}

.region-chart-wrapper {
  position: relative;
  padding: var(--spacing-2) 0;
}

.region-chart-box {
  height: 280px;
}

.chart-empty {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-bg-card);
}

/* 图例 */
.chart-legend {
  display: flex;
  flex-direction: row;
  justify-content: center;
  gap: var(--spacing-6);
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

  .chart-card-wide {
    grid-column: auto;
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
