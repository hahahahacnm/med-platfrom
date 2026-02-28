<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NDataTable, NTag, NButton, NInputNumber, NModal, 
  NForm, NFormItem, useMessage, NIcon, NPageHeader, NAlert, NButtonGroup
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui' // 🔥 引入 Naive UI 的表格列类型
import { 
  AddCircleOutline, RefreshOutline, CopyOutline, DownloadOutline // 🔥 新增 DownloadOutline
} from '@vicons/ionicons5'
import request from '../../utils/request'

const message = useMessage()
const loading = ref(false)
const list = ref([])
const pagination = reactive({ page: 1, pageSize: 15, itemCount: 0 })

// 筛选状态
const filterStatus = ref<string | null>(null) // null=全部, '0'=未使用, '1'=已使用

// 生成卡密弹窗
const showGenerateModal = ref(false)
const generateForm = reactive({ count: 10, points: 500 })
const generating = ref(false)

// 🔥 新增：导出卡密弹窗状态
const showExportModal = ref(false)
const exportPoints = ref(500)
const exporting = ref(false)

// === 表格列定义 ===
const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'ID', width: 60 },
  { 
    title: '激活码 (卡密)', key: 'code', width: 200,
    render(row: any) {
        return h('span', { style: 'font-family: monospace; font-size: 15px; font-weight: bold; color: #334155; letter-spacing: 1px;' }, row.code)
    }
  },
  { 
    title: '额度 (积分)', key: 'points', width: 120,
    render(row: any) {
        return h('span', { style: 'color: #d97706; font-weight: bold;' }, `+ ${row.points} 分`)
    }
  },
  { 
    title: '状态', key: 'status', width: 100,
    render(row: any) {
      if (row.status === 1) return h(NTag, { type: 'error', size: 'small', bordered: false }, { default: () => '已使用' })
      return h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '未使用' })
    }
  },
  { 
    title: '使用者ID', key: 'used_by_id', width: 100,
    render(row: any) {
        if (row.status === 0) return '-'
        return h(NTag, { type: 'default', size: 'small' }, { default: () => `用户 ${row.used_by_id}` })
    }
  },
  { 
    title: '使用时间', key: 'used_at', width: 180,
    render(row: any) {
        if (row.status === 0 || !row.used_at) return '-'
        return new Date(row.used_at).toLocaleString()
    }
  },
  { 
    title: '生成时间', key: 'CreatedAt', width: 180,
    render: (row: any) => new Date(row.CreatedAt).toLocaleString() 
  },
  {
    title: '操作', key: 'actions', fixed: 'right', width: 100,
    render(row: any) {
      return h(NButton, { 
          size: 'tiny', 
          type: 'primary', 
          secondary: true,
          disabled: row.status === 1,
          onClick: () => copyCode(row.code) 
        }, 
        { icon: () => h(NIcon, null, { default: () => h(CopyOutline) }), default: () => '复制' }
      )
    }
  }
]

const copyCode = (code: string) => {
    navigator.clipboard.writeText(code).then(() => {
        message.success('卡密已复制到剪贴板')
    }).catch(() => message.error('复制失败'))
}

// === API 操作 ===
const fetchData = async () => {
  loading.value = true
  try {
    const params: any = { page: pagination.page, page_size: pagination.pageSize }
    if (filterStatus.value !== null) params.status = filterStatus.value

    const res: any = await request.get('/admin/codes', { params })
    list.value = res.data || []
    pagination.itemCount = res.total || 0
  } catch (e) {
    message.error('加载列表失败')
  } finally {
    loading.value = false
  }
}

const handleFilter = (status: string | null) => {
    filterStatus.value = status
    pagination.page = 1
    fetchData()
}

const handlePageChange = (page: number) => { 
    pagination.page = page
    fetchData() 
}

const submitGenerate = async () => {
    generating.value = true
    try {
        const res: any = await request.post('/admin/codes/generate', generateForm)
        message.success(res.message || '批量生成成功')
        showGenerateModal.value = false
        handleFilter(null) // 刷新并回到全部列表
    } catch (e: any) {
        message.error(e.response?.data?.error || '生成失败')
    } finally {
        generating.value = false
    }
}

// 🔥 新增：执行卡密导出下载逻辑
const submitExport = async () => {
    if (!exportPoints.value) return message.warning('请输入要导出的积分额度')
    exporting.value = true
    try {
        // 请求后端接口，要求返回 Blob (二进制流)
        const res: any = await request.get('/admin/codes/export', {
            params: { points: exportPoints.value },
            responseType: 'blob' 
        })
        
        // 拦截后端返回的 JSON 报错 (如果出错，后端返回的是 JSON 格式的 Blob)
        if (res.type === 'application/json') {
            const text = await res.text()
            const err = JSON.parse(text)
            message.error(err.error || '导出失败')
            return
        }

        // 正常下载流程：创建隐形 a 标签触发浏览器下载
        const url = window.URL.createObjectURL(new Blob([res]))
        const a = document.createElement('a')
        a.href = url
        a.download = `卡密-${exportPoints.value}积分.txt`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        window.URL.revokeObjectURL(url)

        message.success('导出成功')
        showExportModal.value = false
    } catch (e: any) {
        message.error('导出失败，请检查该额度下是否有可用卡密')
    } finally {
        exporting.value = false
    }
}

onMounted(fetchData)
</script>

<template>
  <div class="code-manage-container">
    <n-page-header title="🎟️ 激活码与卡密管理" subtitle="生成、分发、追踪卡密兑换状态" style="margin-bottom: 24px;" />
    
    <n-card>
      <div class="toolbar">
        <div class="toolbar-left">
            <n-button-group>
               <n-button :type="filterStatus === null ? 'primary' : 'default'" @click="handleFilter(null)">全部</n-button>
               <n-button :type="filterStatus === '0' ? 'success' : 'default'" @click="handleFilter('0')">未使用</n-button>
               <n-button :type="filterStatus === '1' ? 'error' : 'default'" @click="handleFilter('1')">已核销</n-button>
            </n-button-group>
            <n-button circle @click="fetchData" style="margin-left: 12px;"><template #icon><n-icon><RefreshOutline/></n-icon></template></n-button>
        </div>
        
        <div class="toolbar-right" style="display: flex; gap: 12px;">
            <n-button type="info" @click="showExportModal = true">
                <template #icon><n-icon><DownloadOutline /></n-icon></template>
                导出可用卡密 (TXT)
            </n-button>

            <n-button type="primary" color="#18a058" @click="showGenerateModal = true">
                <template #icon><n-icon><AddCircleOutline /></n-icon></template>
                批量生成新卡密
            </n-button>
        </div>
      </div>

      <n-data-table 
        remote 
        :columns="columns" 
        :data="list" 
        :loading="loading" 
        :pagination="pagination" 
        @update:page="handlePageChange" 
        style="margin-top: 16px;" 
        :scroll-x="1000" 
      />
    </n-card>

    <n-modal v-model:show="showGenerateModal" preset="card" title="⚡ 批量生成卡密" style="width: 450px">
        <n-alert type="info" :show-icon="false" style="margin-bottom: 20px;">
            系统将自动生成 12 位高强度防伪卡密。生成后可直接发给用户，用户在前端输入即可兑换对应积分。
        </n-alert>
        
        <n-form label-placement="top">
            <n-grid cols="2" x-gap="20">
                <n-gi>
                    <n-form-item label="生成数量 (个)">
                        <n-input-number v-model:value="generateForm.count" :min="1" :max="500" style="width: 100%" />
                    </n-form-item>
                </n-gi>
                <n-gi>
                    <n-form-item label="包含积分额度">
                        <n-input-number v-model:value="generateForm.points" :min="10" :step="100" style="width: 100%">
                            <template #suffix>分</template>
                        </n-input-number>
                    </n-form-item>
                </n-gi>
            </n-grid>
        </n-form>
        
        <template #footer>
            <div style="display:flex; justify-content:flex-end">
                <n-button @click="showGenerateModal=false" style="margin-right:12px">取消</n-button>
                <n-button type="primary" :loading="generating" @click="submitGenerate">确认生成</n-button>
            </div>
        </template>
    </n-modal>

    <n-modal v-model:show="showExportModal" preset="card" title="📥 按额度导出卡密" style="width: 400px">
        <n-alert type="success" :show-icon="false" style="margin-bottom: 20px;">
            系统将提取指定额度下所有<strong style="color:red"> 未使用 </strong>的卡密，并保存为 TXT 文本，每行一个。
        </n-alert>
        
        <n-form label-placement="top">
            <n-form-item label="要导出的积分额度">
                <n-input-number v-model:value="exportPoints" :min="10" :step="10" style="width: 100%" size="large">
                    <template #suffix>分</template>
                </n-input-number>
            </n-form-item>
        </n-form>
        
        <template #footer>
            <div style="display:flex; justify-content:flex-end">
                <n-button @click="showExportModal=false" style="margin-right:12px">取消</n-button>
                <n-button type="primary" :loading="exporting" @click="submitExport">立即下载</n-button>
            </div>
        </template>
    </n-modal>
  </div>
</template>

<style scoped>
.code-manage-container { padding: 24px; min-height: 100vh; background-color: #f5f7fa; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.toolbar-left { display: flex; align-items: center; }
</style>