<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { 
  NCard, NSlider, NInputNumber, NStatistic, NGrid, NGi, NButton, 
  NAlert, NSpace, NTag, NDataTable, NIcon, useMessage 
} from 'naive-ui'
import { 
  PricetagOutline, 
  SaveOutline 
} from '@vicons/ionicons5'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)

// === Agent 状态 ===
const myRate = ref(0) // 0 - 20
const saving = ref(false)

// === Admin 状态 ===
const agentList = ref([])
const adminLoading = ref(false)

// 模拟计算 (基于 100 元基准)
const basePrice = 100
const maxProfitRate = 20 // 平台最大释放 20%

// 精度控制
const userPay = computed(() => {
  const val = basePrice * (1 - myRate.value / 100)
  return Number(val.toFixed(2))
})

const userSave = computed(() => {
  const val = basePrice * (myRate.value / 100)
  return Number(val.toFixed(2))
})

const myProfit = computed(() => {
  const maxProfit = basePrice * (maxProfitRate / 100)
  const currentSave = basePrice * (myRate.value / 100)
  const val = maxProfit - currentSave
  return Number(val.toFixed(2))
})

// Admin 列定义
const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '代理名称', key: 'nickname', width: 120 },
  { title: '邀请码', key: 'invitation_code', width: 120, 
    render: (row: any) => h(NTag, { type: 'warning', bordered: false }, { default: () => row.invitation_code }) 
  },
  { 
    title: '当前让利配置', key: 'agent_discount_rate', 
    render: (row: any) => {
      const rate = row.agent_discount_rate || 0
      return h('div', [
        h(NTag, { type: rate > 10 ? 'error' : 'success' }, { default: () => `让利 ${rate}%` }),
        h('span', { style: 'font-size: 12px; color: #999; margin-left: 8px' }, `(自留利润 ${20 - rate}%)`)
      ])
    }
  },
  { title: '加入时间', key: 'created_at', width: 180, 
    render: (row: any) => new Date(row.created_at).toLocaleDateString() 
  }
]

// === 初始化 ===
const initData = async () => {
  if (userStore.role === 'agent') {
    // 代理：获取自己的配置
    loading.value = true
    try {
      const res: any = await request.get('/user/profile')
      
      // 1. 设置折扣
      myRate.value = res.data.agent_discount_rate || 0
      
      // 🔥 2. [修复] 同步邀请码，解决"加载中..."问题
      if (res.data.invitation_code) {
        userStore.invitationCode = res.data.invitation_code
      }
      
    } finally {
      loading.value = false
    }
  } else if (userStore.role === 'admin') {
    // 管理员逻辑
    adminLoading.value = true
    try {
      const res: any = await request.get('/admin/users', { params: { role: 'agent', page_size: 100 } })
      agentList.value = res.data || [] 
    } finally {
      adminLoading.value = false
    }
  }
}

// === Agent 保存 ===
const handleSave = async () => {
  saving.value = true
  try {
    await request.put('/user/profile', {
      agent_discount_rate: Math.floor(myRate.value) 
    })
    message.success('优惠策略已更新，新用户下单将即时生效')
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
    
    <div v-if="userStore.role === 'agent'">
      <n-card title="💰 优惠与利润配置" size="huge" :bordered="false" style="max-width: 800px; margin: 0 auto;">
        <template #header-extra>
          <n-tag type="info">我的邀请码：{{ userStore.invitationCode || '加载中...' }}</n-tag>
        </template>
        
        <n-alert type="info" show-icon title="策略说明" style="margin-bottom: 24px;">
          平台默认释放 <strong>20%</strong> 的利润空间给您。您可以自由决定将这 20% 分配多少给用户（作为折扣），剩下的即为您的净利润。
          <br>
          <i>例如：设置让利 5%，用户打 95 折，您拿 15% 提成。</i>
        </n-alert>

        <div class="control-panel">
          <div class="label">
            <n-icon size="20" color="#2080f0"><PricetagOutline/></n-icon>
            <span>设置给用户的折扣比例 (0% - 20%)</span>
          </div>
          <n-grid cols="12" x-gap="12" style="align-items: center; margin-top: 12px;">
            <n-gi span="8">
              <n-slider v-model:value="myRate" :min="0" :max="20" :step="1" :marks="{0:'不打折', 10:'让利10%', 20:'0利润引流'}" />
            </n-gi>
            <n-gi span="4">
              <n-input-number v-model:value="myRate" size="small" :min="0" :max="20" :precision="0">
                <template #suffix>%</template>
              </n-input-number>
            </n-gi>
          </n-grid>
        </div>

        <n-card embedded title="📊 利润模拟器 (以 100元 商品为例)" style="margin-top: 30px;">
          <n-grid cols="3" style="text-align: center;">
            <n-gi>
              <n-statistic label="用户实付" :value="userPay">
                <template #prefix>¥</template>
                <template #suffix><small style="font-size: 12px; color: #999">({{ 100 - myRate }}折)</small></template>
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="用户节省" :value="userSave" :value-style="{ color: '#18a058' }">
                <template #prefix>¥</template>
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="您的净利润" :value="myProfit" :value-style="{ color: '#d03050', fontWeight: 'bold' }">
                <template #prefix>¥</template>
              </n-statistic>
            </n-gi>
          </n-grid>
        </n-card>

        <div class="actions" style="margin-top: 30px; text-align: center;">
          <n-button type="primary" size="large" @click="handleSave" :loading="saving" style="width: 200px;">
            <template #icon><n-icon><SaveOutline /></n-icon></template>
            保存策略
          </n-button>
          <div style="margin-top: 10px; font-size: 12px; color: #999;">修改后即时生效</div>
        </div>
      </n-card>
    </div>

    <div v-else-if="userStore.role === 'admin'">
      <n-card title="🕵️ 代理定价监控" :bordered="false">
        <template #header-extra>
          <n-button size="small" @click="initData">刷新数据</n-button>
        </template>
        <n-data-table 
          :columns="columns" 
          :data="agentList" 
          :loading="adminLoading" 
          :pagination="{ pageSize: 10 }"
        />
      </n-card>
    </div>

    <div v-else style="text-align: center; margin-top: 50px; color: #999;">
      您没有权限访问此页面
    </div>

  </div>
</template>

<style scoped>
.discount-page { padding: 24px; }
.control-panel { background: #f9f9f9; padding: 20px; border-radius: 8px; border: 1px solid #eee; }
.label { display: flex; align-items: center; gap: 8px; font-weight: bold; color: #333; }
</style>