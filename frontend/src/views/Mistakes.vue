<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { 
  NLayout, NLayoutContent, NLayoutSider, NEmpty, NSpin, 
  useMessage, NButton, NCheckbox, NSelect, NInput, NIcon, NTag, NTree, NTooltip,
  NDrawer, NDrawerContent
} from 'naive-ui'
import type { TreeOption } from 'naive-ui'
import { 
  BookOutline, SearchOutline, TrashOutline, FilterOutline,
  ChevronBackOutline, ChevronForwardOutline, PushOutline, Push, ListOutline, Flame,
  MenuOutline
} from '@vicons/ionicons5'
import QuestionCard from '../components/QuestionCard.vue'
import request from '../utils/request'

const message = useMessage()

interface CustomTreeOption extends TreeOption {
  name?: string
  full?: string
}

// =========================
// 1. 核心状态定义
// =========================
const mistakeTree = ref<CustomTreeOption[]>([])
const loadingTree = ref(false) 
const expandedKeys = ref<Array<string | number>>([])

const filter = ref({ source: null as string | null, keyword: '', category: '' })
const bankOptions = ref<{label: string, value: string}[]>([])

const autoRemove = ref(localStorage.getItem('mistake_auto_remove') === 'true')
watch(autoRemove, (val) => {
  localStorage.setItem('mistake_auto_remove', String(val))
  if (val) message.info('已开启：答对后自动移出')
})

const skeletonList = ref<any[]>([])
const currentIndex = ref(0)
const currentDetail = ref<any>(null)
const loadingSkeleton = ref(false)
const loadingDetail = ref(false)

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
const handleLeftLeave = () => { if (!leftPinned.value && !isDropdownOpen.value) leftCollapsed.value = true }
const toggleLeftPin = () => { leftPinned.value = !leftPinned.value; leftCollapsed.value = !leftPinned.value }
const handleRightEnter = () => { if (!rightPinned.value) rightCollapsed.value = false }
const handleRightLeave = () => { if (!rightPinned.value) rightCollapsed.value = true }
const toggleRightPin = () => { rightPinned.value = !rightPinned.value; rightCollapsed.value = !rightPinned.value }

// =========================
// 2. 目录树逻辑
// =========================
const adaptTreeData = (list: any[], parentPath = ''): CustomTreeOption[] => {
  return list.map(item => {
    let currentFull = item.full
    if (!currentFull) currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name
    return { key: String(item.id), label: item.label || item.name, name: item.name, full: currentFull, isLeaf: item.isLeaf }
  })
}

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    if (res.data) {
      bankOptions.value = res.data.map((b: string) => ({ label: b, value: b }))
      if (!filter.value.source && bankOptions.value.length > 0) { 
        filter.value.source = bankOptions.value[0]?.value || null; 
        handleSourceChange() 
      }
    }
  } catch (e) { console.error(e) }
}

const fetchRootTree = async () => {
  const source = filter.value.source
  if (!source) return
  loadingTree.value = true; mistakeTree.value = [] 
  try {
    const res: any = await request.get('/mistake-tree', { params: { source: source, parent_key: '' } })
    mistakeTree.value = adaptTreeData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

const handleLoad = async (node: TreeOption) => {
  const source = filter.value.source
  if (!source) return
  try {
    const res: any = await request.get('/mistake-tree', { params: { source: source, parent_id: node.key } })
    const n = node as CustomTreeOption
    node.children = adaptTreeData(res.data || [], n.full || n.name || '')
  } catch (e) { node.children = [] }
}

const handleNodeClick = (_keys: Array<string | number>, option: Array<TreeOption | null>) => {
  const node = option[0] as CustomTreeOption | null
  if (node) {
    filter.value.category = node.full || node.name || ''
    if (!isMobile.value) leftPinned.value = false 
    else mobileLeftOpen.value = false
    fetchSkeleton()
  } else { filter.value.category = ''; fetchSkeleton() }
}

// =========================
// 3. 核心辅助函数
// =========================
const getStandardTypeName = (rawType: string) => { 
    const t = (rawType || '').toUpperCase(); 
    if (t.includes('A1')) return 'A1型题'; if (t.includes('A2')) return 'A2型题'; 
    if (t.includes('A3') || t.includes('A4')) return 'A3/A4型题'; 
    if (t.includes('B1')) return 'B1型题'; if (t.includes('X')) return 'X型题'; 
    return rawType || '其他题型' 
}
const TypePriority: Record<string, number> = { 'A1型题': 1, 'A2型题': 2, 'A3/A4型题': 3, 'B1型题': 4, 'X型题': 5, '简答题': 6 }

const getMistakeLevelClass = (count: number) => {
    if (count >= 3) return 'mistake-lv3'
    if (count === 2) return 'mistake-lv2'
    return 'mistake-lv1'
}

// =========================
// 4. 数据加载逻辑
// =========================
const fetchSkeleton = async () => {
  loadingSkeleton.value = true
  skeletonList.value = []; currentIndex.value = 0; currentDetail.value = null
  try {
    const res: any = await request.get('/mistakes/skeleton', { params: filter.value })
    let rawData = res.data || []
    rawData.sort((a: any, b: any) => (TypePriority[getStandardTypeName(a.type)] || 999) - (TypePriority[getStandardTypeName(b.type)] || 999))
    skeletonList.value = rawData.map((q: any, idx: number) => ({ ...q, displayIndex: idx + 1 }))
    if (skeletonList.value.length > 0) loadQuestionDetail(0)
  } catch (e) { message.error('加载失败') } finally { loadingSkeleton.value = false }
}

const loadQuestionDetail = async (index: number) => {
    if (index < 0 || index >= skeletonList.value.length) return
    currentIndex.value = index; const targetItem = skeletonList.value[index]
    loadingDetail.value = true
    try {
        const res: any = await request.get(`/questions/${targetItem.id}`)
        if (res.data) { currentDetail.value = res.data; currentDetail.value.displayIndex = targetItem.displayIndex; currentDetail.value._wrongCount = targetItem.wrong_count }
        document.querySelector('#mistake-scroll-container')?.scrollTo(0, 0)
    } catch(e) { message.error('加载失败') } finally { loadingDetail.value = false }
}

const handleSourceChange = () => { filter.value.category = ''; fetchRootTree(); fetchSkeleton() }
const handleSearch = () => { fetchSkeleton() }
const clearSearch = () => { filter.value.keyword = ''; fetchSkeleton() }
const handleSheetJump = (targetIndex: number) => { if (isMobile.value) mobileRightOpen.value = false; loadQuestionDetail(targetIndex) }
const goPrev = () => { if (currentIndex.value > 0) loadQuestionDetail(currentIndex.value - 1) }
const goNext = () => { if (currentIndex.value < skeletonList.value.length - 1) loadQuestionDetail(currentIndex.value + 1) }

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
      groups[type]?.forEach(q => { items.push({ isHeader: false, globalIndex: q.displayIndex, skeletonIndex: q.skeletonIndex, wrongCount: q.wrong_count || 1, id: q.id, key: `sheet-${q.id}` }) }) 
  })
  return items
})

const handleRemove = async (silent = false) => {
  const targetId = currentDetail.value?.id; if (!targetId) return
  try {
    await request.delete(`/mistakes/${targetId}`)
    skeletonList.value = skeletonList.value.filter(item => item.id !== targetId)
    if (skeletonList.value.length === 0) { currentDetail.value = null } 
    else { if (currentIndex.value >= skeletonList.value.length) currentIndex.value = skeletonList.value.length - 1; loadQuestionDetail(currentIndex.value) }
    if (!silent) message.success('已移出错题本'); fetchRootTree() 
  } catch (e) { if (!silent) message.error('移除失败') }
}

const onAnswerResult = (payload: { id: number, isCorrect: boolean }) => {
  // 1. 如果答对了且开启了自动移出，则执行移出逻辑
  if (payload.isCorrect && autoRemove.value) { 
      setTimeout(() => { 
          handleRemove(true); 
          message.success('🎉 已自动移出') 
      }, 800) 
  } 
  
  // 2. 🔥 实时更新：如果答错了，重新拉取骨架状态以更新错误次数标签
  if (!payload.isCorrect) {
      refreshSingleSkeleton(payload.id)
  }
}

// 🔥 新增：答错后局部刷新骨架数据
const refreshSingleSkeleton = async (qId: number) => {
    try {
        // 重新请求错题骨架（后端会返回最新的 wrong_count）
        const res: any = await request.get('/mistakes/skeleton', { params: filter.value })
        const rawData = res.data || []
        
        // 在本地列表中找到这道题并更新其 wrong_count
        const latestInfo = rawData.find((q: any) => q.id === qId)
        const localItem = skeletonList.value.find(q => q.id === qId)
        
        if (latestInfo && localItem) {
            localItem.wrong_count = latestInfo.wrong_count
            // 同步更新当前正在显示的详情标签
            if (currentDetail.value && currentDetail.value.id === qId) {
                currentDetail.value._wrongCount = latestInfo.wrong_count
            }
        }
    } catch (e) {
        console.error('刷新错题统计失败', e)
    }
}

onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile); fetchBanks() })
</script>

<template>
  <div class="mistakes-container">
    <n-layout has-sider class="main-layout-area">
      <n-layout-sider 
        v-if="!isMobile" bordered collapse-mode="width" :collapsed-width="36" :width="280" :collapsed="leftCollapsed"
        @mouseenter="handleLeftEnter" @mouseleave="handleLeftLeave" content-style="padding: 0; display: flex; flex-direction: column;" class="category-sider auto-expand-sider"
      >
        <div class="collapsed-strip" v-show="leftCollapsed"><n-icon size="20" color="#d03050"><BookOutline /></n-icon></div>
        <div class="expanded-content" v-show="!leftCollapsed">
            <div class="sider-toolbar">
                <span class="toolbar-title"><n-icon color="#d03050" size="18" style="margin-right: 6px; transform: translateY(2px)"><BookOutline /></n-icon>错题本</span>
                <n-tooltip trigger="hover">
                    <template #trigger><n-button text size="small" @click="toggleLeftPin" :type="leftPinned ? 'error' : 'default'"><template #icon><n-icon size="18"><component :is="leftPinned ? Push : PushOutline" /></n-icon></template></n-button></template>
                    {{ leftPinned ? '取消固定' : '固定' }}
                </n-tooltip>
            </div>
            <div class="sider-controls">
                <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" size="small" />
                <n-checkbox v-model:checked="autoRemove" style="margin-top: 8px;"><span style="font-size: 13px; color: #64748b;">答对后自动移出</span></n-checkbox>
            </div>
            <div class="filter-header"><n-icon color="#d03050"><FilterOutline /></n-icon> 错题分布 ({{ skeletonList.length }})</div>
            <div class="sider-scroll-area">
                <n-spin :show="loadingTree">
                  <n-tree block-line v-model:expanded-keys="expandedKeys" :data="mistakeTree" key-field="key" label-field="label" selectable remote :on-load="handleLoad" @update:selected-keys="handleNodeClick" style="font-size: 13px;" />
                </n-spin>
                <div v-if="mistakeTree.length === 0 && !loadingTree" class="empty-hint">暂无记录</div>
            </div>
        </div>
      </n-layout-sider>

      <n-layout has-sider style="flex: 1;">
        <n-layout-content content-style="padding: 24px; display: flex; flexDirection: column; min-height: 100%" id="mistake-scroll-container">
          <n-page-header v-if="filter.category" style="margin-bottom: 20px;">
            <template #title><n-tag type="error" size="small" round style="margin-right: 8px">正在消灭</n-tag> <span>{{ filter.category }}</span></template>
            <template #extra><n-button size="small" @click="handleNodeClick([], [])">查看全部</n-button></template>
          </n-page-header>
          <n-empty v-if="skeletonList.length === 0 && !loadingSkeleton" description="太棒了！这里没有错题！" style="margin-top: 100px"><template #icon><n-icon size="40" color="#10b981"><Flame /></n-icon></template></n-empty>

          <div v-if="skeletonList.length > 0" class="single-question-view">
             <div v-if="loadingSkeleton" style="padding: 50px 0; text-align: center;"><n-spin size="large" /></div>
             <div v-else class="question-wrap" :class="{ 'loading-mask': loadingDetail }">
                <div v-if="currentDetail" class="mistake-top-bar"><n-tag :type="currentDetail._wrongCount >= 3 ? 'error' : 'warning'" round size="large" style="font-weight: 700;"><template #icon><n-icon><Flame/></n-icon></template>本题已错 {{ currentDetail._wrongCount || 1 }} 次</n-tag><n-button size="small" type="error" ghost @click="handleRemove(false)"><template #icon><n-icon><TrashOutline /></n-icon></template> 手动移出</n-button></div>
                <QuestionCard v-if="currentDetail" :question="currentDetail" :serial-number="currentDetail.displayIndex" @answer-result="onAnswerResult" />
             </div>
             <div class="action-bar" v-if="!loadingSkeleton">
                 <n-button size="large" secondary @click="goPrev" :disabled="currentIndex === 0 || loadingDetail"><template #icon><n-icon><ChevronBackOutline/></n-icon></template> 上一题</n-button>
                 <div class="progress-indicator"><strong>{{ currentIndex + 1 }}</strong> / {{ skeletonList.length }}</div>
                 <n-button size="large" type="error" @click="goNext" :disabled="currentIndex === skeletonList.length - 1 || loadingDetail">下一题 <template #icon><n-icon><ChevronForwardOutline/></n-icon></template></n-button>
             </div>
          </div>
        </n-layout-content>

        <n-layout-sider 
          v-if="!isMobile && (skeletonList.length > 0 || rightPinned)" 
          bordered collapse-mode="width" :collapsed-width="36" :width="260" :collapsed="rightCollapsed" placement="right"
          @mouseenter="handleRightEnter" @mouseleave="handleRightLeave" content-style="padding: 0; background-color: #fff; display: flex; flex-direction: column;" class="sheet-sider auto-expand-sider"
        >
             <div class="collapsed-strip" v-show="rightCollapsed"><n-icon size="20" color="#d03050"><ListOutline /></n-icon></div>
             <div class="expanded-content" v-show="!rightCollapsed">
                <div class="sheet-header"><div class="sheet-title">题目导航</div><n-tooltip trigger="hover"><template #trigger><n-button text size="small" @click="toggleRightPin" :type="rightPinned ? 'error' : 'default'"><template #icon><n-icon size="18"><component :is="rightPinned ? Push : PushOutline" /></n-icon></template></n-button></template>{{ rightPinned ? '取消' : '固定' }}</n-tooltip></div>
                <div class="sheet-search"><n-input v-model:value="filter.keyword" placeholder="搜索错题..." size="small" round @keydown.enter="handleSearch" @clear="clearSearch" clearable><template #prefix><n-icon :component="SearchOutline" /></template></n-input></div>
                <div class="sheet-content">
                    <div class="sheet-flow">
                    <template v-for="item in answerSheetItems" :key="item.key">
                        <div v-if="item.isHeader" class="type-header"><span class="type-dot"></span>{{ item.type }}</div>
                        <div v-else class="number-circle" :class="[getMistakeLevelClass(item.wrongCount), { 'active-q': item.skeletonIndex === currentIndex }]" @click="handleSheetJump(item.skeletonIndex)">{{ item.globalIndex }}</div>
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

    <n-drawer v-model:show="mobileLeftOpen" placement="left" width="100%">
       <n-drawer-content title="错题本目录" closable>
           <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" style="margin-bottom: 20px;"/>
           <n-tree block-line v-model:expanded-keys="expandedKeys" :data="mistakeTree" selectable remote :on-load="handleLoad" @update:selected-keys="handleNodeClick" />
       </n-drawer-content>
    </n-drawer>

    <n-drawer v-model:show="mobileRightOpen" placement="right" width="100%">
       <n-drawer-content :title="`题目导航 (${skeletonList.length})`" closable>
           <n-input v-model:value="filter.keyword" placeholder="搜索错题..." round @keydown.enter="handleSearch" clearable style="margin-bottom: 20px;"><template #prefix><n-icon :component="SearchOutline" /></template></n-input>
           <div class="sheet-flow mobile-sheet-flow">
               <template v-for="item in answerSheetItems" :key="item.key">
                   <div v-if="item.isHeader" class="type-header" style="margin-top: 12px;"><span class="type-dot"></span>{{ item.type }}</div>
                   <div v-else class="number-circle mobile-circle" :class="[getMistakeLevelClass(item.wrongCount), { 'active-q': item.skeletonIndex === currentIndex }]" @click="handleSheetJump(item.skeletonIndex)">{{ item.globalIndex }}</div>
               </template>
           </div>
       </n-drawer-content>
    </n-drawer>
  </div>
</template>

<style scoped>
.mistakes-container { height: 100%; display: flex; flex-direction: column; background-color: transparent; }
.main-layout-area { flex: 1; overflow: hidden; background-color: #fff; }
.auto-expand-sider { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); z-index: 50; }
.collapsed-strip { height: 100%; width: 100%; display: flex; justify-content: center; padding-top: 24px; cursor: pointer; }
.expanded-content { height: 100%; display: flex; flex-direction: column; background-color: #fff; }
.sider-toolbar { padding: 20px 20px 12px 20px; display: flex; justify-content: space-between; align-items: center; }
.toolbar-title { font-weight: 800; font-size: 16px; color: #1e293b; display: flex; align-items: center; }
.sider-controls { padding: 0 20px; }
.filter-header { padding: 16px 20px 0 20px; font-weight: 700; color: #333; font-size: 13px; display: flex; align-items: center; gap: 6px; border-top: 1px dashed #e2e8f0; margin-top: 16px;}
.sider-scroll-area { flex: 1; overflow-y: auto; padding: 16px; }

.single-question-view { display: flex; flex-direction: column; flex: 1; justify-content: space-between; min-height: calc(100vh - 120px); }
.question-wrap { flex: 1; transition: opacity 0.3s; }
.mistake-top-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; padding: 12px 20px; background: #fef2f2; border-radius: 16px; border: 1px dashed #fca5a5; }
.action-bar { margin-top: 24px; padding: 16px 24px; background: #fff; border-radius: 16px; box-shadow: 0 -4px 20px rgba(0,0,0,0.03); border: 1px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; }
.progress-indicator strong { font-size: 18px; color: #1e293b; }

.sheet-header { padding: 16px 20px; border-bottom: 1px solid #f1f5f9; display: flex; justify-content: space-between; align-items: center; }
.sheet-search { padding: 12px 16px; border-bottom: 1px solid #f0f0f0; }
.sheet-content { padding: 20px; flex: 1; overflow-y: auto; }
.sheet-flow { display: flex; flex-wrap: wrap; gap: 10px; }

/* 🔥 答题卡基础样式修复 */
.number-circle { 
    width: 36px; height: 36px; border-radius: 10px; font-size: 14px; font-weight: 700; 
    display: flex; align-items: center; justify-content: center; cursor: pointer; 
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); user-select: none;
    background-color: #fff; color: #64748b; border: 1px solid #f1f5f9;
    box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}
.number-circle:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.06); }

/* 🔥 错题程度状态色 */
.mistake-lv1 { background-color: #fef2f2 !important; color: #ef4444 !important; border-color: #fca5a5 !important; }
.mistake-lv2 { background-color: #ef4444 !important; color: #fff !important; border: none !important; box-shadow: 0 2px 8px rgba(239, 68, 68, 0.4); }
.mistake-lv3 { background-color: #7f1d1d !important; color: #fff !important; border: none !important; box-shadow: 0 2px 8px rgba(127, 29, 29, 0.5); }

/* 🔥 高亮状态修复：使用 box-shadow 确保不被状态色覆盖 */
.active-q { 
    border-color: #ef4444 !important; 
    color: #ef4444 !important; 
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.4) !important;
    z-index: 2;
}

/* 如果背景色是红色系，高亮文字反白 */
.mistake-lv2.active-q, .mistake-lv3.active-q { color: #fff !important; border-color: #fff !important; }

.type-header { width: 100%; font-size: 13px; font-weight: 700; color: #94a3b8; margin-top: 20px; margin-bottom: 12px; display: flex; align-items: center; }
.type-dot { width: 6px; height: 6px; background-color: #e2e8f0; border-radius: 50%; margin-right: 8px; }

/* 手机端适配 */
.mobile-fabs { position: fixed; bottom: 80px; left: 20px; right: 20px; height: 0; display: flex; justify-content: space-between; z-index: 1000; pointer-events: none; }
.fab-btn { width: 48px; height: 48px; border-radius: 50%; background: #ef4444; display: flex; align-items: center; justify-content: center; pointer-events: auto; box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3); transition: all 0.2s; }
.fab-btn:active { transform: scale(0.92); }
.mobile-sheet-flow { gap: 12px; }
.mobile-circle { width: 44px; height: 44px; font-size: 16px; }
.empty-hint { text-align: center; color: #ccc; margin-top: 40px; font-size: 12px; }
</style>