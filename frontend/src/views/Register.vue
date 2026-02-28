<script setup lang="ts">
import { ref, reactive, shallowRef, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  useMessage, NIcon, NModal, NSpin, NButton 
} from 'naive-ui'
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
  TicketOutline,
  MailOutline,
  MailOpenOutline // 🔥 新增已发送图标
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useHandler } from '../hooks/useRotateHandler'

// 引入验证码组件及样式
import * as GoCaptchaLib from 'go-captcha-vue'
import 'go-captcha-vue/dist/style.css'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const showCaptcha = ref(false)
const captchaDomRef = ref(null)

// 🔥 邮箱验证专属状态
const isEmailSent = ref(false)
const resendLoading = ref(false)
const resendCountdown = ref(0)
let timer: any = null

const CaptchaComponent = shallowRef<any>(null)
onMounted(() => {
  const lib = GoCaptchaLib as any
  CaptchaComponent.value = lib.Rotate || lib.GocaptchaRotate || lib.default || lib
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const hasAgreed = ref(false)
const showDocument = ref<'agreement' | 'privacy' | 'disclaimer' | null>(null)

const model = reactive({
  username: '', nickname: '', email: '', password: '', confirmPassword: '', invitationCode: ''
})

const validateForm = () => {
  if (!hasAgreed.value) { message.warning('请先阅读并同意用户协议等条款'); return false }
  if (!model.username || !model.nickname || !model.email || !model.password || !model.confirmPassword) { message.warning('请填写所有必填项'); return false }
  
  if (!/^[a-zA-Z][a-zA-Z0-9_]{3,19}$/.test(model.username)) { message.error('账号格式错误：需字母开头，4-20位字符'); return false }
  if (!/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(model.email)) { message.error('邮箱格式不正确'); return false }
  if (model.password.length < 6) { message.warning('密码长度不能少于6位'); return false }
  if (model.password !== model.confirmPassword) { message.warning('两次输入的密码不一致'); return false }
  return true
}

const handleRegister = () => {
  if (!validateForm()) return
  showCaptcha.value = true
  setTimeout(() => handler.requestCaptchaData(), 100)
}

const startCountdown = () => {
  resendCountdown.value = 60
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    resendCountdown.value--
    if (resendCountdown.value <= 0) clearInterval(timer)
  }, 1000)
}

// 提交注册表单
const submitRegister = async (captchaData: any) => {
  loading.value = true
  try {
    await request.post('/auth/register', {
      username: model.username,
      password: model.password,
      nickname: model.nickname,
      email: model.email, 
      invitation_code: model.invitationCode,
      captcha_id: captchaData.key,
      captcha_value: captchaData.angle
    })
    
    showCaptcha.value = false
    isEmailSent.value = true // 🔥 切换到发信成功页面
    startCountdown() // 开启 60 秒防刷
    return Promise.resolve()
  } catch (error) {
    return Promise.reject()
  } finally {
    loading.value = false
  }
}

// 重新发送邮件
const handleResendEmail = async () => {
  if (resendCountdown.value > 0) return
  resendLoading.value = true
  try {
    await request.post('/auth/resend-email', { email: model.email })
    message.success('激活邮件已重新发送，请查收')
    startCountdown()
  } catch (error: any) {
    message.error(error.response?.data?.error || '发送失败')
  } finally {
    resendLoading.value = false
  }
}

const handler = useHandler(captchaDomRef, submitRegister)

const openDocument = (type: 'agreement' | 'privacy' | 'disclaimer') => { showDocument.value = type }
const closeDocument = () => { showDocument.value = null }
const agreeDocument = () => { hasAgreed.value = true; closeDocument() }
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
        <div class="brand-logo"><div class="logo-icon-wrapper"><n-icon size="28" color="white"><CheckmarkCircle /></n-icon></div><span class="brand-name">题酷</span></div>
        <div class="brand-hero-text"><h1 class="hero-title">开启您的医学进阶之旅</h1><p class="hero-subtitle">题酷 提供最权威的医学题库、知识库与 AI 助教服务，助您在医学考试与临床实践中游刃有余。</p></div>
        <div class="brand-footer"><div class="certification-badge"><div class="cert-icon"><n-icon size="16" color="#34d399"><ShieldCheckmarkOutline /></n-icon></div><span>专业认证内容</span></div><p class="slogan">我们永远在这里！</p></div>
      </div>
    </div>

    <div class="auth-form-side">
      <div class="mobile-header">
        <div class="mobile-logo-icon"><n-icon size="20" color="white"><CheckmarkCircle /></n-icon></div>
        <span class="mobile-brand-name">题酷</span>
      </div>

      <div class="form-wrapper">
        
        <div v-if="!isEmailSent" class="fade-enter">
          <div class="form-header">
            <div><h2 class="form-title">创建新账号</h2><p class="form-subtitle">请输入您的认证信息以继续</p></div>
            <button class="toggle-auth-btn" @click="router.push('/login')">已有账号?</button>
          </div>

          <div class="form-content">
            <div class="form-group">
              <label>昵称</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><HappyOutline /></n-icon><input v-model="model.nickname" type="text" placeholder="怎么称呼您？" class="custom-input"/></div>
            </div>
            <div class="form-group">
              <label>账号/用户名</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><PersonOutline /></n-icon><input v-model="model.username" type="text" placeholder="字母开头，4-20位字符" class="custom-input"/></div>
            </div>
            <div class="form-group">
              <label>邮箱</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><MailOutline /></n-icon><input v-model="model.email" type="text" placeholder="用于激活账号与找回密码" class="custom-input"/></div>
            </div>
            <div class="form-group">
              <label>密码</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><LockClosedOutline /></n-icon><input v-model="model.password" :type="showPassword ? 'text' : 'password'" placeholder="至少6位" class="custom-input"/><button type="button" class="eye-btn" @click="showPassword = !showPassword"><n-icon size="18" color="#94a3b8"><EyeOutline v-if="!showPassword" /><EyeOffOutline v-else /></n-icon></button></div>
            </div>
            <div class="form-group">
              <label>确认密码</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><LockClosedOutline /></n-icon><input v-model="model.confirmPassword" :type="showConfirmPassword ? 'text' : 'password'" placeholder="请再次输入密码" class="custom-input"/><button type="button" class="eye-btn" @click="showConfirmPassword = !showConfirmPassword"><n-icon size="18" color="#94a3b8"><EyeOutline v-if="!showConfirmPassword" /><EyeOffOutline v-else /></n-icon></button></div>
            </div>
            <div class="form-group">
              <label>邀请码 (选填)</label>
              <div class="input-wrapper group-focus"><n-icon class="input-icon" size="18" color="#94a3b8"><TicketOutline /></n-icon><input v-model="model.invitationCode" type="text" placeholder="如有邀请码，请输入" class="custom-input" @keydown.enter="handleRegister"/></div>
            </div>

            <div class="agreement-section">
              <label class="agreement-label">
                <input type="checkbox" v-model="hasAgreed" class="agreement-checkbox"/>
                <span class="custom-checkbox-ui"><n-icon size="10" color="white" v-if="hasAgreed"><Checkmark /></n-icon></span>
                <span class="agreement-text">我已仔细阅读并同意题酷<span class="link" @click.prevent="openDocument('agreement')">用户协议</span>、<span class="link" @click.prevent="openDocument('privacy')">隐私政策</span>、<span class="link" @click.prevent="openDocument('disclaimer')">免责声明</span></span>
              </label>
            </div>

            <button :disabled="loading || !hasAgreed" class="submit-btn" @click="handleRegister">
              <span class="btn-text">注册账号 <n-icon size="16"><ArrowForwardOutline /></n-icon></span>
            </button>
          </div>
        </div>

        <div v-else class="email-sent-state fade-enter">
          <div class="sent-icon-wrapper">
            <n-icon size="64" color="#2563eb"><MailOpenOutline /></n-icon>
          </div>
          <h2 class="sent-title">验证邮件已发送</h2>
          <p class="sent-desc">
            我们已向 <strong>{{ model.email }}</strong> 发送了一封激活邮件。请前往邮箱点击无感激活链接完成注册。
          </p>
          <div class="sent-tip">
            💡 链接24小时内有效。如果没有收到，请检查您的垃圾邮件箱。
          </div>
          
          <div class="sent-actions">
            <n-button 
              type="primary" 
              secondary 
              size="large" 
              block 
              :loading="resendLoading" 
              :disabled="resendCountdown > 0" 
              @click="handleResendEmail"
            >
              {{ resendCountdown > 0 ? `${resendCountdown} 秒后可重新发送` : '未收到？重新发送' }}
            </n-button>
            <n-button 
              type="primary" 
              size="large" 
              block 
              @click="router.push('/login')" 
              style="margin-top: 12px"
            >
              已激活，去登录
            </n-button>
          </div>
        </div>

      </div>
    </div>

    <n-modal v-model:show="showCaptcha" transform-origin="center">
      <div class="captcha-box">
        <div v-if="!handler.data.image" class="status-state">
          <n-spin size="medium" />
          <span style="font-size: 14px; margin-top: 8px; color: #666;">安全验证加载中...</span>
        </div>
        <component v-else-if="CaptchaComponent" :is="CaptchaComponent" :data="handler.data" :events="{ close: () => { showCaptcha = false }, refresh: handler.refreshEvent, confirm: handler.confirmEvent }" />
        <div v-else class="status-state" style="color: red;">组件加载失败，请刷新页面</div>
      </div>
    </n-modal>

    <div v-if="showDocument" class="modal-overlay" @click="closeDocument">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title"><n-icon color="#2563eb" size="20"><ShieldCheckmarkOutline /></n-icon><span v-if="showDocument === 'agreement'">题酷 用户协议</span><span v-if="showDocument === 'privacy'">隐私政策</span><span v-if="showDocument === 'disclaimer'">免责声明</span></h3>
          <button class="close-btn" @click="closeDocument"><n-icon size="18"><CloseOutline /></n-icon></button>
        </div>
        <div class="modal-body">
          <div v-if="showDocument === 'agreement'" class="doc-text">
            <h4>1. 协议的接受与修改</h4>
            <p>欢迎使用“题酷”平台！本协议是您与平台之间关于使用本服务所订立的契约。完成注册即表示您已充分阅读、理解并同意接受本协议的全部内容。平台有权依法在必要时修改本协议，修改后的内容一经公布即生效。</p>
            <h4>2. 账号注册与使用规范</h4>
            <p>您应提供真实、准确的注册信息。您的账号仅限您本人使用，严禁以任何形式转让、借用、出租或跨设备违规共享。若因保管不善导致账号被盗，责任由用户自行承担。严禁利用本平台进行任何违法违规、侵犯他人知识产权或破坏平台运行的操作（如恶意利用脚本爬取题库）。</p>
            <h4>3. 知识产权声明</h4>
            <p>“题酷”平台内包含的所有内容（包括但不限于医学题目、深度解析、AI助教回答、图文资源、交互设计等）的知识产权均归平台及相关权利人所有。未经正式书面授权，任何人不得擅自复制、传播或用于其他商业用途。</p>
            <h4>4. 服务的变更与终止</h4>
            <p>平台会尽力保障服务的连贯性和安全性，但有权根据运营情况调整、中止或终止部分服务。若用户严重违反本协议或存在作弊/破解行为，平台有权单方面冻结或封禁违规账号且不予退款。</p>
          </div>

          <div v-if="showDocument === 'privacy'" class="doc-text">
            <h4>1. 信息的收集</h4>
            <p>为了向您提供优质的医学题库和个性化学习服务，我们会在您注册及使用过程中收集以下信息：您主动提供的基础信息（如用户名、邮箱、昵称、学校/专业/年级等），以及您在平台上的学习记录（如答题记录、错题本、检索与浏览行为）。</p>
            <h4>2. 信息的使用</h4>
            <p>我们收集的信息将主要用于：为您提供基础题库服务；通过 AI 模型为您分析薄弱知识点并提供个性化建议；发送服务通知、账号验证或重置邮件；以及在进行数据去标识化处理后，用于整体医学学习数据的统计与算法优化。</p>
            <h4>3. 信息的保护与共享</h4>
            <p>我们采用行业标准的数据安全措施（如加密传输、安全存储机制）来尽力保护您的个人信息。未经您的明确授权，我们绝不会向任何第三方出售或非法共享您的个人数据，除非依照国家法律法规的强制性规定或司法机关的要求。</p>
            <h4>4. 您的权利</h4>
            <p>您拥有访问、更正、更新及删除个人信息的权利。您可以通过“个人中心”自行修改您的资料，或联系管理员申请注销账号。账号注销后，您的个人隐私信息将被脱敏处理或彻底删除。</p>
          </div>

          <div v-if="showDocument === 'disclaimer'" class="doc-text">
            <h4>1. 临床指导限制声明（核心说明）</h4>
            <p>“题酷”平台提供的所有医学题库、知识点解析、AI助教问答及其他相关衍生内容，<strong>仅供医学考试复习、学术交流和医学基础知识学习使用，绝对不能替代专业执业医师的临床诊断、治疗建议或正式医疗指导。</strong>任何因直接或间接参考本平台内容而导致的实际医疗事故、临床偏差或人身伤害，平台概不承担任何法律责任。</p>
            <h4>2. 内容准确性说明</h4>
            <p>平台尽最大努力保障题库及解析的准确性和时效性。但鉴于医学科学的不断发展、医学教材的更新迭代以及各医学院校考察侧重点的不同，平台不对内容的绝对正确性和完整性作任何明示或暗示的担保。如您在刷题过程中发现错漏，请通过“题目纠错”功能向我们反馈。</p>
            <h4>3. 网络服务中断及不可抗力</h4>
            <p>因黑客攻击、计算机病毒侵入、电信部门技术调整、第三方云服务商故障或不可抗力等非平台主观故意的原因，导致的服务异常中断、响应延迟或部分数据丢失，平台不承担相关法律责任，但我们将尽最大努力在第一时间进行修复，减少对您复习备考造成的影响。</p>
          </div>
        </div>
        <div class="modal-footer"><button class="agree-modal-btn" @click="agreeDocument">阅读并同意</button></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 🔥 新增专属落地页样式 */
.email-sent-state { text-align: center; padding: 20px 0; }
.sent-icon-wrapper { margin-bottom: 24px; animation: bounce 1.5s infinite; }
.sent-title { font-size: 1.5rem; font-weight: 700; color: #0f172a; margin-bottom: 16px; }
.sent-desc { font-size: 1rem; color: #475569; line-height: 1.6; margin-bottom: 24px; }
.sent-desc strong { color: #1e293b; font-weight: 700; }
.sent-tip { background-color: #f1f5f9; padding: 12px 16px; border-radius: 8px; font-size: 0.875rem; color: #64748b; margin-bottom: 32px; text-align: left; }
.fade-enter { animation: fadeIn 0.4s ease-out forwards; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
@keyframes bounce { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-8px); } }

/* 验证码与其他原有样式保持不变 */
.captcha-box { background: #fff; padding: 16px; border-radius: 8px; width: 330px; min-height: 280px; display: flex; justify-content: center; align-items: center; flex-direction: column; box-shadow: 0 4px 16px rgba(0,0,0,0.15); }
.status-state { display: flex; flex-direction: column; align-items: center; }
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
.agreement-section { display: flex; align-items: flex-start; gap: 0.5rem; margin-top: 0.25rem; }
.agreement-label { display: flex; align-items: center; cursor: pointer; position: relative; }
.agreement-checkbox { position: absolute; opacity: 0; width: 0; height: 0; }
.custom-checkbox-ui { width: 1rem; height: 1rem; border: 1px solid #cbd5e1; border-radius: 0.25rem; display: flex; align-items: center; justify-content: center; margin-right: 0.5rem; transition: all 0.2s; flex-shrink: 0; }
.agreement-checkbox:checked + .custom-checkbox-ui { background-color: #2563eb; border-color: #2563eb; }
.agreement-checkbox:focus + .custom-checkbox-ui { box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2); }
.agreement-text { font-size: 0.75rem; color: #64748b; line-height: 1.5; user-select: none; }
.link { color: #2563eb; font-weight: 500; cursor: pointer; }
.link:hover { text-decoration: underline; color: #1d4ed8; }
.modal-overlay { position: fixed; inset: 0; z-index: 50; display: flex; align-items: center; justify-content: center; background-color: rgba(0, 0, 0, 0.5); backdrop-filter: blur(4px); padding: 1rem; }
.modal-content { background-color: white; border-radius: 1rem; width: 100%; max-width: 32rem; max-height: 80vh; display: flex; flex-direction: column; overflow: hidden; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25); animation: modalIn 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes modalIn { from { opacity: 0; transform: scale(0.95) translateY(10px); } to { opacity: 1; transform: scale(1) translateY(0); } }
.modal-header { padding: 1rem; border-bottom: 1px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; background-color: white; z-index: 10; }
.modal-title { font-size: 1.125rem; font-weight: 700; color: #0f172a; display: flex; align-items: center; gap: 0.5rem; margin: 0; }
.close-btn { width: 2rem; height: 2rem; border-radius: 9999px; background-color: #f8fafc; border: none; display: flex; align-items: center; justify-content: center; color: #94a3b8; cursor: pointer; transition: all 0.2s; }
.close-btn:hover { background-color: #f1f5f9; color: #475569; }
.modal-body { padding: 1.25rem; overflow-y: auto; color: #475569; font-size: 0.875rem; line-height: 1.625; }
.doc-text h4 { font-weight: 700; color: #1e293b; margin-top: 1rem; margin-bottom: 0.5rem; }
.doc-footer-text { margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid #f1f5f9; font-size: 0.75rem; color: #94a3b8; }
.modal-footer { padding: 1rem; border-top: 1px solid #f1f5f9; background-color: #f8fafc; display: flex; justify-content: flex-end; }
.agree-modal-btn { background-color: #0f172a; color: white; font-weight: 700; padding: 0.5rem 1.5rem; border-radius: 0.5rem; border: none; cursor: pointer; font-size: 0.875rem; transition: all 0.2s; }
.agree-modal-btn:hover { background-color: #2563eb; }
</style>