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
      <el-card class="stat-card client-card animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '80ms' }">
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
        <div class="card-water-zone">
          <div class="wave-bg" aria-hidden="true" :class="{ 'is-filling': filling }">
            <div class="wave-water" :style="{ transform: waterTransform(waterLevels.client) }">
              <div class="wave wave-1"></div>
              <div class="wave wave-4"></div>
              <div class="wave wave-2"></div>
              <div v-if="boatCard === 'client'" class="boat" aria-hidden="true">
                <svg viewBox="0 0 40 30" width="40" height="30" xmlns="http://www.w3.org/2000/svg">
                  <rect x="19.5" y="1" width="1" height="18" fill="#35507A"/>
                  <path d="M19 3 L19 18 L8 18 Z" fill="#FFFFFF"/>
                  <path d="M21 6 L21 18 L31 18 Z" fill="#BFDBFE"/>
                  <path d="M4 20 L36 20 L30 28 L10 28 Z" fill="#35507A"/>
                </svg>
              </div>
              <div class="wave wave-3"></div>
              <span v-for="(b, i) in clientBubbles" :key="i" class="bubble" :style="b"></span>
            </div>
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
        </div>
      </el-card>

      <!-- 卡片 2: Linux 加固 -->
      <el-card class="stat-card linux-card animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '160ms' }">
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
        <div class="card-water-zone">
          <div class="wave-bg" aria-hidden="true" :class="{ 'is-filling': filling }">
            <div class="wave-water" :style="{ transform: waterTransform(waterLevels.linux) }">
              <div class="wave wave-1"></div>
              <div class="wave wave-4"></div>
              <div class="wave wave-2"></div>
              <div v-if="boatCard === 'linux'" class="boat" aria-hidden="true">
                <svg viewBox="0 0 40 30" width="40" height="30" xmlns="http://www.w3.org/2000/svg">
                  <rect x="19.5" y="1" width="1" height="18" fill="#35507A"/>
                  <path d="M19 3 L19 18 L8 18 Z" fill="#FFFFFF"/>
                  <path d="M21 6 L21 18 L31 18 Z" fill="#BFDBFE"/>
                  <path d="M4 20 L36 20 L30 28 L10 28 Z" fill="#35507A"/>
                </svg>
              </div>
              <div class="wave wave-3"></div>
              <span v-for="(b, i) in linuxBubbles" :key="i" class="bubble" :style="b"></span>
            </div>
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
        </div>
      </el-card>

      <!-- 卡片 3: Windows 加固 -->
      <el-card class="stat-card windows-card animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '240ms' }">
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
        <div class="card-water-zone">
          <div class="wave-bg" aria-hidden="true" :class="{ 'is-filling': filling }">
            <div class="wave-water" :style="{ transform: waterTransform(waterLevels.windows) }">
              <div class="wave wave-1"></div>
              <div class="wave wave-4"></div>
              <div class="wave wave-2"></div>
              <div v-if="boatCard === 'windows'" class="boat" aria-hidden="true">
                <svg viewBox="0 0 40 30" width="40" height="30" xmlns="http://www.w3.org/2000/svg">
                  <rect x="19.5" y="1" width="1" height="18" fill="#35507A"/>
                  <path d="M19 3 L19 18 L8 18 Z" fill="#FFFFFF"/>
                  <path d="M21 6 L21 18 L31 18 Z" fill="#BFDBFE"/>
                  <path d="M4 20 L36 20 L30 28 L10 28 Z" fill="#35507A"/>
                </svg>
              </div>
              <div class="wave wave-3"></div>
              <span v-for="(b, i) in windowsBubbles" :key="i" class="bubble" :style="b"></span>
            </div>
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
        </div>
      </el-card>

      <!-- 卡片 4: 区域管理 -->
      <el-card class="stat-card region-card animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '320ms' }">
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
      <!-- 最近新增客户端列表 -->
      <el-card class="chart-card animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '400ms' }">
        <div slot="header" class="chart-card-header">
          <span class="chart-title">最近新增客户端</span>
        </div>
        <div class="recent-client-body">
          <ul v-if="stats.recent_clients && stats.recent_clients.length" class="recent-client-list">
            <li v-for="(cli, i) in stats.recent_clients" :key="i" class="recent-client-item">
              <span class="recent-client-dot" aria-hidden="true"></span>
              <div class="recent-client-info">
                <span class="recent-client-name" :title="cli.device_name">{{ cli.device_name }}</span>
                <span class="recent-client-meta">{{ cli.ip_address }}{{ cli.os_version ? ' · ' + cli.os_version : '' }}</span>
              </div>
              <span class="recent-client-time">{{ formatShortTime(cli.created_at) }}</span>
            </li>
          </ul>
          <div v-else class="recent-client-empty">暂无客户端</div>
        </div>
      </el-card>

      <!-- 各区域合规数条形图 -->
      <el-card class="chart-card chart-card-wide animate-item" :class="{ 'is-loaded': loaded }" :style="{ animationDelay: '480ms' }">
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
        <el-button type="success" plain icon="el-icon-plus" size="small" @click="$router.push('/client-management')">
          新增客户端
        </el-button>
        <el-button type="primary" icon="el-icon-refresh" size="small" :loading="loading" @click="fetchDashboardData">
          刷新数据
        </el-button>
        <el-button type="info" plain icon="el-icon-s-check" size="small" @click="$router.push('/check/linux')">
          Linux 加固检查
        </el-button>
        <el-button type="warning" plain icon="el-icon-s-check" size="small" @click="$router.push('/check/windows')">
          Windows 加固检查
        </el-button>
      </div>
    </div>
  </div>
</template>

<script>
import { getDashboardStats } from '@/api/dashboard'
import * as echarts from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([TitleComponent, TooltipComponent, LegendComponent, GridComponent, BarChart, CanvasRenderer])

export default {
  name: 'Home',
  data() {
    return {
      loading: false,
      loaded: false,
      filling: false,
      // 灌水式水位：弹簧动画逐帧驱动的当前水位（0-100+过冲）
      waterLevels: { client: 0, linux: 0, windows: 0 },
      // 每次刷新随机选一张水卡显示小船
      boatCard: ['client', 'linux', 'windows'][Math.floor(Math.random() * 3)],
      fillRaf: 0,
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
        region_compliance: [],
        recent_clients: []
      },
      // ECharts 实例
      regionChart: null,
      // 数字动画
      animatedTotalClients: 0,
      animatedLinuxHosts: 0,
      animatedWindowsHosts: 0,
      animatedRegions: 0,
      // 水波气泡配置（created 时随机生成一次）
      clientBubbles: [],
      linuxBubbles: [],
      windowsBubbles: []
    }
  },
  computed: {
    // 客户端在线率（卡片 1 水位比例）
    clientOnlineRate() {
      const total = this.stats.total_clients
      if (total === 0) return 0
      return Math.round((this.stats.online_clients / total) * 100)
    },
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
  created() {
    this.clientBubbles = this.makeBubbles(6)
    this.linuxBubbles = this.makeBubbles(6)
    this.windowsBubbles = this.makeBubbles(6)
  },
  mounted() {
    this.$nextTick(() => {
      this.initCharts()
    })
    this.fetchDashboardData()
  },
  beforeDestroy() {
    cancelAnimationFrame(this.fillRaf)
    window.removeEventListener('resize', this.handleResize)
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
          // 数据就绪后触发灌水式水位上升
          this.$nextTick(() => {
            this.startFilling()
          })
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
    // 水位 transform：以浪花顶部为基线（后浪高 18px）；
    // 80% 以上水位线性预留最多 5% 头隙，弹簧过冲在头隙内起伏、不越顶裁切
    waterTransform(level) {
      const r = Math.min(level || 0, 105)
      const headroom = Math.max(0, r - 80) * 0.25
      return 'translateY(calc(' + (100 - r + headroom) + '% + ' + (r * 0.18).toFixed(2) + 'px))'
    },
    // 灌水式水位上升：欠阻尼弹簧阶跃响应（轻微过冲 + 晃荡回落），三卡错峰注入；
    // 落定后进入永久海面涌浪阶段（双正弦小幅起伏），模拟海面高低起伏
    startFilling() {
      const keys = ['client', 'linux', 'windows']
      const targets = [this.clientOnlineRate, this.linuxComplianceRate, this.windowsComplianceRate]
      const starts = keys.map(key => this.waterLevels[key])
      cancelAnimationFrame(this.fillRaf)
      // 减少动态偏好：直接落到目标值，无弹簧与泡沫层
      if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
        this.waterLevels = { client: targets[0], linux: targets[1], windows: targets[2] }
        return
      }
      const delays = [0, 150, 300]
      const T = 2200
      // 欠阻尼弹簧参数：ζ=0.7、ωn=3.0 → 过冲≈4.6%
      const zeta = 0.7
      const wn = 3.0
      const lambda = zeta * wn
      const omega = wn * Math.sqrt(1 - zeta * zeta)
      const total = T + delays[2]
      const startTime = performance.now()
      this.filling = true
      const step = (now) => {
        const elapsed = now - startTime
        const levels = {}
        if (elapsed < total) {
          // 灌水阶段：弹簧上升
          keys.forEach((key, i) => {
            const t = Math.max(elapsed - delays[i], 0) / 1000
            if (t <= 0) {
              levels[key] = starts[i]
            } else {
              // 零初速度欠阻尼阶跃响应：p(t) = 1 - e^(-λt)(cos ωt + (λ/ω) sin ωt)
              const p = 1 - Math.exp(-lambda * t) * (Math.cos(omega * t) + (lambda / omega) * Math.sin(omega * t))
              // 水位钳制 ≤105：5% 头隙容纳过冲（≈4.6%），到顶仍可见弹簧回弹且不裁切
              levels[key] = Math.min(starts[i] + (targets[i] - starts[i]) * p, 105)
            }
          })
        } else {
          // 落定后：永久海面涌浪（2+1 双正弦、峰值 3%），1.5s 内振幅淡入避免跳变；
          // 头隙已容纳 ≤105 水位，涌浪可对称起伏不裁切
          if (this.filling) this.filling = false
          const ts = elapsed - total
          const fade = Math.min(ts / 1500, 1)
          keys.forEach((key, i) => {
            const osc = 2 * Math.sin((2 * Math.PI * ts) / 4600 + i * 1.7) + Math.sin((2 * Math.PI * ts) / 7300 + i * 2.3)
            levels[key] = Math.min(targets[i] + fade * osc, 105)
          })
        }
        this.waterLevels = levels
        this.fillRaf = requestAnimationFrame(step)
      }
      this.fillRaf = requestAnimationFrame(step)
    },
    // 生成水泡动画参数
    makeBubbles(count) {
      const bubbles = []
      for (let i = 0; i < count; i++) {
        const size = 4 + Math.round(Math.random() * 6)
        bubbles.push({
          left: (6 + Math.random() * 88).toFixed(1) + '%',
          width: size + 'px',
          '--s': size + 'px',
          '--d': (5 + Math.random() * 6).toFixed(1) + 's',
          '--delay': (Math.random() * 8).toFixed(1) + 's'
        })
      }
      return bubbles
    },
    // 初始化 ECharts 实例
    initCharts() {
      if (this.$refs.regionChart) {
        this.regionChart = echarts.init(this.$refs.regionChart)
      }
      window.addEventListener('resize', this.handleResize)
      this.renderCharts()
    },
    // 渲染全部图表
    renderCharts() {
      this.renderRegionChart()
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
            rotate: 0,
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
      if (this.regionChart) this.regionChart.resize()
    },
    // 最近新增列表的短时间格式：MM-DD HH:mm
    formatShortTime(t) {
      if (!t) return ''
      const d = new Date(t)
      if (isNaN(d.getTime())) return ''
      const p = n => String(n).padStart(2, '0')
      return p(d.getMonth() + 1) + '-' + p(d.getDate()) + ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
    }
  }
}
</script>

<style scoped>
/* 整页纵向 Flex：占满主内容区高度，头部/统计卡/快捷操作定高，图表区弹性吃掉剩余高度；
   min-height 兜底，极矮窗口回落到外层滚动 */
.home-container {
  max-width: 1400px;
  margin: 0 auto;
  height: 100%;
  min-height: 560px;
  display: flex;
  flex-direction: column;
}

/* 页面头部 */
.page-header {
  flex-shrink: 0;
  margin-bottom: clamp(10px, 1.6vh, 24px);
}

.page-title {
  font-size: clamp(20px, 2.8vh, 30px);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 4px 0;
  letter-spacing: -0.5px;
}

.page-subtitle {
  font-size: 13px;
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
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: clamp(12px, 1.6vh, 24px);
  margin-bottom: clamp(12px, 1.8vh, 24px);
  flex-shrink: 0;
}

/* --card-pad：卡片内边距随视口高度弹性收缩，水区波浪外扩量与其联动 */
.stat-card {
  --card-pad: clamp(14px, 2vh, 24px);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-base);
  background: var(--color-bg-card);
  position: relative;
  overflow: hidden;
}

.stat-card :deep(.el-card__body) {
  padding: var(--card-pad);
}

.stat-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

/* 水位波浪背景 */
/* 水位区域：头部以下，波浪背景向四周延伸至卡片边缘 */
.card-water-zone {
  position: relative;
}

.wave-bg {
  position: absolute;
  top: 0;
  left: calc(var(--card-pad) * -1);
  right: calc(var(--card-pad) * -1);
  bottom: calc(var(--card-pad) * -1);
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

.wave-water {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  /* 顶部透明度与 wave-1 浪花底部完全一致，消除水面硬边横线；水位由 JS 弹簧逐帧驱动 */
  background: linear-gradient(180deg, rgba(96, 165, 250, 0.22), rgba(59, 130, 246, 0.32));
  will-change: transform;
}

/* 前层浪：色深、绘制在后层浪之上 */
.wave-1 {
  z-index: 2;
}

.wave {
  position: absolute;
  top: -14px;
  left: 0;
  width: calc(100% + 140px);
  height: 14px;
  /* 浪花填充用纵向渐变，底部透明度=水体顶部透明度，与水体无缝衔接 */
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 140 14' preserveAspectRatio='none'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0' y1='0' x2='0' y2='1'%3E%3Cstop offset='0' stop-color='%2360A5FA' stop-opacity='0.36'/%3E%3Cstop offset='1' stop-color='%2360A5FA' stop-opacity='0.22'/%3E%3C/linearGradient%3E%3C/defs%3E%3Cpath d='M0 7 Q17.5 0 35 7 T70 7 T105 7 T140 7 V14 H0 Z' fill='url(%23g)'/%3E%3C/svg%3E") repeat-x;
  background-size: 140px 14px;
  background-position: 0 0;
  animation: waveScroll 7s linear infinite;
}

/* 后层浪：更浅更高，波峰露出前浪之上形成层叠；底部渐隐到 0 消除硬边 */
.wave-2 {
  top: -18px;
  height: 32px;
  z-index: 1;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 140 32' preserveAspectRatio='none'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0' y1='0' x2='0' y2='1'%3E%3Cstop offset='0' stop-color='%2393C5FD' stop-opacity='0.5'/%3E%3Cstop offset='1' stop-color='%2393C5FD' stop-opacity='0'/%3E%3C/linearGradient%3E%3C/defs%3E%3Cpath d='M0 7 Q17.5 0 35 7 T70 7 T105 7 T140 7 V32 H0 Z' fill='url(%23g)'/%3E%3C/svg%3E") repeat-x;
  background-size: 140px 32px;
  animation-duration: 13s;
  animation-direction: reverse;
}

/* 灌水期湍流泡沫浪：快速流动，仅灌水期显示，落定后淡出恢复平静 */
.wave-4 {
  top: -14px;
  height: 14px;
  z-index: 2;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 140 14' preserveAspectRatio='none'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='0' y1='0' x2='0' y2='1'%3E%3Cstop offset='0' stop-color='%23FFFFFF' stop-opacity='0.35'/%3E%3Cstop offset='1' stop-color='%23FFFFFF' stop-opacity='0'/%3E%3C/linearGradient%3E%3C/defs%3E%3Cpath d='M0 7 Q17.5 0 35 7 T70 7 T105 7 T140 7 V14 H0 Z' fill='url(%23g)'/%3E%3C/svg%3E") repeat-x;
  background-size: 140px 14px;
  animation: waveScroll 4s linear infinite;
  opacity: 0;
  transition: opacity 0.6s;
}

.is-filling .wave-4 {
  opacity: 1;
}

/* 水面内侧浪花线：满水（100%）时顶部仍可见波动 */
.wave-3 {
  top: 1px;
  height: 14px;
  z-index: 3;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 140 14' preserveAspectRatio='none'%3E%3Cpath d='M0 7 Q17.5 0 35 7 T70 7 T105 7 T140 7' fill='none' stroke='%23FFFFFF' stroke-opacity='0.65' stroke-width='3'/%3E%3C/svg%3E") repeat-x;
  background-size: 140px 14px;
  animation-duration: 9s;
}

/* 海面小船：置于 .wave-water 内随水位起伏；z1 同 wave-2 但 DOM 序后→盖在后浪上，
   前浪 wave-1(z2)/泡沫线 wave-3(z3) 盖住船底形成吃水感 */
.boat {
  position: absolute;
  top: -16px;
  /* 放在卡片右侧空白水面，避免遮挡左侧标签/数字/趋势文字 */
  left: 66%;
  width: 40px;
  height: 30px;
  z-index: 1;
  pointer-events: none;
  will-change: transform;
  animation: boatFloat 9s ease-in-out infinite;
}

/* 漂移 + 摇晃 + 浮动合成一条关键帧，避免多动画 transform 冲突 */
@keyframes boatFloat {
  0%, 100% {
    transform: translateX(-6px) translateY(0) rotate(-2deg);
  }
  25% {
    transform: translateX(-1px) translateY(-2px) rotate(1.5deg);
  }
  50% {
    transform: translateX(3px) translateY(0) rotate(2.5deg);
  }
  75% {
    transform: translateX(6px) translateY(-1.5px) rotate(-1deg);
  }
}

@keyframes waveScroll {
  to {
    transform: translateX(-140px);
  }
}

.bubble {
  position: absolute;
  top: 100%;
  height: 100%;
  opacity: 0;
  animation: bubbleRise var(--d, 8s) linear infinite;
  animation-delay: var(--delay, 0s);
}

.bubble::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: var(--s, 6px);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.45);
  box-shadow: inset 0 0 2px rgba(255, 255, 255, 0.6);
}

@keyframes bubbleRise {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  85% {
    opacity: 0.35;
  }
  100% {
    transform: translateY(-200%);
    opacity: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .wave {
    animation: none;
  }
  .bubble {
    display: none;
  }
  .wave-4 {
    display: none;
  }
  /* 小船仍显示，仅静止 */
  .boat {
    animation: none;
  }
}

/* 卡片头部行 */
.card-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: clamp(8px, 1.2vh, 16px);
  position: relative;
  z-index: 1;
}

/* 图标容器 */
.card-icon-wrapper {
  width: clamp(40px, 5.4vh, 52px);
  height: clamp(40px, 5.4vh, 52px);
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
  margin-bottom: clamp(8px, 1.2vh, 16px);
  position: relative;
  z-index: 1;
}

.card-label {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: clamp(2px, 0.6vh, 8px);
  font-weight: 500;
}

.card-value {
  font-size: clamp(24px, 3.6vh, 36px);
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
  margin-bottom: clamp(4px, 1vh, 12px);
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
  padding-top: clamp(8px, 1.2vh, 16px);
  margin-top: clamp(8px, 1.2vh, 16px);
  position: relative;
  z-index: 1;
}

/* 水区内分隔线：统一为白色泡沫色，与浪花/气泡同色系，避免灰边框在不同水深底色上观感不一 */
.card-water-zone .card-footer {
  border-top-color: rgba(255, 255, 255, 0.75);
}

.card-footer :deep(.el-button) {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  margin: 0 !important;
}

/* 图表区域：4 等分网格与上排统计卡对齐，条形图卡横跨后 3 列（1:3）；
   flex:1 吃掉统计卡与快捷操作之外的全部剩余高度 */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: clamp(12px, 1.6vh, 24px);
  margin-bottom: clamp(12px, 1.8vh, 24px);
  flex: 1 1 auto;
  min-height: 200px;
}

.chart-card-wide {
  grid-column: 2 / -1;
}

.chart-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
  background: var(--color-bg-card);
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

/* 卡片体填满剩余高度，内部纵向 Flex 供图表容器自适应 */
.chart-card :deep(.el-card__header) {
  padding: clamp(10px, 1.5vh, 18px) clamp(14px, 1.4vw, 20px);
}

.chart-card :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: clamp(10px, 1.6vh, 20px) clamp(14px, 1.4vw, 20px);
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

/* ECharts 图表容器 */
.echart-box {
  width: 100%;
}

/* 最近新增客户端列表：无序列表填满卡片剩余高度 */
.recent-client-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: clamp(2px, 0.8vh, 10px) 0;
}

.recent-client-list {
  list-style: none;
  margin: 0;
  padding: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  gap: 4px;
}

.recent-client-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 6px;
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
}

.recent-client-item:hover {
  background: var(--color-bg-hover);
}

.recent-client-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-primary);
  flex-shrink: 0;
}

.recent-client-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.recent-client-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.recent-client-meta {
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.recent-client-time {
  font-size: 12px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
  font-variant-numeric: tabular-nums;
}

.recent-client-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.region-chart-wrapper {
  position: relative;
  flex: 1;
  min-height: 0;
}

.region-chart-box {
  height: 100%;
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

/* 快捷操作 */
.action-section {
  text-align: center;
  flex-shrink: 0;
}

.action-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: clamp(6px, 1vh, 12px);
}

.action-buttons {
  display: flex;
  gap: clamp(8px, 1.4vh, 16px);
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
}
</style>
