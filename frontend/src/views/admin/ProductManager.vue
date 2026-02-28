<script setup lang="ts">
import { ref, onMounted, reactive, shallowRef, onBeforeUnmount } from 'vue'
import { 
  NCard, NButton, NTag, NInput, NModal, useMessage, 
  NForm, NFormItem, NGrid, NGi, NList, NListItem, NThing, 
  NPopconfirm, NIcon, NEmpty, NDivider, NSpin, NSwitch, NInputNumber, 
  NSpace, NTooltip, NSelect, NAlert, NTabs, NTabPane, NDynamicTags, NInputGroup
} from 'naive-ui'
import { 
  CubeOutline, AddCircleOutline, TrashOutline, LinkOutline, 
  UnlinkOutline, CreateOutline, TimeOutline, PricetagOutline,
  WalletOutline, ImageOutline, DocumentTextOutline, CloudUploadOutline,
  FlameOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

// 🔥 引入 vue-cropper 及其样式
import 'vue-cropper/dist/index.css'
import { VueCropper } from 'vue-cropper'

// 🔥 引入 WangEditor 组件与样式
import '@wangeditor/editor/dist/css/style.css'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'

const message = useMessage()

// =======================
// 🎨 富文本编辑器配置 (WangEditor)
// =======================
const editorRef = shallowRef() // 编辑器实例，必须用 shallowRef
const toolbarConfig = {}
const editorConfig = { 
    placeholder: '请输入商品详细介绍内容...',
    MENU_CONF: {
        // 配置图片上传，对接现有的封面上传接口
        uploadImage: {
            server: 'http://localhost:8080/api/v1/admin/products/upload', 
            fieldName: 'file', 
            maxFileSize: 5 * 1024 * 1024,
            // 自定义插入图片，适配后端返回的格式
            customInsert(res: any, insertFn: any) {
                const url = res.url || res.data?.url || res.data
                if (url) {
                    const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`
                    insertFn(fullUrl, '商品详情图', fullUrl)
                } else {
                    message.error('图片插入失败，未获取到链接')
                }
            }
        }
    }
}

// 组件销毁时，也及时销毁编辑器
onBeforeUnmount(() => {
    const editor = editorRef.value
    if (editor == null) return
    editor.destroy()
})

const handleCreated = (editor: any) => {
  editorRef.value = editor
}

// =======================
// 1. 数据定义
// =======================
const products = ref<any[]>([])
const pLoading = ref(false)
const currentProduct = ref<any>(null)

// 模态框状态
const showModal = ref(false)
const modalType = ref<'create' | 'edit'>('create')

// 表单模型
const formModel = reactive({
  id: 0,
  name: '',
  description: '',
  cover_img: '',
  category: '',
  tags: [] as string[],
  detail: '',
  skus: [] as Array<{ id?: number; name: string; points: number; duration_days: number }>
})

const durationPresets = [
  { label: '7天', value: 7 },
  { label: '月卡', value: 30 },
  { label: '年卡', value: 365 },
  { label: '永久', value: -1 }
]

// =======================
// 🎨 图片裁剪专区
// =======================
const fileInput = ref<HTMLInputElement | null>(null)
const showCropperModal = ref(false)
const cropperImg = ref('')
const cropperRef = ref<any>(null)
const uploading = ref(false)

const triggerUpload = () => {
  fileInput.value?.click()
}

const onFileSelected = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    if (!file.type.includes('image/')) {
       message.error('请选择图片文件')
       return
    }
    cropperImg.value = URL.createObjectURL(file)
    showCropperModal.value = true
    target.value = '' 
  }
}

const confirmCrop = () => {
  if (!cropperRef.value) return
  uploading.value = true

  cropperRef.value.getCropBlob(async (data: Blob) => {
    try {
      const formData = new FormData()
      formData.append('file', data, `cover_${Date.now()}.jpg`)

      const res: any = await request.post('/admin/products/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      
      const finalUrl = res.url || res.data?.url || res.data
      if (finalUrl) {
          formModel.cover_img = finalUrl
          message.success('封面图裁剪并上传成功！')
          showCropperModal.value = false
      } else {
          throw new Error('服务器未返回图片URL')
      }
    } catch (err: any) {
      message.error(err.response?.data?.error || '上传失败，请检查后端接口')
    } finally {
      uploading.value = false
    }
  })
}

// =======================
// 2. 商品管理逻辑 (CRUD)
// =======================

const fetchProducts = async () => {
  pLoading.value = true
  try {
    const res: any = await request.get('/market/products', { params: { admin: 1 } })
    products.value = res.data || []
    
    if (currentProduct.value) {
      const fresh = products.value.find((p: any) => p.ID === currentProduct.value.ID) 
      if (fresh) currentProduct.value = fresh
    }
  } catch { 
    message.error('加载商品列表失败') 
  } finally { 
    pLoading.value = false 
  }
}

const getCoverUrl = (url: string | undefined) => {
  if (!url) return 'https://images.unsplash.com/photo-1606326608606-aa0b62935f2b?q=80&w=800&auto=format&fit=crop'
  return url.startsWith('http') ? url : `http://localhost:8080${url}`
}

const openCreateModal = () => {
  modalType.value = 'create'
  formModel.id = 0
  formModel.name = ''
  formModel.description = ''
  formModel.cover_img = ''
  formModel.category = ''
  formModel.tags = []
  formModel.detail = ''
  formModel.skus = [{ name: '月卡', points: 300, duration_days: 30 }]
  showModal.value = true
}

const openEditModal = async (p: any) => {
  modalType.value = 'edit'
  formModel.id = p.ID 
  formModel.name = p.name
  formModel.description = p.description
  formModel.cover_img = p.cover_img || ''
  formModel.category = p.category || ''
  formModel.tags = p.tags ? p.tags.split(',').filter((t: string) => t.trim() !== '') : []
  
  try {
      const res: any = await request.get(`/market/products/${p.ID}`)
      formModel.detail = res.data?.detail || ''
  } catch {
      message.error('商品详细描述获取失败')
  }
  
  if (p.skus && p.skus.length > 0) {
      formModel.skus = p.skus.map((s: any) => ({
          id: s.ID, 
          name: s.name,
          points: s.points, 
          duration_days: s.duration_days
      }))
  } else {
      formModel.skus = [{ name: '标准版', points: 0, duration_days: 30 }]
  }
  showModal.value = true
}

const addSkuRow = () => formModel.skus.push({ name: '', points: 0, duration_days: 30 })
const removeSkuRow = (index: number) => formModel.skus.splice(index, 1)

const handleSave = async () => {
  if (!formModel.name) return message.warning('请输入商品名称')
  if (formModel.skus.length === 0) return message.warning('至少需要一个规格')
  
  for (const s of formModel.skus) {
    if (!s.name) return message.warning('规格名称不能为空')
    if (s.points < 0) return message.warning('积分不能为负数，已拦截！')
  }

  const payload = {
      name: formModel.name,
      description: formModel.description,
      cover_img: formModel.cover_img,
      category: formModel.category,
      tags: formModel.tags.join(','),
      detail: formModel.detail,
      skus: formModel.skus.map(s => ({
          id: s.id || 0, 
          name: s.name,
          points: s.points, 
          duration_days: s.duration_days
      }))
  }

  try {
    if (modalType.value === 'create') {
      await request.post('/admin/products', payload)
      message.success('创建成功')
    } else {
      await request.put(`/admin/products/${formModel.id}`, payload)
      message.success('更新成功')
    }
    showModal.value = false
    fetchProducts()
  } catch (e: any) { 
    message.error(e.response?.data?.error || '保存失败') 
  }
}

const toggleShelf = async (p: any, val: boolean) => {
  p.is_on_shelf = val
  try {
    await request.put(`/admin/products/${p.ID}`, { is_on_shelf: val })
    message.success(val ? '已上架' : '已下架 (暂停兑换)')
  } catch {
    p.is_on_shelf = !val
    message.error('操作失败')
  }
}

const handleDeleteProduct = async (id: number) => {
  try {
    await request.delete(`/admin/products/${id}`)
    message.success('商品已删除')
    if (currentProduct.value?.ID === id) currentProduct.value = null
    fetchProducts()
  } catch { 
    message.error('删除失败') 
  }
}

const selectProduct = (p: any) => {
  currentProduct.value = p
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
  } catch { } finally { cLoading.value = false }
}

const openBindModal = async () => {
  bindModal.value = true
  bindForm.source = null
  bindForm.category = null
  sourceOptions.value = []
  categoryOptions.value = []
  try {
    const res: any = await request.get('/banks')
    const list = res.data || []
    if (list.length > 0) {
        sourceOptions.value = list.map((s: string) => ({ label: s, value: s }))
    } else {
        message.warning('未检测到题库源数据')
    }
  } catch (e) { message.error('加载题库源失败') }
}

const handleSourceChange = async (val: string) => {
  bindForm.category = null
  categoryOptions.value = [] 
  if (!val) return
  try {
    message.loading('正在加载科目...', { duration: 1000 })
    const res: any = await request.get('/category-tree', { params: { source: val } })
    const list = res.data || []
    categoryOptions.value = list.map((c: any) => ({ label: c.name || c, value: c.name || c }))
  } catch (e) { message.error('加载科目失败') }
}

const handleBind = async () => {
  if (!bindForm.source || !bindForm.category) return message.warning('请选择完整')
  try {
    await request.post('/admin/products/bind', {
      product_id: currentProduct.value.ID, 
      source: bindForm.source,
      category: bindForm.category
    })
    message.success('绑定成功')
    bindModal.value = false
    fetchContents(currentProduct.value.ID) 
  } catch { message.error('绑定失败，可能已存在') }
}

const handleUnbind = async (row: any) => {
  try {
    await request.post('/admin/products/unbind', {
      product_id: row.product_id, source: row.source, category: row.category
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
      
      <n-gi :span="9">
        <n-card title="🛍️ 商品管理" class="full-height" content-style="padding: 0;">
          <template #header-extra>
            <n-button size="small" type="primary" @click="openCreateModal">
              <template #icon><n-icon><AddCircleOutline/></n-icon></template> 发布新商品
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
                    <div class="list-desc">{{ p.category ? `[${p.category}] ` : '' }}{{ p.description || '暂无描述' }}</div>
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
                      <n-switch 
                          size="small" 
                          :value="p.is_on_shelf" 
                          @update:value="(v) => toggleShelf(p, v)"
                          @click.stop
                      />
                      <n-button size="tiny" secondary circle type="info" @click.stop="openEditModal(p)">
                        <n-icon><CreateOutline/></n-icon>
                      </n-button>

                      <n-popconfirm @positive-click="handleDeleteProduct(p.ID)">
                        <template #trigger>
                          <n-button size="tiny" secondary circle type="error" @click.stop>
                             <n-icon><TrashOutline/></n-icon>
                          </n-button>
                        </template>
                        确定删除该商品吗？(用户已购权益不受影响)
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
        <n-card class="full-height right-panel">
          <div v-if="currentProduct" class="preview-container">
            
            <div class="product-preview-hero">
                <img :src="getCoverUrl(currentProduct.cover_img)" class="hero-cover" />
                <div class="hero-info">
                   <div class="hero-tags">
                      <n-tag type="primary" size="small" v-if="currentProduct.category">{{ currentProduct.category }}</n-tag>
                      <n-tag type="warning" size="small" v-for="tag in (currentProduct.tags ? currentProduct.tags.split(',') : [])" :key="tag">
                         <template #icon><n-icon><FlameOutline/></n-icon></template> {{ tag }}
                      </n-tag>
                   </div>
                   <h2 class="hero-title">{{ currentProduct.name }}</h2>
                   <p class="hero-desc">{{ currentProduct.description || '未填写简介' }}</p>
                </div>
            </div>
            
            <n-divider />

            <div class="section-title">
               <n-icon><PricetagOutline/></n-icon> 兑换规格 (SKU)
            </div>
            <div class="sku-grid-view">
               <div v-for="sku in currentProduct.skus" :key="sku.ID" class="sku-card">
                  <div class="sku-name">{{ sku.name }}</div>
                  <div class="sku-price">
                      <n-icon><WalletOutline/></n-icon> {{ sku.points === 0 ? '免费' : sku.points + ' 积分' }}
                  </div>
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
                 <n-icon><LinkOutline/></n-icon> 包含权益内容 (解锁权限)
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
                      确定移除该题库的解锁权限？
                    </n-popconfirm>
                  </div>
                </div>
              </div>
              <n-empty v-if="contents.length===0" description="⚠️ 这是一个空壳商品，用户兑换后没有任何系统权限。" style="margin-top:20px" />
            </n-spin>
          </div>

          <div v-else class="empty-placeholder">
            <n-icon size="60" color="#e0e0e0"><CubeOutline/></n-icon>
            <p>请点击左侧商品查看详情</p>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <n-modal v-model:show="showModal" preset="card" :title="modalType==='create'?'发布商品':'编辑商品'" style="width: 900px">
      <n-tabs type="line" animated>
        
        <n-tab-pane name="basic" tab="基础设置">
          <n-form label-placement="left" label-width="90">
            <n-form-item label="商品名称" required>
                <n-input v-model:value="formModel.name" placeholder="例如：2025考研英语核心题库" />
            </n-form-item>
            
            <n-form-item label="一句话简介">
                <n-input v-model:value="formModel.description" type="textarea" :rows="2" placeholder="展示在卡片列表副标题的位置" />
            </n-form-item>

            <n-grid :cols="2" :x-gap="24">
              <n-gi>
                <n-form-item label="商品分类">
                  <n-input v-model:value="formModel.category" placeholder="例如：押题包 / VIP会员" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="营销标签">
                  <n-dynamic-tags v-model:value="formModel.tags" placeholder="输入后按回车" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-form-item label="商品封面图">
              <n-input-group>
                <n-input v-model:value="formModel.cover_img" placeholder="可手动输入链接，或点击右侧上传" />
                <n-button type="primary" @click="triggerUpload" secondary>
                  <template #icon><n-icon><CloudUploadOutline /></n-icon></template>
                  上传并裁剪
                </n-button>
              </n-input-group>
              <input type="file" ref="fileInput" style="display: none" accept="image/*" @change="onFileSelected" />
            </n-form-item>
            <div style="font-size:12px; color:#999; margin: -10px 0 10px 90px;">推荐使用 16:9 的高质量封面图，直接上传会自动呼出裁剪工具。</div>

          </n-form>
        </n-tab-pane>

        <n-tab-pane name="detail" tab="图文详情">
           <n-alert type="info" :bordered="false" style="margin-bottom: 16px;">
              提示：支持直接从系统剪贴板粘贴图片，也支持通过鼠标拖拽修改图片大小。
           </n-alert>
           <div class="editor-wrapper">
              <Toolbar
                style="border-bottom: 1px solid #e2e8f0"
                :editor="editorRef"
                :defaultConfig="toolbarConfig"
                mode="default"
              />
              <Editor
                style="height: 400px; overflow-y: hidden;"
                v-model="formModel.detail"
                :defaultConfig="editorConfig"
                mode="default"
                @onCreated="handleCreated"
              />
           </div>
        </n-tab-pane>

        <n-tab-pane name="sku" tab="价格与规格 (SKU)">
          <div class="sku-editor">
            <div v-for="(sku, idx) in formModel.skus" :key="idx" class="sku-row">
                <n-grid :cols="24" :x-gap="10" align-items="center">
                  <n-gi :span="6">
                      <n-input v-model:value="sku.name" placeholder="规格名 (如: 月卡)" size="small" />
                  </n-gi>
                  <n-gi :span="6">
                      <n-input-number v-model:value="sku.points" :min="0" :show-button="false" placeholder="所需积分" size="small">
                        <template #suffix>分</template>
                      </n-input-number>
                  </n-gi>
                  <n-gi :span="10">
                      <n-input-number v-model:value="sku.duration_days" placeholder="有效期天数 (-1代表永久)" size="small" style="width: 100%">
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
                <template #icon><n-icon><AddCircleOutline/></n-icon></template> 新增规格
            </n-button>
          </div>
        </n-tab-pane>
      </n-tabs>

      <template #footer>
         <n-space justify="end">
            <n-button @click="showModal=false">取消</n-button>
            <n-button type="primary" @click="handleSave">确认提交</n-button>
         </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showCropperModal" preset="card" title="✂️ 调整商品封面图" style="width: 650px">
      <div style="height: 400px; width: 100%; background: #f8fafc; border-radius: 8px; overflow: hidden; border: 1px solid #e2e8f0;">
        <vue-cropper
          ref="cropperRef"
          :img="cropperImg"
          :autoCrop="true"
          :autoCropWidth="320"
          :autoCropHeight="180"
          :fixedBox="false"
          :fixed="true"
          :fixedNumber="[16, 9]"
          outputType="jpeg"
          :info="true"
          :canScale="true"
          :full="false"
        />
      </div>
      <div style="margin-top: 12px; color: #64748b; font-size: 13px;">
        <n-icon><ImageOutline /></n-icon> 请拖拽或滚动鼠标缩放图片，选框比例已锁定为 <b>16:9</b> 以保证最佳展示效果。
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCropperModal = false" :disabled="uploading">取消</n-button>
          <n-button type="primary" :loading="uploading" @click="confirmCrop">
            确认裁剪并上传
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="bindModal" preset="card" title="绑定权益" style="width: 450px">
      <n-form>
        <n-form-item label="选择题库源">
          <n-select v-model:value="bindForm.source" :options="sourceOptions" @update:value="handleSourceChange" />
        </n-form-item>
        <n-form-item label="选择绑定科目">
          <n-select v-model:value="bindForm.category" :options="categoryOptions" :disabled="!bindForm.source" />
        </n-form-item>
      </n-form>
      <template #footer><n-button type="primary" block @click="handleBind">提交绑定</n-button></template>
    </n-modal>
  </div>
</template>

<style scoped>
.product-manager { height: calc(100vh - 100px); }
.full-height { height: 100%; display: flex; flex-direction: column; }
.right-panel :deep(.n-card__content) { overflow-y: auto; padding-top: 0; }
.active { background-color: #f0fdf4; border-right: 3px solid #18a058; }
.list-desc { font-size: 12px; color: #999; margin-bottom: 5px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.empty-placeholder { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #ccc; }

.product-preview-hero { display: flex; gap: 20px; padding: 20px 0; }
.hero-cover { width: 180px; height: 110px; object-fit: cover; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.hero-info { flex: 1; display: flex; flex-direction: column; justify-content: center; }
.hero-tags { display: flex; gap: 8px; margin-bottom: 10px; flex-wrap: wrap; }
.hero-title { margin: 0 0 6px 0; font-size: 20px; font-weight: bold; color: #1e293b; }
.hero-desc { margin: 0; color: #64748b; font-size: 14px; line-height: 1.5; }

.section-title { font-weight: bold; color: #333; display: flex; align-items: center; gap: 6px; margin-bottom: 10px; font-size: 14px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }

.sku-grid-view { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 12px; margin-bottom: 20px; }
.sku-card { border: 1px solid #eee; padding: 12px; border-radius: 8px; background: #fafafa; }
.sku-name { font-weight: bold; font-size: 15px; margin-bottom: 4px; }
.sku-price { color: #d97706; font-size: 16px; font-weight: bold; margin-bottom: 4px; display: flex; align-items: center; gap: 4px; }
.sku-days { font-size: 12px; color: #666; display: flex; align-items: center; gap: 4px; }
.no-data { color: #999; font-size: 12px; padding: 10px; font-style: italic; }

.content-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; }
.content-item { border: 1px solid #eef2f6; padding: 12px; border-radius: 6px; display: flex; align-items: center; justify-content: space-between; background: #fff; }
.tag-source { font-size: 12px; color: #999; margin-right: 8px; }
.tag-cat { font-weight: bold; font-size: 14px; color: #333; flex: 1; }

.sku-editor { background: #f9f9f9; padding: 16px; border-radius: 8px; border: 1px dashed #ddd; }
.sku-row { background: #fff; padding: 12px; margin-bottom: 10px; border-radius: 6px; border: 1px solid #eee; box-shadow: 0 1px 3px rgba(0,0,0,0.02); }

/* 🔥 富文本编辑器外框样式 */
.editor-wrapper {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  margin-top: 10px;
  background: #fff;
  z-index: 100;
}
</style>