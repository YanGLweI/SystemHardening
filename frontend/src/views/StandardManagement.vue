<template>
  <div class="standard-management-container">
    <el-card shadow="never" class="standard-card">
      <!-- 顶部横向 Pill 标签导航 -->
      <div class="standard-tabs">
        <div
          class="tab-pill"
          :class="{ 'active': $route.path === '/standard/linux' }"
          @click="handleNavClick('/standard/linux')"
        >
          <!-- Linux Terminal Icon -->
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="4" width="18" height="16" rx="2" stroke="currentColor" stroke-width="2"/>
            <path d="M7 10L9.5 12.5L7 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M16 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          <span>Linux 标准</span>
        </div>
        <div
          class="tab-pill"
          :class="{ 'active': $route.path === '/standard/windows' }"
          @click="handleNavClick('/standard/windows')"
        >
          <!-- Windows Window Icon -->
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M4 4H11V11H4V4Z" fill="currentColor"/>
            <path d="M13 4H20V11H13V4Z" fill="currentColor" opacity="0.9"/>
            <path d="M4 13H11V20H4V13Z" fill="currentColor" opacity="0.8"/>
            <path d="M13 13H20V20H13V13Z" fill="currentColor" opacity="0.7"/>
          </svg>
          <span>Windows 标准</span>
        </div>
      </div>

      <!-- 内容区 -->
      <div class="standard-content">
        <router-view />
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'StandardManagement',
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
.standard-management-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.standard-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 24px;
  }
}

// 顶部 Pill 标签导航
.standard-tabs {
  flex-shrink: 0;
  display: flex;
  gap: 0;
  padding: 0;
  background: transparent;
}

.tab-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 28px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-regular);
  background: var(--color-gray-100);
  border: 1px solid var(--color-border-light);
  transition: all 0.25s ease;
  user-select: none;

  &:first-child {
    border-radius: 8px 0 0 8px;
  }

  &:last-child {
    border-radius: 0 8px 8px 0;
    border-left: none;
  }

  i {
    font-size: 16px;
    color: var(--color-secondary);
    transition: color 0.25s ease;
  }

  &:hover {
    background: var(--color-primary-alpha-10);
    color: var(--color-primary);

    i {
      color: var(--color-primary);
    }
  }

  &.active {
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
    color: #fff;
    font-weight: 600;
    border-color: var(--color-primary);
    position: relative;
    z-index: 1;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);

    i {
      color: #fff;
    }

    + .tab-pill {
      border-left-color: var(--color-primary);
    }
  }
}

// 内容区
.standard-content {
  flex: 1;
  margin-top: 10px;
  overflow: hidden;
  position: relative;
}
</style>