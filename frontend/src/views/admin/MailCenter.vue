<script setup lang="ts">
import { ref, reactive, shallowRef, onBeforeUnmount, h } from 'vue'
import { 
  NPageHeader, NCard, NForm, NFormItem, NInput, NSelect, NSwitch, 
  NButton, NSpace, NIcon, NAlert, useMessage, NModal, NTag, NSpin, NDivider
} from 'naive-ui'
import { 
  PaperPlaneOutline, 
  SearchOutline, 
  EyeOutline, 
  AlertCircleOutline,
  PawOutline // 🐾 借用类似爪子的图标
} from '@vicons/ionicons5'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

// 引入 WangEditor
import '@wangeditor/editor/dist/css/style.css' 
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

const message = useMessage()
const userStore = useUserStore()

// =============================================================
// 表单数据与状态
// =============================================================
const formRef = ref()
const mailModel = reactive({
  target_type: 'specific', // 'specific' | 'all'
  user_ids: [],
  subject: '',
  content: '<p>在这里写下你的内容...</p>'
})
const sending = ref(false)

// =============================================================
// 用户搜索与选择逻辑
// =============================================================
const searchLoading = ref(false)
const userOptions = ref<Array<{label: string, value: number, disabled?: boolean}>>([])

// 防抖搜索用户
let searchTimer: any = null
const handleSearchUsers = (query: string) => {
  if (!query) {
    userOptions.value = []
    return
  }
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(async () => {
    searchLoading.value = true
    try {
      const res: any = await request.get('/admin/emails/users', { params: { q: query } })
      userOptions.value = res.data.map((u: any) => ({
        label: `${u.nickname || u.username} (${u.email})`,
        value: u.id
      }))
    } catch (e) {
      console.error(e)
    } finally {
      searchLoading.value = false
    }
  }, 300) // 300ms 防抖
}

// 首次加载时拉取一些默认用户展示
const fetchInitialUsers = async () => {
  try {
    const res: any = await request.get('/admin/emails/users')
    userOptions.value = res.data.map((u: any) => ({
      label: `${u.nickname || u.username} (${u.email})`,
      value: u.id
    }))
  } catch (e) {}
}
fetchInitialUsers()

// =============================================================
// WangEditor 配置
// =============================================================
const editorRef = shallowRef()
const mode = 'default'
const toolbarConfig = { excludeKeys: ['group-video'] } 
const editorConfig = { 
  placeholder: '请输入邮件正文...',
  MENU_CONF: {
    uploadImage: {
      server: '/api/v1/forum/upload', // 复用论坛的图片上传接口
      fieldName: 'file',
      maxFileSize: 5 * 1024 * 1024,
      headers: { Authorization: `Bearer ${userStore.token}` },
      customInsert(res: any, insertFn: any) {
        if (res.url) { 
          const fullUrl = res.url.startsWith('http') ? res.url : `http://localhost:8080${res.url}`
          insertFn(fullUrl, '', '') 
        } else { 
          message.error('图片上传失败') 
        }
      }
    }
  }
}
onBeforeUnmount(() => { const editor = editorRef.value; if (editor) editor.destroy() })
const handleCreated = (editor: any) => { editorRef.value = editor }

// =============================================================
// 发送与预览逻辑
// =============================================================
const showPreview = ref(false)

const handleSend = async () => {
  if (!mailModel.subject) return message.warning('请填写邮件标题')
  if (mailModel.target_type === 'specific' && mailModel.user_ids.length === 0) {
    return message.warning('请至少选择一位收件人')
  }

  sending.value = true
  try {
    const res: any = await request.post('/admin/emails/send', mailModel)
    message.success(res.message || '发信任务已提交')
    // 发送成功后清空部分表单
    mailModel.subject = ''
    mailModel.content = '<p></p>'
    mailModel.user_ids = []
  } catch (e: any) {
    message.error(e.response?.data?.error || '任务下发失败')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="mail-center-page">
    <n-page-header title="🐾 邮件营销与通知中心" subtitle="基于异步队列的高性能群发系统" style="margin-bottom: 24px;" />

    <div class="glass-panel">
      <div class="top-decoration"></div>

      <n-form ref="formRef" :model="mailModel" size="large" label-placement="top">
        
        <div class="section-title">收件人设置</div>
        <n-form-item label="发送模式">
          <n-switch v-model:value="mailModel.target_type" checked-value="all" unchecked-value="specific" size="large">
            <template #checked>全员群发模式 (All Users)</template>
            <template #unchecked>精准推送模式 (Specific Users)</template>
          </n-switch>
        </n-form-item>

        <n-alert v-if="mailModel.target_type === 'all'" type="warning" show-icon class="mt-mb">
          <template #icon><n-icon><AlertCircleOutline /></n-icon></template>
          <strong>高能预警：</strong> 您当前开启了全站群发模式！邮件将发送给所有已绑定邮箱的用户。系统将自动采用缓冲队列发送，以防止 SMTP 服务器封停。
        </n-alert>

        <n-form-item v-if="mailModel.target_type === 'specific'" label="选择目标用户">
          <n-select
            v-model:value="mailModel.user_ids"
            multiple
            filterable
            remote
            :options="userOptions"
            :loading="searchLoading"
            placeholder="输入用户名、昵称或邮箱进行搜索并选中..."
            @search="handleSearchUsers"
            clearable
          >
            <template #action>
              <span style="color: #94a3b8; font-size: 12px;">提示：只能搜索到已绑定邮箱的用户哦 🐾</span>
            </template>
          </n-select>
        </n-form-item>

        <n-divider dashed />

        <div class="section-title">信件内容撰写</div>
        <n-form-item label="邮件标题 (Subject)">
          <n-input v-model:value="mailModel.subject" placeholder="例如：平台十一月重大更新通知 🚀" />
        </n-form-item>

        <n-form-item label="邮件正文 (Rich Text)">
          <div class="editor-wrapper">
            <Toolbar style="border-bottom: 1px solid #e2e8f0; background: #f8fafc;" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
            <Editor style="height: 400px; overflow-y: hidden;" v-model="mailModel.content" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
          </div>
        </n-form-item>

        <div class="action-footer">
          <n-button size="large" type="info" secondary @click="showPreview = true" style="width: 150px; margin-right: 20px;">
            <template #icon><n-icon><EyeOutline /></n-icon></template>
            实景预览
          </n-button>
          
          <n-button size="large" type="primary" :loading="sending" @click="handleSend" class="send-btn">
            <template #icon><n-icon><PaperPlaneOutline /></n-icon></template>
            确认并投递邮件
          </n-button>
        </div>
      </n-form>
    </div>

    <n-modal v-model:show="showPreview" transform-origin="center">
      <div class="preview-sandbox">
        <div class="email-container">
          <div class="header-title">{{ mailModel.subject || '（未填写标题）' }}</div>
          
          <div class="content">
            <div class="user-greeting">尊敬的 [用户昵称]，您好！🐾</div>
            <div class="html-content" v-html="mailModel.content"></div>
          </div>

          <div class="footer">
            由 <strong>平台安全系统</strong> 全力驱动<br/>
            感谢每一位支持本站的朋友！
            <span class="paw-footer">🐾 🐾 🐾</span>
          </div>
          <div class="paw-watermark">🐾</div>
        </div>
        
        <div style="text-align: center; margin-top: 20px;">
          <n-button type="primary" ghost @click="showPreview = false">关闭预览</n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.mail-center-page {
  padding: 24px;
  background-color: #f8fafc;
  min-height: calc(100vh - 60px);
}

/* 玻璃拟物主面板 */
.glass-panel {
  max-width: 1000px;
  margin: 0 auto;
  background: #ffffff;
  border-radius: 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
  border: 1px solid #e2e8f0;
  padding: 40px;
  position: relative;
  overflow: hidden;
}

/* 顶部福瑞渐变条 */
.top-decoration {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 8px;
  background: linear-gradient(to right, #ff9a9e, #fad0c4, #3b82f6);
}

.section-title {
  font-size: 18px;
  font-weight: bold;
  color: #1e293b;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}
.section-title::before {
  content: "🐾";
  margin-right: 8px;
  font-size: 20px;
}

.mt-mb { margin-top: 10px; margin-bottom: 24px; }

/* 编辑器容器 */
.editor-wrapper {
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  width: 100%;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
}

/* 底部按钮区 */
.action-footer {
  margin-top: 40px;
  display: flex;
  justify-content: center;
}
.send-btn {
  width: 250px;
  border-radius: 100px;
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  border: none;
  font-weight: bold;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
  transition: transform 0.2s;
}
.send-btn:active { transform: scale(0.96); }

/* ========================================================= */
/* 🐾 实景预览区域样式 (1:1 复刻后端的 custom_notice.html) */
/* ========================================================= */
.preview-sandbox {
  width: 650px;
  background: transparent;
}
.email-container {
  background: white url('https://q1.qlogo.cn/g?b=qq&nk=2219911811&s=640') top 24px right 24px no-repeat;
  background-size: 64px;
  border-radius: 24px;
  padding: 40px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.15);
  position: relative;
  overflow: hidden;
}
.email-container::before {
  content: ""; position: absolute; top: 0; left: 0; width: 100%; height: 8px;
  background: linear-gradient(to right, #ff9a9e, #fad0c4, #3b82f6);
}
.header-title {
  font-size: 24px; font-weight: bold; color: #1e293b; margin-bottom: 30px; 
  padding-right: 80px; border-bottom: 2px dashed #f1f5f9; padding-bottom: 15px;
}
.user-greeting { font-size: 18px; font-weight: 600; color: #1e293b; margin-bottom: 20px; }

/* 注入富文本的沙箱隔离 */
.html-content { font-size: 16px; color: #475569; min-height: 100px; line-height: 1.8; }
.html-content :deep(img) { max-width: 100%; height: auto; border-radius: 12px; margin: 15px 0; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.html-content :deep(a) { color: #3b82f6; text-decoration: none; border-bottom: 1px solid #3b82f6; }

.footer {
  margin-top: 50px; font-size: 13px; color: #94a3b8; text-align: center; 
  border-top: 1px solid #f1f5f9; padding-top: 25px; position: relative;
}
.paw-watermark {
  position: absolute; bottom: -15px; left: -10px; font-size: 90px; 
  opacity: 0.04; transform: rotate(15deg); color: #ff9a9e; pointer-events: none;
}
.paw-footer { font-size: 20px; opacity: 0.3; margin-top: 10px; display: block; }
</style>