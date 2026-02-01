<script setup lang="ts">
import { ref, onMounted, computed, nextTick, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import { 
  NLayout, NLayoutSider, NLayoutContent, 
  NTree, NSpin, NEmpty, NButton, NPageHeader, NTag,
  NPopconfirm, NSpace, NIcon, useMessage, NBackTop, NInput, NSelect, NTooltip,
  NDrawer, NDrawerContent
} from 'naive-ui'
import { 
  SearchOutline, LibraryOutline, HomeOutline, PushOutline, Push, MenuOutline, ListOutline, RefreshOutline
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

const expandedKeys = ref<any[]>([]) // Shared expanded keys

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

// 🔥 Mobile States
const isMobile = ref(false)
const mobileLeftOpen = ref(false)
const mobileRightOpen = ref(false)

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 } // Simple breakpoint

// 🔥 Sidebar Control States
const leftCollapsed = ref(true)
const leftPinned = ref(false)
const rightCollapsed = ref(true)
const rightPinned = ref(false)

// Left Sidebar Logic
const handleLeftEnter = () => { if (!leftPinned.value) leftCollapsed.value = false }
const handleLeftLeave = () => { if (!leftPinned.value) leftCollapsed.value = true }
const toggleLeftPin = () => { 
    leftPinned.value = !leftPinned.value
    if (leftPinned.value) leftCollapsed.value = false // Ensure expanded when pinned
    else leftCollapsed.value = false // Keep expanded until mouse leave usually, or just let hover handle it
}

// Right Sidebar Logic
const handleRightEnter = () => { if (!rightPinned.value) rightCollapsed.value = false }
const handleRightLeave = () => { if (!rightPinned.value) rightCollapsed.value = true }
const toggleRightPin = () => { 
    rightPinned.value = !rightPinned.value
    if (rightPinned.value) rightCollapsed.value = false 
}


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
    
    // Auto-pin sidebar on bank change (no category selected)
    if (!isMobile.value) {
        leftPinned.value = true
        leftCollapsed.value = false
    }

    fetchTreeRoot() 
}

const handleNodeClick = (keys: any, option: any) => {
  if (!option || option.length === 0) return
  const node = option[0]
  resetState()
  currentCategory.value = node.full || node.label 
  
  // Auto-unpin and collapse when category selected
  if (!isMobile.value) {
      leftPinned.value = false
      // We don't force collapse here immediately as mouse is likely over it,
      // let handleLeftLeave take care of it when user moves mouse away.
  }

  fetchQuestions(false) 

  // Force expand on click
  if (!expandedKeys.value.includes(node.key)) {
      expandedKeys.value.push(node.key)
  }
}

const handleMobileNodeClick = (keys: any, option: any) => {
    handleNodeClick(keys, option)
    // Removed auto-close logic based on user feedback. 
    // User can manually close the drawer to view questions.
    /*
    if (option && option.length > 0) {
        const node = option[0]
        if (node.isLeaf) {
            mobileLeftOpen.value = false
        }
    }
    */
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

const setupIntersectionObserver = () => { 
    if (observer) observer.disconnect(); 
    observer = new IntersectionObserver((entries) => { if (entries[0].isIntersecting && hasMore.value && !loadingQuestions.value) { pagination.value.page++; fetchQuestions(true) } }, { root: null, threshold: 0.1, rootMargin: '200px' }); 
    if (loadTrigger.value) observer.observe(loadTrigger.value) 
}

onMounted(async () => {
    checkMobile()
    window.addEventListener('resize', checkMobile)
    await fetchBanks(); 
    if (currentBank.value) {
        fetchTreeRoot();
        // Auto-pin on initial load if no category
        if (!isMobile.value && !currentCategory.value) {
            leftPinned.value = true
            leftCollapsed.value = false
        }
    }
    nextTick(() => { setupIntersectionObserver() }) 
})

watch(() => visibleQuestions.value.length, () => { nextTick(() => { if (loadTrigger.value && observer) { observer.disconnect(); observer.observe(loadTrigger.value) } }) })
</script>

<template>
  <div class="quiz-container">
    <!-- Main Layout -->
    <n-layout has-sider class="main-layout-area">
      
      <!-- 🔥 LEFT SIDER: CHAPTERS (Desktop) 🔥 -->
      <n-layout-sider 
        v-if="!isMobile"
        bordered 
        collapse-mode="width" 
        :collapsed-width="36" 
        :width="280" 
        :collapsed="leftCollapsed"
        @mouseenter="handleLeftEnter"
        @mouseleave="handleLeftLeave"
        content-style="padding: 0; display: flex; flex-direction: column;" 
        :native-scrollbar="false"
        class="category-sider auto-expand-sider"
      >
        <!-- Collapsed Strip -->
        <div class="collapsed-strip" v-show="leftCollapsed">
            <n-icon size="20" color="#999"><MenuOutline /></n-icon>
        </div>

        <!-- Expanded Content -->
        <div class="expanded-content" v-show="!leftCollapsed">
            <div class="sider-toolbar">
                <span class="toolbar-title">章节目录</span>
                <n-tooltip trigger="hover">
                    <template #trigger>
                        <n-button text size="small" @click="toggleLeftPin" :type="leftPinned ? 'primary' : 'default'">
                            <template #icon><n-icon size="18"><component :is="leftPinned ? Push : PushOutline" /></n-icon></template>
                        </n-button>
                    </template>
                    {{ leftPinned ? '取消固定' : '固定侧边栏' }}
                </n-tooltip>
            </div>
            
            <div class="sider-bank-select" style="padding: 0 16px 12px 16px; background: #fafafa; border-bottom: 1px solid #eee; display: flex; align-items: center; gap: 8px;">
               <span style="font-size: 13px; color: #666; white-space: nowrap;">选择学科</span>
               <n-select v-model:value="currentBank" :options="bankOptions" placeholder="切换题库" @update:value="handleBankChange" size="small">
                 <template #prefix><n-icon><LibraryOutline /></n-icon></template>
               </n-select>
            </div>
            
            <div class="sider-scroll-area">
                <n-spin :show="loadingTree">
                <n-tree 
                    block-line 
                    v-model:expanded-keys="expandedKeys"
                    :cancelable="false"
                    :data="treeData" 
                    key-field="key" 
                    label-field="label" 
                    children-field="children" 
                    remote
                    :on-load="handleLoad"
                    @update:selected-keys="handleNodeClick" 
                />
                </n-spin>
                <div v-if="treeData.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 13px;">
                    请先选择题库
                </div>
            </div>
        </div>
      </n-layout-sider>

      <n-layout has-sider sider-placement="right" class="content-layout">
        <n-layout-content 
            :content-style="{ padding: isMobile ? '12px' : '24px', backgroundColor: '#f8fafc' }" 
            :native-scrollbar="true" 
            id="question-scroll-container"
        >
          
          <n-page-header v-if="isSearching" style="margin-bottom: 20px;">
            <template #title>🔍 搜索结果: "{{ searchKeyword }}"</template>
            <template #extra><n-button size="small" @click="clearSearch">清除</n-button></template>
          </n-page-header>
          
          <n-page-header v-else-if="currentCategory" :style="{ marginBottom: isMobile ? '12px' : '20px' }">
            <template #title>
              <div style="font-size: 15px; font-weight: 600; color: #334155; display: flex; align-items: center; flex-wrap: wrap;">
                  <span style="font-size: 13px; color: #94a3b8; margin-right: 6px;">{{ currentBank }} /</span> 
                  <span>{{ currentCategory }}</span>
              </div>
            </template>
            <template #extra>
              <div class="header-actions" style="display: flex; align-items: center; gap: 12px;">
                <n-tag :bordered="false" type="default" size="medium" style="background: transparent; font-weight: 500; color: #64748b;">
                   共 <span style="font-weight: 700; color: #0f172a; margin: 0 2px;">{{ pagination.itemCount }}</span> 道题目
                </n-tag>
                <div style="width: 1px; height: 16px; background: #e2e8f0;"></div>
                <n-popconfirm @positive-click="handleResetChapter">
                  <template #trigger>
                    <n-button size="small" type="error" dashed>
                        <template #icon><n-icon><RefreshOutline /></n-icon></template>
                        重做本章
                    </n-button>
                  </template>
                  确定要清空本章的所有答题记录吗？这将无法恢复。
                </n-popconfirm>
              </div>
            </template>
          </n-page-header>
          
          <n-empty v-else-if="!isSearching" description="请选择左侧章节开始刷题" style="margin-top: 100px">
              <template #icon><n-icon size="40" color="#ddd"><LibraryOutline /></n-icon></template>
          </n-empty>
          
          <div v-if="visibleQuestions.length > 0" class="question-list">
              <QuestionCard v-for="q in visibleQuestions" :key="q.id" :question="q" :serial-number="q.displayIndex" @answer-result="onAnswerResult" />
          </div>
          
          <n-empty v-else-if="!loadingQuestions && isSearching" description="无结果" style="margin-top: 50px"></n-empty>
          
          <div ref="loadTrigger" class="load-trigger" v-if="currentCategory || isSearching">
            <div v-if="loadingQuestions"><n-spin size="small" /> 加载中...</div>
            <div v-else-if="!hasMore && visibleQuestions.length > 0">🎉 到底啦</div>
          </div>
          
          <n-back-top :right="isMobile ? 20 : 40" :bottom="isMobile ? 140 : 40" />
        </n-layout-content>

        <!-- 🔥 RIGHT SIDER: ANSWER SHEET (Desktop) 🔥 -->
        <n-layout-sider 
          v-if="!isMobile && (globalSheetItems.length > 0 || rightPinned)" 
          bordered 
          collapse-mode="width" 
          :collapsed-width="36" 
          :width="260" 
          :collapsed="rightCollapsed"
           @mouseenter="handleRightEnter"
           @mouseleave="handleRightLeave"
          content-style="padding: 0; background-color: #fff; display: flex; flex-direction: column;"
          class="sheet-sider auto-expand-sider"
        >
             <!-- Collapsed Strip -->
            <div class="collapsed-strip" v-show="rightCollapsed">
               <n-icon size="20" color="#999"><ListOutline /></n-icon>
            </div>

             <!-- Expanded Content -->
             <div class="expanded-content" v-show="!rightCollapsed">
                <div class="sheet-header">
                    <div class="sheet-title">题目总览 ({{ globalSheetItems.length }})</div>
                    <n-tooltip trigger="hover">
                        <template #trigger>
                             <n-button text size="small" @click="toggleRightPin" :type="rightPinned ? 'primary' : 'default'">
                                <template #icon><n-icon size="18"><component :is="rightPinned ? Push : PushOutline" /></n-icon></template>
                            </n-button>
                        </template>
                         {{ rightPinned ? '取消固定' : '固定栏' }}
                    </n-tooltip>
                </div>
                
                <div class="sheet-search" style="padding: 12px 16px; border-bottom: 1px solid #f0f0f0;">
                    <n-input v-model:value="searchKeyword" placeholder="搜索题目..." size="small" round @keydown.enter="handleSearch" @clear="clearSearch" clearable>
                       <template #prefix><n-icon :component="SearchOutline" /></template>
                    </n-input>
                </div>

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
             </div>
        </n-layout-sider>
      </n-layout>
    </n-layout>

    <!-- 📱 Mobile Floating Buttons -->
    <div v-if="isMobile" class="mobile-fabs">
       <div class="fab-btn left-fab" @click="mobileLeftOpen = true">
          <n-icon size="24" color="#fff"><MenuOutline /></n-icon>
       </div>
       <div class="fab-btn right-fab" @click="mobileRightOpen = true" v-if="globalSheetItems.length > 0">
          <n-icon size="24" color="#fff"><ListOutline /></n-icon>
          <!-- Badge removed as requested -->
       </div>
    </div>

    <!-- 📱 Mobile Left Drawer (Chapters) -->
    <n-drawer v-model:show="mobileLeftOpen" placement="left" width="100%">
       <n-drawer-content title="章节目录" closable>
           <div class="sider-bank-select" style="padding: 0 0 16px 0; background: #fff; border-bottom: 1px dashed #eee; display: flex; align-items: center; gap: 8px;">
               <span style="font-size: 14px; color: #666; white-space: nowrap;">选择学科</span>
               <n-select v-model:value="currentBank" :options="bankOptions" placeholder="切换题库" @update:value="handleBankChange" >
                 <template #prefix><n-icon><LibraryOutline /></n-icon></template>
               </n-select>
            </div>
            
            <div class="sider-scroll-area">
                <n-spin :show="loadingTree">
                <n-tree 
                    block-line 
                    v-model:expanded-keys="expandedKeys"
                    :cancelable="false"
                    :data="treeData" 
                    key-field="key" 
                    label-field="label" 
                    children-field="children" 
                    remote
                    :on-load="handleLoad"
                    @update:selected-keys="handleMobileNodeClick" 
                />
                </n-spin>
                <div v-if="treeData.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 13px;">
                    请先选择题库
                </div>
            </div>
       </n-drawer-content>
    </n-drawer>

    <!-- 📱 Mobile Right Drawer (Sheet) -->
    <n-drawer v-model:show="mobileRightOpen" placement="right" width="100%">
       <n-drawer-content :title="`题目总览 (${globalSheetItems.length})`" closable>
           <div class="sheet-search" style="padding: 0 0 16px 0; border-bottom: 1px solid #f0f0f0;">
                <n-input v-model:value="searchKeyword" placeholder="搜索题目..." round @keydown.enter="handleSearch" @clear="clearSearch" clearable>
                   <template #prefix><n-icon :component="SearchOutline" /></template>
                </n-input>
            </div>

            <div class="sheet-content">
                <div class="sheet-flow">
                <template v-for="item in answerSheetItems" :key="item.key">
                    <div v-if="item.isHeader" class="type-header"><span class="type-dot"></span>{{ item.type }}</div>
                    
                    <div v-else 
                        class="number-circle" 
                        :class="{ 'sheet-correct': item.status === 'correct', 'sheet-wrong': item.status === 'wrong', 'sheet-partial': item.status === 'partially-correct' }" 
                        @click="() => { handleSheetJump(item.raw); mobileRightOpen = false; }">
                        {{ item.globalIndex }}
                    </div>
                </template>
                </div>
            </div>
       </n-drawer-content>
    </n-drawer>

  </div>
</template>

<style scoped>
/* Only showing new/modified styles */
.quiz-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: transparent;
  /* Removed card styles for full screen flat layout */
}

.page-control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0; /*  Less padding, no horizontal padding to align with content if container has padding, or keep it */
  margin-bottom: 0;
  background-color: transparent; /* Transparent to blend */
  border-bottom: none;
}

.left-controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-title {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
  display: flex;
  align-items: center;
  letter-spacing: -0.5px;
}

.bank-selector { width: 180px; }
.search-box { width: 300px; }

.main-layout-area {
  flex: 1;
  overflow: hidden;
  /* Removed border radius and border for full screen flat layout */
  background-color: #fff;
}

/* Auto Expand Sider Styles */
.auto-expand-sider {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 50;
    position: relative;
    border-right: 1px solid #f1f5f9 !important;
}

.collapsed-strip {
    height: 100%;
    width: 100%;
    display: flex;
    justify-content: center;
    padding-top: 24px;
    background: #fff;
    cursor: pointer;
    transition: background-color 0.2s;
}
.collapsed-strip:hover { background-color: #f8fafc; }

.expanded-content {
    height: 100%;
    display: flex;
    flex-direction: column;
    background-color: #fff;
}

/* 🌟 左侧章节栏美化 */
.sider-toolbar {
    padding: 20px 20px 12px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #fff;
}

.toolbar-title { 
    font-weight: 800; 
    font-size: 16px; 
    color: #1e293b; 
    letter-spacing: -0.02em;
}

.sider-bank-select {
    padding: 0 20px 16px 20px;
    background: #fff;
    border-bottom: 1px dashed #e2e8f0;
}

.sider-scroll-area {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
}
/* 树形控件美化 */
:deep(.n-tree-node) {
    padding: 6px 0;
    border-radius: 8px;
    transition: all 0.2s;
}
:deep(.n-tree-node:hover) {
    background-color: #f8fafc;
}
:deep(.n-tree-node--selected) {
    background-color: #eff6ff !important;
}
:deep(.n-tree-node-content__text) {
    font-size: 14px;
    color: #475569;
    font-weight: 500;
}
:deep(.n-tree-node--selected .n-tree-node-content__text) {
    color: #3b82f6;
    font-weight: 700;
}


/* Question List & Sheet Styles */
.question-list { display: flex; flex-direction: column; padding-bottom: 40px; }
.sheet-header { padding: 16px 20px; border-bottom: 1px solid #f1f5f9; background-color: #fff; display: flex; justify-content: space-between; align-items: center; }
.sheet-title { font-weight: 700; font-size: 15px; color: #1e293b; }
.sheet-content { padding: 20px; flex: 1; overflow-y: auto; }
.sheet-flow { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }

.type-header { 
    width: 100%; font-size: 13px; font-weight: 700; color: #94a3b8; 
    margin-top: 20px; margin-bottom: 12px; 
    display: flex; align-items: center; 
    letter-spacing: 0.5px;
}
.type-header:first-child { margin-top: 0; }
.type-dot { width: 6px; height: 6px; background-color: #e2e8f0; border-radius: 50%; margin-right: 8px; }

.number-circle { 
    width: 36px; 
    height: 36px; 
    border-radius: 12px; /* 方圆形 */
    background-color: #fff; 
    border: 1px solid #f1f5f9;
    color: #64748b; 
    font-size: 14px; 
    font-weight: 600; 
    display: flex; 
    align-items: center; 
    justify-content: center; 
    cursor: pointer; 
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); 
    user-select: none; 
    position: relative;
    box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}

.number-circle:hover { 
    border-color: #cbd5e1;
    color: #1e293b;
    transform: translateY(-2px); 
    box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}

/* Status Styles */
.sheet-correct { 
    background: linear-gradient(135deg, #10b981 0%, #059669 100%) !important; 
    border: none !important;
    color: #fff !important; 
    box-shadow: 0 4px 10px rgba(16, 185, 129, 0.3) !important;
}

.sheet-wrong { 
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%) !important; 
    border: none !important;
    color: #fff !important; 
    box-shadow: 0 4px 10px rgba(239, 68, 68, 0.3) !important;
}

.sheet-partial { 
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%) !important; 
    border: none !important;
    color: #fff !important; 
    box-shadow: 0 4px 10px rgba(245, 158, 11, 0.3) !important;
}

.number-circle.active-q {
    border-color: #3b82f6;
    color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.load-trigger { padding: 32px; text-align: center; color: #94a3b8; font-size: 13px; }

:deep(.highlight-flash) {
    animation: flash-bg 1.5s ease-out;
}

@keyframes flash-bg {
    0% { background-color: rgba(37, 99, 235, 0.1); }
    100% { background-color: transparent; }
}

/* Mobile Floating Action Buttons */
.mobile-fabs {
    position: fixed;
    bottom: 80px; 
    left: 20px;
    right: 20px;
    height: 0; /* Just a container position reference */
    display: flex;
    justify-content: space-between;
    pointer-events: none; /* Let clicks pass through empty space */
    z-index: 1000;
}

.fab-btn {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    pointer-events: auto; /* Enable clicks on buttons */
    transition: all 0.2s;
    position: relative;
    backdrop-filter: blur(4px);
}

.fab-btn:active {
    transform: scale(0.92);
}

.fab-badge {
    position: absolute;
    top: -4px;
    right: -4px;
    background-color: #ef4444;
    color: #fff;
    font-size: 11px;
    height: 18px;
    min-width: 18px;
    border-radius: 9px; /* Pill shape */
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 4px;
    border: 2px solid #fff;
    font-weight: 700;
}

.content-layout {
  transition: all 0.3s ease;
}
</style>
