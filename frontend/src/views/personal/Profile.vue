<script setup lang="ts">
import { ref, onMounted, reactive, computed, watch } from 'vue'
import { 
  NForm, NFormItem, NInput, NButton, NSelect, 
  useMessage, NIcon, NModal, NSpace, NTabs, NTabPane, NDivider
} from 'naive-ui'
import { 
  PersonOutline, LogOutOutline, CameraOutline, 
  SchoolOutline, BookOutline, CalendarOutline,
  CreateOutline, LockClosedOutline
} from '@vicons/ionicons5'
import 'vue-cropper/dist/index.css' 
import { VueCropper } from 'vue-cropper'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'
import { useRouter } from 'vue-router'

const message = useMessage()
const userStore = useUserStore()
const router = useRouter()

// =====================
// 常量与选项
// =====================
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

// =====================
// 状态
// =====================
const user = ref<any>({})
const editLoading = ref(false)

// 头像裁剪
const showCropper = ref(false)
const cropperRef = ref()
const uploadLoading = ref(false)
const cropperOptions = reactive({
  img: '', autoCrop: true, autoCropWidth: 200, autoCropHeight: 200,
  fixedBox: false, fixed: true, fixedNumber: [1, 1], centerBox: true, infoTrue: true
})

// 表单数据
const formModel = reactive({
  nickname: '',
  school: '',
  major: '', 
  grade: null as string | null, 
  qq: '', wechat: '', email: '', gender: 0
})

const majorSelectValue = ref<string | null>(null)
const majorCustomValue = ref('')

const pwdModel = reactive({ old_password: '', new_password: '', confirm_password: '' })
const pwdLoading = ref(false)

// =====================
// 逻辑方法
// =====================
const fetchProfile = async () => {
  try {
    const res: any = await request.get('/user/profile')
    user.value = res.data
    syncFormModel(res.data)
  } catch {}
}

const syncFormModel = (userData: any) => {
    Object.assign(formModel, userData)
    if (!formModel.nickname) formModel.nickname = userData.username
    if (formModel.major) {
        const exists = MAJOR_OPTIONS.some(opt => opt.value === formModel.major)
        if (exists) {
            majorSelectValue.value = formModel.major
        } else {
            majorSelectValue.value = 'other'
            majorCustomValue.value = formModel.major
        }
    }
}

watch([majorSelectValue, majorCustomValue], () => {
    if (majorSelectValue.value === 'other') {
        formModel.major = majorCustomValue.value
    } else {
        formModel.major = majorSelectValue.value || ''
    }
})

const handleUpdateProfile = async () => {
  editLoading.value = true
  try {
    if (!formModel.major) { message.warning('请选择或填写专业'); editLoading.value = false; return }
    if (!formModel.grade) { message.warning('请选择年级'); editLoading.value = false; return }

    await request.put('/user/profile', formModel)
    message.success('资料保存成功')
    
    // 🔥🔥🔥 关键：同时刷新 Store 中的用户信息，保证左侧菜单同步更新 🔥🔥🔥
    await userStore.fetchProfile()
    
    await fetchProfile()
  } catch {
    message.error('保存失败')
  } finally {
    editLoading.value = false
  }
}

const handleChangePwd = async () => {
  if (pwdModel.new_password !== pwdModel.confirm_password) {
    message.error('两次新密码输入不一致'); return
  }
  if (!pwdModel.old_password || !pwdModel.new_password) {
      message.warning('请填写完整'); return
  }
  pwdLoading.value = true
  try {
    await request.put('/user/password', {
      old_password: pwdModel.old_password,
      new_password: pwdModel.new_password
    })
    message.success('密码修改成功，请重新登录')
    userStore.logout()
    router.push('/login')
  } catch(e: any) {
    message.error(e.response?.data?.error || '修改失败')
  } finally {
      pwdLoading.value = false
  }
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// 头像
const fileInput = ref<HTMLInputElement | null>(null)
const triggerFileSelect = () => { fileInput.value?.click() }
const onFileSelected = (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (file) {
        const reader = new FileReader()
        reader.readAsDataURL(file)
        reader.onload = (evt: any) => {
            cropperOptions.img = evt.target.result
            showCropper.value = true
            target.value = '' // reset
        }
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
        // 🔥🔥🔥 关键：使用 Store 的新方法更新并持久化头像 🔥🔥🔥
        userStore.setAvatar(res.url)
        
        user.value.avatar = res.url // 更新当前页面显示
        message.success('头像更新成功')
        showCropper.value = false
      }
    } catch (e) { message.error('上传失败') } finally { uploadLoading.value = false }
  })
}

const getAvatarUrl = (path: string) => {
  if (!path) return undefined // NAvatar 会显示 fallback
  return path.startsWith('http') ? path : `http://localhost:8080${path}`
}

onMounted(fetchProfile)
</script>

<template>
  <div class="profile-page">
    <div class="animate-fade-in space-y-6">
      
      <div class="profile-header-card">
        <div class="logout-btn-wrapper">
            <button class="logout-btn" @click="handleLogout">
                <n-icon size="18"><LogOutOutline /></n-icon>
                <span class="btn-text">退出登录</span>
            </button>
        </div>

        <div class="profile-content">
            <div class="avatar-section">
                <div class="avatar-wrapper group" @click="triggerFileSelect">
                    <div class="avatar-circle" :class="{ 'has-img': !!user.avatar }">
                        <img v-if="user.avatar" :src="getAvatarUrl(user.avatar)" class="avatar-img" />
                        <n-icon v-else size="48" color="#fff"><PersonOutline /></n-icon>
                        <div class="avatar-overlay"><n-icon size="24" color="#fff"><CameraOutline /></n-icon></div>
                    </div>
                </div>
                <button class="edit-avatar-btn" @click="triggerFileSelect">更换头像</button>
                <input type="file" ref="fileInput" class="hidden-input" accept="image/*" @change="onFileSelected" />
            </div>

            <div class="user-details">
                <h2 class="user-nickname">{{ user.nickname || user.username }}</h2>
                <div class="user-meta-row">
                    <span v-if="user.role === 'admin'" class="role-tag admin">管理员</span>
                    <span v-else-if="user.subscriptions && user.subscriptions.length > 0" class="role-tag pro">Pro 会员</span>
                    <span v-else class="role-tag normal">普通用户</span>
                </div>
                <div class="badge-row">
                    <div class="info-badge" v-if="user.school">
                        <n-icon class="badge-icon"><SchoolOutline /></n-icon> 
                        {{ user.school }}
                    </div>
                    <div class="info-badge" v-if="user.major">
                        <n-icon class="badge-icon"><BookOutline /></n-icon> 
                        {{ user.major }}
                    </div>
                    <div class="info-badge" v-if="user.grade">
                        <n-icon class="badge-icon"><CalendarOutline /></n-icon> 
                        {{ user.grade }}
                    </div>
                </div>
            </div>
        </div>
      </div>

      <div class="settings-card">
        <div class="card-header">
            <h3 class="card-title">账户设置</h3>
            <p class="card-desc">修改您的个人资料与安全设置</p>
        </div>
        
        <n-tabs type="line" animated>
            <n-tab-pane name="profile" tab="基本资料">
                <template #tab>
                    <div class="tab-label"><n-icon><CreateOutline /></n-icon> 基本资料</div>
                </template>
                
                <div class="form-wrapper">
                    <n-form label-placement="left" label-width="90" require-mark-placement="right-hanging">
                        <n-form-item label="昵称">
                            <n-input v-model:value="formModel.nickname" placeholder="大家怎么称呼你" />
                        </n-form-item>
                        <n-form-item label="性别">
                            <n-select v-model:value="formModel.gender" :options="genderOptions" />
                        </n-form-item>
                        <n-divider />
                        <n-form-item label="学校">
                            <n-input v-model:value="formModel.school" placeholder="例如：中山大学" />
                        </n-form-item>
                        <n-form-item label="专业">
                             <n-space vertical style="width: 100%">
                                <n-select v-model:value="majorSelectValue" :options="MAJOR_OPTIONS" placeholder="选择专业" />
                                <n-input v-if="majorSelectValue === 'other'" v-model:value="majorCustomValue" placeholder="请输入你的具体专业" />
                            </n-space>
                        </n-form-item>
                        <n-form-item label="年级">
                            <n-select v-model:value="formModel.grade" :options="GRADE_OPTIONS" placeholder="入学年份" />
                        </n-form-item>
                        <n-divider />
                        <n-form-item label="邮箱">
                             <n-input v-model:value="formModel.email" placeholder="Contact Email" />
                        </n-form-item>
                        
                        <div class="form-actions">
                            <n-button type="primary" size="large" @click="handleUpdateProfile" :loading="editLoading" class="save-btn">
                                保存修改
                            </n-button>
                        </div>
                    </n-form>
                </div>
            </n-tab-pane>

            <n-tab-pane name="security" tab="安全设置">
                <template #tab>
                    <div class="tab-label"><n-icon><LockClosedOutline /></n-icon> 安全设置</div>
                </template>
                <div class="form-wrapper">
                     <p class="section-tip">为了您的账号安全，请定期修改密码。</p>
                     <n-form label-placement="left" label-width="90">
                        <n-form-item label="原密码">
                            <n-input type="password" show-password-on="click" v-model:value="pwdModel.old_password" placeholder="请输入当前密码" />
                        </n-form-item>
                        <n-form-item label="新密码">
                            <n-input type="password" show-password-on="click" v-model:value="pwdModel.new_password" placeholder="至少6位" />
                        </n-form-item>
                        <n-form-item label="确认新密码">
                            <n-input type="password" show-password-on="click" v-model:value="pwdModel.confirm_password" placeholder="再次输入新密码" />
                        </n-form-item>
                        
                        <div class="form-actions">
                            <n-button type="warning" size="large" @click="handleChangePwd" :loading="pwdLoading" class="save-btn">
                                修改密码
                            </n-button>
                        </div>
                     </n-form>
                </div>
            </n-tab-pane>
        </n-tabs>
      </div>

    </div>

    <n-modal v-model:show="showCropper" preset="card" title="更换头像" style="width: 600px">
      <div style="width: 100%; height: 400px;">
        <vue-cropper ref="cropperRef" :img="cropperOptions.img" :output-size="1" :output-type="'png'" :info="true" :can-scale="true" :auto-crop="true" :auto-crop-width="200" :auto-crop-height="200" :fixed="true" :fixed-number="[1, 1]" :center-box="true"></vue-cropper>
      </div>
      <template #footer>
        <n-space justify="end">
           <n-button @click="showCropper = false">取消</n-button>
           <n-button type="primary" @click="handleCropConfirm" :loading="uploadLoading">确认并上传</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.profile-page {
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
  box-sizing: border-box;
  padding-bottom: 50px;
}

.animate-fade-in { 
    animation: fadeIn 0.5s ease-out; 
    display: flex;
    flex-direction: column;
    gap: 24px;
}
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* === Header Card === */
.profile-header-card {
    background: #fff;
    border-radius: 20px;
    padding: 32px;
    border: 1px solid #f1f5f9;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
    position: relative;
}

.profile-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
    width: 100%;
}

@media (min-width: 768px) {
    .profile-content {
        flex-direction: row;
        align-items: flex-start; /* Align top or center? Center usually looks better for profile header */
        align-items: center;
        text-align: left;
    }
}

.logout-btn-wrapper { position: absolute; top: 24px; right: 24px; z-index: 10; }
.logout-btn {
    display: flex; align-items: center; gap: 8px; padding: 8px 16px;
    background: #f8fafc; color: #64748b; border: none; border-radius: 12px;
    font-weight: 600; font-size: 14px; cursor: pointer; transition: all 0.2s;
}
.logout-btn:hover { background: #fff1f2; color: #e11d48; }
.btn-text { display: none; }
@media (min-width: 640px) { .btn-text { display: inline; } }

.avatar-section { 
    display: flex; 
    flex-direction: column; 
    align-items: center; 
    gap: 12px; 
    flex-shrink: 0; /* Prevent avatar from shrinking */
}

.avatar-wrapper { position: relative; cursor: pointer; }
.avatar-circle {
    width: 100px; height: 100px; border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6, #2dd4bf);
    display: flex; align-items: center; justify-content: center;
    overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.avatar-circle.has-img { background: #fff; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-overlay {
    position: absolute; inset: 0; background: rgba(0,0,0,0.3);
    display: flex; align-items: center; justify-content: center;
    opacity: 0; transition: opacity 0.2s;
}
.avatar-wrapper:hover .avatar-overlay { opacity: 1; }
.edit-avatar-btn {
    font-size: 12px; font-weight: bold; color: #64748b;
    background: #f1f5f9; padding: 4px 12px; border-radius: 20px;
    border: none; cursor: pointer;
}
.edit-avatar-btn:hover { background: #e2e8f0; }
.hidden-input { display: none; }

.user-details { flex: 1; text-align: center; display: flex; flex-direction: column; justify-content: center; }
@media (min-width: 768px) { 
    .user-details { text-align: left; align-items: flex-start; } 
}
.user-nickname { margin: 0; font-size: 26px; font-weight: 800; color: #0f172a; }
.user-meta-row { margin-top: 8px; }
.role-tag { font-weight: bold; font-size: 13px; }
.role-tag.admin { color: #f59e0b; }
.role-tag.pro { color: #f59e0b; }
.role-tag.normal { color: #64748b; }
.badge-row { margin-top: 16px; display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; }
@media (min-width: 768px) { .badge-row { justify-content: flex-start; } }
.info-badge {
    display: flex; align-items: center; gap: 6px; padding: 6px 14px;
    background: #f1f5f9; color: #475569; border-radius: 8px;
    font-size: 13px; font-weight: 600;
}
.badge-icon { color: #3b82f6; }

/* === Settings Card === */
.settings-card {
    background: #fff;
    border-radius: 20px;
    padding: 32px 32px 40px; /* More padding at bottom */
    border: 1px solid #f1f5f9;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
}

.card-header { margin-bottom: 24px; border-bottom: 1px solid #f1f5f9; padding-bottom: 16px; }
.card-title { margin: 0; font-size: 18px; font-weight: 700; color: #0f172a; }
.card-desc { margin: 4px 0 0; font-size: 14px; color: #64748b; }

.tab-label { display: flex; align-items: center; gap: 6px; font-weight: 600; }
.form-wrapper { max-width: 500px; margin: 24px auto 0; }
.section-tip { font-size: 14px; color: #64748b; margin-bottom: 20px; text-align: center; }

.form-actions { margin-top: 32px; display: flex; justify-content: center; }
.save-btn { width: 100%; font-weight: bold; border-radius: 12px; }

:deep(.n-tabs .n-tabs-nav.n-tabs-nav--line-type .n-tabs-nav-scroll-content) { border-bottom: none; }
</style>