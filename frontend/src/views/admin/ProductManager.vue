<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NButton, NTag, NInput, NModal, useMessage, 
  NForm, NFormItem, NSelect, NGrid, NGi, NList, NListItem, NThing, 
  NPopconfirm, NIcon, NEmpty, NDivider, NSpin, NSwitch, NInputNumber, NSpace, NTooltip
} from 'naive-ui'
import { 
  CubeOutline, AddCircleOutline, TrashOutline, LinkOutline, 
  UnlinkOutline, CreateOutline, TimeOutline, PricetagOutline
} from '@vicons/ionicons5'
import request from '../../utils/request' // 确保路径正确

const message = useMessage()

// =======================
// 1. 数据定义
// =======================
const products = ref<any[]>([])
const pLoading = ref(false)
const currentProduct = ref<any>(null)

// 模态框状态
const showModal = ref(false)
const modalType = ref<'create' | 'edit'>('create')

// 表单模型 (SPU + SKU)
const formModel = reactive({
  id: 0,
  name: '',
  description: '',
  // SKU 列表
  skus: [] as Array<{ id?: number; name: string; price: number; duration_days: number }>
})

// 预设时长选项
const durationPresets = [
  { label: '7天', value: 7 },
  { label: '月卡', value: 30 },
  { label: '年卡', value: 365 },
  { label: '永久', value: -1 }
]

// =======================
// 2. 商品管理逻辑 (CRUD)
// =======================

// 获取商品列表
const fetchProducts = async () => {
  pLoading.value = true
  try {
    const res: any = await request.get('/admin/products')
    products.value = res.data || []
    
    // 刷新选中态的数据
    if (currentProduct.value) {
      // 🔥 核心修正：后端返回的是大写 ID
      const fresh = products.value.find(p => p.ID === currentProduct.value.ID) 
      if (fresh) currentProduct.value = fresh
    }
  } catch { 
    message.error('加载商品列表失败') 
  } finally { 
    pLoading.value = false 
  }
}

// 打开新建窗口
const openCreateModal = () => {
  modalType.value = 'create'
  formModel.id = 0
  formModel.name = ''
  formModel.description = ''
  formModel.skus = [{ name: '月卡', price: 29.9, duration_days: 30 }]
  showModal.value = true
}

// 打开编辑窗口
const openEditModal = (p: any) => {
  modalType.value = 'edit'
  // 🔥 核心修正：使用 p.ID (大写)
  formModel.id = p.ID 
  formModel.name = p.name
  formModel.description = p.description
  
  // 🔥 核心修正：使用 p.skus (小写)
  if (p.skus && p.skus.length > 0) {
      // 映射：SKU 里的 ID 也是大写 ID
      formModel.skus = p.skus.map((s: any) => ({
          id: s.ID, // 这里要把后端的 ID 映射给表单的 id
          name: s.name,
          price: s.price,
          duration_days: s.duration_days
      }))
  } else {
      formModel.skus = []
  }
  
  if (formModel.skus.length === 0) {
    formModel.skus.push({ name: '标准版', price: 0, duration_days: 30 })
  }
  showModal.value = true
}

// SKU 操作
const addSkuRow = () => {
  formModel.skus.push({ name: '', price: 0, duration_days: 30 })
}
const removeSkuRow = (index: number) => {
  formModel.skus.splice(index, 1)
}

// 提交保存 (新建或更新)
const handleSave = async () => {
  if (!formModel.name) return message.warning('请输入商品名称')
  if (formModel.skus.length === 0) return message.warning('至少需要一个规格')
  
  for (const s of formModel.skus) {
    if (!s.name) return message.warning('规格名称不能为空')
  }

  // 构造 payload
  const payload = {
      name: formModel.name,
      description: formModel.description,
      skus: formModel.skus.map(s => ({
          id: s.id || 0, // 新增的没有ID，传0
          name: s.name,
          price: s.price,
          duration_days: s.duration_days
      }))
  }

  try {
    if (modalType.value === 'create') {
      await request.post('/admin/products', payload)
      message.success('创建成功')
    } else {
      // 🔥 核心修正：URL使用 formModel.id (这是我们在 openEditModal 里赋值的)
      await request.put(`/admin/products/${formModel.id}`, payload)
      message.success('更新成功')
    }
    showModal.value = false
    fetchProducts()
  } catch { 
    message.error('保存失败') 
  }
}

// 切换上下架
const toggleShelf = async (p: any, val: boolean) => {
  p.is_on_shelf = val
  try {
    // 🔥 核心修正：使用 p.ID
    await request.put(`/admin/products/${p.ID}`, { is_on_shelf: val })
    message.success(val ? '已上架' : '已下架 (暂停售卖)')
  } catch {
    p.is_on_shelf = !val
    message.error('操作失败')
  }
}

// 硬删除
const handleDeleteProduct = async (id: number) => {
  try {
    await request.delete(`/admin/products/${id}`)
    message.success('商品已彻底删除')
    // 🔥 核心修正：使用 currentProduct.value.ID
    if (currentProduct.value?.ID === id) currentProduct.value = null
    fetchProducts()
  } catch { 
    message.error('删除失败') 
  }
}

const selectProduct = (p: any) => {
  currentProduct.value = p
  // 🔥 核心修正：使用 p.ID
  fetchContents(p.ID)
}

// =======================
// 3. 内容绑定逻辑
// =======================
const contents = ref<any[]>([])
const cLoading = ref(false)
const bindModal = ref(false)
const bindForm = reactive({ source: null, category: null })
const sourceOptions = ref<any[]>([])
const categoryOptions = ref<any[]>([])

const fetchContents = async (pid: number) => {
  cLoading.value = true
  try {
    const res: any = await request.get(`/admin/products/${pid}/contents`)
    contents.value = res.data || []
  } catch { 
  } finally { 
    cLoading.value = false 
  }
}

const openBindModal = async () => {
  bindModal.value = true
  bindForm.source = null; bindForm.category = null
  try {
    const res: any = await request.get('/banks')
    sourceOptions.value = (res.data || []).map((s: string) => ({ label: s, value: s }))
  } catch {}
}

const handleSourceChange = async (val: string) => {
  bindForm.source = val as any
  bindForm.category = null
  try {
    const res: any = await request.get('/category-tree', { params: { source: val, parent_id: 0 } })
    categoryOptions.value = (res.data || []).map((c:any) => ({ label: c.name, value: c.name }))
  } catch {}
}

const handleBind = async () => {
  if (!bindForm.source || !bindForm.category) return message.warning('请选择完整')
  try {
    await request.post('/admin/products/bind', {
      product_id: currentProduct.value.ID, // 🔥 核心修正：使用 ID
      source: bindForm.source,
      category: bindForm.category
    })
    message.success('绑定成功')
    bindModal.value = false
    fetchContents(currentProduct.value.ID) // 🔥 核心修正：使用 ID
  } catch { message.error('绑定失败，可能已存在') }
}

const handleUnbind = async (row: any) => {
  try {
    await request.post('/admin/products/unbind', {
      product_id: row.product_id,
      source: row.source,
      category: row.category
    })
    message.success('已移除')
    fetchContents(currentProduct.value.ID) // 🔥 核心修正：使用 ID
  } catch { message.error('解绑失败') }
}

onMounted(fetchProducts)
</script>

<template>
  <div class="product-manager">
    <n-grid :x-gap="24" :cols="24" style="height: 100%">
      
      <n-gi :span="9">
        <n-card title="🛍️ 商品管理 (SPU)" class="full-height" content-style="padding: 0;">
          <template #header-extra>
            <n-button size="small" type="primary" @click="openCreateModal">
              <template #icon><n-icon><AddCircleOutline/></n-icon></template> 发布商品
            </n-button>
          </template>
          
          <n-spin :show="pLoading">
            <n-list hoverable clickable>
              <n-list-item 
                v-for="p in products" 
                :key="p.ID" 
                @click="selectProduct(p)" 
                :class="{ active: currentProduct?.ID === p.ID }"
              >
                <n-thing>
                  <template #header>
                    <div style="display: flex; align-items: center; gap: 8px;">
                      {{ p.name }}
                      <n-tag v-if="!p.is_on_shelf" type="error" size="small" round :bordered="false">已下架</n-tag>
                    </div>
                  </template>
                  <template #description>
                    <div class="list-desc">{{ p.description || '暂无描述' }}</div>
                  </template>
                  
                  <template #footer>
                     <n-space size="small">
                        <n-tag v-for="sku in p.skus" :key="sku.ID" size="tiny" :bordered="false" type="info">
                           {{ sku.name }}
                        </n-tag>
                     </n-space>
                  </template>

                  <template #header-extra>
                    <n-space align="center">
                      <n-tooltip trigger="hover">
                        <template #trigger>
                           <n-switch 
                              size="small" 
                              :value="p.is_on_shelf" 
                              @update:value="(v) => toggleShelf(p, v)"
                              @click.stop
                           />
                        </template>
                        {{ p.is_on_shelf ? '销售中' : '已下架' }}
                      </n-tooltip>

                      <n-button size="tiny" secondary circle type="info" @click.stop="openEditModal(p)">
                        <n-icon><CreateOutline/></n-icon>
                      </n-button>

                      <n-popconfirm @positive-click="handleDeleteProduct(p.ID)">
                        <template #trigger>
                          <n-button size="tiny" secondary circle type="error" @click.stop>
                             <n-icon><TrashOutline/></n-icon>
                          </n-button>
                        </template>
                        ⚠️ 警告：物理删除！<br>
                        建议使用“下架”功能。<br>
                        确定要彻底销毁该商品及其所有规格吗？
                      </n-popconfirm>
                    </n-space>
                  </template>
                </n-thing>
              </n-list-item>
              <n-empty v-if="products.length===0" description="暂无商品" style="padding: 40px" />
            </n-list>
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi :span="15">
        <n-card class="full-height">
          <div v-if="currentProduct">
            <div class="header-area">
              <div class="title-group">
                 <h2>{{ currentProduct.name }}</h2>
                 <n-tag :type="currentProduct.is_on_shelf ? 'success' : 'error'">
                    {{ currentProduct.is_on_shelf ? '销售中' : '已下架' }}
                 </n-tag>
              </div>
              <p class="desc-text">{{ currentProduct.description }}</p>
            </div>
            
            <n-divider />

            <div class="section-title">
               <n-icon><PricetagOutline/></n-icon> 售卖规格 (SKU)
            </div>
            <div class="sku-grid-view">
               <div v-for="sku in currentProduct.skus" :key="sku.ID" class="sku-card">
                  <div class="sku-name">{{ sku.name }}</div>
                  <div class="sku-price">¥{{ sku.price }}</div>
                  <div class="sku-days">
                     <n-icon><TimeOutline/></n-icon> 
                     {{ sku.duration_days === -1 ? '永久有效' : `${sku.duration_days} 天` }}
                  </div>
               </div>
               <div v-if="!currentProduct.skus?.length" class="no-data">暂无规格配置</div>
            </div>

            <n-divider />

            <div class="header">
              <div class="section-title">
                 <n-icon><LinkOutline/></n-icon> 包含题库内容
              </div>
              <n-button size="small" type="primary" secondary @click="openBindModal">
                <template #icon><n-icon><AddCircleOutline/></n-icon></template> 添加绑定
              </n-button>
            </div>

            <n-spin :show="cLoading">
              <div class="content-grid">
                <div v-for="c in contents" :key="c.ID" class="content-item">
                  <div class="tag-source">{{ c.source }}</div>
                  <div class="tag-cat">{{ c.category }}</div>
                  <div class="action">
                    <n-popconfirm @positive-click="handleUnbind(c)">
                      <template #trigger>
                        <n-button circle size="tiny" type="error" secondary>
                          <n-icon><UnlinkOutline/></n-icon>
                        </n-button>
                      </template>
                      确定移除该科目？用户将失去访问权限。
                    </n-popconfirm>
                  </div>
                </div>
              </div>
              <n-empty v-if="contents.length===0" description="⚠️ 这是一个空壳商品，用户购买后没有任何题库权限。" style="margin-top:20px" />
            </n-spin>
          </div>

          <div v-else class="empty-placeholder">
            <n-icon size="60" color="#e0e0e0"><CubeOutline/></n-icon>
            <p>请点击左侧商品进行管理</p>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-modal v-model:show="showModal" preset="card" :title="modalType==='create'?'发布新商品':'编辑商品'" style="width: 700px">
      <n-form label-placement="left" label-width="80">
        <n-grid :cols="24" :x-gap="24">
           <n-gi :span="24">
              <n-form-item label="商品名称">
                 <n-input v-model:value="formModel.name" placeholder="例如：高考数学冲刺包" />
              </n-form-item>
           </n-gi>
           <n-gi :span="24">
              <n-form-item label="描述备注">
                 <n-input v-model:value="formModel.description" type="textarea" placeholder="后台备注" />
              </n-form-item>
           </n-gi>
        </n-grid>

        <n-divider title-placement="left" style="margin: 10px 0 20px 0; font-size: 12px; color: #999;">
           规格设置 (SKU)
        </n-divider>

        <div class="sku-editor">
           <div v-for="(sku, idx) in formModel.skus" :key="idx" class="sku-row">
              <n-grid :cols="24" :x-gap="10" align-items="center">
                 <n-gi :span="6">
                    <n-input v-model:value="sku.name" placeholder="规格名 (如:月卡)" size="small" />
                 </n-gi>
                 <n-gi :span="5">
                    <n-input-number v-model:value="sku.price" :precision="2" placeholder="价格" size="small">
                       <template #prefix>¥</template>
                    </n-input-number>
                 </n-gi>
                 <n-gi :span="11">
                    <n-input-number v-model:value="sku.duration_days" placeholder="天数" size="small" style="width: 100%">
                       <template #suffix>天</template>
                    </n-input-number>
                    <div style="margin-top: 4px; display: flex; gap: 4px;">
                       <n-tag 
                          v-for="opt in durationPresets" :key="opt.label" 
                          size="tiny" checkable 
                          :checked="sku.duration_days === opt.value"
                          @update:checked="() => sku.duration_days = opt.value"
                       >
                          {{ opt.label }}
                       </n-tag>
                    </div>
                 </n-gi>
                 <n-gi :span="2" style="text-align: right">
                    <n-button circle size="tiny" type="error" secondary @click="removeSkuRow(idx)">
                       <n-icon><TrashOutline/></n-icon>
                    </n-button>
                 </n-gi>
              </n-grid>
           </div>
           <n-button dashed block size="small" @click="addSkuRow" style="margin-top: 10px">
              <template #icon><n-icon><AddCircleOutline/></n-icon></template> 添加规格
           </n-button>
        </div>
      </n-form>
      <template #footer>
         <n-space justify="end">
            <n-button @click="showModal=false">取消</n-button>
            <n-button type="primary" @click="handleSave">保存提交</n-button>
         </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="bindModal" preset="card" title="添加绑定" style="width: 500px">
      <n-alert type="info" :bordered="false" style="margin-bottom: 15px">
        绑定后，持有该商品的用户将立即获得该科目的访问权限。
      </n-alert>
      <n-form>
        <n-form-item label="1. 选择题库源">
          <n-select v-model:value="bindForm.source" :options="sourceOptions" @update:value="handleSourceChange" placeholder="选择年份版本" />
        </n-form-item>
        <n-form-item label="2. 选择科目 (一级目录)">
          <n-select v-model:value="bindForm.category" :options="categoryOptions" :disabled="!bindForm.source" placeholder="选择科目" />
        </n-form-item>
      </n-form>
      <template #footer><n-button type="primary" block @click="handleBind">确认绑定</n-button></template>
    </n-modal>
  </div>
</template>

<style scoped>
.product-manager { height: calc(100vh - 100px); }
.full-height { height: 100%; display: flex; flex-direction: column; }
.active { background-color: #f0fdf4; border-right: 3px solid #18a058; }
.list-desc { font-size: 12px; color: #999; margin-bottom: 5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.empty-placeholder { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #ccc; }

.header-area { padding-bottom: 10px; }
.title-group { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.title-group h2 { margin: 0; font-size: 1.2rem; }
.desc-text { color: #666; font-size: 13px; }

.section-title { font-weight: bold; color: #333; display: flex; align-items: center; gap: 6px; margin-bottom: 10px; font-size: 14px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }

/* SKU Grid View */
.sku-grid-view { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 12px; margin-bottom: 20px; }
.sku-card { border: 1px solid #eee; padding: 12px; border-radius: 8px; background: #fafafa; }
.sku-name { font-weight: bold; font-size: 15px; margin-bottom: 4px; }
.sku-price { color: #d03050; font-size: 16px; font-weight: bold; margin-bottom: 4px; }
.sku-days { font-size: 12px; color: #666; display: flex; align-items: center; gap: 4px; }
.no-data { color: #999; font-size: 12px; padding: 10px; font-style: italic; }

/* Content Grid */
.content-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; }
.content-item { border: 1px solid #eef2f6; padding: 12px; border-radius: 6px; display: flex; align-items: center; justify-content: space-between; background: #fff; box-shadow: 0 1px 2px rgba(0,0,0,0.03); }
.tag-source { font-size: 12px; color: #999; margin-right: 8px; }
.tag-cat { font-weight: bold; font-size: 14px; color: #333; flex: 1; }

/* SKU Editor Modal */
.sku-editor { background: #f9f9f9; padding: 10px; border-radius: 6px; border: 1px dashed #ddd; max-height: 300px; overflow-y: auto; }
.sku-row { background: #fff; padding: 10px; margin-bottom: 8px; border-radius: 4px; border: 1px solid #eee; }
</style>