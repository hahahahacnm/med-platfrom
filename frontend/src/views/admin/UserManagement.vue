<script setup lang="ts">
import { ref, onMounted, reactive, h, computed, watch } from 'vue'
import { 
  NCard, NDataTable, NTag, NButton, NSpace, NInput, NModal, NSelect, 
  NForm, NFormItem, useMessage, NPopconfirm, NIcon, NPageHeader,
  NGrid, NGi, NRadio, NRadioGroup, NAvatar, NUpload, NUploadTrigger, NSpin
} from 'naive-ui'
import { 
  PersonOutline, ShieldCheckmarkOutline, BanOutline, SearchOutline, 
  CreateOutline, LockOpenOutline, KeyOutline, CloudUploadOutline, CheckmarkOutline
} from '@vicons/ionicons5'
import 'vue-cropper/dist/index.css' // 👈 引入样式
import { VueCropper } from 'vue-cropper' // 👈 引入组件
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const list = ref([])
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const keyword = ref('')

// === 常量定义 ===
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

const roleOptions = [
  { label: '普通用户', value: 'user' },
  { label: '机构代理', value: 'agent' },
  { label: '超级管理员', value: 'admin' }
]
const roleMap: Record<string, string> = { user: '普通用户', agent: '机构代理', admin: '超级管理员' }
const banDurationOptions = [
  { label: '1 天', value: 24 }, { label: '3 天', value: 72 }, { label: '1 周', value: 168 },
  { label: '1 个月', value: 720 }, { label: '永久封禁', value: -1 },
]

// === 模态框状态 ===
const showRoleModal = ref(false)
const showBanModal = ref(false)
const showEditModal = ref(false)
const showResetModal = ref(false)

const currentEditUser = ref<any>(null)
const roleForm = ref({ role: '' })
const banForm = ref({ duration: 24 })

const editForm = reactive({
    id: 0, nickname: '', school: '', major: '', grade: null as string|null, qq: '', wechat: '', email: '', gender: 0, 
    avatar: '' 
})
const adminMajorSelect = ref<string|null>(null)
const adminMajorCustom = ref('')

const resetForm = reactive({ id: 0, new_password: '' })
const submitting = ref(false)

// === 🔥🔥🔥 头像剪裁相关状态 🔥🔥🔥 ===
const showCropper = ref(false)
const cropperRef = ref()
const uploadLoading = ref(false)
const cropperOptions = reactive({
  img: '',           // 剪裁图片的地址
  autoCrop: true,    // 是否默认生成截图框
  autoCropWidth: 200,// 默认生成截图框宽度
  autoCropHeight: 200,// 默认生成截图框高度
  fixedBox: false,   // 固定截图框大小
  fixed: true,       // 是否开启截图框宽高固定比例
  fixedNumber: [1, 1], // 强制 1:1
  centerBox: true,   // 截图框是否被限制在图片里面
  infoTrue: true     // 展示真实输出图片宽高
})

// === 监听管理员的专业选择 ===
watch([adminMajorSelect, adminMajorCustom], () => {
    if (adminMajorSelect.value === 'other') {
        editForm.major = adminMajorCustom.value
    } else {
        editForm.major = adminMajorSelect.value || ''
    }
})

// === 表格列定义 ===
const columns = [
  { title: 'ID', key: 'id', width: 60, fixed: 'left' },
  { 
    title: '用户', key: 'username', width: 160, fixed: 'left',
    render(row: any) {
        return h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
            h(NAvatar, { round: true, size: 'small', src: row.avatar ? `http://localhost:8080${row.avatar}` : undefined, fallbackSrc: 'https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg' }),
            h('div', [
                h('div', { style: 'font-weight: bold' }, row.nickname || row.username),
                h('div', { style: 'font-size: 12px; color: #999' }, row.username)
            ])
        ])
    }
  },
  { 
    title: '角色', key: 'role', width: 100,
    render(row: any) {
      const type = row.role === 'admin' ? 'error' : (row.role === 'agent' ? 'warning' : 'default')
      return h(NTag, { type, bordered: false, size: 'small' }, { default: () => roleMap[row.role] || row.role })
    }
  },
  { title: '🏫 学校', key: 'school', width: 140, ellipsis: { tooltip: true }, render: (row: any) => row.school || '-' },
  { title: '📚 专业', key: 'major', width: 120, ellipsis: { tooltip: true }, render: (row: any) => row.major || '-' },
  { 
    title: '🎓 年级', key: 'grade', width: 90,
    render(row: any) {
        if (!row.grade) return '-'
        return h(NTag, { size: 'small', bordered: false, type: 'info', style: 'opacity: 0.8' }, { default: () => row.grade })
    }
  },
  { 
    title: '状态', key: 'status', width: 100,
    render(row: any) {
      if (row.status === 2) {
        return h(NTag, { type: 'error', size: 'small' }, { default: () => '封禁中' })
      }
      return h(NTag, { type: 'success', bordered: false, size: 'small' }, { default: () => '正常' })
    }
  },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 220,
    render(row: any) {
      return h(NSpace, { justify: 'center', size: 'small' }, {
        default: () => [
          h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => openEditModal(row) }, 
            { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => '资料' }),
          h(NButton, { size: 'tiny', onClick: () => openRoleModal(row) }, { default: () => '角色' }),
          h(NButton, { size: 'tiny', type: 'warning', dashed: true, onClick: () => openResetPwdModal(row) }, 
             { icon: () => h(NIcon, null, { default: () => h(KeyOutline) }) }),
          row.status === 1 
            ? h(NButton, { size: 'tiny', type: 'error', ghost: true, onClick: () => openBanModal(row) }, { default: () => '封' })
            : h(NPopconfirm, { onPositiveClick: () => handleUnban(row.id) }, { 
                trigger: () => h(NButton, { size: 'tiny', type: 'success' }, { default: () => '解' }),
                default: () => '确定要解封该用户吗？'
              })
        ]
      })
    }
  }
]

// === API 操作 ===
const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/users', { params: { page: pagination.page, page_size: pagination.pageSize, keyword: keyword.value } })
    list.value = res.data || []; pagination.itemCount = res.total || 0
  } catch (e) { message.error('加载失败') } finally { loading.value = false }
}
const handleSearch = () => { pagination.page = 1; fetchData() }

const openRoleModal = (user: any) => { currentEditUser.value = user; roleForm.value.role = user.role; showRoleModal.value = true }
const submitRole = async () => {
  submitting.value = true
  try { await request.post('/admin/users/role', { user_id: currentEditUser.value.id, new_role: roleForm.value.role }); message.success('角色修改成功'); showRoleModal.value = false; fetchData() } catch { message.error('操作失败') } finally { submitting.value = false }
}
const openBanModal = (user: any) => { currentEditUser.value = user; banForm.value.duration = 24; showBanModal.value = true }
const submitBan = async () => {
  submitting.value = true
  try { await request.post('/admin/users/ban', { user_id: currentEditUser.value.id, duration: banForm.value.duration }); message.success('用户已封禁'); showBanModal.value = false; fetchData() } catch { message.error('操作失败') } finally { submitting.value = false }
}
const handleUnban = async (id: number) => { try { await request.post('/admin/users/unban', { user_id: id }); message.success('已解封'); fetchData() } catch { message.error('操作失败') } }

// === 编辑资料逻辑 ===
const openEditModal = (row: any) => {
    editForm.id = row.id; editForm.nickname = row.nickname; editForm.school = row.school; editForm.major = row.major; editForm.grade = row.grade;
    editForm.qq = row.qq; editForm.wechat = row.wechat; editForm.email = row.email; editForm.gender = row.gender; editForm.avatar = row.avatar;
    
    // 回显专业逻辑
    if (editForm.major) {
        const exists = MAJOR_OPTIONS.some(opt => opt.value === editForm.major)
        if (exists) { adminMajorSelect.value = editForm.major } 
        else { adminMajorSelect.value = 'other'; adminMajorCustom.value = editForm.major }
    } else {
        adminMajorSelect.value = null; adminMajorCustom.value = ''
    }
    showEditModal.value = true
}
const handleSaveUser = async () => {
    try { await request.put(`/admin/users/${editForm.id}`, editForm); message.success('用户资料更新成功'); showEditModal.value = false; fetchData() } catch { message.error('更新失败') }
}

// === 🔥🔥🔥 核心新增：头像上传拦截逻辑 🔥🔥🔥 ===
const onSelectFile = async ({ file }: any) => {
  const reader = new FileReader()
  // 读取文件为 DataURL 以便预览
  reader.readAsDataURL(file.file)
  reader.onload = (e: any) => {
    cropperOptions.img = e.target.result // 塞入剪裁器
    showCropper.value = true // 打开剪裁弹窗
  }
  return false // 阻止默认的自动上传
}

// === 🔥🔥🔥 核心新增：剪裁并上传到 管理员专用接口 🔥🔥🔥 ===
const handleCropConfirm = () => {
  uploadLoading.value = true
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', blob, 'avatar.png') 
      
      // 注意：这里调用的是 /admin/users/:id/avatar 接口
      const res: any = await request.post(`/admin/users/${editForm.id}/avatar`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })

      if (res.url) {
        editForm.avatar = res.url + '?t=' + new Date().getTime() // 刷新编辑框的预览
        message.success('头像强制修改成功')
        showCropper.value = false // 关闭剪裁窗
        fetchData() // 刷新列表
      }
    } catch (e) {
      message.error('上传失败')
    } finally {
      uploadLoading.value = false
    }
  })
}

const openResetPwdModal = (row: any) => { resetForm.id = row.id; resetForm.new_password = ''; showResetModal.value = true }
const handleConfirmReset = async () => {
    if(resetForm.new_password.length < 6) return message.warning('密码至少6位');
    try { await request.put(`/admin/users/${resetForm.id}/password`, { new_password: resetForm.new_password }); message.success('密码重置成功'); showResetModal.value = false } catch { message.error('重置失败') }
}

const handlePageChange = (page: number) => { pagination.page = page; fetchData() }
onMounted(fetchData)
</script>

<template>
  <div class="user-manage-container">
    <n-page-header title="👥 用户管理" subtitle="系统层级：资料修改、角色分配与封号" style="margin-bottom: 24px;"> </n-page-header>
    <n-card>
      <div class="toolbar">
        <n-input v-model:value="keyword" placeholder="搜索用户名/昵称..." style="width: 240px" @keydown.enter="handleSearch"><template #prefix><n-icon><SearchOutline /></n-icon></template></n-input>
        <n-button type="primary" @click="handleSearch">搜索</n-button>
      </div>
      <n-data-table remote :columns="columns" :data="list" :loading="loading" :pagination="pagination" @update:page="handlePageChange" style="margin-top: 16px;" :scroll-x="1200" />
    </n-card>

    <n-modal v-model:show="showRoleModal" preset="card" title="修改用户角色" style="width: 400px">
      <n-form><n-form-item label="当前用户"><n-input :value="currentEditUser?.username" disabled /></n-form-item><n-form-item label="选择新角色"><n-select v-model:value="roleForm.role" :options="roleOptions" /></n-form-item></n-form>
      <template #footer><div style="text-align: right;"><n-button @click="showRoleModal = false" style="margin-right: 10px;">取消</n-button><n-button type="primary" :loading="submitting" @click="submitRole">保存</n-button></div></template>
    </n-modal>
    <n-modal v-model:show="showBanModal" preset="card" title="账号封禁" style="width: 400px">
      <n-form><n-form-item label="封禁对象"><n-input :value="currentEditUser?.username" disabled /></n-form-item><n-form-item label="封禁时长"><n-select v-model:value="banForm.duration" :options="banDurationOptions" /></n-form-item><n-alert type="warning" :show-icon="false" v-if="banForm.duration === -1">注意：永久封禁将导致该用户无法再登录系统。</n-alert></n-form>
      <template #footer><div style="text-align: right;"><n-button @click="showBanModal = false" style="margin-right: 10px;">取消</n-button><n-button type="error" :loading="submitting" @click="submitBan">确认封禁</n-button></div></template>
    </n-modal>

    <n-modal v-model:show="showEditModal" preset="card" title="✏️ 修改用户资料 (上帝模式)" style="width: 500px">
        <div style="display: flex; justify-content: center; margin-bottom: 24px; position: relative;">
             <n-avatar :size="80" round :src="editForm.avatar ? `http://localhost:8080${editForm.avatar}` : ''" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" style="border: 2px solid #eee;" />
             
             <n-upload abstract :show-file-list="false" @before-upload="onSelectFile">
                <n-upload-trigger #="{ handleClick }" abstract>
                    <n-button circle type="primary" size="small" style="position: absolute; bottom: 0; margin-left: 50px; box-shadow: 0 2px 5px rgba(0,0,0,0.2)" @click="handleClick"><template #icon><n-icon><CloudUploadOutline /></n-icon></template></n-button>
                </n-upload-trigger>
             </n-upload>
        </div>

        <n-form label-placement="left" label-width="80">
            <n-form-item label="昵称"><n-input v-model:value="editForm.nickname" /></n-form-item>
            <n-form-item label="学校"><n-input v-model:value="editForm.school" /></n-form-item>
            <n-form-item label="专业">
                <n-space vertical style="width: 100%">
                    <n-select v-model:value="adminMajorSelect" :options="MAJOR_OPTIONS" placeholder="选择专业" />
                    <n-input v-if="adminMajorSelect === 'other'" v-model:value="adminMajorCustom" placeholder="请输入自定义专业" />
                </n-space>
            </n-form-item>
            <n-grid cols="2" x-gap="12">
                <n-gi>
                    <n-form-item label="年级">
                        <n-select v-model:value="editForm.grade" :options="GRADE_OPTIONS" placeholder="入学年份" />
                    </n-form-item>
                </n-gi>
            </n-grid>
            <n-form-item label="QQ"><n-input v-model:value="editForm.qq" /></n-form-item>
            <n-form-item label="微信"><n-input v-model:value="editForm.wechat" /></n-form-item>
            <n-form-item label="邮箱"><n-input v-model:value="editForm.email" /></n-form-item>
            <n-form-item label="性别"><n-radio-group v-model:value="editForm.gender"><n-space><n-radio :value="0">保密</n-radio><n-radio :value="1">男</n-radio><n-radio :value="2">女</n-radio></n-space></n-radio-group></n-form-item>
        </n-form>
        <template #footer><div style="display:flex; justify-content:flex-end"><n-button @click="showEditModal=false" style="margin-right:12px">取消</n-button><n-button type="primary" @click="handleSaveUser">保存修改</n-button></div></template>
    </n-modal>

    <n-modal v-model:show="showCropper" preset="card" title="修改头像 (裁剪)" style="width: 600px">
      <div style="width: 100%; height: 400px;">
        <vue-cropper
          ref="cropperRef"
          :img="cropperOptions.img"
          :output-size="1"
          :output-type="'png'"
          :info="true"
          :can-scale="true"
          :auto-crop="true"
          :auto-crop-width="200"
          :auto-crop-height="200"
          :fixed="true"
          :fixed-number="[1, 1]"
          :center-box="true"
        ></vue-cropper>
      </div>
      <template #footer>
        <n-space justify="end">
           <n-button @click="showCropper = false">取消</n-button>
           <n-button type="primary" @click="handleCropConfirm" :loading="uploadLoading">
             <template #icon><n-icon><CheckmarkOutline /></n-icon></template>
             确认并上传
           </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showResetModal" preset="card" title="🔒 强制重置密码" style="width: 400px">
        <n-card :bordered="false" size="small" style="background: #fff8f8; color: #d03050; margin-bottom: 12px;">⚠️ 警告：该操作将强制覆盖用户原有密码，用户需使用新密码登录。</n-card>
        <n-form><n-form-item label="输入新密码"><n-input type="password" show-password-on="click" v-model:value="resetForm.new_password" placeholder="建议设置为 user123 或 123456" /></n-form-item></n-form>
        <template #footer><div style="display:flex; justify-content:flex-end"><n-button @click="showResetModal=false" style="margin-right:12px">取消</n-button><n-button type="error" @click="handleConfirmReset">确认重置</n-button></div></template>
    </n-modal>
  </div>
</template>

<style scoped>
.user-manage-container { padding: 24px; min-height: 100vh; background-color: #f5f7fa; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>