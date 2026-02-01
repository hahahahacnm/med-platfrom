<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { 
  NCard, NForm, NFormItem, NInput, NSelect, NButton, NUpload, NIcon, 
  useMessage, NDivider, NList, NListItem, NThing, NTag, NImage, NImageGroup,
  NEmpty, NSpin, NTimeline, NTimelineItem
} from 'naive-ui'
import { 
  CloudUploadOutline, PaperPlaneOutline, ChatboxEllipsesOutline, 
  ImagesOutline, AlertCircleOutline, CheckmarkCircleOutline, 
  TimeOutline, CloseCircleOutline
} from '@vicons/ionicons5'
import request from '../utils/request' // 请根据实际路径调整

const message = useMessage()

// =======================
// 1. 常量定义
// =======================
const MAX_IMAGES = 4 // 🔥 限制最多上传 4 张
const TYPE_OPTIONS = [
  { label: '🐛 功能异常 (Bug)', value: '功能异常' },
  { label: '💡 产品建议', value: '产品建议' },
  { label: '💳 充值/账号问题', value: '账号问题' },
  { label: '🚮 题库内容报错', value: '内容报错' }, // 虽然有专门的纠错，这里也可以留一个入口
  { label: '📝 其他', value: '其他' }
]

// =======================
// 2. 提交表单逻辑
// =======================
const submitLoading = ref(false)
const form = reactive({
  type: null as string | null,
  content: '',
  contact: '',
  images: [] as string[] // 存储图片URL字符串
})

// 文件列表用于 UI 显示 (Naive UI Upload 组件需要的数据结构)
const fileList = ref<any[]>([]) 

// 自定义上传处理
const handleCustomRequest = async ({ file, onFinish, onError }: any) => {
  if (form.images.length >= MAX_IMAGES) {
    message.warning(`最多只能上传 ${MAX_IMAGES} 张图片`)
    onError()
    return
  }

  const formData = new FormData()
  formData.append('file', file.file)

  try {
    // 🔥 复用通用的上传接口 (请确保后端 router 里有 /notes/upload 或 /common/upload)
    const res: any = await request.post('/notes/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    
    // 成功后，将 URL 存入 form.images
    form.images.push(res.url)
    message.success('图片上传成功')
    onFinish()
  } catch (e) {
    message.error('上传失败')
    onError()
  }
}

// 删除图片时的同步处理
const handleRemove = (data: { file: any, fileList: any[] }) => {
  // Naive UI 的 fileList index 和 form.images 的 index 是一一对应的
  const index = fileList.value.findIndex(f => f.id === data.file.id)
  if (index !== -1) {
    form.images.splice(index, 1)
  }
  return true
}

const submitFeedback = async () => {
  if (!form.type) return message.warning('请选择反馈类型')
  if (!form.content.trim()) return message.warning('请描述具体情况')

  submitLoading.value = true
  try {
    await request.post('/platform-feedback', {
      type: form.type,
      content: form.content,
      contact: form.contact,
      images: form.images
    })
    message.success('反馈提交成功，感谢您的声音！')
    
    // 重置表单
    form.content = ''
    form.type = null
    form.images = []
    fileList.value = [] // 清空上传组件视图
    
    // 刷新列表
    fetchHistory()
  } catch (e: any) {
    message.error(e.response?.data?.error || '提交失败')
  } finally {
    submitLoading.value = false
  }
}

// =======================
// 3. 历史记录逻辑
// =======================
const historyLoading = ref(false)
const historyList = ref<any[]>([])

const fetchHistory = async () => {
  historyLoading.value = true
  try {
    const res: any = await request.get('/platform-feedback', { params: { page: 1, page_size: 50 } })
    historyList.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    historyLoading.value = false
  }
}

// 辅助：解析 JSON 图片数组
const parseImages = (jsonStr: any) => {
  try {
    // 如果后端存的是 datatype.JSON，这里可能已经是数组了，或者需要 JSON.parse
    if (Array.isArray(jsonStr)) return jsonStr
    return JSON.parse(jsonStr) || []
  } catch {
    return []
  }
}

// 辅助：状态样式
const getStatusTag = (status: number) => {
  switch (status) {
    case 0: return { type: 'default', text: '⏳ 待处理', icon: TimeOutline }
    case 1: return { type: 'info', text: '🏃 处理中', icon: TimeOutline }
    case 2: return { type: 'success', text: '✅ 已解决', icon: CheckmarkCircleOutline }
    case 3: return { type: 'error', text: '🚫 已驳回', icon: CloseCircleOutline }
    default: return { type: 'default', text: '未知', icon: AlertCircleOutline }
  }
}

const getFullUrl = (path: string) => path.startsWith('http') ? path : `http://localhost:8080${path}`

onMounted(() => {
  fetchHistory()
})
</script>

<template>
  <div class="feedback-page">
    <div class="page-container">
      
      <div class="column form-column">
        <div class="header-box">
          <h2><n-icon color="#2080f0"><ChatboxEllipsesOutline /></n-icon> 意见反馈</h2>
          <p>您的每一条建议都是我们进步的阶梯。</p>
        </div>

        <n-card :bordered="false" class="form-card">
          <n-form size="large">
            <n-form-item label="反馈类型" path="type" required>
              <n-select v-model:value="form.type" :options="TYPE_OPTIONS" placeholder="请选择问题类型" />
            </n-form-item>

            <n-form-item label="详细描述" path="content" required>
              <n-input 
                v-model:value="form.content" 
                type="textarea" 
                placeholder="请详细描述您遇到的问题或建议，如果是Bug请提供复现步骤..." 
                :autosize="{ minRows: 4, maxRows: 8 }" 
              />
            </n-form-item>

            <n-form-item label="图片凭证 (选填)">
              <div class="upload-wrapper">
                <n-upload
                  v-model:file-list="fileList"
                  list-type="image-card"
                  :custom-request="handleCustomRequest"
                  :on-remove="handleRemove"
                  :max="MAX_IMAGES"
                  accept="image/png,image/jpeg,image/gif"
                >
                  <div class="upload-trigger">
                    <n-icon size="24" color="#999"><ImagesOutline /></n-icon>
                    <span class="upload-text">{{ fileList.length }}/{{ MAX_IMAGES }}</span>
                  </div>
                </n-upload>
                <div class="tip">提供截图能帮我们更快定位问题 (最多{{ MAX_IMAGES }}张)</div>
              </div>
            </n-form-item>

            <n-form-item label="联系方式 (选填)">
              <n-input v-model:value="form.contact" placeholder="QQ / 微信 / 邮箱，方便我们联系您" />
            </n-form-item>

            <n-button type="primary" block size="large" @click="submitFeedback" :loading="submitLoading" class="submit-btn">
              <template #icon><n-icon><PaperPlaneOutline /></n-icon></template>
              提交反馈
            </n-button>
          </n-form>
        </n-card>
      </div>

      <div class="column list-column">
        <h3 class="list-title">反馈进度</h3>
        
        <div class="history-scroll">
          <n-spin :show="historyLoading">
            <div v-if="historyList.length === 0 && !historyLoading" class="empty-state">
              <n-empty description="暂无反馈记录，您现在就可以去提一个！" />
            </div>

            <div v-else class="feed-list">
              <div v-for="item in historyList" :key="item.id" class="feed-item">
                <div class="feed-header">
                  <n-tag size="small" :bordered="false" type="info" class="type-tag">{{ item.type }}</n-tag>
                  <n-tag size="small" :bordered="false" :type="getStatusTag(item.status).type as any">
                    {{ getStatusTag(item.status).text }}
                  </n-tag>
                </div>
                
                <div class="feed-content">{{ item.content }}</div>
                
                <div v-if="parseImages(item.images).length > 0" class="feed-imgs">
                  <n-image-group>
                    <n-space>
                      <n-image 
                        v-for="(img, idx) in parseImages(item.images)" 
                        :key="idx" 
                        :src="getFullUrl(img)" 
                        width="60" 
                        height="60" 
                        object-fit="cover" 
                        class="thumb"
                      />
                    </n-space>
                  </n-image-group>
                </div>

                <div class="feed-meta">
                  <span class="time">{{ new Date(item.created_at).toLocaleString() }}</span>
                </div>

                <div v-if="item.admin_reply" class="admin-reply-box">
                  <div class="reply-head">
                    <n-icon color="#18a058"><ChatboxEllipsesOutline /></n-icon> 管理员回复：
                  </div>
                  <div class="reply-text">{{ item.admin_reply }}</div>
                </div>
              </div>
            </div>
          </n-spin>
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
.feedback-page {
  padding: 24px;
  background-color: #f8fafc;
  min-height: calc(100vh - 64px);
}

.page-container {
  max-width: 1100px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1.2fr 1fr; /* 左宽右窄 */
  gap: 32px;
  align-items: start;
}

/* 左侧样式 */
.header-box h2 {
  display: flex; align-items: center; gap: 8px; margin: 0 0 8px 0; color: #334155;
}
.header-box p { color: #64748b; margin: 0 0 24px 0; font-size: 14px; }

.form-card {
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
}

.upload-wrapper { width: 100%; }
.upload-trigger {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  width: 100%; height: 100%;
}
.upload-text { font-size: 12px; color: #999; margin-top: 4px; }
.tip { font-size: 12px; color: #94a3b8; margin-top: 8px; }

.submit-btn {
  margin-top: 12px;
  font-weight: bold;
  border-radius: 8px;
}

/* 右侧样式 */
.list-title { margin: 0 0 16px 0; color: #334155; font-size: 18px; }

.feed-list { display: flex; flex-direction: column; gap: 16px; }

.feed-item {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #eef2f6;
  transition: all 0.2s;
}
.feed-item:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.05); }

.feed-header { display: flex; justify-content: space-between; margin-bottom: 10px; }
.type-tag { font-weight: bold; }

.feed-content { font-size: 14px; color: #334155; line-height: 1.6; white-space: pre-wrap; margin-bottom: 10px; }

.feed-imgs { margin-bottom: 10px; }
.thumb { border-radius: 6px; border: 1px solid #eee; }

.feed-meta { font-size: 12px; color: #cbd5e1; text-align: right; }

.admin-reply-box {
  margin-top: 12px;
  background: #f0fdf4; /* 浅绿色背景 */
  border-left: 3px solid #18a058;
  padding: 10px 12px;
  border-radius: 0 6px 6px 0;
}
.reply-head { font-size: 13px; font-weight: bold; color: #166534; display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.reply-text { font-size: 13px; color: #15803d; line-height: 1.5; }

.empty-state { padding: 40px 0; }

/* 响应式 */
@media (max-width: 800px) {
  .page-container { grid-template-columns: 1fr; }
  .list-column { margin-top: 20px; border-top: 1px solid #eee; padding-top: 24px; }
}
</style>