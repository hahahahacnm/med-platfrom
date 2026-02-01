<script setup lang="ts">
import { ref, onMounted, reactive, h, shallowRef, onBeforeUnmount, computed } from 'vue'
import { 
  NCard, NDataTable, NButton, NInput, NSpace, NTag, NModal, NForm, NFormItem, 
  NSelect, NSwitch, NInputNumber, NPopconfirm, useMessage, NIcon, NAvatar, NTabs, NTabPane,
  NBadge, NTooltip, NEllipsis, NSpin, NDivider, NUpload
} from 'naive-ui'
import { 
  AddOutline, RefreshOutline, CreateOutline, TrashOutline, 
  DocumentTextOutline, LockClosedOutline, LockOpenOutline,
  MegaphoneOutline, ChatboxOutline, AlertCircleOutline, CheckmarkCircleOutline,
  EyeOutline, CloseCircleOutline, CloudUploadOutline, CropOutline
} from '@vicons/ionicons5'
import type { DataTableColumns, UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

// 🔥 1. 引入 WangEditor
import '@wangeditor/editor/dist/css/style.css' 
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

// 🔥 2. 引入 VueCropper (交互式剪裁)
// ✅ 核心修复：Vue 3 版本 (vue-cropper@next) 的样式文件名为 index.css
import 'vue-cropper/dist/index.css' 
import { VueCropper } from 'vue-cropper'

const message = useMessage()
const loading = ref(false)

// 通用辅助
const stripHtml = (html: string, length = 30) => {
  if (!html) return ''
  const tmp = document.createElement("DIV"); tmp.innerHTML = html
  let text = tmp.textContent || tmp.innerText || ""
  text = text.replace(/\s+/g, ' ').trim()
  return text.length > length ? text.slice(0, length) + '...' : text
}

const processUrl = (url: string) => {
    if (!url) return ''
    if (url.startsWith('http')) return url
    return `http://localhost:8080${url}`
}

const processContent = (html: string) => {
    if (!html) return ''
    return html.replace(/src="\/uploads\//g, 'src="http://localhost:8080/uploads/')
}

// -------------------------------------------------------------
// WangEditor 配置
// -------------------------------------------------------------
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
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      customInsert(res: any, insertFn: any) {
        if (res.url) { insertFn(processUrl(res.url), '', '') } else { message.error('上传失败') }
      }
    }
  }
}
onBeforeUnmount(() => { const editor = editorRef.value; if (editor) editor.destroy() })
const handleCreated = (editor: any) => { editorRef.value = editor }

// =============================================================
// 🟢 Tab 1: 板块管理 (交互式剪裁版)
// =============================================================
const boardList = ref<any[]>([])
const showBoardModal = ref(false)
const boardModel = ref({ id: 0, name: '', description: '', icon: '', sort_order: 0, is_locked: false })

// 🔥 剪裁器相关状态
const showCropperModal = ref(false)
const cropperRef = ref() // 剪裁器实例
const cropperImg = ref('') // 待剪裁图片 Base64
const cropperLoading = ref(false)
const cropperOptions = reactive({
  img: '', 
  autoCrop: true, 
  autoCropWidth: 200, 
  autoCropHeight: 200, 
  fixed: true, // 固定宽高比
  fixedNumber: [1, 1], // 1:1 正方形
  centerBox: true,
  infoTrue: true // 展示真实输出大小
})

const fetchBoards = async () => {
  const res: any = await request.get('/forum/boards') 
  if (res.data) boardList.value = res.data
}
const handleEditBoard = (row: any) => { boardModel.value = { ...row }; showBoardModal.value = true }
const handleAddBoard = () => { boardModel.value = { id: 0, name: '', description: '', icon: '', sort_order: 0, is_locked: false }; showBoardModal.value = true }

// 🔥 1. 选择图片：拦截默认上传，打开剪裁弹窗
const onSelectFile = async (data: { file: UploadFileInfo, fileList: UploadFileInfo[] }) => {
  const file = data.file.file
  if (!file) return false
  
  // 转为 Base64 供剪裁器预览
  const reader = new FileReader()
  reader.readAsDataURL(file)
  reader.onload = (e) => {
    cropperOptions.img = e.target?.result as string
    showCropperModal.value = true // 打开剪裁弹窗
  }
  return false // 阻止 NaiveUI 自动上传
}

// 🔥 2. 确认剪裁并上传
const uploadCroppedImage = () => {
  cropperLoading.value = true
  // 获取剪裁后的 Blob 数据
  cropperRef.value.getCropBlob(async (blob: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', blob, 'icon_cropped.png')
      
      const res: any = await request.post('/forum/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      
      boardModel.value.icon = res.url
      message.success('图标设置成功')
      showCropperModal.value = false // 关闭剪裁窗
    } catch (e) {
      message.error('上传失败')
    } finally {
      cropperLoading.value = false
    }
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
  try { await request.delete(`/admin/forum/boards/${id}`); message.success('删除成功'); fetchBoards() } catch (e: any) { message.error(e.response?.data?.error || '删除失败') }
}

const boardColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '图标', key: 'icon', width: 80, align: 'center',
    render: (row: any) => row.icon 
      ? h(NAvatar, { src: processUrl(row.icon), size: 40, style: 'background: transparent;' }) 
      : h(NIcon, { size: 30, color: '#ddd' }, { default: () => h(AddOutline) })
  },
  { title: '名称', key: 'name', width: 150, render: (row: any) => h('b', row.name) },
  { title: '描述', key: 'description', render: (row: any) => h(NEllipsis, { style: 'max-width: 300px' }, { default: () => row.description }) },
  { title: '排序', key: 'sort_order', width: 80, align: 'center' },
  { 
    title: '权限', key: 'is_locked', width: 120, align: 'center',
    render: (row: any) => row.is_locked 
      ? h(NTag, { type: 'error', size: 'small', round: true }, { default: () => '🔒 锁定', icon: () => h(NIcon, null, { default: () => h(LockClosedOutline) }) })
      : h(NTag, { type: 'success', size: 'small', round: true }, { default: () => '🔓 开放', icon: () => h(NIcon, null, { default: () => h(LockOpenOutline) }) })
  },
  {
    title: '操作', key: 'actions', width: 180, align: 'center',
    render: (row: any) => h(NSpace, { justify: 'center' }, { default: () => [
      h(NButton, { size: 'small', secondary: true, type: 'primary', onClick: () => handleEditBoard(row) }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }) }),
      h(NPopconfirm, { onPositiveClick: () => deleteBoard(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }) }), default: () => '确定删除？板块内必须无帖子才可删除。' })
    ]})
  }
]

// =============================================================
// 🔵 Tab 2 & 3 & 4 (保持不变)
// =============================================================
const postList = ref<any[]>([])
const postPagination = reactive({ page: 1, pageSize: 10, itemCount: 0, onChange: (p: number) => { postPagination.page = p; fetchPosts() } })
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
  if (!postModel.value.title || !postModel.value.board_id) return message.warning('请补全信息')
  postLoading.value = true
  try {
    await request.post('/forum/posts', postModel.value)
    message.success('发布成功'); showPostModal.value = false; fetchPosts()
  } catch (e) { message.error('失败') } finally { postLoading.value = false }
}
const deletePost = async (id: number) => {
  try { await request.delete(`/admin/forum/posts/${id}`); message.success('已删除'); fetchPosts() } catch (e) { message.error('失败') }
}
const postColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { title: '标题', key: 'title', width: 250, render: (row: any) => h('div', { style: 'display:flex; align-items:center;' }, [ row.is_pinned ? h(NTag, { type: 'warning', size: 'small', style: 'margin-right:6px' }, { default: () => '顶' }) : null, h(NEllipsis, { style: 'max-width: 200px; font-weight: 600;' }, { default: () => row.title }) ]) },
  { title: '摘要', key: 'summary', render: (row: any) => h(NEllipsis, { style: 'max-width: 300px; color: #666;' }, { default: () => stripHtml(row.content, 50) }) },
  { title: '板块', key: 'board', width: 120, render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => row.board?.name }) },
  { title: '作者', key: 'author', width: 140, render: (row: any) => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [ h(NAvatar, { round: true, size: 24, src: row.author?.avatar ? processUrl(row.author.avatar) : undefined }), h('span', row.author?.nickname || row.author?.username) ]) },
  { title: '时间', key: 'created_at', width: 160, render: (row: any) => new Date(row.created_at).toLocaleString() },
  { title: '操作', key: 'actions', fixed: 'right', width: 100, align: 'center', render: (row: any) => h(NPopconfirm, { onPositiveClick: () => deletePost(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { default: () => '删除' }), default: () => '确定删除？' }) }
]

const commentList = ref<any[]>([])
const commentPagination = reactive({ page: 1, pageSize: 10, itemCount: 0, onChange: (p: number) => { commentPagination.page = p; fetchComments() } })
const commentSearch = ref('')
const fetchComments = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/forum/comments', { params: { page: commentPagination.page, page_size: commentPagination.pageSize, q: commentSearch.value } })
    if (res.data) { commentList.value = res.data; commentPagination.itemCount = res.total }
  } finally { loading.value = false }
}
const deleteComment = async (id: number) => {
  try { await request.delete(`/admin/forum/comments/${id}`); message.success('已删除'); fetchComments() } catch (e) { message.error('失败') }
}
const commentColumns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { title: '评论内容', key: 'content', render: (row: any) => h(NEllipsis, { style: 'max-width: 400px', lineClamp: 2 }, { default: () => stripHtml(row.content, 100) }) },
  { title: '发布者', key: 'author', width: 140, render: (row: any) => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [ h(NAvatar, { round: true, size: 24, src: row.author?.avatar ? processUrl(row.author.avatar) : undefined }), h('span', row.author?.nickname || row.author?.username) ]) },
  { title: '帖子ID', key: 'post_id', width: 100, align: 'center', render: (row: any) => h(NTag, { size: 'small' }, { default: () => `#${row.post_id}` }) },
  { title: '时间', key: 'created_at', width: 160, render: (row: any) => new Date(row.created_at).toLocaleString() },
  { title: '操作', key: 'actions', fixed: 'right', width: 100, align: 'center', render: (row: any) => h(NPopconfirm, { onPositiveClick: () => deleteComment(row.id) }, { trigger: () => h(NButton, { size: 'small', type: 'error', secondary: true }, { default: () => '删除' }), default: () => '删除？' }) }
]

const reportList = ref<any[]>([])
const reportPagination = reactive({ page: 1, pageSize: 10, itemCount: 0, onChange: (p: number) => { reportPagination.page = p; fetchReports() } })
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
const resolveReport = async (id: number) => { try { await request.put(`/admin/forum/reports/${id}/resolve`); message.success('已标记'); fetchReports() } catch (e) { message.error('失败') } }
const deleteTargetAndResolve = async (row: any) => {
  try {
    if (row.target_type === 'post') await request.delete(`/admin/forum/posts/${row.target_id}`)
    else await request.delete(`/admin/forum/comments/${row.target_id}`)
    await request.put(`/admin/forum/reports/${row.id}/resolve`)
    message.success('已清理违规内容'); fetchReports()
  } catch (e) { message.error('清理失败') }
}
const openPreview = async (row: any) => {
  showPreview.value = true; previewLoading.value = true
  try {
    const res: any = await request.get('/admin/forum/reports/preview', { params: { target_type: row.target_type, target_id: row.target_id } })
    previewContent.value = res
  } catch (e) { previewContent.value = { title: '内容已删除', content: '...', author: '未知', type: 'error' }
  } finally { previewLoading.value = false }
}
const reportColumns: DataTableColumns = [
  { title: '状态', key: 'status', width: 90, align: 'center', render: (row: any) => row.status === 0 ? h(NTag, { type: 'error', size: 'small', round: true }, { default: () => '待处理' }) : h(NTag, { type: 'success', size: 'small', round: true, bordered: false, style: 'opacity: 0.7' }, { default: () => '已处理' }) },
  { title: '类型', key: 'target_type', width: 80, align: 'center', render: (row: any) => row.target_type === 'post' ? h(NTag, { type: 'info', size: 'small' }, { default: () => '帖子' }) : h(NTag, { type: 'warning', size: 'small' }, { default: () => '评论' }) },
  { title: '理由', key: 'reason', width: 120, render: (row: any) => h('span', { style: 'color: #d03050; font-weight: bold;' }, row.reason) },
  { title: '举报人', key: 'reporter', width: 140, render: (row: any) => h('div', { style: 'display:flex;align-items:center;gap:8px' }, [ h(NAvatar, { round: true, size: 20, src: row.reporter?.avatar ? processUrl(row.reporter.avatar) : undefined }), h('span', row.reporter?.nickname || row.reporter?.username) ]) },
  { title: '操作', key: 'actions', fixed: 'right', width: 180, align: 'center', render: (row: any) => row.status === 0 ? h(NSpace, { justify: 'center' }, [ h(NTooltip, null, { trigger: () => h(NButton, { size: 'small', circle: true, type: 'info', onClick: () => openPreview(row) }, { icon: () => h(NIcon, null, { default: () => h(EyeOutline) }) }), default: () => '查看详情' }), h(NPopconfirm, { onPositiveClick: () => resolveReport(row.id) }, { trigger: () => h(NButton, { size: 'small', circle: true, secondary: true }, { icon: () => h(NIcon, null, { default: () => h(CheckmarkCircleOutline) }) }), default: () => '忽略' }), h(NPopconfirm, { onPositiveClick: () => deleteTargetAndResolve(row) }, { trigger: () => h(NButton, { size: 'small', circle: true, type: 'error' }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }) }), default: () => '删除违规内容' }) ]) : h('span', { style: 'color:#ccc' }, '已归档') }
]

onMounted(() => { fetchBoards(); fetchPosts(); fetchComments(); fetchReports() })
</script>

<template>
  <div class="page-container">
    <n-card :bordered="false" content-style="padding: 0;">
      <n-tabs type="line" size="large" :tabs-padding="20" pane-style="padding: 20px;">
        
        <n-tab-pane name="board" tab="板块配置">
           <div class="tab-header">
             <div class="desc">管理论坛版块图标及权限。</div>
             <n-button type="primary" @click="handleAddBoard"><template #icon><n-icon><AddOutline /></n-icon></template>新建</n-button>
           </div>
           <n-data-table :columns="boardColumns" :data="boardList" :loading="loading" />
        </n-tab-pane>

        <n-tab-pane name="post" tab="帖子管理">
            <div class="tab-header">
             <div class="desc">管理全站帖子与公告。</div>
             <n-space>
               <n-button @click="fetchPosts"><template #icon><n-icon><RefreshOutline /></n-icon></template>刷新</n-button>
               <n-button type="primary" color="#8a2be2" @click="handleCreatePost"><template #icon><n-icon><MegaphoneOutline /></n-icon></template>发公告</n-button>
             </n-space>
           </div>
           <n-data-table :columns="postColumns" :data="postList" :loading="loading" :pagination="postPagination" remote />
        </n-tab-pane>

        <n-tab-pane name="comment" tab="评论管理">
            <div class="tab-header">
             <div class="desc">查看并管理全站评论。</div>
             <n-space>
               <n-input v-model:value="commentSearch" placeholder="搜索评论..." @keydown.enter="fetchComments" size="small" />
               <n-button @click="fetchComments"><template #icon><n-icon><RefreshOutline /></n-icon></template>刷新</n-button>
             </n-space>
           </div>
           <n-data-table :columns="commentColumns" :data="commentList" :loading="loading" :pagination="commentPagination" remote />
        </n-tab-pane>

        <n-tab-pane name="report" tab="举报中心">
            <div class="tab-header">
             <div class="desc">处理用户提交的举报。</div>
             <n-button @click="fetchReports"><template #icon><n-icon><RefreshOutline /></n-icon></template>刷新</n-button>
           </div>
           <n-data-table :columns="reportColumns" :data="reportList" :loading="loading" :pagination="reportPagination" remote />
        </n-tab-pane>

      </n-tabs>
    </n-card>

    <n-modal v-model:show="showBoardModal" preset="card" title="板块配置" style="width: 500px">
      <n-form label-placement="left" label-width="80px">
        <n-form-item label="图标">
           <div style="display: flex; align-items: center; gap: 15px;">
              <n-upload 
                :show-file-list="false" 
                accept="image/*"
                @before-upload="onSelectFile"
              >
                 <n-button secondary type="primary">
                    <template #icon><n-icon><CropOutline /></n-icon></template>
                    选择图片
                 </n-button>
              </n-upload>
              
              <div v-if="boardModel.icon" style="position: relative;">
                 <n-avatar :src="processUrl(boardModel.icon)" size="large" style="background:transparent; border: 1px solid #eee;" />
                 <n-icon size="18" color="#d03050" style="position: absolute; top: -5px; right: -5px; cursor: pointer; background: #fff; border-radius: 50%;" @click="boardModel.icon = ''">
                    <CloseCircleOutline />
                 </n-icon>
              </div>
              <span v-else style="color: #999; font-size: 12px;">(支持 1:1 剪裁)</span>
           </div>
        </n-form-item>
        <n-form-item label="名称"><n-input v-model:value="boardModel.name" /></n-form-item>
        <n-form-item label="描述"><n-input v-model:value="boardModel.description" type="textarea" /></n-form-item>
        <n-form-item label="排序"><n-input-number v-model:value="boardModel.sort_order" /></n-form-item>
        <n-form-item label="锁定"><n-switch v-model:value="boardModel.is_locked" /></n-form-item>
      </n-form>
      <template #footer><div style="text-align:right"><n-button type="primary" @click="submitBoard">保存</n-button></div></template>
    </n-modal>

    <n-modal v-model:show="showCropperModal" preset="card" title="剪裁图标" style="width: 600px;">
        <div style="height: 400px; width: 100%;">
            <VueCropper
                ref="cropperRef"
                :img="cropperOptions.img"
                :outputSize="1"
                :outputType="'png'"
                :info="true"
                :canScale="true"
                :autoCrop="true"
                :autoCropWidth="200"
                :autoCropHeight="200"
                :fixed="true"
                :fixedNumber="[1, 1]"
                :centerBox="true"
            ></VueCropper>
        </div>
        <template #footer>
            <div style="display:flex; justify-content:flex-end; gap:12px;">
                <n-button @click="showCropperModal = false">取消</n-button>
                <n-button type="primary" :loading="cropperLoading" @click="uploadCroppedImage">
                    确认并上传
                </n-button>
            </div>
        </template>
    </n-modal>

    <n-modal v-model:show="showPostModal" preset="card" title="发布公告" style="width: 900px;">
      <n-form>
        <n-form-item label="标题"><n-input v-model:value="postModel.title" /></n-form-item>
        <n-form-item label="板块"><n-select v-model:value="postModel.board_id" :options="boardOptions" /></n-form-item>
        <div style="border: 1px solid #ccc; margin-top: 10px;">
            <Toolbar style="border-bottom: 1px solid #ccc" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
            <Editor style="height: 400px; overflow-y: hidden;" v-model="postModel.content" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
        </div>
      </n-form>
      <template #footer><n-button type="primary" :loading="postLoading" @click="submitPost">发布</n-button></template>
    </n-modal>

    <n-modal v-model:show="showPreview" preset="card" title="违规内容预览" style="width: 600px;">
        <n-spin :show="previewLoading">
            <div v-if="previewContent.type !== 'error'">
                <div style="margin-bottom: 15px;">
                    <n-tag type="info" size="small">{{ previewContent.type }}</n-tag>
                    <span style="font-weight: bold; margin-left: 10px;">{{ previewContent.title }}</span>
                </div>
                <n-divider />
                <div class="rich-content-preview" v-html="processContent(previewContent.content)"></div>
            </div>
        </n-spin>
    </n-modal>
  </div>
</template>

<style scoped>
.page-container { padding: 24px; }
.tab-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; background: #f9fafb; padding: 12px 16px; border-radius: 8px; border: 1px solid #f3f4f6; }
.tab-header .desc { color: #6b7280; font-size: 14px; }
:deep(.w-e-panel-content-emotion-list) { height: 200px !important; overflow-y: auto !important; padding: 10px; }
.rich-content-preview { font-size: 14px; line-height: 1.6; color: #333; max-height: 60vh; overflow-y: auto; }
.rich-content-preview :deep(img) { max-width: 100%; border-radius: 4px; }
</style>