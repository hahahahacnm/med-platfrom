<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NCard, NDataTable, NButton, NTag, NInput, NPopconfirm, useMessage, 
  NImage, NIcon, NTabs, NTabPane, NAvatar, NModal, NDivider, NTooltip
} from 'naive-ui'
import { 
  SearchOutline, TrashOutline, AlertCircleOutline, ChatboxEllipsesOutline, EyeOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const data = ref([])
const showPreview = ref(false)
const currentQ = ref<any>(null)

const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const filter = reactive({ keyword: '', onlyReported: 'false', activeTab: 'all' })

/* 🔍 辅助函数 */
const parseOptions = (opts: any) => {
    if (!opts) return {}
    try { return typeof opts === 'string' ? JSON.parse(opts) : opts } catch { return {} }
}
const handlePreviewQ = (q: any) => {
    if (!q) return message.warning('题目数据不完整')
    currentQ.value = q; showPreview.value = true
}
const fmtDate = (s: string) => s ? new Date(s).toLocaleString() : '-'

/* 📊 列定义 */
const columns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '发布者', key: 'user', width: 180,
    render(row: any) {
      const u = row.user || {}, url = u.avatar ? `http://localhost:8080${u.avatar}` : undefined
      return h('div', { style: 'display:flex;align-items:center;gap:10px' }, [
        h(NAvatar, { round: true, size: 'small', src: url, fallbackSrc: "https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg", style: 'border:1px solid #eee' }),
        h('div', { style: 'line-height:1.2' }, [
            h('div', { style: 'font-weight:bold;font-size:13px' }, u.nickname || u.username || `User ${row.user_id}`),
            h('div', { style: 'font-size:12px;color:#999' }, `ID: ${u.id || row.user_id}`)
        ])
      ])
    }
  },
  { 
    title: '内容', key: 'content', width: 350,
    render(row: any) {
      let imgs = row.images || [], txt = row.content || ''
      if (!imgs.length) {
          const m = [...txt.matchAll(/\[图片:(.*?)\]/g)]
          if (m.length) { imgs = m.map(x => x[1]); txt = txt.replace(/\[图片:.*?\]/g, '') }
      }
      return h('div', { style: 'padding:4px 0' }, [
        h('div', { style: 'margin-bottom:6px;max-height:42px;overflow:hidden;text-overflow:ellipsis;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;font-size:14px;color:#333' }, txt.trim() || '(纯图片)'),
        imgs.length ? h('div', { style: 'display:flex;gap:4px' }, imgs.map((u: string) => 
            h(NImage, { src: `http://localhost:8080${u}`, width: 40, height: 40, objectFit: 'cover', style: 'border-radius:4px;border:1px solid #eee' })
        )) : null
      ])
    }
  },
  { 
    title: '关联题目', key: 'question', width: 140,
    render(row: any) {
        const q = row.question || {}
        return h('div', { style: 'cursor:pointer', onClick: () => handlePreviewQ(q), title: '点击查看' }, [
            h(NTag, { size: 'small', type: 'info', bordered: false, style: 'margin-bottom:4px' }, { default: () => `ID: ${q.id||row.question_id}`, icon: () => h(NIcon, null, { default: () => h(EyeOutline) }) }),
            h('div', { style: 'font-size:12px;color:#666' }, q.type || '未知题型')
        ])
    }
  },
  { 
    title: '状态', key: 'status', width: 110,
    render(row: any) {
        if (!row.is_reported) return h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '正常' })
        
        // 悬停显示举报理由
        const reasons = (row.reports || []).map((r: any) => `👤 用户${r.user_id}: ${r.reason}`).join('\n') || '暂无详细理由'
        return h(NTooltip, { trigger: 'hover' }, {
            trigger: () => h(NTag, { type: 'error', round: true, size: 'small', style: 'cursor:help' }, { icon: () => h(NIcon, null, { default: () => h(AlertCircleOutline) }), default: () => `举报 ${row.report_count}` }),
            default: () => h('div', { style: 'white-space:pre-wrap;max-width:300px' }, reasons)
        })
    }
  },
  { title: '时间', key: 'created_at', width: 150, render: (row: any) => fmtDate(row.created_at) },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 140,
    render(row: any) {
      const btns = []
      if (row.is_reported) btns.push(h(NButton, { size: 'tiny', type: 'warning', secondary: true, style: 'margin-right:8px', onClick: () => ignore(row) }, { default: () => '忽略' }))
      btns.push(h(NPopconfirm, { onPositiveClick: () => del(row) }, {
          trigger: () => h(NButton, { size: 'tiny', type: 'error', dashed: true }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }), default: () => '删除' }),
          default: () => '确定删除？举报记录也会一并清除。'
      }))
      return h('div', btns)
    }
  }
]

/* 🚀 交互逻辑 */
const fetch = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/notes', { params: { page: pagination.page, page_size: pagination.pageSize, keyword: filter.keyword, reported: filter.onlyReported } })
    data.value = res.data || []; pagination.itemCount = res.total
  } catch { message.error('加载失败') } finally { loading.value = false }
}
const onPage = (p: number) => { pagination.page = p; fetch() }
const ignore = async (row: any) => { try { await request.post(`/admin/notes/${row.id}/ignore`); message.success('已忽略'); fetch() } catch { message.error('失败') } }
const del = async (row: any) => { try { await request.delete(`/notes/${row.id}`); message.success('已删除'); fetch() } catch { message.error('失败') } }
const onTab = (v: string) => { filter.activeTab = v; filter.onlyReported = v === 'reported' ? 'true' : 'false'; pagination.page = 1; fetch() }

onMounted(fetch)
</script>

<template>
  <div class="page">
    <n-card class="card" :bordered="false">
        <div class="head">
            <h2 class="title"><n-icon color="#2080f0"><ChatboxEllipsesOutline/></n-icon> 评论管理</h2>
            <div class="acts">
                <n-input v-model:value="filter.keyword" placeholder="搜内容/用户" style="width:200px" clearable @keyup.enter="fetch"><template #prefix><n-icon><SearchOutline/></n-icon></template></n-input>
                <n-button type="primary" @click="fetch">查询</n-button>
            </div>
        </div>
        <n-tabs type="line" :value="filter.activeTab" @update:value="onTab" animated>
            <n-tab-pane name="all" tab="全部" />
            <n-tab-pane name="reported" tab="🚨 待审核" />
        </n-tabs>
        <n-data-table remote :columns="columns" :data="data" :loading="loading" :pagination="pagination" @update:page="onPage" :row-key="r=>r.id" style="margin-top:10px;height:calc(100vh - 220px)" flex-height />
    </n-card>

    <n-modal v-model:show="showPreview" preset="card" style="width:600px;max-width:90%" title="题目详情">
        <div v-if="currentQ">
            <n-tag type="success" size="small" style="margin-bottom:10px">{{ currentQ.type }}</n-tag>
            <div class="q-html" v-html="currentQ.stem"></div>
            <div class="q-opts" v-if="currentQ.options">
                <div v-for="(txt, k) in parseOptions(currentQ.options)" :key="k" class="opt"><span class="k">{{k}}.</span> <span v-html="txt"></span></div>
            </div>
            <n-divider dashed />
            <div class="q-an">
                <div style="font-weight:bold;color:#18a058;margin-bottom:4px">✅ 答案：{{ currentQ.correct||currentQ.answer }}</div>
                <div style="color:#666;font-size:13px;background:#f9f9f9;padding:8px;border-radius:4px"><strong>解析：</strong>{{ currentQ.analysis||'无' }}</div>
            </div>
        </div>
    </n-modal>
  </div>
</template>

<style scoped>
.page { padding: 16px; height: 100vh; background-color: #f5f7f9; }
.card { height: 100%; display: flex; flex-direction: column; }
.head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.title { margin: 0; font-size: 18px; display: flex; align-items: center; gap: 8px; }
.acts { display: flex; gap: 10px; }
.q-html { font-size: 15px; color: #333; line-height: 1.5; margin-bottom: 12px; }
.q-opts { display: flex; flex-direction: column; gap: 6px; }
.opt { font-size: 14px; color: #555; }
.k { font-weight: bold; color: #2080f0; margin-right: 6px; }
</style>