<script setup lang="ts">
import { ref, onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { 
  PersonOutline, 
  LockClosedOutline, 
  EyeOutline, 
  EyeOffOutline, 
  CheckmarkCircle, 
  ShieldCheckmarkOutline, 
  ArrowForwardOutline 
} from '@vicons/ionicons5'
import { useMessage, NIcon, NModal, NSpin } from 'naive-ui'
import { useUserStore } from '../stores/user'
import { useHandler } from '../hooks/useRotateHandler'
import * as GoCaptchaLib from 'go-captcha-vue'

const userStore = useUserStore()
const router = useRouter()
const message = useMessage()

const CaptchaComponent = shallowRef<any>(null)

onMounted(() => {
  const lib = GoCaptchaLib as any
  const comp = lib.Rotate || lib.GocaptchaRotate || lib.default || lib
  if (comp) {
    CaptchaComponent.value = comp
    console.log('✅ 验证码组件加载成功')
  } else {
    console.error('❌ 验证码组件加载失败')
  }
})

const formModel = ref({
  username: '', 
  password: ''
})

const loading = ref(false)
const showPassword = ref(false)
const showCaptcha = ref(false)
// 🔥 修改 1：定义 ref 变量，用于传给 hook 控制组件
const captchaDomRef = ref<any>(null)

const handleAuthConfirm = async (captchaData: any) => {
  loading.value = true
  try {
    const loginParams = {
      ...formModel.value,
      captcha_id: captchaData.key,
      captcha_value: captchaData.angle
    }

    const success = await userStore.login(loginParams)
    
    if (success) {
      message.success('登录成功！')
      showCaptcha.value = false
      const role = localStorage.getItem('role')
      if (role === 'admin' || role === 'agent') {
        router.push('/admin')
      } else {
        router.push('/')
      }
      return Promise.resolve()
    } else {
      // 🔥 修改 2：登录业务失败也要抛出异常，触发 hook 的重试逻辑
      return Promise.reject()
    }
  } catch (e) {
    return Promise.reject()
  } finally {
    loading.value = false
  }
}

// 传入 ref
const handler = useHandler(captchaDomRef, handleAuthConfirm)

const handleLogin = async () => {
  if (!formModel.value.username || !formModel.value.password) {
    message.warning('请输入账号和密码')
    return
  }
  showCaptcha.value = true
  setTimeout(() => {
    handler.requestCaptchaData()
  }, 100)
}
</script>

<template>
  <div class="auth-container">
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
            <n-icon size="28" color="white"><CheckmarkCircle /></n-icon>
          </div>
          <span class="brand-name">题酷</span>
        </div>
        <div class="brand-hero-text">
          <h1 class="hero-title">欢迎回到您的医学殿堂</h1>
          <p class="hero-subtitle">权威题库 · 智能助教 · 全栈式备考方案</p>
        </div>
        <div class="brand-footer">
          <div class="certification-badge">
            <div class="cert-icon"><n-icon size="16" color="#34d399"><ShieldCheckmarkOutline /></n-icon></div>
            <span>专业认证内容</span>
          </div>
          <p class="slogan">我们永远在这里！</p>
        </div>
      </div>
    </div>

    <div class="auth-form-side">
      <div class="mobile-header">
        <div class="mobile-logo-icon"><n-icon size="20" color="white"><CheckmarkCircle /></n-icon></div>
        <span class="mobile-brand-name">题酷</span>
      </div>

      <div class="form-wrapper">
        <div class="form-header">
          <div><h2 class="form-title">账号登录</h2><p class="form-subtitle">请输入您的认证信息以继续</p></div>
          <button class="toggle-auth-btn" @click="router.push('/register')">免费注册</button>
        </div>

        <div class="form-content">
          <div class="form-group">
            <label>账号</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8"><PersonOutline /></n-icon>
              <input v-model="formModel.username" type="text" placeholder="请输入用户名" class="custom-input" />
            </div>
          </div>

          <div class="form-group">
            <div class="label-row"><label>密码</label></div>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8"><LockClosedOutline /></n-icon>
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

          <button :disabled="loading" class="submit-btn" @click="handleLogin">
            <span class="btn-text">立即登录 <n-icon size="16"><ArrowForwardOutline /></n-icon></span>
          </button>
        </div>
      </div>
    </div>

    <n-modal v-model:show="showCaptcha" transform-origin="center">
      <div class="captcha-wrapper">
        <div v-if="!handler.data.image" class="status-box">
          <n-spin size="medium" />
          <span class="loading-text">正在加载验证码...</span>
        </div>

        <component 
          v-else-if="CaptchaComponent"
          ref="captchaDomRef"
          :is="CaptchaComponent"
          :data="handler.data"
          :events="{
            close: () => { showCaptcha = false },
            refresh: handler.refreshEvent,
            confirm: handler.confirmEvent,
          }"
        />

        <div v-else class="status-box error">
          组件加载异常，请刷新页面
        </div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
/* 🔥 修改 4：CSS 适配，防止手机端弹窗过大 */
.captcha-wrapper {
  background: white;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  
  /* 关键修改：宽度自适应，但在大屏上保持合适大小 */
  width: auto;
  max-width: 90vw; /* 手机上不超过屏幕 */
  min-height: 290px;
  
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
}

/* 针对极小屏幕 */
@media (max-width: 360px) {
  .captcha-wrapper {
    padding: 10px;
    width: 95vw;
  }
}

.status-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #666;
}
.status-box.error { color: #ff4d4f; }
.loading-text { font-size: 14px; }

/* 其他样式保持原有不变 */
.auth-container { height: 100vh; width: 100%; display: flex; background-color: #f8fafc; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; color: #0f172a; overflow: hidden; }
.auth-brand-side { display: none; background-color: #0f172a; position: relative; color: white; flex-direction: column; justify-content: space-between; padding: 2.5rem; width: 41.666667%; }
@media (min-width: 768px) { .auth-brand-side { display: flex; } }
@media (min-width: 1024px) { .auth-brand-side { padding: 4rem; width: 40%; } }
.background-decoration { position: absolute; top: 0; left: 0; width: 100%; height: 100%; opacity: 0.2; pointer-events: none; }
.bg-svg { height: 100%; width: 100%; }
.brand-content { position: relative; z-index: 10; height: 100%; display: flex; flex-direction: column; justify-content: space-between; }
.brand-logo { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 2.5rem; }
.logo-icon-wrapper { background-color: #2563eb; padding: 0.625rem; border-radius: 0.75rem; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1); display: flex; align-items: center; justify-content: center; }
.brand-name { font-size: 1.875rem; font-weight: 700; letter-spacing: -0.025em; }
.hero-title { font-size: 2.25rem; font-weight: 700; line-height: 1.25; margin-bottom: 1.5rem; }
@media (min-width: 1024px) { .hero-title { font-size: 3rem; } }
.hero-subtitle { color: #94a3b8; font-size: 1.125rem; line-height: 1.625; }
.brand-footer { position: relative; z-index: 10; margin-top: 2.5rem; }
.certification-badge { display: flex; align-items: center; gap: 0.75rem; font-weight: 500; color: #cbd5e1; margin-bottom: 0.5rem; }
.cert-icon { width: 2rem; height: 2rem; border-radius: 9999px; background-color: #1e293b; border: 1px solid #334155; display: flex; align-items: center; justify-content: center; }
.slogan { font-size: 0.75rem; color: #64748b; }
.auth-form-side { width: 100%; height: 100%; display: flex; flex-direction: column; justify-content: center; padding: 1.5rem; background-color: white; overflow-y: auto; position: relative; }
@media (min-width: 768px) { .auth-form-side { width: 58.333333%; overflow: hidden; } }
@media (min-width: 1024px) { .auth-form-side { width: 60%; } }
.mobile-header { display: flex; align-items: center; justify-content: center; gap: 0.5rem; margin-bottom: 1.5rem; flex-shrink: 0; }
@media (min-width: 768px) { .mobile-header { display: none; } }
.mobile-logo-icon { background-color: #2563eb; padding: 0.5rem; border-radius: 0.5rem; display: flex; align-items: center; justify-content: center; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
.mobile-brand-name { font-size: 1.5rem; font-weight: 700; letter-spacing: -0.025em; color: #0f172a; }
.form-wrapper { max-width: 28rem; width: 100%; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1.25rem; }
.form-title { font-size: 1.5rem; font-weight: 700; color: #0f172a; margin: 0 0 0.25rem 0; }
@media (min-width: 768px) { .form-title { font-size: 1.875rem; } }
.form-subtitle { font-size: 0.875rem; color: #64748b; margin: 0; }
.toggle-auth-btn { font-size: 0.875rem; font-weight: 700; color: #2563eb; background: none; border: none; cursor: pointer; padding: 0.375rem 0.75rem; border-radius: 0.5rem; transition: all 0.2s; }
.toggle-auth-btn:hover { color: #1d4ed8; background-color: #eff6ff; }
.form-content { display: flex; flex-direction: column; gap: 1rem; }
.form-group { display: flex; flex-direction: column; gap: 0.25rem; }
.form-group label { font-size: 0.75rem; font-weight: 700; color: #334155; margin-left: 0.25rem; }
.input-wrapper { position: relative; display: flex; align-items: center; }
.input-icon { position: absolute; left: 0.75rem; top: 50%; transform: translateY(-50%); pointer-events: none; transition: color 0.2s; }
.group-focus:focus-within .input-icon { color: #3b82f6 !important; }
.custom-input { width: 100%; background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 0.75rem; padding: 0.625rem 1rem 0.625rem 2.5rem; font-size: 0.875rem; font-weight: 500; color: #1e293b; outline: none; transition: all 0.2s; }
.custom-input:focus { box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2); border-color: #3b82f6; }
.eye-btn { position: absolute; right: 0.75rem; top: 50%; transform: translateY(-50%); background: none; border: none; cursor: pointer; padding: 0.25rem; display: flex; align-items: center; justify-content: center; color: #94a3b8; }
.eye-btn:hover { color: #475569; }
.submit-btn { width: 100%; font-weight: 700; padding: 0.75rem; border-radius: 0.75rem; box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1); transition: all 0.2s; display: flex; align-items: center; justify-content: center; gap: 0.5rem; font-size: 0.875rem; background-color: #0f172a; color: white; border: none; cursor: pointer; }
.submit-btn:disabled { background-color: #e2e8f0; color: #94a3b8; cursor: not-allowed; box-shadow: none; }
.submit-btn:not(:disabled):hover { background-color: #2563eb; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1); }
.submit-btn:not(:disabled):active { transform: scale(0.99); }
.btn-text { display: flex; align-items: center; gap: 0.5rem; }
</style>