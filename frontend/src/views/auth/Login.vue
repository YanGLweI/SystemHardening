<template>
  <div class="login-container">
    <!-- 入场动画背景 -->
    <div class="animated-background">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>
    
    <!-- 登录卡片 - 入场动画 -->
    <transition 
      name="login-card-enter"
      enter-active-class="animate__animated animate__fadeInUp"
      leave-active-class="animate__animated animate__fadeOutDown"
    >
      <el-card 
        v-if="!loading"
        class="login-card"
      >
        <!-- 自定义表头 -->
        <div class="custom-header">
          <h2 class="title">系统加固平台</h2>
        </div>
        
        <div class="form-content">
          <el-form 
            ref="loginForm" 
            :model="loginForm" 
            :rules="validationRules"
            label-position="top"
            size="medium"
          >
            <!-- 用户名输入框 -->
            <el-form-item prop="username" label="用户名">
              <el-input
                ref="usernameInput"
                v-model="loginForm.username"
                placeholder="请输入您的域账号"
                prefix-icon="el-icon-user"
                clearable
                @keyup.enter.native="handleLogin"
              />
            </el-form-item>
            
            <!-- 密码输入框 -->
            <el-form-item prop="password" label="密码">
              <el-input
                ref="passwordInput"
                v-model="loginForm.password"
                type="password"
                :placeholder="passwordPlaceholder"
                prefix-icon="el-icon-lock"
                show-password
                @keyup.enter.native="handleLogin"
              />
            </el-form-item>
            
            <!-- 登录按钮 -->
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                class="login-btn"
                :loading="submitting"
                @click="handleLogin"
              >
                <i v-if="!submitting" class="el-icon-key"></i>
                <span>{{ submitting ? '正在认证...' : '安全登录' }}</span>
              </el-button>
            </el-form-item>
          </el-form>
          
          <!-- 辅助信息 -->
          <div class="help-info">
            <el-divider></el-divider>
            <div class="info-row">
              <i class="el-icon-info"></i>
              <span>如有登录问题请联系 IT 部</span>
            </div>
            <div class="security-tip">
              <i class="el-icon-s-security"></i>
              采用 LDAPS 加密传输，保障账号安全
            </div>
          </div>
        </div>
      </el-card>
    </transition>
    
    <!-- 加载中全屏遮罩 -->
    <transition name="fade-in">
      <div v-if="submitting" class="loading-overlay">
        <div class="loader-container">
          <div class="ldap-animation">
            <div class="secure-badge">
              <i class="el-icon-s-security"></i>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'Login',
  data() {
    return {
      loading: true,
      submitting: false,
      passwordPlaceholder: '请输入密码',
      loginForm: {
        username: '',
        password: ''
      },
      validationRules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' },
          { min: 3, max: 50, message: '长度在 3 到 50 个字符', trigger: 'blur' },
          { pattern: /^[a-zA-Z0-9@._-]+$/, message: '用户名格式不正确', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, max: 100, message: '密码至少 6 个字符', trigger: 'blur' }
        ]
      }
    }
  },
  mounted() {
    // 模拟加载动画，卡片渲染后自动聚焦用户名输入框
    setTimeout(() => {
      this.loading = false
      this.$nextTick(() => {
        this.$refs.usernameInput && this.$refs.usernameInput.focus()
      })
    }, 800)
  },
  methods: {
    async handleLogin() {
      const valid = await this.$refs.loginForm.validate()
      if (!valid) return
      
      this.submitting = true
      
      try {
        const response = await this.$axios.post('/api/auth/login', {
          username: this.loginForm.username.trim(),
          password: this.loginForm.password
        })
        
        // 保存认证信息
        const { token, user_info } = response.data
        localStorage.setItem('token', token)
        localStorage.setItem('username', user_info?.username || this.loginForm.username)
        
        // 成功提示 + 跳转
        this.$message.success({
          message: '登录成功！正在进入系统...',
          duration: 1500
        })
        
        // 路由跳转到首页
        this.$router.push('/')
        
      } catch (error) {
        // 错误处理
        const errorMsg = error.response?.data?.error || 
                        error.message || 
                        '认证失败，请检查用户名和密码'
        
        this.$message.error({
          message: errorMsg,
          duration: 3000,
          onClose: () => {
            this.submitting = false
            this.passwordPlaceholder = '密码错误，请重试'
            // 清空密码并重新聚焦
            this.loginForm.password = ''
            this.$nextTick(() => {
              this.$refs.passwordInput && this.$refs.passwordInput.focus()
            })
          }
        })
        
        this.submitting = false
        
      } finally {
        if (!this.submitting) {
          this.passwordPlaceholder = '请输入密码'
        }
      }
    },
    
    handleKeyPress(event) {
      if (event.key === 'Enter' && !this.submitting) {
        this.handleLogin()
      }
    }
  },
  watch: {
    '$route.query.redirect'(newPath) {
      if (newPath) {
        this.$router.push(newPath)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
/* ========== 容器布局 - 薄荷绿主题 ========== */
.login-container {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #064E3B 0%, #059669 35%, #10B981 65%, #34D399 100%);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ========== 动态背景效果 ========== */
.animated-background {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.4;
  animation: float 20s infinite ease-in-out;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(16, 185, 129, 0.5) 0%, transparent 70%);
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(5, 150, 105, 0.5) 0%, transparent 70%);
  bottom: -80px;
  right: -80px;
  animation-delay: 5s;
}

.orb-3 {
  width: 350px;
  height: 350px;
  background: radial-gradient(circle, rgba(52, 211, 153, 0.4) 0%, transparent 70%);
  bottom: -120px;
  left: 20%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -50px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
}

/* ========== 登录卡片样式 ========== */
.login-card {
  position: relative;
  z-index: 10;
  width: 480px;
  max-width: 90%;
  box-shadow: 
    0 8px 32px rgba(16, 185, 129, 0.2),
    0 4px 16px rgba(0, 0, 0, 0.15);
  border-radius: var(--radius-xl);
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid rgba(16, 185, 129, 0.1);
  
  & >>> .el-card__body {
    padding: 30px !important;
  }
}

.custom-header {
  padding: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  margin: -30px -30px 30px -30px;
  
  .title {
    margin: 0;
    font-size: 26px;
    font-weight: 600;
    color: white;
    letter-spacing: 1.5px;
    text-align: center;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }
}

.form-content {
  padding: 30px;
}

/* ========== 表单输入框样式 - 薄荷绿主题 ========== */
.el-form-item {
  margin-bottom: 24px;
  
  &.el-form-item--label-position-top >>> .el-form-item__label {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-primary);
    margin-bottom: 8px;
  }
}

.el-input {
  & >>> .el-input__inner {
    height: 48px !important;
    line-height: 48px !important;
    font-size: 16px;
    border: 2px solid var(--color-border-light);
    border-radius: var(--radius-md);
    transition: all var(--transition-base);
    background-color: white !important;
    color: var(--color-text-primary) !important;
    padding-left: 15px !important;
    padding-right: 30px !important;
    box-sizing: border-box !important;
    
    &:hover {
      border-color: var(--color-primary);
      box-shadow: 0 0 0 3px var(--color-primary-alpha-10);
    }
    
    &:focus {
      border-color: var(--color-primary);
      box-shadow: 0 0 0 3px var(--color-primary-alpha-20);
      outline: none;
    }
  }
  
  &__icon {
    cursor: pointer;
    color: var(--color-secondary);
  }
}

/* ========== 登录按钮样式 - 薄荷绿主题 ========== */
.login-btn {
  width: 100%;
  height: 52px;
  font-size: 16px;
  font-weight: 600;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%) !important;
  border: none;
  transition: all var(--transition-base);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
  }
  
  &:active {
    transform: translateY(0);
  }
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  i, span {
    margin-right: 8px;
  }
}

/* ========== 辅助信息区域 ========== */
.help-info {
  margin-top: 30px;
  
  .el-divider {
    margin-bottom: 20px;
    background-color: #dcdfe6;
  }
  
  .info-row,
  .security-tip {
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    padding: 8px 0;
    
    i {
      margin-right: 8px;
      font-size: 16px;
    }
  }
  
  .info-row {
    color: var(--color-text-secondary);
    
    i {
      color: var(--color-info);
    }
  }
  
  .security-tip {
    color: var(--color-success);
    font-weight: 500;
    background: var(--color-primary-alpha-10);
    padding: 10px 16px;
    border-radius: var(--radius-md);
    
    i {
      color: var(--color-success);
    }
  }
}

/* ========== 入场动画类 ========== */
.login-card-enter {
  &-enter {
    opacity: 0;
    transform: translateY(30px);
  }
  
  &-enter-active {
    transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  }
  
  &-leave-to {
    opacity: 0;
    transform: translateY(30px);
  }
  
  &-leave-active {
    transition: all 0.4s ease;
  }
}

.fade-in {
  &-enter,
  &-leave-to {
    opacity: 0;
  }
  
  &-enter-active,
  &-leave-active {
    transition: opacity 0.3s ease;
  }
}

/* ========== 加载遮罩层 - 薄荷绿主题 ========== */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(16, 185, 129, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loader-container {
  text-align: center;
}

.ldap-animation {
  .secure-badge {
    width: 80px;
    height: 80px;
    margin: 0 auto 20px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-success) 0%, var(--color-primary-dark) 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    animation: security-glow 2s infinite;
    
    i {
      font-size: 40px;
      color: white;
    }
    
    @keyframes security-glow {
      0%, 100% {
        box-shadow: 0 0 20px rgba(16, 185, 129, 0.6);
        transform: scale(1);
      }
      50% {
        box-shadow: 0 0 40px rgba(16, 185, 129, 0.9);
        transform: scale(1.05);
      }
    }
  }
}

/* ========== 响应式设计 ========== */
@media (max-width: 768px) {
  .login-card {
    width: 100%;
    max-width: 400px;
    margin: 20px;
  }
  
  .form-content {
    padding: 24px;
  }
  
  .gradient-orb {
    filter: blur(40px);
  }
  
  .orb-1 {
    width: 250px;
    height: 250px;
  }
  
  .orb-2,
  .orb-3 {
    width: 200px;
    height: 200px;
  }
}


/* ========== 无障碍支持 - 薄荷绿主题 ========== */
.login-btn:focus-visible {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
}

.el-input__inner:focus-visible {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
