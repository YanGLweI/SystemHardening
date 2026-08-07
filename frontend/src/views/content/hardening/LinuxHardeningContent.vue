<template>
  <div class="content-container">
  <!-- 操作栏 -->
  <div class="action-bar">
    <div class="action-title">
      <h2>Linux 加固检查</h2>
      <p>对 Linux 系统进行安全合规检查和评估</p>
    </div>
  </div>
  
  <!-- 搜索栏 -->
  <div class="search-bar">
    <div class="search-left">
      <el-input
        v-model="keyword"
        placeholder="搜索计算机名或 IP"
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
        <el-table-column prop="ip" label="IP" min-width="120"></el-table-column>
        <el-table-column prop="operasystem" label="系统" min-width="200"></el-table-column>
        <el-table-column label="合规状态" min-width="100">
          <template slot-scope="{row}">
            <el-tag :type="row.compliance_status === 'compliant' ? 'success' : 'danger'">
              {{ row.compliance_status === 'compliant' ? '合规' : '不合规' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template slot-scope="{row}">
            <div style="display: inline-flex; align-items: center; gap: 8px;">
              <el-button 
                size="small" 
                type="primary"
                class="trigger-check-btn"
                icon="el-icon-time"
                @click="handleTriggerCheck(row)"
              >
                立即检查
              </el-button>
              
              <el-button 
                size="small" 
                type="primary"
                @click="handleDetail(row)"
              >
                详情
              </el-button>
            </div>
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
            :class="{'non-compliant': isNonCompliant('dnf_conf_gpgcheck') && !isExemptedField('dnf_conf_gpgcheck')}"
          >
            {{ currentDetail.dnf_conf_gpgcheck }}
            <span v-if="isNonCompliant('dnf_conf_gpgcheck') && !isExemptedField('dnf_conf_gpgcheck')" class="standard-hint">
              (标准：{{ getStandardValue('dnf_conf_gpgcheck') }})
            </span>
            <el-tag v-if="isExemptedField('dnf_conf_gpgcheck')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="redhat.repo_gpgcheck"
            :class="{'non-compliant': isNonCompliant('redhat_repo_gpgcheck') && !isExemptedField('redhat_repo_gpgcheck')}"
          >
            {{ currentDetail.redhat_repo_gpgcheck }}
            <span v-if="isNonCompliant('redhat_repo_gpgcheck') && !isExemptedField('redhat_repo_gpgcheck')" class="standard-hint">
              (标准：{{ getStandardValue('redhat_repo_gpgcheck') }})
            </span>
            <el-tag v-if="isExemptedField('redhat_repo_gpgcheck')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 用户账户策略 -->
      <el-tab-pane name="user-policy" label="用户账户策略">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="PASS_MAX_DAYS"
            :class="{'non-compliant': isNonCompliant('pass_max_days') && !isExemptedField('pass_max_days')}"
          >
            {{ currentDetail.pass_max_days }}
            <span v-if="isNonCompliant('pass_max_days') && !isExemptedField('pass_max_days')" class="standard-hint">
              (标准：{{ getStandardValue('pass_max_days') }})
            </span>
            <el-tag v-if="isExemptedField('pass_max_days')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PASS_MIN_DAYS"
            :class="{'non-compliant': isNonCompliant('pass_min_days') && !isExemptedField('pass_min_days')}"
          >
            {{ currentDetail.pass_min_days }}
            <span v-if="isNonCompliant('pass_min_days') && !isExemptedField('pass_min_days')" class="standard-hint">
              (标准：{{ getStandardValue('pass_min_days') }})
            </span>
            <el-tag v-if="isExemptedField('pass_min_days')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PASS_MIN_LEN"
            :class="{'non-compliant': isNonCompliant('pass_min_len') && !isExemptedField('pass_min_len')}"
          >
            {{ currentDetail.pass_min_len }}
            <span v-if="isNonCompliant('pass_min_len') && !isExemptedField('pass_min_len')" class="standard-hint">
              (标准：{{ getStandardValue('pass_min_len') }})
            </span>
            <el-tag v-if="isExemptedField('pass_min_len')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PASS_WARN_AGE"
            :class="{'non-compliant': isNonCompliant('pass_warn_age') && !isExemptedField('pass_warn_age')}"
          >
            {{ currentDetail.pass_warn_age }}
            <span v-if="isNonCompliant('pass_warn_age') && !isExemptedField('pass_warn_age')" class="standard-hint">
              (标准：{{ getStandardValue('pass_warn_age') }})
            </span>
            <el-tag v-if="isExemptedField('pass_warn_age')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="INACTIVE"
            :class="{'non-compliant': isNonCompliant('inactive') && !isExemptedField('inactive')}"
          >
            {{ currentDetail.inactive }}
            <span v-if="isNonCompliant('inactive') && !isExemptedField('inactive')" class="standard-hint">
              (标准：{{ getStandardValue('inactive') }})
            </span>
            <el-tag v-if="isExemptedField('inactive')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="GID (root)"
            :class="{'non-compliant': isNonCompliant('gid') && !isExemptedField('gid')}"
          >
            {{ currentDetail.gid }}
            <span v-if="isNonCompliant('gid') && !isExemptedField('gid')" class="standard-hint">
              (标准：{{ getStandardValue('gid') }})
            </span>
            <el-tag v-if="isExemptedField('gid')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="TMOUT"
            :class="{'non-compliant': isNonCompliant('tmout') && !isExemptedField('tmout')}"
          >
            {{ currentDetail.tmout }}
            <span v-if="isNonCompliant('tmout') && !isExemptedField('tmout')" class="standard-hint">
              (标准：{{ getStandardValue('tmout') }})
            </span>
            <el-tag v-if="isExemptedField('tmout')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 计划任务配置 -->
      <el-tab-pane name="cron-config" label="计划任务">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="Cron 守护进程"
            :class="{'non-compliant': isNonCompliant('cron') && !isExemptedField('cron')}"
          >
            {{ currentDetail.cron }}
            <span v-if="isNonCompliant('cron') && !isExemptedField('cron')" class="standard-hint">
              (标准：{{ getStandardValue('cron') }})
            </span>
            <el-tag v-if="isExemptedField('cron')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="crontab 权限"
            :class="{'non-compliant': isNonCompliant('crontab') && !isExemptedField('crontab')}"
            :span="2"
          >
            {{ currentDetail.crontab }}
            <span v-if="isNonCompliant('crontab') && !isExemptedField('crontab')" class="standard-hint">
              (标准：{{ getStandardValue('crontab') }})
            </span>
            <el-tag v-if="isExemptedField('crontab')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.hourly 权限"
            :class="{'non-compliant': isNonCompliant('cron_hourly') && !isExemptedField('cron_hourly')}"
            :span="2"
          >
            {{ currentDetail.cron_hourly }}
            <span v-if="isNonCompliant('cron_hourly') && !isExemptedField('cron_hourly')" class="standard-hint">
              (标准：{{ getStandardValue('cron_hourly') }})
            </span>
            <el-tag v-if="isExemptedField('cron_hourly')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.daily 权限"
            :class="{'non-compliant': isNonCompliant('cron_daily') && !isExemptedField('cron_daily')}"
            :span="2"
          >
            {{ currentDetail.cron_daily }}
            <span v-if="isNonCompliant('cron_daily') && !isExemptedField('cron_daily')" class="standard-hint">
              (标准：{{ getStandardValue('cron_daily') }})
            </span>
            <el-tag v-if="isExemptedField('cron_daily')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.weekly 权限"
            :class="{'non-compliant': isNonCompliant('cron_weekly') && !isExemptedField('cron_weekly')}"
            :span="2"
          >
            {{ currentDetail.cron_weekly }}
            <span v-if="isNonCompliant('cron_weekly') && !isExemptedField('cron_weekly')" class="standard-hint">
              (标准：{{ getStandardValue('cron_weekly') }})
            </span>
            <el-tag v-if="isExemptedField('cron_weekly')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.monthly 权限"
            :class="{'non-compliant': isNonCompliant('cron_monthly') && !isExemptedField('cron_monthly')}"
            :span="2"
          >
            {{ currentDetail.cron_monthly }}
            <span v-if="isNonCompliant('cron_monthly') && !isExemptedField('cron_monthly')" class="standard-hint">
              (标准：{{ getStandardValue('cron_monthly') }})
            </span>
            <el-tag v-if="isExemptedField('cron_monthly')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.deny"
            :class="{'non-compliant': isNonCompliant('cron_deny') && !isExemptedField('cron_deny')}"
            :span="2"
          >
            {{ currentDetail.cron_deny }}
            <span v-if="isNonCompliant('cron_deny') && !isExemptedField('cron_deny')" class="standard-hint">
              (标准：{{ getStandardValue('cron_deny') }})
            </span>
            <el-tag v-if="isExemptedField('cron_deny')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="at.deny"
            :class="{'non-compliant': isNonCompliant('at_deny') && !isExemptedField('at_deny')}"
            :span="2"
          >
            {{ currentDetail.at_deny }}
            <span v-if="isNonCompliant('at_deny') && !isExemptedField('at_deny')" class="standard-hint">
              (标准：{{ getStandardValue('at_deny') }})
            </span>
            <el-tag v-if="isExemptedField('at_deny')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="cron.allow"
            :class="{'non-compliant': isNonCompliant('cron_allow') && !isExemptedField('cron_allow')}"
            :span="2"
          >
            {{ currentDetail.cron_allow }}
            <span v-if="isNonCompliant('cron_allow') && !isExemptedField('cron_allow')" class="standard-hint">
              (标准：{{ getStandardValue('cron_allow') }})
            </span>
            <el-tag v-if="isExemptedField('cron_allow')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="at.allow"
            :class="{'non-compliant': isNonCompliant('at_allow') && !isExemptedField('at_allow')}"
            :span="2"
          >
            {{ currentDetail.at_allow }}
            <span v-if="isNonCompliant('at_allow') && !isExemptedField('at_allow')" class="standard-hint">
              (标准：{{ getStandardValue('at_allow') }})
            </span>
            <el-tag v-if="isExemptedField('at_allow')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- SSH 服务器配置 -->
      <el-tab-pane name="ssh-config" label="SSH 配置">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="sshd_config 权限"
            :class="{'non-compliant': isNonCompliant('sshd_config') && !isExemptedField('sshd_config')}"
            :span="2"
          >
            {{ currentDetail.sshd_config }}
            <span v-if="isNonCompliant('sshd_config') && !isExemptedField('sshd_config')" class="standard-hint">
              (标准：{{ getStandardValue('sshd_config') }})
            </span>
            <el-tag v-if="isExemptedField('sshd_config')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="LogLevel"
            :class="{'non-compliant': isNonCompliant('log_level') && !isExemptedField('log_level')}"
          >
            {{ currentDetail.log_level }}
            <span v-if="isNonCompliant('log_level') && !isExemptedField('log_level')" class="standard-hint">
              (标准：{{ getStandardValue('log_level') }})
            </span>
            <el-tag v-if="isExemptedField('log_level')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="X11Forwarding"
            :class="{'non-compliant': isNonCompliant('x11_forwarding') && !isExemptedField('x11_forwarding')}"
          >
            {{ currentDetail.x11_forwarding }}
            <span v-if="isNonCompliant('x11_forwarding') && !isExemptedField('x11_forwarding')" class="standard-hint">
              (标准：{{ getStandardValue('x11_forwarding') }})
            </span>
            <el-tag v-if="isExemptedField('x11_forwarding')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="MaxAuthTries"
            :class="{'non-compliant': isNonCompliant('max_auth_tries') && !isExemptedField('max_auth_tries')}"
          >
            {{ currentDetail.max_auth_tries }}
            <span v-if="isNonCompliant('max_auth_tries') && !isExemptedField('max_auth_tries')" class="standard-hint">
              (标准：{{ getStandardValue('max_auth_tries') }})
            </span>
            <el-tag v-if="isExemptedField('max_auth_tries')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="IgnoreRhosts"
            :class="{'non-compliant': isNonCompliant('ignore_rhosts') && !isExemptedField('ignore_rhosts')}"
          >
            {{ currentDetail.ignore_rhosts }}
            <span v-if="isNonCompliant('ignore_rhosts') && !isExemptedField('ignore_rhosts')" class="standard-hint">
              (标准：{{ getStandardValue('ignore_rhosts') }})
            </span>
            <el-tag v-if="isExemptedField('ignore_rhosts')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="HostbasedAuthentication"
            :class="{'non-compliant': isNonCompliant('hostbased_authentication') && !isExemptedField('hostbased_authentication')}"
          >
            {{ currentDetail.hostbased_authentication }}
            <span v-if="isNonCompliant('hostbased_authentication') && !isExemptedField('hostbased_authentication')" class="standard-hint">
              (标准：{{ getStandardValue('hostbased_authentication') }})
            </span>
            <el-tag v-if="isExemptedField('hostbased_authentication')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PermitRootLogin"
            :class="{'non-compliant': isNonCompliant('permit_root_login') && !isExemptedField('permit_root_login')}"
          >
            {{ currentDetail.permit_root_login }}
            <span v-if="isNonCompliant('permit_root_login') && !isExemptedField('permit_root_login')" class="standard-hint">
              (标准：{{ getStandardValue('permit_root_login') }})
            </span>
            <el-tag v-if="isExemptedField('permit_root_login')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PermitEmptyPasswords"
            :class="{'non-compliant': isNonCompliant('permit_empty_passwords') && !isExemptedField('permit_empty_passwords')}"
          >
            {{ currentDetail.permit_empty_passwords }}
            <span v-if="isNonCompliant('permit_empty_passwords') && !isExemptedField('permit_empty_passwords')" class="standard-hint">
              (标准：{{ getStandardValue('permit_empty_passwords') }})
            </span>
            <el-tag v-if="isExemptedField('permit_empty_passwords')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="PermitUserEnvironment"
            :class="{'non-compliant': isNonCompliant('permit_user_environment') && !isExemptedField('permit_user_environment')}"
          >
            {{ currentDetail.permit_user_environment }}
            <span v-if="isNonCompliant('permit_user_environment') && !isExemptedField('permit_user_environment')" class="standard-hint">
              (标准：{{ getStandardValue('permit_user_environment') }})
            </span>
            <el-tag v-if="isExemptedField('permit_user_environment')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="ClientAliveInterval"
            :class="{'non-compliant': isNonCompliant('client_alive_interval') && !isExemptedField('client_alive_interval')}"
          >
            {{ currentDetail.client_alive_interval }}
            <span v-if="isNonCompliant('client_alive_interval') && !isExemptedField('client_alive_interval')" class="standard-hint">
              (标准：{{ getStandardValue('client_alive_interval') }})
            </span>
            <el-tag v-if="isExemptedField('client_alive_interval')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="ClientAliveCountMax"
            :class="{'non-compliant': isNonCompliant('client_alive_count_max') && !isExemptedField('client_alive_count_max')}"
          >
            {{ currentDetail.client_alive_count_max }}
            <span v-if="isNonCompliant('client_alive_count_max') && !isExemptedField('client_alive_count_max')" class="standard-hint">
              (标准：{{ getStandardValue('client_alive_count_max') }})
            </span>
            <el-tag v-if="isExemptedField('client_alive_count_max')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="LoginGraceTime"
            :class="{'non-compliant': isNonCompliant('login_grace_time') && !isExemptedField('login_grace_time')}"
          >
            {{ currentDetail.login_grace_time }}
            <span v-if="isNonCompliant('login_grace_time') && !isExemptedField('login_grace_time')" class="standard-hint">
              (标准：{{ getStandardValue('login_grace_time') }})
            </span>
            <el-tag v-if="isExemptedField('login_grace_time')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 密码策略与复杂度 -->
      <el-tab-pane name="password-policy" label="密码策略">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="minlen"
            :class="{'non-compliant': isNonCompliant('minlen') && !isExemptedField('minlen')}"
          >
            {{ currentDetail.minlen }}
            <span v-if="isNonCompliant('minlen') && !isExemptedField('minlen')" class="standard-hint">
              (标准：{{ getStandardValue('minlen') }})
            </span>
            <el-tag v-if="isExemptedField('minlen')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="minclass"
            :class="{'non-compliant': isNonCompliant('minclass') && !isExemptedField('minclass')}"
          >
            {{ currentDetail.minclass }}
            <span v-if="isNonCompliant('minclass') && !isExemptedField('minclass')" class="standard-hint">
              (标准：{{ getStandardValue('minclass') }})
            </span>
            <el-tag v-if="isExemptedField('minclass')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="dcredit"
            :class="{'non-compliant': isNonCompliant('dcredit') && !isExemptedField('dcredit')}"
          >
            {{ currentDetail.dcredit }}
            <span v-if="isNonCompliant('dcredit') && !isExemptedField('dcredit')" class="standard-hint">
              (标准：{{ getStandardValue('dcredit') }})
            </span>
            <el-tag v-if="isExemptedField('dcredit')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="ucredit"
            :class="{'non-compliant': isNonCompliant('ucredit') && !isExemptedField('ucredit')}"
          >
            {{ currentDetail.ucredit }}
            <span v-if="isNonCompliant('ucredit') && !isExemptedField('ucredit')" class="standard-hint">
              (标准：{{ getStandardValue('ucredit') }})
            </span>
            <el-tag v-if="isExemptedField('ucredit')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="lcredit"
            :class="{'non-compliant': isNonCompliant('lcredit') && !isExemptedField('lcredit')}"
          >
            {{ currentDetail.lcredit }}
            <span v-if="isNonCompliant('lcredit') && !isExemptedField('lcredit')" class="standard-hint">
              (标准：{{ getStandardValue('lcredit') }})
            </span>
            <el-tag v-if="isExemptedField('lcredit')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="ocredit"
            :class="{'non-compliant': isNonCompliant('ocredit') && !isExemptedField('ocredit')}"
          >
            {{ currentDetail.ocredit }}
            <span v-if="isNonCompliant('ocredit') && !isExemptedField('ocredit')" class="standard-hint">
              (标准：{{ getStandardValue('ocredit') }})
            </span>
            <el-tag v-if="isExemptedField('ocredit')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="password_remember"
            :class="{'non-compliant': isNonCompliant('password_remember') && !isExemptedField('password_remember')}"
            :span="2"
          >
            {{ currentDetail.password_remember }}
            <span v-if="isNonCompliant('password_remember') && !isExemptedField('password_remember')" class="standard-hint">
              (标准：{{ getStandardValue('password_remember') }})
            </span>
            <el-tag v-if="isExemptedField('password_remember')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 文件系统权限 -->
      <el-tab-pane name="file-permissions" label="文件权限">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="passwd 权限"
            :class="{'non-compliant': isNonCompliant('passwd') && !isExemptedField('passwd')}"
          >
            {{ currentDetail.passwd }}
            <span v-if="isNonCompliant('passwd') && !isExemptedField('passwd')" class="standard-hint">
              (标准：{{ getStandardValue('passwd') }})
            </span>
            <el-tag v-if="isExemptedField('passwd')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="passwd- 权限"
            :class="{'non-compliant': isNonCompliant('passwd_minus') && !isExemptedField('passwd_minus')}"
          >
            {{ currentDetail.passwd_minus }}
            <span v-if="isNonCompliant('passwd_minus') && !isExemptedField('passwd_minus')" class="standard-hint">
              (标准：{{ getStandardValue('passwd_minus') }})
            </span>
            <el-tag v-if="isExemptedField('passwd_minus')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="group 权限"
            :class="{'non-compliant': isNonCompliant('group') && !isExemptedField('group')}"
          >
            {{ currentDetail.group }}
            <span v-if="isNonCompliant('group') && !isExemptedField('group')" class="standard-hint">
              (标准：{{ getStandardValue('group') }})
            </span>
            <el-tag v-if="isExemptedField('group')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="group- 权限"
            :class="{'non-compliant': isNonCompliant('group_minus') && !isExemptedField('group_minus')}"
          >
            {{ currentDetail.group_minus }}
            <span v-if="isNonCompliant('group_minus') && !isExemptedField('group_minus')" class="standard-hint">
              (标准：{{ getStandardValue('group_minus') }})
            </span>
            <el-tag v-if="isExemptedField('group_minus')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="shadow 权限"
            :class="{'non-compliant': isNonCompliant('shadow') && !isExemptedField('shadow')}"
          >
            {{ currentDetail.shadow }}
            <span v-if="isNonCompliant('shadow') && !isExemptedField('shadow')" class="standard-hint">
              (标准：{{ getStandardValue('shadow') }})
            </span>
            <el-tag v-if="isExemptedField('shadow')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="shadow- 权限"
            :class="{'non-compliant': isNonCompliant('shadow_minus') && !isExemptedField('shadow_minus')}"
          >
            {{ currentDetail.shadow_minus }}
            <span v-if="isNonCompliant('shadow_minus') && !isExemptedField('shadow_minus')" class="standard-hint">
              (标准：{{ getStandardValue('shadow_minus') }})
            </span>
            <el-tag v-if="isExemptedField('shadow_minus')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="gshadow 权限"
            :class="{'non-compliant': isNonCompliant('gshadow') && !isExemptedField('gshadow')}"
          >
            {{ currentDetail.gshadow }}
            <span v-if="isNonCompliant('gshadow') && !isExemptedField('gshadow')" class="standard-hint">
              (标准：{{ getStandardValue('gshadow') }})
            </span>
            <el-tag v-if="isExemptedField('gshadow')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="gshadow- 权限"
            :class="{'non-compliant': isNonCompliant('gshadow_minus') && !isExemptedField('gshadow_minus')}"
          >
            {{ currentDetail.gshadow_minus }}
            <span v-if="isNonCompliant('gshadow_minus') && !isExemptedField('gshadow_minus')" class="standard-hint">
              (标准：{{ getStandardValue('gshadow_minus') }})
            </span>
            <el-tag v-if="isExemptedField('gshadow_minus')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <!-- 加密与时钟同步 -->
      <el-tab-pane name="crypto-sync" label="加密与时钟">
        <el-descriptions :column="2" border>
          <el-descriptions-item 
            label="CryptoPolicies"
            :class="{'non-compliant': isNonCompliant('crypto_policies') && !isExemptedField('crypto_policies')}"
            :span="2"
          >
            {{ currentDetail.crypto_policies }}
            <span v-if="isNonCompliant('crypto_policies') && !isExemptedField('crypto_policies')" class="standard-hint">
              (标准：{{ getStandardValue('crypto_policies') }})
            </span>
            <el-tag v-if="isExemptedField('crypto_policies')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
          <el-descriptions-item 
            label="NTPServer"
            :class="{'non-compliant': isNonCompliant('ntp_server') && !isExemptedField('ntp_server')}"
            :span="2"
          >
            {{ currentDetail.ntp_server }}
            <span v-if="isNonCompliant('ntp_server') && !isExemptedField('ntp_server')" class="standard-hint">
              (标准：{{ getStandardValue('ntp_server') }})
            </span>
            <el-tag v-if="isExemptedField('ntp_server')" size="mini" type="warning" class="exempt-tag">例外</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
  
  <!-- 立即检查对话框 -->
  <check-trigger-dialog 
    v-show="checkDialogVisible"
    :visible.sync="checkDialogVisible"
    :client="currentClient"
    :task-id.sync="taskId"
  />
  </div>
</template>

<script>
import { getList, getDetail } from '@/api/linux-checks'
import CheckTriggerDialog from '@/components/CheckTriggerDialog.vue'
import { triggerCheck, getClientLatestTask } from '@/api/task-check'

export default {
  name: 'LinuxHardeningContent',
  components: {
    CheckTriggerDialog
  },
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
      tableMaxHeight: 500,
      // 立即检查相关
      checkDialogVisible: false,
      currentClient: {},
      taskId: null // 当前任务的 ID
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
    // 检查字段是否被例外豁免
    isExemptedField(fieldName) {
      if (!this.complianceData || !this.complianceData.exempted_fields) {
        return false
      }
      return this.complianceData.exempted_fields.includes(fieldName)
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
    
    // 触发立即检查
    async handleTriggerCheck(row) {
      try {
        // 先检查该客户端是否有正在执行的任务
        try {
          const taskRes = await getClientLatestTask(row.client_uuid)
          if (taskRes && ['pending', 'sent', 'executing'].includes(taskRes.status)) {
            // 有进行中的任务，直接显示弹窗查看状态，不重复创建
            this.currentClient = row
            // ✅ 先设置 taskId，再打开弹窗
            this.taskId = taskRes.task_id
            this.checkDialogVisible = true
            return
          }
        } catch (error) {
          // 如果查询失败，继续尝试创建任务（可能是前端路由问题）
          console.warn('查询任务状态失败，将尝试创建新任务:', error)
        }
        
        // 没有任务或任务已完成，创建新任务
        const res = await triggerCheck(row.client_uuid)
        
        this.currentClient = row
        // ✅ 先设置 taskId，再打开弹窗
        this.taskId = res.task_id
        this.checkDialogVisible = true
        
        // 对话框打开时会自动开始轮询任务状态
      } catch (error) {
        if (error.response && error.response.status === 409) {
          this.$message.warning(error.response.data.error)
        } else {
          this.$message.error(error.response.data.error || '操作失败')
        }
      }
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

.compliant .el-descriptions-item__content {
  color: var(--color-success);
  font-weight: 500;
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

/* 🟢 例外标签 */
.exempt-tag {
  margin-left: 8px;
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
  
  .el-table__header-wrapper,
  .el-table__body-wrapper {
    ::-webkit-scrollbar {
      height: 6px;
    }
  }
}

/* 🟢 立即检查按钮样式 */
.trigger-check-btn {
  background: var(--color-primary-alpha-10);
  border: 1px solid var(--color-primary);
  color: var(--color-primary);
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  cursor: pointer;

  &:hover {
    background: var(--color-primary);
    color: white;
  }

  &:active {
    transform: scale(0.98);
  }

  i {
    margin-right: 4px;
    font-size: 14px;
  }
}

/* 🟢 详情按钮样式 */
.detail-btn {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);

  &:hover {
    border-color: var(--color-info);
    color: var(--color-info);
  }
}
</style>
