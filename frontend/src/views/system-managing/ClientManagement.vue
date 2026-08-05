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
      </div>
      
      <!-- 表格卡片 -->
      <el-card class="table-card" shadow="never">
      <el-table :data="tableData" v-loading="loading" style="width: 100%">
        <el-table-column type="index" label="#" width="50"></el-table-column>
        <el-table-column prop="device_name" label="主机名" min-width="120"></el-table-column>
        <el-table-column prop="ip_address" label="IP" min-width="130"></el-table-column>
        <el-table-column prop="os_version" label="系统" min-width="180"></el-table-column>
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
  </div>
</template>

<script>
import { getClientList, deleteClient } from '@/api/clients'
import { formatTime } from '@/utils/index.js'

export default {
  name: 'ClientManagement',
  data() {
    return {
      loading: false,
      tableData: [],
      currentPage: 1,
      pageSize: 20,
      total: 0
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
        const res = await getClientList({
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

<style scoped lang="scss">
.client-management-container {
  max-width: 100%;
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

.el-table__header-wrapper,
.el-table__body-wrapper {
  ::-webkit-scrollbar {
    height: 6px;
  }
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
}
</style>
