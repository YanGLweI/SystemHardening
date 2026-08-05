<template>
  <div class="mail-notification-container">
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <!-- Tab 1: 邮件配置 -->
        <el-tab-pane label="邮件配置" name="config">
          <el-form :model="form" label-width="120px" :rules="configRules" ref="configForm">
            <el-form-item label="SMTP 服务器地址" prop="smtp_host">
              <el-input v-model="form.smtp_host" placeholder="例如：smtp.gmail.com"></el-input>
            </el-form-item>
            <el-form-item label="SMTP 端口" prop="smtp_port">
              <el-input-number 
                v-model="form.smtp_port" 
                :min="1" 
                :max="65535" 
                :step="1"
                style="width: 100%;"
              ></el-input-number>
            </el-form-item>
            <el-form-item label="账号" prop="username">
              <el-input v-model="form.username"></el-input>
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input 
                v-model="form.password" 
                type="password" 
                show-password
                placeholder="输入 SMTP 认证密码/授权码"
              ></el-input>
            </el-form-item>
            <el-form-item label="发件人邮箱" prop="from_email">
              <el-input v-model="form.from_email" placeholder="默认为 SMTP 账号"></el-input>
            </el-form-item>
            <el-form-item label="启用服务">
                <el-switch v-model="form.is_enabled"></el-switch>
              </el-form-item>
            <el-form-item>
              <el-button type="primary" icon="el-icon-check" @click="saveConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
          
          <!-- 测试邮件 -->
          <el-divider></el-divider>
          <h4 style="color: #059669; margin: 0 0 15px 0;">📧 测试邮件</h4>
          <div style="display: flex; gap: 12px; align-items: center;">
            <el-input 
              v-model="testRecipient" 
              placeholder="输入收件人邮箱地址" 
              style="flex: 1;"
            ></el-input>
            <el-button type="success" icon="el-icon-s-share" @click="sendTest">发送测试邮件</el-button>
          </div>
        </el-tab-pane>
        
        <!-- Tab 2: 报告计划 -->
        <el-tab-pane label="报告计划" name="schedule">
          <!-- 工具栏 -->
          <div class="action-bar">
            <el-button type="primary" icon="el-icon-plus" @click="openDialog()">新建计划</el-button>
          </div>
          
          <!-- 列表表格 -->
          <el-table 
            :data="scheduleList" 
            v-loading="loading" 
            stripe
            style="width: 100%"
            max-height="600"
          >
            <el-table-column type="index" label="#" width="50"></el-table-column>
            <el-table-column prop="name" label="报告名称" min-width="120"></el-table-column>
            <el-table-column prop="subject" label="邮件主题" min-width="200"></el-table-column>
            <el-table-column label="频率" min-width="120">
              <template slot-scope="{row}">
                {{ formatFrequency(row) }}
              </template>
            </el-table-column>
            <el-table-column prop="send_time" label="发送时间" width="100"></el-table-column>
            <el-table-column prop="recipients" label="收件人" min-width="180"></el-table-column>
            <el-table-column label="状态" width="80">
              <template slot-scope="{row}">
                <el-tag :type="row.is_enabled ? 'success' : 'danger'" size="small">
                  {{ row.is_enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="最近发送" width="180">
              <template slot-scope="{row}">
                <span v-if="row.last_run_at">{{ formatDate(row.last_run_at) }}</span>
                <span v-else style="color: #999;">未发送</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240" fixed="right">
              <template slot-scope="{row}">
                <el-button 
                  size="small" 
                  type="primary"
                  @click="immediateSend(row.id)"
                >
                  立即发送
                </el-button>
                <el-button 
                  size="small" 
                  @click="editSchedule(row)"
                >
                  编辑
                </el-button>
                <el-button 
                  size="small" 
                  type="danger"
                  @click="deleteSchedule(row.id)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      
      <!-- 新建/编辑弹窗 -->
      <el-dialog 
        :title="dialogTitle" 
        :visible.sync="dialogVisible" 
        width="700px"
        append-to-body
      >
        <el-form :model="dialogForm" :rules="dialogRules" label-width="100px" ref="dialogForm">
          <el-form-item label="报告名称" prop="name">
            <el-input v-model="dialogForm.name" placeholder="例如：每周安全合规报告"></el-input>
          </el-form-item>
          <el-form-item label="邮件主题" prop="subject">
            <el-input v-model="dialogForm.subject" :placeholder="'例如：系统加固周报 - ' + currentDate"></el-input>
          </el-form-item>
          <el-form-item label="发送频率" prop="schedule_type">
            <el-select 
              v-model="dialogForm.schedule_type" 
              @change="onScheduleTypeChange"
              style="width: 100%;"
            >
              <el-option label="每日" value="daily"></el-option>
              <el-option label="每几日" value="every_n_days"></el-option>
              <el-option label="每周" value="weekly"></el-option>
              <el-option label="每几周" value="every_n_weeks"></el-option>
              <el-option label="每月" value="monthly"></el-option>
              <el-option label="每几个月" value="every_n_months"></el-option>
            </el-select>
          </el-form-item>
          
          <!-- 动态字段 -->
          <el-form-item 
            label="间隔数量" 
            prop="interval_value"
            v-if="['every_n_days', 'every_n_weeks', 'every_n_months'].includes(dialogForm.schedule_type)"
          >
            <el-input-number 
              v-model="intervalValue" 
              :min="1" 
              :max="365" 
              :step="1"
            ></el-input-number>
            <span style="margin-left: 10px;">
              {{ intervalValue > 1 ? `每${frequencyLabels[dialogForm.schedule_type]}${intervalValue}个` : `每个${frequencyLabels[dialogForm.schedule_type]}` }}
            </span>
          </el-form-item>
          
          <el-form-item 
            label="星期几" 
            prop="weekday"
            v-if="dialogForm.schedule_type === 'weekly'"
          >
            <el-select v-model="dialogForm.weekday" style="width: 100%;">
              <el-option :label="'星期一'" :value="1"></el-option>
              <el-option :label="'星期二'" :value="2"></el-option>
              <el-option :label="'星期三'" :value="3"></el-option>
              <el-option :label="'星期四'" :value="4"></el-option>
              <el-option :label="'星期五'" :value="5"></el-option>
              <el-option :label="'星期六'" :value="6"></el-option>
              <el-option :label="'星期日'" :value="7"></el-option>
            </el-select>
          </el-form-item>
          
          <el-form-item 
            label="每月日期" 
            prop="day_of_month"
            v-if="dialogForm.schedule_type === 'monthly'"
          >
            <el-input-number 
              v-model="dialogForm.day_of_month" 
              :min="1" 
              :max="31"
              style="width: 100%;"
            ></el-input-number>
          </el-form-item>
          
          <el-form-item label="发送时间" prop="send_time">
            <el-time-picker 
              v-model="dialogForm.send_time" 
              format="HH:mm"
              value-format="HH:mm"
              style="width: 100%;"
            ></el-time-picker>
          </el-form-item>
          <el-form-item label="收件人" prop="recipients">
            <el-input 
              v-model="dialogForm.recipients" 
              type="textarea" 
              :rows="3"
              placeholder="多个邮箱用英文逗号分隔，例如：user1@example.com,user2@example.com"
            ></el-input>
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="dialogForm.is_enabled"></el-switch>
          </el-form-item>
        </el-form>
        <span slot="footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitDialog">确定</el-button>
        </span>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import {
  getMailConfig,
  saveMailConfig,
  testEmail,
  listSchedules,
  createSchedule,
  updateSchedule,
  deleteSchedule,
  immediateSend
} from '@/api/mail'

export default {
  name: 'MailNotification',
  data() {
    return {
      activeTab: 'schedule',
      form: {},
      testRecipient: '',
      scheduleList: [],
      loading: false,
      dialogVisible: false,
      dialogType: 'create',
      editId: null,
      dialogForm: {
        name: '',
        subject: '',
        schedule_type: 'daily',
        send_time: '09:00',
        interval_days: 1,
        weekday: 1,
        day_of_month: 1,
        interval_weeks: 1,
        interval_months: 1,
        recipients: '',
        is_enabled: true,
        created_by: '',
        last_updated_by: ''
      },
      intervalValue: 1,
      frequencyLabels: {
        'every_n_days': '日',
        'weekly': '周',
        'every_n_weeks': '周',
        'monthly': '月',
        'every_n_months': '月'
      },
      configRules: {
        smtp_host: [{ required: true, message: '请输入 SMTP 服务器地址', trigger: 'blur' }],
        smtp_port: [{ required: true, message: '请输入 SMTP 端口', trigger: 'blur' }],
        username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
      },
      dialogRules: {
        name: [{ required: true, message: '请输入报告名称', trigger: 'blur' }],
        schedule_type: [{ required: true, message: '请选择发送频率', trigger: 'change' }],
        send_time: [{ required: true, message: '请选择发送时间', trigger: 'change' }],
        recipients: [
          { required: true, message: '请输入收件人', trigger: 'blur' },
          {
            validator: (rule, value, callback) => {
              if (!value) return callback()
              const emails = value.split(/[,，]/).map(e => e.trim()).filter(Boolean)
              const emailRegex = /^[\w.-]+@[\w.-]+\.[a-zA-Z]+$/
              for (const email of emails) {
                if (!emailRegex.test(email)) {
                  return callback(new Error('邮箱格式不正确'))
                }
              }
              callback()
            },
            trigger: 'blur'
          }
        ]
      }
    }
  },
  computed: {
    dialogTitle() {
      return this.dialogType === 'create' ? '新建报告计划' : '编辑报告计划'
    },
    currentDate() {
      const now = new Date()
      return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
    }
  },
  methods: {
    // 加载配置
    loadConfig() {
      getMailConfig().then(res => {
        if (res) {
          this.form = res
        }
      }).catch(() => {
        this.$message.error('获取配置失败')
      })
    },
    
    // 保存配置
    saveConfig() {
      this.$refs.configForm.validate(valid => {
        if (!valid) return
        
        saveMailConfig(this.form).then(() => {
          this.$message.success('配置保存成功')
        }).catch(err => {
          const msg = (err.response && err.response.data && err.response.data.error) || '保存配置失败'
          this.$message.error(msg)
        })
      })
    },
    
    // 测试邮件
    sendTest() {
      if (!this.testRecipient || !this.testRecipient.includes('@')) {
        this.$message.error('请输入有效的邮箱地址')
        return
      }
      
      this.loading = true
      testEmail(this.testRecipient).then(() => {
        this.$message.success({ message: '测试邮件已发送！请检查收件箱', duration: 3000 })
      }).catch(err => {
        const msg = (err.response && err.response.data && err.response.data.error) || '发送测试邮件失败'
        this.$message.error(msg)
      }).finally(() => {
        this.loading = false
      })
    },
    
    // 加载计划列表
    loadSchedules() {
      this.loading = true
      listSchedules().then(res => {
        if (res && res.list) {
          this.scheduleList = res.list
        }
      }).catch(() => {
        this.$message.error('加载计划失败')
      }).finally(() => {
        this.loading = false
      })
    },
    
    // 打开新建/编辑对话框
    openDialog(data = null) {
      this.dialogType = data ? 'edit' : 'create'
      this.editId = data ? data.id : null
      
      if (data) {
        this.dialogForm = JSON.parse(JSON.stringify(data))
        this.intervalValue = this.getIntervalValue(data)
      } else {
        this.resetDialogForm()
      }
      
      this.dialogVisible = true
      this.$nextTick(() => {
        this.updateIntervalInput()
      })
    },
    
    // 提交对话框
    submitDialog() {
      this.$refs.dialogForm.validate(valid => {
        if (!valid) return
        
        const payload = { ...this.dialogForm }
        this.fillIntervals(payload)
        
        const promise = this.dialogType === 'create' 
          ? createSchedule(payload) 
          : updateSchedule(this.editId, payload)
        
        promise.then(() => {
          this.$message.success('保存成功')
          this.loadSchedules()
          this.dialogVisible = false
        }).catch(err => {
          const msg = (err.response && err.response.data && err.response.data.error) || '保存失败'
          this.$message.error(msg)
        })
      })
    },
    
    // 立即发送
    immediateSend(id) {
      this.$confirm('确定要立即发送此报告吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        immediateSend(id).then(() => {
          this.$message.success('报告发送成功')
          this.loadSchedules()
        }).catch(err => {
          const msg = (err.response && err.response.data && err.response.data.error) || '发送报告失败'
          this.$message.error(msg)
        })
      }).catch(() => {})
    },
    
    // 编辑计划
    editSchedule(row) {
      this.openDialog(row)
    },
    
    // 删除计划
    deleteSchedule(id) {
      this.$confirm('确定要删除此计划吗？删除后无法恢复。', '警告', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        deleteSchedule(id).then(() => {
          this.$message.success('删除成功')
          this.loadSchedules()
        }).catch(() => {})
      }).catch(() => {})
    },
    
    // 格式化频率显示
    formatFrequency(row) {
      switch(row.schedule_type) {
        case 'daily': return '每日'
        case 'every_n_days': return `每${row.interval_days || 1}日`
        case 'weekly': return `每${row.weekday || 1}星期`
        case 'every_n_weeks': return `每${row.interval_weeks || 1}周`
        case 'monthly': return `每月 ${row.day_of_month || 1}日`
        case 'every_n_months': return `每${row.interval_months || 1}月`
        default: return '-'
      }
    },
    
    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return ''
      try {
        const date = new Date(dateStr)
        return date.toLocaleString('zh-CN', { 
          year: 'numeric', 
          month: '2-digit', 
          day: '2-digit', 
          hour: '2-digit', 
          minute: '2-digit'
        })
      } catch (e) {
        return dateStr
      }
    },
    
    // 频率类型切换联动
    onScheduleTypeChange() {
      this.updateIntervalInput()
    },
    
    // 重置对话框表单
    resetDialogForm() {
      Object.assign(this.dialogForm, {
        name: '',
        subject: '',
        schedule_type: 'daily',
        send_time: '09:00',
        interval_days: 1,
        weekday: 1,
        day_of_month: 1,
        interval_weeks: 1,
        interval_months: 1,
        recipients: '',
        is_enabled: true,
        created_by: '',
        last_updated_by: ''
      })
      this.intervalValue = 1
    },
    
    // 填充间隔值到 payload
    fillIntervals(payload) {
      switch(payload.schedule_type) {
        case 'every_n_days':
          payload.interval_days = this.intervalValue
          break
        case 'every_n_weeks':
          payload.interval_weeks = this.intervalValue
          break
        case 'every_n_months':
          payload.interval_months = this.intervalValue
          break
      }
    },
    
    // 获取间隔值
    getIntervalValue(row) {
      if (row.schedule_type === 'every_n_days') return row.interval_days || 1
      if (row.schedule_type === 'every_n_weeks') return row.interval_weeks || 1
      if (row.schedule_type === 'every_n_months') return row.interval_months || 1
      return 1
    },
    
    // 更新间隔输入框
    updateIntervalInput() {
      const t = this.dialogForm.schedule_type
      if (['daily', 'weekly', 'monthly'].includes(t)) {
        this.intervalValue = 1
      } else {
        this.intervalValue = this.getIntervalValue(this.dialogForm) || 1
      }
    },
    
    // Tab 切换处理
    handleTabClick(tab) {
      if (tab.name === 'config') {
        this.loadConfig()
      } else if (tab.name === 'schedule') {
        this.loadSchedules()
      }
    }
  },
  created() {
    // 初始化用户信息（如果从 localStorage 获取）
    this.form.created_by = localStorage.getItem('username') || 'admin'
  },
  mounted() {
    this.activeTab === 'schedule' ? this.loadSchedules() : this.loadConfig()
  }
}
</script>

<style scoped lang="scss">
.mail-notification-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  
  .action-bar {
    margin-bottom: var(--spacing-6);
  }
  
  :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 24px;
  }
  
  :deep(.el-tabs__content) {
    flex: 1;
    overflow-y: auto;
    padding-top: 20px;
  }
}
</style>
