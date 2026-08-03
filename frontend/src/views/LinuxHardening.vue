<template>
  <div class="linux-hardening-container">
    <!-- 表格 -->
    <el-table :data="tableData" v-loading="loading" style="width: 100%">
      <el-table-column type="index" label="#" width="50"></el-table-column>
      <el-table-column prop="hostname" label="计算机名" min-width="120"></el-table-column>
      <el-table-column prop="ip" label="IP" min-width="120"></el-table-column>
      <el-table-column prop="operasystem" label="系统" min-width="200"></el-table-column>
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
      title="加固检查详情" 
      :visible.sync="dialogVisible" 
      width="60%"
      max-height="80vh"
      append-to-body
    >
      <el-tabs v-if="currentDetail">
        <!-- 基本信息 -->
        <el-tab-pane label="基本信息">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="计算机名">{{ currentDetail.hostname }}</el-descriptions-item>
            <el-descriptions-item label="IP">{{ currentDetail.ip }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ currentDetail.operasystem }}</el-descriptions-item>
            <el-descriptions-item label="内核版本">{{ currentDetail.kernel }}</el-descriptions-item>
            <el-descriptions-item label="检查时间" :span="2">{{ currentDetail.date }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 系统更新配置 -->
        <el-tab-pane label="系统更新">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="dnf.conf_gpgcheck"
              :class="{'non-compliant': isNonCompliant('dnf_conf_gpgcheck')}"
            >
              {{ currentDetail.dnf_conf_gpgcheck }}
              <span v-if="isNonCompliant('dnf_conf_gpgcheck')" class="standard-hint">
                (标准: {{ getStandardValue('dnf_conf_gpgcheck') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="redhat.repo_gpgcheck"
              :class="{'non-compliant': isNonCompliant('redhat_repo_gpgcheck')}"
            >
              {{ currentDetail.redhat_repo_gpgcheck }}
              <span v-if="isNonCompliant('redhat_repo_gpgcheck')" class="standard-hint">
                (标准: {{ getStandardValue('redhat_repo_gpgcheck') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 用户账户策略 -->
        <el-tab-pane name="user-policy" label="用户账户策略">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="PASS_MAX_DAYS"
              :class="{'non-compliant': isNonCompliant('pass_max_days')}"
            >
              {{ currentDetail.pass_max_days }}
              <span v-if="isNonCompliant('pass_max_days')" class="standard-hint">
                (标准: {{ getStandardValue('pass_max_days') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PASS_MIN_DAYS"
              :class="{'non-compliant': isNonCompliant('pass_min_days')}"
            >
              {{ currentDetail.pass_min_days }}
              <span v-if="isNonCompliant('pass_min_days')" class="standard-hint">
                (标准: {{ getStandardValue('pass_min_days') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PASS_MIN_LEN"
              :class="{'non-compliant': isNonCompliant('pass_min_len')}"
            >
              {{ currentDetail.pass_min_len }}
              <span v-if="isNonCompliant('pass_min_len')" class="standard-hint">
                (标准: {{ getStandardValue('pass_min_len') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PASS_WARN_AGE"
              :class="{'non-compliant': isNonCompliant('pass_warn_age')}"
            >
              {{ currentDetail.pass_warn_age }}
              <span v-if="isNonCompliant('pass_warn_age')" class="standard-hint">
                (标准: {{ getStandardValue('pass_warn_age') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="INACTIVE"
              :class="{'non-compliant': isNonCompliant('inactive')}"
            >
              {{ currentDetail.inactive }}
              <span v-if="isNonCompliant('inactive')" class="standard-hint">
                (标准: {{ getStandardValue('inactive') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="GID (root)"
              :class="{'non-compliant': isNonCompliant('gid')}"
            >
              {{ currentDetail.gid }}
              <span v-if="isNonCompliant('gid')" class="standard-hint">
                (标准: {{ getStandardValue('gid') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="TMOUT"
              :class="{'non-compliant': isNonCompliant('tmout')}"
            >
              {{ currentDetail.tmout }}
              <span v-if="isNonCompliant('tmout')" class="standard-hint">
                (标准: {{ getStandardValue('tmout') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 计划任务配置 -->
        <el-tab-pane name="cron-config" label="计划任务">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="Cron 守护进程"
              :class="{'non-compliant': isNonCompliant('cron')}"
            >
              {{ currentDetail.cron }}
              <span v-if="isNonCompliant('cron')" class="standard-hint">
                (标准：{{ getStandardValue('cron') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="crontab 权限"
              :class="{'non-compliant': isNonCompliant('crontab')}"
              :span="2"
            >
              {{ currentDetail.crontab }}
              <span v-if="isNonCompliant('crontab')" class="standard-hint">
                (标准：{{ getStandardValue('crontab') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.hourly 权限"
              :class="{'non-compliant': isNonCompliant('cron_hourly')}"
              :span="2"
            >
              {{ currentDetail.cron_hourly }}
              <span v-if="isNonCompliant('cron_hourly')" class="standard-hint">
                (标准：{{ getStandardValue('cron_hourly') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.daily 权限"
              :class="{'non-compliant': isNonCompliant('cron_daily')}"
              :span="2"
            >
              {{ currentDetail.cron_daily }}
              <span v-if="isNonCompliant('cron_daily')" class="standard-hint">
                (标准：{{ getStandardValue('cron_daily') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.weekly 权限"
              :class="{'non-compliant': isNonCompliant('cron_weekly')}"
              :span="2"
            >
              {{ currentDetail.cron_weekly }}
              <span v-if="isNonCompliant('cron_weekly')" class="standard-hint">
                (标准：{{ getStandardValue('cron_weekly') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.monthly 权限"
              :class="{'non-compliant': isNonCompliant('cron_monthly')}"
              :span="2"
            >
              {{ currentDetail.cron_monthly }}
              <span v-if="isNonCompliant('cron_monthly')" class="standard-hint">
                (标准：{{ getStandardValue('cron_monthly') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.deny"
              :class="{'non-compliant': isNonCompliant('cron_deny')}"
              :span="2"
            >
              {{ currentDetail.cron_deny }}
              <span v-if="isNonCompliant('cron_deny')" class="standard-hint">
                (标准：{{ getStandardValue('cron_deny') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="at.deny"
              :class="{'non-compliant': isNonCompliant('at_deny')}"
              :span="2"
            >
              {{ currentDetail.at_deny }}
              <span v-if="isNonCompliant('at_deny')" class="standard-hint">
                (标准：{{ getStandardValue('at_deny') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="cron.allow"
              :class="{'non-compliant': isNonCompliant('cron_allow')}"
              :span="2"
            >
              {{ currentDetail.cron_allow }}
              <span v-if="isNonCompliant('cron_allow')" class="standard-hint">
                (标准：{{ getStandardValue('cron_allow') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="at.allow"
              :class="{'non-compliant': isNonCompliant('at_allow')}"
              :span="2"
            >
              {{ currentDetail.at_allow }}
              <span v-if="isNonCompliant('at_allow')" class="standard-hint">
                (标准：{{ getStandardValue('at_allow') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- SSH 服务器配置 -->
        <el-tab-pane name="ssh-config" label="SSH 配置">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="sshd_config 权限"
              :class="{'non-compliant': isNonCompliant('sshd_config')}"
              :span="2"
            >
              {{ currentDetail.sshd_config }}
              <span v-if="isNonCompliant('sshd_config')" class="standard-hint">
                (标准：{{ getStandardValue('sshd_config') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="LogLevel"
              :class="{'non-compliant': isNonCompliant('log_level')}"
            >
              {{ currentDetail.log_level }}
              <span v-if="isNonCompliant('log_level')" class="standard-hint">
                (标准：{{ getStandardValue('log_level') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="X11Forwarding"
              :class="{'non-compliant': isNonCompliant('x11_forwarding')}"
            >
              {{ currentDetail.x11_forwarding }}
              <span v-if="isNonCompliant('x11_forwarding')" class="standard-hint">
                (标准：{{ getStandardValue('x11_forwarding') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="MaxAuthTries"
              :class="{'non-compliant': isNonCompliant('max_auth_tries')}"
            >
              {{ currentDetail.max_auth_tries }}
              <span v-if="isNonCompliant('max_auth_tries')" class="standard-hint">
                (标准：{{ getStandardValue('max_auth_tries') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="IgnoreRhosts"
              :class="{'non-compliant': isNonCompliant('ignore_rhosts')}"
            >
              {{ currentDetail.ignore_rhosts }}
              <span v-if="isNonCompliant('ignore_rhosts')" class="standard-hint">
                (标准：{{ getStandardValue('ignore_rhosts') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="HostbasedAuthentication"
              :class="{'non-compliant': isNonCompliant('hostbased_authentication')}"
            >
              {{ currentDetail.hostbased_authentication }}
              <span v-if="isNonCompliant('hostbased_authentication')" class="standard-hint">
                (标准：{{ getStandardValue('hostbased_authentication') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PermitRootLogin"
              :class="{'non-compliant': isNonCompliant('permit_root_login')}"
            >
              {{ currentDetail.permit_root_login }}
              <span v-if="isNonCompliant('permit_root_login')" class="standard-hint">
                (标准：{{ getStandardValue('permit_root_login') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PermitEmptyPasswords"
              :class="{'non-compliant': isNonCompliant('permit_empty_passwords')}"
            >
              {{ currentDetail.permit_empty_passwords }}
              <span v-if="isNonCompliant('permit_empty_passwords')" class="standard-hint">
                (标准：{{ getStandardValue('permit_empty_passwords') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="PermitUserEnvironment"
              :class="{'non-compliant': isNonCompliant('permit_user_environment')}"
            >
              {{ currentDetail.permit_user_environment }}
              <span v-if="isNonCompliant('permit_user_environment')" class="standard-hint">
                (标准：{{ getStandardValue('permit_user_environment') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="ClientAliveInterval"
              :class="{'non-compliant': isNonCompliant('client_alive_interval')}"
            >
              {{ currentDetail.client_alive_interval }}
              <span v-if="isNonCompliant('client_alive_interval')" class="standard-hint">
                (标准：{{ getStandardValue('client_alive_interval') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="ClientAliveCountMax"
              :class="{'non-compliant': isNonCompliant('client_alive_count_max')}"
            >
              {{ currentDetail.client_alive_count_max }}
              <span v-if="isNonCompliant('client_alive_count_max')" class="standard-hint">
                (标准：{{ getStandardValue('client_alive_count_max') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="LoginGraceTime"
              :class="{'non-compliant': isNonCompliant('login_grace_time')}"
            >
              {{ currentDetail.login_grace_time }}
              <span v-if="isNonCompliant('login_grace_time')" class="standard-hint">
                (标准：{{ getStandardValue('login_grace_time') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 密码策略与复杂度 -->
        <el-tab-pane name="password-policy" label="密码策略">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="minlen"
              :class="{'non-compliant': isNonCompliant('minlen')}"
            >
              {{ currentDetail.minlen }}
              <span v-if="isNonCompliant('minlen')" class="standard-hint">
                (标准：{{ getStandardValue('minlen') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="minclass"
              :class="{'non-compliant': isNonCompliant('minclass')}"
            >
              {{ currentDetail.minclass }}
              <span v-if="isNonCompliant('minclass')" class="standard-hint">
                (标准：{{ getStandardValue('minclass') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="dcredit"
              :class="{'non-compliant': isNonCompliant('dcredit')}"
            >
              {{ currentDetail.dcredit }}
              <span v-if="isNonCompliant('dcredit')" class="standard-hint">
                (标准：{{ getStandardValue('dcredit') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="ucredit"
              :class="{'non-compliant': isNonCompliant('ucredit')}"
            >
              {{ currentDetail.ucredit }}
              <span v-if="isNonCompliant('ucredit')" class="standard-hint">
                (标准：{{ getStandardValue('ucredit') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="lcredit"
              :class="{'non-compliant': isNonCompliant('lcredit')}"
            >
              {{ currentDetail.lcredit }}
              <span v-if="isNonCompliant('lcredit')" class="standard-hint">
                (标准：{{ getStandardValue('lcredit') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="ocredit"
              :class="{'non-compliant': isNonCompliant('ocredit')}"
            >
              {{ currentDetail.ocredit }}
              <span v-if="isNonCompliant('ocredit')" class="standard-hint">
                (标准：{{ getStandardValue('ocredit') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="password_remember"
              :class="{'non-compliant': isNonCompliant('password_remember')}"
              :span="2"
            >
              {{ currentDetail.password_remember }}
              <span v-if="isNonCompliant('password_remember')" class="standard-hint">
                (标准：{{ getStandardValue('password_remember') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 文件系统权限 -->
        <el-tab-pane name="file-permissions" label="文件权限">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="passwd 权限"
              :class="{'non-compliant': isNonCompliant('passwd')}"
            >
              {{ currentDetail.passwd }}
              <span v-if="isNonCompliant('passwd')" class="standard-hint">
                (标准：{{ getStandardValue('passwd') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="passwd- 权限"
              :class="{'non-compliant': isNonCompliant('passwd_minus')}"
            >
              {{ currentDetail.passwd_minus }}
              <span v-if="isNonCompliant('passwd_minus')" class="standard-hint">
                (标准：{{ getStandardValue('passwd_minus') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="group 权限"
              :class="{'non-compliant': isNonCompliant('group')}"
            >
              {{ currentDetail.group }}
              <span v-if="isNonCompliant('group')" class="standard-hint">
                (标准：{{ getStandardValue('group') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="group- 权限"
              :class="{'non-compliant': isNonCompliant('group_minus')}"
            >
              {{ currentDetail.group_minus }}
              <span v-if="isNonCompliant('group_minus')" class="standard-hint">
                (标准：{{ getStandardValue('group_minus') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="shadow 权限"
              :class="{'non-compliant': isNonCompliant('shadow')}"
            >
              {{ currentDetail.shadow }}
              <span v-if="isNonCompliant('shadow')" class="standard-hint">
                (标准：{{ getStandardValue('shadow') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="shadow- 权限"
              :class="{'non-compliant': isNonCompliant('shadow_minus')}"
            >
              {{ currentDetail.shadow_minus }}
              <span v-if="isNonCompliant('shadow_minus')" class="standard-hint">
                (标准：{{ getStandardValue('shadow_minus') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="gshadow 权限"
              :class="{'non-compliant': isNonCompliant('gshadow')}"
            >
              {{ currentDetail.gshadow }}
              <span v-if="isNonCompliant('gshadow')" class="standard-hint">
                (标准：{{ getStandardValue('gshadow') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="gshadow- 权限"
              :class="{'non-compliant': isNonCompliant('gshadow_minus')}"
            >
              {{ currentDetail.gshadow_minus }}
              <span v-if="isNonCompliant('gshadow_minus')" class="standard-hint">
                (标准：{{ getStandardValue('gshadow_minus') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 加密与时钟同步 -->
        <el-tab-pane name="crypto-sync" label="加密与时钟">
          <el-descriptions :column="2" border>
            <el-descriptions-item 
              label="CryptoPolicies"
              :class="{'non-compliant': isNonCompliant('crypto_policies')}"
              :span="2"
            >
              {{ currentDetail.crypto_policies }}
              <span v-if="isNonCompliant('crypto_policies')" class="standard-hint">
                (标准：{{ getStandardValue('crypto_policies') }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item 
              label="NTPServer"
              :class="{'non-compliant': isNonCompliant('ntp_server')}"
              :span="2"
            >
              {{ currentDetail.ntp_server }}
              <span v-if="isNonCompliant('ntp_server')" class="standard-hint">
                (标准：{{ getStandardValue('ntp_server') }})
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script>
import { getList, getDetail } from '@/api/linux-checks'

export default {
  name: 'LinuxHardening',
  data() {
    return {
      loading: false,
      tableData: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      dialogVisible: false,
      currentDetail: null,
      complianceData: null // 合规比对结果
    }
  },
  created() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const res = await getList({
          page: this.currentPage,
          pageSize: this.pageSize
        })
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
    }
  }
}
</script>

<style scoped>
.linux-hardening-container {
  padding: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

/* 不合规字段高亮样式 */
.non-compliant {
  background-color: #fef0f0 !important;
  border-color: #fde2e2 !important;
}

.non-compliant .el-descriptions-item__content {
  color: #f56c6c;
  font-weight: bold;
}

/* 标准值提示 */
.standard-hint {
  color: #e6a23c;
  font-size: 12px;
  margin-left: 8px;
  font-weight: normal;
}
</style>
