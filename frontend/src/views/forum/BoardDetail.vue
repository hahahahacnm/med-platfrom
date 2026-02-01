<script setup lang="ts">
import { ref, onMounted, reactive, shallowRef, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  NCard, NButton, NIcon, NInput, NAvatar, NTag, NPagination, NSpin, 
  NGrid, NGi, NList, NListItem, NModal, NForm, NFormItem, useMessage, 
  NBreadcrumb, NBreadcrumbItem
} from 'naive-ui'
import { 
  CreateOutline, SearchOutline, ChatboxEllipsesOutline, EyeOutline, 
  Flame, HomeOutline // 🔥 修正：Fire 改为 Flame，移除未使用的 ArrowBackOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'
// 移除未使用的 userStore 引用

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
const hotPosts = ref<any[]>([]) // 侧边栏热门

// 发帖
const showPostModal = ref(false)
const postLoading = ref(false)
const postModel = ref({ title: '', content: '<p></p>', board_id: boardId })

// WangEditor
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
        if (res.url) insertFn(processUrl(res.url), '', '') 
      } 
    } 
  } 
}
onBeforeUnmount(() => { const editor = editorRef.value; if (editor == null) return; editor.destroy() })
const handleCreated = (editor: any) => { editorRef.value = editor }

// 辅助函数
const processUrl = (url: string) => url && url.startsWith('http') ? url : `http://localhost:8080${url}`
const getAvatar = (url: string) => url ? processUrl(url) : undefined
const stripHtml = (html: string) => { 
  if(!html) return ''; 
  const tmp = document.createElement("DIV"); 
  tmp.innerHTML = html; 
  let text = tmp.textContent || tmp.innerText || ""; 
  return text.length > 60 ? text.slice(0, 60) + '...' : text 
}
const formatDate = (str: string) => new Date(str).toLocaleDateString()

// 业务逻辑
const fetchBoardInfo = async () => {
  const res: any = await request.get('/forum/boards')
  if (res.data) {
    boardInfo.value = res.data.find((b: any) => b.id === boardId)
  }
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
  fetchBoardInfo()
  fetchPosts()
  fetchHotPosts()
})
</script>

<template>
  <div class="board-detail-page">
    <div class="container">
      
      <div class="breadcrumb-bar">
        <n-breadcrumb>
          <n-breadcrumb-item @click="router.push('/forum')"><n-icon><HomeOutline /></n-icon> 社区首页</n-breadcrumb-item>
          <n-breadcrumb-item v-if="boardInfo">{{ boardInfo.name }}</n-breadcrumb-item>
        </n-breadcrumb>
      </div>

      <n-grid x-gap="24" cols="1 900:3" item-responsive>
        <n-gi span="2">
          
          <div class="board-header-card" v-if="boardInfo">
             <div class="bh-icon">
                <img :src="processUrl(boardInfo.icon) || 'https://cdn-icons-png.flaticon.com/512/2659/2659360.png'" />
             </div>
             <div class="bh-info">
                <h1 class="bh-title">{{ boardInfo.name }}</h1>
                <p class="bh-desc">{{ boardInfo.description || '暂无介绍' }}</p>
                <div class="bh-stats">
                   <span>帖子: {{ boardInfo.post_count }}</span>
                   <span class="divider">/</span>
                   <span>评论: {{ boardInfo.comment_count }}</span>
                </div>
             </div>
             <div class="bh-action">
                <n-button type="primary" color="#36ad6a" size="large" round @click="showPostModal = true">
                   <template #icon><n-icon><CreateOutline /></n-icon></template> 发布新帖
                </n-button>
             </div>
          </div>

          <div class="filter-bar">
             <div class="left">
                <span class="active-tab">全部动态</span>
             </div>
             <div class="right">
                <n-input v-model:value="searchText" round placeholder="搜索本版..." size="small" @keydown.enter="fetchPosts">
                   <template #prefix><n-icon><SearchOutline /></n-icon></template>
                </n-input>
             </div>
          </div>

          <div class="post-list-box">
             <n-spin :show="loading">
                <div v-if="posts.length > 0">
                   <div v-for="post in posts" :key="post.id" class="post-row" @click="router.push(`/post/${post.id}`)">
                      <div class="pr-main">
                         <div class="pr-header">
                            <n-tag v-if="post.is_pinned" type="warning" size="small" class="pin">置顶</n-tag>
                            <h3 class="title">{{ post.title }}</h3>
                         </div>
                         <p class="summary">{{ stripHtml(post.content) }}</p>
                         <div class="pr-meta">
                            <div class="user">
                               <n-avatar round size="small" :src="getAvatar(post.author?.avatar)" class="avt" />
                               {{ post.author?.nickname || post.author?.username }}
                            </div>
                            <span class="time">{{ formatDate(post.created_at) }}</span>
                         </div>
                      </div>
                      <div class="pr-stats">
                         <div class="stat"><n-icon><EyeOutline /></n-icon> {{ post.view_count }}</div>
                         <div class="stat"><n-icon><ChatboxEllipsesOutline /></n-icon> {{ post.comment_count }}</div>
                      </div>
                   </div>
                   <div class="pagination-box">
                      <n-pagination v-model:page="pagination.page" :page-size="pagination.pageSize" :item-count="pagination.total" @update:page="fetchPosts" />
                   </div>
                </div>
                <div v-else class="empty">暂无帖子，来抢沙发！</div>
             </n-spin>
          </div>
        </n-gi>

        <n-gi span="1">
           <n-card :bordered="false" class="sidebar-card">
              <template #header>
                  <div class="sh-title">
                      <n-icon color="#d03050"><Flame /></n-icon> 
                      全站热点
                  </div>
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

    <n-modal v-model:show="showPostModal" preset="card" title="发布新帖" style="width: 900px; max-width: 95vw;">
      <n-form>
        <n-form-item label="标题"><n-input v-model:value="postModel.title" /></n-form-item>
        <n-tag type="success" style="margin-bottom: 15px;">发布到：{{ boardInfo?.name }}</n-tag>
        <div style="border: 1px solid #ccc;"><Toolbar style="border-bottom: 1px solid #ccc" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" /><Editor style="height: 350px; overflow-y: hidden;" v-model="postModel.content" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" /></div>
      </n-form>
      <template #footer><div style="text-align:right"><n-button type="primary" :loading="postLoading" @click="submitPost">发布</n-button></div></template>
    </n-modal>
  </div>
</template>

<style scoped>
.board-detail-page { background: #f7f9fa; min-height: 100vh; padding: 20px; }
.container { max-width: 1200px; margin: 0 auto; }
.breadcrumb-bar { margin-bottom: 20px; font-size: 16px; }

/* 板块头部 */
.board-header-card { background: #fff; border-radius: 12px; padding: 24px; display: flex; align-items: center; gap: 24px; margin-bottom: 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.03); }
.bh-icon { width: 80px; height: 80px; border-radius: 12px; overflow: hidden; background: #f5f5f5; flex-shrink: 0; }
.bh-icon img { width: 100%; height: 100%; object-fit: cover; }
.bh-info { flex: 1; }
.bh-title { margin: 0 0 8px 0; font-size: 24px; color: #333; }
.bh-desc { color: #666; font-size: 14px; margin: 0 0 12px 0; }
.bh-stats { color: #999; font-size: 13px; }
.bh-stats .divider { margin: 0 8px; color: #eee; }

/* 筛选条 */
.filter-bar { background: #fff; padding: 12px 20px; border-radius: 12px 12px 0 0; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #f0f0f0; }
.active-tab { color: #36ad6a; font-weight: bold; border-bottom: 2px solid #36ad6a; padding-bottom: 10px; margin-bottom: -13px; display: inline-block; }

/* 列表 */
.post-list-box { background: #fff; border-radius: 0 0 12px 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.02); }
.post-row { padding: 20px; border-bottom: 1px solid #f5f5f5; display: flex; justify-content: space-between; align-items: center; cursor: pointer; transition: background 0.2s; }
.post-row:hover { background: #fafafa; }
.pr-main { flex: 1; margin-right: 20px; }
.pr-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.title { font-size: 16px; font-weight: 600; color: #333; margin: 0; }
.summary { font-size: 13px; color: #888; margin: 0 0 10px 0; }
.pr-meta { display: flex; align-items: center; gap: 15px; font-size: 12px; color: #999; }
.user { display: flex; align-items: center; gap: 6px; }
.avt { width: 20px; height: 20px; }
.pr-stats { display: flex; gap: 16px; color: #bbb; font-size: 12px; }
.stat { display: flex; align-items: center; gap: 4px; }
.pagination-box { padding: 24px; display: flex; justify-content: center; }
.empty { padding: 40px; text-align: center; color: #999; }

/* 侧边栏 */
.sidebar-card { border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.03); }
.sh-title { font-weight: bold; display: flex; align-items: center; gap: 6px; }
.hot-row { display: flex; gap: 10px; align-items: center; }
.rank { background: #eee; color: #999; width: 18px; height: 18px; border-radius: 4px; text-align: center; line-height: 18px; font-size: 12px; flex-shrink: 0; }
.rank.top { background: #ff7875; color: #fff; }
.txt { font-size: 13px; color: #444; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>