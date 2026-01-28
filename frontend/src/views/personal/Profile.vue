<script setup lang="ts">
import { ref, onMounted, reactive, computed, watch } from 'vue'
import { 
  NCard, NTabs, NTabPane, NForm, NFormItem, NInput, NButton, NUpload, NAvatar, 
  NGrid, NGi, NSelect, NDivider, useMessage, NTag, NIcon, NModal, NSpace, NUploadTrigger 
} from 'naive-ui'
import { 
  SchoolOutline, MailOutline, CloudUploadOutline, CheckmarkOutline 
} from '@vicons/ionicons5'
import 'vue-cropper/dist/index.css' // 👈 别忘了引入样式
import { VueCropper } from 'vue-cropper'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const message = useMessage()
const userStore = useUserStore()

// === 1. 定义常量数据 ===
// 常见医学专业列表
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

// 动态生成近12年的年级
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

// === 基础状态 ===
const user = ref<any>({})
const loading = ref(false)

// === 头像剪裁状态 ===
const showCropper = ref(false)
const cropperRef = ref()
const uploadLoading = ref(false)
const cropperOptions = reactive({
  img: '', autoCrop: true, autoCropWidth: 200, autoCropHeight: 200,
  fixedBox: false, fixed: true, fixedNumber: [1, 1], centerBox: true, infoTrue: true
})

// === 表单数据 ===
const formModel = reactive({
  nickname: '',
  school: '',
  major: '', 
  grade: null as string | null, 
  qq: '', wechat: '', email: '', gender: 0
})

// 🔥 辅助变量：用于控制专业选择逻辑
const majorSelectValue = ref<string | null>(null) // 下拉框选的值
const majorCustomValue = ref('') // 自定义输入框的值

const pwdModel = reactive({ old_password: '', new_password: '', confirm_password: '' })

// === 逻辑方法 ===

// 获取资料并回显
const fetchProfile = async () => {
  try {
    const res: any = await request.get('/user/profile')
    user.value = res.data
    Object.assign(formModel, res.data)
    if (!formModel.nickname) formModel.nickname = user.value.username

    // 🔥 回显专业逻辑：
    if (formModel.major) {
        const exists = MAJOR_OPTIONS.some(opt => opt.value === formModel.major)
        if (exists) {
            majorSelectValue.value = formModel.major
        } else {
            majorSelectValue.value = 'other'
            majorCustomValue.value = formModel.major
        }
    }
  } catch {}
}

// 🔥 监听专业选择变化，同步到 formModel.major
watch([majorSelectValue, majorCustomValue], () => {
    if (majorSelectValue.value === 'other') {
        formModel.major = majorCustomValue.value // 取输入框的值
    } else {
        formModel.major = majorSelectValue.value || '' // 取下拉框的值
    }
})

// 更新资料
const handleUpdateProfile = async () => {
  loading.value = true
  try {
    // 简单校验
    if (!formModel.major) {
        message.warning('请选择或填写专业')
        loading.value = false
        return
    }
    if (!formModel.grade) {
        message.warning('请选择年级')
        loading.value = false
        return
    }

    await request.put('/user/profile', formModel)
    message.success('资料保存成功')
    await fetchProfile()
    userStore.username = formModel.nickname || user.value.username
  } catch {
    message.error('保存失败')
  } finally {
    loading.value = false
  }
}

const handleChangePwd = async () => {
  if (pwdModel.new_password !== pwdModel.confirm_password) {
    message.error('两次新密码输入不一致'); return
  }
  try {
    await request.put('/user/password', {
      old_password: pwdModel.old_password,
      new_password: pwdModel.new_password
    })
    message.success('密码修改成功，请重新登录')
    userStore.logout()
    window.location.href = '/login'
  } catch(e: any) {
    message.error(e.response?.data?.error || '修改失败')
  }
}

// 头像相关
const onSelectFile = async ({ file }: any) => {
  const reader = new FileReader()
  reader.readAsDataURL(file.file)
  reader.onload = (e: any) => { cropperOptions.img = e.target.result; showCropper.value = true }
  return false
}
const handleCropConfirm = () => {
  uploadLoading.value = true
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', blob, 'avatar.png') 
      const res: any = await request.post('/user/avatar', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
      if (res.url) {
        user.value.avatar = res.url + '?t=' + new Date().getTime()
        userStore.avatar = res.url // 同步更新store
        message.success('头像更新成功')
        showCropper.value = false
      }
    } catch (e) { message.error('上传失败') } finally { uploadLoading.value = false }
  })
}

const getAvatarUrl = (path: string) => {
  if (!path) return ''
  return `http://localhost:8080${path}`
}

onMounted(fetchProfile)
</script>

<template>
  <div class="profile-container">
    <n-grid x-gap="24" cols="1 600:3">
      <n-gi span="1">
        <n-card class="profile-card">
          <div class="avatar-box">
            <n-avatar round :size="120" :src="getAvatarUrl(user.avatar)" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
            <n-upload abstract :show-file-list="false" @before-upload="onSelectFile">
              <n-upload-trigger #="{ handleClick }" abstract>
                <n-button size="small" secondary circle class="upload-btn" @click="handleClick">
                  <template #icon><n-icon><CloudUploadOutline /></n-icon></template>
                </n-button>
              </n-upload-trigger>
            </n-upload>
          </div>
          <h2 class="username">{{ user.nickname || user.username }}</h2>
          <div class="tags">
            <n-tag type="info" size="small" v-if="user.role === 'admin'">超级管理员</n-tag>
            <n-tag type="warning" size="small" v-else-if="user.role === 'agent'">代理商</n-tag>
            <n-tag type="success" size="small" v-else>普通用户</n-tag>
            <n-tag :bordered="false" size="small">{{ user.school || '院校未填' }}</n-tag>
          </div>
          <n-divider />
          <div class="info-item"><n-icon><SchoolOutline /></n-icon> <span>{{ user.major || '未填专业' }} {{ user.grade }}</span></div>
          <div class="info-item"><n-icon><MailOutline /></n-icon> <span>{{ user.email || '未绑定邮箱' }}</span></div>
        </n-card>
      </n-gi>

      <n-gi span="2">
        <n-card>
          <n-tabs type="line" animated>
            <n-tab-pane name="basic" tab="📝 编辑资料">
              <n-form label-placement="left" label-width="80" style="max-width: 500px; margin-top: 20px">
                <n-form-item label="昵称">
                  <n-input v-model:value="formModel.nickname" placeholder="大家怎么称呼你" />
                </n-form-item>
                <n-form-item label="性别">
                  <n-select v-model:value="formModel.gender" :options="genderOptions" />
                </n-form-item>
                
                <n-divider title-placement="left" style="font-size: 12px; color: #999">学籍信息 (请真实填写)</n-divider>
                
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
                    <n-select 
                        v-model:value="formModel.grade" 
                        :options="GRADE_OPTIONS" 
                        placeholder="入学年份" 
                    />
                </n-form-item>

                <n-divider title-placement="left" style="font-size: 12px; color: #999">联系方式</n-divider>

                <n-form-item label="QQ"><n-input v-model:value="formModel.qq" /></n-form-item>
                <n-form-item label="微信号"><n-input v-model:value="formModel.wechat" /></n-form-item>
                <n-form-item label="邮箱"><n-input v-model:value="formModel.email" /></n-form-item>

                <div style="display: flex; justify-content: flex-end">
                  <n-button type="primary" @click="handleUpdateProfile" :loading="loading">保存修改</n-button>
                </div>
              </n-form>
            </n-tab-pane>

            <n-tab-pane name="security" tab="🔒 账号安全">
              <n-form label-placement="left" label-width="100" style="max-width: 400px; margin-top: 20px">
                <n-form-item label="当前密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.old_password" /></n-form-item>
                <n-form-item label="新密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.new_password" placeholder="至少6位" /></n-form-item>
                <n-form-item label="确认新密码"><n-input type="password" show-password-on="click" v-model:value="pwdModel.confirm_password" /></n-form-item>
                <div style="display: flex; justify-content: flex-end"><n-button type="warning" @click="handleChangePwd">修改密码</n-button></div>
              </n-form>
            </n-tab-pane>
          </n-tabs>
        </n-card>
      </n-gi>
    </n-grid>

    <n-modal v-model:show="showCropper" preset="card" title="修改头像" style="width: 600px">
      <div style="width: 100%; height: 400px;">
        <vue-cropper ref="cropperRef" :img="cropperOptions.img" :output-size="1" :output-type="'png'" :info="true" :can-scale="true" :auto-crop="true" :auto-crop-width="200" :auto-crop-height="200" :fixed="true" :fixed-number="[1, 1]" :center-box="true"></vue-cropper>
      </div>
      <template #footer>
        <n-space justify="end">
           <n-button @click="showCropper = false">取消</n-button>
           <n-button type="primary" @click="handleCropConfirm" :loading="uploadLoading"><template #icon><n-icon><CheckmarkOutline /></n-icon></template>确认并上传</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.profile-container { padding: 24px; }
.profile-card { text-align: center; }
.avatar-box { position: relative; display: inline-block; margin-bottom: 16px; }
.upload-btn { position: absolute; bottom: 0; right: 0; box-shadow: 0 2px 8px rgba(0,0,0,0.2); z-index: 10; cursor: pointer; }
.username { margin: 0 0 8px; font-size: 20px; font-weight: bold; }
.tags { display: flex; justify-content: center; gap: 8px; flex-wrap: wrap; margin-bottom: 20px; }
.info-item { display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: 8px; color: #666; }
</style>