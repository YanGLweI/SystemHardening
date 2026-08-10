<template>
  <div class="region-management-container">
    <!-- 卡片容器 -->
    <el-card shadow="never">
      <!-- 操作栏 -->
      <div class="action-bar">
        <div class="action-title">
          <h2>区域管理</h2>
          <p>管理不同区域的客户端分组和关联关系</p>
        </div>
        <el-button class="add-button" type="primary" @click="handleAdd">
          <i class="el-icon-plus"></i> 新建区域
        </el-button>
      </div>

      <!-- 表格容器 -->
      <el-card class="table-card" shadow="never">
        <el-table :data="regionList" stripe style="width: 100%">
          <el-table-column type="index" label="#" width="50" align="center"></el-table-column>
          <el-table-column prop="name" label="区域名" min-width="180">
            <template slot-scope="{row}">
              <span class="region-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="clients" label="区域客户端" min-width="350">
            <template slot-scope="{row}">
              <div v-if="row.clients && row.clients.length > 0" class="client-tags">
                <el-tag
                  v-for="client in row.clients"
                  :key="client.id"
                  size="small"
                  style="margin-right: 6px; margin-bottom: 4px;"
                >
                  {{ client.device_name }} ({{ client.ip_address }})
                </el-tag>
              </div>
              <div v-else class="empty-clients">
                <el-empty :image-size="60" description="未关联客户端"></el-empty>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="300" fixed="right" align="center">
            <template slot-scope="{row, $index}">
              <el-button
                size="small"
                type="warning"
                icon="el-icon-edit"
                @click="handleEdit(row)"
              >
                编辑
              </el-button>
              <el-button
                size="small"
                type="primary"
                icon="el-icon-connection"
                @click="handleAssociate(row)"
              >
                关联
              </el-button>
              <el-button
                size="small"
                type="danger"
                icon="el-icon-delete"
                @click="handleDelete($index, row)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-card>

    <!-- 新建区域 Dialog -->
    <el-dialog
      :title="isEditMode ? '编辑区域' : '新建区域'"
      :visible.sync="dialogVisible"
      width="500px"
      append-to-body
      class="region-dialog"
    >
      <el-form label-width="100px">
        <el-form-item label="区域名称" required>
          <el-input
            v-model="formModel"
            placeholder="请输入区域名称"
            clearable
            :disabled="false"
            autofocus
          ></el-input>
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

    <!-- 关联客户端 Transfer Dialog -->
    <el-dialog
      title="关联客户端"
      :visible.sync="transferDialogVisible"
      width="700px"
      append-to-body
      class="transfer-dialog"
    >
      <el-transfer
        v-model="selectedClientIds"
        :data="allClients"
        :titles="['可选客户端', '已选客户端']"
        :button-texts="['移除', '添加']"
        filterable
        filter-placeholder="搜索主机名或 IP"
        show-checkboxes
        :props="{ key: 'id', label: 'label' }"
      ></el-transfer>

      <span slot="footer" class="dialog-footer">
        <el-button @click="transferDialogVisible = false" style="border-color: var(--color-border); color: var(--color-text-regular);">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="handleSaveAssociation"
          style="background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%); border:none;"
        >
          保存关联
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { listRegions, createRegion, updateRegionClients, deleteRegion, updateRegion } from '@/api/regions'
import { getClientList } from '@/api/clients'
import { formatTime } from '@/utils/index.js'

export default {
  name: 'RegionManagement',
  data() {
    return {
      loading: false,
      regionList: [], // [{id, name, clients: [{id, device_name, ip_address}]}]
      dialogVisible: false,
      isEditMode: false,
      formModel: '', // 区域名称
      transferDialogVisible: false,
      currentRegionId: null,
      allClients: [], // 所有可用客户端 [{id, label: 'hostname(ip)'}]
      selectedClientIds: [] // 已关联的客户端 ID 数组
    }
  },
  created() {
    this.fetchData()
    // 每 60 秒自动刷新一次
    this.timer = setInterval(() => {
      this.fetchData()
    }, 60000)
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
        const res = await listRegions()
        this.regionList = res || []
      } catch (error) {
        console.error('获取数据失败:', error)
        this.$message.error('获取数据失败')
      } finally {
        this.loading = false
      }
    },

    handleAdd() {
      this.isEditMode = false
      this.formModel = ''
      this.dialogVisible = true
    },

    handleSubmit() {
      if (!this.formModel || !this.formModel.trim()) {
        this.$message.warning('请填写区域名称')
        return
      }

      const name = this.formModel.trim()
      this.loading = true

      if (this.isEditMode) {
        // 更新模式
        updateRegion(this.currentRegionId, { name })
          .then(res => {
            this.$message.success(res.message || '更新成功')
            this.dialogVisible = false
            this.fetchData()
          })
          .catch(error => {
            console.error('更新失败:', error)
            const errorMsg = error.response?.data?.error || error.message || '更新失败'
            this.$message.error(errorMsg)
          })
          .finally(() => {
            this.loading = false
          })
      } else {
        // 创建模式
        createRegion({ name })
          .then(res => {
            this.$message.success(res.message || '创建成功')
            this.dialogVisible = false
            this.fetchData()
          })
          .catch(error => {
            console.error('创建失败:', error)
            const errorMsg = error.response?.data?.error || error.message || '创建失败'
            this.$message.error(errorMsg)
          })
          .finally(() => {
            this.loading = false
          })
      }
    },

    handleEdit(row) {
      this.isEditMode = true
      this.currentRegionId = row.id
      this.formModel = row.name
      this.dialogVisible = true
    },

    handleAssociate(region) {
      this.currentRegionId = region.id
      
      // 获取当前已关联的客户端 ID
      const currentClientIds = region.clients ? region.clients.map(c => c.id) : []
      
      // 收集已被其他区域占用的客户端 ID（一个客户端只能属于一个区域）
      const occupiedClientIds = new Set()
      this.regionList.forEach(r => {
        if (r.id !== region.id && r.clients) {
          r.clients.forEach(c => occupiedClientIds.add(c.id))
        }
      })
      
      // 加载所有客户端列表
      this.loading = true
      getClientList()
        .then(res => {
          const clients = res.list || res || []
          
          // 转换为穿梭框所需格式，排除已被其他区域占用的客户端
          this.allClients = clients
            .filter(client => !occupiedClientIds.has(client.id))
            .map(client => ({
              id: client.id,
              label: `${client.device_name} (${client.ip_address})`,
              device_name: client.device_name,
              ip_address: client.ip_address
            }))
          
          this.selectedClientIds = [...currentClientIds]
          this.transferDialogVisible = true
        })
        .catch(error => {
          console.error('获取客户端列表失败:', error)
          this.$message.error('获取客户端列表失败')
        })
        .finally(() => {
          this.loading = false
        })
    },

    handleSaveAssociation() {
      if (this.currentRegionId === null) return
      
      this.loading = true
      updateRegionClients(this.currentRegionId, this.selectedClientIds)
        .then(res => {
          this.$message.success(res.message || '关联成功')
          this.transferDialogVisible = false
          this.fetchData()
        })
        .catch(error => {
          console.error('保存关联失败:', error)
          this.$message.error(error.response?.data?.error || '保存关联失败')
        })
        .finally(() => {
          this.loading = false
        })
    },

    handleDelete(index, row) {
      this.$confirm(
        `确定要删除区域「${row.name}」吗？此操作将同步清理该区域与客户端的关联关系，且不可恢复。`,
        '警告',
        {
          confirmButtonText: '确定删除',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          await deleteRegion(row.id)
          this.$message.success('删除成功')
          this.fetchData()
        } catch (error) {
          console.error('删除失败:', error)
          this.$message.error('删除失败')
        }
      }).catch(() => {})
    },

    formatTime(time) {
      return formatTime(time)
    }
  }
}
</script>

<style scoped lang="scss">
.region-management-container {
  background: var(--color-bg-page);
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

.add-button {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
  border: none;
  height: 44px;
  padding: 0 24px;
  font-size: 15px;
  font-weight: 500;
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  transition: all var(--transition-base);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
  }
  
  &:active {
    transform: translateY(0);
  }
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

.el-table {
  border-radius: var(--radius-lg);
  overflow: hidden;
  font-size: 14px;
}

.el-table__row {
  transition: all var(--transition-base);
}

.region-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text-primary);
}

.client-tags {
  display: flex;
  flex-wrap: wrap;
  max-height: 100px;
  overflow-y: auto;
}

.empty-clients {
  padding: 10px 0;
  color: var(--color-text-secondary);
  text-align: center;
}

/* 🟢 对话框样式 */
.region-dialog,
.transfer-dialog {
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
      }
    }
    
    .el-dialog__body {
      padding: var(--spacing-6);
    }
    
    .el-dialog__footer {
      border-top: 1px solid var(--color-border-light);
      padding: var(--spacing-4);
    }
  }
}

/* 🔄 响应式设计 */
@media screen and (max-width: 768px) {
  .region-management-container {
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
