<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NDataTable, NTag, NButton, NSpace, NInput, NModal, useMessage, 
  NForm, NFormItem, NSelect, NPageHeader, NList, NListItem, NThing, 
  NPopconfirm, NIcon, NEmpty, NDivider, NRadio, NRadioGroup, NPopover, NBadge
} from 'naive-ui'
import { 
  SearchOutline, WalletOutline, TimeOutline, GiftOutline, InfiniteOutline
} from '@vicons/ionicons5'
import request from '../../utils/request' 
import { format, differenceInDays } from 'date-fns'

const message = useMessage()
const loading = ref(false)
const list = ref([])
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const keyword = ref('')

// ==========================================
// 🛠️ 工具函数
// ==========================================
const isPermanent = (dateStr: string) => {
    if (!dateStr) return false
    return new Date(dateStr).getFullYear() > 2090
}

const formatFriendlyDate = (dateStr: string) => {
    if (!dateStr) return '未知'
    const date = new Date(dateStr)
    if (date.getFullYear() > 2090) {
        return '永久有效'
    }
    return format(date, 'yyyy-MM-dd HH:mm')
}

// === 1. 列表展示逻辑 ===
const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '客户用户名', key: 'username', width: 150 },
  { 
    title: '当前持仓概览', 
    key: 'products',
    width: 300,
    render(row: any) {
       const products = row.user_products || []

       // 情况A：穷光蛋
       if (products.length === 0) {
           return h(NTag, { bordered: false, size: 'small' }, { default: () => '暂无授权' })
       }

       // 情况B：有资产 -> 悬浮气泡查看详情
       return h(NPopover, { trigger: 'hover', style: { maxWidth: '350px' } }, {
           trigger: () => h(NTag, 
               { type: 'success', bordered: false, style: 'cursor: pointer' }, 
               { default: () => `已授权 ${products.length} 项权益` }
           ),
           default: () => {
               return h(NList, { size: 'small', bordered: false }, {
                   default: () => products.map((up: any) => {
                       const expireDate = new Date(up.expire_at)
                       const daysLeft = differenceInDays(expireDate, new Date())
                       const isPerm = isPermanent(up.expire_at)
                       
                       let tagType = 'success'
                       let tagText = `剩 ${daysLeft} 天`
                       
                       if (isPerm) {
                           tagType = 'info' 
                           tagText = '永久'
                       } else if (daysLeft < 0) {
                           tagType = 'error'
                           tagText = '已过期'
                       } else if (daysLeft < 30) {
                           tagType = 'warning'
                       }

                       return h(NListItem, {}, {
                           default: () => h(NThing, 
                               { 
                                   title: up.product_name || up.product?.name || '未知商品', 
                                   titleExtra: h(NTag, { size: 'small', type: tagType, bordered: false }, { default: () => tagText })
                               },
                               { 
                                   description: () => isPerm 
                                    ? h('span', { style: 'color: #2080f0' }, [h(NIcon, { component: InfiniteOutline }), ' 永久有效'])
                                    : `有效期至: ${format(expireDate, 'yyyy-MM-dd')}` 
                               }
                           )
                       })
                   })
               })
           }
       })
    }
  },
  {
    title: '业务办理', key: 'actions', fixed: 'right', width: 150, align: 'center',
    render(row: any) {
      return h(NButton, { size: 'small', type: 'primary', onClick: () => openAuthModal(row) }, 
        { icon: () => h(NIcon, null, { default: () => h(WalletOutline) }), default: () => '发证/核销' })
    }
  }
]

// API
const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/users', {
      params: { 
        page: pagination.page, 
        page_size: pagination.pageSize, 
        keyword: keyword.value,
        // 🔥🔥🔥 核心修改：只查询普通用户，过滤掉 admin 和 agent 🔥🔥🔥
        role: 'user' 
      }
    })
    list.value = res.data || []
    pagination.itemCount = res.total || 0
  } catch { message.error('加载失败') } finally { loading.value = false }
}
const handleSearch = () => { pagination.page = 1; fetchData() }
const handlePageChange = (page: number) => { pagination.page = page; fetchData() }

// === 2. 授权窗口逻辑 ===
const showAuthModal = ref(false)
const currentCustomer = ref<any>({})
const userProducts = ref<any[]>([])
const allProducts = ref<any[]>([])
const grantForm = reactive({ productId: null, days: 365 })
const granting = ref(false)

const openAuthModal = (user: any) => {
    currentCustomer.value = user
    showAuthModal.value = true
    grantForm.productId = null
    fetchUserProducts(user.id) 
    fetchAllProducts()
}

// 获取持有列表
const fetchUserProducts = async (uid: number) => {
    try {
        const res: any = await request.get(`/admin/users/${uid}/products`)
        userProducts.value = res.data || []
    } catch {}
}
// 获取商品列表
const fetchAllProducts = async () => {
    try {
        const res: any = await request.get('/admin/products')
        const validProducts = (res.data || []).filter((p:any) => p.is_on_shelf)
        allProducts.value = validProducts.map((p:any) => ({ label: p.name, value: p.ID }))
    } catch {}
}
// 发放
const handleGrant = async () => {
    if(!grantForm.productId) return message.warning('请选择商品')
    granting.value = true
    try {
        await request.post('/admin/users/grant', {
            user_id: currentCustomer.value.id, 
            product_id: grantForm.productId, 
            duration_days: grantForm.days
        })
        message.success('发放成功')
        await fetchUserProducts(currentCustomer.value.id) 
        await fetchData()
    } catch { 
        message.error('发放失败') 
    } finally { 
        granting.value = false 
    }
}
// 收回
const handleRevoke = async (pid: number) => {
    try {
        await request.post('/admin/users/revoke', { user_id: currentCustomer.value.id, product_id: pid })
        message.success('已收回')
        await fetchUserProducts(currentCustomer.value.id)
        await fetchData() 
    } catch { message.error('操作失败') }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container">
    <n-page-header title="💼 业务授权大厅" subtitle="管理用户的付费商品持有情况（发证/核销）" style="margin-bottom: 20px;" />
    
    <n-card>
      <div class="toolbar">
        <n-input v-model:value="keyword" placeholder="输入客户用户名..." style="width: 240px" @keydown.enter="handleSearch">
            <template #prefix><n-icon><SearchOutline/></n-icon></template>
        </n-input>
        <n-button type="primary" @click="handleSearch">查询客户</n-button>
      </div>
      <n-data-table 
        remote 
        :columns="columns" 
        :data="list" 
        :loading="loading" 
        :pagination="pagination" 
        @update:page="handlePageChange" 
        :row-key="(row) => row.id"
        style="margin-top: 16px;" 
      />
    </n-card>

    <n-modal v-model:show="showAuthModal" preset="card" title="业务办理窗口" style="width: 600px">
        <template #header-extra>
            客户：<span style="font-weight: bold; color: #18a058">{{ currentCustomer.username }}</span>
        </template>

        <div class="section-title">📦 该客户持有的商品凭证：</div>
        <div class="product-list-box">
            <n-list hoverable>
                <n-list-item v-for="up in userProducts" :key="up.ID">
                    <n-thing>
                        <template #header>
                            <span :style="{ color: up.product_name ? '#333' : '#999' }">
                                {{ up.product_name || up.product?.name || '未知商品' }}
                            </span>
                        </template>
                        <template #description>
                            <span v-if="isPermanent(up.expire_at)" class="text-permanent">
                                <n-icon class="icon-fix"><InfiniteOutline/></n-icon> 永久有效
                            </span>
                            <span v-else :class="new Date(up.expire_at) > new Date() ? 'text-valid' : 'text-expired'">
                                <n-icon class="icon-fix"><TimeOutline/></n-icon> 
                                有效期至：{{ formatFriendlyDate(up.expire_at) }}
                                <span v-if="new Date(up.expire_at) < new Date()">(已过期)</span>
                            </span>
                        </template>
                    </n-thing>
                    <template #suffix>
                        <n-popconfirm @positive-click="handleRevoke(up.product_id)">
                            <template #trigger>
                                <n-button size="small" type="error" secondary>收回</n-button>
                            </template>
                            ⚠️ 高危操作：<br>
                            确定要强制收回该凭证吗？<br>
                            该操作将计入审计日志。
                        </n-popconfirm>
                    </template>
                </n-list-item>
                <n-empty v-if="userProducts.length === 0" description="暂未持有任何有效商品" style="padding: 20px 0" />
            </n-list>
        </div>

        <n-divider dashed />

        <div class="section-title">🎁 发放新凭证：</div>
        <div style="background: #f9f9f9; padding: 16px; border-radius: 8px;">
            <n-form label-placement="left" label-width="80">
                <n-form-item label="选择商品">
                    <n-select v-model:value="grantForm.productId" :options="allProducts" placeholder="请选择商品" filterable />
                </n-form-item>
                
                <n-form-item label="授权时长">
                    <n-radio-group v-model:value="grantForm.days" name="durationGroup">
                        <n-space>
                            <n-radio :value="7">7天</n-radio>
                            <n-radio :value="30">30天</n-radio>
                            <n-radio :value="120">一学期</n-radio>
                            <n-radio :value="365">一年</n-radio>
                            <n-radio :value="-1">
                                <span style="font-weight: bold; color: #2080f0">永久</span>
                            </n-radio>
                        </n-space>
                    </n-radio-group>
                </n-form-item>

                <n-button type="primary" block @click="handleGrant" :loading="granting" :disabled="!grantForm.productId">
                    <template #icon><n-icon><GiftOutline/></n-icon></template> 立即发放
                </n-button>
            </n-form>
        </div>
    </n-modal>
  </div>
</template>

<style scoped>
.page-container { padding: 24px; }
.toolbar { display: flex; gap: 12px; }
.section-title { font-weight: bold; margin-bottom: 10px; color: #333; font-size: 15px; }
.product-list-box { border: 1px solid #eee; border-radius: 4px; max-height: 250px; overflow-y: auto; }

/* 状态颜色 */
.text-valid { color: #18a058; }
.text-expired { color: #d03050; }
.text-permanent { color: #2080f0; font-weight: bold; }

.icon-fix { position: relative; top: 2px; margin-right: 4px; }
</style>