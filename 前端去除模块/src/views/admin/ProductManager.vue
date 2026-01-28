<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NButton, NTag, NInput, NModal, useMessage, 
  NForm, NFormItem, NSelect, NGrid, NGi, NList, NListItem, NThing, 
  NPopconfirm, NIcon, NEmpty, NDivider, NSpin, NAlert
} from 'naive-ui'
import { 
  CubeOutline, AddCircleOutline, TrashOutline, LinkOutline, 
  UnlinkOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()

// =======================
// 1. 商品列表逻辑 (左侧)
// =======================
const products = ref<any[]>([])
const pLoading = ref(false)
const currentProduct = ref<any>(null) // 当前选中的商品

const createModal = ref(false)
const createForm = reactive({ name: '', description: '' })

const fetchProducts = async () => {
  pLoading.value = true
  try {
    const res: any = await request.get('/admin/products')
    products.value = res.data || []
  } catch { message.error('加载商品列表失败') } 
  finally { pLoading.value = false }
}

const handleCreateProduct = async () => {
  if(!createForm.name) return message.warning('请输入商品名称')
  try {
    await request.post('/admin/products', createForm)
    message.success('创建成功')
    createModal.value = false
    createForm.name = ''; createForm.description = ''
    fetchProducts()
  } catch { message.error('创建失败') }
}

const handleDeleteProduct = async (id: number) => {
  try {
    await request.delete(`/admin/products/${id}`)
    message.success('商品已下架')
    if (currentProduct.value?.ID === id) currentProduct.value = null
    fetchProducts()
  } catch { message.error('删除失败') }
}

const selectProduct = (p: any) => {
  currentProduct.value = p
  fetchContents(p.ID) // 注意：Gorm Model 默认 ID 是大写
}

// =======================
// 2. 内容绑定逻辑 (右侧)
// =======================
const contents = ref<any[]>([])
const cLoading = ref(false)
const bindModal = ref(false)

// 绑定表单
const bindForm = reactive({ source: null, category: null })
const sourceOptions = ref<any[]>([])
const categoryOptions = ref<any[]>([])

const fetchContents = async (pid: number) => {
  cLoading.value = true
  try {
    const res: any = await request.get(`/admin/products/${pid}/contents`)
    contents.value = res.data || []
  } catch { message.error('加载内容失败') } 
  finally { cLoading.value = false }
}

// 1. 加载题库源
const openBindModal = async () => {
  bindModal.value = true
  bindForm.source = null; bindForm.category = null
  try {
    const res: any = await request.get('/banks')
    sourceOptions.value = (res.data || []).map((s: string) => ({ label: s, value: s }))
  } catch {}
}

// 2. 加载一级科目 (根据源)
const handleSourceChange = async (val: string) => {
  bindForm.source = val as any
  bindForm.category = null
  try {
    const res: any = await request.get('/category-tree', { params: { source: val, parent_id: 0 } })
    categoryOptions.value = (res.data || []).map((c:any) => ({ label: c.name, value: c.name }))
  } catch {}
}

// 3. 提交绑定
const handleBind = async () => {
  if (!bindForm.source || !bindForm.category) return message.warning('请选择完整')
  try {
    await request.post('/admin/products/bind', {
      product_id: currentProduct.value.ID, // ID 大写
      source: bindForm.source,
      category: bindForm.category
    })
    message.success('绑定成功')
    bindModal.value = false
    fetchContents(currentProduct.value.ID)
  } catch { message.error('绑定失败，可能已存在') }
}

// 4. 解绑 (🔥 修复点：这里也要改小写)
const handleUnbind = async (row: any) => {
  try {
    await request.post('/admin/products/unbind', {
      product_id: row.product_id, // 🔥 改为小写 (json tag)
      source: row.source,         // 🔥 改为小写
      category: row.category      // 🔥 改为小写
    })
    message.success('已移除')
    fetchContents(currentProduct.value.ID)
  } catch { message.error('解绑失败') }
}

onMounted(fetchProducts)
</script>

<template>
  <div class="product-manager">
    <n-grid :x-gap="24" :cols="24" style="height: 100%">
      
      <n-gi :span="8">
        <n-card title="📦 商品(身份)定义" class="full-height" content-style="padding: 0;">
          <template #header-extra>
            <n-button size="small" type="primary" dashed @click="createModal=true">
              <template #icon><n-icon><AddCircleOutline/></n-icon></template> 新建
            </n-button>
          </template>
          
          <n-spin :show="pLoading">
            <n-list hoverable clickable>
              <n-list-item v-for="p in products" :key="p.ID" @click="selectProduct(p)" :class="{ active: currentProduct?.ID === p.ID }">
                <n-thing :title="p.name" :description="p.description || '暂无描述'">
                  <template #header-extra>
                    <n-popconfirm @positive-click="handleDeleteProduct(p.ID)">
                      <template #trigger>
                        <n-button size="tiny" type="error" text @click.stop>
                           <n-icon><TrashOutline/></n-icon>
                        </n-button>
                      </template>
                      确定下架删除该商品？用户的持有记录也会被清理。
                    </n-popconfirm>
                  </template>
                </n-thing>
              </n-list-item>
              <n-empty v-if="products.length===0" description="暂无商品" style="padding: 20px" />
            </n-list>
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi :span="16">
        <n-card class="full-height">
          <div v-if="currentProduct">
            <div class="header">
              <h3>
                <span style="color: #666">配置内容：</span>
                <span style="color: #18a058">{{ currentProduct.name }}</span>
              </h3>
              <n-button type="primary" @click="openBindModal">
                <template #icon><n-icon><LinkOutline/></n-icon></template> 添加科目绑定
              </n-button>
            </div>
            
            <n-divider />

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
                      确定将该科目从商品中移除？
                    </n-popconfirm>
                  </div>
                </div>
              </div>
              <n-empty v-if="contents.length===0" description="该商品暂未绑定任何题库内容，用户购买后将是空的。" />
            </n-spin>
          </div>

          <div v-else class="empty-placeholder">
            <n-icon size="60" color="#ddd"><CubeOutline/></n-icon>
            <p>请点击左侧商品进行配置</p>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-modal v-model:show="createModal" preset="card" title="新建商品" style="width: 400px">
      <n-form>
        <n-form-item label="商品名称"><n-input v-model:value="createForm.name" placeholder="例如：2025年儿科学单科" /></n-form-item>
        <n-form-item label="描述备注"><n-input v-model:value="createForm.description" type="textarea" placeholder="仅后台可见" /></n-form-item>
      </n-form>
      <template #footer><n-button type="primary" block @click="handleCreateProduct">创建</n-button></template>
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
.active { background-color: #e7f5ee; border-right: 3px solid #18a058; }
.empty-placeholder { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #ccc; }
.header { display: flex; justify-content: space-between; align-items: center; }
.content-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; }
.content-item { border: 1px solid #eee; padding: 12px; border-radius: 6px; display: flex; align-items: center; justify-content: space-between; background: #fafafa; }
.tag-source { font-size: 12px; color: #999; margin-right: 8px; }
.tag-cat { font-weight: bold; font-size: 15px; color: #333; flex: 1; }
</style>