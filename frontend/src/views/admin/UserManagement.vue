<script setup lang="ts">
import { ref, onMounted, reactive, h, computed, watch } from 'vue'
import { 
  NCard, NDataTable, NTag, NButton, NSpace, NInput, NModal, NSelect, 
  NForm, NFormItem, useMessage, NPopconfirm, NIcon, NPageHeader,
  NGrid, NGi, NRadio, NRadioGroup, NAvatar, NUpload, NUploadTrigger, NSpin,
  NTooltip, NInputNumber, NStatistic
} from 'naive-ui'
import { 
  PersonOutline, ShieldCheckmarkOutline, BanOutline, SearchOutline, 
  CreateOutline, LockOpenOutline, KeyOutline, CloudUploadOutline, CheckmarkOutline,
  CopyOutline, TicketOutline, WalletOutline
} from '@vicons/ionicons5'
import 'vue-cropper/dist/index.css' 
import { VueCropper } from 'vue-cropper' 
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const list = ref([])
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const keyword = ref('')

const MAJOR_OPTIONS = [ { label: '临床医学', value: '临床医学' }, { label: '医学影像学', value: '医学影像学' }, { label: '麻醉学', value: '麻醉学' }, { label: '口腔医学', value: '口腔医学' }, { label: '基础医学', value: '基础医学' }, { label: '预防医学', value: '预防医学' }, { label: '护理学', value: '护理学' }, { label: '药学', value: '药学' }, { label: '中医学', value: '中医学' }, { label: '其他 (自定义)', value: 'other' } ]
const GRADE_OPTIONS = computed(() => { const currentYear = new Date().getFullYear() + 1; const list = []; for (let i = 0; i < 12; i++) { const y = currentYear - i; list.push({ label: `${y}级`, value: `${y}级` }) }; return list })
const roleOptions = [ { label: '普通用户', value: 'user' }, { label: '机构代理', value: 'agent' }, { label: '超级管理员', value: 'admin' } ]
const roleMap: Record<string, string> = { user: '普通用户', agent: '机构代理', admin: '超级管理员' }
const banDurationOptions = [ { label: '1 天', value: 24 }, { label: '3 天', value: 72 }, { label: '1 周', value: 168 }, { label: '1 个月', value: 720 }, { label: '永久封禁', value: -1 } ]

// === 模态框状态 ===
const showRoleModal = ref(false)
const showBanModal = ref(false)
const showEditModal = ref(false)
const showResetModal = ref(false)

// 🔥🔥🔥 优化：积分操作状态 🔥🔥🔥
const showPointsModal = ref(false)
const pointsForm = reactive({ user_id: 0, username: '', current_points: 0, points: 0 })
const pointsLoading = ref(false)
// 实时计算最终积分
const calculatedFinalPoints = computed(() => pointsForm.current_points + (pointsForm.points || 0))

const currentEditUser = ref<any>(null)
const roleForm = ref({ role: '' })
const banForm = ref({ duration: 24 })
const editForm = reactive({ id: 0, nickname: '', school: '', major: '', grade: null as string|null, qq: '', wechat: '', email: '', gender: 0, avatar: '' })
const adminMajorSelect = ref<string|null>(null)
const adminMajorCustom = ref('')
const resetForm = reactive({ id: 0, new_password: '' })
const submitting = ref(false)

// 头像剪裁相关状态
const showCropper = ref(false)
const cropperRef = ref()
const uploadLoading = ref(false)
const cropperOptions = reactive({ img: '', autoCrop: true, autoCropWidth: 200, autoCropHeight: 200, fixedBox: false, fixed: true, fixedNumber: [1, 1], centerBox: true, infoTrue: true })

watch([adminMajorSelect, adminMajorCustom], () => {
    if (adminMajorSelect.value === 'other') editForm.major = adminMajorCustom.value
    else editForm.major = adminMajorSelect.value || ''
})

const copyCode = (code: string) => {
    navigator.clipboard.writeText(code).then(() => { message.success('邀请码已复制') }).catch(() => { message.error('复制失败') })
}

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
      const type = row.role === 'admin' ? 'error' : (row.role === 'agent' ? 'info' : 'default')
      const label = roleMap[row.role] || row.role
      return h(NTag, { type, bordered: false, size: 'small' }, { default: () => label })
    }
  },
  // 🔥🔥🔥 优化：新增积分展示列 🔥🔥🔥
  {
    title: '当前积分', key: 'points', width: 100,
    render(row: any) {
        return h('div', { style: 'color: #d97706; font-weight: 800; font-family: monospace; font-size: 15px;' }, row.points || 0)
    }
  },
  {
    title: '邀请码', key: 'invitation_code', width: 130,
    render(row: any) {
        if (row.role !== 'agent' || !row.invitation_code) return '-'
        return h(NTag, { type: 'warning', size: 'small', style: 'cursor: pointer', onClick: () => copyCode(row.invitation_code) }, { 
            default: () => [ h(NIcon, { style: 'margin-right: 4px; vertical-align: text-bottom' }, { default: () => h(TicketOutline) }), row.invitation_code ]
        })
    }
  },
  { title: '🏫 学校', key: 'school', width: 140, ellipsis: { tooltip: true }, render: (row: any) => row.school || '-' },
  { title: '📚 专业', key: 'major', width: 120, ellipsis: { tooltip: true }, render: (row: any) => row.major || '-' },
  { 
    title: '🎓 年级', key: 'grade', width: 90,
    render(row: any) {
        if (!row.grade) return '-'
        return h(NTag, { size: 'small', bordered: false, type: 'default', style: 'opacity: 0.8' }, { default: () => row.grade })
    }
  },
  { 
    title: '状态', key: 'status', width: 80,
    render(row: any) {
      if (row.status === 2) return h(NTag, { type: 'error', size: 'small' }, { default: () => '封禁' })
      return h(NTag, { type: 'success', bordered: false, size: 'small' }, { default: () => '正常' })
    }
  },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 250,
    render(row: any) {
      return h(NSpace, { justify: 'center', size: 'small', wrap: false }, {
        default: () => [
          // 资料修改
          h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => openEditModal(row) }, 
            { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => '资料' }),
          // 🔥 优化：增扣积分按钮
          h(NTooltip, { trigger: 'hover' }, {
              trigger: () => h(NButton, { size: 'tiny', type: 'success', dashed: true, onClick: () => openPointsModal(row) }, 
                 { icon: () => h(NIcon, null, { default: () => h(WalletOutline) }) }),
              default: () => '增扣积分'
          }),
          // 角色
          h(NButton, { size: 'tiny', onClick: () => openRoleModal(row) }, { default: () => '角色' }),
          // 密码
          h(NTooltip, { trigger: 'hover' }, {
              trigger: () => h(NButton, { size: 'tiny', type: 'warning', dashed: true, onClick: () => openResetPwdModal(row) }, 
                 { icon: () => h(NIcon, null, { default: () => h(KeyOutline) }) }),
              default: () => '重置密码'
          }),
          // 封号
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

// 🔥🔥🔥 优化：积分操作逻辑 🔥🔥🔥
const openPointsModal = (row: any) => {
    pointsForm.user_id = row.id
    pointsForm.username = row.nickname || row.username
    pointsForm.current_points = row.points || 0 // 获取当前积分
    pointsForm.points = 0 // 每次打开输入框清零
    showPointsModal.value = true
}

const submitPoints = async () => {
    if (pointsForm.points === 0) return message.warning('变更额度不能为0')
    if (calculatedFinalPoints.value < 0) return message.error('扣除后积分不能小于0')
    
    pointsLoading.value = true
    try {
        await request.post('/admin/users/points', {
            user_id: pointsForm.user_id,
            points: pointsForm.points
        })
        message.success('积分操作成功')
        showPointsModal.value = false
        fetchData()
    } catch (e: any) {
        message.error(e.response?.data?.error || '操作失败')
    } finally {
        pointsLoading.value = false
    }
}

// 其余常规逻辑...
const openRoleModal = (user: any) => { currentEditUser.value = user; roleForm.value.role = user.role; showRoleModal.value = true }
const submitRole = async () => { submitting.value = true; try { await request.post('/admin/users/role', { user_id: currentEditUser.value.id, new_role: roleForm.value.role }); message.success('角色修改成功'); showRoleModal.value = false; fetchData() } catch { message.error('操作失败') } finally { submitting.value = false } }
const openBanModal = (user: any) => { currentEditUser.value = user; banForm.value.duration = 24; showBanModal.value = true }
const submitBan = async () => { submitting.value = true; try { await request.post('/admin/users/ban', { user_id: currentEditUser.value.id, duration: banForm.value.duration }); message.success('用户已封禁'); showBanModal.value = false; fetchData() } catch { message.error('操作失败') } finally { submitting.value = false } }
const handleUnban = async (id: number) => { try { await request.post('/admin/users/unban', { user_id: id }); message.success('已解封'); fetchData() } catch { message.error('操作失败') } }

const openEditModal = (row: any) => {
    editForm.id = row.id; editForm.nickname = row.nickname; editForm.school = row.school; editForm.major = row.major; editForm.grade = row.grade;
    editForm.qq = row.qq; editForm.wechat = row.wechat; editForm.email = row.email; editForm.gender = row.gender; editForm.avatar = row.avatar;
    if (editForm.major) {
        const exists = MAJOR_OPTIONS.some(opt => opt.value === editForm.major)
        if (exists) { adminMajorSelect.value = editForm.major } 
        else { adminMajorSelect.value = 'other'; adminMajorCustom.value = editForm.major }
    } else { adminMajorSelect.value = null; adminMajorCustom.value = '' }
    showEditModal.value = true
}
const handleSaveUser = async () => { try { await request.put(`/admin/users/${editForm.id}`, editForm); message.success('更新成功'); showEditModal.value = false; fetchData() } catch { message.error('更新失败') } }

const onSelectFile = async ({ file }: any) => {
  const reader = new FileReader(); reader.readAsDataURL(file.file); reader.onload = (e: any) => { cropperOptions.img = e.target.result; showCropper.value = true }; return false 
}
const handleCropConfirm = () => {
  uploadLoading.value = true
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData(); formData.append('file', blob, 'avatar.png') 
      const res: any = await request.post(`/admin/users/${editForm.id}/avatar`, formData, { headers: { 'Content-Type': 'multipart/form-data' } })
      if (res.url) { editForm.avatar = res.url + '?t=' + new Date().getTime(); message.success('头像修改成功'); showCropper.value = false; fetchData() }
    } catch (e) { message.error('上传失败') } finally { uploadLoading.value = false }
  })
}

const openResetPwdModal = (row: any) => { resetForm.id = row.id; resetForm.new_password = ''; showResetModal.value = true }
const handleConfirmReset = async () => {
    if(resetForm.new_password.length < 6) return message.warning('密码至少6位');
    try { await request.put(`/admin/users/${resetForm.id}/password`, { new_password: resetForm.new_password }); message.success('重置成功'); showResetModal.value = false } catch { message.error('重置失败') }
}

const handlePageChange = (page: number) => { pagination.page = page; fetchData() }
onMounted(fetchData)
</script>

<template>
  <div class="user-manage-container">
    <n-page-header title="👥 用户管理" subtitle="系统层级：资料修改、角色分配、积分与封号" style="margin-bottom: 24px;" />
    <n-card>
      <div class="toolbar">
        <n-input v-model:value="keyword" placeholder="搜索用户名/昵称..." style="width: 240px" @keydown.enter="handleSearch"><template #prefix><n-icon><SearchOutline /></n-icon></template></n-input>
        <n-button type="primary" @click="handleSearch">搜索</n-button>
      </div>
      <n-data-table remote :columns="columns" :data="list" :loading="loading" :pagination="pagination" @update:page="handlePageChange" style="margin-top: 16px;" :scroll-x="1300" />
    </n-card>

    <n-modal v-model:show="showPointsModal" preset="card" title="💰 积分发放与扣除" style="width: 450px">
        <div class="points-header-box">
            <div class="ph-label">正在操作目标用户：<strong style="color:#18a058">{{ pointsForm.username }}</strong></div>
            <n-statistic label="该用户当前积分" :value="pointsForm.current_points">
               <template #prefix><n-icon color="#f59e0b"><WalletOutline /></n-icon></template>
            </n-statistic>
        </div>

        <n-form>
            <n-form-item label="操作额度 (正数增加，负数扣除)">
                <n-input-number v-model:value="pointsForm.points" :step="100" style="width: 100%" size="large" clearable>
                    <template #suffix>分</template>
                </n-input-number>
            </n-form-item>
            
            <div class="preview-box" :class="{ 'is-danger': calculatedFinalPoints < 0 }">
                <span>操作后最终余额：</span>
                <span class="preview-val">{{ calculatedFinalPoints }}</span>
                <span v-if="calculatedFinalPoints < 0" style="color: #ef4444; font-size: 12px; margin-left: 8px;">(余额不可为负数)</span>
            </div>

            <p style="font-size: 12px; color: #999; margin-top: 16px; line-height: 1.5;">
                说明：在此处手动增加积分，系统不会计算代理提成。常用于私下微信/支付宝转账后的手动补发，或是活动赠送。
            </p>
        </n-form>
        <template #footer>
            <div style="display:flex; justify-content:flex-end">
                <n-button @click="showPointsModal=false" style="margin-right:12px">取消</n-button>
                <n-button type="primary" :loading="pointsLoading" @click="submitPoints" :disabled="calculatedFinalPoints < 0">确认执行</n-button>
            </div>
        </template>
    </n-modal>

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
                <n-gi><n-form-item label="年级"><n-select v-model:value="editForm.grade" :options="GRADE_OPTIONS" placeholder="入学年份" /></n-form-item></n-gi>
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
        <vue-cropper ref="cropperRef" :img="cropperOptions.img" :output-size="1" :output-type="'png'" :info="true" :can-scale="true" :auto-crop="true" :auto-crop-width="200" :auto-crop-height="200" :fixed="true" :fixed-number="[1, 1]" :center-box="true"></vue-cropper>
      </div>
      <template #footer>
        <n-space justify="end"><n-button @click="showCropper = false">取消</n-button><n-button type="primary" @click="handleCropConfirm" :loading="uploadLoading"><template #icon><n-icon><CheckmarkOutline /></n-icon></template>确认并上传</n-button></n-space>
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

/* 积分弹窗特殊样式 */
.points-header-box { background: #fefce8; border: 1px solid #fef08a; padding: 16px; border-radius: 8px; margin-bottom: 24px; text-align: center; }
.ph-label { font-size: 14px; color: #854d0e; margin-bottom: 8px; }
.preview-box { background: #f8fafc; border: 1px solid #e2e8f0; padding: 12px 16px; border-radius: 8px; display: flex; align-items: baseline; font-size: 14px; color: #475569; transition: all 0.3s; }
.preview-box.is-danger { background: #fef2f2; border-color: #fecaca; color: #ef4444; }
.preview-val { font-size: 24px; font-weight: bold; margin-left: 8px; font-family: monospace; color: #0f172a; }
.preview-box.is-danger .preview-val { color: #ef4444; }
</style>