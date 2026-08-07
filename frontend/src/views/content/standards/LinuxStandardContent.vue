<template>
  <div class="content-container">
    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="action-title">
        <h2>Linux 标准配置管理</h2>
        <p>定义和维护 Linux 系统的安全合规标准值</p>
      </div>
      <el-button class="add-button" type="primary" @click="handleAdd">
        <i class="el-icon-plus"></i> 添加标准
      </el-button>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <div class="search-left">
        <el-input
          v-model="keyword"
          placeholder="搜索字段名或标签"
          prefix-icon="el-icon-search"
          clearable
          class="search-input"
          @keyup.enter.native="handleSearch"
          @clear="handleSearch"
        ></el-input>
        <el-select
          v-model="selectedGroup"
          placeholder="选择类型分组"
          clearable
          class="group-select"
          @change="handleSearch"
        >
          <el-option
            v-for="group in groupOptions"
            :key="group"
            :label="group"
            :value="group"
          ></el-option>
        </el-select>
        <el-button type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
      </div>
      <el-button icon="el-icon-refresh" @click="handleRefresh">刷新</el-button>
    </div>

    <div class="table-wrapper">
      <el-table :data="sortedStandards" stripe style="width: 100%" :max-height="tableMaxHeight">
        <el-table-column prop="group_name" label="类型" width="120"></el-table-column>
        <el-table-column prop="field_label" label="字段名" width="180"></el-table-column>
        <el-table-column prop="standard_value" label="标准值" min-width="200">
          <template slot-scope="{row}">
            <el-tag v-if="isRegex(row.standard_value)" size="small" type="warning" style="margin-right: 6px; background: #FFFBEB; border-color: #FCD34D; color: #92400E;">正则</el-tag>
            {{ row.standard_value }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250">
          <template slot-scope="{row}">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" @click="handleExempt(row)">例外{{ getExemptCount(row.field_name) ? `(${getExemptCount(row.field_name)})` : '' }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 添加/编辑弹框 -->
    <el-dialog
      :title="isEditMode ? '编辑标准配置' : '添加标准配置'"
      :visible.sync="dialogVisible"
      width="700px"
      max-height="85vh"
      append-to-body
      class="standard-dialog"
    >
      <el-form label-width="120px">
        <div class="standard-row" :class="{ 'highlight': index === 0 }" v-for="(row, index) in rows" :key="index">
          <!-- 标题栏 - 所有行都显示 -->
          <div class="row-header">
            <span class="row-title">标准配置 #{{ index + 1 }}</span>
            <el-button
              size="small"
              type="danger"
              @click="removeRow(index)"
              v-if="rows.length > 1"
            >
              删除此行
            </el-button>
          </div>

          <!-- 字段选择 -->
          <el-form-item
            v-if="!isEditMode"
            label="字段名"
            required
          >
            <el-select
              v-model="row.field_name"
              filterable
              placeholder="请选择字段"
              style="width: 100%"
              @change="handleFieldChange(row, row.field_name)"
            >
              <el-option
                v-for="option in availableFields.filter(f => {
                  // 排除数据库已存在的字段
                  if (existingFieldNames.has(f.field_name)) return false
                  // 排除其他行已选择的字段（但允许当前行显示它）
                  const isUsedInOtherRows = rows.some(r => r.field_name === f.field_name && r !== row)
                  return !isUsedInOtherRows
                })"
                :key="option.field_name"
                :label="`${option.group_name} / ${option.field_label}`"
                :value="option.field_name"
              ></el-option>
            </el-select>
          </el-form-item>

          <!-- 编辑模式显示字段标签 -->
          <el-form-item label="字段名" v-else>
            <el-input v-model="row.field_label" disabled></el-input>
          </el-form-item>

          <!-- 标准值输入 -->
          <el-form-item label="标准值">
            <el-input
              v-model="row.standard_value"
              placeholder="请输入标准值，支持正则格式如 /^\d+$/"
              clearable
            ></el-input>
            <div class="regex-hint">支持正则表达式，格式：/正则模式/，例如 `/^\d+$/` 或 `/^[a-zA-Z0-9.-]+$/</div>
          </el-form-item>
        </div>

        <!-- 添加新行按钮 -->
        <el-form-item v-if="!isEditMode">
          <el-button size="small" type="primary" @click="addRow">
            <i class="el-icon-plus"></i> 添加一行
          </el-button>
        </el-form-item>
      </el-form>

      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false" style="border-color: var(--color-border); color: var(--color-text-regular);">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="handleSubmit"
          style="background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%); border:none;"
        >
          提交
        </el-button>
      </span>
    </el-dialog>

    <!-- 例外客户端穿梭框弹框 -->
    <el-dialog
      :title="`例外设置 - ${currentExemptRow ? currentExemptRow.field_label : ''}`"
      :visible.sync="exemptDialogVisible"
      width="700px"
      append-to-body
      class="transfer-dialog"
    >
      <p class="exempt-tip">已例外的客户端不要求比对该字段的标准值（影响加固检查与邮件报告）</p>
      <el-transfer
        v-model="selectedClientUuids"
        :data="allClients"
        :titles="['可选客户端', '例外客户端']"
        :button-texts="['移除', '添加']"
        filterable
        filter-placeholder="搜索主机名或 IP"
        :props="{ key: 'client_uuid', label: 'label' }"
      ></el-transfer>

      <span slot="footer" class="dialog-footer">
        <el-button @click="exemptDialogVisible = false" style="border-color: var(--color-border); color: var(--color-text-regular);">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="handleSaveExemptions"
          style="background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%); border:none;"
        >
          保存例外
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { createStandards, listStandards, updateStandard, deleteStandard, getAvailableFields, getStandardExemptions, updateStandardExemptions } from '@/api/linux-checks'
import { getClientList } from '@/api/clients'

export default {
  name: 'LinuxStandardContent',
  data() {
    return {
      loading: false,
      allStandards: [], // 原始全量数据
      groupedData: {}, // key: group_name, value: [field records]
      dialogVisible: false,
      isEditMode: false,
      availableFields: [], // 从后端获取的可用字段列表
      rows: [{  // 初始化为包含一条空记录的数组
        field_name: '',
        field_label: '',
        standard_value: '',
        group_name: ''
      }],
      existingFieldNames: new Set(),
      keyword: '',       // 搜索关键词
      selectedGroup: '', // 选中的分组类型
      tableMaxHeight: 500,
      // 例外配置相关
      exemptionList: [],      // 全部例外记录 [{field_name, client_uuid, device_name, ip_address}]
      exemptionCountMap: {},  // field_name -> 例外客户端数量
      exemptDialogVisible: false,
      currentExemptRow: null, // 当前设置例外的标准行
      allClients: [],         // 穿梭框候选客户端 [{client_uuid, label}]
      selectedClientUuids: [] // 已选例外客户端 UUID 数组
    }
  },
  computed: {
    // 所有分组选项（基于原始全量数据）
    groupOptions() {
      const groups = new Set()
      this.allStandards.forEach(item => {
        if (item.group_name) {
          groups.add(item.group_name)
        }
      })
      return Array.from(groups).sort()
    },
    
    // 排序后的所有数据（按 group_name 排序）
    sortedStandards() {
      return [...Object.values(this.groupedData)].flat()
        .sort((a, b) => {
          // 按 group_name 排序
          const groupDiff = a.group_name.localeCompare(b.group_name)
          if (groupDiff !== 0) return groupDiff
          // 同组内按 field_name 排序
          return a.field_name.localeCompare(b.field_name)
        })
    },
    
    // 根据数据库中已有的字段和当前表单已提交的字段计算可用选项
    availableOptions() {
      return this.availableFields.filter(field => !this.existingFieldNames.has(field.field_name))
    },
    
    // 当前行数
    currentRowsLength() {
      return this.rows.length
    },
    
    // 获取当前所有行中已选择的字段名集合
    selectedFieldNames() {
      return new Set(this.rows.map(row => row.field_name).filter(name => name))
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
  watch: {
    selectedGroup() {
      this.applyFilters()
    },
    keyword() {
      this.applyFilters()
    }
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
        // 始终获取全量数据，在前端过滤；同时拉取例外配置
        const [res] = await Promise.all([listStandards({}), this.fetchExemptions()])
        this.allStandards = res || []
        this.applyFilters()
      } catch (error) {
        console.error('获取数据失败:', error)
        this.$message.error('获取数据失败')
      } finally {
        this.loading = false
      }
    },

    // 拉取例外列表并构建字段 -> 数量映射
    async fetchExemptions() {
      try {
        const res = await getStandardExemptions()
        this.exemptionList = res || []
        const map = {}
        this.exemptionList.forEach(item => {
          map[item.field_name] = (map[item.field_name] || 0) + 1
        })
        this.exemptionCountMap = map
      } catch (error) {
        console.error('获取例外列表失败:', error)
      }
    },

    getExemptCount(fieldName) {
      return this.exemptionCountMap[fieldName] || 0
    },

    // 打开例外设置穿梭框
    async handleExempt(row) {
      this.currentExemptRow = row
      try {
        const res = await getClientList({ os_type: 'linux', pageSize: 100 })
        const clients = res.list || res || []
        this.allClients = clients.map(client => ({
          client_uuid: client.client_uuid,
          label: `${client.device_name} (${client.ip_address})`
        }))
        this.selectedClientUuids = this.exemptionList
          .filter(e => e.field_name === row.field_name)
          .map(e => e.client_uuid)
        this.exemptDialogVisible = true
      } catch (error) {
        console.error('获取客户端列表失败:', error)
        this.$message.error('获取客户端列表失败')
      }
    },

    // 保存例外配置（全量替换）
    handleSaveExemptions() {
      if (!this.currentExemptRow) return
      this.loading = true
      updateStandardExemptions(this.currentExemptRow.id, this.selectedClientUuids)
        .then(res => {
          this.$message.success(res.message || '保存成功')
          this.exemptDialogVisible = false
          this.fetchExemptions()
        })
        .catch(error => {
          console.error('保存例外失败:', error)
          this.$message.error(error.response?.data?.error || '保存例外失败')
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    applyFilters() {
      let filtered = this.allStandards
      
      // 按分组过滤
      if (this.selectedGroup) {
        filtered = filtered.filter(item => item.group_name === this.selectedGroup)
      }
      
      // 按关键词过滤
      if (this.keyword) {
        const kw = this.keyword.toLowerCase()
        filtered = filtered.filter(item =>
          item.field_name.toLowerCase().includes(kw) ||
          item.field_label.toLowerCase().includes(kw) ||
          (item.standard_value && item.standard_value.toLowerCase().includes(kw))
        )
      }
      
      // 按 group_name 分组
      const temp = {}
      filtered.forEach(item => {
        if (!temp[item.group_name]) {
          temp[item.group_name] = []
        }
        temp[item.group_name].push(item)
      })
      
      // 每个组内按 field_name 排序
      Object.keys(temp).forEach(key => {
        temp[key].sort((a, b) => a.field_name.localeCompare(b.field_name))
      })
      
      this.groupedData = temp
      this.existingFieldNames.clear()
      Object.values(this.groupedData).flat().forEach(item => {
        this.existingFieldNames.add(item.field_name)
      })
    },
    
    async fetchAvailableFields() {
      try {
        const res = await getAvailableFields()
        this.availableFields = res || []
      } catch (error) {
        console.error('获取可用字段失败:', error)
      }
    },
    
    isRegex(value) {
      return value && value.length >= 2 && value.startsWith('/') && value.endsWith('/')
    },
    async handleAdd() {
      this.isEditMode = false
      this.dialogVisible = true
      
      // 点击添加时才获取可用字段列表
      await this.fetchAvailableFields()
      
      // 初始化第一行数据
      if (this.availableFields && this.availableFields.length > 0) {
        const firstField = this.availableFields[0]
        this.rows = [{
          field_name: firstField.field_name,
          field_label: firstField.field_label,
          standard_value: '',
          group_name: firstField.group_name
        }]
      } else {
        this.rows = [{
          field_name: '',
          field_label: '',
          standard_value: '',
          group_name: ''
        }]
      }
    },
    
    // 当字段改变时自动更新表单中的分组信息
    handleFieldChange(row, fieldName) {
      if (!fieldName) return
      
      // 查找字段的完整信息（直接从 availableFields）
      const field = this.availableFields.find(f => f.field_name === fieldName)
      if (field) {
        row.field_label = field.field_label
        row.group_name = field.group_name
      }
    },

    
    addRow() {
      this.rows.push({
        field_name: '',
        field_label: '',
        standard_value: '',
        group_name: ''
      })
    },
    
    removeRow(index) {
      if (this.rows.length > 1) {
        this.rows.splice(index, 1)
      }
    },
    
    handleSubmit() {
      // 验证：检查是否有空字段或空标准值
      const invalidRows = this.rows.filter(r => !r.field_name || r.standard_value === null || !r.standard_value)
      if (invalidRows.length > 0) {
        this.$message.warning('请填写完整的字段名和标准值')
        return
      }
      
      // 验证重复字段（同一提交中不允许相同字段）
      const fieldSet = new Set()
      for (const row of this.rows) {
        if (fieldSet.has(row.field_name)) {
          this.$message.error('字段不能重复')
          return
        }
        fieldSet.add(row.field_name)
      }
      
      this.loading = true
      
      // 判断是创建还是更新模式
      if (this.isEditMode) {
        // 更新模式：使用 updateStandard，不包含重复检查
        const rowData = this.rows[0]
        const data = {
          field_name: rowData.field_name,
          field_label: rowData.field_label,
          standard_value: rowData.standard_value,
          group_name: rowData.group_name
        }
        updateStandard(rowData.id, data)
          .then(res => {
            this.$message.success('更新成功')
            this.dialogVisible = false
            this.fetchData()
          })
          .catch(error => {
            console.error('更新失败:', error)
            this.$message.error(error.response?.data?.error || '更新失败')
          })
          .finally(() => {
            this.loading = false
          })
      } else {
        // 创建模式：使用 createStandards
        createStandards(this.rows)
          .then(res => {
            this.$message.success(`成功添加 ${res.count} 条标准配置`)
            this.dialogVisible = false
            this.fetchData()
          })
          .catch(error => {
            console.error('提交失败:', error)
            this.$message.error(error.response?.data?.error || '提交失败')
          })
          .finally(() => {
            this.loading = false
          })
      }
    },
    
    async handleEdit(row) {
      this.isEditMode = true
      this.dialogVisible = true
      this.rows = [{ ...row }]  // 只有一条记录用于编辑
    },
    
    async handleDelete(row) {
      this.$confirm(`确定要删除字段「${row.field_label}」的标准配置吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          await deleteStandard(row.id)
          this.$message.success('删除成功')
          this.fetchData()
        } catch (error) {
          console.error('删除失败:', error)
          this.$message.error('删除失败')
        }
      }).catch(() => {})
    },
    
    // 处理搜索
    handleSearch() {
      this.fetchData()
    },
    
    // 刷新数据
    handleRefresh() {
      this.keyword = ''
      this.selectedGroup = ''
      this.fetchData()
    }
  }
}</script>

<style scoped lang="scss">
.content-container {
  background: transparent;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 🟢 操作栏 */
.action-bar {
  flex-shrink: 0;
  padding-bottom: var(--spacing-4);
  padding-left: var(--spacing-2);
  padding-right: var(--spacing-2);
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

/* 🟢 搜索栏 */
.search-bar {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: var(--spacing-3);
  padding-left: var(--spacing-2);
  padding-right: var(--spacing-2);

  .search-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search-input {
    width: 240px;
  }

  .group-select {
    width: 180px;
  }
}

.add-button {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
  border: none;
  height: 40px;
  padding: 0 20px;
  font-size: 14px;
  font-weight: 500;
  border-radius: var(--radius-md);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.25);
  transition: all var(--transition-base);

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }

  &:active {
    transform: translateY(0);
  }
}

/* 🟢 表格容器 */
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

/* 🟢 对话框优化 */
.standard-dialog {
  :deep(.el-dialog) {
    border-radius: var(--radius-xl);
    overflow: hidden;
    max-height: calc(100vh - 100px);
    
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
      }
    }
    
    .el-dialog__body {
      padding: var(--spacing-6);
      max-height: calc(100vh - 200px);
      overflow-y: auto;
    }
    
    .el-dialog__footer {
      border-top: 1px solid var(--color-border-light);
      padding: var(--spacing-4);
    }
  }
}

/* 🟢 表单行样式 */
.standard-row {
  margin-bottom: var(--spacing-6);
  border: 1px solid var(--color-border-light);
  padding: var(--spacing-6);
  border-radius: var(--radius-md);
  background-color: var(--color-bg-page);
  position: relative;
  transition: all var(--transition-base);
  
  &:hover {
    box-shadow: var(--shadow-md);
    border-color: var(--color-primary);
  }
}

.standard-row.highlight {
  border-color: var(--color-primary);
  border-left: 4px solid var(--color-primary);
  background: linear-gradient(to right, var(--color-primary-alpha-10), transparent);
}

.row-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-5);
  padding-bottom: var(--spacing-4);
  border-bottom: 1px solid var(--color-border-light);
}

.row-title {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text-primary);
}

.regex-hint {
  color: var(--color-text-secondary);
  font-size: 12px;
  margin-top: var(--spacing-2);
  line-height: 1.4;
  background: var(--color-primary-alpha-10);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  border-left: 2px solid var(--color-primary);
}

/* 🟢 例外弹框 */
.exempt-tip {
  margin: 0 0 var(--spacing-4) 0;
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-warning-alpha-10, rgba(245, 158, 11, 0.1));
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  border-left: 2px solid #f59e0b;
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  .content-container {
    padding: var(--spacing-4);
  }
  
  .action-bar {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-4);
  }
  
  .add-button {
    width: 100%;
  }
  
  :deep(.el-table) {
    font-size: 12px;
  }
  
  :deep(.el-dialog) {
    max-width: 100% !important;
    margin-top: 10vh !important;
    
    .el-dialog__body {
      padding: var(--spacing-4) !important;
    }
  }
}
</style>
