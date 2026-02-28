<script setup lang="ts">
import { ref, onMounted, computed, h, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import request from '../utils/request'
import { 
  NLayout, NLayoutSider, NLayoutContent, 
  NTree, NSpin, NEmpty, NButton, NPageHeader, NTag,
  NPopconfirm, NIcon, useMessage, NInput, NSelect, NTooltip,
  NDrawer, NDrawerContent, NProgress, NText, NNumberAnimation // 🔥 增加数字动画增强体验
} from 'naive-ui'
import { 
  SearchOutline, LibraryOutline, PushOutline, Push, 
  MenuOutline, ListOutline, RefreshOutline,
  ChevronBackOutline, ChevronForwardOutline, AnalyticsOutline
} from '@vicons/ionicons5'
import QuestionCard from '../components/QuestionCard.vue'

const router = useRouter()
const userStore = useUserStore()
const message = useMessage()

// =======================
// 1. 核心状态定义
// =======================
const treeData = ref<any[]>([]) 
const expandedKeys = ref<any[]>([]) 

const loadingTree = ref(false)  
const loadingSkeleton = ref(false)
const loadingDetail = ref(false)

const currentCategory = ref('') 
const searchKeyword = ref('')
const isSearching = ref(false)

const bankOptions = ref<any[]>([]) 
const currentBank = ref<string | null>(null)

const skeletonList = ref<any[]>([]) 
const currentIndex = ref(0)         
const currentDetail = ref<any>(null)

// 🔥 新增：本章科学统计数据
const chapterSummary = ref<{
  correct_num: number;
  attempted_num: number;
  total_num: number;
  accuracy_rate: string;
  mastery_rate: string;
} | null>(null)

const isMobile = ref(false)
const mobileLeftOpen = ref(false)
const mobileRightOpen = ref(false)
const leftCollapsed = ref(false) 
const leftPinned = ref(true)
const rightCollapsed = ref(true)
const rightPinned = ref(false)
const isDropdownOpen = ref(false)

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }

const handleLeftEnter = () => { if (!leftPinned.value) leftCollapsed.value = false }
const handleLeftLeave = () => { if (!leftPinned.value && !isDropdownOpen.value) {leftCollapsed.value = true }}
const toggleLeftPin = () => { leftPinned.value = !leftPinned.value; leftCollapsed.value = !leftPinned.value }

const handleRightEnter = () => { if (!rightPinned.value) rightCollapsed.value = false }
const handleRightLeave = () => { if (!rightPinned.value) rightCollapsed.value = true }
const toggleRightPin = () => { rightPinned.value = !rightPinned.value; rightCollapsed.value = !rightPinned.value }

// =======================
// 2. 目录树渲染引擎
// =======================
const renderTreeLabel = ({ option }: { option: any }) => {
  const total = option.total_count || 0
  const done = option.done_count || 0
  const percentage = total > 0 ? Math.round((done / total) * 100) : 0
  const status = percentage >= 100 ? 'success' : 'default'
  const isLong = option.label.length > 8

  return h(
    'div',
    { style: 'display: flex; align-items: center; justify-content: space-between; width: 100%; padding: 4px 0; overflow: hidden;' },
    [
      h('div', { 
        style: {
          flex: 1, marginRight: '12px', fontSize: isLong ? '12px' : '13.5px',
          lineHeight: '1.25', color: '#334155', fontWeight: '500',
          display: '-webkit-box', '-webkit-line-clamp': '2', '-webkit-box-orient': 'vertical',
          overflow: 'hidden', wordBreak: 'break-all'
        }
      }, option.label),
      h('div', { style: 'display: flex; flex-direction: column; align-items: flex-end; gap: 2px; flex-shrink: 0; min-width: 48px;' }, [
        h(NText, { depth: 3, style: 'font-size: 10px; font-family: monospace; color: #94a3b8; transform: scale(0.9);' }, () => `${done}/${total}`),
        h('div', { style: 'width: 36px' }, [
          h(NProgress, { type: 'line', percentage: percentage, showIndicator: false, height: 3.5, borderRadius: 2, status: status, processing: percentage > 0 && percentage < 100 })
        ])
      ])
    ]
  )
}

const adaptData = (list: any[], parentPath = '') => {
  return list.map(item => {
    let currentFull = item.full
    if (!currentFull) { currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name }
    return { key: item.id, label: item.name, full: currentFull, level: item.level, isLeaf: item.is_leaf, total_count: item.total_count, done_count: item.done_count, children: null }
  })
}

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    bankOptions.value = (res.data || []).map((item: string) => ({ label: item, value: item }))
    if (!currentBank.value && bankOptions.value.length > 0) currentBank.value = bankOptions.value[0].value 
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
      const res: any = await request.get('/category-tree', { params: { parent_id: node.key, source: currentBank.value } })
      node.children = adaptData(res.data || [], node.full || node.label)
      resolve()
    } catch (e) { node.children = []; resolve() }
  })
}

const handleBankChange = (val: string) => { 
  currentBank.value = val; resetState(); treeData.value = []
  if (!isMobile.value) { leftPinned.value = true; leftCollapsed.value = false }
  fetchTreeRoot() 
}

// 1. 修改点击逻辑：不要在这里立即设置 currentCategory
const handleNodeClick = (keys: any[], option: any[]) => {
  if (!option || option.length === 0) return;
  const node = option[0];
  
  // 仅获取目标分类名，先不赋值给响应式变量 currentCategory，防止 UI 标题抢跑
  const targetCategory = node.full || node.label;
  
  if (!isMobile.value) leftPinned.value = false;
  else mobileLeftOpen.value = false;
  
  if (!expandedKeys.value.includes(node.key)) expandedKeys.value.push(node.key);
  
  // 将目标传递给 fetch 函数
  fetchQuestions(targetCategory); 
}

const resetState = () => {
  searchKeyword.value = ''; 
  isSearching.value = false; 
  currentCategory.value = ''; // 重置后标题会消失，显示 n-empty
  skeletonList.value = []; 
  currentIndex.value = 0; 
  currentDetail.value = null; 
  chapterSummary.value = null;
}

// 2. 修改获取逻辑：请求成功后再更新 UI 状态
const fetchQuestions = async (targetCat?: string) => {
  // 确定要请求的分类：如果是点击触发则用传入的 targetCat，否则用现有的
  const catName = targetCat || currentCategory.value;
  if (!catName && !searchKeyword.value) return;
  
  loadingSkeleton.value = true;
  try {
    if (isSearching.value && searchKeyword.value) {
      const res: any = await request.get('/questions', { params: { q: searchKeyword.value, source: currentBank.value, page: 1, page_size: 100 } });
      skeletonList.value = (res.data || []).map((q: any, idx: number) => ({ id: q.id, type: q.type, status: q.user_record ? (q.user_record.is_correct ? 'correct' : 'wrong') : 'unfilled', displayIndex: idx + 1, _fullData: q }));
      chapterSummary.value = null;
    } else {
      const res: any = await request.get('/questions/skeleton', { params: { category: catName, source: currentBank.value } });
      
      // 🔥 关键改动点：只有当请求成功（不报 403/500）到达这里时，才正式更新标题和数据
      currentCategory.value = catName; 
      skeletonList.value = (res.data || []).map((q: any, idx: number) => ({ ...q, displayIndex: idx + 1 }));
      chapterSummary.value = res.summary || null;
    }
    
    if (skeletonList.value.length > 0) await loadQuestionDetail(0);
  } catch (e) {
    // 🔥 关键改动点：如果请求失败（如 403 Forbidden），执行 resetState()
    // 这会清空 currentCategory，使 UI 自动退回到“请选择章节”的 <n-empty> 状态
    resetState();
    console.error('获取题目失败', e);
  } finally {
    loadingSkeleton.value = false;
  }
}

const loadQuestionDetail = async (index: number) => {
  if (index < 0 || index >= skeletonList.value.length) return
  currentIndex.value = index
  const targetItem = skeletonList.value[index]
  if (targetItem._fullData) {
    currentDetail.value = targetItem._fullData; currentDetail.value.displayIndex = targetItem.displayIndex
    document.querySelector('#question-scroll-container')?.scrollTo(0, 0); return
  }
  loadingDetail.value = true
  try {
    const res: any = await request.get(`/questions/${targetItem.id}`)
    currentDetail.value = res.data; currentDetail.value.displayIndex = targetItem.displayIndex
    document.querySelector('#question-scroll-container')?.scrollTo(0, 0)
  } catch(e) { message.error('加载单题失败') } finally { loadingDetail.value = false }
}

// 🔥 核心逻辑：本地实时重算正确率
const onAnswerResult = (payload: { id: number, isCorrect: boolean }) => {
  const item = skeletonList.value[currentIndex.value]
  if (!item || !chapterSummary.value) return
  
  const oldStatus = item.status
  
  if (oldStatus === 'unfilled') {
    item.status = payload.isCorrect ? 'correct' : 'wrong'
    chapterSummary.value.attempted_num++ // 第一次做，已做数+1
    if (payload.isCorrect) chapterSummary.value.correct_num++ // 且做对了，正确数+1
    updateTreeCount(treeData.value, currentCategory.value)
  } else if (oldStatus === 'wrong' && payload.isCorrect) {
    // 之前错了，现在重做对了
    item.status = 'correct'
    chapterSummary.value.correct_num++
  } else if (oldStatus === 'correct' && !payload.isCorrect) {
    // 之前对了，现在重做错了 (极少见但需兼容)
    item.status = 'wrong'
    chapterSummary.value.correct_num--
  }

  // 实时重新计算百分比
  const s = chapterSummary.value
  s.accuracy_rate = s.attempted_num > 0 ? ((s.correct_num / s.attempted_num) * 100).toFixed(1) : "0.0"
  s.mastery_rate = s.total_num > 0 ? ((s.correct_num / s.total_num) * 100).toFixed(1) : "0.0"
}

const updateTreeCount = (nodes: any[], targetFull: string) => {
  for (let node of nodes) {
    if (node.full === targetFull) { node.done_count = (node.done_count || 0) + 1; return true }
    if (node.children && updateTreeCount(node.children, targetFull)) return true
  }
  return false
}

const handlePageJump = (idx: number) => { if (isMobile.value) mobileRightOpen.value = false; loadQuestionDetail(idx) }
const goPrev = () => { if (currentIndex.value > 0) loadQuestionDetail(currentIndex.value - 1) }
const goNext = () => { if (currentIndex.value < skeletonList.value.length - 1) loadQuestionDetail(currentIndex.value + 1) }

const handleResetChapter = async () => {
  try {
    await request.post('/questions/reset-chapter', { category: currentCategory.value, source: currentBank.value })
    message.success('本章答题记录已清空'); fetchQuestions()
  } catch (e) { message.error('清空记录失败') }
}

const handleSearch = () => { if (searchKeyword.value) { isSearching.value = true; fetchQuestions() } }

const getStandardTypeName = (rawType: string) => { 
  const t = (rawType || '').toUpperCase(); 
  if (t.includes('A1')) return 'A1型题'; if (t.includes('A2')) return 'A2型题'; if (t.includes('A3') || t.includes('A4')) return 'A3/A4型题'; if (t.includes('B1')) return 'B1型题'; if (t.includes('X')) return 'X型题'; return rawType || '其他题型' 
}
const TypePriority: Record<string, number> = { 'A1型题': 1, 'A2型题': 2, 'A3/A4型题': 3, 'B1型题': 4, 'X型题': 5, '简答题': 6 }

const answerSheetItems = computed(() => {
  if (!skeletonList.value.length) return []
  const groups: Record<string, any[]> = {}
  skeletonList.value.forEach((item, realIndex) => { 
    const s = getStandardTypeName(item.type); if (!groups[s]) groups[s] = []; groups[s].push({ ...item, skeletonIndex: realIndex }) 
  })
  const sortedTypes = Object.keys(groups).sort((a, b) => (TypePriority[a] || 999) - (TypePriority[b] || 999))
  const items: any[] = []
  sortedTypes.forEach(type => { 
    items.push({ isHeader: true, type: type, key: `header-${type}` }); 
    groups[type]?.forEach(q => { items.push({ isHeader: false, globalIndex: q.displayIndex, skeletonIndex: q.skeletonIndex, status: q.status, id: q.id, key: `sheet-${q.id}` }) }) 
  })
  return items
})

onMounted(async () => {
  checkMobile(); window.addEventListener('resize', checkMobile); await fetchBanks(); 
  if (currentBank.value) { fetchTreeRoot(); if (!isMobile.value && !currentCategory.value) { leftPinned.value = true; leftCollapsed.value = false } }
})
</script>

<template>
  <div class="quiz-container">
    <n-layout has-sider class="main-layout-area">
      
      <n-layout-sider 
        v-if="!isMobile" bordered collapse-mode="width" :collapsed-width="36" :width="260" resizable :min-width="220" :max-width="450" :collapsed="leftCollapsed"
        @mouseenter="handleLeftEnter" @mouseleave="handleLeftLeave" content-style="padding: 0; display: flex; flex-direction: column;" class="category-sider auto-expand-sider"
      >
        <div class="collapsed-strip" v-show="leftCollapsed"><n-icon size="20" color="#999"><MenuOutline /></n-icon></div>
        <div class="expanded-content" v-show="!leftCollapsed">
          <div class="sider-toolbar">
            <span class="toolbar-title">章节目录</span>
            <n-button text size="small" @click="toggleLeftPin" :type="leftPinned ? 'primary' : 'default'"><template #icon><n-icon size="18"><component :is="leftPinned ? Push : PushOutline" /></n-icon></template></n-button>
          </div>
          <div class="sider-bank-select">
            <n-select v-model:value="currentBank" :options="bankOptions" placeholder="切换题库" @update:value="handleBankChange" size="small" @update:show="(show) => isDropdownOpen = show" />
          </div>
          <div class="sider-scroll-area">
            <n-spin :show="loadingTree">
              <n-tree block-line v-model:expanded-keys="expandedKeys" :data="treeData" remote :on-load="handleLoad" :render-label="renderTreeLabel" @update:selected-keys="handleNodeClick" />
            </n-spin>
            <div v-if="treeData.length === 0 && !loadingTree" class="empty-tree-hint">请先选择题库</div>
          </div>
        </div>
      </n-layout-sider>

      <n-layout has-sider sider-placement="right" class="content-layout">
        <n-layout-content :content-style="{ padding: isMobile ? '12px' : '24px', backgroundColor: '#f8fafc', display: 'flex', flexDirection: 'column' }" id="question-scroll-container">
          
          <n-page-header v-if="isSearching" style="margin-bottom: 20px;">
            <template #title>🔍 搜索: "{{ searchKeyword }}"</template>
            <template #extra><n-button size="small" @click="resetState(); fetchTreeRoot()">退出</n-button></template>
          </n-page-header>

          <n-page-header v-else-if="currentCategory" :style="{ marginBottom: isMobile ? '12px' : '20px' }">
            <template #title>
              <div class="category-header-title">
                <span class="bank-prefix">{{ currentBank }} /</span> <span>{{ currentCategory }}</span>
              </div>
            </template>
            
            <template #subtitle v-if="chapterSummary && !isMobile">
               <div class="scientific-stats">
                  <div class="stat-item">
                    <span class="label">掌握度</span>
                    <span class="value mastery">{{ chapterSummary.mastery_rate }}%</span>
                  </div>
                  <div class="stat-divider"></div>
                  <div class="stat-item">
                    <span class="label">正确率</span>
                    <span class="value" :class="Number(chapterSummary.accuracy_rate) < 60 ? 'accuracy-low' : 'accuracy-high'">
                       {{ chapterSummary.accuracy_rate }}%
                    </span>
                  </div>
               </div>
            </template>

            <template #extra>
              <div style="display: flex; gap: 8px; align-items: center;">
                <n-popconfirm @positive-click="handleResetChapter">
                  <template #trigger><n-button size="small" type="error" dashed><template #icon><n-icon><RefreshOutline /></n-icon></template>重做</n-button></template>
                  确定要清空本章答题记录吗？
                </n-popconfirm>
              </div>
            </template>
          </n-page-header>

          <n-empty v-if="!currentCategory && !isSearching" description="请选择章节开始刷题" style="margin-top: 100px"><template #icon><n-icon size="40" color="#ddd"><LibraryOutline /></n-icon></template></n-empty>

          <div v-if="skeletonList.length > 0" class="single-question-view">
            <div v-if="loadingSkeleton" style="padding: 50px 0; text-align: center;"><n-spin size="large" /></div>
            <div v-else class="question-wrap" :class="{ 'loading-mask': loadingDetail }">
              <div v-if="isMobile && chapterSummary" class="mobile-summary-bar">
                 <span>🎯 掌握: {{ chapterSummary.mastery_rate }}%</span>
                 <span>✅ 正确: {{ chapterSummary.accuracy_rate }}%</span>
              </div>
              <QuestionCard v-if="currentDetail" :question="currentDetail" :serial-number="currentDetail.displayIndex" @answer-result="onAnswerResult" />
            </div>
            <div class="action-bar" v-if="!loadingSkeleton">
              <n-button size="large" secondary @click="goPrev" :disabled="currentIndex === 0 || loadingDetail">
                <template #icon><n-icon><ChevronBackOutline/></n-icon></template>
                上一题
              </n-button>

              <div class="progress-indicator">
                <strong>{{ currentIndex + 1 }}</strong> 
                <span style="margin: 0 4px; opacity: 0.3;">/</span> 
                {{ skeletonList.length }}
              </div>

              <n-button 
                size="large" 
                type="primary" 
                @click="goNext" 
                icon-placement="right"
                :disabled="currentIndex === skeletonList.length - 1 || loadingDetail"
              >
                <template #icon><n-icon><ChevronForwardOutline/></n-icon></template>
                下一题
              </n-button>
            </div>
          </div>
        </n-layout-content>

        <n-layout-sider 
          v-if="!isMobile && skeletonList.length > 0" bordered collapse-mode="width" :collapsed-width="36" :width="260" :collapsed="rightCollapsed"
          @mouseenter="handleRightEnter" @mouseleave="handleRightLeave" content-style="padding: 0; background-color: #fff; display: flex; flex-direction: column;" class="sheet-sider auto-expand-sider"
        >
          <div class="collapsed-strip" v-show="rightCollapsed"><n-icon size="20" color="#999"><ListOutline /></n-icon></div>
          <div class="expanded-content" v-show="!rightCollapsed">
            <div class="sheet-header">
              <div class="sheet-title">答题卡 ({{ skeletonList.length }})</div>
              <n-button text size="small" @click="toggleRightPin" :type="rightPinned ? 'primary' : 'default'"><template #icon><n-icon size="18"><component :is="rightPinned ? Push : PushOutline" /></n-icon></template></n-button>
            </div>
            <div class="sheet-search"><n-input v-model:value="searchKeyword" placeholder="搜索题目..." size="small" round @keydown.enter="handleSearch" clearable><template #prefix><n-icon :component="SearchOutline" /></template></n-input></div>
            <div class="sheet-content">
              <div class="sheet-flow">
                <template v-for="item in answerSheetItems" :key="item.key">
                  <div v-if="item.isHeader" class="type-header"><span class="type-dot"></span>{{ item.type }}</div>
                  <div v-else class="number-circle" :class="{'sheet-correct': item.status==='correct', 'sheet-wrong': item.status==='wrong', 'active-q': item.skeletonIndex===currentIndex}" @click="handlePageJump(item.skeletonIndex)">{{ item.globalIndex }}</div>
                </template>
              </div>
            </div>
          </div>
        </n-layout-sider>
      </n-layout>
    </n-layout>

    <div v-if="isMobile" class="mobile-fabs">
      <div class="fab-btn left-fab" @click="mobileLeftOpen = true"><n-icon size="24" color="#fff"><MenuOutline /></n-icon></div>
      <div class="fab-btn right-fab" @click="mobileRightOpen = true" v-if="skeletonList.length > 0"><n-icon size="24" color="#fff"><ListOutline /></n-icon></div>
    </div>

    <n-drawer v-model:show="mobileLeftOpen" width="100%" placement="left">
      <n-drawer-content title="章节目录" closable>
        <div class="mobile-drawer-inner">
          <n-select v-model:value="currentBank" :options="bankOptions" @update:value="handleBankChange" style="margin-bottom: 20px;" placeholder="选择题库"/>
          <n-spin :show="loadingTree">
            <n-tree block-line :data="treeData" remote :on-load="handleLoad" :render-label="renderTreeLabel" @update:selected-keys="handleNodeClick" />
          </n-spin>
        </div>
      </n-drawer-content>
    </n-drawer>

    <n-drawer v-model:show="mobileRightOpen" width="100%" placement="right">
      <n-drawer-content :title="`答题卡 (${skeletonList.length})`" closable>
        <div class="mobile-drawer-inner">
          <n-input v-model:value="searchKeyword" placeholder="搜索本章题目..." round @keydown.enter="handleSearch" clearable style="margin-bottom: 24px;">
            <template #prefix><n-icon :component="SearchOutline"/></template>
          </n-input>
          <div class="sheet-flow mobile-sheet-flow">
            <template v-for="item in answerSheetItems" :key="item.key">
              <div v-if="item.isHeader" class="type-header"><span class="type-dot"></span>{{ item.type }}</div>
              <div v-else class="number-circle mobile-circle" :class="{'sheet-correct': item.status==='correct', 'sheet-wrong': item.status==='wrong', 'active-q': item.skeletonIndex===currentIndex}" @click="handlePageJump(item.skeletonIndex)">{{ item.globalIndex }}</div>
            </template>
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<style scoped>
/* 容器与基础布局 */
.quiz-container { height: 100%; display: flex; flex-direction: column; background-color: transparent; }
.main-layout-area { flex: 1; overflow: hidden; background-color: #fff; }
.auto-expand-sider { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); z-index: 50; }
.collapsed-strip { height: 100%; width: 100%; display: flex; justify-content: center; padding-top: 24px; cursor: pointer; }
.expanded-content { height: 100%; display: flex; flex-direction: column; background-color: #fff; }
.sider-toolbar { padding: 20px 20px 12px 20px; display: flex; justify-content: space-between; align-items: center; }
.toolbar-title { font-weight: 800; font-size: 16px; color: #1e293b; }
.sider-bank-select { padding: 0 16px 12px 16px; border-bottom: 1px solid #eee; }
.sider-scroll-area { flex: 1; overflow-y: auto; padding: 12px; }

/* 目录树优化 */
:deep(.n-tree-node) { padding: 4px 0; border-radius: 6px; margin-bottom: 2px; }
:deep(.n-tree-node-content) { align-items: center; }
:deep(.n-tree-node-indent) { width: 14px !important; }

/* 🔥 科学统计看板样式 */
.scientific-stats {
  display: flex;
  align-items: center;
  background: #f1f5f9;
  padding: 4px 16px;
  border-radius: 100px;
  margin-left: 16px;
  border: 1px solid #e2e8f0;
}
.stat-item { display: flex; align-items: center; gap: 8px; }
.stat-item .label { font-size: 12px; color: #64748b; font-weight: 500; }
.stat-item .value { font-size: 14px; font-weight: 800; font-family: 'JetBrains Mono', monospace; }
.stat-divider { width: 1px; height: 14px; background: #cbd5e1; margin: 0 12px; }
.mastery { color: #3b82f6; }
.accuracy-high { color: #10b981; }
.accuracy-low { color: #f59e0b; }

/* QuizBank.vue 的 style 部分 */
.action-bar {
  margin-top: 32px;
  padding: 12px 24px;
  background: #fff;
  border-radius: 18px;
  border: 1px solid #f1f5f9;
  
  /* 🔥 核心修复逻辑 */
  display: flex !important;
  flex-direction: row !important; /* 强制横向 */
  align-items: center !important;  /* 强制垂直居中 */
  justify-content: space-between;
  
  box-shadow: 0 10px 25px -5px rgba(0,0,0,0.04);
}

/* 确保中间的进度文字也是居中的 */
.progress-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #94a3b8;
  font-weight: 600;
  line-height: 1; /* 统一行高防止偏移 */
}

.progress-indicator strong {
  font-size: 20px;
  color: #0f172a;
  font-family: 'JetBrains Mono', monospace;
  line-height: 1;
}

/* 手机端统计栏 */
.mobile-summary-bar {
  display: flex;
  justify-content: space-around;
  background: #fff;
  padding: 10px;
  border-radius: 12px;
  margin-bottom: 16px;
  font-size: 13px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  color: #334155;
}

/* 答题卡样式 */
.sheet-header { padding: 16px 20px; border-bottom: 1px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; }
.sheet-content { padding: 20px; flex: 1; overflow-y: auto; }
.sheet-flow { display: flex; flex-wrap: wrap; gap: 10px; }

.number-circle { 
  width: 36px; height: 36px; border-radius: 10px; border: 1px solid #f1f5f9; 
  font-size: 14px; font-weight: 600; display: flex; align-items: center; justify-content: center; 
  cursor: pointer; transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); 
  background-color: #fff; color: #64748b; box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}
.number-circle:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.06); }

.active-q { 
  border-color: #3b82f6 !important; color: #3b82f6 !important; 
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.3) !important;
  z-index: 2;
}

.sheet-correct { background: linear-gradient(135deg, #10b981 0%, #059669 100%) !important; color: #fff !important; border: none !important; }
.sheet-wrong { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%) !important; color: #fff !important; border: none !important; }

/* 手机端专用 */
.mobile-fabs { position: fixed; bottom: 80px; left: 20px; right: 20px; height: 0; display: flex; justify-content: space-between; z-index: 1000; pointer-events: none; }
.fab-btn { width: 48px; height: 48px; border-radius: 50%; background: #3b82f6; display: flex; align-items: center; justify-content: center; pointer-events: auto; box-shadow: 0 4px 12px rgba(0,0,0,0.2); }
.mobile-drawer-inner { padding: 8px 4px 40px 4px; }
.mobile-sheet-flow { gap: 12px; }
.mobile-circle { width: 44px; height: 44px; font-size: 16px; }

.type-header { width: 100%; font-size: 13px; font-weight: 700; color: #94a3b8; margin-top: 20px; margin-bottom: 12px; display: flex; align-items: center; }
.type-dot { width: 6px; height: 6px; background-color: #e2e8f0; border-radius: 50%; margin-right: 8px; }
</style>