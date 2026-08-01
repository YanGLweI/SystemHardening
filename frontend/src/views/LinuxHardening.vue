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
          <span>-</span>
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
            <el-descriptions-item label="dnf.conf_gpgcheck">{{ currentDetail.dnf_conf_gpgcheck }}</el-descriptions-item>
            <el-descriptions-item label="redhat.repo_gpgcheck">{{ currentDetail.redhat_repo_gpgcheck }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 用户账户策略 -->
        <el-tab-pane name="user-policy" label="用户账户策略">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="PASS_MAX_DAYS">{{ currentDetail.pass_max_days }}</el-descriptions-item>
            <el-descriptions-item label="PASS_MIN_DAYS">{{ currentDetail.pass_min_days }}</el-descriptions-item>
            <el-descriptions-item label="PASS_MIN_LEN">{{ currentDetail.pass_min_len }}</el-descriptions-item>
            <el-descriptions-item label="PASS_WARN_AGE">{{ currentDetail.pass_warn_age }}</el-descriptions-item>
            <el-descriptions-item label="INACTIVE">{{ currentDetail.inactive }}</el-descriptions-item>
            <el-descriptions-item label="GID (root)">{{ currentDetail.gid }}</el-descriptions-item>
            <el-descriptions-item label="TMOUT">{{ currentDetail.tmout }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 计划任务配置 -->
        <el-tab-pane name="cron-config" label="计划任务">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="Cron 守护进程">{{ currentDetail.cron }}</el-descriptions-item>
            <el-descriptions-item label="crontab 权限" :span="2">{{ currentDetail.crontab }}</el-descriptions-item>
            <el-descriptions-item label="cron.hourly 权限" :span="2">{{ currentDetail.cron_hourly }}</el-descriptions-item>
            <el-descriptions-item label="cron.daily 权限" :span="2">{{ currentDetail.cron_daily }}</el-descriptions-item>
            <el-descriptions-item label="cron.weekly 权限" :span="2">{{ currentDetail.cron_weekly }}</el-descriptions-item>
            <el-descriptions-item label="cron.monthly 权限" :span="2">{{ currentDetail.cron_monthly }}</el-descriptions-item>
            <el-descriptions-item label="cron.deny" :span="2">{{ currentDetail.cron_deny }}</el-descriptions-item>
            <el-descriptions-item label="at.deny" :span="2">{{ currentDetail.at_deny }}</el-descriptions-item>
            <el-descriptions-item label="cron.allow" :span="2">{{ currentDetail.cron_allow }}</el-descriptions-item>
            <el-descriptions-item label="at.allow" :span="2">{{ currentDetail.at_allow }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- SSH 服务器配置 -->
        <el-tab-pane name="ssh-config" label="SSH 配置">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="sshd_config 权限" :span="2">{{ currentDetail.sshd_config }}</el-descriptions-item>
            <el-descriptions-item label="LogLevel">{{ currentDetail.log_level }}</el-descriptions-item>
            <el-descriptions-item label="X11Forwarding">{{ currentDetail.x11_forwarding }}</el-descriptions-item>
            <el-descriptions-item label="MaxAuthTries">{{ currentDetail.max_auth_tries }}</el-descriptions-item>
            <el-descriptions-item label="IgnoreRhosts">{{ currentDetail.ignore_rhosts }}</el-descriptions-item>
            <el-descriptions-item label="HostbasedAuthentication">{{ currentDetail.hostbased_authentication }}</el-descriptions-item>
            <el-descriptions-item label="PermitRootLogin">{{ currentDetail.permit_root_login }}</el-descriptions-item>
            <el-descriptions-item label="PermitEmptyPasswords">{{ currentDetail.permit_empty_passwords }}</el-descriptions-item>
            <el-descriptions-item label="PermitUserEnvironment">{{ currentDetail.permit_user_environment }}</el-descriptions-item>
            <el-descriptions-item label="ClientAliveInterval">{{ currentDetail.client_alive_interval }}</el-descriptions-item>
            <el-descriptions-item label="ClientAliveCountMax">{{ currentDetail.client_alive_count_max }}</el-descriptions-item>
            <el-descriptions-item label="LoginGraceTime">{{ currentDetail.login_grace_time }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 密码策略与复杂度 -->
        <el-tab-pane name="password-policy" label="密码策略">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="minlen">{{ currentDetail.minlen }}</el-descriptions-item>
            <el-descriptions-item label="minclass">{{ currentDetail.minclass }}</el-descriptions-item>
            <el-descriptions-item label="dcredit">{{ currentDetail.dcredit }}</el-descriptions-item>
            <el-descriptions-item label="ucredit">{{ currentDetail.ucredit }}</el-descriptions-item>
            <el-descriptions-item label="lcredit">{{ currentDetail.lcredit }}</el-descriptions-item>
            <el-descriptions-item label="ocredit">{{ currentDetail.ocredit }}</el-descriptions-item>
            <el-descriptions-item label="password_remember" :span="2">{{ currentDetail.password_remember }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 文件系统权限 -->
        <el-tab-pane name="file-permissions" label="文件权限">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="passwd 权限">{{ currentDetail.passwd }}</el-descriptions-item>
            <el-descriptions-item label="passwd- 权限">{{ currentDetail.passwd_minus }}</el-descriptions-item>
            <el-descriptions-item label="group 权限">{{ currentDetail.group }}</el-descriptions-item>
            <el-descriptions-item label="group- 权限">{{ currentDetail.group_minus }}</el-descriptions-item>
            <el-descriptions-item label="shadow 权限">{{ currentDetail.shadow }}</el-descriptions-item>
            <el-descriptions-item label="shadow- 权限">{{ currentDetail.shadow_minus }}</el-descriptions-item>
            <el-descriptions-item label="gshadow 权限">{{ currentDetail.gshadow }}</el-descriptions-item>
            <el-descriptions-item label="gshadow- 权限">{{ currentDetail.gshadow_minus }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        
        <!-- 加密与时钟同步 -->
        <el-tab-pane name="crypto-sync" label="加密与时钟">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="CryptoPolicies">{{ currentDetail.crypto_policies }}</el-descriptions-item>
            <el-descriptions-item label="NTPServer" :span="2">{{ currentDetail.ntp_server }}</el-descriptions-item>
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
      currentDetail: null
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
        this.currentDetail = res
      }).catch(error => {
        console.error('获取详情失败:', error)
        this.$message.error('获取详情失败')
      }).finally(() => {
        this.loading = false
      })
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
</style>
