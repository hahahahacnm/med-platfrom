<script setup lang="ts">
import { ref, onMounted, shallowRef, onBeforeUnmount, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  NCard, NButton, NIcon, NSkeleton, NAvatar, NTag, NDivider, useMessage, 
  NSpace, NModal, NRadioGroup, NRadio, NPopconfirm, NAlert, NTooltip, NEmpty
} from 'naive-ui'
import { 
  ArrowBackOutline, TimeOutline, EyeOutline, 
  AlertCircleOutline, ChatbubbleEllipsesOutline, CloseOutline,
  TrashOutline, ShareSocialOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

import '@wangeditor/editor/dist/css/style.css' 
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const loading = ref(true)
const post = ref<any>(null)
const rawComments = ref<any[]>([]) 
const commentLoading = ref(false)
const replyTarget = ref<any>(null) 

// 举报相关
const showReportModal = ref(false)
const reportReason = ref('垃圾广告')
const reportTargetId = ref(0)
const reportType = ref('post') 
const reportLoading = ref(false)

// 大图预览
const showImagePreview = ref(false)
const previewImageUrl = ref('')

// 编辑器配置
const editorRef = shallowRef()
const contentHtml = ref('<p></p>') 
const mode = 'simple' 
const toolbarConfig = {
  toolbarKeys: ['bold', 'italic', 'color', '|', 'emotion', 'uploadImage', '|', 'clearStyle']
}
const editorConfig = { 
  placeholder: '善语结善缘，聊聊你的看法...',
  MENU_CONF: {
    uploadImage: {
      server: '/api/v1/forum/upload',
      fieldName: 'file',
      maxFileSize: 5 * 1024 * 1024,
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
      customInsert(res: any, insertFn: any) {
        if (res.url) {
            const url = res.url.startsWith('http') ? res.url : `http://localhost:8080${res.url}`
            insertFn(url, '', '')
        } else {
          message.error('图片上传失败')
        }
      }
    }
  }
}

onBeforeUnmount(() => {
    const editor = editorRef.value
    if (editor == null) return
    editor.destroy()
})
const handleCreated = (editor: any) => { editorRef.value = editor }

const commentTree = computed(() => {
  const map: any = {}
  const roots: any[] = []
  rawComments.value.forEach(c => {
    c.children = [] 
    map[c.id] = c
  })
  rawComments.value.forEach(c => {
    if (c.parent_id && map[c.parent_id]) {
      map[c.parent_id].children.push(c)
    } else {
      roots.push(c)
    }
  })
  return roots
})

const processContent = (html: string) => {
  if (!html) return ''
  return html.replace(/src="\/uploads\//g, 'src="http://localhost:8080/uploads/')
}

const fetchDetail = async () => {
  loading.value = true
  try {
    const res: any = await request.get(`/forum/posts/${route.params.id}`)
    if (res.data) {
      post.value = res.data
      fetchComments() 
    }
  } catch (e) {
    message.error('获取帖子详情失败')
  } finally {
    loading.value = false
  }
}

const fetchComments = async () => {
  const res: any = await request.get('/forum/comments', { params: { post_id: route.params.id } })
  if (res.data) rawComments.value = res.data
}

const handleReplyClick = (comment: any) => {
  replyTarget.value = comment
  const editorDom = document.querySelector('.comment-editor-box')
  editorDom?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

const submitComment = async () => {
  if (contentHtml.value === '<p><br></p>' || !contentHtml.value.trim()) {
    return message.warning('评论内容不能为空')
  }
  commentLoading.value = true
  try {
    const payload: any = { 
      post_id: Number(route.params.id), 
      content: contentHtml.value 
    }
    if (replyTarget.value) payload.parent_id = replyTarget.value.id

    await request.post('/forum/comments', payload)
    message.success('发布成功')
    contentHtml.value = '<p></p>'
    replyTarget.value = null
    fetchComments()
  } catch (e) {
    message.error('发布失败')
  } finally {
    commentLoading.value = false
  }
}

const deletePost = async () => {
  try {
    await request.delete(`/forum/posts/${post.value.id}`)
    message.success('帖子已删除')
    router.replace('/forum')
  } catch (e) { message.error('删除失败') }
}

const deleteComment = async (id: number) => {
  try {
    await request.delete(`/forum/comments/${id}`)
    message.success('评论已删除')
    fetchComments()
  } catch (e) { message.error('删除失败') }
}

const openReport = (id: number, type: 'post'|'comment') => {
  reportTargetId.value = id
  reportType.value = type
  showReportModal.value = true
}

const submitReport = async () => {
  reportLoading.value = true
  try {
    await request.post('/forum/report', {
      target_id: reportTargetId.value,
      target_type: reportType.value,
      reason: reportReason.value
    })
    message.success('投诉已受理')
    showReportModal.value = false
  } catch (e) {
    message.error('提交失败')
  } finally {
    reportLoading.value = false
  }
}

const handleContentClick = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (target && target.tagName.toLowerCase() === 'img') {
    previewImageUrl.value = (target as HTMLImageElement).src
    showImagePreview.value = true
  }
}

const getAvatar = (path: string) => path && path.startsWith('http') ? path : `http://localhost:8080${path}`
const formatDate = (str: string) => new Date(str).toLocaleString('zh-CN', { hour12: false })

onMounted(() => { fetchDetail() })
</script>

<template>
  <div class="post-detail-wrapper">
    <div class="top-nav-bar animate-in">
        <n-button quaternary circle class="nav-btn" @click="router.back()">
          <template #icon><n-icon size="22"><ArrowBackOutline /></n-icon></template>
        </n-button>
        <div class="nav-title" v-if="post">{{ post.title }}</div>
        <n-button quaternary circle class="nav-btn" @click="message.info('链接已复制')">
          <template #icon><n-icon size="20"><ShareSocialOutline /></n-icon></template>
        </n-button>
    </div>

    <div class="main-layout">
      <n-card :bordered="false" v-if="loading" class="glass-card shadow-soft">
        <n-skeleton text style="width: 40%; height: 32px; margin-bottom: 24px;" />
        <n-skeleton text :repeat="8" />
      </n-card>

      <div v-else-if="post" class="post-content-section animate-in" style="animation-delay: 0.1s;">
        <n-card :bordered="false" class="glass-card shadow-soft">
          <div class="post-header">
            <n-tag :bordered="false" type="primary" size="small" class="board-badge">
              {{ post.board?.name }}
            </n-tag>
            <h1 class="post-main-title">{{ post.title }}</h1>
            
            <div class="author-bar">
              <div class="u-info">
                <n-avatar round :size="44" :src="getAvatar(post.author?.avatar)" class="author-avatar" />
                <div class="u-meta">
                  <span class="u-name">{{ post.author?.nickname || post.author?.username }}</span>
                  <span class="u-time">{{ formatDate(post.created_at) }}</span>
                </div>
              </div>
              <div class="post-actions">
                 <n-tooltip trigger="hover">
                    <template #trigger>
                      <n-button text class="action-btn" @click="openReport(post.id, 'post')">
                        <n-icon size="20"><AlertCircleOutline /></n-icon>
                      </n-button>
                    </template>
                    举报违规内容
                 </n-tooltip>
                 
                 <n-popconfirm v-if="userStore.role === 'admin' || Number(userStore.id) === Number(post.author_id)" @positive-click="deletePost">
                   <template #trigger>
                     <n-button text type="error" class="action-btn delete-btn">
                        <n-icon size="20"><TrashOutline /></n-icon>
                     </n-button>
                   </template>
                   确定要永久删除这篇帖子吗？
                 </n-popconfirm>
              </div>
            </div>
          </div>

          <n-divider class="content-divider" />

          <div class="rich-content-area" @click="handleContentClick">
             <div class="w-e-text-container slate-view" v-html="processContent(post.content)"></div>
          </div>

          <div class="post-footer-stats">
             <span class="stat-item"><n-icon size="16"><EyeOutline /></n-icon> {{ post.view_count }} 次阅读</span>
          </div>
        </n-card>

        <div class="comment-container animate-in" style="animation-delay: 0.2s">
          <div class="comment-header">
            <span class="title">全部评论</span>
            <span class="count">{{ rawComments.length }}</span>
          </div>

          <n-card :bordered="false" class="comment-editor-box shadow-soft">
             <n-alert v-if="replyTarget" type="info" closable @close="replyTarget = null" class="reply-alert">
                正在回复 <b>{{ replyTarget.author?.nickname || replyTarget.author?.username }}</b>
             </n-alert>
             <div class="editor-ui">
                <Toolbar class="t-bar" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
                <Editor class="e-body" v-model="contentHtml" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
             </div>
             <div class="editor-footer">
               <n-button type="primary" round :loading="commentLoading" @click="submitComment" class="send-btn">
                 {{ replyTarget ? '发送回复' : '发布评论' }}
               </n-button>
             </div>
          </n-card>

          <div class="comment-list-flow">
             <div v-for="item in commentTree" :key="item.id" class="comment-card-wrapper">
                <n-card :bordered="false" class="comment-card shadow-soft">
                  <div class="cm-main">
                    <n-avatar round :size="38" :src="getAvatar(item.author?.avatar)" class="cm-avatar" />
                    <div class="cm-content-wrap">
                      <div class="cm-user-row">
                        <span class="cm-username">{{ item.author?.nickname || item.author?.username }}</span>
                        <span class="cm-date">{{ formatDate(item.created_at) }}</span>
                      </div>
                      <div class="cm-text rich-content-area" @click="handleContentClick">
                        <div v-html="processContent(item.content)"></div>
                      </div>
                      <div class="cm-actions">
                        <n-button text size="small" class="reply-btn" @click="handleReplyClick(item)">
                          <template #icon><n-icon><ChatbubbleEllipsesOutline /></n-icon></template>回复
                        </n-button>
                        <n-button text size="small" class="report-btn" @click="openReport(item.id, 'comment')">举报</n-button>
                        <n-button v-if="userStore.role === 'admin' || Number(userStore.id) === Number(item.author_id)" text size="small" type="error" class="delete-btn" @click="deleteComment(item.id)">删除</n-button>
                      </div>
                    </div>
                  </div>

                  <div v-if="item.children?.length" class="sub-comment-box">
                    <div v-for="sub in item.children" :key="sub.id" class="sub-cm-item">
                      <n-avatar round :size="28" :src="getAvatar(sub.author?.avatar)" />
                      <div class="sub-cm-body">
                        <div class="sub-cm-user">
                          <span class="name">{{ sub.author?.nickname || sub.author?.username }}</span>
                          <span class="reply-text">回复</span>
                          <span class="name">{{ item.author?.nickname || item.author?.username }}</span>
                          <span class="date">{{ formatDate(sub.created_at) }}</span>
                        </div>
                        <div class="sub-cm-text rich-content-area" @click="handleContentClick">
                          <div v-html="processContent(sub.content)"></div>
                        </div>
                        <div class="cm-actions sub-actions">
                          <n-button text size="small" class="reply-btn" @click="handleReplyClick(sub)">回复</n-button>
                          <n-button v-if="userStore.role === 'admin' || Number(userStore.id) === Number(sub.author_id)" text size="small" type="error" class="delete-btn" @click="deleteComment(sub.id)">删除</n-button>
                        </div>
                      </div>
                    </div>
                  </div>
                </n-card>
             </div>
             
             <div v-if="!loading && rawComments.length === 0" class="empty-comment-state">
                <img src="https://img.icons8.com/bubbles/100/000000/comments.png" alt="No comments" />
                <p>暂无评论，快来抢下沙发吧~</p>
             </div>
          </div>
        </div>
      </div>
    </div>

    <n-modal v-model:show="showImagePreview" :mask-style="{ backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }" :style="{ background: 'transparent' }">
      <div class="image-zoom-overlay" @click="showImagePreview = false">
        <n-button circle class="close-overlay" @click.stop="showImagePreview = false">
           <n-icon size="24"><CloseOutline /></n-icon>
        </n-button>
        <img :src="previewImageUrl" class="zoom-img" @click.stop />
      </div>
    </n-modal>

    <n-modal v-model:show="showReportModal" preset="card" title="内容违规反馈" :style="{ width: '400px' }">
       <n-radio-group v-model:value="reportReason" style="margin-top: 16px;">
          <n-space vertical size="large">
            <n-radio value="垃圾广告">💸 垃圾广告 / 推广</n-radio>
            <n-radio value="违规内容">🚫 违法违规 / 色情暴力</n-radio>
            <n-radio value="恶意攻击">🤬 人身攻击 / 语言不文明</n-radio>
            <n-radio value="其他">❓ 其他原因</n-radio>
          </n-space>
       </n-radio-group>
       <template #footer>
          <n-space justify="end">
            <n-button @click="showReportModal = false">取消</n-button>
            <n-button type="primary" :loading="reportLoading" @click="submitReport">确认提交</n-button>
          </n-space>
       </template>
    </n-modal>
  </div>
</template>

<style scoped>
/* ================= 全局排版 ================= */
.post-detail-wrapper { 
    background-color: #f1f5f9; 
    min-height: 100vh; 
    padding-bottom: 80px;
}
.main-layout { 
    max-width: 780px; 
    margin: 0 auto; 
    padding: 0 16px; 
    margin-top: 84px;
}

/* ================= 毛玻璃导航 ================= */
.top-nav-bar {
    position: fixed; top: 0; left: 0; right: 0; height: 60px;
    background: rgba(255, 255, 255, 0.75); 
    backdrop-filter: blur(16px); 
    -webkit-backdrop-filter: blur(16px);
    display: flex; align-items: center; justify-content: space-between;
    padding: 0 24px; z-index: 999; 
    border-bottom: 1px solid rgba(226, 232, 240, 0.6);
    box-shadow: 0 2px 10px rgba(0,0,0,0.02);
}
.nav-title { font-weight: 800; font-size: 16px; color: #0f172a; max-width: 60%; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.nav-btn { color: #475569; }

/* ================= 卡片与主体 ================= */
.glass-card { background: #ffffff; border-radius: 20px; transition: all 0.3s ease;}
.shadow-soft { box-shadow: 0 4px 20px rgba(148, 163, 184, 0.08); }

.post-header { padding: 12px 8px 0; }
.board-badge { margin-bottom: 16px; font-weight: 700; border-radius: 8px; padding: 0 12px; height: 26px; }
.post-main-title { font-size: 30px; font-weight: 900; color: #0f172a; line-height: 1.4; margin-bottom: 24px; letter-spacing: 0.5px;}

.author-bar { display: flex; justify-content: space-between; align-items: center; background: #f8fafc; padding: 12px 16px; border-radius: 16px;}
.u-info { display: flex; align-items: center; gap: 14px; }
.author-avatar { border: 2px solid #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
.u-meta { display: flex; flex-direction: column; }
.u-name { font-weight: 800; color: #1e293b; font-size: 15px; }
.u-time { font-size: 12px; color: #94a3b8; margin-top: 3px; }

.post-actions { display: flex; gap: 8px; }
.action-btn { color: #94a3b8; transition: color 0.2s;}
.action-btn:hover { color: #64748b; }
.delete-btn:hover { color: #ef4444 !important; background: #fee2e2; border-radius: 8px;}

.content-divider { margin: 20px 0 28px 0; opacity: 0.6; }

/* ================= 富文本正文展示 ================= */
.rich-content-area { font-size: 17px; line-height: 1.85; color: #334155; word-wrap: break-word; padding: 0 8px;}
.rich-content-area :deep(img) {
    max-width: 180px; 
    max-height: 180px; 
    border-radius: 12px;
    cursor: zoom-in;
    object-fit: cover;
    margin: 8px 6px 8px 0;
    border: 1px solid #e2e8f0;
    box-shadow: 0 2px 8px rgba(0,0,0,0.04);
    transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}
.rich-content-area :deep(img:hover) { transform: translateY(-2px); box-shadow: 0 8px 16px rgba(0,0,0,0.1); }
.rich-content-area :deep(p) { margin-bottom: 1em; }
.rich-content-area :deep(blockquote) { border-left: 4px solid #cbd5e1; background-color: #f8fafc; margin: 10px 0; padding: 12px; border-radius: 0 8px 8px 0;}

.post-footer-stats { margin-top: 48px; padding: 0 8px; font-size: 14px; color: #94a3b8; display: flex; gap: 24px; }
.stat-item { display: flex; align-items: center; gap: 6px; }

/* ================= 评论区 ================= */
.comment-container { margin-top: 32px; }
.comment-header { display: flex; align-items: center; gap: 10px; margin-bottom: 24px; padding-left: 8px;}
.comment-header .title { font-size: 18px; font-weight: 900; color: #0f172a; }
.comment-header .count { background: #e2e8f0; color: #475569; padding: 2px 10px; border-radius: 100px; font-size: 13px; font-weight: 800; }

.comment-editor-box { padding: 20px; border-radius: 20px; margin-bottom: 24px; border: 1px solid rgba(226, 232, 240, 0.8); }
.reply-alert { border-radius: 12px; margin-bottom: 16px; border: none; background: #f0f9ff; }
/* 🔥 修复 1：去掉 overflow: hidden，加上相对定位和 z-index，防止面板被下方评论遮挡 */
.editor-ui { 
    border: 1px solid #e2e8f0; 
    border-radius: 12px; 
    background: #fff;
    position: relative; 
    z-index: 99; 
}

/* 🔥 修复 2：因为去掉了 overflow: hidden，需要单独给工具栏和输入区加上对应的圆角 */
.t-bar { 
    background: #f8fafc !important; 
    border-bottom: 1px solid #e2e8f0 !important; 
    border-radius: 12px 12px 0 0; 
}
.e-body { 
    height: 130px; 
    background: #fff !important; 
    border-radius: 0 0 12px 12px; 
}

/* 🔥 修复 3：强制提升 WangEditor 弹出面板的层级，确保在最顶层显示 */
:deep(.w-e-panel-container) {
    z-index: 9999 !important;
}
:deep(.w-e-menu-tooltip) {
    z-index: 10000 !important;
}
.editor-footer { display: flex; justify-content: flex-end; margin-top: 16px; }
.send-btn { padding: 0 36px; font-weight: 800; height: 40px; box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);}

/* 🔥🔥🔥 核心修复：强制表情面板支持滚动 🔥🔥🔥 */
:deep(.w-e-panel-content-emotion-list) {
    height: 200px !important;       
    overflow-y: auto !important;    
    padding: 10px;                  
}

/* 评论卡片 */
.comment-list-flow { display: flex; flex-direction: column; gap: 16px; }
.comment-card-wrapper { margin-bottom: 16px; }
.comment-card { border-radius: 20px; padding: 20px 16px; }
.cm-main { display: flex; gap: 16px; }
.cm-avatar { box-shadow: 0 2px 6px rgba(0,0,0,0.06); }
.cm-content-wrap { flex: 1; min-width: 0; }
.cm-user-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.cm-username { font-weight: 800; color: #1e293b; font-size: 15px; }
.cm-date { font-size: 12px; color: #94a3b8; font-weight: 500;}

.cm-text { font-size: 15px; line-height: 1.7; color: #334155; margin: 6px 0 10px; }
.cm-text :deep(img) { max-width: 120px; max-height: 120px; } 

.cm-actions { display: flex; gap: 16px; align-items: center; opacity: 0.8;}
.reply-btn, .report-btn { color: #64748b; font-weight: 600;}

/* 嵌套楼中楼 */
.sub-comment-box { background: #f8fafc; border-radius: 16px; padding: 16px 20px; margin-top: 16px; display: flex; flex-direction: column; gap: 20px; border: 1px solid #f1f5f9;}
.sub-cm-item { display: flex; gap: 12px; }
.sub-cm-body { flex: 1; min-width: 0; }
.sub-cm-user { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 6px; }
.sub-cm-user .name { font-weight: 800; font-size: 13px; color: #1e293b; }
.sub-cm-user .reply-text { font-size: 12px; color: #94a3b8; }
.sub-cm-user .date { font-size: 11px; color: #cbd5e1; margin-left: auto; }
.sub-cm-text { font-size: 14px; color: #334155; line-height: 1.6; }
.sub-actions { margin-top: 8px; }

/* 缺省状态 */
.empty-comment-state { text-align: center; padding: 60px 0; color: #94a3b8; font-size: 14px; font-weight: 600;}
.empty-comment-state img { width: 80px; opacity: 0.5; margin-bottom: 10px; filter: grayscale(100%);}

/* ================= 大图预览 ================= */
.image-zoom-overlay { width: 100vw; height: 100vh; display: flex; align-items: center; justify-content: center; position: relative;}
.zoom-img { max-width: 90%; max-height: 85%; border-radius: 12px; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5); object-fit: contain;}
.close-overlay { position: absolute; top: 40px; right: 40px; background: rgba(255,255,255,0.15) !important; color: #fff !important; border: 1px solid rgba(255,255,255,0.3); backdrop-filter: blur(4px); transition: background 0.3s;}
.close-overlay:hover { background: rgba(255,255,255,0.3) !important; }

/* ================= 动画与响应式 ================= */
.animate-in { animation: fadeInUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) both; }
@keyframes fadeInUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 640px) {
    .main-layout { margin-top: 76px; padding: 0 12px; }
    .post-main-title { font-size: 24px; margin-bottom: 20px; }
    .nav-title { display: none; }
    .author-bar { padding: 10px 12px; }
    .rich-content-area :deep(img) { max-width: 130px; max-height: 130px; }
    .cm-text :deep(img) { max-width: 100px; max-height: 100px; }
    .comment-card { padding: 16px 12px; }
    .sub-comment-box { padding: 12px 10px; border-radius: 12px;}
    .close-overlay { top: 20px; right: 20px; }
}
</style>