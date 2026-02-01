<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NDataTable, NButton, NTag, NInput, NModal, useMessage, 
  NTabs, NTabPane, NSpace, NAvatar, NDivider, NIcon, NTooltip
} from 'naive-ui'
import { 
  CheckmarkCircleOutline, CloseCircleOutline, BuildOutline, 
  EyeOutline, ChatboxEllipsesOutline 
} from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()
const loading = ref(false)
const data = ref([])
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const activeStatus = ref('0') // '0':待办, '1':已修复, '2':忽略

// 弹窗状态
const processModal = reactive({ show: false, id: 0, reply: '', status: 0, loading: false })
const previewModal = reactive({ show: false, q: null as any })

// 辅助函数：解析题目选项
const parseOptions = (opts: any) => {
    try { return typeof opts === 'string' ? JSON.parse(opts) : opts } catch { return {} }
}

// 打开题目预览
const openPreview = (q: any) => {
    if (!q) return message.warning('关联题目不存在')
    previewModal.q = q
    previewModal.show = true
}

const columns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '关联题目', key: 'question_id', width: 120,
    render(row: any) {
        return h(NButton, { 
            size: 'tiny', secondary: true, type: 'info', 
            onClick: () => openPreview(row.question) 
        }, { 
            icon: () => h(NIcon, null, { default: () => h(EyeOutline) }),
            default: () => `题号 ${row.question_id}` 
        })
    }
  },
  { 
    title: '反馈类型', key: 'type', width: 120, 
    render: (row: any) => h(NTag, { type: 'warning', bordered: false, size: 'small' }, { default: () => row.type }) 
  },
  { 
    title: '问题描述', key: 'content', width: 300,
    render(row: any) {
        return h('div', { style: 'white-space: pre-wrap; font-size: 13px;' }, row.content)
    }
  },
  { 
    title: '提交人', key: 'user', width: 150,
    render(row: any) {
        const u = row.user || {}
        return h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
            h(NAvatar, { round: true, size: 24, src: u.avatar ? `http://localhost:8080${u.avatar}` : undefined, fallbackSrc: "https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" }),
            h('span', { style: 'font-size:12px' }, u.nickname || u.username)
        ])
    }
  },
  { 
    title: '时间', key: 'created_at', width: 160, 
    render: (r: any) => new Date(r.created_at).toLocaleString() 
  },
  {
    title: '操作/状态', key: 'actions', fixed: 'right', width: 180,
    render(row: any) {
        // 如果是待处理状态 (0)
        if (row.status === 0) {
            return h(NSpace, { size: 'small' }, { default: () => [
                h(NButton, { size: 'tiny', type: 'primary', onClick: () => openProcess(row, 1) }, { icon: () => h(NIcon, null, { default: () => h(CheckmarkCircleOutline) }), default: () => '修复' }),
                h(NButton, { size: 'tiny', type: 'error', dashed: true, onClick: () => openProcess(row, 2) }, { icon: () => h(NIcon, null, { default: () => h(CloseCircleOutline) }), default: () => '忽略' })
            ]})
        }
        // 如果已处理
        const isFixed = row.status === 1
        return h(NTooltip, { trigger: 'hover' }, {
            trigger: () => h(NTag, { type: isFixed ? 'success' : 'default', bordered: false }, { default: () => isFixed ? '✅ 已修复' : '🚫 已忽略' }),
            default: () => `管理员回复: ${row.admin_reply || '无'}`
        })
    }
  }
]

const fetchData = async () => {
    loading.value = true
    try {
        const res: any = await request.get('/admin/feedbacks', {
            params: { page: pagination.page, page_size: pagination.pageSize, status: activeStatus.value }
        })
        data.value = res.data || []
        pagination.itemCount = res.total
    } finally { loading.value = false }
}

const openProcess = (row: any, status: number) => {
    processModal.id = row.id
    processModal.status = status
    // 自动填充默认回复
    processModal.reply = status === 1 ? '感谢反馈，经核查已修正该问题！' : '经核查，该题目内容无误，感谢您的反馈。'
    processModal.show = true
}

const submitProcess = async () => {
    processModal.loading = true
    try {
        await request.put(`/admin/feedbacks/${processModal.id}`, {
            status: processModal.status,
            admin_reply: processModal.reply
        })
        message.success('处理完成')
        processModal.show = false
        fetchData()
    } catch { message.error('操作失败') }
    finally { processModal.loading = false }
}

onMounted(fetchData)
</script>

<template>
  <div class="feedback-page">
    <n-card class="main-card" :bordered="false">
        <div class="header">
            <h2 class="title"><n-icon color="#f0a020"><BuildOutline/></n-icon> 题目纠错看板</h2>
            <n-button size="small" @click="fetchData">刷新</n-button>
        </div>

        <n-tabs type="line" v-model:value="activeStatus" @update:value="()=>{pagination.page=1;fetchData()}">
            <n-tab-pane name="0" tab="⏳ 待处理" />
            <n-tab-pane name="1" tab="✅ 已修复" />
            <n-tab-pane name="2" tab="🚫 已忽略" />
        </n-tabs>

        <n-data-table 
            remote 
            :columns="columns" 
            :data="data" 
            :loading="loading" 
            :pagination="pagination" 
            @update:page="(p)=>{pagination.page=p;fetchData()}" 
            flex-height 
            style="height: calc(100vh - 220px); margin-top: 12px;"
        />
    </n-card>

    <n-modal v-model:show="processModal.show" preset="dialog" :title="processModal.status===1 ? '🛠️ 确认修复' : '🚫 确认忽略'">
        <div style="padding: 10px 0;">
            <div style="margin-bottom: 8px; font-weight: bold; color: #666;">回复用户（可选）：</div>
            <n-input 
                v-model:value="processModal.reply" 
                type="textarea" 
                :autosize="{ minRows: 3, maxRows: 5 }"
                placeholder="给用户一句暖心的回复吧..." 
            />
        </div>
        <template #action>
            <n-button @click="processModal.show=false">取消</n-button>
            <n-button :type="processModal.status===1?'primary':'error'" :loading="processModal.loading" @click="submitProcess">
                {{ processModal.status===1 ? '确认标记为已修复' : '确认忽略' }}
            </n-button>
        </template>
    </n-modal>

    <n-modal v-model:show="previewModal.show" preset="card" style="width: 600px; max-width: 90%;" title="题目详情">
        <div v-if="previewModal.q">
            <n-tag type="success" size="small" style="margin-bottom: 12px">{{ previewModal.q.type }}</n-tag>
            <div class="q-content" v-html="previewModal.q.stem"></div>
            <div class="q-options" v-if="previewModal.q.options">
                <div v-for="(txt, k) in parseOptions(previewModal.q.options)" :key="k" class="opt">
                    <span class="k">{{k}}.</span> <span v-html="txt"></span>
                </div>
            </div>
            <n-divider dashed />
            <div class="q-ans">
                <div style="font-weight:bold;color:#18a058;margin-bottom:4px">✅ 正确答案：{{ previewModal.q.correct || previewModal.q.answer }}</div>
                <div style="background:#f9f9f9;padding:8px;border-radius:4px;font-size:13px;color:#666">
                    <strong>解析：</strong> {{ previewModal.q.analysis || '暂无' }}
                </div>
            </div>
        </div>
    </n-modal>
  </div>
</template>

<style scoped>
.feedback-page { padding: 16px; height: 100vh; background-color: #f5f7f9; }
.main-card { height: 100%; display: flex; flex-direction: column; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.title { margin: 0; font-size: 18px; display: flex; align-items: center; gap: 8px; }
.q-content { font-size: 15px; color: #333; line-height: 1.6; margin-bottom: 12px; }
.q-options { display: flex; flex-direction: column; gap: 6px; }
.opt { font-size: 14px; color: #555; }
.k { font-weight: bold; color: #2080f0; margin-right: 6px; }
</style>