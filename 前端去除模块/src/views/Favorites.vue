<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NLayout, NLayoutHeader, NLayoutContent, NLayoutSider, NEmpty, NSpin, 
  useMessage, NButton, NSelect, NInput, NIcon, NPagination, NTag, NAlert, NTree
} from 'naive-ui'
import { 
  HeartOutline, Heart, SearchOutline, TrashOutline, MenuOutline, FilterOutline, StarOutline 
} from '@vicons/ionicons5'
import QuestionCard from '../components/QuestionCard.vue'
import request from '../utils/request'

const router = useRouter()
const message = useMessage()

// =========================
// 1. 状态定义
// =========================
const loading = ref(false)
const list = ref<any[]>([])

// 🔥 收藏目录树 (懒加载模式)
const favTree = ref<any[]>([])
const loadingTree = ref(false) 

const pagination = reactive({ page: 1, pageSize: 5, itemCount: 0 })
const filter = reactive({ 
  source: null as string | null, 
  keyword: '',
  category: '' 
})
const bankOptions = ref<any[]>([])

// =========================
// 2. 数据获取逻辑
// =========================

// 🔥🔥🔥 核心适配器：自动拼接完整路径 🔥🔥🔥
const adaptTreeData = (list: any[], parentPath = '') => {
  return list.map(item => {
    // 优先用后端的 full，如果没有则前端拼接
    let currentFull = item.full
    if (!currentFull) {
        currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name
    }

    return {
      key: item.id,
      label: item.label || item.name, // 后端返回 label 带数量
      name: item.name,                // 原始名称
      full: currentFull,              // ✅ 完整路径
      isLeaf: item.isLeaf,
      children: null                  // 懒加载占位
    }
  })
}

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    if (res.data) {
      bankOptions.value = res.data.map((b: string) => ({ label: b, value: b }))
      if (!filter.source && bankOptions.value.length > 0) {
        filter.source = bankOptions.value[0].value
        // 初始加载
        handleSourceChange()
      }
    }
  } catch (e) { console.error(e) }
}

// 🔥 1. 初始加载：只获取一级目录
const fetchRootTree = async () => {
  if (!filter.source) return
  loadingTree.value = true
  favTree.value = [] 
  try {
    const res: any = await request.get('/favorite-tree', { 
      params: { source: filter.source, parent_key: '' } 
    })
    // 根目录，父路径为空
    favTree.value = adaptTreeData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

// 🔥 2. 懒加载：点击箭头时，加载子节点
const handleLoad = async (node: any) => {
  return new Promise<void>(async (resolve) => {
    try {
      const res: any = await request.get('/favorite-tree', { 
        params: { source: filter.source, parent_key: node.key } // 后端接收 parent_key (兼容 parent_id)
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
    // 🔥 使用拼接好的 full 发送请求
    filter.category = node.full || node.name
    pagination.page = 1
    fetchData()
  } else {
    filter.category = ''
    fetchData()
  }
}

const safeParse = (val: any) => {
  if (typeof val === 'string') { try { return JSON.parse(val) } catch(e) { return {} } }
  return val
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/favorites', {
      params: {
        page: pagination.page, 
        page_size: pagination.pageSize,
        source: filter.source, 
        keyword: filter.keyword, 
        category: filter.category
      }
    })
    
    list.value = (res.data || []).map((item: any) => {
      if (item.question) {
        item.question.options = safeParse(item.question.options)
        if (item.question.children) {
          item.question.children.forEach((child: any) => child.options = safeParse(child.options))
        }
      }
      return item
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

const handleSearch = () => {
  pagination.page = 1
  fetchData()
}

const handlePageChange = (page: number) => { pagination.page = page; fetchData() }

// =========================
// 3. 交互操作
// =========================

const handleRemove = async (questionId: number) => {
  try {
    // 调用 Toggle 接口取消收藏
    await request.post(`/favorites/${questionId}`)
    
    // 列表移除
    list.value = list.value.filter(item => item.question.id !== questionId)
    pagination.itemCount-- 
    
    if (list.value.length === 0) { 
      fetchData()
    }
    message.success('已取消收藏')
  } catch (e) { message.error('操作失败') }
}

const onAnswerResult = (payload: { questionId: number, isCorrect: boolean }) => {
  if(payload.isCorrect) {
      message.success("回答正确！🎉")
  }
}

const goBack = () => router.push('/')
const isValidQuestion = (q: any) => q && ( (q.options && Object.keys(q.options).length > 0) || (q.children && q.children.length > 0) )

onMounted(() => { fetchBanks() })
</script>

<template>
  <div class="favorites-container">
    <n-layout style="height: 100vh;">
      
      <n-layout-header bordered style="padding: 12px 24px; display: flex; justify-content: space-between; align-items: center; height: 64px;">
        <div style="display: flex; align-items: center; gap: 16px;">
          <n-button quaternary circle @click="goBack"><template #icon><n-icon><MenuOutline /></n-icon></template></n-button>
          <div style="font-size: 18px; font-weight: bold; color: #f0a020; display: flex; align-items: center; gap: 8px;">
            <n-icon><StarOutline /></n-icon> 我的收藏夹
          </div>
          <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" style="width: 180px" size="small" />
        </div>
        
        <div style="display: flex; align-items: center; gap: 16px;">
          <n-input v-model:value="filter.keyword" placeholder="搜索收藏题目..." size="small" style="width: 200px" @keydown.enter="handleSearch" clearable>
            <template #prefix><n-icon><SearchOutline /></n-icon></template>
          </n-input>
          <n-button type="warning" ghost size="small" @click="handleSearch">搜索</n-button>
        </div>
      </n-layout-header>

      <n-layout has-sider position="absolute" style="top: 64px; bottom: 0;">
        
        <n-layout-sider bordered collapse-mode="width" :collapsed-width="0" :width="260" show-trigger="arrow-circle" content-style="padding: 12px;" style="background-color: #fafafa;">
          <div style="font-weight: bold; color: #333; margin-bottom: 12px; padding-left: 8px; font-size: 14px; display: flex; align-items: center; gap: 6px;">
            <n-icon color="#f0a020"><FilterOutline /></n-icon> 收藏分布 ({{ pagination.itemCount }})
          </div>
          
          <n-spin :show="loadingTree">
            <n-tree
              block-line
              expand-on-click
              :data="favTree"
              key-field="key"
              label-field="label"
              selectable
              remote
              :on-load="handleLoad" 
              @update:selected-keys="handleNodeClick"
              style="font-size: 13px;"
            />
            <div v-if="favTree.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 12px;">
              当前暂无收藏记录
            </div>
          </n-spin>
        </n-layout-sider>

        <n-layout-content content-style="padding: 24px; max-width: 960px; margin: 0 auto;" :native-scrollbar="true">
          
          <div v-if="filter.category" style="margin-bottom: 16px;">
             <n-tag closable type="warning" @close="filter.category = ''; fetchData()">
               筛选: {{ filter.category }}
             </n-tag>
          </div>

          <n-spin :show="loading">
            <div v-if="list.length > 0">
              <div v-for="(item, index) in list" :key="item.id" class="fav-item-wrapper">
                
                <div class="fav-toolbar">
                  <div class="info-badges">
                    <n-tag type="warning" size="small" :bordered="false" style="margin-right: 8px;">
                      收藏 #{{ (pagination.page - 1) * pagination.pageSize + index + 1 }}
                    </n-tag>
                    <span style="font-size: 12px; color: #999;">
                      收藏于 {{ new Date(item.created_at).toLocaleDateString() }}
                    </span>
                  </div>
                  
                  <n-button size="tiny" type="warning" ghost @click="handleRemove(item.question.id)">
                     <template #icon><n-icon><Heart /></n-icon></template>
                     取消收藏
                  </n-button>
                </div>

                <QuestionCard 
                  v-if="isValidQuestion(item.question)"
                  :question="item.question" 
                  :serial-number="(pagination.page - 1) * pagination.pageSize + index + 1"
                  @answer-result="onAnswerResult"
                />

                <n-alert v-else type="warning" title="数据异常" style="margin-top: 10px;">
                  题目内容缺失 (ID: {{ item.question?.id }})
                </n-alert>

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

            <n-empty v-else-if="!loading" description="空空如也，快去收藏一些好题吧！" style="margin-top: 100px">
              <template #extra>
                <n-button type="primary" @click="filter.category = ''; fetchData()">查看全部</n-button>
              </template>
            </n-empty>
          </n-spin>
        </n-layout-content>
      </n-layout>
    </n-layout>
  </div>
</template>

<style scoped>
.fav-item-wrapper { margin-bottom: 30px; }
.fav-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; padding: 0 4px; }
.info-badges { display: flex; align-items: center; }
</style>