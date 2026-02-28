<script setup lang="ts">
import { ref, onMounted, reactive, shallowRef, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  NButton, NIcon, NAvatar, NTag, NPagination, NSpin, 
  NGrid, NGi, NList, NListItem, NModal, NForm, NFormItem, useMessage, 
  NBreadcrumb, NBreadcrumbItem, NInput, NCard
} from 'naive-ui'
import { 
  CreateOutline, SearchOutline, EyeOutline, ChatboxEllipsesOutline,
  HomeOutline, Flame
} from '@vicons/ionicons5'
import request from '../../utils/request'
import '@wangeditor/editor/dist/css/style.css' 
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()

// 状态
const boardId = Number(route.params.id)
const boardInfo = ref<any>(null)
const posts = ref<any[]>([])
const loading = ref(false)
const searchText = ref('')
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const hotPosts = ref<any[]>([]) 

// 发帖
const showPostModal = ref(false)
const postLoading = ref(false)
const postModel = ref({ title: '', content: '<p></p>', board_id: boardId })

// Editor 配置
const editorRef = shallowRef()
const mode = 'default'
const toolbarConfig = { excludeKeys: ['group-video', 'fullScreen'] } 
const editorConfig = { 
  placeholder: '在此输入正文...', 
  MENU_CONF: { 
    uploadImage: { 
      server: '/api/v1/forum/upload', 
      fieldName: 'file', 
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }, 
      customInsert(res: any, insertFn: any) { 
        if (res.url) insertFn(urlProxy(res.url), '', '') 
      } 
    } 
  } 
}
onBeforeUnmount(() => { const editor = editorRef.value; if (editor == null) return; editor.destroy() })
const handleCreated = (editor: any) => { editorRef.value = editor }

const urlProxy = (url: string) => url && url.startsWith('http') ? url : `http://localhost:8080${url}`
const stripHtml = (html: string) => { 
  if(!html) return ''; 
  const tmp = document.createElement("DIV"); 
  tmp.innerHTML = html; 
  let text = tmp.textContent || tmp.innerText || ""; 
  return text.length > 60 ? text.slice(0, 60) + '...' : text 
}
const formatDate = (str: string) => new Date(str).toLocaleDateString()

const fetchBoardInfo = async () => {
  const res: any = await request.get('/forum/boards')
  if (res.data) boardInfo.value = res.data.find((b: any) => b.id === boardId)
}

const fetchPosts = async () => {
  loading.value = true
  try {
    const params: any = { page: pagination.page, page_size: pagination.pageSize, q: searchText.value, board_id: boardId }
    const res: any = await request.get('/forum/posts', { params })
    if (res.data) { posts.value = res.data; pagination.total = res.total }
  } finally { loading.value = false }
}

const fetchHotPosts = async () => {
  const res: any = await request.get('/forum/posts', { params: { page: 1, page_size: 10 } })
  if (res.data) hotPosts.value = res.data.sort((a: any, b: any) => b.view_count - a.view_count).slice(0, 5)
}

const submitPost = async () => {
  if (!postModel.value.title || postModel.value.content === '<p><br></p>') return message.warning('请填写完整')
  postLoading.value = true
  try {
    await request.post('/forum/posts', postModel.value)
    message.success('发布成功')
    showPostModal.value = false
    fetchPosts()
  } catch (e) { message.error('失败') } finally { postLoading.value = false }
}

onMounted(() => {
  fetchBoardInfo(); fetchPosts(); fetchHotPosts()
})
</script>

<template>
  <div class="board-page">
    <div class="container">
      
      <div class="breadcrumb-wrapper">
        <n-breadcrumb>
          <n-breadcrumb-item @click="router.push('/forum')"><n-icon><HomeOutline /></n-icon> 社区</n-breadcrumb-item>
          <n-breadcrumb-item v-if="boardInfo">{{ boardInfo.name }}</n-breadcrumb-item>
        </n-breadcrumb>
      </div>

      <n-grid x-gap="20" y-gap="20" cols="1 m:3" item-responsive>
        <n-gi span="2">
          <div class="board-info-card" v-if="boardInfo">
             <div class="bi-main">
                <img :src="urlProxy(boardInfo.icon)" class="bi-icon" />
                <div class="bi-text">
                   <h1 class="bi-title">{{ boardInfo.name }}</h1>
                   <p class="bi-desc">{{ boardInfo.description || '暂无介绍' }}</p>
                </div>
             </div>
             <div class="bi-action">
                <n-button type="primary" round class="create-btn" @click="showPostModal = true">
                   <template #icon><n-icon><CreateOutline /></n-icon></template> 发布帖子
                </n-button>
             </div>
          </div>

          <div class="post-container">
             <div class="list-toolbar">
                <span class="tab active">最新动态</span>
                <div class="search-box">
                   <n-input v-model:value="searchText" round placeholder="搜索..." size="small" @keydown.enter="fetchPosts">
                      <template #prefix><n-icon><SearchOutline /></n-icon></template>
                   </n-input>
                </div>
             </div>

             <n-spin :show="loading">
                <div class="post-list" v-if="posts.length > 0">
                   <div v-for="post in posts" :key="post.id" class="post-item" @click="router.push(`/post/${post.id}`)">
                      <div class="pi-body">
                         <div class="pi-title-row">
                            <n-tag v-if="post.is_pinned" type="warning" size="small" class="pin-tag">置顶</n-tag>
                            <h3 class="pi-title">{{ post.title }}</h3>
                         </div>
                         <p class="pi-summary">{{ stripHtml(post.content) }}</p>
                         
                         <div class="pi-meta">
                            <div class="author">
                               <n-avatar round size="small" :src="urlProxy(post.author?.avatar)" />
                               <span>{{ post.author?.nickname || '匿名' }}</span>
                            </div>
                            <span class="dot">·</span>
                            <span class="time">{{ formatDate(post.created_at) }}</span>
                         </div>
                      </div>
                      
                      <div class="pi-stats mobile-hidden">
                         <div class="stat"><n-icon><EyeOutline /></n-icon> {{ post.view_count }}</div>
                         <div class="stat"><n-icon><ChatboxEllipsesOutline /></n-icon> {{ post.comment_count }}</div>
                      </div>
                   </div>
                </div>
                <div v-else class="empty-box">暂无内容，快来发布第一篇吧！</div>
                
                <div class="pagination-wrapper" v-if="posts.length > 0">
                   <n-pagination v-model:page="pagination.page" :page-size="pagination.pageSize" :item-count="pagination.total" @update:page="fetchPosts" />
                </div>
             </n-spin>
          </div>
        </n-gi>

        <n-gi span="1">
           <n-card size="small" :bordered="false" class="sidebar-card">
              <template #header>
                  <div class="sh-header"><n-icon color="#d03050"><Flame /></n-icon> 热门讨论</div>
              </template>
              <n-list hoverable clickable>
                 <n-list-item v-for="(post, idx) in hotPosts" :key="post.id" @click="router.push(`/post/${post.id}`)">
                    <div class="hot-row">
                       <span class="rank" :class="{'top': idx<3}">{{ idx+1 }}</span>
                       <span class="txt">{{ post.title }}</span>
                    </div>
                 </n-list-item>
              </n-list>
           </n-card>
        </n-gi>
      </n-grid>
    </div>

    <n-modal v-model:show="showPostModal" preset="card" title="发布新帖" class="editor-modal" style="width: 800px; max-width: 100vw;">
      <n-form>
        <n-form-item label="标题"><n-input v-model:value="postModel.title" placeholder="请输入标题" /></n-form-item>
        <div class="editor-box">
           <Toolbar style="border-bottom: 1px solid #ccc" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
           <Editor style="height: 300px; overflow-y: hidden;" v-model="postModel.content" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
        </div>
      </n-form>
      <template #footer><n-button type="primary" block @click="submitPost" :loading="postLoading">立即发布</n-button></template>
    </n-modal>
  </div>
</template>

<style scoped>
.board-page { background: #f8fafc; min-height: 100vh; padding: 20px; }
.container { max-width: 1100px; margin: 0 auto; }
.breadcrumb-wrapper { margin-bottom: 16px; }

/* 板块信息卡片 */
.board-info-card {
  background: #fff; border-radius: 12px; padding: 20px; margin-bottom: 20px;
  display: flex; align-items: center; justify-content: space-between;
  border: 1px solid #e2e8f0; flex-wrap: wrap; gap: 16px;
}
.bi-main { display: flex; align-items: center; gap: 16px; flex: 1; }
.bi-icon { width: 64px; height: 64px; border-radius: 12px; object-fit: cover; background: #f1f5f9; }
.bi-title { margin: 0; font-size: 20px; color: #1e293b; }
.bi-desc { margin: 4px 0 0 0; color: #64748b; font-size: 13px; }
.bi-action { flex-shrink: 0; }

/* 响应式适配 */
@media (max-width: 640px) {
  .board-page { padding: 12px; }
  .board-info-card { flex-direction: column; align-items: flex-start; }
  .bi-action { width: 100%; }
  .create-btn { width: 100%; }
  .mobile-hidden { display: none !important; }
}

/* 帖子列表容器 */
.post-container { background: #fff; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; }
.list-toolbar { padding: 12px 20px; border-bottom: 1px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; }
.tab.active { font-weight: bold; color: #2080f0; position: relative; }
.tab.active::after { content: ''; position: absolute; bottom: -13px; left: 0; width: 100%; height: 2px; background: #2080f0; }

.post-item { padding: 16px 20px; border-bottom: 1px solid #f8fafc; display: flex; justify-content: space-between; align-items: center; cursor: pointer; transition: background 0.2s; }
.post-item:hover { background: #f8fafc; }
.pi-body { flex: 1; min-width: 0; margin-right: 12px; }
.pi-title-row { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.pi-title { margin: 0; font-size: 16px; font-weight: 600; color: #334155; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pi-summary { color: #94a3b8; font-size: 13px; margin: 0 0 10px 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pi-meta { display: flex; align-items: center; gap: 8px; font-size: 12px; color: #94a3b8; }
.author { display: flex; align-items: center; gap: 4px; }

.pi-stats { display: flex; flex-direction: column; align-items: flex-end; gap: 4px; color: #cbd5e1; font-size: 12px; min-width: 60px; }
.stat { display: flex; align-items: center; gap: 4px; }

.empty-box { padding: 40px; text-align: center; color: #94a3b8; }
.pagination-wrapper { padding: 20px; display: flex; justify-content: center; }

/* 侧边栏 */
.sidebar-card { border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
.sh-header { font-weight: bold; display: flex; align-items: center; gap: 6px; }
.hot-row { display: flex; gap: 10px; align-items: center; }
.rank { width: 18px; height: 18px; background: #f1f5f9; color: #64748b; font-size: 12px; text-align: center; line-height: 18px; border-radius: 4px; flex-shrink: 0; }
.rank.top { background: #fee2e2; color: #ef4444; }
.txt { font-size: 13px; color: #475569; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.editor-box { border: 1px solid #ccc; border-radius: 4px; }
</style>