<template>
  <div class="client-management-container">
    <!-- 卡片容器 -->
    <el-card shadow="never">
      <!-- 操作栏 -->
      <div class="action-bar">
        <div class="action-title">
          <h2>客户端管理</h2>
          <p>管理和监控系统加固客户端的状态和合规性</p>
        </div>
        
        <!-- 操作按钮组 -->
        <div class="action-buttons">
          <el-button 
            size="small" 
            type="primary"
            icon="el-icon-download"
            @click="showDownloadDialog = true"
          >客户端下载
          </el-button>
          <div class="icon-btn clock" title="检查计划" @click="openScheduleDialog()">
            <i class="el-icon-alarm-clock"></i>
          </div>
          <div class="icon-btn gear" title="安装包上传" @click="openUploadDialog()">
            <i class="el-icon-setting"></i>
          </div>
        </div>
      </div>
      
      <!-- 表格卡片 -->
      <el-card class="table-card" shadow="never">
      <!-- 搜索筛选栏 -->
      <div class="filter-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索主机名或 IP"
          prefix-icon="el-icon-search"
          clearable
          class="filter-input"
          @keyup.enter.native="handleSearch"
          @clear="handleSearch"
        ></el-input>
        <el-select
          v-model="statusFilter"
          placeholder="状态"
          clearable
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="全部状态" value=""></el-option>
          <el-option label="在线" value="online"></el-option>
          <el-option label="离线" value="offline"></el-option>
        </el-select>
        <el-select
          v-model="osFilter"
          placeholder="系统"
          clearable
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="全部系统" value=""></el-option>
          <el-option label="Windows" value="windows"></el-option>
          <el-option label="Linux" value="linux"></el-option>
        </el-select>
        <el-button type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
        <el-button icon="el-icon-refresh" @click="resetFilters">重置</el-button>
      </div>
      
      <el-table :data="tableData" v-loading="loading" style="width: 100%">
        <el-table-column type="index" label="#" width="50"></el-table-column>
        <el-table-column prop="device_name" label="主机名" min-width="120"></el-table-column>
        <el-table-column prop="ip_address" label="IP" min-width="130"></el-table-column>
        <el-table-column prop="os_version" label="系统" min-width="180"></el-table-column>
        <el-table-column label="客户端版本" width="100" align="center">
          <template slot-scope="{row}">
            <el-tag size="small" type="info">{{ row.client_version || 'unknown' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100" align="center">
          <template slot-scope="{row}">
            <el-tag 
              :type="row.status === 'online' ? 'success' : 'info'"
              effect="plain"
            >
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后心跳时间" min-width="170">
          <template slot-scope="{row}">
            {{ formatTime(row.last_check_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" align="center">
          <template slot-scope="{row, $index}">
            <el-button 
              size="small" 
              type="danger"
              @click="handleDelete($index, row)"
            >
              删除
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
    </el-card>
    </el-card>
    
    <!-- 下载对话框 -->
    <el-dialog
      title="客户端安装包下载"
      :visible.sync="showDownloadDialog"
      width="720px"
      center-line
      @open="openDownloadDialog()"
      append-to-body
    >
      
      <div class="modal-content">
        <!-- Linux 下载项 -->
        <div class="download-card" v-loading="loadingLinux">
          <div class="card-left">
            <div class="platform-icon linux-bg">
              <i class="el-icon-setting"></i>
            </div>
            <div class="info">
              <h4 class="platform-name">Linux 客户端</h4>
              <p class="platform-desc">适用于各种 Linux 发行版的加固客户端</p>
              <div class="package-meta" v-if="linuxPackageInfo">
                <span class="meta-item version-item" v-if="linuxPackageInfo.version">
                  <i class="el-icon-collection-tag"></i>
                  <span class="version-text">v{{ linuxPackageInfo.version }}</span>
                </span>
                <el-tooltip :content="formatFileSize(linuxPackageInfo.size)" placement="top">
                  <span class="meta-item">
                    <i class="el-icon-document"></i>
                    {{ formatFileSize(linuxPackageInfo.size) }}
                  </span>
                </el-tooltip>
                <span class="meta-item hash-item" v-if="linuxPackageInfo.hash">
                  <i class="el-icon-verification"></i>
                  <span class="hash-text">{{ linuxPackageInfo.hash.substring(0, 16) }}</span>
                  <el-button
                    size="mini"
                    type="text"
                    icon="el-icon-document-copy"
                    class="copy-btn"
                    @click="copyHash(linuxPackageInfo.hash)"
                  ></el-button>
                </span>
              </div>
              <div class="package-meta empty-meta" v-else>
                <span>暂无安装包</span>
              </div>
            </div>
          </div>
          <div class="card-right">
            <el-button
              v-if="linuxPackageInfo && linuxPackageInfo.exists"
              type="primary"
              size="medium"
              class="download-btn"
              :loading="downloadingLinux"
              @click="downloadLinux()"
            >
              下载
            </el-button>
            <el-button
              v-else
              type="default"
              size="medium"
              disabled
            >
              无可用版本
            </el-button>
          </div>
        </div>
        
        <!-- Windows 下载项 -->
        <div class="download-card" v-loading="loadingWindows">
          <div class="card-left">
            <div class="platform-icon windows-bg">
              <i class="el-icon-monitor"></i>
            </div>
            <div class="info">
              <h4 class="platform-name">Windows 客户端</h4>
              <p class="platform-desc">适用于 Windows Server 或 Windows 的加固客户端</p>
              <div class="package-meta" v-if="windowsPackageInfo">
                <span class="meta-item version-item" v-if="windowsPackageInfo.version">
                  <i class="el-icon-collection-tag"></i>
                  <span class="version-text">v{{ windowsPackageInfo.version }}</span>
                </span>
                <el-tooltip :content="formatFileSize(windowsPackageInfo.size)" placement="top">
                  <span class="meta-item">
                    <i class="el-icon-document"></i>
                    {{ formatFileSize(windowsPackageInfo.size) }}
                  </span>
                </el-tooltip>
                <span class="meta-item hash-item" v-if="windowsPackageInfo.hash">
                  <i class="el-icon-verification"></i>
                  <span class="hash-text">{{ windowsPackageInfo.hash.substring(0, 16) }}</span>
                  <el-button
                    size="mini"
                    type="text"
                    icon="el-icon-document-copy"
                    class="copy-btn"
                    @click="copyHash(windowsPackageInfo.hash)"
                  ></el-button>
                </span>
              </div>
              <div class="package-meta empty-meta" v-else>
                <span>暂无安装包</span>
              </div>
            </div>
          </div>
          <div class="card-right">
            <el-button
              v-if="windowsPackageInfo && windowsPackageInfo.exists"
              type="primary"
              size="medium"
              class="download-btn"
              :loading="downloadingWindows"
              @click="downloadWindows()"
            >
              下载
            </el-button>
            <el-button
              v-else
              type="default"
              size="medium"
              disabled
            >
              无可用版本
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>
    
    <!-- 上传对话框 -->
    <el-dialog
      title="安装包上传"
      :visible.sync="showUploadDialog"
      width="600px"
      center-line
      @close="resetUploadForm"
    >
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="安装包类型">
          <el-select v-model="uploadForm.type" placeholder="请选择安装包类型" style="width: 100%">
            <el-option label="Linux 安装包" value="linux"></el-option>
            <el-option label="Windows 安装包" value="windows"></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="安装包文件">
          <el-upload
            ref="upload"
            drag
            action=""
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :before-upload="handleBeforeUpload"
            :limit="1"
            accept=".zip,.exe"
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
            <div class="el-upload__tip" slot="tip">
              支持 .zip 或 .exe 格式，单个文件不超过 200MB<br/>
              <strong>重要提示</strong>: 文件名必须包含版本号，系统将自动从文件名中提取:<br/>
              • Linux: <code>linux-hardening-client_v1.1.0.zip</code><br/>
              • Windows: <code>WindowsHardeningClient_Setup_1.1.0.exe</code><br/>
              <em>系统将验证版本号格式，不符合要求的文件将拒绝上传</em>
            </div>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <div slot="footer" class="dialog-footer">
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="submitUpload">
          开始上传
        </el-button>
      </div>
    </el-dialog>
    
    <!-- 检查计划对话框 -->
    <el-dialog
      title="加固检查计划"
      :visible.sync="showScheduleDialog"
      width="480px"
      @open="loadSchedule()"
      append-to-body
    >
      <el-form :model="scheduleForm" label-width="100px" v-loading="scheduleLoading">
        <el-form-item label="检查频率">
          <el-radio-group v-model="scheduleForm.schedule_type">
            <el-radio label="daily">每天</el-radio>
            <el-radio label="weekly">每周</el-radio>
            <el-radio label="monthly">每月</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="星期" v-if="scheduleForm.schedule_type === 'weekly'">
          <el-select v-model="scheduleForm.weekday" style="width: 100%">
            <el-option
              v-for="(label, idx) in weekdayOptions"
              :key="idx"
              :label="label"
              :value="idx + 1"
            ></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="日期" v-if="scheduleForm.schedule_type === 'monthly'">
          <el-select v-model="scheduleForm.day_of_month" style="width: 100%">
            <el-option
              v-for="d in 31"
              :key="d"
              :label="d + ' 日'"
              :value="d"
            ></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="检查时间">
          <el-select v-model="scheduleForm.check_time" style="width: 100%">
            <el-option
              v-for="t in timeOptions"
              :key="t"
              :label="t"
              :value="t"
            ></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      
      <div slot="footer" class="dialog-footer">
        <el-button @click="showScheduleDialog = false">取消</el-button>
        <el-button type="primary" :loading="scheduleSaving" @click="submitSchedule">
          保存
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getClientList, deleteClient, getCheckSchedule, saveCheckSchedule } from '@/api/clients'
import { formatTime } from '@/utils/index.js'
import { uploadPackage, getPackageInfo, downloadPackage } from '@/api/packages'

export default {
  name: 'ClientManagement',
  data() {
    return {
      loading: false,
      tableData: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      // 搜索筛选
      searchKeyword: '',
      statusFilter: '',
      osFilter: '',
      // 下载对话框
      showDownloadDialog: false,
      loadingLinux: true,
      loadingWindows: true,
      linuxPackageInfo: null,
      windowsPackageInfo: null,
      downloadingLinux: false,
      downloadingWindows: false,
      // 上传对话框
      showUploadDialog: false,
      uploadForm: {
        type: '',
        file: null
      },
      uploadLoading: false,
      // 检查计划对话框
      showScheduleDialog: false,
      scheduleLoading: false,
      scheduleSaving: false,
      scheduleForm: {
        schedule_type: 'daily',
        check_time: '02:00',
        weekday: 1,
        day_of_month: 1
      },
      weekdayOptions: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
    }
  },
  computed: {
    // 半小时粒度的时间选项（00:00 ~ 23:30，共 48 个）
    timeOptions() {
      const options = []
      for (let h = 0; h < 24; h++) {
        const hh = String(h).padStart(2, '0')
        options.push(hh + ':00')
        options.push(hh + ':30')
      }
      return options
    }
  },
  created() {
    this.fetchData()
    // 每 30 秒自动刷新一次
    this.timer = setInterval(() => {
      this.fetchData()
    }, 30000)
  },
  beforeDestroy() {
    if (this.timer) {
      clearInterval(this.timer)
    }
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = {
          page: this.currentPage,
          pageSize: this.pageSize
        }
        if (this.searchKeyword) params.search = this.searchKeyword
        if (this.statusFilter) params.status = this.statusFilter
        if (this.osFilter) params.os_type = this.osFilter
        const res = await getClientList(params)
        if (res.list && res.total !== undefined) {
          this.tableData = res.list
          this.total = res.total
        } else {
          this.$message.error('数据格式错误')
        }
      } catch (error) {
        console.error('获取数据失败:', error)
        // 不显示错误提示，静默处理（避免频繁打扰用户）
      } finally {
        this.loading = false
      }
    },
    
    handleDelete(index, row) {
      this.$confirm(
        `确定要删除客户端"${row.device_name}"吗？此操作将同步删除该客户端关联的所有加固信息，且不可恢复。`,
        '警告',
        {
          confirmButtonText: '确定删除',
          cancelButtonText: '取消',
          type: 'warning',
          center: true
        }
      )
        .then(async () => {
          try {
            await deleteClient(row.id)
            this.$message.success('删除成功')
            
            // 如果当前页只剩一条数据，跳回上一页
            if (this.tableData.length === 1 && this.currentPage > 1) {
              this.currentPage--
            }
            
            this.fetchData()
          } catch (error) {
            console.error('删除失败:', error)
            this.$message.error('删除失败')
          }
        })
        .catch(() => {
          // 取消删除不做任何处理
        })
    },
    
    formatTime(timeStr) {
      if (!timeStr) return '未连接'
      return formatTime(timeStr, 'YYYY-MM-DD HH:mm:ss') || timeStr
    },
    
    handleSearch() {
      this.currentPage = 1
      this.fetchData()
    },
    
    resetFilters() {
      this.searchKeyword = ''
      this.statusFilter = ''
      this.osFilter = ''
      this.currentPage = 1
      this.fetchData()
    },
    
    handleSizeChange(val) {
      this.pageSize = val
      this.fetchData()
    },
    
    handleCurrentChange(val) {
      this.currentPage = val
      this.fetchData()
    },
    
    // 打开下载对话框时获取包信息
    openDownloadDialog() {
      this.showDownloadDialog = true
      this.getLinuxPackageInfo()
      this.getWindowsPackageInfo()
    },
    
    // 获取 Linux 包信息
    async getLinuxPackageInfo() {
      this.loadingLinux = true
      try {
        const res = await getPackageInfo('linux')
        console.log('🔍 Linux 包信息响应:', res)
        if (res && res.exists) {
          this.linuxPackageInfo = res
          console.log('✅ 设置 Linux包信息:', this.linuxPackageInfo)
        } else {
          console.warn('⚠️ Linux 安装包不存在')
          this.linuxPackageInfo = null
        }
      } catch (error) {
        console.error('❌ 获取 Linux 包信息失败:', error)
        this.linuxPackageInfo = null
      } finally {
        this.loadingLinux = false
      }
    },
    
    // 获取 Windows 包信息
    async getWindowsPackageInfo() {
      this.loadingWindows = true
      try {
        const res = await getPackageInfo('windows')
        console.log('🔍 Windows 包信息响应:', res)
        if (res && res.exists) {
          this.windowsPackageInfo = res
          console.log('✅ 设置 Windows 包信息:', this.windowsPackageInfo)
        } else {
          console.warn('⚠️ Windows 安装包不存在')
          this.windowsPackageInfo = null
        }
      } catch (error) {
        console.error('❌ 获取 Windows 包信息失败:', error)
        this.windowsPackageInfo = null
      } finally {
        this.loadingWindows = false
      }
    },
    
    // 格式化文件大小
    formatFileSize(bytes) {
      if (bytes === 0 || bytes === undefined) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },
    
    // 复制哈希值
    copyHash(hash) {
      if (!hash) {
        this.$message.warning('暂无哈希值')
        return
      }
      // 优先使用 Clipboard API，降级使用 execCommand
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(hash).then(() => {
          this.$message.success('已复制到剪贴板')
        }).catch(() => {
          this.fallbackCopy(hash)
        })
      } else {
        this.fallbackCopy(hash)
      }
    },
    // 降级复制方案（兼容非 HTTPS 环境）
    fallbackCopy(text) {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.left = '-9999px'
      textarea.style.top = '-9999px'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.focus()
      textarea.select()
      try {
        const success = document.execCommand('copy')
        if (success) {
          this.$message.success('已复制到剪贴板')
        } else {
          this.$message.error('复制失败，请手动复制')
        }
      } catch (err) {
        this.$message.error('复制失败，请手动复制')
      } finally {
        document.body.removeChild(textarea)
      }
    },
    
    // 下载 Linux 安装包
    async downloadLinux() {
      this.downloadingLinux = true
      try {
        const blob = await downloadPackage('linux')
        this.triggerDownload(blob, 'system-hardening-linux-client.zip')
      } catch (error) {
        console.error('下载 Linux 安装包失败:', error)
        this.$message.error('下载失败')
      } finally {
        this.downloadingLinux = false
      }
    },
    
    // 下载 Windows 安装包
    async downloadWindows() {
      this.downloadingWindows = true
      try {
        const blob = await downloadPackage('windows')
        this.triggerDownload(blob, 'system-hardening-windows-client.exe')
      } catch (error) {
        console.error('下载 Windows 安装包失败:', error)
        this.$message.error('下载失败')
      } finally {
        this.downloadingWindows = false
      }
    },
    
    // 触发浏览器下载
    triggerDownload(blob, filename) {
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', filename)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    },
    
    // 打开上传对话框
    openUploadDialog() {
      this.showUploadDialog = true
      this.uploadLoading = false
    },
    
    // 重置上传表单
    resetUploadForm() {
      this.uploadForm = {
        type: '',
        file: null
      }
      this.$refs.upload && this.$refs.upload.clearFiles()
    },
    
    // 文件改变事件
    handleFileChange(file) {
      this.uploadForm.file = file.raw
    },
    
    // 文件移除事件
    handleFileRemove() {
      this.uploadForm.file = null
    },
    
    // 上传前验证
    handleBeforeUpload(file) {
      const isZip = file.name.endsWith('.zip') || file.name.endsWith('.exe')
      if (!isZip) {
        this.$message.warning('仅支持 zip 或 exe 格式的文件')
        return false
      }
      const maxSize = 200 * 1024 * 1024 // 200MB
      if (file.size > maxSize) {
        this.$message.warning(`文件大小不能超过${maxSize / 1024 / 1024}MB`) 
        return false
      }
      return true
    },
    
    // 提交上传
    async submitUpload() {
      // 验证表单
      if (!this.uploadForm.type) {
        this.$message.warning('请选择安装包类型')
        return
      }
      if (!this.uploadForm.file) {
        this.$message.warning('请选择安装包文件')
        return
      }
      
      this.uploadLoading = true
      try {
        const response = await uploadPackage(this.uploadForm.type, this.uploadForm.file)
        this.$message.success(response.message || '上传成功')
        // 刷新包信息
        this.getLinuxPackageInfo()
        this.getWindowsPackageInfo()
        // 关闭对话框
        this.showUploadDialog = false
        // 重置表单
        this.resetUploadForm()
      } catch (error) {
        console.error('上传失败:', error)
        this.$message.error(error.response?.data?.error || '上传失败')
      } finally {
        this.uploadLoading = false
      }
    },
    
    // 打开检查计划对话框
    openScheduleDialog() {
      this.showScheduleDialog = true
    },
    
    // 加载当前检查计划
    async loadSchedule() {
      this.scheduleLoading = true
      try {
        const res = await getCheckSchedule()
        if (res && res.schedule_type) {
          this.scheduleForm = {
            schedule_type: res.schedule_type,
            check_time: res.check_time || '02:00',
            weekday: res.weekday || 1,
            day_of_month: res.day_of_month || 1
          }
        }
      } catch (error) {
        console.error('加载检查计划失败:', error)
        this.$message.error('加载检查计划失败')
      } finally {
        this.scheduleLoading = false
      }
    },
    
    // 保存检查计划
    async submitSchedule() {
      this.scheduleSaving = true
      try {
        await saveCheckSchedule({
          schedule_type: this.scheduleForm.schedule_type,
          check_time: this.scheduleForm.check_time,
          weekday: this.scheduleForm.weekday,
          day_of_month: this.scheduleForm.day_of_month
        })
        this.$message.success('检查计划保存成功，客户端将在 5 分钟内生效')
        this.showScheduleDialog = false
      } catch (error) {
        console.error('保存检查计划失败:', error)
        this.$message.error(error.response?.data?.error || '保存检查计划失败')
      } finally {
        this.scheduleSaving = false
      }
    },
  }
}
</script>

<style scoped lang="scss">
.client-management-container {
  max-width: 100%;
}

/* 🟢 操作栏 */
.action-bar {
  margin-bottom: var(--spacing-6);
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.action-buttons {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  width: auto;

  .el-button {
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 10px var(--color-primary-alpha-30);
    }
  }
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  cursor: pointer;
  background: none;
  border: none;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);

  i {
    font-size: 20px;
    color: #666;
    transition: color 0.3s ease, transform 0.3s ease;
  }

  &:hover {
    background: var(--color-primary-alpha-10);
    transform: scale(1.1);

    i {
      color: var(--color-primary);
    }
  }

  &.clock:hover i {
    animation: clock-swing 0.6s ease;
    transform-origin: 50% 50%;
  }

  &.gear:hover i {
    transform: rotate(360deg);
    transition: color 0.3s ease, transform 0.6s ease;
  }
}

@keyframes clock-swing {
  0% { transform: rotate(0); }
  25% { transform: rotate(-15deg); }
  50% { transform: rotate(12deg); }
  75% { transform: rotate(-8deg); }
  100% { transform: rotate(0); }
}

/* 🟢 表格容器 */
.table-card {
  background: white;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
  
  &:hover {
    box-shadow: var(--shadow-lg);
  }
}

/* 🔍 搜索筛选栏 */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.filter-input {
  width: 240px;
}

.filter-select {
  width: 140px;
}

.el-table {
  border-radius: var(--radius-lg);
  overflow: hidden;
  font-size: 14px;
}

.el-table__header-wrapper,
.el-table__body-wrapper {
  ::-webkit-scrollbar {
    height: 6px;
  }
}

/* 📥 下载对话框样式 */
.client-download-modal :deep(.el-dialog__header) {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eaf0 100%);
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  padding: 24px 24px 16px 24px;
}

.modal-content {
  padding: 0;
}

.download-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  background: #fff;
  border: 1px solid var(--color-border-light);
  margin-bottom: 12px;
  transition: all 0.2s ease;
  position: relative;
}

.download-card:last-child {
  margin-bottom: 0;
}

.download-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.card-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.platform-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.linux-bg {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.windows-bg {
  background: linear-gradient(135deg, #3b82f6 0%, #10b981 100%);
}

.platform-icon i {
  font-size: 24px;
  color: white;
}

.info {
  flex: 1;
}

.platform-name {
  margin: 0 0 6px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.platform-desc {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.package-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
  background: #f8fafc;
  padding: 6px 10px;
  border-radius: 4px;
}

.meta-item i {
  color: #9ca3af;
  font-size: 14px;
}

.hash-item .hash-text {
  font-family: monospace;
  color: var(--color-primary);
  font-weight: 500;
}

.version-item .version-text {
  color: var(--color-primary);
  font-weight: 600;
}

.copy-btn {
  padding: 2px;
  color: var(--color-primary);
  opacity: 0.7;
  transition: all 0.2s ease;
}

.copy-btn:hover {
  opacity: 1;
  transform: scale(1.1);
}

.empty-meta {
  color: var(--color-text-placeholder);
  background: transparent;
  padding: 0;
}

.card-right {
  flex-shrink: 0;
  min-width: 120px;
}

.download-btn {
  width: 100%;
  padding: 10px 20px;
  font-size: 14px;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
}

.download-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.3);
}

/* 📤 上传对话框样式 */
.upload-section {
  margin-bottom: 24px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: var(--radius-lg);
}

.upload-section .el-radio-group {
  margin-left: 20px;
}

.upload-area {
  padding: 16px 0;
}

.dialog-footer {
  text-align: right;
  padding-top: 20px;
}

/* 🟢 分页样式 */
.pagination {
  margin-top: var(--spacing-6);
  text-align: right;
  padding: var(--spacing-4) 0;
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  .client-management-container {
    padding: var(--spacing-4);
  }
  
  .pagination {
    text-align: center;
  }
  
  :deep(.el-table) {
    font-size: 12px;
  }
  
  /* 移动端优化 */
  .download-card {
    flex-direction: column;
    gap: 16px;
  }
  
  .card-right {
    width: 100%;
  }
  
  .modal-content {
    max-height: 60vh;
    overflow-y: auto;
  }
}
</style>
