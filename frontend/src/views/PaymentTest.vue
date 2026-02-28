<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { 
  NCard, NButton, NSpin, NIcon, NEmpty, NBadge, useMessage, NModal, NTag, NDivider, NSkeleton
} from 'naive-ui'
import { 
  CloseOutline, CheckmarkCircle, WalletOutline, GiftOutline, PricetagOutline, FlameOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'

const message = useMessage()
const userStore = useUserStore()
const router = useRouter()

const loading = ref(false)
const products = ref<any[]>([])
const currentCategory = ref<string>('') // 当前选中的分类

// 🛒 弹窗与详情控制
const showSkuModal = ref(false)
const currentProduct = ref<any>(null)
const selectedSkuId = ref<number>(0)
const exchanging = ref(false)

// 详情独立加载状态
const detailLoading = ref(false)
const fullProductDetail = ref<any>(null)

// 动态提取存在的分类
const availableCategories = computed(() => {
  const cats = new Set(products.value.map(p => p.category).filter(c => c))
  return ['全部', ...Array.from(cats)]
})

// =======================
// 1. 数据获取
// =======================
const fetchMarketData = async (category: string = '') => {
  loading.value = true
  currentCategory.value = category === '全部' ? '' : category
  try {
    const res: any = await request.get('/market/products', {
      params: { category: currentCategory.value }
    })
    products.value = res.data || []
    userStore.fetchProfile() // 进页面刷新积分
  } catch (error) {
    message.error('数据加载失败')
  } finally {
    loading.value = false
  }
}

// =======================
// 2. 辅助与格式化计算
// =======================
const getMinPoints = (p: any) => {
  if (!p.skus || p.skus.length === 0) return 0
  return Math.min(...p.skus.map((s: any) => s.points))
}

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

const parseTags = (tagStr: string) => {
  if (!tagStr) return []
  return tagStr.split(',').filter(t => t.trim() !== '')
}

const getCoverUrl = (url: string | undefined) => {
  if (!url) return 'https://images.unsplash.com/photo-1606326608606-aa0b62935f2b?q=80&w=800&auto=format&fit=crop' // 默认好看的占位图
  return url.startsWith('http') ? url : `http://localhost:8080${url}`
}

// =======================
// 3. 交互逻辑
// =======================

// 🔥 打开弹窗并独立请求富文本详情
const openSkuModal = async (p: any) => {
  currentProduct.value = p
  if (p.skus && p.skus.length > 0) {
    selectedSkuId.value = p.skus[0].ID
  } else {
    selectedSkuId.value = 0
  }
  showSkuModal.value = true
  
  // 重置并拉取详情
  fullProductDetail.value = null
  detailLoading.value = true
  try {
    const res: any = await request.get(`/market/products/${p.ID}`)
    fullProductDetail.value = res.data
  } catch (e) {
    message.error('商品详情获取失败')
  } finally {
    detailLoading.value = false
  }
}

const handleExchange = async () => {
  if (!currentSelectedSku.value) return
  
  const cost = currentSelectedSku.value.points
  if (userStore.points < cost) {
      message.warning('您的积分不足，请先获取积分')
      return
  }
  
  exchanging.value = true
  try {
    await request.post('/product/exchange', {
      sku_id: currentSelectedSku.value.ID
    })
    
    message.success('🎉 兑换成功！权益已即时生效')
    showSkuModal.value = false
    await userStore.fetchProfile() // 更新头部积分
  } catch (error: any) {
    message.error(error.response?.data?.error || '兑换失败')
  } finally {
    exchanging.value = false
  }
}

const goToRecharge = () => {
    router.push('/profile')
}

onMounted(() => fetchMarketData('全部'))
</script>

<template>
  <div class="market-page">
    <div class="market-header">
        <div class="title-area">
          <h2 class="page-title">
            <n-icon color="#f59e0b" class="title-icon"><GiftOutline/></n-icon> 
            <span>积分兑换中心</span>
          </h2>
          <p class="page-sub">使用积分兑换高级题库、课程与周边权益。</p>
        </div>
        
        <div class="my-points-card">
            <div class="points-info">
              <span class="label">当前余额</span>
              <div class="val-group">
                <span class="val">{{ userStore.points }}</span>
                <span class="unit">分</span>
              </div>
            </div>
            <n-button size="small" secondary type="primary" class="top-btn" @click="goToRecharge">
              获取积分
            </n-button>
        </div>
    </div>

    <div class="category-tabs" v-if="availableCategories.length > 1">
        <div 
          v-for="cat in availableCategories" 
          :key="cat"
          class="cat-pill"
          :class="{ active: (currentCategory === cat) || (currentCategory === '' && cat === '全部') }"
          @click="fetchMarketData(cat)"
        >
          {{ cat }}
        </div>
    </div>

    <n-spin :show="loading" size="large">
      <div class="product-grid">
        <div v-for="p in products" :key="p.ID" class="grid-item">
          <n-badge :value="p.skus?.length ? '' : '缺货'" :type="p.skus?.length ? 'info' : 'default'" dot :offset="[-5, 5]">
            <n-card hoverable class="product-card" :bordered="false" @click="openSkuModal(p)">
              
              <div class="cover-wrapper">
                <img :src="getCoverUrl(p.cover_img)" class="p-cover" alt="封面" />
                <div class="cat-tag" v-if="p.category">{{ p.category }}</div>
              </div>

              <div class="card-content">
                <div class="info-section">
                  <h3 class="p-name">{{ p.name }}</h3>
                  
                  <div class="tags-row" v-if="p.tags">
                    <n-tag v-for="tag in parseTags(p.tags)" :key="tag" size="small" type="warning" round :bordered="false">
                      <template #icon><n-icon><FlameOutline /></n-icon></template>
                      {{ tag }}
                    </n-tag>
                  </div>

                  <p class="p-desc" :class="{'mt-2': !p.tags}">{{ p.description || '暂无描述' }}</p>
                </div>
                
                <div class="footer-section">
                  <div class="price-box">
                    <n-icon size="16"><WalletOutline/></n-icon>
                    <span class="amount" v-if="getMinPoints(p) > 0">{{ getMinPoints(p) }}</span>
                    <span class="amount free-text" v-else>免费</span>
                    <span class="unit" v-if="getMinPoints(p) > 0">起</span>
                  </div>
                  
                  <n-button secondary round size="small" type="primary" class="action-btn">
                    立即查看
                  </n-button>
                </div>
              </div>
            </n-card>
          </n-badge>
        </div>
      </div>

      <div v-if="!loading && products.length === 0" class="empty-box">
         <n-empty description="该分类下暂无商品" />
      </div>
    </n-spin>

    <n-modal v-model:show="showSkuModal" transform-origin="center">
      <n-card 
        class="sku-modal-card" 
        :bordered="false" 
        size="huge" 
        role="dialog" 
        aria-modal="true"
        :style="{ maxWidth: '850px', padding: 0 }"
      >
        <template #header>
          <div class="modal-header">
            <span>商品详情</span>
            <n-button text class="close-btn" @click="showSkuModal = false">
              <n-icon size="24"><CloseOutline /></n-icon>
            </n-button>
          </div>
        </template>

        <div v-if="currentProduct" class="modal-split-layout">
          <div class="modal-left-detail">
             <img :src="getCoverUrl(currentProduct.cover_img)" class="detail-hero-img" />
             <h2 class="detail-title">{{ currentProduct.name }}</h2>
             <p class="detail-subtitle">{{ currentProduct.description }}</p>
             <n-divider dashed>图文介绍</n-divider>
             
             <div class="rich-text-content" v-if="!detailLoading && fullProductDetail">
               <div v-html="fullProductDetail.detail || '<p style=\'color:#94a3b8;text-align:center;\'>暂无图文详情</p>'"></div>
             </div>
             <div v-else-if="detailLoading" class="loading-skeleton">
                <n-skeleton text :repeat="4" /> <n-skeleton text style="width: 60%" />
             </div>
          </div>

          <div class="modal-right-action">
            <div class="action-sticky-wrap">
              <div class="modal-sku-section">
                <div class="section-label">选择兑换时长/规格</div>
                <div class="sku-grid-list">
                  <div 
                    v-for="sku in currentProduct.skus" 
                    :key="sku.ID"
                    class="sku-option"
                    :class="{ active: selectedSkuId === sku.ID }"
                    @click="selectedSkuId = sku.ID"
                  >
                    <div class="option-left">
                      <div class="option-name">{{ sku.name }}</div>
                      <div class="option-duration">{{ formatDuration(sku.duration_days) }}</div>
                    </div>
                    
                    <div class="option-right">
                       <div class="option-price">{{ sku.points === 0 ? '限免' : sku.points + '分' }}</div>
                       <div class="check-mark" v-if="selectedSkuId === sku.ID">
                          <n-icon><CheckmarkCircle /></n-icon>
                       </div>
                    </div>
                  </div>
                  <div v-if="!currentProduct.skus || currentProduct.skus.length === 0" style="color:red; font-size:13px;">
                    该商品规格异常，暂不可兑换
                  </div>
                </div>
              </div>

              <div class="modal-footer">
                <div class="footer-info" v-if="currentSelectedSku">
                  <span class="total-label">消耗:</span>
                  <span class="total-price" v-if="currentSelectedSku.points > 0">{{ currentSelectedSku.points }}</span>
                  <span class="total-price" v-else>0 (免费)</span>
                  <span class="total-unit" v-if="currentSelectedSku.points > 0">积分</span>
                </div>
                <div v-else class="footer-info">暂无规格</div>

                <div class="footer-action">
                    <n-button 
                      v-if="currentSelectedSku && userStore.points >= currentSelectedSku.points"
                      type="primary" 
                      size="large" 
                      class="pay-btn" 
                      :loading="exchanging" 
                      @click="handleExchange"
                    >
                      确认兑换
                    </n-button>
                    <n-button 
                      v-else-if="currentSelectedSku"
                      type="warning" 
                      dashed
                      size="large" 
                      class="pay-btn"
                      @click="goToRecharge"
                    >
                      积分不足 (去获取)
                    </n-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </n-card>
    </n-modal>

  </div>
</template>

<style scoped>
.market-page { padding: 24px; min-height: 100%; max-width: 1200px; margin: 0 auto; }

/* 头部样式 */
.market-header { 
  display: flex; justify-content: space-between; align-items: center; 
  margin-bottom: 24px; flex-wrap: wrap; gap: 16px;
}
.page-title { margin: 0; font-size: 24px; font-weight: 800; color: #1e293b; display: flex; align-items: center; }
.title-icon { margin-right: 8px; }
.page-sub { margin: 4px 0 0 32px; color: #64748b; font-size: 14px; }

.my-points-card {
    background: linear-gradient(to right, #fffbeb, #fff7ed); 
    border: 1px solid #fcd34d; padding: 12px 24px; border-radius: 16px;
    display: flex; align-items: center; gap: 24px;
    box-shadow: 0 4px 12px rgba(245, 158, 11, 0.1);
}
.points-info { display: flex; flex-direction: column; }
.my-points-card .label { font-size: 12px; font-weight: 700; color: #b45309; text-transform: uppercase; letter-spacing: 0.5px; }
.val-group { display: flex; align-items: baseline; gap: 4px; margin-top: 2px;}
.my-points-card .val { font-size: 24px; font-weight: 900; color: #d97706; font-family: monospace; line-height: 1; }
.my-points-card .unit { font-size: 13px; color: #b45309; font-weight: bold; }

/* 分类 Tab */
.category-tabs { display: flex; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
.cat-pill { 
  padding: 6px 16px; border-radius: 100px; font-size: 14px; font-weight: 600;
  background: #f1f5f9; color: #475569; cursor: pointer; transition: all 0.2s;
}
.cat-pill:hover { background: #e2e8f0; }
.cat-pill.active { background: #3b82f6; color: #fff; box-shadow: 0 4px 10px rgba(59, 130, 246, 0.3); }

/* 网格布局 */
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px; }

/* 卡片样式 */
.product-card {
  height: 100%; border-radius: 16px; background: #fff; cursor: pointer;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04); transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid #f1f5f9; overflow: hidden; padding: 0;
}
.product-card:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); border-color: #e2e8f0; }

.cover-wrapper { position: relative; width: 100%; aspect-ratio: 16/9; background: #f8fafc; overflow: hidden; }
.p-cover { width: 100%; height: 100%; object-fit: cover; transition: transform 0.5s; }
.product-card:hover .p-cover { transform: scale(1.05); }
.cat-tag { 
  position: absolute; top: 12px; left: 12px; background: rgba(0,0,0,0.6); 
  color: #fff; font-size: 11px; padding: 4px 8px; border-radius: 6px; backdrop-filter: blur(4px);
}

.card-content { display: flex; flex-direction: column; height: 100%; padding: 16px; }
.info-section { flex: 1; }
.p-name { margin: 0 0 8px 0; font-size: 17px; font-weight: 800; color: #1e293b; line-height: 1.4; }
.tags-row { display: flex; gap: 6px; margin-bottom: 8px; flex-wrap: wrap; }
.p-desc { font-size: 13px; color: #64748b; margin: 0; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.mt-2 { margin-top: 8px; }

.footer-section { margin-top: 20px; display: flex; justify-content: space-between; align-items: center; }
.price-box { color: #d97706; font-weight: 800; display: flex; align-items: baseline; gap: 4px; }
.amount { font-size: 22px; font-family: monospace; }
.free-text { color: #10b981; font-size: 18px; font-family: inherit; }
.unit { font-size: 12px; color: #94a3b8; font-weight: normal; }

/* ================== 弹窗双栏样式 ================== */
:deep(.n-card__content) { padding: 0 !important; }
.modal-header { display: flex; justify-content: space-between; align-items: center; font-size: 18px; font-weight: bold; padding: 16px 24px; border-bottom: 1px solid #f1f5f9;}

.modal-split-layout { display: flex; flex-direction: row; height: 60vh; min-height: 500px; }

/* 左侧富文本 */
.modal-left-detail { flex: 6; padding: 24px; overflow-y: auto; border-right: 1px solid #f1f5f9; background: #fafaf9;}
.detail-hero-img { width: 100%; border-radius: 12px; margin-bottom: 20px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
.detail-title { margin: 0 0 8px 0; font-size: 24px; font-weight: 900; color: #1e293b; }
.detail-subtitle { color: #64748b; font-size: 14px; line-height: 1.6; margin-bottom: 20px; }

.rich-text-content { font-size: 15px; color: #334155; line-height: 1.8; }
/* 深度重置富文本样式，防止后台传来的脏 HTML 破环排版 */
:deep(.rich-text-content img) { max-width: 100%; border-radius: 8px; margin: 12px 0; }
:deep(.rich-text-content p) { margin-bottom: 1em; }

/* 右侧购买操作 */
.modal-right-action { flex: 4; padding: 24px; background: #fff; position: relative; }
.action-sticky-wrap { position: sticky; top: 0; display: flex; flex-direction: column; height: 100%; justify-content: space-between;}

.section-label { font-size: 15px; font-weight: 800; color: #334155; margin-bottom: 16px; }

.sku-grid-list { display: flex; flex-direction: column; gap: 12px; }
.sku-option { 
  border: 2px solid #e2e8f0; border-radius: 12px; padding: 16px; cursor: pointer; 
  display: flex; justify-content: space-between; align-items: center;
  transition: all 0.2s; background: #fff;
}
.sku-option:hover { border-color: #fbbf24; background: #fffbeb; }
.sku-option.active { border-color: #f59e0b; background: #fff7ed; box-shadow: 0 4px 12px rgba(245, 158, 11, 0.15); }

.option-left { display: flex; flex-direction: column; gap: 4px; }
.option-name { font-weight: 800; font-size: 16px; color: #1e293b; }
.option-duration { font-size: 13px; color: #94a3b8; font-weight: 500;}

.option-right { display: flex; align-items: center; gap: 8px; }
.option-price { font-size: 18px; font-weight: 900; color: #d97706; font-family: monospace;}
.check-mark { color: #f59e0b; font-size: 24px; display: flex; }

.modal-footer { margin-top: 32px; padding-top: 24px; border-top: 2px dashed #f1f5f9; }
.footer-info { display: flex; align-items: baseline; gap: 6px; margin-bottom: 16px; justify-content: center;}
.total-label { font-size: 14px; font-weight: bold; color: #64748b; }
.total-price { color: #d97706; font-size: 32px; font-weight: 900; line-height: 1; font-family: monospace;}
.total-unit { font-size: 13px; font-weight: bold; color: #b45309; }
.pay-btn { width: 100%; height: 48px; font-size: 16px; font-weight: bold; border-radius: 12px; background: linear-gradient(135deg, #f59e0b, #d97706); border: none; box-shadow: 0 4px 12px rgba(245, 158, 11, 0.3);}
.pay-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(245, 158, 11, 0.4); }

.empty-box { padding: 100px 0; display: flex; justify-content: center; }

/* =========================================
   📱 移动端响应式适配
   ========================================= */
@media (max-width: 768px) {
  .market-page { padding: 16px; }
  .market-header { flex-direction: column; align-items: stretch; gap: 16px; }
  .page-sub { margin-left: 0; margin-top: 8px;}
  .my-points-card { justify-content: space-between; }
  .product-grid { grid-template-columns: 1fr; gap: 16px; }
  
  /* 移动端弹窗改为上下堆叠 */
  .modal-split-layout { flex-direction: column; height: 80vh; }
  .modal-left-detail { border-right: none; border-bottom: 8px solid #f1f5f9; padding: 16px; flex: auto; overflow-y: visible; }
  .modal-right-action { padding: 16px; flex: auto; }
  .sku-modal-card { width: 100% !important; max-width: 100vw !important; border-radius: 16px 16px 0 0 !important; margin: 0; position: absolute; bottom: 0;}
}
</style>