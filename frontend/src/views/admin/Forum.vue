<script setup lang="ts">
import { ref, onMounted, reactive, h, shallowRef, onBeforeUnmount, computed } from 'vue'
import { 
  NCard, NDataTable, NButton, NInput, NSpace, NTag, NModal, NForm, NFormItem, 
  NSelect, NSwitch, NInputNumber, NPopconfirm, useMessage, NIcon, NAvatar, NTabs, NTabPane,
  NEllipsis, NSpin, NDivider, NUpload, NTooltip, NPageHeader
} from 'naive-ui'
import { 
  AddOutline, RefreshOutline, CreateOutline, TrashOutline, 
  LockClosedOutline, LockOpenOutline, MegaphoneOutline, 
  CheckmarkCircleOutline, EyeOutline, CloseCircleOutline, CropOutline
} from '@vicons/ionicons5'
import type { DataTableColumns, UploadFileInfo } from 'naive-ui'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

// 引入 WangEditor
import '@wangeditor/editor/dist/css/style.css' 
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
// 引入 VueCropper
import 'vue-cropper/dist/index.css' 
import { VueCropper } from 'vue-cropper'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)

// 环境变量或全局配置 (提取出来方便后续修改)
const BASE_URL = 'http://localhost:8080'

// --- 辅助函数 ---
const stripHtml = (html: string, length = 50) => {
  if (!html) return ''
  const tmp = document.createElement("DIV"); tmp.innerHTML = html
  let text = tmp.textContent || tmp.innerText || ""
  text = text.replace(/\s+/g, ' ').trim()
  return text.length > length ? text.slice(0, length) + '...' : text
}

const processUrl = (url: string) => {
    if (!url) return ''
    if (url.startsWith('http')) return url
    return `${BASE_URL}${url}`
}

// 🔥 修复富文本内的相对路径图片，使其在后台也能正常显示
const processContent = (html: string) => {
    if (!html) return '<span style="color:#999">内容为空</span>'
    return html.replace(/src="\/uploads\//g, `src="${BASE_URL}/uploads/`)
}

// =============================================================
// WangEditor 配置
// =============================================================
const editorRef = shallowRef()
const mode = 'default'
const toolbarConfig = { excludeKeys: ['group-video'] } 
const editorConfig = { 
  placeholder: '请输入正文...',
  MENU_CONF: {
    uploadImage: {
      server: '/api/v1/forum/upload',
      fieldName: 'file',
      maxFileSize: 5 * 1024 * 1024,
      headers: { Authorization: `Bearer ${userStore.token}` },
      customInsert(res: any, insertFn: any) {
        if (res.url) { insertFn(processUrl(res.url), '', '') } else { message.error('上传图片失败') }
      }
    }
  }
}
onBeforeUnmount(() => { const editor = editorRef.value; if (editor) editor.destroy() })
const handleCreated = (editor: any) => { editorRef.value = editor }

// =============================================================
// 🟢 板块管理 (Board)
// =============================================================
const boardList = ref<any[]>([])
const showBoardModal = ref(false)
const boardModel = ref({ id: 0, name: '', description: '', icon: '', sort_order: 0, is_locked: false })

// 剪裁器状态
const showCropperModal = ref(false)
const cropperRef = ref() 
const cropperOptions = reactive({ img: '', autoCropWidth: 200, autoCropHeight: 200 })
const cropperLoading = ref(false)

const fetchBoards = async () => {
  const res: any = await request.get('/forum/boards') 
  if (res.data) boardList.value = res.data
}

const handleEditBoard = (row: any) => { boardModel.value = { ...row }; showBoardModal.value = true }
const handleAddBoard = () => { boardModel.value = { id: 0, name: '', description: '', icon: '', sort_order: 0, is_locked: false }; showBoardModal.value = true }

const onSelectFile = async (data: { file: UploadFileInfo }) => {
  const file = data.file.file
  if (!file) return false
  const reader = new FileReader()
  reader.readAsDataURL(file)
  reader.onload = (e) => { cropperOptions.img = e.target?.result as string; showCropperModal.value = true }
  return false 
}

const uploadCroppedImage = () => {
  cropperLoading.value = true
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', blob, 'icon_cropped.png')
      const res: any = await request.post('/forum/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
      boardModel.value.icon = res.url
      message.success('图标设置成功')
      showCropperModal.value = false 
    } catch (e) {
      message.error('上传失败')
    } finally { cropperLoading.value = false }
  })
}

const submitBoard = async () => {
  try {
    if (boardModel.value.id) await request.put(`/admin/forum/boards/${boardModel.value.id}`, boardModel.value)
    else await request.post('/admin/forum/boards', boardModel.value)
    message.success('操作成功'); showBoardModal.value = false; fetchBoards()
  } catch (e) { message.error('操作失败') }
}

const deleteBoard = async (id: number) => {
  try { await request.delete(`/admin/forum/boards/${id}`); message.success('删除成功'); fetchBoards() } 
  catch (e: any) { message.error(e.response?.data?.error || '删除失败') }
}

const boardColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '图标', key: 'icon', width: 80, align: 'center',
    render: (row: any) => row.icon 
      ? h(NAvatar, { src: processUrl(row.icon), size: 42, style: 'background: transparent; border-radius: 8px; border: 1px solid #eee' }) 
      : h(NAvatar, { size: 42, color: '#f3f4f6', style: 'border-radius: 8px;' }, { default: () => h(NIcon, { color: '#9ca3af' }, { default: () => h(AddOutline) }) })
  },
  { title: '板块名称', key: 'name', minWidth: 150, render: (row: any) => h('span', { style: 'font-weight: 600; font-size: 15px;' }, row.name) },
  { title: '描述说明', key: 'description', minWidth: 200, render: (row: any) => h(NEllipsis, { style: 'color: #64748b' }, { default: () => row.description || '暂无描述' }) },
  // 🔥 新增：展示后端返回的帖子与评论统计
  { title: '主题数', key: 'post_count', width: 90, align: 'center', render: (row: any) => h(NTag, { type: 'info', bordered: false }, { default: () => row.post_count || 0 }) },
  { title: '评论数', key: 'comment_count', width: 90, align: 'center', render: (row: any) => h(NTag, { type: 'default', bordered: false }, { default: () => row.comment_count || 0 }) },
  { title: '排序权重', key: 'sort_order', width: 90, align: 'center' },
  { 
    title: '发帖权限', key: 'is_locked', width: 120, align: 'center',
    render: (row: any) => row.is_locked 
      ? h(NTag, { type: 'error', size: 'small', round: true }, { default: () => '🔒 仅超管', icon: () => h(NIcon, null, { default: () => h(LockClosedOutline) }) })
      : h(NTag, { type: 'success', size: 'small', round: true }, { default: () => '🔓 全员开放', icon: () => h(NIcon, null, { default: () => h(LockOpenOutline) }) })
  },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 140, align: 'center',
    render: (row: any) => h(NSpace, { justify: 'center', wrap: false }, { default: () => [
      h(NButton, { size: 'small', type: 'primary', secondary: true, onClick: () => handleEditBoard(row) }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }) }),
      h(NPopconfirm, { onPositiveClick: () => deleteBoard(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }) }), default: () => '警告：板块内必须无帖子才可删除，确定？' })
    ]})
  }
]

// =============================================================
// 🔵 帖子管理 (Post)
// =============================================================
const postList = ref<any[]>([])
const postPagination = reactive({ page: 1, pageSize: 12, itemCount: 0, onChange: (p: number) => { postPagination.page = p; fetchPosts() } })
const showPostModal = ref(false)
const postModel = ref({ title: '', board_id: null as number | null, summary: '', content: '<p></p>', is_pinned: false })
const postLoading = ref(false)
const boardOptions = computed(() => boardList.value.map(b => ({ label: b.name, value: b.id })))

const fetchPosts = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/forum/posts', { params: { page: postPagination.page, page_size: postPagination.pageSize } })
    if (res.data) { postList.value = res.data; postPagination.itemCount = res.total }
  } finally { loading.value = false }
}

const handleCreatePost = () => { postModel.value = { title: '', board_id: null, summary: '', content: '<p></p>', is_pinned: false }; showPostModal.value = true }

const submitPost = async () => {
  if (!postModel.value.title || !postModel.value.board_id) return message.warning('请补全发布信息')
  postLoading.value = true
  try {
    await request.post('/forum/posts', postModel.value)
    message.success('发布成功'); showPostModal.value = false; fetchPosts()
  } catch (e) { message.error('发布失败') } finally { postLoading.value = false }
}

const deletePost = async (id: number) => {
  try { await request.delete(`/admin/forum/posts/${id}`); message.success('已删除'); fetchPosts() } 
  catch (e) { message.error('删除失败') }
}

const postColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 80, align: 'center' },
  { 
    title: '标题与摘要', key: 'title', minWidth: 300, 
    render: (row: any) => h('div', { class: 'table-title-cell' }, [ 
      h('div', { style: 'display: flex; align-items: center; margin-bottom: 4px;' }, [
        row.is_pinned ? h(NTag, { type: 'warning', size: 'small', style: 'margin-right: 8px;' }, { default: () => '置顶' }) : null,
        h('span', { style: 'font-weight: 600; font-size: 15px; color: #1e293b' }, row.title)
      ]),
      h(NEllipsis, { style: 'color: #64748b; font-size: 13px;' }, { default: () => row.summary || stripHtml(row.content, 50) })
    ]) 
  },
  { title: '所属板块', key: 'board', width: 140, render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => row.board?.name || '未知板块' }) },
  { title: '作者', key: 'author', width: 160, render: (row: any) => h('div', { class: 'author-cell' }, [ h(NAvatar, { round: true, size: 28, src: row.author?.avatar ? processUrl(row.author.avatar) : undefined }), h('span', row.author?.nickname || row.author?.username) ]) },
  { title: '发布时间', key: 'created_at', width: 180, render: (row: any) => new Date(row.created_at).toLocaleString() },
  { title: '操作', key: 'actions', fixed: 'right', width: 100, align: 'center', render: (row: any) => h(NPopconfirm, { onPositiveClick: () => deletePost(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { default: () => '删除' }), default: () => '确定要删除该帖子及其所有评论吗？' }) }
]

// =============================================================
// 🟡 评论管理 (Comment)
// =============================================================
const commentList = ref<any[]>([])
const commentPagination = reactive({ page: 1, pageSize: 12, itemCount: 0, onChange: (p: number) => { commentPagination.page = p; fetchComments() } })
const commentSearch = ref('')

const fetchComments = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/forum/comments', { params: { page: commentPagination.page, page_size: commentPagination.pageSize, q: commentSearch.value } })
    if (res.data) { commentList.value = res.data; commentPagination.itemCount = res.total }
  } finally { loading.value = false }
}

const deleteComment = async (id: number) => {
  try { await request.delete(`/admin/forum/comments/${id}`); message.success('评论已删除'); fetchComments() } 
  catch (e) { message.error('删除失败') }
}

const commentColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 80, align: 'center' },
  { 
    title: '评论内容', key: 'content', minWidth: 350, 
    render: (row: any) => h('div', { class: 'table-title-cell' }, [
      h(NEllipsis, { lineClamp: 2, style: 'font-size: 14px; color: #334155; line-height: 1.6;' }, { default: () => stripHtml(row.content, 150) }),
      row.parent_id ? h('div', { style: 'font-size: 12px; color: #94a3b8; margin-top: 4px;' }, `↳ 回复了评论 ID: #${row.parent_id}`) : null
    ])
  },
  { title: '关联帖子', key: 'post_id', width: 120, align: 'center', render: (row: any) => h(NTag, { size: 'small', bordered: false }, { default: () => `帖子 #${row.post_id}` }) },
  { title: '发布者', key: 'author', width: 160, render: (row: any) => h('div', { class: 'author-cell' }, [ h(NAvatar, { round: true, size: 28, src: row.author?.avatar ? processUrl(row.author.avatar) : undefined }), h('span', row.author?.nickname || row.author?.username) ]) },
  { title: '时间', key: 'created_at', width: 180, render: (row: any) => new Date(row.created_at).toLocaleString() },
  { title: '操作', key: 'actions', fixed: 'right', width: 100, align: 'center', render: (row: any) => h(NPopconfirm, { onPositiveClick: () => deleteComment(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { default: () => '删除' }), default: () => '直接删除此评论？' }) }
]

// =============================================================
// 🔴 举报中心 (Report)
// =============================================================
const reportList = ref<any[]>([])
const reportPagination = reactive({ page: 1, pageSize: 12, itemCount: 0, onChange: (p: number) => { reportPagination.page = p; fetchReports() } })
const showPreview = ref(false)
const previewContent = ref<any>({ title: '', content: '', author: '', type: '' })
const previewLoading = ref(false)

const fetchReports = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/forum/reports', { params: { page: reportPagination.page, page_size: reportPagination.pageSize } })
    if (res.data) { reportList.value = res.data; reportPagination.itemCount = res.total }
  } finally { loading.value = false }
}

const resolveReport = async (id: number) => { 
  try { await request.put(`/admin/forum/reports/${id}/resolve`); message.success('已标记为处理完成'); fetchReports() } 
  catch (e) { message.error('标记失败') } 
}

const deleteTargetAndResolve = async (row: any) => {
  try {
    if (row.target_type === 'post') await request.delete(`/admin/forum/posts/${row.target_id}`)
    else await request.delete(`/admin/forum/comments/${row.target_id}`)
    
    await request.put(`/admin/forum/reports/${row.id}/resolve`)
    message.success('已强制删除违规内容'); fetchReports()
  } catch (e) { message.error('处理失败') }
}

const openPreview = async (row: any) => {
  showPreview.value = true
  previewLoading.value = true
  try {
    const res: any = await request.get('/admin/forum/reports/preview', { params: { target_type: row.target_type, target_id: row.target_id } })
    previewContent.value = res
  } catch (e) { 
    previewContent.value = { title: '内容已被删除或不存在', content: '<p style="color:red;">无法获取内容</p>', author: '系统', type: 'error' }
  } finally { 
    previewLoading.value = false 
  }
}

const reportColumns: DataTableColumns = [
  { title: '工单号', key: 'id', width: 80, align: 'center' },
  { 
    title: '处理状态', key: 'status', width: 100, align: 'center', 
    render: (row: any) => row.status === 0 
      ? h(NTag, { type: 'error', size: 'small', round: true }, { default: () => '🔴 待处理' }) 
      : h(NTag, { type: 'default', size: 'small', round: true }, { default: () => '⚪ 已归档' }) 
  },
  { 
    title: '违规类型', key: 'target_type', width: 100, align: 'center', 
    render: (row: any) => row.target_type === 'post' 
      ? h(NTag, { type: 'info', size: 'small' }, { default: () => '主贴 (Post)' }) 
      : h(NTag, { type: 'warning', size: 'small' }, { default: () => '评论 (Comment)' }) 
  },
  { 
    title: '举报理由', key: 'reason', minWidth: 200, 
    render: (row: any) => h('span', { style: 'color: #be123c; font-weight: bold; background: #fff1f2; padding: 4px 8px; border-radius: 4px;' }, row.reason) 
  },
  { title: '举报人', key: 'reporter', width: 160, render: (row: any) => h('div', { class: 'author-cell' }, [ h(NAvatar, { round: true, size: 24, src: row.reporter?.avatar ? processUrl(row.reporter.avatar) : undefined }), h('span', row.reporter?.nickname || row.reporter?.username) ]) },
  { title: '举报时间', key: 'created_at', width: 180, render: (row: any) => new Date(row.created_at).toLocaleString() },
  { 
    title: '操作面板', key: 'actions', fixed: 'right', width: 200, align: 'center', 
    render: (row: any) => row.status === 0 ? h(NSpace, { justify: 'center', wrap: false }, [ 
      h(NTooltip, null, { trigger: () => h(NButton, { size: 'small', type: 'primary', secondary: true, onClick: () => openPreview(row) }, { icon: () => h(NIcon, null, { default: () => h(EyeOutline) }), default: () => '审查内容' }), default: () => '还原并预览真实图文内容' }), 
      h(NPopconfirm, { onPositiveClick: () => resolveReport(row.id) }, { trigger: () => h(NButton, { size: 'small', secondary: true }, { default: () => '忽略' }), default: () => '确认为无效举报？' }), 
      h(NPopconfirm, { onPositiveClick: () => deleteTargetAndResolve(row) }, { trigger: () => h(NButton, { size: 'small', type: 'error' }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }) }), default: () => '警告：将直接删除该违规贴/评论！' }) 
    ]) : h('span', { style: 'color:#94a3b8; font-size: 13px;' }, '无需操作') 
  }
]

onMounted(() => { fetchBoards(); fetchPosts(); fetchComments(); fetchReports() })
</script>

<template>
  <div class="page-container">
    <n-page-header title="💬 社区与论坛管理" subtitle="维护社区秩序，配置交流板块与审核违规内容" style="margin-bottom: 24px;" />

    <n-card :bordered="false" class="main-card">
      <n-tabs type="line" size="large" justify-content="start" :tabs-padding="24" pane-style="padding: 24px;">
        
        <n-tab-pane name="board" tab="板块架构配置">
           <div class="tab-toolbar">
             <div class="toolbar-info">
               <h3>📁 社区板块列表</h3>
               <p>控制前端论坛的分类展示、访问权限及视觉图标。</p>
             </div>
             <n-button type="primary" size="large" @click="handleAddBoard">
               <template #icon><n-icon><AddOutline /></n-icon></template>创建新板块
             </n-button>
           </div>
           <n-data-table :columns="boardColumns" :data="boardList" :loading="loading" size="large" />
        </n-tab-pane>

        <n-tab-pane name="post" tab="全站帖子大厅">
            <div class="tab-toolbar">
             <div class="toolbar-info">
               <h3>📝 帖子动态大厅</h3>
               <p>审查用户发帖，或以官方身份发布全局公告与活动。</p>
             </div>
             <n-space>
               <n-button size="large" @click="fetchPosts"><template #icon><n-icon><RefreshOutline /></n-icon></template>刷新视图</n-button>
               <n-button size="large" type="primary" color="#8a2be2" @click="handleCreatePost">
                 <template #icon><n-icon><MegaphoneOutline /></n-icon></template>发布官方置顶帖
               </n-button>
             </n-space>
           </div>
           <n-data-table :columns="postColumns" :data="postList" :loading="loading" :pagination="postPagination" remote size="large" />
        </n-tab-pane>

        <n-tab-pane name="comment" tab="互动评论流">
            <div class="tab-toolbar">
             <div class="toolbar-info">
               <h3>🗣️ 实时评论流监控</h3>
               <p>全文检索并管理分布在各帖子下的用户回复。</p>
             </div>
             <n-space>
               <n-input v-model:value="commentSearch" placeholder="输入关键词检索违规词..." @keydown.enter="fetchComments" size="large" style="width: 300px;" clearable />
               <n-button size="large" type="primary" secondary @click="fetchComments">检索</n-button>
             </n-space>
           </div>
           <n-data-table :columns="commentColumns" :data="commentList" :loading="loading" :pagination="commentPagination" remote size="large" />
        </n-tab-pane>

        <n-tab-pane name="report" tab="风控与举报中心">
            <div class="tab-toolbar">
             <div class="toolbar-info">
               <h3>🚨 违规风控收件箱</h3>
               <p>处理来自用户社区巡查的举报反馈，支持图文原貌溯源。</p>
             </div>
             <n-button size="large" @click="fetchReports"><template #icon><n-icon><RefreshOutline /></n-icon></template>刷新工单</n-button>
           </div>
           <n-data-table :columns="reportColumns" :data="reportList" :loading="loading" :pagination="reportPagination" remote size="large" />
        </n-tab-pane>

      </n-tabs>
    </n-card>

    <n-modal v-model:show="showBoardModal" preset="card" title="板块属性设置" style="width: 550px">
      <n-form label-placement="left" label-width="90px" size="large">
        <n-form-item label="板块视觉">
           <div style="display: flex; align-items: center; gap: 20px;">
              <n-upload :show-file-list="false" accept="image/*" @before-upload="onSelectFile">
                 <n-button secondary type="info">上传新图标</n-button>
              </n-upload>
              <div v-if="boardModel.icon" class="icon-preview-box">
                 <n-avatar :src="processUrl(boardModel.icon)" :size="64" style="background:#f8fafc; border-radius:12px; border:1px solid #e2e8f0;" />
                 <n-icon class="icon-delete" @click="boardModel.icon = ''"><CloseCircleOutline /></n-icon>
              </div>
              <span v-else style="color: #94a3b8; font-size: 13px;">(支持自动等比剪裁)</span>
           </div>
        </n-form-item>
        <n-form-item label="显示名称"><n-input v-model:value="boardModel.name" placeholder="例如：考研交流区" /></n-form-item>
        <n-form-item label="详细说明"><n-input v-model:value="boardModel.description" type="textarea" placeholder="描述该板块的主题范围..." :rows="3" /></n-form-item>
        <n-form-item label="展示权重"><n-input-number v-model:value="boardModel.sort_order" style="width:100%" placeholder="数字越大越靠前" /></n-form-item>
        <n-form-item label="权限控制">
          <n-switch v-model:value="boardModel.is_locked" size="large">
            <template #checked>已锁定 (仅管理可发帖)</template>
            <template #unchecked>开放 (全员可发帖)</template>
          </n-switch>
        </n-form-item>
      </n-form>
      <template #footer><div style="text-align:right"><n-button size="large" type="primary" @click="submitBoard">保存变更</n-button></div></template>
    </n-modal>

    <n-modal v-model:show="showCropperModal" preset="card" title="框选版块图标" style="width: 600px;">
        <div style="height: 400px; width: 100%;">
            <VueCropper
                ref="cropperRef" :img="cropperOptions.img" :outputSize="1" :outputType="'png'"
                :info="true" :canScale="true" :autoCrop="true" :autoCropWidth="200" :autoCropHeight="200"
                :fixed="true" :fixedNumber="[1, 1]" :centerBox="true"
            />
        </div>
        <template #footer>
            <div style="display:flex; justify-content:flex-end; gap:12px;">
                <n-button @click="showCropperModal = false">取消</n-button>
                <n-button type="primary" :loading="cropperLoading" @click="uploadCroppedImage">裁剪并保存</n-button>
            </div>
        </template>
    </n-modal>

    <n-modal v-model:show="showPostModal" preset="card" title="📝 撰写全站公告" style="width: 1000px;">
      <n-form size="large">
        <n-form-item label="公告标题"><n-input v-model:value="postModel.title" placeholder="输入醒目的标题..." /></n-form-item>
        <n-form-item label="归属板块">
          <n-select v-model:value="postModel.board_id" :options="boardOptions" placeholder="选择公告要发布到的板块" />
        </n-form-item>
        <n-form-item label="置顶属性">
          <n-switch v-model:value="postModel.is_pinned"><template #checked>将此帖在板块内全局置顶</template></n-switch>
        </n-form-item>
        <div style="border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; margin-top: 10px;">
            <Toolbar style="border-bottom: 1px solid #e2e8f0; background: #f8fafc;" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
            <Editor style="height: 500px; overflow-y: hidden;" v-model="postModel.content" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
        </div>
      </n-form>
      <template #footer><div style="text-align:right"><n-button size="large" type="primary" :loading="postLoading" @click="submitPost" style="width: 150px">立即发布</n-button></div></template>
    </n-modal>

    <n-modal v-model:show="showPreview" preset="card" title="🔍 违规内容实景审查" style="width: 800px;">
        <n-spin :show="previewLoading">
            <div v-if="previewContent.type !== 'error'" class="preview-sandbox">
                
                <div class="preview-header">
                  <div class="preview-type-badge">
                    <n-tag :type="previewContent.type === 'post' ? 'info' : 'warning'" size="large">
                      {{ previewContent.type === 'post' ? '主贴' : '评论' }}
                    </n-tag>
                  </div>
                  <div class="preview-meta">
                    <h2 class="preview-title">{{ previewContent.title }}</h2>
                    <div class="preview-author">发布者：{{ previewContent.author }}</div>
                  </div>
                </div>

                <n-divider style="margin: 16px 0;" />
                
                <div class="editor-content-view" v-html="processContent(previewContent.content)"></div>
            </div>
            <div v-else class="preview-error">
               <n-icon size="60" color="#fca5a5"><CloseCircleOutline/></n-icon>
               <p>内容已被删除或不存在，无法溯源</p>
            </div>
        </n-spin>
    </n-modal>
  </div>
</template>

<style scoped>
.page-container { padding: 24px; background: #f1f5f9; min-height: 100vh; }
.main-card { border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05); }

/* Tab 头部栏设计 */
.tab-toolbar { 
  display: flex; justify-content: space-between; align-items: flex-end; 
  margin-bottom: 24px; padding-bottom: 20px; border-bottom: 1px solid #f1f5f9;
}
.toolbar-info h3 { margin: 0 0 8px 0; font-size: 20px; color: #1e293b; }
.toolbar-info p { margin: 0; color: #64748b; font-size: 14px; }

/* 表格内元素优化 */
.table-title-cell { display: flex; flex-direction: column; gap: 4px; padding: 4px 0; }
.author-cell { display: flex; align-items: center; gap: 10px; font-weight: 500; color: #334155; }

/* 弹窗小组件 */
.icon-preview-box { position: relative; display: inline-block; }
.icon-delete { position: absolute; top: -8px; right: -8px; cursor: pointer; color: #ef4444; font-size: 20px; background: #fff; border-radius: 50%; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }

/* 🔥 富文本预览沙箱样式 (修复图片与排版错乱的核心) */
.preview-sandbox { padding: 10px; }
.preview-header { display: flex; align-items: flex-start; gap: 16px; }
.preview-type-badge { padding-top: 4px; }
.preview-title { margin: 0 0 6px 0; font-size: 22px; color: #0f172a; line-height: 1.4; }
.preview-author { font-size: 14px; color: #64748b; }
.preview-error { text-align: center; padding: 40px; color: #ef4444; font-size: 16px; }

/* 🛡️ 专门针对后端 WangEditor 导出的 HTML 标签进行重置与限制 */
.editor-content-view {
  font-size: 16px;
  line-height: 1.8;
  color: #334155;
  max-height: 60vh;
  overflow-y: auto;
  padding: 10px;
  background: #f8fafc;
  border-radius: 8px;
}
.editor-content-view :deep(img) {
  max-width: 100%;
  height: auto !important; /* 防止内联 height 导致拉伸 */
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  margin: 12px 0;
  display: block;
}
.editor-content-view :deep(p) { margin: 0 0 12px 0; }
.editor-content-view :deep(blockquote) {
  border-left: 4px solid #cbd5e1;
  padding-left: 12px;
  color: #64748b;
  margin: 12px 0;
  background: #f1f5f9;
  padding: 10px 12px;
}
.editor-content-view :deep(ul), .editor-content-view :deep(ol) {
  padding-left: 20px;
  margin-bottom: 12px;
}
.editor-content-view :deep(a) { color: #2563eb; text-decoration: none; }
.editor-content-view :deep(a:hover) { text-decoration: underline; }
</style>