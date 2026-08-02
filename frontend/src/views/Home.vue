<template>
  <div class="home-container">
    <el-card class="development-card">
      <div class="card-header">
        <i class="el-icon-s-platform"></i>
        <span>系统看板</span>
      </div>
      <div class="card-content">
        <i style="font-size: 80px; margin-bottom: 20px; color: #e6a23c;" class="el-icon-s-platform"></i>
        <h3>欢迎使用系统看板</h3>
        <p>这是系统的主看板页面</p>
        <p>您可以在此查看系统的整体运行状态</p>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Home',
  data() {
    return {}
  }
}
</script>

<style scoped>
.home-container {
  height: 100vh;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.development-card {
  width: 80%;
  max-width: 800px;
}

.card-header {
  display: flex;
  align-items: center;
  color: #303133;
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 30px;
}

.card-header i {
  margin-right: 8px;
}

.card-content {
  padding: 60px 20px;
  text-align: center;
  color: #606266;
}

.card-content h3 {
  font-size: 24px;
  color: #303133;
  margin-bottom: 20px;
  font-weight: 500;
}

.card-content p {
  font-size: 16px;
  margin: 10px 0;
  color: #909399;
}
</style>
