<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { 
  NCard, NButton, NTag, NSpin, NIcon, NEmpty, NBadge, useMessage, NModal, NGrid, NGi
} from 'naive-ui'
import { 
  FlameOutline, TimeOutline, CartOutline, CloseOutline, CheckmarkCircle
} from '@vicons/ionicons5'
import request from '../utils/request'

const message = useMessage()
const loading = ref(false)
const products = ref<any[]>([])

// 🛒 弹窗控制
const showSkuModal = ref(false)
const currentProduct = ref<any>(null)
const selectedSkuId = ref<number>(0)
const paying = ref(false)

// =======================
// 1. 数据获取
// =======================
const fetchMarketData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/market/products')
    products.value = res.data || []
  } catch (error) {
    message.error('数据加载失败')
  } finally {
    loading.value = false
  }
}

// =======================
// 2. 辅助计算
// =======================
// 获取商品的最低起步价
const getMinPrice = (p: any) => {
  if (!p.skus || p.skus.length === 0) return 0
  return Math.min(...p.skus.map((s: any) => s.price))
}

// 获取当前弹窗中选中的 SKU 对象
const currentSelectedSku = computed(() => {
  if (!currentProduct.value || !currentProduct.value.skus) return null
  return currentProduct.value.skus.find((s: any) => s.ID === selectedSkuId.value)
})

const formatDuration = (days: number) => {
  if (days === -1) return '永久有效'
  if (days === 365) return '1年'
  if (days >= 30 && days % 30 === 0) return `${days / 30}个月`
  return `${days}天`
}

// =======================
// 3. 交互逻辑
// =======================
const openSkuModal = (p: any) => {
  currentProduct.value = p
  // 默认选中第一个规格
  if (p.skus && p.skus.length > 0) {
    selectedSkuId.value = p.skus[0].ID
  } else {
    selectedSkuId.value = 0
  }
  showSkuModal.value = true
}

const handlePay = async () => {
  if (!currentSelectedSku.value) return
  
  paying.value = true
  try {
    const res: any = await request.post('/pay/create', {
      sku_id: currentSelectedSku.value.ID, 
      channel: 'mock'
    })
    if (res.pay_url) {
        message.loading('正在跳转收银台...')
        setTimeout(() => { window.location.href = res.pay_url }, 500)
    }
  } catch (error: any) {
    message.error(error.response?.data?.error || '下单失败')
    paying.value = false
  }
}

onMounted(fetchMarketData)
</script>

<template>
  <div class="market-page">
    <n-spin :show="loading" size="large">
      <h2 class="page-title">精选资源</h2>

      <div class="product-grid">
        <div v-for="p in products" :key="p.ID" class="grid-item">
          
          <n-badge :value="p.skus?.length ? '' : '缺货'" :type="p.skus?.length ? 'info' : 'default'" dot :offset="[-5, 5]">
            <n-card hoverable class="product-card" :bordered="false" @click="openSkuModal(p)">
              <div class="card-content">
                <div class="info-section">
                  <h3 class="p-name">{{ p.name }}</h3>
                  <p class="p-desc">{{ p.description || '暂无描述' }}</p>
                </div>
                
                <div class="footer-section">
                  <div class="price-box">
                    <span class="currency">¥</span>
                    <span class="amount">{{ getMinPrice(p) }}</span>
                    <span class="unit">起</span>
                  </div>
                  
                  <n-button secondary round size="small" type="primary" class="action-btn">
                    购买
                  </n-button>
                </div>
              </div>
            </n-card>
          </n-badge>

        </div>
      </div>

      <div v-if="!loading && products.length === 0" class="empty-box">
         <n-empty description="暂无商品上架" />
      </div>
    </n-spin>

    <n-modal v-model:show="showSkuModal" transform-origin="center">
      <n-card 
        class="sku-modal-card" 
        :bordered="false" 
        size="huge" 
        role="dialog" 
        aria-modal="true"
        :style="{ width: '500px', maxWidth: '90vw' }"
      >
        <template #header>
          <div class="modal-header">
            <span>选购详情</span>
            <n-button text class="close-btn" @click="showSkuModal = false">
              <n-icon size="24"><CloseOutline /></n-icon>
            </n-button>
          </div>
        </template>

        <div v-if="currentProduct">
          <div class="modal-product-info">
            <h3 class="modal-title">{{ currentProduct.name }}</h3>
            <p class="modal-desc">{{ currentProduct.description || '暂无详细描述' }}</p>
          </div>

          <div class="modal-sku-section">
            <div class="section-label">选择规格套餐</div>
            <div class="sku-grid-list">
              <div 
                v-for="sku in currentProduct.skus" 
                :key="sku.ID"
                class="sku-option"
                :class="{ active: selectedSkuId === sku.ID }"
                @click="selectedSkuId = sku.ID"
              >
                <div class="option-name">{{ sku.name }}</div>
                <div class="option-price">¥{{ sku.price }}</div>
                <div class="check-mark" v-if="selectedSkuId === sku.ID">
                  <n-icon><CheckmarkCircle /></n-icon>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <div class="footer-left" v-if="currentSelectedSku">
              <div class="total-label">总计</div>
              <div class="total-price">
                <span class="symbol">¥</span>{{ currentSelectedSku.price }}
              </div>
              <div class="duration-tag">
                <n-icon><TimeOutline/></n-icon> {{ formatDuration(currentSelectedSku.duration_days) }}
              </div>
            </div>
            <div v-else class="footer-left">暂无规格</div>

            <n-button 
              type="primary" 
              size="large" 
              class="pay-btn" 
              :loading="paying" 
              :disabled="!currentSelectedSku"
              @click="handlePay"
            >
              立即支付
            </n-button>
          </div>
        </div>
      </n-card>
    </n-modal>

  </div>
</template>

<style scoped>
.market-page { padding: 24px; min-height: 100%; }
.page-title { margin: 0 0 24px 0; font-size: 18px; font-weight: 600; color: #333; }

/* 网格布局：自动填充 */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 20px;
}

/* 外层极简卡片 */
.product-card {
  height: 100%; border-radius: 12px; background: #fff; cursor: pointer;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); transition: all 0.2s ease;
  border: 1px solid #f3f3f3;
}
.product-card:hover { transform: translateY(-3px); box-shadow: 0 8px 20px rgba(0,0,0,0.08); }
.card-content { display: flex; flex-direction: column; height: 100%; padding: 10px 5px; }
.info-section { flex: 1; }
.p-name { margin: 0 0 6px 0; font-size: 16px; font-weight: 700; color: #1a1a1a; }
.p-desc { font-size: 13px; color: #888; margin: 0; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.footer-section { margin-top: 20px; display: flex; justify-content: space-between; align-items: flex-end; }
.price-box { color: #d03050; font-weight: 800; line-height: 1; }
.currency { font-size: 14px; margin-right: 1px; }
.amount { font-size: 22px; }
.unit { font-size: 12px; color: #999; font-weight: normal; margin-left: 2px; }
.action-btn { font-weight: 600; }

/* ================== 弹窗样式 ================== */
.modal-header { display: flex; justify-content: space-between; align-items: center; font-size: 18px; font-weight: bold; }
.modal-product-info { margin-bottom: 24px; padding-bottom: 16px; border-bottom: 1px solid #eee; }
.modal-title { margin: 0 0 8px 0; font-size: 20px; }
.modal-desc { color: #666; font-size: 14px; line-height: 1.6; }

.modal-sku-section { margin-bottom: 30px; }
.section-label { font-size: 14px; font-weight: bold; color: #333; margin-bottom: 12px; }

/* SKU 网格列表 */
.sku-grid-list { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.sku-option { 
  border: 1px solid #e0e0e0; border-radius: 8px; padding: 12px; cursor: pointer; 
  position: relative; transition: all 0.2s; background: #fff;
}
.sku-option:hover { border-color: #36ad6a; background: #f0fdf4; }
.sku-option.active { border-color: #18a058; background: #e7f5ee; color: #18a058; box-shadow: 0 0 0 1px #18a058 inset; }
.option-name { font-weight: 600; font-size: 14px; margin-bottom: 4px; }
.option-price { font-size: 16px; font-weight: bold; }
.check-mark { position: absolute; top: 8px; right: 8px; color: #18a058; font-size: 18px; }

/* 底部结算条 */
.modal-footer { display: flex; justify-content: space-between; align-items: center; margin-top: 20px; padding-top: 20px; border-top: 1px dashed #eee; }
.footer-left { display: flex; align-items: baseline; gap: 8px; }
.total-label { font-size: 12px; color: #666; }
.total-price { color: #d03050; font-size: 28px; font-weight: 800; line-height: 1; }
.total-price .symbol { font-size: 16px; margin-right: 2px; }
.duration-tag { background: #f5f5f5; color: #666; padding: 2px 8px; border-radius: 4px; font-size: 12px; display: flex; align-items: center; gap: 4px; }
.pay-btn { padding: 0 32px; font-weight: bold; border-radius: 8px; background: linear-gradient(to right, #18a058, #2080f0); border: none; }
.pay-btn:hover { opacity: 0.9; }

.empty-box { padding: 100px 0; display: flex; justify-content: center; }
</style>