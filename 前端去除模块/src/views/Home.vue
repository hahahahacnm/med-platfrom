<script setup lang="ts">
import { ref, onMounted, computed, nextTick, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import Dashboard from '../components/Dashboard.vue'
import { 
  NLayout, NLayoutSider, NLayoutHeader, NLayoutContent, 
  NTree, NSpin, NEmpty, NButton, NPageHeader, NTag,
  NPopconfirm, NSpace, NIcon, useMessage, NBackTop, NInput, NSelect, NTooltip,
  NAvatar, NDropdown 
} from 'naive-ui'
import { 
  SearchOutline, LibraryOutline, SettingsOutline, 
  BookOutline, StarOutline, JournalOutline, 
  PersonOutline, LogOutOutline 
} from '@vicons/ionicons5'
import QuestionCard from '../components/QuestionCard.vue'

const router = useRouter()
const userStore = useUserStore()
const message = useMessage()

// =======================
// 1. 状态定义
// =======================
const treeData = ref<any[]>([]) 
const visibleQuestions = ref<any[]>([]) 
const globalSheetItems = ref<any[]>([]) 

const loadingTree = ref(false)  
const loadingQuestions = ref(false) 
const currentCategory = ref('') 
const hasMore = ref(true) 
const searchKeyword = ref('')
const isSearching = ref(false)
const bankOptions = ref<any[]>([]) 
const currentBank = ref<string | null>(null)
const answerStatusMap = ref<Record<string, boolean>>({}) 
const loadTrigger = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null
const globalQuestionCounter = ref(0) 
const pagination = ref({ page: 1, pageSize: 200, itemCount: 0 })

// =======================
// 2. 核心逻辑：数据适配与加载
// =======================

const adaptData = (list: any[], parentPath = '') => {
  return list.map(item => {
    let currentFull = item.full
    if (!currentFull) {
        currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name
    }
    return {
        key: item.id,
        label: item.name,
        full: currentFull, 
        isLeaf: item.is_leaf,
        children: null 
    }
  })
}

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    const list = res.data || []
    bankOptions.value = list.map((item: string) => ({ label: item, value: item }))
    if (!currentBank.value && bankOptions.value.length > 0) { 
        currentBank.value = bankOptions.value[0].value 
    }
  } catch (e) { console.error(e) }
}

const fetchTreeRoot = async () => {
  if (!currentBank.value) return 
  loadingTree.value = true
  try {
    const res: any = await request.get('/category-tree', { params: { source: currentBank.value, parent_id: 0 } })
    treeData.value = adaptData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

const handleLoad = async (node: any) => {
  return new Promise<void>(async (resolve) => {
    try {
      const res: any = await request.get('/category-tree', { 
          params: { parent_id: node.key, source: currentBank.value }
      })
      const currentPath = node.full || node.label
      node.children = adaptData(res.data || [], currentPath)
      resolve()
    } catch (e) { 
        node.children = []
        resolve() 
    }
  })
}

const handleBankChange = (val: string) => { 
    currentBank.value = val
    resetState()
    treeData.value = []
    fetchTreeRoot() 
}

const handleNodeClick = (keys: any, option: any) => {
  if (!option || option.length === 0) return
  const node = option[0]
  resetState()
  currentCategory.value = node.full || node.label 
  fetchQuestions(false) 
}

const resetState = () => {
  searchKeyword.value = ''
  isSearching.value = false
  currentCategory.value = ''
  pagination.value.page = 1 
  hasMore.value = true
  answerStatusMap.value = {}
  visibleQuestions.value = []
  globalSheetItems.value = []
  globalQuestionCounter.value = 0 
}

const fetchQuestions = async (isLoadMore = false, specificPage?: number) => {
  if (!currentCategory.value && !searchKeyword.value) return
  if (loadingQuestions.value) return 
  if (isLoadMore && !hasMore.value) return 

  loadingQuestions.value = true
  const requestPage = specificPage || pagination.value.page

  try {
    const params: any = { page: requestPage, page_size: pagination.value.pageSize }
    if (currentBank.value) params.source = currentBank.value
    
    if (searchKeyword.value.trim()) { 
        params.q = searchKeyword.value.trim()
        isSearching.value = true 
    } else { 
        params.category = currentCategory.value 
        isSearching.value = false 
    }

    const res: any = await request.get('/questions', { params })
    const newRawList = res.data || []
    
    if (!isLoadMore && !specificPage) {
      visibleQuestions.value = []
      globalSheetItems.value = []
      globalQuestionCounter.value = 0
      pagination.value.itemCount = res.total || 0 
      document.querySelector('#question-scroll-container')?.scrollTo(0, 0)
    }

    if (newRawList.length < pagination.value.pageSize) hasMore.value = false; else hasMore.value = true
    
    syncInitialStatus(newRawList)
    const sortedBatch = sortBatch([...newRawList])
    
    const processedBatch = sortedBatch.map((q: any) => {
        const domId = `question-anchor-${q.id}`
        let displayIndex = 0
        if (q.children && q.children.length > 0) {
            q.children = q.children.map((child: any) => {
                globalQuestionCounter.value++ 
                return { ...child, displayIndex: globalQuestionCounter.value }
            })
            displayIndex = q.children[0].displayIndex 
        } else {
            globalQuestionCounter.value++
            displayIndex = globalQuestionCounter.value
        }
        return { ...q, displayIndex, domId }
    })

    if (!specificPage) {
        processedBatch.forEach(item => {
            if (item.children && item.children.length > 0) {
                item.children.forEach((child: any) => {
                    globalSheetItems.value.push({
                        id: child.id,
                        type: item.type,
                        displayIndex: child.displayIndex,
                        domId: `question-anchor-${child.id}`,
                        parentId: item.id
                    })
                })
            } else {
                globalSheetItems.value.push({ id: item.id, type: item.type, displayIndex: item.displayIndex, domId: item.domId })
            }
        })
    }

    if (specificPage) {
        visibleQuestions.value = processedBatch
        pagination.value.page = specificPage 
    } else {
        visibleQuestions.value.push(...processedBatch)
        const MAX_DOM_Nodes = 500
        if (visibleQuestions.value.length > MAX_DOM_Nodes) {
            visibleQuestions.value.splice(0, visibleQuestions.value.length - 400)
        }
    }
  } catch (e) { console.error(e) } finally { loadingQuestions.value = false }
}

const handleSheetJump = async (item: any) => {
    let el = document.getElementById(item.domId)
    if (!el && item.parentId) {
        const parentDomId = `question-anchor-${item.parentId}`
        el = document.getElementById(parentDomId)
    }
    if (el) { 
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
        el.classList.add('highlight-flash')
        setTimeout(() => el?.classList.remove('highlight-flash'), 1500)
        return 
    }
    message.info("题目可能在之前的页面，请尝试重新加载章节") 
}

const TypePriority: Record<string, number> = { 'A1型题': 1, 'A2型题': 2, 'A3/A4型题': 3, 'B1型题': 4, 'X型题': 5, '简答题': 6, '名词解释': 7, '问答题': 8, '论述题': 9, '案例分析题': 10 }
const getStandardTypeName = (rawType: string) => { const t = (rawType || '').toUpperCase(); if (t.includes('A1')) return 'A1型题'; if (t.includes('A2')) return 'A2型题'; if (t.includes('A3') || t.includes('A4')) return 'A3/A4型题'; if (t.includes('B1')) return 'B1型题'; if (t.includes('X')) return 'X型题'; return t || '其他题型' }
const sortBatch = (list: any[]) => { return list.sort((a: any, b: any) => { const nameA = getStandardTypeName(a.type); const nameB = getStandardTypeName(b.type); return (TypePriority[nameA] || 999) - (TypePriority[nameB] || 999) }) }

const onAnswerResult = (payload: { id: string, isCorrect: boolean }) => { answerStatusMap.value[payload.id] = payload.isCorrect }
const syncInitialStatus = (list: any[]) => { list.forEach(q => { if (q.user_record) { answerStatusMap.value[q.id] = q.user_record.is_correct } ; if (q.children) { q.children.forEach((child: any) => { if (child.user_record) { answerStatusMap.value[child.id] = child.user_record.is_correct } }) } }) }

const answerSheetItems = computed(() => {
  if (!globalSheetItems.value.length) return []
  const groups: Record<string, any[]> = {}
  globalSheetItems.value.forEach(item => { 
      const s = getStandardTypeName(item.type); 
      if (!groups[s]) groups[s] = []; 
      groups[s].push({ globalIndex: item.displayIndex, domId: item.domId, id: item.id, status: getSheetStatus(item.id), raw: item }) 
  })
  const sortedTypes = Object.keys(groups).sort((a, b) => (TypePriority[a] || 999) - (TypePriority[b] || 999))
  const items: any[] = []
  sortedTypes.forEach(type => { 
      items.push({ isHeader: true, type: type, key: `header-${type}` }); 
      groups[type].forEach(q => { items.push({ isHeader: false, globalIndex: q.globalIndex, domId: q.domId, key: `sheet-${q.id}`, status: q.status, raw: q }) }) 
  })
  return items
})

const getSheetStatus = (id: string) => { const s = answerStatusMap.value[id]; if (s === undefined) return 'none'; return s ? 'correct' : 'wrong' }

const handleSearch = () => { const val = searchKeyword.value.trim(); if (!val) { message.warning('请输入关键词'); return }; resetState(); searchKeyword.value = val; fetchQuestions(false) }
const clearSearch = () => { resetState(); fetchQuestions(false) }

const handleResetChapter = async () => { if (!currentCategory.value) return; try { await request.delete('/answers/reset-chapter', { params: { category: currentCategory.value } }); message.success('已清空'); resetState(); fetchQuestions(false) } catch (e) { console.error(e) } }

const handleLogout = () => { userStore.logout(); router.push('/login') }

// ===================================
// 🔥🔥🔥 核心修复：下拉菜单权限控制 🔥🔥🔥
// ===================================

const baseDropdownOptions = [
    { label: '个人中心', key: 'profile', icon: () => h(NIcon, null, { default: () => h(PersonOutline) }) },
    { label: '后台管理', key: 'admin', icon: () => h(NIcon, null, { default: () => h(SettingsOutline) }) },
    { label: '退出登录', key: 'logout', icon: () => h(NIcon, null, { default: () => h(LogOutOutline) }) }
]

const finalDropdownOptions = computed(() => {
    // 过滤逻辑：
    // 如果是 'admin' 菜单项，只有当 userStore.role 是 'admin' 时才保留
    return baseDropdownOptions.filter(item => {
        if (item.key === 'admin') {
            return userStore.role === 'admin'
        }
        return true
    })
})

const handleUserSelect = (key: string) => {
    if (key === 'profile') router.push('/profile')
    if (key === 'admin') router.push('/admin')
    if (key === 'logout') handleLogout()
}

// 获取用户头像
const userAvatar = computed(() => {
    if (userStore.avatar) return `http://localhost:8080${userStore.avatar}`
    return undefined 
})

const setupIntersectionObserver = () => { 
    if (observer) observer.disconnect(); 
    observer = new IntersectionObserver((entries) => { if (entries[0].isIntersecting && hasMore.value && !loadingQuestions.value) { pagination.value.page++; fetchQuestions(true) } }, { root: null, threshold: 0.1, rootMargin: '200px' }); 
    if (loadTrigger.value) observer.observe(loadTrigger.value) 
}

onMounted(async () => { 
    await fetchBanks(); 
    
    // 🔥 重要：获取最新的个人资料，并同步角色到 store
    // 这样刷新页面后，role 信息依然准确
    try { 
        const res: any = await request.get('/user/profile'); 
        if (res.data) {
            userStore.avatar = res.data.avatar
            userStore.username = res.data.nickname || res.data.username
            userStore.role = res.data.role // 👈 关键：同步角色
        }
    } catch {}

    if (currentBank.value) fetchTreeRoot(); 
    nextTick(() => { setupIntersectionObserver() }) 
})

watch(() => visibleQuestions.value.length, () => { nextTick(() => { if (loadTrigger.value && observer) { observer.disconnect(); observer.observe(loadTrigger.value) } }) })
</script>

<template>
  <div class="home-container">
    <n-layout style="height: 100vh">
      <n-layout-header bordered style="padding: 0 24px; display: flex; justify-content: space-between; align-items: center; height: 64px; background: #fff; box-shadow: 0 1px 4px rgba(0,21,41,0.08); z-index: 100;">
        <div style="display: flex; align-items: center; gap: 24px;">
          <div style="font-size: 20px; font-weight: 800; color: #18a058; cursor: pointer; display: flex; align-items: center; gap: 8px;" @click="resetState">
              <span style="font-size: 24px;">🏥</span> 医考刷题
          </div>
          <div style="width: 220px;">
              <n-select v-model:value="currentBank" :options="bankOptions" placeholder="切换题库" @update:value="handleBankChange" size="medium"><template #prefix><n-icon><LibraryOutline /></n-icon></template></n-select>
          </div>
          
          <div class="review-group">
              <n-tooltip trigger="hover">
                  <template #trigger>
                      <n-button text class="nav-btn" @click="$router.push('/mistakes')">
                        <template #icon><n-icon size="18"><BookOutline /></n-icon></template>
                        错题本
                      </n-button>
                  </template>
                  查看历史错题
              </n-tooltip>

              <n-tooltip trigger="hover">
                  <template #trigger>
                      <n-button text class="nav-btn" @click="$router.push('/favorites')">
                        <template #icon><n-icon size="18"><StarOutline /></n-icon></template>
                        收藏夹
                      </n-button>
                  </template>
                  查看收藏题目
              </n-tooltip>

              <n-tooltip trigger="hover">
                  <template #trigger>
                      <n-button text class="nav-btn" @click="$router.push('/my-notes')">
                        <template #icon><n-icon size="18"><JournalOutline /></n-icon></template>
                        笔记本
                      </n-button>
                  </template>
                  查看我的笔记
              </n-tooltip>
          </div>
        </div>
        
        <div style="flex: 1; max-width: 500px; margin: 0 40px;">
          <n-input v-model:value="searchKeyword" placeholder="输入关键词搜索题目..." round size="large" @keydown.enter="handleSearch" @clear="clearSearch" clearable>
            <template #prefix><n-icon :component="SearchOutline" /></template>
          </n-input>
        </div>
        
        <div>
          <n-dropdown trigger="hover" :options="finalDropdownOptions" @select="handleUserSelect">
            <div style="display: flex; align-items: center; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: background 0.3s;" class="user-trigger">
                <n-avatar round size="medium" :src="userAvatar" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" style="border: 1px solid #eee; margin-right: 8px;" />
                <span style="font-weight: 500; color: #333; font-size: 15px;">{{ userStore.username || '医考生' }}</span>
            </div>
          </n-dropdown>
        </div>
      </n-layout-header>

      <n-layout has-sider position="absolute" style="top: 64px; bottom: 0">
        <n-layout-sider bordered collapse-mode="width" :collapsed-width="0" :width="280" show-trigger="arrow-circle" content-style="padding: 12px;" :native-scrollbar="false">
          <n-spin :show="loadingTree">
            <n-tree 
              block-line 
              expand-on-click 
              :data="treeData" 
              key-field="key" 
              label-field="label" 
              children-field="children" 
              remote
              :on-load="handleLoad"
              @update:selected-keys="handleNodeClick" 
            />
          </n-spin>
        </n-layout-sider>

        <n-layout has-sider sider-placement="right">
          <n-layout-content content-style="padding: 24px; background-color: #f5f7fa;" :native-scrollbar="true" id="question-scroll-container">
            <n-page-header v-if="isSearching" style="margin-bottom: 20px;"><template #title>🔍 搜索结果: "{{ searchKeyword }}"</template><template #extra><n-button size="small" @click="clearSearch">清除</n-button></template></n-page-header>
            
            <n-page-header v-else-if="currentCategory" style="margin-bottom: 20px;">
              <template #title><span style="font-size: 14px; color: #666;">{{ currentBank }} / </span> {{ currentCategory }}</template>
              <template #extra><n-space><n-popconfirm @positive-click="handleResetChapter"><template #trigger><n-button size="small" type="warning" ghost>重做本章</n-button></template>确定清空记录吗？</n-popconfirm><n-tag type="primary" size="small" round>共 {{ pagination.itemCount }} 大题</n-tag></n-space></template>
            </n-page-header>
            
            <div v-else><Dashboard /></div>
            
            <div v-if="visibleQuestions.length > 0" class="question-list">
                <QuestionCard v-for="q in visibleQuestions" :key="q.id" :question="q" :serial-number="q.displayIndex" @answer-result="onAnswerResult" />
            </div>
            
            <n-empty v-else-if="!loadingQuestions && isSearching" description="无结果" style="margin-top: 50px"></n-empty>
            <div ref="loadTrigger" class="load-trigger" v-if="currentCategory || isSearching"><div v-if="loadingQuestions"><n-spin size="small" /> 加载中...</div><div v-else-if="!hasMore && visibleQuestions.length > 0">🎉 到底啦</div></div>
            <n-back-top :right="300" :bottom="50" />
          </n-layout-content>

          <n-layout-sider v-if="globalSheetItems.length > 0" bordered collapse-mode="width" :collapsed-width="0" :width="260" show-trigger="arrow-circle" content-style="padding: 0; background-color: #fff;">
            <div class="sheet-header"><div class="sheet-title">📝 答题卡 ({{ globalSheetItems.length }})</div></div>
            <div class="sheet-content">
              <div class="sheet-flow">
                <template v-for="item in answerSheetItems" :key="item.key">
                  <div v-if="item.isHeader" class="type-header"><span class="type-dot"></span>{{ item.type }}</div>
                  
                  <div v-else 
                       class="number-circle" 
                       :class="{ 'sheet-correct': item.status === 'correct', 'sheet-wrong': item.status === 'wrong', 'sheet-partial': item.status === 'partially-correct' }" 
                       @click="handleSheetJump(item.raw)">
                       {{ item.globalIndex }}
                  </div>
                </template>
              </div>
            </div>
          </n-layout-sider>
        </n-layout>
      </n-layout>
    </n-layout>
  </div>
</template>

<style scoped>
.review-group {
    display: flex;
    gap: 16px;
    padding-left: 20px;
    border-left: 1px solid #f0f0f0;
}

.nav-btn {
    font-size: 15px;
    color: #555;
    transition: all 0.2s;
}

.nav-btn:hover {
    color: #18a058;
}

.user-trigger:hover {
    background-color: #f5f5f5;
}

.question-list { display: flex; flex-direction: column; padding-bottom: 20px; }
.sheet-header { padding: 16px; border-bottom: 1px solid #f0f0f0; background-color: #fff; position: sticky; top: 0; z-index: 10; font-weight: bold; text-align: center; }
.sheet-content { padding: 16px; }
.sheet-flow { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
.type-header { width: 100%; font-size: 12px; font-weight: bold; color: #999; margin-top: 10px; margin-bottom: 4px; display: flex; align-items: center; }
.type-dot { width: 6px; height: 6px; background-color: #18a058; border-radius: 50%; margin-right: 6px; }
.number-circle { width: 34px; height: 34px; border-radius: 8px; background-color: #f5f7fa; color: #666; font-size: 13px; font-weight: 500; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: all 0.2s; user-select: none; }
.number-circle:hover { background-color: #e0e0e0; transform: translateY(-2px); }
.sheet-correct { background-color: #18a058 !important; color: #fff !important; }
.sheet-wrong { background-color: #d03050 !important; color: #fff !important; }
.sheet-partial { background-color: #f0a020 !important; color: #fff !important; }
.load-trigger { padding: 20px; text-align: center; color: #999; }

:deep(.highlight-flash) {
    animation: flash-bg 1.5s ease-out;
}

@keyframes flash-bg {
    0% { background-color: rgba(24, 160, 88, 0.2); }
    100% { background-color: transparent; }
}
</style>