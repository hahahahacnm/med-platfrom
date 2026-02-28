<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NDataTable, NButton, NTag, NInput, NModal, useMessage, 
  NTabs, NTabPane, NSpace, NAvatar, NImage, NImageGroup, NDivider,
  NIcon, NRadioGroup, NRadio
} from 'naive-ui'
import { 
  ChatboxEllipsesOutline, TimeOutline, CheckmarkCircleOutline, 
  CloseCircleOutline, AlertCircleOutline, EyeOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()
const loading = ref(false)
const data = ref([])
const pagination = reactive({ page: 1, pageSize: 10, itemCount: 0 })
const activeStatus = ref('0') // '0':待办, '1':处理中, '2':已解决, '3':已驳回, 'all':全部

// 🔎 辅助函数：解析图片
const parseImages = (imgData: any) => {
    try {
        if (!imgData) return []
        if (Array.isArray(imgData)) return imgData
        return JSON.parse(imgData) || []
    } catch { return [] }
}

const getFullUrl = (path: string) => path.startsWith('http') ? path : `http://localhost:8080${path}`

// 🏷️ 状态字典
const statusMap: Record<number, { text: string, type: string }> = {
    0: { text: '待处理', type: 'default' },
    1: { text: '处理中', type: 'info' },
    2: { text: '已解决', type: 'success' },
    3: { text: '已驳回', type: 'error' }
}

// 📋 表格列定义
const columns = [
  { title: 'ID', key: 'id', width: 60, align: 'center' },
  { 
    title: '用户', key: 'user', width: 140,
    render(row: any) {
        const u = row.user || {}
        return h('div', { style: 'display:flex;align-items:center;gap:8px' }, [
            h(NAvatar, { round: true, size: 24, src: u.avatar ? getFullUrl(u.avatar) : undefined, fallbackSrc: "https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" }),
            h('div', { style: 'display:flex;flex-direction:column' }, [
                h('span', { style: 'font-size:12px;font-weight:bold' }, u.nickname || u.username),
                h('span', { style: 'font-size:11px;color:#999' }, `ID:${u.id}`)
            ])
        ])
    }
  },
  { 
    title: '类型', key: 'type', width: 100,
    render: (row: any) => h(NTag, { type: 'info', bordered: false, size: 'small' }, { default: () => row.type })
  },
  { 
    title: '内容摘要', key: 'content', width: 250, ellipsis: { tooltip: true }
  },
  {
    title: '图片', key: 'images', width: 80,
    render(row: any) {
        const imgs = parseImages(row.images)
        if (imgs.length === 0) return '-'
        return h(NTag, { size: 'small', round: true }, { default: () => `${imgs.length} 图` })
    }
  },
  { 
    title: '时间', key: 'created_at', width: 160, 
    render: (r: any) => new Date(r.created_at).toLocaleString() 
  },
  { 
    title: '状态', key: 'status', width: 100, fixed: 'right',
    render: (row: any) => {
        const s = statusMap[row.status] || { text: '未知', type: 'default' }
        return h(NTag, { type: s.type as any, bordered: false, size: 'small' }, { default: () => s.text })
    }
  },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 100,
    render(row: any) {
        return h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => openProcess(row) }, { default: () => '处理' })
    }
  }
]

// 📥 获取数据
const fetchData = async () => {
    loading.value = true
    try {
        const params: any = { page: pagination.page, page_size: pagination.pageSize }
        if (activeStatus.value !== 'all') params.status = activeStatus.value
        
        const res: any = await request.get('/admin/platform-feedbacks', { params })
        data.value = res.data || []
        pagination.itemCount = res.total
    } finally { loading.value = false }
}

// 🛠️ 处理弹窗逻辑
const modal = reactive({ show: false, loading: false, data: null as any })
const form = reactive({ status: 2, reply: '' })

const openProcess = (row: any) => {
    modal.data = row
    // 初始化表单
    form.status = row.status === 0 ? 2 : row.status // 默认选中"已解决"
    form.reply = row.admin_reply || ''
    modal.show = true
}

const submitProcess = async () => {
    modal.loading = true
    try {
        await request.put(`/admin/platform-feedbacks/${modal.data.id}`, {
            status: form.status,
            admin_reply: form.reply
        })
        message.success('处理完成')
        modal.show = false
        fetchData()
    } catch { message.error('操作失败') }
    finally { modal.loading = false }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-root">
    <n-card :bordered="false" class="main-card">
        <div class="header">
            <h2 class="title"><n-icon color="#2080f0"><ChatboxEllipsesOutline/></n-icon> 平台意见反馈</h2>
            <n-button size="small" @click="fetchData">刷新</n-button>
        </div>

        <n-tabs type="line" v-model:value="activeStatus" @update:value="()=>{pagination.page=1;fetchData()}">
            <n-tab-pane name="0" tab="⏳ 待处理" />
            <n-tab-pane name="1" tab="🏃 处理中" />
            <n-tab-pane name="2" tab="✅ 已解决" />
            <n-tab-pane name="3" tab="🚫 已驳回" />
            <n-tab-pane name="all" tab="全部记录" />
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

    <n-modal v-model:show="modal.show" preset="card" title="📝 反馈详情与处理" style="width: 600px; max-width: 95%">
        <div v-if="modal.data">
            <div class="user-info-box">
                <n-avatar round size="small" :src="modal.data.user?.avatar ? getFullUrl(modal.data.user.avatar) : undefined" />
                <span class="uname">{{ modal.data.user?.nickname || modal.data.user?.username }}</span>
                <n-divider vertical />
                <span class="ucontact" v-if="modal.data.contact">📞 联系方式: {{ modal.data.contact }}</span>
                <span class="ucontact" v-else>未留联系方式</span>
            </div>

            <div class="content-box">
                <n-tag type="info" size="small" style="margin-bottom: 8px">{{ modal.data.type }}</n-tag>
                <div class="text">{{ modal.data.content }}</div>
                
                <div class="img-box" v-if="parseImages(modal.data.images).length > 0">
                    <n-image-group>
                        <n-space>
                            <n-image 
                                v-for="(url, i) in parseImages(modal.data.images)" 
                                :key="i"
                                :src="getFullUrl(url)"
                                width="80" height="80" object-fit="cover"
                                style="border-radius: 4px; border: 1px solid #eee"
                            />
                        </n-space>
                    </n-image-group>
                </div>
            </div>

            <n-divider dashed />

            <div class="process-form">
                <div class="form-label">处理状态：</div>
                <n-radio-group v-model:value="form.status" name="radiogroup">
                    <n-space>
                        <n-radio :value="1">🏃 处理中</n-radio>
                        <n-radio :value="2">✅ 已解决</n-radio>
                        <n-radio :value="3">🚫 驳回</n-radio>
                    </n-space>
                </n-radio-group>

                <div class="form-label" style="margin-top: 16px;">回复用户：</div>
                <n-input 
                    v-model:value="form.reply" 
                    type="textarea" 
                    placeholder="请输入回复内容，用户将在反馈中心看到..." 
                    :autosize="{ minRows: 3, maxRows: 6 }"
                />
            </div>
        </div>
        
        <template #footer>
            <div style="display: flex; justify-content: flex-end; gap: 12px;">
                <n-button @click="modal.show=false">取消</n-button>
                <n-button type="primary" :loading="modal.loading" @click="submitProcess">确认提交</n-button>
            </div>
        </template>
    </n-modal>
  </div>
</template>

<style scoped>
.page-root { padding: 16px; height: 100vh; background-color: #f5f7f9; }
.main-card { height: 100%; display: flex; flex-direction: column; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.title { margin: 0; font-size: 18px; display: flex; align-items: center; gap: 8px; }

.user-info-box { display: flex; align-items: center; gap: 8px; background: #f9f9f9; padding: 8px 12px; border-radius: 8px; font-size: 13px; color: #666; }
.uname { font-weight: bold; color: #333; }

.content-box { margin-top: 16px; padding: 0 4px; }
.text { font-size: 14px; color: #333; line-height: 1.6; white-space: pre-wrap; }
.img-box { margin-top: 12px; }

.form-label { font-weight: bold; margin-bottom: 8px; color: #333; font-size: 14px; }
</style>