<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NIcon, NModal, NCard } from 'naive-ui'
import { 
  PersonOutline, 
  LockClosedOutline, 
  HappyOutline, 
  CheckmarkCircle, 
  ShieldCheckmarkOutline,
  ArrowForwardOutline,
  EyeOutline,
  EyeOffOutline,
  Checkmark,
  CloseOutline,
  MailOutline
} from '@vicons/ionicons5'
import request from '../utils/request'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Agreement State
const hasAgreed = ref(false)
const showDocument = ref<'agreement' | 'privacy' | 'disclaimer' | null>(null)

// 表单数据
const model = reactive({
  username: '',
  nickname: '', // 昵称
  password: '',
  confirmPassword: ''
})

const handleRegister = async () => {
  if (!hasAgreed.value) {
    message.warning('请先阅读并同意用户协议等条款')
    return
  }
  
  // Basic validation
  if (!model.username || !model.nickname || !model.password || !model.confirmPassword) {
    message.warning('请填写所有必填项')
    return
  }
  
  if (model.password.length < 6) {
    message.warning('密码长度不能少于6位')
    return
  }

  if (model.password !== model.confirmPassword) {
    message.warning('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    // 调用后端注册接口
    await request.post('/auth/register', {
      username: model.username,
      password: model.password,
      nickname: model.nickname
    })
    
    message.success('注册成功！请登录')
    router.push('/login')
  } catch (error) {
    // 错误处理交给了 request 拦截器
  } finally {
    loading.value = false
  }
}

const openDocument = (type: 'agreement' | 'privacy' | 'disclaimer') => {
  showDocument.value = type
}

const closeDocument = () => {
  showDocument.value = null
}

const agreeDocument = () => {
  hasAgreed.value = true
  closeDocument()
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
          <h1 class="hero-title">开启您的医学进阶之旅</h1>
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
            <h2 class="form-title">创建新账号</h2>
            <p class="form-subtitle">请输入您的认证信息以继续</p>
          </div>
          <button class="toggle-auth-btn" @click="router.push('/login')">
            已有账号?
          </button>
        </div>

        <div class="form-content">
          <div class="form-group">
            <label>姓名</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <HappyOutline />
              </n-icon>
              <input 
                v-model="model.nickname" 
                type="text" 
                placeholder="怎么称呼您？" 
                class="custom-input"
              />
            </div>
          </div>

          <div class="form-group">
            <label>账号</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <PersonOutline />
              </n-icon>
              <input 
                v-model="model.username" 
                type="text" 
                placeholder="请输入用户名 (唯一标识)" 
                class="custom-input"
              />
            </div>
          </div>

          <div class="form-group">
            <label>密码</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <LockClosedOutline />
              </n-icon>
              <input 
                v-model="model.password" 
                :type="showPassword ? 'text' : 'password'" 
                placeholder="请输入密码" 
                class="custom-input"
              />
              <button type="button" class="eye-btn" @click="showPassword = !showPassword">
                <n-icon size="18" color="#94a3b8">
                  <EyeOutline v-if="!showPassword" />
                  <EyeOffOutline v-else />
                </n-icon>
              </button>
            </div>
          </div>

          <div class="form-group">
            <label>确认密码</label>
            <div class="input-wrapper group-focus">
              <n-icon class="input-icon" size="18" color="#94a3b8">
                <LockClosedOutline />
              </n-icon>
              <input 
                v-model="model.confirmPassword" 
                :type="showConfirmPassword ? 'text' : 'password'" 
                placeholder="请再次输入密码" 
                class="custom-input"
                @keydown.enter="handleRegister"
              />
              <button type="button" class="eye-btn" @click="showConfirmPassword = !showConfirmPassword">
                <n-icon size="18" color="#94a3b8">
                  <EyeOutline v-if="!showConfirmPassword" />
                  <EyeOffOutline v-else />
                </n-icon>
              </button>
            </div>
          </div>

          <!-- Agreement Checkbox -->
          <div class="agreement-section">
            <label class="agreement-label">
              <input 
                type="checkbox" 
                v-model="hasAgreed"
                class="agreement-checkbox"
              />
              <span class="custom-checkbox-ui">
                <n-icon size="10" color="white" v-if="hasAgreed">
                  <Checkmark />
                </n-icon>
              </span>
              <span class="agreement-text">
                我已仔细阅读并同意题酷
                <span class="link" @click.prevent="openDocument('agreement')">用户协议</span>、
                <span class="link" @click.prevent="openDocument('privacy')">隐私政策</span>、
                <span class="link" @click.prevent="openDocument('disclaimer')">免责声明</span>
              </span>
            </label>
          </div>

          <button 
            :disabled="loading || !hasAgreed" 
            class="submit-btn" 
            @click="handleRegister"
          >
            <span v-if="loading" class="spinner"></span>
            <span v-else class="btn-text">注册账号 <n-icon size="16"><ArrowForwardOutline /></n-icon></span>
          </button>
        </div>
      </div>
    </div>

    <!-- Document Modal Overlay -->
    <div v-if="showDocument" class="modal-overlay" @click="closeDocument">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">
            <n-icon color="#2563eb" size="20"><ShieldCheckmarkOutline /></n-icon>
            <span v-if="showDocument === 'agreement'">题酷 用户协议</span>
            <span v-if="showDocument === 'privacy'">隐私政策</span>
            <span v-if="showDocument === 'disclaimer'">免责声明</span>
          </h3>
          <button class="close-btn" @click="closeDocument">
            <n-icon size="18"><CloseOutline /></n-icon>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="showDocument === 'agreement'" class="doc-text">
            <h4>1. 服务条款的确认和接纳</h4>
            <p>题酷提供的服务完全按照其发布的服务条款和操作规则严格执行。用户必须完全同意所有服务条款并完成注册程序，才能成为题酷的正式用户。</p>
            <h4>2. 服务说明</h4>
            <p>题酷运用自己的系统通过互联网向用户提供包括医学题库、AI助教等在内的网络服务。</p>
            <h4>3. 用户的帐号，密码和安全性</h4>
            <p>用户一旦注册成功，成为题酷的合法用户，将得到一个密码和用户名。用户将对用户名和密码安全负全部责任。</p>
          </div>
          <div v-if="showDocument === 'privacy'" class="doc-text">
            <h4>1. 信息收集</h4>
            <p>我们在您注册、使用服务时收集您的个人信息，包括但不限于姓名、邮箱、学习记录等。</p>
            <h4>2. 信息使用</h4>
            <p>我们使用这些信息来为您提供个性化的学习体验、改进我们的服务以及通知您有关产品更新的信息。</p>
          </div>
          <div v-if="showDocument === 'disclaimer'" class="doc-text">
             <h4>1. 内容免责</h4>
            <p>题酷提供的所有医学知识、题目解析仅供学习参考，不能替代专业的医疗建议、诊断或治疗。</p>
          </div>
          <div class="doc-footer-text">最后更新日期：2026年1月21日</div>
        </div>
        <div class="modal-footer">
          <button class="agree-modal-btn" @click="agreeDocument">
            阅读并同意
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Reuse styles from Login.vue by copy-padding or simply duplicating here for isolation */

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

/* Agreement & Modal */
.agreement-section {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.agreement-label {
  display: flex;
  align-items: center;
  cursor: pointer;
  position: relative;
}

.agreement-checkbox {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.custom-checkbox-ui {
  width: 1rem;
  height: 1rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 0.5rem;
  transition: all 0.2s;
  flex-shrink: 0;
}

.agreement-checkbox:checked + .custom-checkbox-ui {
  background-color: #2563eb;
  border-color: #2563eb;
}

.agreement-checkbox:focus + .custom-checkbox-ui {
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2);
}

.agreement-text {
  font-size: 0.75rem; /* text-xs */
  color: #64748b; /* slate-500 */
  line-height: 1.5;
  user-select: none;
}

.link {
  color: #2563eb;
  font-weight: 500;
  cursor: pointer;
}

.link:hover {
  text-decoration: underline;
  color: #1d4ed8;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  padding: 1rem;
}

.modal-content {
  background-color: white;
  border-radius: 1rem;
  width: 100%;
  max-width: 32rem; /* max-w-lg */
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  animation: modalIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  padding: 1rem;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: white;
  z-index: 10;
}

.modal-title {
  font-size: 1.125rem; /* text-lg */
  font-weight: 700;
  color: #0f172a;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
}

.close-btn {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  background-color: #f8fafc;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background-color: #f1f5f9;
  color: #475569;
}

.modal-body {
  padding: 1.25rem;
  overflow-y: auto;
  color: #475569; /* slate-600 */
  font-size: 0.875rem; /* text-sm */
  line-height: 1.625;
}

.doc-text h4 {
  font-weight: 700;
  color: #1e293b; /* slate-800 */
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.doc-footer-text {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid #f1f5f9;
  font-size: 0.75rem;
  color: #94a3b8;
}

.modal-footer {
  padding: 1rem;
  border-top: 1px solid #f1f5f9;
  background-color: #f8fafc;
  display: flex;
  justify-content: flex-end;
}

.agree-modal-btn {
  background-color: #0f172a;
  color: white;
  font-weight: 700;
  padding: 0.5rem 1.5rem;
  border-radius: 0.5rem;
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.agree-modal-btn:hover {
  background-color: #2563eb;
}
</style>