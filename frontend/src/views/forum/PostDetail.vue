<script setup lang="ts">
import { ref, onMounted, shallowRef, onBeforeUnmount, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  NCard, NButton, NIcon, NSkeleton, NAvatar, NTag, NDivider, useMessage, 
  NThing, NSpace, NModal, NRadioGroup, NRadio, NPopconfirm, NAlert
} from 'naive-ui'
import { 
  ArrowBackOutline, TimeOutline, EyeOutline, 
  AlertCircleOutline, ChatbubbleEllipsesOutline
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

// 编辑器
const editorRef = shallowRef()
const contentHtml = ref('<p></p>') 
const mode = 'simple' 
const toolbarConfig = {
  toolbarKeys: ['bold', 'italic', 'color', '|', 'emotion', 'uploadImage', 'insertLink', '|', 'clearStyle']
}
const editorConfig = { 
  placeholder: '发表你的看法...',
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
          message.error('上传失败')
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

// 树形评论
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
  document.querySelector('.comment-editor-box')?.scrollIntoView({ behavior: 'smooth' })
}

const cancelReply = () => {
  replyTarget.value = null
}

const submitComment = async () => {
  if (contentHtml.value === '<p><br></p>' || !contentHtml.value.trim()) {
    return message.warning('写点什么吧')
  }
  
  commentLoading.value = true
  try {
    const payload: any = { 
      post_id: Number(route.params.id), 
      content: contentHtml.value 
    }
    if (replyTarget.value) {
      payload.parent_id = replyTarget.value.id
    }

    await request.post('/forum/comments', payload)
    message.success('回复成功')
    contentHtml.value = '<p></p>'
    replyTarget.value = null
    fetchComments()
  } catch (e) {
    message.error('评论失败')
  } finally {
    commentLoading.value = false
  }
}

const deletePost = async () => {
  try {
    await request.delete(`/forum/posts/${post.value.id}`)
    message.success('已删除')
    router.replace('/forum')
  } catch (e) { message.error('删除失败') }
}

const deleteComment = async (id: number) => {
  try {
    await request.delete(`/forum/comments/${id}`)
    message.success('已删除')
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
    message.success('举报已提交')
    showReportModal.value = false
  } catch (e) {
    message.error('提交失败')
  } finally {
    reportLoading.value = false
  }
}

const getAvatar = (path: string) => path && path.startsWith('http') ? path : `http://localhost:8080${path}`
const formatDate = (str: string) => new Date(str).toLocaleString()

onMounted(() => {
  fetchDetail()
})
</script>

<template>
  <div class="post-detail-container">
    <div class="main-content">
      <n-button text @click="router.back()" style="margin-bottom: 16px;">
        <template #icon><n-icon><ArrowBackOutline /></n-icon></template>
        返回列表
      </n-button>

      <n-card :bordered="false" v-if="loading">
        <n-skeleton text style="width: 60%; height: 40px; margin-bottom: 20px;" />
        <n-skeleton text :repeat="6" />
      </n-card>

      <n-card :bordered="false" v-else-if="post" class="post-card">
        <div class="post-header-area">
          <h1 class="post-title">{{ post.title }}</h1>
          <n-space>
             <n-button size="tiny" secondary type="error" @click="openReport(post.id, 'post')">
               <template #icon><n-icon><AlertCircleOutline /></n-icon></template>举报
             </n-button>
             <n-popconfirm v-if="userStore.role === 'admin' || userStore.id === post.author_id" @positive-click="deletePost">
               <template #trigger>
                 <n-button size="tiny" secondary type="error">删除</n-button>
               </template>
               确定彻底删除该帖？
             </n-popconfirm>
          </n-space>
        </div>
        
        <div class="post-meta">
          <div class="meta-left">
            <n-tag :bordered="false" type="primary" size="small" class="board-tag">
              {{ post.board?.name }}
            </n-tag>
            <div class="author-info">
              <n-avatar round size="small" :src="getAvatar(post.author?.avatar)" />
              <span class="author-name">{{ post.author?.nickname || post.author?.username }}</span>
            </div>
            <span class="time">
              <n-icon><TimeOutline /></n-icon> {{ formatDate(post.created_at) }}
            </span>
          </div>
          <div class="meta-right">
            <span><n-icon><EyeOutline /></n-icon> {{ post.view_count }}</span>
          </div>
        </div>

        <n-divider />

        <div class="rich-content w-e-text-container" style="min-height: auto; border: none;">
             <div data-slate-editor v-html="processContent(post.content)"></div>
        </div>

        <n-divider />

        <div class="comment-section">
          <h3>参与讨论 ({{ rawComments.length }})</h3>
          
          <div class="comment-editor-box">
             <n-alert v-if="replyTarget" type="info" closable @close="cancelReply" style="margin-bottom: 10px;">
                正在回复 <b>{{ replyTarget.author?.nickname || replyTarget.author?.username }}</b> 的评论...
             </n-alert>

             <div class="editor-wrapper">
                <Toolbar style="border-bottom: 1px solid #eee" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
                <Editor style="height: 150px; overflow-y: hidden;" v-model="contentHtml" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
             </div>

             <div class="ci-actions">
               <n-button type="primary" :loading="commentLoading" @click="submitComment">
                 {{ replyTarget ? '回复评论' : '发表评论' }}
               </n-button>
             </div>
          </div>

          <div class="comment-list">
             <div v-for="item in commentTree" :key="item.id" class="root-comment">
                <n-thing class="comment-item">
                   <template #avatar><n-avatar round size="small" :src="getAvatar(item.author?.avatar)" /></template>
                   <template #header>{{ item.author?.nickname || item.author?.username }}</template>
                   <template #header-extra><span class="cm-time">{{ formatDate(item.created_at) }}</span></template>
                   <template #description>
                      <div class="rich-content w-e-text-container comment-content-html">
                         <div v-html="processContent(item.content)"></div>
                      </div>
                   </template>
                   <template #action>
                      <n-space>
                        <n-button size="tiny" text @click="handleReplyClick(item)">
                           <template #icon><n-icon><ChatbubbleEllipsesOutline /></n-icon></template> 回复
                        </n-button>
                        <n-button size="tiny" text @click="openReport(item.id, 'comment')">举报</n-button>
                        <n-popconfirm v-if="userStore.role === 'admin' || userStore.id === item.author_id" @positive-click="deleteComment(item.id)">
                           <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
                           删除
                        </n-popconfirm>
                      </n-space>
                   </template>
                </n-thing>

                <div v-if="item.children && item.children.length > 0" class="sub-comments">
                   <n-thing v-for="sub in item.children" :key="sub.id" class="comment-item sub-item">
                      <template #avatar><n-avatar round size="tiny" :src="getAvatar(sub.author?.avatar)" /></template>
                      <template #header>
                         {{ sub.author?.nickname || sub.author?.username }} 
                         <span style="color:#999;font-weight:normal;"> 回复 </span>
                         {{ item.author?.nickname || item.author?.username }}
                      </template>
                      <template #header-extra><span class="cm-time">{{ formatDate(sub.created_at) }}</span></template>
                      <template #description>
                         <div class="rich-content w-e-text-container comment-content-html">
                            <div v-html="processContent(sub.content)"></div>
                         </div>
                      </template>
                      <template #action>
                         <n-space>
                           <n-button size="tiny" text @click="handleReplyClick(sub)">回复</n-button>
                           <n-button size="tiny" text @click="openReport(sub.id, 'comment')">举报</n-button>
                           <n-popconfirm v-if="userStore.role === 'admin' || userStore.id === sub.author_id" @positive-click="deleteComment(sub.id)">
                              <template #trigger><n-button size="tiny" text type="error">删除</n-button></template>
                              删除
                           </n-popconfirm>
                         </n-space>
                      </template>
                   </n-thing>
                </div>
             </div>
          </div>
        </div>

      </n-card>

      <div v-else class="empty-state">
        内容不存在或已被删除
      </div>
    </div>

    <n-modal v-model:show="showReportModal" preset="dialog" title="我要举报">
       <n-radio-group v-model:value="reportReason" name="radiogroup">
          <n-space vertical>
            <n-radio value="垃圾广告">垃圾广告</n-radio>
            <n-radio value="违规内容">违规/色情/暴力内容</n-radio>
            <n-radio value="恶意攻击">恶意攻击/谩骂</n-radio>
            <n-radio value="其他">其他</n-radio>
          </n-space>
       </n-radio-group>
       <template #action>
          <n-button @click="showReportModal = false">取消</n-button>
          <n-button type="primary" :loading="reportLoading" @click="submitReport">提交</n-button>
       </template>
    </n-modal>
  </div>
</template>

<style scoped>
.post-detail-container { max-width: 900px; margin: 0 auto; padding: 24px; }
.post-card { border-radius: 12px; min-height: 80vh; }
.post-header-area { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.post-title { margin: 0; font-size: 28px; color: #1f2937; line-height: 1.4; flex: 1; }

.post-meta { display: flex; justify-content: space-between; align-items: center; color: #6b7280; font-size: 13px; }
.meta-left { display: flex; align-items: center; gap: 16px; }
.author-info { display: flex; align-items: center; gap: 6px; font-weight: 500; color: #374151; }
.time { display: flex; align-items: center; gap: 4px; }

.rich-content { font-size: 16px; line-height: 1.8; color: #374151; }
.rich-content :deep(img) { max-width: 100%; height: auto; border-radius: 8px; margin: 10px 0; cursor: pointer; }
.rich-content :deep(blockquote) { border-left: 4px solid #ccc; background-color: #f8f8f8; margin: 10px 0; padding: 10px; }

.comment-section { margin-top: 40px; }
.comment-editor-box { margin-bottom: 30px; position: relative; z-index: 100; }
.ci-actions { margin-top: 10px; text-align: right; }

.editor-wrapper {
  border: 1px solid #ccc; 
  border-radius: 4px; 
  /* 务必不要加 overflow: hidden，否则弹窗会被切掉 */
}

/* 🔥🔥🔥 核心修复：强制表情面板支持滚动 🔥🔥🔥 */
:deep(.w-e-panel-content-emotion-list) {
    height: 200px !important;       /* 强制固定高度 */
    overflow-y: auto !important;    /* 强制开启纵向滚动条 */
    padding: 10px;                  /* 加点内边距更好看 */
}

.comment-list { display: flex; flex-direction: column; gap: 24px; }
.root-comment { border-bottom: 1px solid #f3f4f6; padding-bottom: 16px; }
.sub-comments { margin-left: 48px; margin-top: 12px; background: #f9fafb; padding: 12px; border-radius: 8px; display: flex; flex-direction: column; gap: 12px;}
.sub-item { border-bottom: 1px dashed #eee; padding-bottom: 8px; }
.sub-item:last-child { border-bottom: none; padding-bottom: 0; }

.comment-content-html { background: transparent !important; border: none !important; padding: 0 !important; min-height: auto !important; font-size: 14px; margin-top: 4px; }
.cm-time { font-size: 12px; color: #9ca3af; }

@media (max-width: 700px) {
  .post-detail-container { padding: 16px; }
  .sub-comments { margin-left: 20px; }
}
</style>