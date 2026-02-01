<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NLayout, NLayoutHeader, NLayoutContent, NLayoutSider, NEmpty, NSpin, 
  useMessage, NButton, NCheckbox, NSelect, NInput, NIcon, NPagination, NTag, NAlert, NTree
} from 'naive-ui'
import { 
  BookOutline, SearchOutline, TrashOutline, MenuOutline, FilterOutline 
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

// 🔥 错题目录树 (懒加载模式)
const mistakeTree = ref<any[]>([])
const loadingTree = ref(false) // 仅用于初始加载

const pagination = reactive({ page: 1, pageSize: 5, itemCount: 0 })
const filter = reactive({ 
  source: null as string | null, 
  keyword: '',
  category: '' 
})
const bankOptions = ref<any[]>([])

// 自动移除开关
const autoRemove = ref(localStorage.getItem('mistake_auto_remove') === 'true')
watch(autoRemove, (val) => {
  localStorage.setItem('mistake_auto_remove', String(val))
  if (val) message.info('已开启：答对后自动移出')
})

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
      label: item.label || item.name, // 错题树后端返回 label 带数量
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
        // 初始加载：获取一级目录 + 错题列表
        handleSourceChange()
      }
    }
  } catch (e) { console.error(e) }
}

// 🔥 1. 初始加载：只获取一级目录
const fetchRootTree = async () => {
  if (!filter.source) return
  loadingTree.value = true
  mistakeTree.value = [] // 清空旧树
  try {
    const res: any = await request.get('/mistake-tree', { 
      params: { source: filter.source, parent_key: '' }
    })
    // 根目录，父路径为空
    mistakeTree.value = adaptTreeData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

// 🔥 2. 懒加载：点击箭头时，加载子节点
const handleLoad = async (node: any) => {
  return new Promise<void>(async (resolve) => {
    try {
      const res: any = await request.get('/mistake-tree', { 
        params: { source: filter.source, parent_id: node.key } // 注意：后端改用 parent_id 接收 ID
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

// 点击节点筛选错题
const handleNodeClick = (keys: any, option: any) => {
  if (option && option.length > 0) {
    const node = option[0]
    // 🔥 使用拼接好的 full 发送请求
    filter.category = node.full || node.name 
    pagination.page = 1
    fetchData()
  } else {
    // 取消选中
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
    const res: any = await request.get('/mistakes', {
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
        // 注入单题历史
        if (!item.question.children || item.question.children.length === 0) {
           item.question.user_record = { choice: item.choice, is_correct: false }
        }
      }
      return item
    })
    pagination.itemCount = res.total || 0
  } catch (e) { message.error('加载失败') } finally { loading.value = false }
}

// 切换题库时：重置树 + 重置列表
const handleSourceChange = () => {
  filter.category = ''
  pagination.page = 1
  fetchRootTree() 
  fetchData()
}

// 仅搜索关键词时：只刷新列表 (保持树的状态，不让它缩回去)
const handleSearch = () => {
  pagination.page = 1
  fetchData()
}

const handlePageChange = (page: number) => { pagination.page = page; fetchData() }

// =========================
// 3. 交互操作
// =========================

const handleRemove = async (mistakeId: number, silent = false) => {
  try {
    await request.delete(`/mistakes/${mistakeId}`)
    list.value = list.value.filter(item => item.id !== mistakeId)
    pagination.itemCount-- 
    if (list.value.length === 0) { 
      // 列表空了，刷新列表
      fetchData()
      // 注意：懒加载模式下不建议强行刷新树，否则树会折叠，用户体验不好
      // 数字不准没关系，下次刷新页面就准了
    }
    if (!silent) message.success('已移出错题本')
  } catch (e) { if (!silent) message.error('移除失败') }
}

const onAnswerResult = (payload: { questionId: number, isCorrect: boolean }) => {
  const mistakeItem = list.value.find(item => {
    if (item.question?.id === payload.questionId) return true
    if (item.question?.children?.some((c:any) => c.id === payload.questionId)) return true
    return false
  })
  
  if (mistakeItem && payload.isCorrect && autoRemove.value) {
    setTimeout(() => {
      handleRemove(mistakeItem.id, true) 
      message.success('🎉 恭喜攻克！已自动移出')
    }, 800)
  }
}

const goBack = () => router.push('/')
const isValidQuestion = (q: any) => q && ( (q.options && Object.keys(q.options).length > 0) || (q.children && q.children.length > 0) )

onMounted(() => { fetchBanks() })
</script>

<template>
  <div class="mistakes-container">
    <div class="page-control-bar">
      <div class="left-controls">
        <h2 class="page-title">
          <n-icon color="#d03050" style="margin-right: 8px; vertical-align: bottom;"><BookOutline /></n-icon>
          我的错题本
        </h2>
        <div class="bank-selector">
          <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" size="small" />
        </div>
      </div>
      
      <div class="right-controls">
        <n-checkbox v-model:checked="autoRemove">
          <span style="font-size: 13px; font-weight: 500; color: #666;">答对自动移除</span>
        </n-checkbox>
        
        <div class="search-box">
          <n-input v-model:value="filter.keyword" placeholder="搜索关键词..." size="small" @keydown.enter="handleSearch" clearable>
            <template #prefix><n-icon><SearchOutline /></n-icon></template>
          </n-input>
        </div>
        <n-button type="primary" size="small" secondary @click="handleSearch">搜索</n-button>
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
          <n-icon color="#d03050"><FilterOutline /></n-icon> 错题分布 ({{ pagination.itemCount }})
        </div>
        
        <n-spin :show="loadingTree">
          <n-tree
            block-line
            expand-on-click
            :data="mistakeTree"
            key-field="key"
            label-field="label"
            selectable
            remote
            :on-load="handleLoad" 
            @update:selected-keys="handleNodeClick"
            style="font-size: 13px;"
          />
          <div v-if="mistakeTree.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 12px;">
            当前题库暂无错题记录
          </div>
        </n-spin>
      </n-layout-sider>

      <n-layout-content content-style="padding: 24px; max-width: 960px; margin: 0 auto;" :native-scrollbar="true">
        
        <div v-if="filter.category" style="margin-bottom: 16px;">
            <n-tag closable type="warning" @close="filter.category = ''; fetchData()">
              正在筛选: {{ filter.category }}
            </n-tag>
        </div>

        <n-spin :show="loading">
          <div v-if="list.length > 0">
            <div v-for="(item, index) in list" :key="item.id" class="mistake-item-wrapper">
              
              <div class="mistake-toolbar">
                <div class="info-badges">
                  <n-tag type="error" size="small" :bordered="false" style="margin-right: 8px;">
                    错题 #{{ (pagination.page - 1) * pagination.pageSize + index + 1 }}
                  </n-tag>
                  <span style="font-size: 12px; color: #999;">
                    收录于 {{ new Date(item.created_at).toLocaleDateString() }}
                  </span>
                </div>
                
                <n-button size="tiny" type="error" ghost @click="handleRemove(item.id)">
                    <template #icon><n-icon><TrashOutline /></n-icon></template>
                    直接移除
                </n-button>
              </div>

              <QuestionCard 
                v-if="isValidQuestion(item.question)"
                :question="item.question" 
                :serial-number="(pagination.page - 1) * pagination.pageSize + index + 1"
                @answer-result="onAnswerResult"
              />

              <n-alert v-else type="warning" title="数据验证异常" style="margin-top: 10px;">
                题目内容可能已被删除 (ID: {{ item.question?.id }})
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

          <n-empty v-else-if="!loading" description="太棒了！该分类下已没有错题！" style="margin-top: 100px">
            <template #extra>
              <n-button type="primary" @click="filter.category = ''; fetchData()">查看全部错题</n-button>
            </template>
          </n-empty>
        </n-spin>
      </n-layout-content>
    </n-layout>
  </div>
</template>

<style scoped>
.mistakes-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: transparent;
}

.page-control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  margin-bottom: 0;
  background-color: transparent;
  border-bottom: none;
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
.search-box { width: 200px; }
.main-layout-area { 
  flex: 1; 
  overflow: hidden; 
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background-color: #fff; /* Ensure white background for the content area */
}

.mistake-item-wrapper { margin-bottom: 30px; }
.mistake-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; padding: 0 4px; }
.info-badges { display: flex; align-items: center; }
</style>