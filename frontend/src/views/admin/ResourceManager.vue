<script setup lang="ts">
import { ref, onMounted, h, reactive, computed, nextTick } from 'vue'
import { 
  NCard, NDataTable, NButton, NSpace, NTag, NInput, NTree, NLayout, NLayoutSider, NLayoutContent,
  NForm, NFormItem, NModal, useMessage, NPopconfirm, NIcon, NSpin, NGrid, NGi,
  NScrollbar, NDivider, NInputNumber, NEmpty, NUpload, NSelect, NAlert, NSwitch
} from 'naive-ui'
import type { UploadFileInfo, TreeDropInfo } from 'naive-ui'
import { 
  SearchOutline, TrashOutline, CreateOutline, RefreshOutline, FolderOpenOutline, BookOutline,
  SaveOutline, AddCircleOutline, RemoveCircleOutline, LibraryOutline, 
  ArrowForwardOutline, CloudUploadOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()

// ========================================================================
// 1. 🌟 超级目录树
// ========================================================================
const treeData = ref<any[]>([])
const treeLoading = ref(false)
const selectedKeys = ref<string[]>([])

const currentContext = reactive({
  type: 'none', 
  source: '',     
  categoryId: 0,
  categoryPath: '',
})

const createSourceNode = (name: string) => ({
  label: name,
  key: `__source__:${name}`, 
  type: 'source',
  isLeaf: false,
  prefix: () => h(NIcon, { color: '#18a058' }, { default: () => h(LibraryOutline) }),
  sourceName: name,
  draggable: false 
})

const createCategoryNode = (node: any, sourceName: string) => ({
  label: node.name,
  key: node.id, 
  type: 'category',
  isLeaf: node.is_leaf,
  prefix: () => h(NIcon, { color: node.level === 1 ? '#f0a020' : '#999' }, { default: () => h(node.level === 1 ? FolderOpenOutline : BookOutline) }),
  id: node.id,
  fullPath: node.full_path || node.name, 
  sourceName: sourceName,
  children: node.is_leaf ? [] : undefined 
})

const fetchRootSources = async () => {
  treeLoading.value = true
  try {
    const res: any = await request.get('/banks') 
    const list = res.data || []
    treeData.value = list.map(createSourceNode)
  } catch (e) { message.error('加载题库失败') } 
  finally { treeLoading.value = false }
}

const handleLoadTree = async (node: any) => {
  return new Promise<void>(async (resolve) => {
    try {
      let params: any = {}
      if (node.type === 'source') params = { source: node.sourceName, parent_id: 0 }
      else if (node.type === 'category') params = { source: node.sourceName, parent_id: node.id }

      const res: any = await request.get('/category-tree', { params })
      const rawData = Array.isArray(res) ? res : (res.data || [])
      node.children = rawData.map((item: any) => createCategoryNode(item, node.sourceName))
      resolve()
    } catch { node.children = []; resolve() }
  })
}

const handleNodeSelect = (keys: string[], option: any[]) => {
  selectedKeys.value = keys
  checkedRowKeys.value = [] 
  if (!keys.length || !option || !option.length) { currentContext.type = 'none'; return }
  
  const node = option[0]
  currentContext.source = node.sourceName

  if (node.type === 'source') {
    currentContext.type = 'source'; currentContext.categoryId = 0; currentContext.categoryPath = ''
  } else {
    currentContext.type = 'category'; currentContext.categoryId = node.id; currentContext.categoryPath = node.fullPath
    pagination.page = 1; fetchQuestions()
  }
}

const handleDrop = async ({ node, dragNode, dropPosition }: TreeDropInfo) => {
  if (node.sourceName !== dragNode.sourceName) { message.warning('不支持跨题库移动'); return }
  if (node.type === 'source' || dragNode.type === 'source') return

  const findSiblings = (nodes: any[]): any[] | null => {
    for (const item of nodes) {
      if (item.key === node.key) return nodes
      if (item.children) { const res = findSiblings(item.children); if (res) return res }
    }
    return null
  }
  const siblings = findSiblings(treeData.value)
  if (!siblings) return

  const dragIndex = siblings.findIndex((x: any) => x.key === dragNode.key)
  if (dragIndex === -1) return
  const [dragItem] = siblings.splice(dragIndex, 1)
  
  let targetIndex = siblings.findIndex((x: any) => x.key === node.key)
  if (dropPosition === 'before') siblings.splice(targetIndex, 0, dragItem)
  else if (dropPosition === 'after') siblings.splice(targetIndex + 1, 0, dragItem)
  else if (dropPosition === 'inside') {
     siblings.splice(dragIndex, 0, dragItem)
     message.info('暂不支持直接拖入内部')
     return
  }

  const updateItems = siblings.map((item: any, index: number) => ({ id: item.id, sort_order: index + 1 }))
  try { 
    await request.post('/admin/categories/reorder', { items: updateItems })
    message.success('顺序已更新') 
  } catch (e) { message.error('排序失败'); fetchRootSources() }
}

// ========================================================================
// 2. 🏦 控制台逻辑
// ========================================================================
const showImportModal = ref(false)
const showRenameModal = ref(false)
const importForm = ref({ bankName: '' })
const importFileList = ref<UploadFileInfo[]>([])
const importing = ref(false)
const renameForm = ref({ newName: '' })
const renaming = ref(false)

const handleImport = async () => {
  if (!importForm.value.bankName || !importFileList.value.length) { message.warning('请补全信息'); return }
  const formData = new FormData()
  formData.append('file', importFileList.value[0].file as File)
  formData.append('bank_name', importForm.value.bankName)
  importing.value = true
  try {
    await request.post('/admin/questions/import', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    message.success('导入成功')
    showImportModal.value = false; fetchRootSources()
  } catch(e) { message.error('导入失败') } finally { importing.value = false }
}

const handleRenameSource = async () => {
  if (!renameForm.value.newName) return; renaming.value = true
  try {
    await request.post('/admin/banks/rename', { old_name: currentContext.source, new_name: renameForm.value.newName })
    message.success('改名成功')
    showRenameModal.value = false; fetchRootSources(); currentContext.type = 'none'
  } catch(e) { message.error('失败') } finally { renaming.value = false }
}

const handleDeleteSource = async () => {
  try {
    await request.post('/admin/banks/delete', { source_name: currentContext.source })
    message.success('题库已彻底粉碎')
    fetchRootSources(); currentContext.type = 'none'
  } catch(e) { message.error('删除失败') }
}

// ========================================================================
// 3. 📋 题目列表逻辑
// ========================================================================
const tableLoading = ref(false)
const questionData = ref([])
const filters = reactive({ q: '', difficulty: null })
const checkedRowKeys = ref<number[]>([])

const pagination = reactive({ 
  page: 1, pageSize: 20, itemCount: 0, showSizePicker: true, pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }: any) => `共 ${itemCount} 道题`,
  onChange: (page: number) => { pagination.page = page; fetchQuestions() },
  onUpdatePageSize: (pageSize: number) => { pagination.pageSize = pageSize; pagination.page = 1; fetchQuestions() }
})

const columns = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '题型', key: 'type', width: 90, align: 'center',
    render: (row: any) => {
      const isParent = row.children && row.children.length > 0
      return h(NTag, { type: isParent ? 'warning' : 'success', size: 'small', bordered: false }, { default: () => row.type + (isParent ? ' (组)' : '') }) 
    }
  },
  { 
    title: '题干预览', key: 'stem', 
    render: (row: any) => {
      let text = (row.stem || '').replace(/<[^>]+>/g, '')
      if (text.length > 50) text = text.substring(0, 50) + '...'
      if (row.children && row.children.length > 0) {
         if (row.type.includes('B1')) return h('span', { style: 'color: #f0a020; font-weight: bold' }, '【共用选项】 ' + text)
         if (row.type.includes('A3') || row.type.includes('A4')) return h('span', { style: 'color: #18a058; font-weight: bold' }, '【案例描述】 ' + text)
      }
      return text
    }
  },
  { title: '操作', key: 'actions', width: 120, fixed: 'right', align: 'center',
    render(row: any) {
      return h(NSpace, { justify: 'center' }, { default: () => [
        h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => openEditor(row) }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }) }),
        h(NPopconfirm, { onPositiveClick: () => handleBatchDelete([row.id]) }, { trigger: () => h(NButton, { size: 'tiny', type: 'error', secondary: true }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }) }), default: () => '⚠️ 彻底删除？' })
      ]})
    }
  }
]

const fetchQuestions = async () => {
  if (currentContext.type !== 'category') return
  tableLoading.value = true
  checkedRowKeys.value = []
  try {
    const res: any = await request.get('/questions', { 
      params: { 
        page: pagination.page, page_size: pagination.pageSize, 
        category: currentContext.categoryPath, source: currentContext.source, q: filters.q 
      } 
    })
    
    let list = [], total = 0
    if (res && Array.isArray(res.data)) { list = res.data; total = res.total || 0 }
    else if (res && res.data && Array.isArray(res.data.data)) { list = res.data.data; total = res.data.total || 0 }
    else if (res && res.data) { list = res.data; total = res.total || 0 }

    questionData.value = list
    pagination.itemCount = total
  } catch (e) { message.error('加载失败') } finally { tableLoading.value = false }
}

const handleBatchDelete = async (ids: number[] = []) => {
  const targetIDs = ids.length > 0 ? ids : checkedRowKeys.value
  if (targetIDs.length === 0) return
  try {
    await request.post('/admin/questions/batch-delete', { ids: targetIDs })
    message.success(`已彻底删除 ${targetIDs.length} 项数据`)
    checkedRowKeys.value = [] 
    fetchQuestions()
  } catch(e) { message.error('删除失败') }
}

const handleCategoryDelete = async () => {
  try { 
    await request.delete('/admin/questions/by-category', { params: { category_path: currentContext.categoryPath, source: currentContext.source } })
    message.success('章节及目录结构已彻底粉碎')
    fetchRootSources(); currentContext.type = 'source'; questionData.value = []
  } catch { message.error('操作失败') }
}

// ========================================================================
// 4. ✏️ 全能编辑器 (🔥 适配主观题 🔥)
// ========================================================================
const showEditor = ref(false)
const saving = ref(false)
const editorModel = ref<any>({ id: 0, type: '', stem: '', options: {}, children: [] })
const FULL_KEYS = ['A', 'B', 'C', 'D', 'E', 'F']
const parentOptionKeys = ref<string[]>([]) 

const isGroup = computed(() => ['A3','A4','B1','案例'].some(t => editorModel.value.type?.toUpperCase().includes(t)))
const isB1 = computed(() => editorModel.value.type?.toUpperCase().includes('B1'))
const diffOptions = ['易','较易','中','较难','难'].map(x=>({label:x,value:x}))

// 🔥 判断是否为主观题 (简答、论述、名解、填空)
const checkIsSubjective = (typeStr: string) => {
    if (!typeStr) return false
    const t = typeStr.toUpperCase()
    return ['简答', '论述', '名解', '名词解释', '填空', '问答'].some(k => t.includes(k))
}

const isSubjective = computed(() => checkIsSubjective(editorModel.value.type))

const initDynamicKeys = (options: any) => {
    if (!options) return ['A', 'B', 'C', 'D'] 
    const keys = Object.keys(options).filter(k => FULL_KEYS.includes(k)).sort()
    return keys.length < 4 ? ['A', 'B', 'C', 'D'] : keys
}
const addOption = (targetKeys: string[]) => {
    const nextKey = FULL_KEYS.find(k => !targetKeys.includes(k))
    if (nextKey) { targetKeys.push(nextKey); targetKeys.sort() }
}
const removeOption = (targetKeys: string[], keyToRemove: string, targetOptionsObj: any) => {
    if (targetKeys.length <= 4) { message.warning('至少保留 4 个选项'); return }
    const idx = targetKeys.indexOf(keyToRemove)
    if (idx > -1) { targetKeys.splice(idx, 1); if (targetOptionsObj) delete targetOptionsObj[keyToRemove] }
}

const openEditor = async (row: any) => {
  try {
    const res: any = await request.get(`/questions/${row.id}`)
    const fullData = res.data || res
    const parseOpts = (opts: any) => {
        if (!opts) return {}
        if (typeof opts === 'string') { try { return JSON.parse(opts) } catch { return {} } }
        return opts
    }
    const parentOpts = parseOpts(fullData.options)
    editorModel.value = { ...fullData, options: parentOpts }
    
    // 如果不是主观题，初始化选项 Keys
    if (!checkIsSubjective(editorModel.value.type)) {
        parentOptionKeys.value = initDynamicKeys(parentOpts)
    }
    
    if (editorModel.value.children?.length > 0) {
        editorModel.value.children = editorModel.value.children.map((child: any) => {
            const childOpts = parseOpts(child.options)
            const isSub = checkIsSubjective(child.type)
            return { 
                ...child, 
                options: childOpts, 
                dynamicKeys: isSub ? [] : initDynamicKeys(childOpts),
                isSubjective: isSub 
            }
        })
    } else { editorModel.value.children = [] }
    showEditor.value = true
  } catch (e) { message.error('加载详情失败') }
}

const setCorrectAnswer = (optionKey: string, childIndex: number = -1) => {
  if (childIndex === -1) editorModel.value.correct = optionKey
  else editorModel.value.children[childIndex].correct = optionKey
}

const handleSaveAll = async () => {
  saving.value = true
  try {
    const cleanOptions = (opts: any, activeKeys: string[]) => {
        if (!activeKeys || activeKeys.length === 0) return null 
        const result: any = {}
        let hasContent = false
        activeKeys.forEach(k => { if (opts[k]) { result[k] = opts[k]; hasContent = true } })
        return hasContent ? result : null
    }

    if (isB1.value && editorModel.value.children?.length > 0) {
        editorModel.value.stem = editorModel.value.children[0].stem
    }

    // 父题选项处理：主观题不需要 Options
    let parentOptsToSave = null
    if (!isSubjective.value) {
        if (isB1.value || !isGroup.value) {
            parentOptsToSave = cleanOptions(editorModel.value.options, parentOptionKeys.value)
        }
    }

    const parentPayload = { ...editorModel.value, options: parentOptsToSave }
    delete parentPayload.children 
    await request.put(`/admin/questions/${editorModel.value.id}`, parentPayload)

    // 子题保存
    if (editorModel.value.children?.length > 0) {
        for (const child of editorModel.value.children) {
            // 如果子题是主观题，Options 设为 null
            const childKeys = (!isB1.value && !child.isSubjective) ? child.dynamicKeys : []
            await request.put(`/admin/questions/${child.id}`, { 
                ...child, 
                options: cleanOptions(child.options, childKeys) 
            })
        }
    }
    message.success('保存成功')
    showEditor.value = false
    fetchQuestions()
  } catch (e) { message.error('保存失败') } finally { saving.value = false }
}

onMounted(() => { fetchRootSources() })
</script>

<template>
  <div class="resource-manager">
    <n-layout has-sider class="full-height">
      <n-layout-sider bordered width="320" content-style="padding: 12px; background: #fff;" collapse-mode="width" show-trigger>
        <div class="sider-header">
          <div style="font-weight: 800; font-size: 16px; color: #333;">🗂️ 资源管理器</div>
          <n-space>
             <n-button size="tiny" secondary type="primary" @click="showImportModal = true" title="导入">
                <template #icon><n-icon><CloudUploadOutline /></n-icon></template>
             </n-button>
             <n-button size="tiny" circle secondary @click="fetchRootSources">
                <template #icon><n-icon><RefreshOutline /></n-icon></template>
             </n-button>
          </n-space>
        </div>
        
        <n-spin :show="treeLoading">
          <n-tree 
            block-line remote draggable 
            :data="treeData" 
            :selected-keys="selectedKeys"
            :on-load="handleLoadTree" 
            @update:selected-keys="handleNodeSelect"
            @drop="handleDrop"
            expand-on-click
            class="custom-tree"
          />
        </n-spin>
        <div style="font-size: 12px; color: #999; margin-top: 10px; text-align: center;">
           * 可拖拽调整同级章节顺序
        </div>
      </n-layout-sider>

      <n-layout-content content-style="padding: 0; background-color: #f5f7fa;">
        
        <div v-if="currentContext.type === 'source'" class="dashboard-panel">
           <n-empty description="当前选中题库源" size="large" style="margin-top: 50px;">
              <template #extra>
                 <div style="text-align: center;">
                    <h2 style="margin: 0 0 20px 0;">{{ currentContext.source }}</h2>
                    <n-space justify="center">
                       <n-button size="large" @click="showRenameModal = true"><template #icon><n-icon><CreateOutline/></n-icon></template> 重命名</n-button>
                       <n-popconfirm @positive-click="handleDeleteSource">
                          <template #trigger>
                             <n-button size="large" type="error" ghost><template #icon><n-icon><TrashOutline/></n-icon></template> 删库跑路</n-button>
                          </template>
                          危险操作：确定删除【{{ currentContext.source }}】及其所有内容吗？
                       </n-popconfirm>
                    </n-space>
                    <p style="color: #999; margin-top: 20px;">请点击左侧展开目录，查看具体章节题目</p>
                 </div>
              </template>
           </n-empty>
        </div>

        <div v-else-if="currentContext.type === 'category'" class="table-panel">
           <div class="panel-header">
              <div class="breadcrumb">
                 <n-tag :bordered="false" type="success">{{ currentContext.source }}</n-tag>
                 <span class="sep">/</span>
                 <span class="path">{{ currentContext.categoryPath }}</span>
              </div>
              
              <div class="actions">
                 <n-space>
                    <n-popconfirm v-if="checkedRowKeys.length > 0" @positive-click="handleBatchDelete()">
                       <template #trigger>
                          <n-button size="small" type="error">
                             <template #icon><n-icon><TrashOutline/></n-icon></template> 批量删除 ({{ checkedRowKeys.length }})
                          </n-button>
                       </template>
                       确定将选中的 {{ checkedRowKeys.length }} 道题目彻底删除？
                    </n-popconfirm>

                    <n-popconfirm @positive-click="handleCategoryDelete">
                       <template #trigger>
                          <n-button size="small" type="error" dashed><template #icon><n-icon><TrashOutline/></n-icon></template> 删除本章</n-button>
                       </template>
                       ⚠️ 危险：将彻底删除本章节目录及其下所有题目！
                    </n-popconfirm>
                 </n-space>
              </div>
           </div>

           <div class="filter-bar">
              <n-input v-model:value="filters.q" placeholder="搜索本章题目..." style="width: 300px" @keyup.enter="fetchQuestions">
                 <template #prefix><n-icon><SearchOutline/></n-icon></template>
              </n-input>
              <n-button type="primary" @click="fetchQuestions">查询</n-button>
           </div>

           <div class="table-wrapper">
              <n-data-table 
                 remote 
                 v-model:checked-row-keys="checkedRowKeys"
                 :columns="columns" 
                 :data="questionData" 
                 :loading="tableLoading" 
                 :pagination="pagination" 
                 :row-key="r=>r.id"
                 children-key="hw_ignore_children"
                 :max-height="650"
                 style="min-height: 200px"
              />
           </div>
        </div>

        <div v-else class="empty-state">
           <n-icon size="64" color="#ddd"><ArrowForwardOutline /></n-icon>
           <p style="margin-top: 10px; color: #999">请在左侧选择题库或章节</p>
        </div>

      </n-layout-content>
    </n-layout>

    <n-modal v-model:show="showImportModal" preset="card" title="导入新题库" style="width: 500px">
       <n-form>
          <n-form-item label="1. 题库名称"><n-input v-model:value="importForm.bankName" placeholder="例如：2025年真题" /></n-form-item>
          <n-form-item label="2. Excel文件">
             <n-upload v-model:file-list="importFileList" :max="1" accept=".xlsx"><n-button>选择文件</n-button></n-upload>
          </n-form-item>
       </n-form>
       <template #footer><n-button type="primary" @click="handleImport" :loading="importing" block>开始导入</n-button></template>
    </n-modal>

    <n-modal v-model:show="showRenameModal" preset="card" title="重命名题库" style="width: 400px">
       <n-input v-model:value="renameForm.newName" placeholder="新名称" />
       <template #footer><n-button type="primary" @click="handleRenameSource" :loading="renaming" block>保存</n-button></template>
    </n-modal>

    <n-modal v-model:show="showEditor" style="width: 900px; max-width: 98%;" preset="card" :title="isGroup ? `题组编辑器 (${editorModel.type})` : '单题编辑器'" :bordered="false">
      <n-scrollbar style="max-height: 75vh; padding-right: 12px;">
        <div v-if="isGroup">
           <n-card :bordered="false" style="margin-bottom: 20px; border-left: 4px solid #18a058; background-color: #fcfcfc;">
              <div style="font-weight: bold; margin-bottom: 12px; color: #666; font-size: 15px;">
                {{ isB1 ? '🧩 共用备选答案 (父题)' : '📚 案例描述 / 共用题干 (父题)' }}
              </div>
              <n-input v-if="!isB1" type="textarea" v-model:value="editorModel.stem" :rows="3" placeholder="在此编辑案例材料..." />
              
              <div v-else>
                 <n-grid :cols="1" :y-gap="8">
                    <n-gi v-for="key in parentOptionKeys" :key="key">
                       <div style="display: flex; align-items: center;">
                          <n-input v-model:value="editorModel.options[key]" :placeholder="`共用选项 ${key}`" style="flex: 1"><template #prefix><b style="color: #d68b00">{{key}}.</b></template></n-input>
                          <n-button v-if="parentOptionKeys.length > 4" circle size="small" type="error" text style="margin-left: 8px" @click="removeOption(parentOptionKeys, key, editorModel.options)"><template #icon><n-icon size="20"><RemoveCircleOutline /></n-icon></template></n-button>
                       </div>
                    </n-gi>
                 </n-grid>
                 <div v-if="parentOptionKeys.length < 6" style="margin-top: 10px; text-align: center;"><n-button dashed size="small" type="primary" @click="addOption(parentOptionKeys)">增加选项 (至 F)</n-button></div>
              </div>
           </n-card>
           <n-divider dashed>👇 子题列表 (请在此处修改子题)</n-divider>
           
           <div v-for="(child, index) in editorModel.children" :key="child.id" style="margin-bottom: 30px; border: 1px solid #eee; padding: 15px; border-radius: 8px;">
              <div style="display: flex; justify-content: space-between; margin-bottom: 10px;">
                 <span style="font-weight: bold; color: #333;">第 {{ index + 1 }} 小题 <n-tag size="small">{{ child.type }}</n-tag></span>
              </div>
              <n-input v-model:value="child.stem" type="textarea" :rows="2" placeholder="编辑小题题干..." style="margin-bottom: 12px;" />
              
              <div v-if="child.isSubjective" style="background: #fff8f0; padding: 12px; border: 1px solid #ffeeba; border-radius: 6px; margin-bottom: 12px;">
                 <div style="margin-bottom: 8px; font-size: 12px; font-weight: bold; color: #d68b00;">✏️ 参考答案 (主观题)</div>
                 <n-input v-model:value="child.correct" type="textarea" :rows="3" placeholder="在此输入参考答案文本..." />
              </div>

              <div v-else-if="isB1" style="background: #f9f9f9; padding: 12px; border-radius: 6px; margin-bottom: 12px;">
                 <div style="margin-bottom: 8px; font-size: 12px; color: #999;"> 👇 设定正确答案</div>
                 <n-space><n-button v-for="k in parentOptionKeys" :key="k" circle :type="child.correct === k ? 'success' : 'default'" @click="setCorrectAnswer(k, index)">{{ k }}</n-button></n-space>
              </div>

              <div v-else style="background: #fdfdfd; padding: 12px; border: 1px solid #f0f0f0; border-radius: 6px; margin-bottom: 12px;">
                 <div style="margin-bottom: 8px; font-size: 12px; font-weight: bold; color: #666;">选项与答案</div>
                 <n-grid :cols="1" :y-gap="12">
                    <n-gi v-for="k in child.dynamicKeys" :key="k">
                       <div style="display: flex; align-items: center;">
                          <n-button circle size="small" :type="child.correct === k ? 'success' : 'default'" @click="setCorrectAnswer(k, index)" style="margin-right: 12px; font-weight: bold;">{{ k }}</n-button>
                          <n-input v-model:value="child.options[k]" :placeholder="`选项 ${k} 内容`" style="flex: 1" />
                          <n-button v-if="child.dynamicKeys.length > 4" circle size="small" type="error" text style="margin-left: 5px" @click="removeOption(child.dynamicKeys, k, child.options)"><n-icon><RemoveCircleOutline /></n-icon></n-button>
                       </div>
                    </n-gi>
                 </n-grid>
                 <div v-if="child.dynamicKeys.length < 6" style="margin-top: 10px; text-align: center;"><n-button dashed size="small" type="primary" @click="addOption(child.dynamicKeys)">加选项</n-button></div>
              </div>

              <div style="background: #fff; border-top: 1px dashed #eee; padding-top: 12px;">
                 <n-grid :cols="4" :x-gap="12" :y-gap="12">
                    <n-gi><div style="font-size: 12px; color: #999;">难度</div><n-select v-model:value="child.difficulty" size="small" :options="diffOptions" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">区分度</div><n-input-number v-model:value="child.diff_value" size="small" :step="0.1" :min="0" :max="1" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">认知层次</div><n-input v-model:value="child.cognitive_level" size="small" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">大纲要求</div><n-input v-model:value="child.syllabus" size="small" /></n-gi>
                    <n-gi :span="4"><div style="font-size: 12px; color: #999;">答案解析</div><n-input v-model:value="child.analysis" type="textarea" :rows="2" /></n-gi>
                 </n-grid>
              </div>
           </div>
        </div>

        <div v-else>
           <n-card :bordered="false" style="border-radius: 8px;">
              <div style="font-weight: bold; margin-bottom: 8px; font-size: 16px;">
                 题目内容 <n-tag size="small" type="info" style="margin-left: 8px">{{ editorModel.type }}</n-tag>
              </div>
              <n-input type="textarea" v-model:value="editorModel.stem" :rows="4" placeholder="在此输入题干..." />
              <n-divider />
              
              <div v-if="isSubjective" style="background: #fff8f0; padding: 16px; border: 1px solid #ffeeba; border-radius: 8px;">
                 <div style="margin-bottom: 8px; font-weight: bold; color: #d68b00;">✏️ 参考答案 (主观题)</div>
                 <n-input v-model:value="editorModel.correct" type="textarea" :rows="5" placeholder="在此输入参考答案文本..." />
              </div>

              <div v-else>
                 <div style="font-weight: bold; margin-bottom: 12px;">选项与答案</div>
                 <div style="background: #fdfdfd; padding: 16px; border: 1px solid #f0f0f0; border-radius: 8px;">
                    <n-grid :cols="1" :y-gap="12">
                       <n-gi v-for="key in parentOptionKeys" :key="key">
                          <div style="display: flex; align-items: center;">
                             <n-button circle size="small" :type="editorModel.correct === key ? 'success' : 'default'" @click="setCorrectAnswer(key)" style="margin-right: 12px; font-weight: bold;">{{ key }}</n-button>
                             <n-input v-model:value="editorModel.options[key]" :placeholder="`选项 ${key} 内容`" style="flex: 1" />
                             <n-button v-if="parentOptionKeys.length > 4" circle size="small" type="error" text style="margin-left: 8px" @click="removeOption(parentOptionKeys, key, editorModel.options)"><template #icon><n-icon size="20"><RemoveCircleOutline /></n-icon></template></n-button>
                          </div>
                       </n-gi>
                    </n-grid>
                    <div v-if="parentOptionKeys.length < 6" style="margin-top: 15px; text-align: center;"><n-button dashed size="small" type="primary" @click="addOption(parentOptionKeys)">增加选项</n-button></div>
                 </div>
              </div>

              <n-divider />
              <div style="font-weight: bold; margin-bottom: 12px;">题目属性 & 解析</div>
              <div style="background: #f9f9f9; padding: 16px; border-radius: 8px;">
                 <n-grid :cols="4" :x-gap="12" :y-gap="12">
                    <n-gi><div style="font-size: 12px; color: #999;">难度</div><n-select v-model:value="editorModel.difficulty" :options="diffOptions" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">区分度</div><n-input-number v-model:value="editorModel.diff_value" :step="0.1" :min="0" :max="1" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">认知层次</div><n-input v-model:value="editorModel.cognitive_level" /></n-gi>
                    <n-gi><div style="font-size: 12px; color: #999;">大纲要求</div><n-input v-model:value="editorModel.syllabus" /></n-gi>
                    <n-gi :span="4"><div style="font-size: 12px; color: #999;">答案解析</div><n-input v-model:value="editorModel.analysis" type="textarea" :rows="3" /></n-gi>
                 </n-grid>
              </div>
           </n-card>
        </div>
      </n-scrollbar>
      <template #footer>
         <div style="display: flex; justify-content: space-between; align-items: center;">
            <div style="color: #999; font-size: 12px;"><span v-if="isGroup">* 保存将同步更新父题及 {{ editorModel.children.length }} 道子题</span></div>
            <n-space><n-button @click="showEditor = false">取消</n-button><n-button type="primary" size="large" @click="handleSaveAll" :loading="saving"><template #icon><n-icon><SaveOutline /></n-icon></template>保存全部修改</n-button></n-space>
         </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.resource-manager { height: 100vh; background-color: #fff; }
.full-height { height: 100%; }
.sider-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #f0f0f0; }
.dashboard-panel { height: 100%; display: flex; justify-content: center; align-items: center; background: #fff; }
.table-panel { height: 100%; display: flex; flex-direction: column; background: #f5f7fa; padding: 24px; }
.empty-state { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #999; }
.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.breadcrumb { font-size: 16px; font-weight: bold; color: #333; display: flex; align-items: center; gap: 8px; }
.sep { color: #ccc; font-weight: normal; }
.filter-bar { background: #fff; padding: 12px; border-radius: 8px; margin-bottom: 12px; display: flex; gap: 12px; }
.table-wrapper { flex: 1; background: #fff; padding: 12px; border-radius: 8px; overflow: hidden; }
:deep(.n-tree-node) { padding: 4px 0; }
</style>