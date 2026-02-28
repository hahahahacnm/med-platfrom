<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { 
  NCard, NSlider, NInputNumber, NStatistic, NGrid, NGi, NButton, 
  NAlert, NSpace, NTag, NDataTable, NIcon, useMessage, NPageHeader, NDivider,
  NTooltip
} from 'naive-ui'
import { 
  PricetagOutline, 
  SaveOutline,
  WalletOutline,
  CashOutline,
  TrendingUpOutline,
  RibbonOutline,
  InformationCircleOutline,
  TicketOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const invitationCode = ref('')

// === 比例配置 ===
const myRate = ref(0) 
const saving = ref(false)
const maxProfitRate = ref(20)      // 直购渠道最大提成 (变量)
const maxCardProfitRate = ref(15)  // 卡密渠道提成 (固定)

// === Admin 状态 ===
const agentList = ref([])
const adminLoading = ref(false)

const basePrice = 100

const sliderMarks = computed(() => ({
  0: '不打折',
  [maxProfitRate.value]: '0利润引流'
}))

// --- 直购渠道模拟 ---
const userPay = computed(() => Number((basePrice * (1 - myRate.value / 100)).toFixed(2)))
const userSave = computed(() => Number((basePrice * (myRate.value / 100)).toFixed(2)))
const myProfit = computed(() => {
  const maxProfit = basePrice * (maxProfitRate.value / 100)
  return Number((maxProfit - userSave.value).toFixed(2))
})

// --- 卡密渠道模拟 (固定) ---
const cardProfit = computed(() => Number((basePrice * (maxCardProfitRate.value / 100)).toFixed(2)))

// Admin 列定义
const columns = computed(() => [
  { title: 'ID', key: 'id', width: 60 },
  { title: '代理名称', key: 'nickname', width: 120 },
  { title: '邀请码', key: 'invitation_code', width: 120, 
    render: (row: any) => h(NTag, { type: 'warning', bordered: false }, { default: () => row.invitation_code }) 
  },
  { 
    title: '直购让利配置', key: 'agent_discount_rate', 
    render: (row: any) => {
      const rate = row.agent_discount_rate || 0
      return h('div', [
        h(NTag, { type: rate > (maxProfitRate.value / 2) ? 'error' : 'success', size: 'small' }, { default: () => `让利 ${rate}%` }),
        h('span', { style: 'font-size: 12px; color: #999; margin-left: 8px' }, `(直购自留 ${maxProfitRate.value - rate}%)`)
      ])
    }
  },
  { title: '加入时间', key: 'created_at', width: 180, 
    render: (row: any) => new Date(row.created_at).toLocaleDateString() 
  }
])

const initData = async () => {
  if (userStore.role === 'agent') {
    loading.value = true
    try {
      const res: any = await request.get('/user/profile')
      // 适配后端传入的两种比例
      if (res.data.global_profit_rate) maxProfitRate.value = Math.round(res.data.global_profit_rate * 100)
      if (res.data.card_profit_rate) maxCardProfitRate.value = Math.round(res.data.card_profit_rate * 100)
      
      myRate.value = res.data.agent_discount_rate || 0
      invitationCode.value = res.data.invitation_code || ''
    } finally {
      loading.value = false
    }
  } else if (userStore.role === 'admin') {
    adminLoading.value = true
    try {
      const confRes: any = await request.get('/admin/configs')
      // Admin 同时获取两项配置进行预览计算
      const rateConf = confRes.data?.find((c: any) => c.key === 'AGENT_COMMISSION_RATE_DIRECT')
      const cardConf = confRes.data?.find((c: any) => c.key === 'AGENT_COMMISSION_RATE_CARD')
      if (rateConf) maxProfitRate.value = Math.round(parseFloat(rateConf.value) * 100)
      if (cardConf) maxCardProfitRate.value = Math.round(parseFloat(cardConf.value) * 100)

      const res: any = await request.get('/admin/users', { params: { role: 'agent', page_size: 100 } })
      agentList.value = res.data || [] 
    } finally {
      adminLoading.value = false
    }
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    await request.put('/user/profile', { agent_discount_rate: Math.floor(myRate.value) })
    message.success('优惠策略已更新')
  } catch (e) {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(initData)
</script>

<template>
  <div class="discount-page">
    <n-page-header title="💎 推广与分润体系" subtitle="全渠道分润透明化，直购让利自定义" style="margin-bottom: 24px;" />
    
    <div v-if="userStore.role === 'agent'">
      <div class="glass-panel">
        <div class="panel-header">
          <div class="header-title">
            <div class="icon-wrapper"><n-icon size="24" color="#fff"><RibbonOutline/></n-icon></div>
            <span class="title-text">分润策略中枢</span>
          </div>
          <div class="header-extra">
            <span class="invite-label">我的邀请码：</span>
            <n-tag type="info" size="large" round bordered style="font-size: 16px; font-weight: bold;">{{ invitationCode || '...' }}</n-tag>
          </div>
        </div>

        <n-alert type="info" show-icon class="custom-alert">
          <template #header>渠道分润规则声明</template>
          1. <strong>直购渠道：</strong> 平台释放 {{ maxProfitRate }}% 利润。您可以自定义分配给用户的折扣，余下为您的收益。<br/>
          2. <strong>卡密渠道：</strong> 考虑到发卡平台费率，平台为您提供固定 <strong>{{ maxCardProfitRate }}%</strong> 提成。卡密不参与系统内让利。
        </n-alert>

        <div class="slider-container">
          <div class="slider-label">
            <n-icon size="22" color="#2563eb"><PricetagOutline/></n-icon>
            <span class="label-text">设置直购让利比例 (0% - {{ maxProfitRate }}%)</span>
            <n-tooltip trigger="hover">
              <template #trigger><n-icon size="18" color="#94a3b8"><InformationCircleOutline /></n-icon></template>
              当用户通过邀请链接在线支付时，将自动享受此折扣。
            </n-tooltip>
          </div>
          
          <n-grid cols="1 s:12" responsive="screen" x-gap="24" y-gap="16" style="align-items: center;">
            <n-gi span="9">
              <n-slider v-model:value="myRate" :min="0" :max="maxProfitRate" :step="1" :marks="sliderMarks" class="custom-slider" />
            </n-gi>
            <n-gi span="3">
              <n-input-number v-model:value="myRate" size="large" :min="0" :max="maxProfitRate" :precision="0" button-placement="both" />
            </n-gi>
          </n-grid>
        </div>

        <n-divider dashed>多渠道收益对比 (以 100 元商品为例)</n-divider>

        <n-grid cols="1 s:2" responsive="screen" x-gap="20" y-gap="20">
          <n-gi>
            <div class="channel-card direct">
              <div class="channel-tag">在线直充 (含让利)</div>
              <n-grid cols="2">
                <n-gi><n-statistic label="用户支付" :value="userPay"><template #prefix>¥</template></n-statistic></n-gi>
                <n-gi><n-statistic label="您的净利" :value="myProfit" :value-style="{color:'#e11d48', fontWeight:'bold'}"><template #prefix>¥</template></n-statistic></n-gi>
              </n-grid>
              <div class="channel-footer">用户享受 {{ 100-myRate }} 折优惠</div>
            </div>
          </n-gi>
          <n-gi>
            <div class="channel-card card">
              <div class="channel-tag">卡密/兑换码 (固定)</div>
              <n-grid cols="2">
                <n-gi><n-statistic label="用户支付" :value="100"><template #prefix>¥</template></n-statistic></n-gi>
                <n-gi><n-statistic label="您的净利" :value="cardProfit" :value-style="{color:'#2563eb', fontWeight:'bold'}"><template #prefix>¥</template></n-statistic></n-gi>
              </n-grid>
              <div class="channel-footer">用户按外部平台价格购买</div>
            </div>
          </n-gi>
        </n-grid>

        <div class="action-footer">
          <n-button type="primary" size="large" @click="handleSave" :loading="saving" class="save-btn">
            <template #icon><n-icon><SaveOutline /></n-icon></template>
            同步让利策略
          </n-button>
          <p class="action-hint">卡密渠道提成由系统自动发放，无需额外配置</p>
        </div>
      </div>
    </div>

    <div v-else-if="userStore.role === 'admin'">
      <n-card title="🕵️ 全平台代理分润监控" :bordered="false" class="admin-card">
        <n-data-table :columns="columns" :data="agentList" :loading="adminLoading" :pagination="{ pageSize: 12 }" striped />
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.discount-page { padding: 24px; background-color: #f8fafc; min-height: calc(100vh - 60px); }
.glass-panel { max-width: 900px; margin: 0 auto; background: #fff; border-radius: 16px; box-shadow: 0 4px 24px rgba(0,0,0,0.05); padding: 32px; border: 1px solid #e2e8f0; }
.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.icon-wrapper { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); width: 42px; height: 42px; border-radius: 12px; display: flex; align-items: center; justify-content: center; }
.title-text { font-size: 20px; font-weight: 800; margin-left: 12px; }
.custom-alert { margin-bottom: 32px; border-radius: 12px; }

/* 渠道卡片样式 */
.channel-card { padding: 24px; border-radius: 16px; position: relative; border: 1px solid #f1f5f9; transition: transform 0.3s ease; }
.channel-card:hover { transform: translateY(-5px); }
.channel-card.direct { background: linear-gradient(145deg, #fff1f2 0%, #ffffff 100%); border-color: #ffe4e6; }
.channel-card.card { background: linear-gradient(145deg, #f0f9ff 0%, #ffffff 100%); border-color: #e0f2fe; }
.channel-tag { position: absolute; top: -12px; left: 20px; background: #1e293b; color: #fff; padding: 2px 12px; border-radius: 100px; font-size: 12px; }
.channel-footer { margin-top: 16px; font-size: 12px; color: #94a3b8; text-align: center; border-top: 1px dashed #e2e8f0; padding-top: 12px; }

.slider-container { background: #f8fafc; padding: 24px; border-radius: 12px; margin-bottom: 24px; }
.action-footer { margin-top: 40px; text-align: center; }
.save-btn { width: 240px; border-radius: 100px; font-weight: bold; }
</style>