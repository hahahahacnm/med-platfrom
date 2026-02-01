<script setup lang="ts">
import { ref } from 'vue'
import { 
  PersonOutline, 
  LockClosedOutline, 
  EyeOutline, 
  EyeOffOutline, 
  CheckmarkCircle, 
  ShieldCheckmarkOutline,
  ArrowForwardOutline 
} from '@vicons/ionicons5'
import { useMessage, NIcon } from 'naive-ui'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'
import request from '../utils/request'

const userStore = useUserStore()
const router = useRouter()
const message = useMessage()

// 表单数据
const formModel = ref({
  username: '', 
  password: ''
})

const loading = ref(false)
const showPassword = ref(false)

const requireCaptcha = ref(false)
const captchaId = ref('')
const captchaImg = ref('')
const captchaVal = ref('')

const fetchCaptcha = async () => {
  try {
    const res: any = await request.get('/auth/captcha')
    captchaId.value = res.id
    captchaImg.value = res.image
  } catch (e) {
    console.error(e)
  }
}

const handleLogin = async () => {
  if (!formModel.value.username || !formModel.value.password) {
    message.warning('请输入账号和密码')
    return
  }
  
  if (requireCaptcha.value && !captchaVal.value) {
    message.warning('请输入验证码')
    return
  }

  loading.value = true
  
  try {
    const success = await userStore.login({
        ...formModel.value,
        captcha_id: requireCaptcha.value ? captchaId.value : undefined,
        captcha_val: requireCaptcha.value ? captchaVal.value : undefined
    })
    
    // login throws if failed, so here is success
    message.success('登录成功！')
    const role = localStorage.getItem('role')
    if (role === 'admin' || role === 'agent') {
      router.push('/admin')
    } else {
      router.push('/')
    }

  } catch (e: any) {
    // Check if captcha is required
    if (e.response && e.response.data && e.response.data.require_captcha) {
        if (!requireCaptcha.value) {
            message.info('由于多次尝试失败，请进行安全验证')
        }
        requireCaptcha.value = true
        fetchCaptcha()
        // Clear captcha val if specific error? or keep it
        captchaVal.value = ''
    } else {
        // Other errors handled by interceptor but we might want to refresh captcha if it was wrong
        if (requireCaptcha.value) {
           captchaVal.value = ''
           fetchCaptcha()
        }
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <!-- Left Side: Brand & Visuals (Desktop) -->
    <div class="auth-brand-side">
      <div class="background-decoration">
        <svg class="bg-svg" viewBox="0 0 100 100" preserveAspectRatio="none">
          <path d="M0 100 C 20 0 50 0 100 100 Z" fill="url(#grad1)" />
          <defs>
            <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" style="stop-color: #3b82f6; stop-opacity: 1" />
              <stop offset="100%" style="stop-color: #14b8a6; stop-opacity: 1" />
            </linearGradient>
          </defs>
        </svg>
      </div>

      <div class="brand-content">
        <div class="brand-logo">
          <div class="logo-icon-wrapper">
            <n-icon size="28" color="white">
              <CheckmarkCircle />
            </n-icon>
          </div>
          <span class="brand-name">题酷</span>
        </div>

        <div class="brand-hero-text">
          <h1 class="hero-title">欢迎回到您的医学殿堂</h1>
          <p class="hero-subtitle">
            题酷 提供最权威的医学题库、知识库与 AI 助教服务，助您在医学考试与临床实践中游刃有余。
          </p>
        </div>

        <div class="brand-footer">
          <div class="certification-badge">
            <div class="cert-icon">
              <n-icon size="16" color="#34d399">
                <ShieldCheckmarkOutline />
              </n-icon>
            </div>
            <span>专业认证内容</span>
          </div>
          <p class="slogan">我们永远在这里！</p>
        </div>
      </div>
    </div>

    <!-- Right Side: Form -->
    <div class="auth-form-side">
      <!-- Mobile Header -->
      <div class="mobile-header">
        <div class="mobile-logo-icon">
          <n-icon size="20" color="white">
            <CheckmarkCircle />
          </n-icon>
        </div>
        <span class="mobile-brand-name">题酷</span>
      </div>

      <div class="form-wrapper">
        <div class="form-header">
          <div>
            <h2 class="form-title">账号登录</h2>
            <p class="form-subtitle">请输入您的认证信息以继续</p>
          </div>
          <button class="toggle-auth-btn" @click="router.push('/register')">
            免费注册
          </button>
        </div>

        <div class="form-content">
          <div class="form-group">
            <label>账号</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <PersonOutline />
              </n-icon>
              <input 
                v-model="formModel.username" 
                type="text" 
                placeholder="请输入用户名" 
                class="custom-input"
              />
            </div>
          </div>

          <div class="form-group">
            <div class="label-row">
              <label>密码</label>
            </div>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <LockClosedOutline />
              </n-icon>
              <input 
                v-model="formModel.password" 
                :type="showPassword ? 'text' : 'password'" 
                placeholder="••••••••" 
                class="custom-input"
                @keydown.enter="handleLogin"
              />
              <button type="button" class="eye-btn" @click="showPassword = !showPassword">
                <n-icon size="18" color="#94a3b8">
                  <EyeOutline v-if="!showPassword" />
                  <EyeOffOutline v-else />
                </n-icon>
              </button>
            </div>
          </div>

          <div class="form-group" v-if="requireCaptcha">
            <label>安全验证</label>
            <div style="display: flex; gap: 10px;">
                <div class="input-wrapper group-focus" style="flex: 1;">
                    <input v-model="captchaVal" type="text" placeholder="输入验证码" class="custom-input" @keydown.enter="handleLogin" />
                </div>
                <div style="width: 100px; height: 38px; cursor: pointer; border-radius: 0.75rem; overflow: hidden; border: 1px solid #e2e8f0; display: flex;" @click="fetchCaptcha">
                    <img v-if="captchaImg" :src="captchaImg" style="width: 100%; height: 100%; object-fit: fill;" />
                </div>
            </div>
          </div>

          <button 
            :disabled="loading" 
            class="submit-btn" 
            @click="handleLogin"
          >
            <span v-if="loading" class="spinner"></span>
            <span v-else class="btn-text">立即登录 <n-icon size="16"><ArrowForwardOutline /></n-icon></span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Layout */
.auth-container {
  height: 100vh;
  width: 100%;
  display: flex;
  background-color: #f8fafc; /* slate-50 */
  font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  color: #0f172a; /* slate-900 */
  overflow: hidden;
}

/* Left Side (Desktop) */
.auth-brand-side {
  display: none;
  background-color: #0f172a; /* slate-900 */
  position: relative;
  color: white;
  flex-direction: column;
  justify-content: space-between;
  padding: 2.5rem;
  width: 41.666667%; /* w-5/12 */
}

@media (min-width: 768px) {
  .auth-brand-side {
    display: flex;
  }
}

@media (min-width: 1024px) {
  .auth-brand-side {
    padding: 4rem; /* p-16 */
    width: 40%; /* lg:w-2/5 */
  }
}

.background-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0.2;
  pointer-events: none;
}

.bg-svg {
  height: 100%;
  width: 100%;
}

.brand-content {
  position: relative;
  z-index: 10;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.brand-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 2.5rem;
}

.logo-icon-wrapper {
  background-color: #2563eb; /* blue-600 */
  padding: 0.625rem;
  border-radius: 0.75rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-name {
  font-size: 1.875rem; /* text-3xl */
  font-weight: 700;
  letter-spacing: -0.025em;
}

.hero-title {
  font-size: 2.25rem; /* text-4xl */
  font-weight: 700;
  line-height: 1.25;
  margin-bottom: 1.5rem;
}

@media (min-width: 1024px) {
  .hero-title {
    font-size: 3rem; /* lg:text-5xl */
  }
}

.hero-subtitle {
  color: #94a3b8; /* slate-400 */
  font-size: 1.125rem; /* text-lg */
  line-height: 1.625;
}

.brand-footer {
  position: relative;
  z-index: 10;
  margin-top: 2.5rem;
}

.certification-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 500;
  color: #cbd5e1; /* slate-300 */
  margin-bottom: 0.5rem;
}

.cert-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  background-color: #1e293b; /* slate-800 */
  border: 1px solid #334155; /* slate-700 */
  display: flex;
  align-items: center;
  justify-content: center;
}

.slogan {
  font-size: 0.75rem; /* text-xs */
  color: #64748b; /* slate-500 */
}

/* Right Side (Form) */
.auth-form-side {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 1.5rem;
  background-color: white;
  overflow-y: auto;
  position: relative;
}

@media (min-width: 768px) {
  .auth-form-side {
    width: 58.333333%; /* w-7/12 */
    overflow: hidden;
  }
}

@media (min-width: 1024px) {
  .auth-form-side {
    width: 60%; /* lg:w-3/5 */
  }
}

.mobile-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  flex-shrink: 0;
}

@media (min-width: 768px) {
  .mobile-header {
    display: none;
  }
}

.mobile-logo-icon {
  background-color: #2563eb;
  padding: 0.5rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.mobile-brand-name {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.025em;
  color: #0f172a;
}

.form-wrapper {
  max-width: 28rem; /* max-w-md */
  width: 100%;
  margin: 0 auto;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 1.25rem;
}

.form-title {
  font-size: 1.5rem; /* text-2xl */
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.25rem 0;
}

@media (min-width: 768px) {
  .form-title {
    font-size: 1.875rem; /* md:text-3xl */
  }
}

.form-subtitle {
  font-size: 0.875rem; /* text-sm */
  color: #64748b; /* slate-500 */
  margin: 0;
}

.toggle-auth-btn {
  font-size: 0.875rem; /* text-sm */
  font-weight: 700;
  color: #2563eb;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.375rem 0.75rem;
  border-radius: 0.5rem;
  transition: all 0.2s;
}

.toggle-auth-btn:hover {
  color: #1d4ed8;
  background-color: #eff6ff;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1rem; /* space-y-3 */
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-group label {
  font-size: 0.75rem; /* text-xs */
  font-weight: 700;
  color: #334155; /* slate-700 */
  margin-left: 0.25rem;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  transition: color 0.2s;
}

.group-focus:focus-within .input-icon {
  color: #3b82f6 !important; /* blue-500 */
}

.custom-input {
  width: 100%;
  background-color: #f8fafc; /* slate-50 */
  border: 1px solid #e2e8f0; /* slate-200 */
  border-radius: 0.75rem; /* rounded-xl */
  padding: 0.625rem 1rem 0.625rem 2.5rem; /* py-2.5 pl-10 */
  font-size: 0.875rem; /* text-sm */
  font-weight: 500;
  color: #1e293b; /* slate-800 */
  outline: none;
  transition: all 0.2s;
}

.custom-input:focus {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
}

.eye-btn {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}

.eye-btn:hover {
  color: #475569;
}

.submit-btn {
  width: 100%;
  font-weight: 700;
  padding: 0.75rem;
  border-radius: 0.75rem; /* rounded-xl */
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 0.875rem; /* text-sm */
  background-color: #0f172a; /* slate-900 */
  color: white;
  border: none;
  cursor: pointer;
}

.submit-btn:disabled {
  background-color: #e2e8f0; /* slate-200 */
  color: #94a3b8; /* slate-400 */
  cursor: not-allowed;
  box-shadow: none;
}

.submit-btn:not(:disabled):hover {
  background-color: #2563eb; /* blue-600 */
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.submit-btn:not(:disabled):active {
  transform: scale(0.99);
}

.spinner {
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid #94a3b8;
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
</style>