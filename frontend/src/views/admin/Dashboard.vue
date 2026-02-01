<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { 
  NCard, NGrid, NGi, NStatistic, NIcon, NButton, NTable, NTag, NSpace, 
  NModal, NForm, NFormItem, NUpload, useMessage, NAlert, NPopconfirm,
  NEmpty, NPopover, NImage
} from 'naive-ui'
import { 
  WalletOutline, PeopleOutline, TrendingUpOutline, 
  CardOutline, TimeOutline, CopyOutline, QrCodeOutline,
  TrashOutline, RefreshOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'
import { useUserStore } from '../../stores/user'

const userStore = useUserStore()
const message = useMessage()
const loading = ref(false)

// 数据源
const stats = ref<any>({})

// 提现相关
const showWithdrawModal = ref(false)
const withdrawLoading = ref(false)
const withdrawForm = reactive({ payment_image: '' })
const isEditingPayment = ref(false) // 是否正在修改收款码

// 审核/删除相关
const auditLoading = ref(false)
const deleteLoading = ref(false)
const clearLoading = ref(false)

// === 初始化加载 ===
const initData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/dashboard/stats')
    stats.value = res.data || {}
    // 如果已有收款码，初始化 form
    if (stats.value.payment_image) {
      withdrawForm.payment_image = stats.value.payment_image
      isEditingPayment.value = false
    } else {
      isEditingPayment.value = true // 没码就强制进入上传模式
    }
  } finally {
    loading.value = false
  }
}

// === 代理：复制邀请码 ===
const copyCode = () => {
  if (!stats.value.invitation_code) return
  navigator.clipboard.writeText(stats.value.invitation_code)
  message.success('邀请码已复制')
}

// === 代理：发起提现 ===
const handleUploadFinish = ({ file, event }: any) => {
  const res = JSON.parse(event.target.response)
  withdrawForm.payment_image = res.url
  // 不立即保存，等点击确认提现时一起提交
  message.success('收款码上传成功')
}

const submitWithdraw = async () => {
  if (!withdrawForm.payment_image) return message.warning('请上传收款码')
  
  withdrawLoading.value = true
  try {
    // 提交时，如果有新图片，后端会自动更新到 Profile
    await request.post('/admin/withdraw/apply', { 
      payment_image: withdrawForm.payment_image 
    })
    message.success('提现申请已提交')
    showWithdrawModal.value = false
    initData() // 刷新数据
  } catch (e) {
    // 错误由拦截器处理
  } finally {
    withdrawLoading.value = false
  }
}

// 切换修改模式
const toggleEditPayment = () => {
  isEditingPayment.value = !isEditingPayment.value
}

// === 管理员：审核提现 ===
const handleAudit = async (id: number, action: 'APPROVED' | 'REJECTED') => {
  auditLoading.value = true
  try {
    await request.post('/admin/withdraw/handle', {
      request_id: id,
      action: action,
      comment: action === 'APPROVED' ? '同意打款' : '信息有误，请核实'
    })
    message.success(action === 'APPROVED' ? '已通过并标记为已打款' : '已驳回申请')
    initData()
  } finally {
    auditLoading.value = false
  }
}

// === 管理员：删除单条记录 ===
const handleDelete = async (id: number) => {
  deleteLoading.value = true
  try {
    await request.delete(`/admin/withdraw/${id}`)
    message.success('记录已删除')
    initData()
  } finally {
    deleteLoading.value = false
  }
}

// === 管理员：一键清空 ===
const handleClear = async () => {
  clearLoading.value = true
  try {
    const res:any = await request.delete('/admin/withdraw/clear')
    message.success(res.message || '清理完成')
    initData()
  } finally {
    clearLoading.value = false
  }
}

const formatStatus = (status: string) => {
  switch(status) {
    case 'PENDING': return { type: 'warning', text: '待审核' }
    case 'APPROVED': return { type: 'success', text: '已打款' }
    case 'REJECTED': return { type: 'error', text: '已驳回' }
    default: return { type: 'default', text: status }
  }
}

onMounted(initData)
</script>

<template>
  <div class="dashboard-container">
    
    <div v-if="userStore.role === 'agent'">
      <div class="welcome-banner">
        <h2>👋 欢迎回来，合伙人 {{ userStore.nickname }}</h2>
        <p>这是您今日的战果，继续保持！</p>
      </div>

      <n-grid cols="1 s:3" responsive="screen" x-gap="12" y-gap="12">
        <n-gi>
          <n-card class="stat-card">
            <n-statistic label="可提现余额" :value="stats.available_balance || 0" :precision="2">
              <template #prefix>¥</template>
              <template #suffix>
                 <n-button size="tiny" type="primary" class="withdraw-btn" 
                   :disabled="stats.available_balance < 1 || stats.has_pending_withdraw"
                   @click="showWithdrawModal = true">
                   {{ stats.has_pending_withdraw ? '审核中' : '提现' }}
                 </n-button>
              </template>
            </n-statistic>
            <div class="stat-icon green"><n-icon><WalletOutline /></n-icon></div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="stat-card">
            <n-statistic label="累计总收益" :value="stats.total_profit || 0" :precision="2">
              <template #prefix>¥</template>
            </n-statistic>
            <div class="stat-icon blue"><n-icon><TrendingUpOutline /></n-icon></div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="stat-card">
            <n-statistic label="累计邀请用户" :value="stats.invite_count || 0">
              <template #suffix>人</template>
            </n-statistic>
            <div class="stat-icon purple"><n-icon><PeopleOutline /></n-icon></div>
          </n-card>
        </n-gi>
      </n-grid>

      <n-grid cols="1 s:2" responsive="screen" x-gap="12" y-gap="12" style="margin-top: 20px;">
        <n-gi>
          <n-card title="🚀 您的专属推广" size="small">
            <div class="invite-box" @click="copyCode">
              <span class="label">我的邀请码</span>
              <span class="code">{{ stats.invitation_code || '生成中...' }}</span>
              <n-icon class="copy-icon"><CopyOutline /></n-icon>
            </div>
            <p class="hint">点击卡片即可复制邀请码，发送给用户注册时填写。</p>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card title="📢 平台公告" size="small">
            <n-alert type="info" :show-icon="false">
              结算规则升级：现在支持一键全额提现，满 1 元即可申请。
            </n-alert>
          </n-card>
        </n-gi>
      </n-grid>

      <n-card title="🧾 最近入账记录" style="margin-top: 20px;">
        <n-table :bordered="false" :single-line="false">
          <thead>
            <tr>
              <th>订单号</th>
              <th>商品</th>
              <th>用户实付</th>
              <th>您的利润</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in stats.recent_sales" :key="item.ID">
              <td>{{ item.order_id.substring(0,8) }}...</td>
              <td>{{ item.description }}</td>
              <td>¥{{ item.final_amount.toFixed(2) }}</td>
              <td style="color: #d03050; font-weight: bold;">+¥{{ item.agent_profit.toFixed(2) }}</td>
              <td>
                <n-tag size="small" :type="item.withdraw_status === 2 ? 'success' : (item.withdraw_status === 1 ? 'warning' : (item.withdraw_status === 3 ? 'error' : 'default'))">
                  {{ item.withdraw_status === 2 ? '已到账' : (item.withdraw_status === 1 ? '审核中' : (item.withdraw_status === 3 ? '已驳回' : '未提现')) }}
                </n-tag>
              </td>
            </tr>
            <tr v-if="!stats.recent_sales || stats.recent_sales.length === 0">
              <td colspan="5" style="text-align: center; color: #999;">暂无收益，快去推广吧！</td>
            </tr>
          </tbody>
        </n-table>
      </n-card>
    </div>

    <div v-else-if="userStore.role === 'admin'">
      <div class="welcome-banner">
        <h2>🛡️ 系统监控台</h2>
        <p>全站数据概览与财务审核</p>
      </div>

      <n-grid cols="1 s:3" responsive="screen" x-gap="12" y-gap="12">
        <n-gi>
          <n-card class="stat-card">
            <n-statistic label="平台总流水" :value="stats.total_revenue || 0" :precision="2">
              <template #prefix>¥</template>
            </n-statistic>
            <div class="stat-icon green"><n-icon><CardOutline /></n-icon></div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="stat-card">
            <n-statistic label="总注册用户" :value="stats.total_users || 0">
               <template #suffix>人</template>
            </n-statistic>
            <div class="stat-icon blue"><n-icon><PeopleOutline /></n-icon></div>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card class="stat-card" :class="{ 'urgent': stats.pending_count > 0 }">
            <n-statistic label="待审核提现" :value="stats.pending_count || 0">
               <template #suffix>笔</template>
            </n-statistic>
            <div class="stat-icon orange"><n-icon><TimeOutline /></n-icon></div>
          </n-card>
        </n-gi>
      </n-grid>

      <n-card title="🛍️ 用户购买动态 (最新10笔)" style="margin-top: 20px;">
        <n-table size="small">
          <thead>
            <tr>
              <th>时间</th>
              <th>购买用户</th>
              <th>购买商品</th>
              <th>支付金额</th>
              <th>订单号</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in stats.recent_orders" :key="order.order_no">
              <td>{{ new Date(order.created_at).toLocaleString() }}</td>
              <td>{{ order.username || '未知用户' }}</td>
              <td>{{ order.product_name || '未知商品' }}</td>
              <td style="font-weight: bold;">¥{{ order.amount.toFixed(2) }}</td>
              <td style="font-family: monospace;">{{ order.order_no.substring(0,8) }}...</td>
              <td><n-tag size="small" type="success">{{ order.status }}</n-tag></td>
            </tr>
            <tr v-if="!stats.recent_orders || stats.recent_orders.length === 0">
              <td colspan="6" style="text-align: center; color: #999;">暂无购买记录</td>
            </tr>
          </tbody>
        </n-table>
      </n-card>

      <n-card title="🏦 提现申请管理" style="margin-top: 20px;">
        <template #header-extra>
          <n-space>
             <n-button size="small" @click="initData">刷新列表</n-button>
             <n-popconfirm @positive-click="handleClear">
               <template #trigger>
                 <n-button size="small" type="error" ghost>一键清空历史</n-button>
               </template>
               确定要永久删除所有“已打款”和“已驳回”的记录吗？<br>待审核记录将被保留。
             </n-popconfirm>
          </n-space>
        </template>
        
        <n-table>
          <thead>
            <tr>
              <th>ID</th>
              <th>代理人</th>
              <th>金额</th>
              <th>收款码</th>
              <th>申请时间</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in stats.withdraw_list" :key="item.ID">
              <td>#{{ item.ID }}</td>
              <td>{{ item.agent_name }}</td>
              <td style="font-weight: bold; font-size: 16px;">¥{{ item.amount.toFixed(2) }}</td>
              <td>
                <n-popover trigger="hover">
                  <template #trigger>
                    <n-icon size="24" style="cursor: pointer; color: #2080f0"><QRCodeOutline /></n-icon>
                  </template>
                  <img :src="'http://localhost:8080' + item.payment_image" style="width: 200px; height: 200px; object-fit: contain;">
                </n-popover>
              </td>
              <td>{{ new Date(item.CreatedAt).toLocaleString() }}</td>
              <td>
                <n-tag size="small" :type="formatStatus(item.status).type">
                  {{ formatStatus(item.status).text }}
                </n-tag>
              </td>
              <td>
                <n-space v-if="item.status === 'PENDING'">
                   <n-popconfirm @positive-click="handleAudit(item.ID, 'APPROVED')">
                     <template #trigger>
                       <n-button size="small" type="success">打款</n-button>
                     </template>
                     确认已线下打款 ¥{{ item.amount }} 给对方？
                   </n-popconfirm>
                   
                   <n-popconfirm @positive-click="handleAudit(item.ID, 'REJECTED')">
                     <template #trigger>
                       <n-button size="small" type="error" ghost>驳回</n-button>
                     </template>
                     确定驳回？资金将被冻结。
                   </n-popconfirm>
                </n-space>

                <div v-else>
                   <n-popconfirm @positive-click="handleDelete(item.ID)">
                     <template #trigger>
                       <n-button size="small" type="default" circle>
                         <template #icon><n-icon><TrashOutline /></n-icon></template>
                       </n-button>
                     </template>
                     删除此记录？
                   </n-popconfirm>
                </div>
              </td>
            </tr>
            <tr v-if="!stats.withdraw_list || stats.withdraw_list.length === 0">
              <td colspan="7" style="text-align: center; padding: 30px;">
                <n-empty description="暂无提现申请记录" />
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-card>
    </div>
    
    <n-modal v-model:show="showWithdrawModal" preset="card" title="申请提现" style="width: 400px">
      <div style="text-align: center; margin-bottom: 20px;">
        <h2 style="color: #d03050; margin: 0;">¥{{ (stats.available_balance || 0).toFixed(2) }}</h2>
        <p style="color: #999; margin: 5px 0 0;">本次提现金额 (全额)</p>
      </div>
      
      <div v-if="!isEditingPayment && withdrawForm.payment_image" style="text-align: center;">
        <p style="margin-bottom: 10px; font-weight: bold;">收款码预览</p>
        <n-image 
          width="200" 
          :src="'http://localhost:8080' + withdrawForm.payment_image" 
          style="border-radius: 8px; border: 1px solid #eee;"
        />
        <div style="margin-top: 10px;">
          <n-button text type="primary" size="small" @click="toggleEditPayment">
            <template #icon><n-icon><RefreshOutline/></n-icon></template>
            更换收款码
          </n-button>
        </div>
      </div>

      <n-form v-else>
        <n-form-item label="请上传您的收款码 (微信/支付宝)">
          <n-upload 
            action="http://localhost:8080/api/v1/upload/payment" 
            :headers="{ Authorization: 'Bearer ' + userStore.token }"
            :max="1"
            list-type="image-card"
            name="file"
            @finish="handleUploadFinish"
          />
        </n-form-item>
        <div v-if="stats.payment_image" style="text-align: right;">
          <n-button text size="small" @click="toggleEditPayment">取消修改</n-button>
        </div>
      </n-form>
      
      <template #footer>
        <n-button type="primary" block size="large" @click="submitWithdraw" :loading="withdrawLoading">
          {{ isEditingPayment ? '保存并申请提现' : '确认申请' }}
        </n-button>
      </template>
    </n-modal>

  </div>
</template>

<style scoped>
.dashboard-container { padding: 20px; }
.welcome-banner { margin-bottom: 24px; }
.welcome-banner h2 { margin: 0; color: #333; }
.welcome-banner p { margin: 5px 0 0; color: #666; }

.stat-card { position: relative; overflow: hidden; }
.stat-icon { 
  position: absolute; right: 20px; top: 20px; 
  font-size: 40px; opacity: 0.15; 
}
.green { color: #18a058; }
.blue { color: #2080f0; }
.purple { color: #8a2be2; }
.orange { color: #f0a020; }

.withdraw-btn { margin-left: 10px; position: relative; top: -2px; }

.invite-box {
  background: linear-gradient(135deg, #f0f9ff 0%, #e6f7ff 100%);
  border: 1px dashed #2080f0;
  padding: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s;
}
.invite-box:hover { background: #e0f2fe; }
.invite-box .label { font-size: 12px; color: #666; }
.invite-box .code { font-size: 24px; font-weight: bold; color: #2080f0; letter-spacing: 2px; }
.hint { font-size: 12px; color: #999; margin-top: 8px; }

.urgent { border: 1px solid #f0a020; }
</style>