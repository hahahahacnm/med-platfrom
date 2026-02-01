<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NLayout, NLayoutHeader, NLayoutContent, NLayoutSider, NEmpty, NSpin, 
  useMessage, NButton, NSelect, NIcon, NPagination, NTag, NTree, NPageHeader
} from 'naive-ui'
import { 
  JournalOutline, MenuOutline, FilterOutline, SearchOutline, 
  LibraryOutline, BookOutline, StarOutline 
} from '@vicons/ionicons5'
import QuestionCard from '../components/QuestionCard.vue'
import request from '../utils/request'
import { useUserStore } from '../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

// =========================
// 1. 状态定义
// =========================
const loading = ref(false)
const list = ref<any[]>([]) 
const noteTree = ref<any[]>([])
const loadingTree = ref(false) 

const pagination = reactive({ page: 1, pageSize: 5, itemCount: 0 }) 
const filter = reactive({ 
  source: null as string | null, 
  keyword: '',
  category: '' 
})
const bankOptions = ref<any[]>([])

// =========================
// 2. 辅助函数
// =========================
const safeParse = (val: any) => {
  if (typeof val === 'string') { try { return JSON.parse(val) } catch(e) { return {} } }
  return val
}

// 🔥🔥🔥 核心修复：适配器增强 (修复 isLeaf 不显示问题) 🔥🔥🔥
const adaptTreeData = (list: any[], parentPath = '') => {
  return list.map(item => {
    // 1. 路径拼接逻辑
    let currentFull = item.full || item.full_path
    if (!currentFull) {
        currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name
    }

    // 2. isLeaf 兼容逻辑 (后端返回 isLeaf, 前端之前只读 is_leaf)
    let isLeafNode = false
    if (item.isLeaf !== undefined) isLeafNode = item.isLeaf
    else if (item.IsLeaf !== undefined) isLeafNode = item.IsLeaf
    else if (item.is_leaf !== undefined) isLeafNode = item.is_leaf

    return {
      key: item.id,
      // 后端返回的 label 包含了数量 "内科学 (5)"，优先使用
      label: item.label || item.name, 
      name: item.name,
      full: currentFull, 
      isLeaf: isLeafNode, 
      children: null 
    }
  })
}

// 递归查找 Label 用于 Tag 显示
const findLabelInTree = (nodes: any[], targetKey: string): string => {
  for (const node of nodes) {
    if (String(node.key) === targetKey) return node.label
    if (Array.isArray(node.children) && node.children.length > 0) {
      const found = findLabelInTree(node.children, targetKey)
      if (found) return found
    }
  }
  return ''
}

// =========================
// 3. 数据获取逻辑
// =========================

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    if (res.data) {
      bankOptions.value = res.data.map((b: string) => ({ label: b, value: b }))
      if (!filter.source && bankOptions.value.length > 0) {
        filter.source = bankOptions.value[0].value
        handleSourceChange()
      }
    }
  } catch (e) { console.error(e) }
}

// 🔥 1. 初始加载：只获取一级目录
const fetchRootTree = async () => {
  if (!filter.source) return
  loadingTree.value = true
  noteTree.value = [] 
  try {
    const res: any = await request.get('/notes/note-tree', { 
      params: { source: filter.source, parent_id: 0 } 
    })
    // 根目录，父路径为空
    noteTree.value = adaptTreeData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

// 🔥 2. 懒加载：点击箭头时，加载子节点
const handleLoad = async (node: any) => {
  return new Promise<void>(async (resolve) => {
    try {
      const res: any = await request.get('/notes/note-tree', { 
        params: { source: filter.source, parent_id: node.key } 
      })
      
      // 🔥 将当前节点的完整路径传给子节点
      const currentPath = node.full || node.name
      node.children = adaptTreeData(res.data || [], currentPath)
      
      resolve()
    } catch (e) {
      node.children = []
      resolve()
    }
  })
}

// 点击节点筛选
const handleNodeClick = (keys: any, option: any) => {
  if (option && option.length > 0) {
    const node = option[0]
    // 这里的 category 对应后端查询参数 category_id
    filter.category = String(node.key)
    pagination.page = 1
    fetchData()
  } else {
    filter.category = ''
    fetchData()
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/notes/my', {
      params: {
        page: pagination.page, 
        page_size: pagination.pageSize,
        category_id: filter.category,
        source: filter.source 
      }
    })
    
    list.value = (res.data || []).map((q: any) => {
      q.options = safeParse(q.options)
      if (q.children) {
           q.children.forEach((child: any) => child.options = safeParse(child.options))
      }
      return q
    })
    
    pagination.itemCount = res.total || 0
  } catch (e) { message.error('加载失败') } finally { loading.value = false }
}

const handleSourceChange = () => {
  filter.category = ''
  pagination.page = 1
  fetchRootTree() 
  fetchData()
}

const handlePageChange = (page: number) => { 
    pagination.page = page
    fetchData() 
    document.querySelector('.n-layout-content')?.scrollTo(0, 0)
}

const goBack = () => router.push('/')

onMounted(() => { fetchBanks() })
</script>

<template>
  <div class="notes-container">
    <div class="page-control-bar">
      <div class="left-controls">
        <h2 class="page-title">
          <n-icon color="#18a058" style="margin-right: 8px; vertical-align: bottom;"><JournalOutline /></n-icon>
          我的笔记本
        </h2>
        <div class="bank-selector">
          <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" size="small" />
        </div>
      </div>
      
      <div class="right-controls">
         <!-- Add search or filters here if needed later -->
      </div>
    </div>

    <n-layout has-sider class="main-layout-area">
      <n-layout-sider 
        bordered 
        collapse-mode="width" 
        :collapsed-width="0" 
        :width="260" 
        show-trigger="arrow-circle" 
        content-style="padding: 12px;" 
        style="background-color: #fafafa;"
      >
        <div style="font-weight: bold; color: #333; margin-bottom: 12px; padding-left: 8px; font-size: 14px; display: flex; align-items: center; gap: 6px;">
          <n-icon color="#18a058"><FilterOutline /></n-icon> 笔记分布
        </div>
        
        <n-spin :show="loadingTree">
          <n-tree
            block-line 
            expand-on-click 
            :data="noteTree" 
            key-field="key" 
            label-field="label" 
            selectable
            remote
            :on-load="handleLoad" 
            @update:selected-keys="handleNodeClick"
            style="font-size: 13px;"
          />
          <div v-if="noteTree.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 12px;">
            暂无笔记记录
          </div>
        </n-spin>
      </n-layout-sider>

      <n-layout-content content-style="padding: 24px; max-width: 960px; margin: 0 auto;" :native-scrollbar="true">
        
        <div v-if="filter.category" style="margin-bottom: 16px;">
            <n-tag closable type="success" @close="filter.category = ''; fetchData()">
              正在筛选: {{ findLabelInTree(noteTree, filter.category) || '当前章节' }}
            </n-tag>
        </div>

        <n-spin :show="loading">
          <div v-if="list.length > 0">
            <div v-for="(q, index) in list" :key="q.id" class="note-item-wrapper">
              
              <div class="note-toolbar">
                <div class="info-badges">
                  <n-tag type="success" size="small" :bordered="false" style="margin-right: 8px;">
                    记录 #{{ (pagination.page - 1) * pagination.pageSize + index + 1 }}
                  </n-tag>
                  <span style="font-size: 12px; color: #999;">
                      {{ q.source }} / {{ q.category_path || '未知章节' }}
                  </span>
                </div>
              </div>

              <QuestionCard 
                :question="q" 
                :serial-number="(pagination.page - 1) * pagination.pageSize + index + 1"
                :init-show-notes="true" 
              />

            </div>
            
            <div style="display: flex; justify-content: center; margin: 40px 0;">
                <n-pagination
                  v-model:page="pagination.page"
                  :item-count="pagination.itemCount"
                  :page-size="pagination.pageSize"
                  @update:page="handlePageChange"
                />
            </div>
          </div>

          <n-empty v-else-if="!loading" description="该题库下暂无笔记，去刷题吧！" style="margin-top: 100px">
            <template #extra>
              <n-button type="primary" @click="router.push('/home')">去刷题</n-button>
            </template>
          </n-empty>
        </n-spin>
      </n-layout-content>
    </n-layout>
  </div>
</template>

<style scoped>
.notes-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fff;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  border: 1px solid #e2e8f0;
}

.page-control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: #fff;
  border-bottom: 1px solid #f0f0f0;
}

.left-controls, .right-controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-title {
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
  display: flex;
  align-items: center;
}

.bank-selector { width: 150px; }
.main-layout-area { flex: 1; overflow: hidden; }

.note-item-wrapper { margin-bottom: 30px; }
.note-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; padding: 0 4px; }
.info-badges { display: flex; align-items: center; }
</style>