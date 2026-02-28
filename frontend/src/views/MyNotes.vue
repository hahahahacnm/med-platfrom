<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NLayout, NLayoutContent, NLayoutSider, NEmpty, NSpin, 
  useMessage, NButton, NSelect, NInput, NIcon, NPagination, NTag, NTree,
  NDrawer, NDrawerContent, NModal, NRadioGroup, NRadio, NSpace, NUpload, NTooltip
} from 'naive-ui'
import type { TreeOption } from 'naive-ui'
import { 
  JournalOutline, SearchOutline, FilterOutline, MenuOutline, 
  Push, PushOutline, ChatbubbleOutline, AddOutline
} from '@vicons/ionicons5'
import NoteFeedCard from '../components/NoteFeedCard.vue' 
import QuestionCard from '../components/QuestionCard.vue' // 🔥 引入 QuestionCard 用于抽屉内展示题目
import request from '../utils/request'
import { useUserStore } from '../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

interface CustomTreeOption extends TreeOption {
  name?: string
  full?: string
}

// =========================
// 1. 核心状态定义
// =========================
const activeTab = ref<'published' | 'collected'>('published') 
const loading = ref(false)
const list = ref<any[]>([]) 
const noteTree = ref<CustomTreeOption[]>([])
const loadingTree = ref(false) 
const expandedKeys = ref<Array<string | number>>([])

const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 }) 
const filter = reactive({ source: null as string | null, keyword: '', category: '' })
const bankOptions = ref<{label: string, value: string}[]>([])

// 响应式控制
const isMobile = ref(false)
const mobileLeftOpen = ref(false)
const leftCollapsed = ref(false)
const leftPinned = ref(true)
const isDropdownOpen = ref(false)

const checkMobile = () => { isMobile.value = window.innerWidth <= 768 }
const handleLeftEnter = () => { if (!leftPinned.value) leftCollapsed.value = false }
const handleLeftLeave = () => { if (!leftPinned.value && !isDropdownOpen.value) leftCollapsed.value = true }
const toggleLeftPin = () => { leftPinned.value = !leftPinned.value; leftCollapsed.value = !leftPinned.value }

// =========================
// 2. 目录树与选项过滤逻辑
// =========================
const adaptTreeData = (list: any[], parentPath = ''): CustomTreeOption[] => {
  return list.map(item => {
    let currentFull = item.full || item.full_path
    if (!currentFull) currentFull = parentPath ? `${parentPath} > ${item.name}` : item.name

    let isLeafNode = false
    if (item.isLeaf !== undefined) isLeafNode = item.isLeaf
    else if (item.IsLeaf !== undefined) isLeafNode = item.IsLeaf
    else if (item.is_leaf !== undefined) isLeafNode = item.is_leaf

    return {
      key: String(item.id),
      label: item.label || item.name, 
      name: item.name,
      full: currentFull, 
      isLeaf: isLeafNode
    }
  })
}

const findLabelInTree = (nodes: any[], targetKey: string): string => {
  for (const node of nodes) {
    if (String(node.key) === targetKey) return String(node.label)
    if (Array.isArray(node.children) && node.children.length > 0) {
      const found = findLabelInTree(node.children, targetKey)
      if (found) return found
    }
  }
  return ''
}

const fetchBanks = async () => {
  try {
    const res: any = await request.get('/banks')
    if (res.data) {
      bankOptions.value = res.data.map((b: string) => ({ label: b, value: b }))
      if (!filter.source && bankOptions.value.length > 0 && bankOptions.value[0]) {
        filter.source = bankOptions.value[0].value
        handleSourceChange()
      }
    }
  } catch (e) { console.error(e) }
}

const fetchRootTree = async () => {
  if (!filter.source) return
  loadingTree.value = true
  noteTree.value = [] 
  try {
    const res: any = await request.get('/notes/note-tree', { 
      params: { source: filter.source, parent_id: 0, tab: activeTab.value } 
    })
    noteTree.value = adaptTreeData(res.data || [], '')
  } catch (e) { console.error(e) } finally { loadingTree.value = false }
}

const handleLoad = async (node: TreeOption) => {
  return new Promise<void>(async (resolve) => {
    try {
      const res: any = await request.get('/notes/note-tree', { 
        params: { source: filter.source, parent_id: node.key, tab: activeTab.value } 
      })
      const n = node as CustomTreeOption
      const currentPath = n.full || n.name || ''
      node.children = adaptTreeData(res.data || [], currentPath)
      resolve()
    } catch (e) {
      node.children = []
      resolve()
    }
  })
}

const handleNodeClick = (_keys: Array<string | number>, option: Array<TreeOption | null>) => {
  if (option && option.length > 0 && option[0]) {
    const node = option[0] as CustomTreeOption
    filter.category = String(node.key)
    pagination.page = 1
    if (isMobile.value) mobileLeftOpen.value = false
    fetchData()
  } else {
    filter.category = ''
    fetchData()
  }
}

// =========================
// 3. 获取信息流数据
// =========================
const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/notes/my', {
      params: {
        page: pagination.page, 
        page_size: pagination.pageSize,
        category_id: filter.category,
        source: filter.source,
        tab: activeTab.value
      }
    })
    
    list.value = res.data || []
    pagination.itemCount = res.total || 0
  } catch (e) { message.error('加载失败') } finally { loading.value = false }
}

const handleTabSwitch = (tab: 'published' | 'collected') => {
    if (activeTab.value === tab) return
    activeTab.value = tab
    filter.category = ''
    pagination.page = 1
    fetchRootTree()
    fetchData()
}

const handleSourceChange = () => { filter.category = ''; pagination.page = 1; fetchRootTree(); fetchData() }
const handleSearch = () => { pagination.page = 1; fetchData() }
const handlePageChange = (page: number) => { 
    pagination.page = page; fetchData() 
    document.querySelector('#feed-scroll-container')?.scrollTo(0, 0)
}

// =========================
// 4. 🔥 卡片交互与题目回溯抽屉 🔥
// =========================
const handleLike = async (n: any) => {
  const old = n.is_liked; n.is_liked = !n.is_liked; n.like_count += n.is_liked ? 1 : -1
  try { await request.post(`/notes/${n.id}/like`) } catch { n.is_liked = old; n.like_count -= n.is_liked ? 1 : -1; message.error('操作失败') }
}

const handleCollect = async (n: any) => {
  const old = n.is_collected; n.is_collected = !n.is_collected
  try { 
      await request.post(`/notes/${n.id}/collect`); 
      message.success(n.is_collected ? '已收藏' : '已取消') 
      if (activeTab.value === 'collected' && !n.is_collected) {
          list.value = list.value.filter(item => item.id !== n.id)
          pagination.itemCount--
      }
  } catch { n.is_collected = old; message.error('操作失败') }
}

const handleDelete = async (noteId: number) => {
    try { 
        await request.delete(`/notes/${noteId}`); 
        message.success('已删除'); 
        list.value = list.value.filter(item => item.id !== noteId)
        pagination.itemCount--
        fetchRootTree()
    } catch { message.error('删除失败') }
}

// 🌟 题目回溯抽屉状态
const showQuestionDrawer = ref(false)
const jumpQuestionDetail = ref<any>(null)
const loadingJumpQuestion = ref(false)

// 🔥 核心重构：不再进行页面跳转，而是原地呼出单题抽屉
const handleJumpQuestion = async (qId: number) => {
    showQuestionDrawer.value = true // 呼出抽屉
    loadingJumpQuestion.value = true
    jumpQuestionDetail.value = null // 清空旧数据
    try {
        // 调用单题查询接口，获取完整题目（含共用题干/子题等）
        const res: any = await request.get(`/questions/${qId}`)
        if (res.data) {
            jumpQuestionDetail.value = res.data
            jumpQuestionDetail.value.displayIndex = 1 // 抽屉内默认显示序号1
        }
    } catch(e) {
        message.error('获取题目详情失败')
        showQuestionDrawer.value = false
    } finally {
        loadingJumpQuestion.value = false
    }
}

// =========================
// 5. 回复与举报模态框
// =========================
const rpt = reactive({ show: false, noteId: 0, type: '营销广告', desc: '', loading: false })
const rptTypes = ['言语辱骂', '虚假消息', '营销广告', '涉政有害', '违法违规', '垃圾消息','其他']
const openReport = (noteId: number) => { rpt.noteId = noteId; rpt.type = '营销广告'; rpt.desc = ''; rpt.show = true }
const submitReport = async () => {
    rpt.loading = true
    try { await request.post(`/notes/${rpt.noteId}/report`, { reason: `${rpt.type}${rpt.desc ? '：'+rpt.desc : ''}` }); message.success('举报已提交'); rpt.show = false } catch (e: any) { message.error(e.response?.data?.error || '提交失败') } finally { rpt.loading = false }
}

const replyState = reactive({ show: false, parentId: 0, questionId: 0, content: '', replyToUser: '', images: [] as string[], loading: false })
const openReply = (n: any) => {
    Object.assign(replyState, { show: true, parentId: n.id, questionId: n.question_id, content: '', replyToUser: n.user?.nickname || n.user?.username, images: [], loading: false })
}
const handleUpload = async ({ file }: { file: any }) => {
  if (replyState.images.length >= 5) return message.warning('最多上传5张')
  const form = new FormData(); form.append('file', file.file)
  try { const res: any = await request.post('/notes/upload', form, { headers: { 'Content-Type': 'multipart/form-data' } }); replyState.images.push(res.url) } catch (e: any) { message.error('上传失败') }
}
const submitReply = async () => {
    if (!replyState.content.trim() && !replyState.images.length) return message.warning('内容不能为空')
    replyState.loading = true
    try {
        await request.post('/notes', { id: 0, question_id: replyState.questionId, content: replyState.content, is_public: true, parent_id: replyState.parentId, images: replyState.images })
        message.success('回复成功！'); replyState.show = false;
        if(activeTab.value === 'published') fetchData()
    } catch(e:any) { message.error(e.response?.data?.error || '操作失败') } finally { replyState.loading = false }
}

onMounted(() => { checkMobile(); window.addEventListener('resize', checkMobile); fetchBanks() })
</script>

<template>
  <div class="notes-container">
    <n-layout has-sider class="main-layout-area">
      
      <n-layout-sider 
        v-if="!isMobile" bordered collapse-mode="width" :collapsed-width="36" :width="280" :collapsed="leftCollapsed"
        @mouseenter="handleLeftEnter" @mouseleave="handleLeftLeave" content-style="padding: 0; display: flex; flex-direction: column;" 
        class="category-sider auto-expand-sider"
      >
        <div class="collapsed-strip" v-show="leftCollapsed"><n-icon size="20" color="#18a058"><JournalOutline /></n-icon></div>

        <div class="expanded-content" v-show="!leftCollapsed">
            <div class="sider-toolbar">
                <span class="toolbar-title"><n-icon color="#18a058" size="18" style="margin-right: 6px; transform: translateY(2px)"><JournalOutline /></n-icon>我的笔记本</span>
                <n-tooltip trigger="hover">
                    <template #trigger><n-button text size="small" @click="toggleLeftPin" :type="leftPinned ? 'success' : 'default'"><template #icon><n-icon size="18"><component :is="leftPinned ? Push : PushOutline" /></n-icon></template></n-button></template>
                    {{ leftPinned ? '取消固定' : '固定' }}
                </n-tooltip>
            </div>
            
            <div class="sider-controls">
                <div class="control-item">
                    <n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" size="small" />
                </div>
                <div class="control-item">
                    <n-input v-model:value="filter.keyword" placeholder="搜索笔记内容..." size="small" round @keydown.enter="handleSearch" clearable>
                        <template #prefix><n-icon><SearchOutline /></n-icon></template>
                    </n-input>
                </div>
            </div>

            <div class="filter-header"><n-icon color="#18a058"><FilterOutline /></n-icon> 笔记分布 ({{ pagination.itemCount }})</div>
            
            <div class="sider-scroll-area">
                <n-spin :show="loadingTree">
                <n-tree block-line v-model:expanded-keys="expandedKeys" :data="noteTree" key-field="key" label-field="label" selectable remote :on-load="handleLoad" @update:selected-keys="handleNodeClick" style="font-size: 13px;" />
                </n-spin>
                <div v-if="noteTree.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 12px;">
                    暂无记录
                </div>
            </div>
        </div>
      </n-layout-sider>

      <n-layout-content content-style="padding: 0; background-color: #f8fafc; display: flex; flex-direction: column;" id="feed-scroll-container" :native-scrollbar="true">
        
        <div class="feed-header">
            <div class="feed-tabs">
                <div class="feed-tab" :class="{ 'active-tab': activeTab === 'published' }" @click="handleTabSwitch('published')">
                    我发布的
                    <div v-if="activeTab === 'published'" class="tab-indicator"></div>
                </div>
                <div class="feed-tab" :class="{ 'active-tab': activeTab === 'collected' }" @click="handleTabSwitch('collected')">
                    我收藏的
                    <div v-if="activeTab === 'collected'" class="tab-indicator"></div>
                </div>
            </div>
        </div>

        <div class="feed-main-scroll">
            <div class="feed-content-wrapper">
                <div v-if="filter.category" class="filter-tag-row">
                    <n-tag closable type="success" @close="filter.category = ''; fetchData()" round size="large" style="font-weight: 600;">
                        📍 {{ findLabelInTree(noteTree, filter.category) || '当前筛选节点' }}
                    </n-tag>
                </div>

                <n-spin :show="loading">
                    <div v-if="list.length > 0">
                        <NoteFeedCard
                            v-for="note in list" 
                            :key="note.id" 
                            :note="note"
                            :is-owner="note.user_id === userStore.id"
                            :is-admin="userStore.role === 'admin'"
                            @like="handleLike"
                            @collect="handleCollect"
                            @delete="handleDelete"
                            @report="openReport"
                            @reply="openReply"
                            @jump-question="handleJumpQuestion"
                        />
                        
                        <div class="pagination-wrapper">
                            <n-pagination
                                v-model:page="pagination.page"
                                :item-count="pagination.itemCount"
                                :page-size="pagination.pageSize"
                                @update:page="handlePageChange"
                            />
                        </div>
                    </div>

                    <n-empty v-else-if="!loading" :description="activeTab === 'published' ? '你还没有发表过笔记，去题库留下你的脚印吧！' : '你还没有收藏过宝藏笔记，快去探索吧！'" style="margin-top: 120px">
                        <template #icon>
                            <n-icon size="48" :color="activeTab === 'published' ? '#10b981' : '#f59e0b'"><ChatbubbleOutline /></n-icon>
                        </template>
                    </n-empty>
                </n-spin>
            </div>
        </div>
      </n-layout-content>
    </n-layout>

    <div v-if="isMobile" class="mobile-fabs">
       <div class="fab-btn left-fab" @click="mobileLeftOpen = true">
          <n-icon size="24" color="#fff"><MenuOutline /></n-icon>
       </div>
    </div>

    <n-drawer v-model:show="mobileLeftOpen" placement="left" width="100%">
       <n-drawer-content title="笔记本目录" closable>
           <div class="sider-controls" style="padding-bottom: 16px; border-bottom: 1px dashed #eee;">
               <div class="control-item"><n-select v-model:value="filter.source" :options="bankOptions" placeholder="选择题库" @update:value="handleSourceChange" size="small" /></div>
           </div>
           <div class="filter-header" style="padding-top: 16px; margin-top: 0; border-top: none;"><n-icon color="#18a058"><FilterOutline /></n-icon> 目录分布 ({{ pagination.itemCount }})</div>
           <div class="sider-scroll-area" style="padding: 0; margin-top: 16px;">
               <n-spin :show="loadingTree">
               <n-tree block-line v-model:expanded-keys="expandedKeys" :data="noteTree" key-field="key" label-field="label" selectable remote :on-load="handleLoad" @update:selected-keys="handleNodeClick" style="font-size: 13px;" />
               </n-spin>
               <div v-if="noteTree.length === 0 && !loadingTree" style="text-align: center; color: #ccc; margin-top: 40px; font-size: 12px;">暂无记录</div>
           </div>
       </n-drawer-content>
    </n-drawer>

    <n-modal v-model:show="replyState.show" preset="card" style="width: 500px; border-radius: 16px;" title="💬 回复探讨">
        <div style="margin-bottom: 12px; color: #18a058; font-weight: 600;">回复 @{{ replyState.replyToUser }}：</div>
        <n-input v-model:value="replyState.content" type="textarea" placeholder="输入友善的回复..." :autosize="{minRows:3, maxRows:6}" maxlength="200" show-count />
        <div style="margin-top: 16px;">
            <n-upload :show-file-list="false" @change="handleUpload" accept="image/*" :disabled="replyState.images.length >= 5">
                <n-button size="small" dashed type="success"><template #icon><n-icon><AddOutline/></n-icon></template>附加图片 ({{ replyState.images.length }}/5)</n-button>
            </n-upload>
        </div>
        <div v-if="replyState.images.length > 0" class="mini-img-preview">
            <div v-for="(url, i) in replyState.images" :key="i" class="mini-img-item" @click="replyState.images.splice(i,1)">
                <img :src="`http://localhost:8080${url}`" />
                <div class="overlay">删除</div>
            </div>
        </div>
        <template #footer>
            <div style="display: flex; justify-content: flex-end; gap: 12px;">
                <n-button round @click="replyState.show = false">取消</n-button>
                <n-button round type="primary" :loading="replyState.loading" @click="submitReply">发送</n-button>
            </div>
        </template>
    </n-modal>

    <n-modal v-model:show="rpt.show" preset="dialog" title="🚨 举报违规">
        <div style="padding:10px 0">
            <div style="margin-bottom:8px;font-weight:bold;color:#666">理由：</div>
            <n-radio-group v-model:value="rpt.type"><n-space vertical><n-radio v-for="r in rptTypes" :key="r" :value="r">{{r}}</n-radio></n-space></n-radio-group>
            <div style="margin-top:16px;font-weight:bold;color:#666">说明：</div>
            <n-input v-model:value="rpt.desc" type="textarea" placeholder="细节描述..." :autosize="{minRows:2,maxRows:4}"/>
        </div>
        <template #action><n-button @click="rpt.show=false">取消</n-button><n-button type="error" :loading="rpt.loading" @click="submitReport">提交</n-button></template>
    </n-modal>

    <n-drawer 
        v-model:show="showQuestionDrawer" 
        :width="isMobile ? '100%' : 600" 
        placement="right"
    >
       <n-drawer-content title="📝 原题重温" closable style="background-color: #f8fafc;">
           <n-spin :show="loadingJumpQuestion" style="min-height: 200px;">
               <div v-if="jumpQuestionDetail" style="padding: 16px 0;">
                   <QuestionCard 
                       :question="jumpQuestionDetail" 
                       :serial-number="1" 
                       :init-show-notes="false" 
                   />
               </div>
               <n-empty v-else-if="!loadingJumpQuestion" description="题目数据走丢了..." style="margin-top: 50px;" />
           </n-spin>
       </n-drawer-content>
    </n-drawer>

  </div>
</template>

<style scoped>
.notes-container { height: 100%; display: flex; flex-direction: column; background-color: transparent; }
.main-layout-area { flex: 1; overflow: hidden; background-color: #fff; }

/* 侧边栏样式 */
.auto-expand-sider { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); z-index: 50; position: relative; border-right: 1px solid #f1f5f9 !important; background-color: #fff;}
.collapsed-strip { height: 100%; width: 100%; display: flex; justify-content: center; padding-top: 24px; background: #fff; cursor: pointer; transition: background-color 0.2s; }
.collapsed-strip:hover { background-color: #f0fdf4; }
.expanded-content { height: 100%; display: flex; flex-direction: column; background-color: #fff; }
.sider-toolbar { padding: 20px 20px 12px 20px; display: flex; justify-content: space-between; align-items: center; }
.toolbar-title { font-weight: 800; font-size: 16px; color: #1e293b; display: flex; align-items: center; }
.sider-controls { padding: 0 20px; display: flex; flex-direction: column; gap: 12px; }
.filter-header { padding: 16px 20px 0 20px; font-weight: 700; color: #333; font-size: 13px; display: flex; align-items: center; gap: 6px; border-top: 1px dashed #e2e8f0; margin-top: 16px;}
.sider-scroll-area { flex: 1; overflow-y: auto; padding: 16px; }

/* 🔥 Feed 顶部双 Tab 设计 */
.feed-header {
    background-color: #fff;
    padding: 0 40px;
    border-bottom: 1px solid #e2e8f0;
    position: sticky;
    top: 0;
    z-index: 10;
}
.feed-tabs {
    display: flex;
    gap: 32px;
}
.feed-tab {
    padding: 16px 0;
    font-size: 16px;
    font-weight: 600;
    color: #64748b;
    cursor: pointer;
    position: relative;
    transition: color 0.2s;
}
.feed-tab:hover { color: #1e293b; }
.feed-tab.active-tab { color: #10b981; }

.tab-indicator {
    position: absolute;
    bottom: 0;
    left: 10%;
    width: 80%;
    height: 3px;
    background-color: #10b981;
    border-radius: 3px 3px 0 0;
}

/* Feed 瀑布流容器 */
.feed-main-scroll {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
}
.feed-content-wrapper {
    max-width: 720px;
    margin: 0 auto;
}

.filter-tag-row { margin-bottom: 20px; }
.pagination-wrapper { display: flex; justify-content: center; margin: 40px 0 60px 0; }

/* 模态框图片预览 */
.mini-img-preview { display: flex; gap: 8px; margin-top: 12px; }
.mini-img-item { position: relative; width: 60px; height: 60px; border-radius: 8px; overflow: hidden; cursor: pointer; }
.mini-img-item img { width: 100%; height: 100%; object-fit: cover; }
.overlay { position: absolute; inset: 0; background: rgba(0,0,0,0.6); color: #fff; font-size: 12px; display: flex; align-items: center; justify-content: center; opacity: 0; transition: opacity 0.2s; }
.mini-img-item:hover .overlay { opacity: 1; }

/* 📱 移动端悬浮球 */
.mobile-fabs { position: fixed; bottom: 80px; left: 20px; right: 20px; height: 0; display: flex; justify-content: space-between; pointer-events: none; z-index: 1000; }
.fab-btn { width: 48px; height: 48px; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; pointer-events: auto; transition: all 0.2s; backdrop-filter: blur(4px); }
.fab-btn:active { transform: scale(0.92); }
.left-fab { background: linear-gradient(135deg, #10b981 0%, #059669 100%); box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4); } 

@media (max-width: 768px) {
    .feed-header { padding: 0 20px; }
    .feed-main-scroll { padding: 16px; }
}
</style>