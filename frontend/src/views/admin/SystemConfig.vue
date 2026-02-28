<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { 
  NCard, NForm, NFormItem, NInput, NButton, NSpace, 
  useMessage, NPageHeader, NAlert, NDivider, NGrid, NGi, NIcon, NSpin,
  NTooltip, NInputNumber
} from 'naive-ui'
import { SettingsOutline, SendOutline, SaveOutline, InformationCircleOutline } from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()
const loading = ref(false)
const testing = ref(false)
const testEmail = ref('')

// 🔥 核心改动 1：配置项模型中引入分润比例参数
const configs = reactive({
  SMTP_HOST: '',
  SMTP_PORT: '465',
  SMTP_USER: '',
  SMTP_PASS: '',
  SMTP_SENDER_NAME: '', 
  FRONTEND_URL: window.location.origin,
  // 新增代理分润变量，默认给个兜底值字符串
  AGENT_COMMISSION_RATE_DIRECT: '0.20', 
  AGENT_COMMISSION_RATE_CARD: '0.15'
})

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/configs')
    if (res.data) {
      res.data.forEach((item: any) => {
        if (Object.keys(configs).includes(item.key)) {
          (configs as any)[item.key] = item.value
        }
      })
    }
  } catch (e) {
    message.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async (key: string, desc: string) => {
  try {
    await request.post('/admin/configs', {
      key: key,
      value: String((configs as any)[key]), // 确保传给后端的是字符串
      description: desc
    })
  } catch (e) {
    throw new Error(`保存${key}失败`)
  }
}

const handleSaveAll = async () => {
  loading.value = true
  try {
    // 保存邮件配置
    await saveConfig('SMTP_HOST', 'SMTP服务器地址')
    await saveConfig('SMTP_PORT', '端口(465/587)')
    await saveConfig('SMTP_USER', '发信邮箱账号')
    await saveConfig('SMTP_PASS', '发信授权码')
    await saveConfig('SMTP_SENDER_NAME', '发件人昵称')
    await saveConfig('FRONTEND_URL', '前端访问地址')
    
    // 🔥 核心改动 2：保存分润配置
    await saveConfig('AGENT_COMMISSION_RATE_DIRECT', '在线支付分润比例 (0~1之间)')
    await saveConfig('AGENT_COMMISSION_RATE_CARD', '卡密兑换分润比例 (0~1之间)')

    message.success('所有配置已同步至内存并实时生效！')
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

const handleTestEmail = async () => {
  // ... 保持原有发送测试邮件逻辑不变 ...
  if (!testEmail.value) return message.warning('请输入接收测试的邮箱')
  testing.value = true
  try {
    await request.post('/admin/configs/test-email', { target_email: testEmail.value })
    message.success('测试邮件已发出，请注意查收')
  } catch (e: any) {
    message.error(e.response?.data?.error || '发送失败')
  } finally {
    testing.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="sys-config-container">
    <n-page-header title="⚙️ 平台参数管理" subtitle="管理系统全局变量、密钥及第三方接口配置" />
    
    <n-spin :show="loading">
      <n-grid cols="1 s:1 m:1 l:2" responsive="screen" :x-gap="20" :y-gap="20" style="margin-top: 24px;">
        
        <n-gi>
          <n-card title="📧 邮件服务 (SMTP)" :segmented="{ content: true }" size="small">
             <n-form label-placement="top" size="medium">
              <n-grid :cols="2" :x-gap="12">
                <n-gi><n-form-item label="SMTP地址"><n-input v-model:value="configs.SMTP_HOST" placeholder="如: smtp.qq.com" /></n-form-item></n-gi>
                <n-gi><n-form-item label="SMTP端口"><n-input v-model:value="configs.SMTP_PORT" placeholder="465 / 587" /></n-form-item></n-gi>
              </n-grid>
              <n-form-item label="发信邮箱"><n-input v-model:value="configs.SMTP_USER" placeholder="用于发送通知的邮箱账号" /></n-form-item>
              <n-form-item label="发件人昵称"><n-input v-model:value="configs.SMTP_SENDER_NAME" placeholder="收件箱里显示的名称" /></n-form-item>
              <n-form-item label="授权码/密码"><n-input v-model:value="configs.SMTP_PASS" type="password" show-password-on="click" /></n-form-item>
              <n-divider title-placement="left" style="margin: 12px 0">基础链路</n-divider>
              <n-form-item label="前端访问 URL"><n-input v-model:value="configs.FRONTEND_URL" placeholder="http://domain.com" /></n-form-item>
            </n-form>
          </n-card>
        </n-gi>

        <n-gi>
          <n-space vertical :size="20">
            
            <n-card title="💰 财务与代理分润配置" size="small">
              <n-alert type="warning" :show-icon="false" style="margin-bottom: 16px;">
                提示：分润比例请填写小数，例如填写 <strong>0.15</strong> 代表 <strong>15%</strong>。此修改对下一笔订单立即生效！
              </n-alert>
              <n-form label-placement="left" label-width="140">
                <n-form-item label="直充支付分润比例">
                  <n-input v-model:value="configs.AGENT_COMMISSION_RATE_DIRECT" placeholder="例如：0.20" />
                </n-form-item>
                <n-form-item label="卡密兑换分润比例">
                  <n-input v-model:value="configs.AGENT_COMMISSION_RATE_CARD" placeholder="例如：0.15" />
                </n-form-item>
              </n-form>
            </n-card>

            <n-card title="🧪 发信测试" size="small">
              <n-form label-placement="top">
                <n-form-item label="测试收件邮箱"><n-input v-model:value="testEmail" placeholder="输入您的常用邮箱" /></n-form-item>
                <n-button :loading="testing" type="success" secondary block @click="handleTestEmail">
                  <template #icon><n-icon><SendOutline /></n-icon></template>发送测试邮件
                </n-button>
              </n-form>
            </n-card>

          </n-space>
        </n-gi>
      </n-grid>

      <div style="margin-top: 24px; text-align: center;">
        <n-button type="primary" size="large" @click="handleSaveAll" :loading="loading" style="width: 200px;">
          <template #icon><n-icon><SaveOutline /></n-icon></template>
          保存所有系统配置
        </n-button>
      </div>
    </n-spin>
  </div>
</template>

<style scoped>
.sys-config-container {
  padding: 24px;
  background-color: #f9fbfe;
  min-height: 100vh;
}

.guide-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.guide-item strong {
  color: #333;
  font-size: 14px;
  display: block;
  margin-bottom: 4px;
}

.guide-item p {
  color: #666;
  font-size: 13px;
  margin: 0;
  line-height: 1.6;
}

:deep(.n-card-header__title) {
  font-weight: bold;
}
</style>