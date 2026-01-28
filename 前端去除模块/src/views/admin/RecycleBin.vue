<script setup lang="ts">
import { ref, onMounted, h, reactive } from 'vue'
import { 
  NCard, NDataTable, NButton, NSpace, NTag, NPopconfirm, useMessage, 
  NPageHeader, NIcon, NEmpty, NModal, NScrollbar, NSpin, NText 
} from 'naive-ui'
import { 
  RefreshOutline, CloseCircleOutline, ArrowBackOutline, EyeOutline 
} from '@vicons/ionicons5'
import { useRouter } from 'vue-router'
import request from '../../utils/request' // 确保这里指向您的 axios 封装
import AdminQuestionPreview from '../../components/AdminQuestionPreview.vue' // 您的预览组件

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const data = ref([])

// --- 预览相关 ---
const showPreview = ref(false)
const currentQuestion = ref<any>(null)
const previewLoading = ref(false)

const pagination = reactive({
  page: 1, 
  pageSize: 10, 
  itemCount: 0, 
  showSizePicker: true, 
  pageSizes: [10, 20, 50, 100],
  onChange: (p: number) => { pagination.page = p; fetchData() },
  onUpdatePageSize: (ps: number) => { pagination.pageSize = ps; pagination.page = 1; fetchData() }
})

// 🚀 核心：点击预览时，重新请求详情以获取子题 (Children)
const handlePreview = async (row: any) => {
  showPreview.value = true
  previewLoading.value = true
  currentQuestion.value = null // 先置空
  
  try {
    // 注意：这里请求的是详情接口。
    // 如果您的后端 GetDetail 接口默认过滤掉了已删除的题目，这里可能会报 404。
    // 如果报 404，我们做个降级处理，显示列表里的基本信息。
    const res: any = await request.get(`/questions/${row.id}`)
    
    // 兼容后端直接返回对象或返回 { data: ... }
    const qData = res.data || res 
    
    if (qData && qData.id) {
      currentQuestion.value = qData
    } else {
      // 降级：如果详情查不到（比如后端做了软删除过滤），就用列表里的数据凑合显示
      currentQuestion.value = { ...row, children: [] } 
      message.warning('无法加载完整子题详情，仅显示基础信息')
    }
  } catch (e) {
    console.error(e)
    // 降级显示
    currentQuestion.value = { ...row, children: [] }
    message.warning('预览详情加载失败，显示基础信息')
  } finally {
    previewLoading.value = false
  }
}

const columns = [
  { title: 'ID', key: 'id', width: 70 },
  { 
    title: '题型', key: 'type', width: 90, 
    render: (row: any) => h(NTag, { 
      type: ['A3', 'A4', 'B1', '案例分析'].some(t => row.type?.includes(t)) ? 'warning' : 'default', 
      size: 'small' 
    }, { default: () => row.type }) 
  },
  { 
    title: '原题库', key: 'source', width: 120, 
    render: (row:any) => h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.source || '未知' }) 
  },
  { 
    title: '分类路径', key: 'category_path', ellipsis: { tooltip: true },
    render: (row: any) => {
        if(!row.category_path) return '-'
        const parts = row.category_path.split(' > ')
        return parts.length > 2 ? '... > ' + parts.slice(-2).join(' > ') : row.category_path
    }
  },
  
  // 🔥🔥🔥 核心修复位置：题干列 🔥🔥🔥
  { 
    title: '题干 (点击预览)', 
    key: 'stem', 
    // 1. 设置 ellipsis 让表格列不要无限撑开，超出部分显示 ... 并提供 tooltip
    ellipsis: { tooltip: true }, 
    render(row: any) {
      // 去除HTML标签
      const text = (row.stem || '').replace(/<[^>]+>/g, '')
      
      // 组合题提示
      const isGroup = ['A3', 'A4', 'B1', '案例'].some(t => row.type?.toUpperCase().includes(t))
      const extraInfo = isGroup ? ' [组合题]' : ''
      const displayText = text + extraInfo

      // 2. 渲染按钮，但内部加一个 div 强制限制样式
      return h(NButton, 
        { 
            text: true, 
            type: 'primary', 
            onClick: () => handlePreview(row),
            // 关键：防止按钮撑破单元格
            style: { maxWidth: '100%', verticalAlign: 'middle' } 
        }, 
        { 
            // 使用 div 包裹文字，利用 CSS 强制截断
            default: () => h('div', {
                style: {
                    maxWidth: '400px', // 限制文字最大宽度，防止撞车
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                    display: 'inline-block',
                    verticalAlign: 'bottom'
                }
            }, displayText) 
        }
      )
    }
  },
  
  { 
    title: '删除时间', key: 'deleted_at', width: 170, 
    render: (row:any) => row.deleted_at ? new Date(row.deleted_at).toLocaleString() : '-' 
  },
  { 
    title: '操作', key: 'actions', width: 160, fixed: 'right',
    render(row: any) {
      return h(NSpace, {}, { default: () => [
          h(NPopconfirm, 
            { onPositiveClick: () => handleRestore(row.id) }, 
            { trigger: () => h(NButton, { size: 'tiny', type: 'primary', secondary: true }, { default: () => '恢复' }), default: () => '确定恢复？' }
          ),
          h(NPopconfirm, 
            { onPositiveClick: () => handleHardDelete(row.id) }, 
            { trigger: () => h(NButton, { size: 'tiny', type: 'error', secondary: true }, { default: () => '粉碎' }), default: () => '彻底删除？' }
          )
      ]})
    }
  }
]

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/recycle-bin', { 
        params: { page: pagination.page, page_size: pagination.pageSize } 
    })
    
    // 适配后端返回结构 { data: [], total: 100, page: 1 ... }
    if (res && Array.isArray(res.data)) {
       data.value = res.data
       pagination.itemCount = res.total || 0
    } else {
       // 容错处理
       data.value = []
       pagination.itemCount = 0
    }
  } catch (e) { 
      console.error(e)
      message.error('获取回收站列表失败')
  } finally { 
      loading.value = false 
  }
}

const handleRestore = async (id: number) => { 
    try { 
        await request.post(`/admin/recycle-bin/${id}/restore`)
        message.success('已恢复')
        fetchData() 
    } catch { 
        message.error('恢复失败') 
    } 
}

const handleHardDelete = async (id: number) => { 
    try { 
        await request.delete(`/admin/recycle-bin/${id}`)
        message.success('已彻底粉碎')
        fetchData() 
    } catch { 
        message.error('删除失败') 
    } 
}

const handleEmptyAll = async () => { 
    try { 
        await request.delete('/admin/recycle-bin/empty')
        message.success('回收站已清空')
        fetchData() 
    } catch { 
        message.error('清空失败') 
    } 
}

onMounted(fetchData)
</script>

<template>
  <div class="recycle-container">
    <n-page-header @back="router.back()" style="margin-bottom: 24px;">
      <template #title>
          <span style="font-weight: 800; font-size: 20px; color: #333;">🗑️ 题目回收站</span>
      </template>
      <template #icon><n-icon><ArrowBackOutline /></n-icon></template>
      <template #extra>
        <n-space>
          <n-popconfirm @positive-click="handleEmptyAll">
            <template #trigger>
                <n-button type="error" :disabled="pagination.itemCount === 0">
                    <template #icon><n-icon><CloseCircleOutline /></n-icon></template>
                    一键清空
                </n-button>
            </template>
            <div style="max-width: 300px;">
                <p style="color: red; font-weight: bold;">⚠️ 高危操作警告</p>
                <p>确定要清空回收站内的所有题目吗？</p>
                <p>包括所有关联的<b>收藏、错题、笔记</b>都将被物理粉碎，且<b>无法找回</b>！</p>
            </div>
          </n-popconfirm>
          <n-button secondary circle @click="fetchData">
              <template #icon><n-icon><RefreshOutline /></n-icon></template>
          </n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-card :bordered="false" style="border-radius: 12px; min-height: 75vh; box-shadow: 0 2px 12px rgba(0,0,0,0.05);">
      <n-data-table 
        remote 
        :columns="columns" 
        :data="data" 
        :loading="loading" 
        :pagination="pagination" 
        :bordered="false" 
        :row-key="(row) => row.id"
        size="large"
      />
    </n-card>

    <n-modal v-model:show="showPreview" style="width: 900px; max-width: 95%;">
       <n-card 
         title="题目详情预览" 
         :bordered="false" 
         size="huge" 
         role="dialog" 
         aria-modal="true" 
         closable 
         @close="showPreview = false"
       >
          <template #header-extra>
             <n-tag v-if="currentQuestion?.type" type="success">{{ currentQuestion.type }}</n-tag>
          </template>
          
          <n-scrollbar style="max-height: 75vh; padding-right: 12px;">
             <n-spin :show="previewLoading" description="正在加载子题数据...">
                <div v-if="currentQuestion">
                   <AdminQuestionPreview :question="currentQuestion" />
                   
                   <n-empty 
                     v-if="['A3','A4','B1'].some(t=>currentQuestion.type?.includes(t)) && (!currentQuestion.children || currentQuestion.children.length === 0)" 
                     description="未检测到子题数据（请确认后端 GetDetail 接口是否包含 children）"
                     style="margin-top: 20px;"
                   />
                </div>
                <NEmpty v-else-if="!previewLoading" description="暂无数据" />
             </n-spin>
          </n-scrollbar>
       </n-card>
    </n-modal>
  </div>
</template>

<style scoped>
.recycle-container { 
    min-height: 100vh; 
    background-color: #f5f7fa; 
    padding: 24px; 
}
:deep(.n-data-table .n-data-table-td) {
    vertical-align: middle;
}
</style>