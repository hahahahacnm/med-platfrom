<script setup lang="ts">
import { ref, onMounted, reactive, computed, watch, onUnmounted } from 'vue'
import { 
  NForm, NFormItem, NInput, NButton, NSelect, 
  useMessage, NIcon, NModal, NSpace, NTabs, NTabPane, NDivider, NGrid, NGridItem,
  NTag, NInputNumber, NRadioGroup, NRadioButton, NEmpty // 🔥 新增 NEmpty
} from 'naive-ui'
import { 
  PersonOutline, LogOutOutline, CameraOutline, 
  SchoolOutline, BookOutline, CalendarOutline,
  CreateOutline, LockClosedOutline, MailOutline, MaleOutline, FemaleOutline,
  LogoWechat, WalletOutline, AddCircleOutline, TicketOutline, LogoAlipay,
  DiamondOutline, TimeOutline // 🔥 新增权益专属图标
} from '@vicons/ionicons5'
import 'vue-cropper/dist/index.css' 
import { VueCropper } from 'vue-cropper'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'
import { useRouter } from 'vue-router'

const message = useMessage()
const userStore = useUserStore()
const router = useRouter()

// === 常量与选项 ===
const MAJOR_OPTIONS = [
  { label: '临床医学', value: '临床医学' },
  { label: '医学影像学', value: '医学影像学' },
  { label: '麻醉学', value: '麻醉学' },
  { label: '口腔医学', value: '口腔医学' },
  { label: '基础医学', value: '基础医学' },
  { label: '预防医学', value: '预防医学' },
  { label: '护理学', value: '护理学' },
  { label: '药学', value: '药学' },
  { label: '中医学', value: '中医学' },
  { label: '其他 (自定义)', value: 'other' }
]

const GRADE_OPTIONS = computed(() => {
  const currentYear = new Date().getFullYear() + 1 
  const list = []
  for (let i = 0; i < 12; i++) {
    const y = currentYear - i
    list.push({ label: `${y}级`, value: `${y}级` })
  }
  return list
})

const genderOptions = [
  { label: '保密', value: 0 },
  { label: '男', value: 1 },
  { label: '女', value: 2 }
]

const user = ref<any>({})
const editLoading = ref(false)

const showRechargeModal = ref(false)
const rechargeType = ref('cdkey')
const redeemCode = ref('')
const redeemLoading = ref(false)
const rechargeAmount = ref(10) 
const predictedPoints = computed(() => Math.floor(rechargeAmount.value * 10))
const BUY_CDKEY_URL = 'https://your-faka-website.com' 

const showCropper = ref(false)
const cropperRef = ref()
const uploadLoading = ref(false)
const cropperOptions = reactive({
  img: '', autoCrop: true, autoCropWidth: 200, autoCropHeight: 200,
  fixedBox: false, fixed: true, fixedNumber: [1, 1], centerBox: true, infoTrue: true
})

// 表单数据 
const formModel = reactive({
  nickname: '', school: '', major: '', grade: null as string | null, gender: 0
})
const displayEmail = ref('') 
const majorSelectValue = ref<string | null>(null)
const majorCustomValue = ref('')

const pwdModel = reactive({ old_password: '', new_password: '', confirm_password: '' })
const pwdLoading = ref(false)

// 邮箱换绑状态
const showBindModal = ref(false)
const newEmail = ref('')
const bindLoading = ref(false)
const bindCountdown = ref(0)
let bindTimer: any = null

onUnmounted(() => {
  if (bindTimer) clearInterval(bindTimer)
})

const startBindCountdown = () => {
  bindCountdown.value = 60
  if (bindTimer) clearInterval(bindTimer)
  bindTimer = setInterval(() => {
    bindCountdown.value--
    if (bindCountdown.value <= 0) clearInterval(bindTimer)
  }, 1000)
}

const handleBindEmail = async () => {
  if (!/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(newEmail.value)) {
    return message.error('邮箱格式不正确')
  }
  if (bindCountdown.value > 0) return

  bindLoading.value = true
  try {
    await request.post('/user/email/bind', { email: newEmail.value })
    message.success('验证邮件已发送至新邮箱，请查收！(30分钟内有效)')
    startBindCountdown()
  } catch (error: any) {
    message.error(error.response?.data?.error || '发送失败')
  } finally {
    bindLoading.value = false
  }
}

// 🔥 新增：计算用户的商品权益列表
const myProducts = computed(() => {
  return user.value.UserProducts || user.value.user_products || []
})

// 🔥 新增：格式化到期时间 (兼容 2099 年判定为永久有效)
const formatExpireTime = (timeStr: string | undefined) => {
  if (!timeStr) return '永久有效'
  const t = new Date(timeStr)
  if (isNaN(t.getTime())) return '永久有效'
  if (t.getFullYear() > 2090) return '永久有效'
  return t.toLocaleDateString() + ' ' + t.toLocaleTimeString()
}

// 🔥 新增：判断商品是否已过期
const isExpired = (timeStr: string | undefined) => {
  if (!timeStr) return false
  const t = new Date(timeStr)
  if (isNaN(t.getTime())) return false
  if (t.getFullYear() > 2090) return false
  return t.getTime() < Date.now()
}

const fetchProfile = async () => {
  try {
    const res: any = await request.get('/user/profile')
    user.value = res.data
    userStore.points = res.data.points
    userStore.role = res.data.role 
    displayEmail.value = res.data.email
    syncFormModel(res.data)
  } catch {}
}

const syncFormModel = (userData: any) => {
    formModel.nickname = userData.nickname || userData.username
    formModel.school = userData.school
    formModel.grade = userData.grade
    formModel.gender = userData.gender
    if (userData.major) {
        formModel.major = userData.major
        const exists = MAJOR_OPTIONS.some(opt => opt.value === userData.major)
        if (exists) { majorSelectValue.value = userData.major } 
        else { majorSelectValue.value = 'other'; majorCustomValue.value = userData.major }
    }
}

watch([majorSelectValue, majorCustomValue], () => {
    if (majorSelectValue.value === 'other') formModel.major = majorCustomValue.value
    else formModel.major = majorSelectValue.value || ''
})

const handleRedeem = async () => {
    if (!redeemCode.value.trim()) return message.warning('请输入激活码')
    redeemLoading.value = true
    try {
        await request.post('/codes/redeem', { code: redeemCode.value })
        message.success('🎉 兑换成功！积分已到账')
        showRechargeModal.value = false
        redeemCode.value = ''
        await fetchProfile()
    } catch (e: any) { message.error(e.response?.data?.error || '兑换失败') } 
    finally { redeemLoading.value = false }
}

const handleUpdateProfile = async () => {
  editLoading.value = true
  try {
    if (!formModel.major) { message.warning('请选择或填写专业'); editLoading.value = false; return }
    if (!formModel.grade) { message.warning('请选择年级'); editLoading.value = false; return }
    await request.put('/user/profile', formModel)
    message.success('资料保存成功')
    await userStore.fetchProfile()
    await fetchProfile()
  } catch { message.error('保存失败') } finally { editLoading.value = false }
}

const handleChangePwd = async () => {
  if (pwdModel.new_password !== pwdModel.confirm_password) return message.error('两次新密码输入不一致')
  if (!pwdModel.old_password || !pwdModel.new_password) return message.warning('请填写完整')
  pwdLoading.value = true
  try {
    await request.put('/user/password', { old_password: pwdModel.old_password, new_password: pwdModel.new_password })
    message.success('密码修改成功，请重新登录')
    userStore.logout()
    router.push('/login')
  } catch(e: any) { message.error(e.response?.data?.error || '修改失败') } finally { pwdLoading.value = false }
}

const handleLogout = () => { userStore.logout(); router.push('/login') }

const fileInput = ref<HTMLInputElement | null>(null)
const triggerFileSelect = () => { fileInput.value?.click() }
const onFileSelected = (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (file) {
        const reader = new FileReader()
        reader.readAsDataURL(file)
        reader.onload = (evt: any) => { cropperOptions.img = evt.target.result; showCropper.value = true; target.value = '' }
    }
}

const handleCropConfirm = () => {
  uploadLoading.value = true
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', blob, 'avatar.png') 
      const res: any = await request.post('/user/avatar', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
      if (res.url) {
        userStore.setAvatar(res.url)
        user.value.avatar = res.url
        message.success('头像更新成功')
        showCropper.value = false
      }
    } catch (e) { message.error('上传失败') } finally { uploadLoading.value = false }
  })
}

const getAvatarUrl = (path: string) => {
  if (!path) return undefined 
  return path.startsWith('http') ? path : `http://localhost:8080${path}`
}

onMounted(fetchProfile)
</script>

<template>
  <div class="profile-container">
    <div class="profile-bg-decoration"></div>

    <div class="content-wrapper animate-slide-up">
      <div class="profile-header-card">
        <div class="header-inner">
            <div class="action-area">
                <button class="logout-btn" @click="handleLogout">
                    <n-icon size="18"><LogOutOutline /></n-icon>
                    <span class="btn-text">退出</span>
                </button>
            </div>

            <div class="avatar-column">
                <div class="avatar-wrapper group" @click="triggerFileSelect">
                    <div class="avatar-ring">
                        <div class="avatar-circle">
                            <img v-if="user.avatar" :src="getAvatarUrl(user.avatar)" class="avatar-img" />
                            <n-icon v-else size="48" color="#cbd5e1"><PersonOutline /></n-icon>
                            <div class="avatar-overlay"><n-icon size="24" color="#fff"><CameraOutline /></n-icon><span class="overlay-text">更换</span></div>
                        </div>
                    </div>
                </div>
                <input type="file" ref="fileInput" class="hidden-input" accept="image/*" @change="onFileSelected" />
            </div>

            <div class="info-column">
                <div class="name-row">
                    <h2 class="user-nickname">{{ user.nickname || user.username }}</h2>
                    <div class="gender-icon" v-if="user.gender === 1"><n-icon color="#3b82f6"><MaleOutline /></n-icon></div>
                    <div class="gender-icon" v-if="user.gender === 2"><n-icon color="#ec4899"><FemaleOutline /></n-icon></div>
                </div>
                
                <div class="role-row">
                    <span v-if="user.role === 'admin'" class="role-badge admin">管理员</span>
                    <span v-else-if="user.role === 'agent'" class="role-badge agent">授权代理</span>
                    <span v-else-if="user.subscriptions && user.subscriptions.length > 0" class="role-badge pro">Pro 会员</span>
                    <span v-else class="role-badge normal">学生用户</span>
                    <span class="uid-text">ID: {{ user.id }}</span>
                </div>

                <div class="points-row">
                    <div class="points-badge">
                        <n-icon size="16" color="#f59e0b"><WalletOutline /></n-icon>
                        <span class="points-val">{{ userStore.points || 0 }}</span> 
                        <span class="points-label">积分</span>
                    </div>
                    <n-button size="small" type="primary" secondary round @click="showRechargeModal = true">
                        <template #icon><n-icon><AddCircleOutline /></n-icon></template>
                        获取积分
                    </n-button>
                </div>

                <div class="meta-tags">
                    <div class="meta-item" v-if="user.school"><n-icon><SchoolOutline /></n-icon> {{ user.school }}</div>
                    <div class="meta-item" v-if="user.major"><n-icon><BookOutline /></n-icon> {{ user.major }}</div>
                    <div class="meta-item" v-if="user.grade"><n-icon><CalendarOutline /></n-icon> {{ user.grade }}</div>
                </div>
            </div>
        </div>
      </div>

      <div class="settings-card">
        <n-tabs type="line" animated size="large" justify-content="space-evenly" class="custom-tabs">
            
            <n-tab-pane name="profile" tab="基本资料">
                <template #tab><div class="tab-label"><n-icon size="18"><CreateOutline /></n-icon> <span>编辑资料</span></div></template>
                <div class="form-container">
                    <n-form label-placement="top" size="medium">
                        <n-grid :x-gap="24" :y-gap="12" :cols="1" responsive="screen" item-responsive>
                            <n-grid-item span="0:1 640:1"><n-form-item label="昵称"><n-input v-model:value="formModel.nickname" placeholder="大家怎么称呼你" maxlength="20" show-count /></n-form-item></n-grid-item>
                            <n-grid-item span="0:1 640:1"><n-form-item label="性别"><n-select v-model:value="formModel.gender" :options="genderOptions" /></n-form-item></n-grid-item>
                        </n-grid>
                        <n-divider dashed style="margin: 12px 0 24px 0" />
                        <n-grid :x-gap="24" :y-gap="12" :cols="2" responsive="screen">
                            <n-grid-item span="2 s:1"><n-form-item label="所在学校"><n-input v-model:value="formModel.school" placeholder="例如：中山大学" ><template #prefix><n-icon color="#94a3b8"><SchoolOutline /></n-icon></template></n-input></n-form-item></n-grid-item>
                            <n-grid-item span="2 s:1">
                                <n-form-item label="主修专业">
                                    <div class="major-input-group">
                                        <n-select v-model:value="majorSelectValue" :options="MAJOR_OPTIONS" placeholder="选择专业" />
                                        <n-input v-if="majorSelectValue === 'other'" v-model:value="majorCustomValue" placeholder="请输入具体专业" style="margin-top: 8px" />
                                    </div>
                                </n-form-item>
                            </n-grid-item>
                        </n-grid>
                        <n-grid :x-gap="24" :y-gap="12" :cols="2" responsive="screen">
                            <n-grid-item span="2 s:1"><n-form-item label="入学年份"><n-select v-model:value="formModel.grade" :options="GRADE_OPTIONS" placeholder="选择年级" /></n-form-item></n-grid-item>
                            
                            <n-grid-item span="2 s:1">
                                <n-form-item label="绑定邮箱">
                                    <n-input :value="displayEmail" disabled placeholder="尚未绑定邮箱">
                                        <template #prefix><n-icon color="#94a3b8"><MailOutline /></n-icon></template>
                                        <template #suffix>
                                            <n-button type="primary" text @click="showBindModal = true">更换</n-button>
                                        </template>
                                    </n-input>
                                </n-form-item>
                            </n-grid-item>
                        </n-grid>

                        <div class="form-actions"><n-button type="primary" size="large" @click="handleUpdateProfile" :loading="editLoading" class="save-btn">保存更改</n-button></div>
                    </n-form>
                </div>
            </n-tab-pane>

            <n-tab-pane name="entitlements" tab="我的权益">
                <template #tab><div class="tab-label"><n-icon size="18"><DiamondOutline /></n-icon> <span>我的权益</span></div></template>
                <div class="form-container" style="background: #fafaf9;">
                    <div v-if="myProducts.length === 0" style="padding: 40px 0;">
                        <n-empty description="您当前暂无任何生效的权益，去获取积分兑换吧~" />
                    </div>
                    
                    <n-grid v-else :x-gap="16" :y-gap="16" :cols="1" responsive="screen" item-responsive>
                        <n-grid-item v-for="(item, index) in myProducts" :key="index" span="0:1 640:1">
                            <div class="product-card">
                                <div class="p-icon"><n-icon size="28" color="#3b82f6"><DiamondOutline/></n-icon></div>
                                <div class="p-info">
                                    <div class="p-name">{{ item.Product?.name || item.product?.name || item.name || '已购权益' }}</div>
                                    <div class="p-time">
                                        <n-icon style="vertical-align: middle; margin-right: 4px;"><TimeOutline/></n-icon>
                                        <span style="vertical-align: middle;">到期时间：{{ formatExpireTime(item.expire_at || item.expires_at || item.expire_time) }}</span>
                                    </div>
                                </div>
                                <div class="p-status">
                                    <n-tag :type="isExpired(item.expire_at || item.expires_at || item.expire_time) ? 'error' : 'success'" round :bordered="false" style="font-weight: bold;">
                                        {{ isExpired(item.expire_at || item.expires_at || item.expire_time) ? '已过期' : '生效中' }}
                                    </n-tag>
                                </div>
                            </div>
                        </n-grid-item>
                    </n-grid>
                </div>
            </n-tab-pane>

            <n-tab-pane name="security" tab="安全设置">
                <template #tab><div class="tab-label"><n-icon size="18"><LockClosedOutline /></n-icon> <span>修改密码</span></div></template>
                <div class="form-container narrow">
                     <div class="security-tip"><n-icon size="20" color="#f59e0b"><LockClosedOutline /></n-icon><p>定期修改密码可以保护您的账号安全。</p></div>
                     <n-form label-placement="top" size="large">
                        <n-form-item label="当前密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.old_password" placeholder="验证当前密码" /></n-form-item>
                        <n-form-item label="新密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.new_password" placeholder="至少6位" /></n-form-item>
                        <n-form-item label="确认新密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.confirm_password" placeholder="再次输入新密码" /></n-form-item>
                        <div class="form-actions"><n-button type="warning" size="large" @click="handleChangePwd" :loading="pwdLoading" class="save-btn">确认修改</n-button></div>
                     </n-form>
                </div>
            </n-tab-pane>
        </n-tabs>
      </div>
    </div>

    <n-modal v-model:show="showBindModal" preset="card" title="更换绑定邮箱" style="width: 400px; max-width: 95vw;">
      <n-form label-placement="top">
        <n-form-item label="输入新的邮箱地址">
          <n-input v-model:value="newEmail" placeholder="如：yourname@example.com" size="large">
             <template #prefix><n-icon color="#94a3b8"><MailOutline /></n-icon></template>
          </n-input>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-button type="primary" block size="large" :loading="bindLoading" :disabled="bindCountdown > 0" @click="handleBindEmail">
          {{ bindCountdown > 0 ? `${bindCountdown} 秒后可重新发送` : '发送确认邮件' }}
        </n-button>
      </template>
    </n-modal>

    <n-modal v-model:show="showCropper" preset="card" title="裁剪头像" style="width: 600px; max-width: 90vw;">
      <div style="width: 100%; height: 350px;">
        <vue-cropper ref="cropperRef" :img="cropperOptions.img" :output-size="1" :output-type="'png'" :info="true" :can-scale="true" :auto-crop="true" :auto-crop-width="200" :auto-crop-height="200" :fixed="true" :fixed-number="[1, 1]" :center-box="true"></vue-cropper>
      </div>
      <template #footer>
        <n-space justify="end"><n-button @click="showCropper = false">取消</n-button><n-button type="primary" @click="handleCropConfirm" :loading="uploadLoading">保存头像</n-button></n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showRechargeModal" preset="card" title="获取积分" style="width: 440px; max-width: 95vw;">
      <div style="margin-bottom: 24px; text-align: center;">
         <n-radio-group v-model:value="rechargeType" size="large">
            <n-radio-button value="cdkey"><n-icon style="margin-right:4px; position:relative; top:2px"><TicketOutline /></n-icon>卡密兑换</n-radio-button>
            <n-radio-button value="online"><n-icon style="margin-right:4px; position:relative; top:2px"><LogoWechat /></n-icon>在线直充</n-radio-button>
         </n-radio-group>
      </div>
      <div v-if="rechargeType === 'cdkey'" class="recharge-content">
        <div class="cdkey-guide"><span>还没有激活码？</span><a :href="BUY_CDKEY_URL" target="_blank" class="buy-link">点击前往发卡网购买 ></a></div>
        <n-form-item label="请输入您的激活码"><n-input v-model:value="redeemCode" placeholder="例如：TK-8A9B-C7D6" size="large" clearable @keydown.enter="handleRedeem"><template #prefix><n-icon color="#18a058"><TicketOutline /></n-icon></template></n-input></n-form-item>
        <n-button type="primary" block size="large" @click="handleRedeem" :loading="redeemLoading" color="#18a058" style="margin-top: 10px">立即兑换</n-button>
      </div>
      <div v-else class="recharge-content online-placeholder">
         <div class="payment-icons"><n-icon size="32" color="#07c160"><LogoWechat /></n-icon><n-icon size="32" color="#1677ff"><LogoAlipay /></n-icon></div>
         <n-form-item label="赞助金额 (元)"><n-input-number v-model:value="rechargeAmount" :min="1" :step="10" size="large" style="width: 100%; text-align: center;" disabled><template #prefix>￥</template></n-input-number></n-form-item>
        <div class="points-preview"><div class="preview-label">预计获得积分</div><div class="preview-val">{{ predictedPoints }}</div><div class="preview-rate">1 元 = 10 积分</div></div>
        <n-button disabled block size="large" style="margin-top: 20px">接口接驳中，敬请期待...</n-button>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
/* 全局容器及原生保持样式 */
.profile-container { min-height: 100%; position: relative; background-color: #f8fafc; }
.profile-bg-decoration { height: 160px; width: 100%; background: linear-gradient(120deg, #dbeafe 0%, #eff6ff 50%, #f0f9ff 100%); position: absolute; top: 0; left: 0; z-index: 0; }
.content-wrapper { position: relative; z-index: 1; max-width: 1000px; margin: 0 auto; padding: 24px 20px 60px; display: flex; flex-direction: column; gap: 24px; }
.animate-slide-up { animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

/* 头部信息卡片 */
.profile-header-card { background: rgba(255, 255, 255, 0.9); backdrop-filter: blur(10px); border-radius: 16px; padding: 24px; box-shadow: 0 4px 20px -4px rgba(0, 0, 0, 0.05); border: 1px solid #fff; margin-top: 40px; }
.header-inner { display: flex; flex-direction: column; align-items: center; position: relative; gap: 20px; }
@media (min-width: 768px) { .header-inner { flex-direction: row; align-items: flex-start; text-align: left; padding-left: 10px; } }
.action-area { position: absolute; top: 0; right: 0; }
.logout-btn { display: flex; align-items: center; gap: 6px; padding: 6px 12px; background: #fff0f0; color: #ef4444; border: 1px solid #fecaca; border-radius: 20px; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.logout-btn:hover { background: #fee2e2; transform: translateY(-1px); }
.btn-text { display: none; }
@media (min-width: 640px) { .btn-text { display: inline; } }

.avatar-column { flex-shrink: 0; }
.avatar-ring { padding: 4px; background: #fff; border-radius: 50%; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.avatar-circle { width: 100px; height: 100px; border-radius: 50%; background: #f1f5f9; display: flex; align-items: center; justify-content: center; overflow: hidden; position: relative; cursor: pointer; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-overlay { position: absolute; inset: 0; background: rgba(0,0,0,0.4); display: flex; flex-direction: column; align-items: center; justify-content: center; opacity: 0; transition: opacity 0.2s; }
.avatar-wrapper:hover .avatar-overlay { opacity: 1; }
.overlay-text { color: #fff; font-size: 12px; margin-top: 4px; font-weight: 500; }
.hidden-input { display: none; }

.info-column { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 8px; }
@media (min-width: 768px) { .info-column { align-items: flex-start; } }
.name-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.user-nickname { font-size: 24px; font-weight: 800; color: #0f172a; margin: 0; line-height: 1.2; }
.gender-icon { display: flex; align-items: center; margin-top: 4px; }
.role-row { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; flex-wrap: wrap; justify-content: center; }
@media (min-width: 768px) { .role-row { justify-content: flex-start; } }
.role-badge { padding: 2px 10px; border-radius: 6px; font-size: 12px; font-weight: 700; letter-spacing: 0.5px; }
.role-badge.admin { background: #fffbeb; color: #b45309; border: 1px solid #fcd34d; }
.role-badge.pro { background: #eff6ff; color: #2563eb; border: 1px solid #bfdbfe; }
.role-badge.normal { background: #f1f5f9; color: #64748b; border: 1px solid #e2e8f0; }
.role-badge.agent { background: #f3e8ff; color: #7c3aed; border: 1px solid #d8b4fe; }
.uid-text { font-size: 12px; color: #94a3b8; font-family: monospace; }

/* 积分栏样式 */
.points-row { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; background: #fefce8; padding: 6px 12px; border-radius: 8px; border: 1px solid #fef08a; }
.points-badge { display: flex; align-items: center; gap: 6px; font-weight: bold; color: #b45309; }
.points-val { font-size: 18px; font-family: monospace; }
.points-label { font-size: 12px; opacity: 0.8; }
.meta-tags { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
@media (min-width: 768px) { .meta-tags { justify-content: flex-start; } }
.meta-item { display: flex; align-items: center; gap: 5px; background: #f8fafc; color: #475569; padding: 6px 12px; border-radius: 8px; font-size: 13px; font-weight: 600; }

/* 设置卡片 */
.settings-card { background: #fff; border-radius: 16px; box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03); overflow: hidden; }
:deep(.n-tabs .n-tabs-nav) { padding: 0 16px; border-bottom: 1px solid #f1f5f9; }
.tab-label { display: flex; align-items: center; gap: 6px; font-weight: 600; font-size: 15px; }
.form-container { padding: 32px 24px 40px; }
.form-container.narrow { max-width: 480px; margin: 0 auto; }
.major-input-group { display: flex; flex-direction: column; width: 100%; }
.form-actions { margin-top: 32px; display: flex; justify-content: center; }
@media (min-width: 768px) { .form-actions { justify-content: flex-end; } }
.save-btn { width: 100%; font-weight: bold; border-radius: 8px; }
@media (min-width: 640px) { .save-btn { width: 180px; } }
.security-tip { background: #fffbeb; border: 1px solid #fcd34d; padding: 12px 16px; border-radius: 8px; display: flex; align-items: center; gap: 10px; margin-bottom: 24px; color: #b45309; font-size: 14px; }

/* 🔥 新增：商品权益卡片专属样式 */
.product-card {
  display: flex;
  align-items: center;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  padding: 18px 20px;
  border-radius: 16px;
  gap: 16px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.02);
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.product-card:hover {
  background: #f0f9ff;
  border-color: #bae6fd;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.08);
}
.p-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  background: #eff6ff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.p-info { flex: 1; }
.p-name {
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 6px;
}
.p-time {
  font-size: 13px;
  color: #64748b;
  display: flex;
  align-items: center;
}

/* 弹窗样式 */
.recharge-content { padding: 10px 0; }
.cdkey-guide {
    background: #f0fdf4; border: 1px solid #bbf7d0; padding: 12px 16px; border-radius: 8px;
    display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; font-size: 13px;
}
.cdkey-guide span { color: #166534; }
.buy-link { color: #15803d; font-weight: bold; text-decoration: none; transition: color 0.2s; }
.buy-link:hover { color: #16a34a; text-decoration: underline; }
.online-placeholder { opacity: 0.8; }
.payment-icons { display: flex; justify-content: center; gap: 20px; margin-bottom: 20px; }
.points-preview { 
    margin-top: 20px; background: #f8fafc; border: 1px solid #e2e8f0; 
    padding: 16px; border-radius: 8px; text-align: center; 
}
.preview-label { font-size: 13px; color: #64748b; margin-bottom: 4px; }
.preview-val { font-size: 28px; font-weight: 800; color: #334155; font-family: monospace; }
.preview-rate { font-size: 12px; color: #94a3b8; margin-top: 4px; }
</style>