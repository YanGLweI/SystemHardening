<template>
  <div class="content-container">
  <!-- 操作栏 -->
  <div class="action-bar">
    <div class="action-title">
      <h2>Windows 加固检查</h2>
      <p>对 Windows 系统进行安全合规检查和评估</p>
    </div>
  </div>

  <!-- 搜索栏 -->
  <div class="search-bar">
    <div class="search-left">
      <el-input
        v-model="keyword"
        placeholder="搜索计算机名、IP 或域名"
        prefix-icon="el-icon-search"
        clearable
        class="search-input"
        @keyup.enter.native="handleSearch"
        @clear="handleSearch"
      ></el-input>
      <el-select
        v-model="complianceStatus"
        placeholder="合规状态"
        clearable
        class="status-select"
        @change="handleSearch"
      >
        <el-option label="合规" value="compliant"></el-option>
        <el-option label="不合规" value="non_compliant"></el-option>
      </el-select>
      <el-button type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
    </div>
    <el-button icon="el-icon-refresh" @click="handleRefresh">刷新</el-button>
  </div>

  <!-- 表格 -->
  <div class="table-wrapper">
      <el-table :data="tableData" v-loading="loading" style="width: 100%" :max-height="tableMaxHeight">
        <el-table-column type="index" label="#" width="50"></el-table-column>
        <el-table-column prop="hostname" label="计算机名" min-width="120"></el-table-column>
        <el-table-column prop="domainname" label="域名" min-width="120"></el-table-column>
        <el-table-column prop="ip" label="IP" min-width="120"></el-table-column>
        <el-table-column prop="operasystem" label="操作系统" min-width="200"></el-table-column>
        <el-table-column label="合规状态" min-width="100">
          <template slot-scope="{row}">
            <el-tag :type="row.compliance_status === 'compliant' ? 'success' : 'danger'">
              {{ row.compliance_status === 'compliant' ? '合规' : '不合规' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template slot-scope="{row}">
            <el-button
              size="small"
              type="primary"
              @click="handleDetail(row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页 -->
    <el-pagination
      class="pagination"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      :current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
      :page-sizes="[10, 20, 50, 100]"
    ></el-pagination>

  <!-- 详情弹窗 -->
  <el-dialog
    title="Windows 加固检查详情"
    :visible.sync="dialogVisible"
    width="70%"
    max-height="80vh"
    append-to-body
  >
    <el-tabs v-if="currentDetail" type="border-card">
      <!-- 基本信息 -->
      <el-tab-pane label="基本信息">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="计算机名">{{ currentDetail.hostname }}</el-descriptions-item>
          <el-descriptions-item label="域名">{{ currentDetail.domainname }}</el-descriptions-item>
          <el-descriptions-item label="IP 地址">{{ currentDetail.ip }}</el-descriptions-item>
          <el-descriptions-item label="操作系统">{{ currentDetail.operasystem }}</el-descriptions-item>
          <el-descriptions-item 
            label="激活状态"
            :class="{'non-compliant': isNonCompliant('license_result')}"
          >
            {{ currentDetail.LicenseResult || '-' }}
            <span v-if="isNonCompliant('license_result')" class="standard-hint">
              (标准：{{ formatStandardValue('license_result') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="检查时间">{{ currentDetail.date }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 账户密码策略 -->
      <el-tab-pane name="password-policy" label="账户密码策略">
        <el-descriptions :column="2" border>
          <el-descriptions-item
            label="密码最短使用天数"
            :class="{'non-compliant': isNonCompliant('minimum_password_age')}"
          >
            {{ formatValue(currentDetail.minimum_password_age) }}
            <span v-if="isNonCompliant('minimum_password_age')" class="standard-hint">
              (标准：{{ formatStandardValue('minimum_password_age') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="密码最长使用天数"
            :class="{'non-compliant': isNonCompliant('maximum_password_age')}"
          >
            {{ formatValue(currentDetail.maximum_password_age) }}
            <span v-if="isNonCompliant('maximum_password_age')" class="standard-hint">
              (标准：{{ formatStandardValue('maximum_password_age') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="密码最小长度"
            :class="{'non-compliant': isNonCompliant('minimum_password_length')}"
          >
            {{ formatValue(currentDetail.minimum_password_length) }}
            <span v-if="isNonCompliant('minimum_password_length')" class="standard-hint">
              (标准：{{ formatStandardValue('minimum_password_length') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="密码复杂度"
            :class="{'non-compliant': isNonCompliant('password_complexity')}"
          >
            {{ formatBoolean(currentDetail.password_complexity) }}
            <span v-if="isNonCompliant('password_complexity')" class="standard-hint">
              (标准：{{ formatStandardValue('password_complexity') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="密码历史记录数"
            :class="{'non-compliant': isNonCompliant('password_history_size')}"
          >
            {{ formatValue(currentDetail.password_history_size) }}
            <span v-if="isNonCompliant('password_history_size')" class="standard-hint">
              (标准：{{ formatStandardValue('password_history_size') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="账户锁定阈值"
            :class="{'non-compliant': isNonCompliant('lockout_bad_count')}"
          >
            {{ formatValue(currentDetail.lockout_bad_count) }}
            <span v-if="isNonCompliant('lockout_bad_count')" class="standard-hint">
              (标准：{{ formatStandardValue('lockout_bad_count') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="锁定持续时间 (分钟)"
            :class="{'non-compliant': isNonCompliant('lockout_duration')}"
          >
            {{ formatValue(currentDetail.lockout_duration) }}
            <span v-if="isNonCompliant('lockout_duration')" class="standard-hint">
              (标准：{{ formatStandardValue('lockout_duration') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="重置锁定计数 (分钟)"
            :class="{'non-compliant': isNonCompliant('reset_lockout_count')}"
          >
            {{ formatValue(currentDetail.reset_lockout_count) }}
            <span v-if="isNonCompliant('reset_lockout_count')" class="standard-hint">
              (标准：{{ formatStandardValue('reset_lockout_count') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="登录后必须更改密码"
            :class="{'non-compliant': isNonCompliant('require_logon_to_change_password')}"
          >
            {{ formatBoolean(currentDetail.require_logon_to_change_password) }}
            <span v-if="isNonCompliant('require_logon_to_change_password')" class="standard-hint">
              (标准：{{ formatStandardValue('require_logon_to_change_password') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="管理员名称"
            :class="{'non-compliant': isNonCompliant('new_administrator_name')}"
          >
            {{ formatValue(currentDetail.new_administrator_name) }}
            <span v-if="isNonCompliant('new_administrator_name')" class="standard-hint">
              (标准：{{ formatStandardValue('new_administrator_name') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="来宾名称"
            :class="{'non-compliant': isNonCompliant('new_guest_name')}"
          >
            {{ formatValue(currentDetail.new_guest_name) }}
            <span v-if="isNonCompliant('new_guest_name')" class="standard-hint">
              (标准：{{ formatStandardValue('new_guest_name') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="明文密码存储"
            :class="{'non-compliant': isNonCompliant('clear_text_password')}"
          >
            {{ formatBoolean(currentDetail.clear_text_password) }}
            <span v-if="isNonCompliant('clear_text_password')" class="standard-hint">
              (标准：{{ formatStandardValue('clear_text_password') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="LSA 匿名名称查找"
            :class="{'non-compliant': isNonCompliant('lsa_anonymous_name_lookup')}"
          >
            {{ formatBoolean(currentDetail.lsa_anonymous_name_lookup) }}
            <span v-if="isNonCompliant('lsa_anonymous_name_lookup')" class="standard-hint">
              (标准：{{ formatStandardValue('lsa_anonymous_name_lookup') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="启用管理员账户"
            :class="{'non-compliant': isNonCompliant('enable_admin_account')}"
          >
            {{ formatBoolean(currentDetail.enable_admin_account) }}
            <span v-if="isNonCompliant('enable_admin_account')" class="standard-hint">
              (标准：{{ formatStandardValue('enable_admin_account') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="启用来宾账户"
            :class="{'non-compliant': isNonCompliant('enable_guest_account')}"
          >
            {{ formatBoolean(currentDetail.enable_guest_account) }}
            <span v-if="isNonCompliant('enable_guest_account')" class="standard-hint">
              (标准：{{ formatStandardValue('enable_guest_account') }})
            </span>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 审计策略 -->
      <el-tab-pane name="audit-policy" label="审计策略">
        <el-descriptions :column="3" border>
          <el-descriptions-item
            label="系统事件"
            :class="{'non-compliant': isNonCompliant('audit_system_events')}"
          >
            {{ getAuditLevel(currentDetail.audit_system_events) }}
            <span v-if="isNonCompliant('audit_system_events')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_system_events') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="登录事件"
            :class="{'non-compliant': isNonCompliant('audit_logon_events')}"
          >
            {{ getAuditLevel(currentDetail.audit_logon_events) }}
            <span v-if="isNonCompliant('audit_logon_events')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_logon_events') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="对象访问"
            :class="{'non-compliant': isNonCompliant('audit_object_access')}"
          >
            {{ getAuditLevel(currentDetail.audit_object_access) }}
            <span v-if="isNonCompliant('audit_object_access')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_object_access') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="特权使用"
            :class="{'non-compliant': isNonCompliant('audit_privilege_use')}"
          >
            {{ getAuditLevel(currentDetail.audit_privilege_use) }}
            <span v-if="isNonCompliant('audit_privilege_use')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_privilege_use') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="策略更改"
            :class="{'non-compliant': isNonCompliant('audit_policy_change')}"
          >
            {{ getAuditLevel(currentDetail.audit_policy_change) }}
            <span v-if="isNonCompliant('audit_policy_change')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_policy_change') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="账户管理"
            :class="{'non-compliant': isNonCompliant('audit_account_manage')}"
          >
            {{ getAuditLevel(currentDetail.audit_account_manage) }}
            <span v-if="isNonCompliant('audit_account_manage')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_account_manage') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="进程跟踪"
            :class="{'non-compliant': isNonCompliant('audit_process_tracking')}"
          >
            {{ getAuditLevel(currentDetail.audit_process_tracking) }}
            <span v-if="isNonCompliant('audit_process_tracking')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_process_tracking') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="DS 访问"
            :class="{'non-compliant': isNonCompliant('audit_ds_access')}"
          >
            {{ getAuditLevel(currentDetail.audit_ds_access) }}
            <span v-if="isNonCompliant('audit_ds_access')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_ds_access') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="账户登录"
            :class="{'non-compliant': isNonCompliant('audit_account_logon')}"
          >
            {{ getAuditLevel(currentDetail.audit_account_logon) }}
            <span v-if="isNonCompliant('audit_account_logon')" class="standard-hint">
              (标准：{{ formatStandardValue('audit_account_logon') }})
            </span>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 设备控制与屏幕保护 -->
      <el-tab-pane name="device-screensaver" label="设备与屏保">
        <el-descriptions :column="2" border>
          <el-descriptions-item
            label="移动存储设备"
            :class="{'non-compliant': isNonCompliant('storage_devices')}"
          >
            {{ formatBoolean(currentDetail.storage_devices) }}
            <span v-if="isNonCompliant('storage_devices')" class="standard-hint">
              (标准：{{ formatStandardValue('storage_devices') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="屏幕保护启用"
            :class="{'non-compliant': isNonCompliant('screen_saver_active')}"
          >
            {{ formatBoolean(currentDetail.screen_saver_active) }}
            <span v-if="isNonCompliant('screen_saver_active')" class="standard-hint">
              (标准：{{ formatStandardValue('screen_saver_active') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="屏幕保护安全"
            :class="{'non-compliant': isNonCompliant('screen_saver_secure')}"
          >
            {{ formatBoolean(currentDetail.screen_saver_secure) }}
            <span v-if="isNonCompliant('screen_saver_secure')" class="standard-hint">
              (标准：{{ formatStandardValue('screen_saver_secure') }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item
            label="屏保超时 (秒)"
            :class="{'non-compliant': isNonCompliant('screen_save_timeout')}"
          >
            {{ formatValue(currentDetail.screen_save_timeout) }}
            <span v-if="isNonCompliant('screen_save_timeout')" class="standard-hint">
              (标准：{{ formatStandardValue('screen_save_timeout') }})
            </span>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <!-- 合规详情 -->
      <el-tab-pane name="compliance-detail" label="合规详情">
        <div v-if="!hasNonCompliantFields" style="text-align: center; padding: 20px;">
          <i class="el-icon-success" style="color: #67C23A; font-size: 48px;"></i>
          <p style="margin-top: 16px; color: #67C23A; font-weight: bold;">全部合规</p>
        </div>
        <div v-else>
          <el-alert title="发现以下不合规项：" type="warning" show-icon :closable="false">
            <el-table :data="complianceData.non_compliant_fields" style="width: 100%; margin-top: 12px;">
              <el-table-column prop="label" label="字段"></el-table-column>
              <el-table-column prop="actual" label="实际值"></el-table-column>
              <el-table-column prop="standard" label="标准值"></el-table-column>
            </el-table>
          </el-alert>
        </div>
      </el-tab-pane>
    </el-tabs>
    <span slot="footer" class="dialog-footer">
      <el-button @click="dialogVisible = false" style="border-color: var(--color-border); color: var(--color-text-regular);">关闭</el-button>
    </span>
  </el-dialog>
  </div>
</template>

<script>
import { getList, getDetail } from '@/api/windows-checks'

export default {
  name: 'WindowsHardeningContent',
  data() {
    return {
      loading: false,
      tableData: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      dialogVisible: false,
      currentDetail: null,
      complianceData: null, // 合规比对结果
      keyword: '',
      complianceStatus: '',
      tableMaxHeight: 500
    }
  },
  computed: {
    hasNonCompliantFields() {
      return this.complianceData &&
        this.complianceData.non_compliant_fields &&
        this.complianceData.non_compliant_fields.length > 0
    }
  },
  created() {
    this.fetchData()
  },
  mounted() {
    this.$nextTick(() => {
      this.updateTableMaxHeight()
    })
    window.addEventListener('resize', this.updateTableMaxHeight)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.updateTableMaxHeight)
  },
  methods: {
    updateTableMaxHeight() {
      // 基于实际容器高度精确计算
      this.$nextTick(() => {
        const contentContainer = this.$el.querySelector('.content-container')
        if (contentContainer) {
          const actionBar = contentContainer.querySelector('.action-bar')
          const searchBar = contentContainer.querySelector('.search-bar')
          const used = (actionBar ? actionBar.offsetHeight : 0) +
                       (searchBar ? searchBar.offsetHeight : 0)
          this.tableMaxHeight = contentContainer.clientHeight - used
        } else {
          this.tableMaxHeight = window.innerHeight - 340
        }
      })
    },

    async fetchData() {
      this.loading = true
      try {
        const params = {
          page: this.currentPage,
          pageSize: this.pageSize
        }
        if (this.keyword) {
          params.keyword = this.keyword
        }
        if (this.complianceStatus) {
          params.compliance_status = this.complianceStatus
        }
        const res = await getList(params)
        if (res.list && res.total !== undefined) {
          this.tableData = res.list
          this.total = res.total
        } else {
          this.$message.error('数据格式错误')
        }
      } catch (error) {
        console.error('获取数据失败:', error)
        this.$message.error('获取数据失败')
      } finally {
        this.loading = false
      }
    },
    handleDetail(row) {
      this.dialogVisible = true
      this.loading = true
      getDetail(row.id).then(res => {
        // 后端返回 {check, compliance}，我们需要提取 check 对象
        this.currentDetail = res.check || res
        this.complianceData = res.compliance || null
      }).catch(error => {
        console.error('获取详情失败:', error)
        this.$message.error('获取详情失败')
      }).finally(() => {
        this.loading = false
      })
    },
    // 检查字段是否不合规
    isNonCompliant(fieldName) {
      if (!this.complianceData || !this.complianceData.non_compliant_fields) {
        return false
      }
      return this.complianceData.non_compliant_fields.some(field => field.field === fieldName)
    },
    // 获取字段的标准值
    getStandardValue(fieldName) {
      if (!this.complianceData || !this.complianceData.non_compliant_fields) {
        return ''
      }
      const field = this.complianceData.non_compliant_fields.find(f => f.field === fieldName)
      return field ? field.standard : ''
    },
    handleSizeChange(val) {
      this.pageSize = val
      this.fetchData()
    },
    handleCurrentChange(val) {
      this.currentPage = val
      this.fetchData()
    },
    handleSearch() {
      this.currentPage = 1
      this.fetchData()
    },
    handleRefresh() {
      this.fetchData()
    },
    formatBoolean(value) {
      if (value === null || value === undefined || value === '') return '-'
      const val = parseInt(value)
      if (val === 0) return '否'
      if (val === 1) return '是'
      return value
    },
    getAuditLevel(value) {
      if (value === null || value === undefined || value === '') return '-'
      const levels = {
        '0': '无审计',
        '1': '成功',
        '2': '失败',
        '3': '成功与失败'
      }
      return levels[value] || value
    },
    formatValue(value) {
      if (value === null || value === undefined || value === '') return '-'
      return value
    },
    formatStandardValue(fieldName) {
      const value = this.getStandardValue(fieldName)
      if (!value) return value

      // 布尔类型字段（0/1 → 否/是）
      const booleanFields = [
        'password_complexity', 'require_logon_to_change_password',
        'clear_text_password', 'lsa_anonymous_name_lookup',
        'enable_admin_account', 'enable_guest_account',
        'storage_devices', 'screen_saver_active', 'screen_saver_secure'
      ]

      // 审计级别字段（0/1/2/3 → 无审计/成功/失败/成功与失败）
      const auditFields = [
        'audit_system_events', 'audit_logon_events', 'audit_object_access',
        'audit_privilege_use', 'audit_policy_change', 'audit_account_manage',
        'audit_process_tracking', 'audit_ds_access', 'audit_account_logon'
      ]

      if (booleanFields.includes(fieldName)) {
        return this.formatBoolean(value)
      }

      if (auditFields.includes(fieldName)) {
        return this.getAuditLevel(value)
      }

      // 数值和字符串字段直接返回原始值
      return value
    }
  }
}
</script>

<style scoped lang="scss">
/* 🟢 内容容器 */
.content-container {
  background: transparent;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 🟢 操作栏 */
.action-bar {
  margin-bottom: var(--spacing-6);
}

.action-title {
  h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--color-text-primary);
  }

  p {
    margin: 4px 0 0 0;
    font-size: 13px;
    color: var(--color-text-secondary);
  }
}

/* 🟢 搜索栏 */
.search-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
  flex-shrink: 0;

  .search-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search-input {
    width: 240px;
  }

  .status-select {
    width: 140px;
  }
}

.table-wrapper {
  flex: 1;
  overflow: hidden;
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);

  &:hover {
    box-shadow: var(--shadow-lg);
  }

  :deep(.el-table) {
    border-radius: var(--radius-lg);
  }
}

.el-table__row {
  transition: all var(--transition-base);
}

/* 🟢 分页样式 */
.pagination {
  margin-top: var(--spacing-6);
  text-align: right;
  padding: var(--spacing-4) 0;
  flex-shrink: 0;
}

/* 🟢 不合规字段高亮 - 薄荷绿主题 */
.non-compliant {
  background-color: var(--color-primary-alpha-10) !important;
  border-left: 3px solid var(--color-danger) !important;
}

.non-compliant .el-descriptions-item__content {
  color: var(--color-danger);
  font-weight: 600;
}

/* 🟢 标准值提示 */
.standard-hint {
  color: var(--color-warning);
  font-size: 12px;
  margin-left: 8px;
  font-weight: normal;
  background: var(--color-warning-alpha-10, rgba(245, 158, 11, 0.1));
  padding: 2px 6px;
  border-radius: 4px;
}

/* 🟢 弹窗优化 */
:deep(.el-dialog) {
  border-radius: var(--radius-xl);
  overflow: hidden;

  .el-dialog__header {
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
    padding: 20px 24px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.3);

    .el-dialog__title {
      color: white;
      font-weight: 600;
      font-size: 20px;
    }

    .el-dialog__headerbtn .el-dialog__close {
      color: white;
      opacity: 0.8;

      &:hover {
        opacity: 1;
      }
    }
  }

  .el-dialog__body {
    padding: var(--spacing-6);
    background: var(--color-bg-page);
  }

  .el-dialog__footer {
    border-top: 1px solid var(--color-border-light);
    padding: var(--spacing-4);
  }
}

/* 🟢 Tabs 样式 */
:deep(.el-tabs) {
  background: white;
  border-radius: var(--radius-md);
  padding: var(--spacing-4);
  box-shadow: var(--shadow-sm);

  .el-tabs__header {
    margin-bottom: var(--spacing-4);

    .el-tabs__item {
      height: 44px;
      line-height: 44px;
      padding: 0 16px;
      font-weight: 500;

      &.is-active {
        color: var(--color-primary);
        background: var(--color-primary-alpha-10);
        font-weight: 600;
      }

      &:hover {
        color: var(--color-primary);
      }
    }

    .el-tabs__active-bar {
      background: var(--color-primary);
    }

    .el-tabs__nav-wrap::after {
      background: var(--color-border-light);
    }
  }
}

/* 🟢 Descriptions 优化 */
:deep(.el-descriptions) {
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);

  .el-descriptions-item__label {
    font-weight: 600;
    color: var(--color-text-primary);
  }

  .el-descriptions-item__content {
    color: var(--color-text-regular);
    font-weight: 500;
  }
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  .search-bar {
    flex-direction: column;
    gap: 12px;

    .search-left {
      flex-wrap: wrap;
      width: 100%;

      .search-input {
        width: 100%;
      }

      .status-select {
        width: 100%;
      }
    }
  }

  .pagination {
    text-align: center;
  }

  :deep(.el-table) {
    font-size: 12px;
  }
}
</style>
