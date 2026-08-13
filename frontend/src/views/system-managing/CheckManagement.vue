<template>
  <div class="check-management-container">
    <!-- 卡片容器 -->
    <el-card shadow="never" class="check-card">
      <!-- 横向标签页 -->
      <div class="check-tabs">
        <div 
          class="tab-pill" 
          :class="{ 'active': $route.path === '/check/linux' }" 
          @click="handleNavClick('/check/linux')"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
            <path d="M7 10L9.5 12.5L7 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M16 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          <span>Linux 加固检查</span>
        </div>
        <div 
          class="tab-pill" 
          :class="{ 'active': $route.path === '/check/windows' }" 
          @click="handleNavClick('/check/windows')"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <path d="M4 4H11V11H4V4Z" fill="currentColor"/>
            <path d="M13 4H20V11H13V4Z" fill="currentColor" opacity="0.9"/>
            <path d="M4 13H11V20H4V13Z" fill="currentColor" opacity="0.8"/>
            <path d="M13 13H20V20H13V13Z" fill="currentColor" opacity="0.7"/>
          </svg>
          <span>Windows 加固检查</span>
        </div>
      </div>
      
      <!-- 路由内容区域 -->
      <div class="check-content">
        <router-view />
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'CheckManagement',
  methods: {
    handleNavClick(path) {
      if (this.$route.path !== path) {
        this.$router.push(path)
      }
    }
  }
}
</script>

<style scoped lang="scss">
.check-management-container {
  max-width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 🟢 卡片容器 */
.check-card {
  background: transparent;
  height: 100%;
  display: flex;
  flex-direction: column;
  
  :deep(.el-card__body) {
    padding: var(--spacing-6);
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
}

/* 🟢 横向标签页 */
.check-tabs {
  display: flex;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
  flex-shrink: 0;
}

.tab-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  color: var(--color-text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: all var(--transition-base);
  white-space: nowrap;
  user-select: none;
  
  &:hover {
    color: var(--color-primary);
    background: var(--color-primary-alpha-10);
  }
  
  &.active {
    color: var(--color-primary);
    background: var(--color-primary-alpha-10);
    font-weight: 600;
  }
}

/* 路由内容区域 */
.check-content {
  flex: 1;
  overflow: hidden;
}

// 响应式设计
@media screen and (max-width: 768px) {
  .check-tabs {
    flex-wrap: wrap;
  }
}
</style>
