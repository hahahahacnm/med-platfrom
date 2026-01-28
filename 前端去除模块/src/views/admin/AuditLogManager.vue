<script setup lang="ts">
import { ref, onMounted, reactive, h } from 'vue'
import { 
  NCard, NDataTable, NTag, NPageHeader, NIcon, NInput, NButton, NSpace, NDatePicker 
} from 'naive-ui'
import { SearchOutline, ShieldCheckmarkOutline, AlertCircleOutline } from '@vicons/ionicons5'
import request from '../../utils/request'
import { format } from 'date-fns'

const loading = ref(false)
const list = ref([])
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

// 筛选条件 (目前支持按ID查)
const filter = reactive({
  operatorId: '',
  targetId: ''
})

const columns = [
  { title: '日志ID', key: 'ID', width: 80 },
  { 
    title: '操作时间', key: 'CreatedAt', width: 180,
    render(row: any) {
      return format(new Date(row.CreatedAt), 'yyyy-MM-dd HH:mm:ss')
    }
  },
  { 
    title: '操作员 (管理员/代理)', key: 'operator_name', width: 150,
    render(row: any) {
      return h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.operator_name || `ID:${row.operator_id}` })
    }
  },
  { 
    title: '动作类型', key: 'action', width: 100, align: 'center',
    render(row: any) {
      const isGrant = row.action === 'GRANT'
      return h(NTag, 
        { type: isGrant ? 'success' : 'error' }, 
        { default: () => isGrant ? '发放授权' : '强制收回' }
      )
    }
  },
  // 🔥🔥🔥 [修改点]：显示详细的客户信息 (用户名 + ID) 🔥🔥🔥
  { 
    title: '目标客户', key: 'target_info', width: 160,
    render(row: any) {
      return h('div', [
        // 第一行：显示用户名 (绿色高亮)
        h('div', { style: 'font-weight: bold; color: #18a058' }, row.target_user_name || '未知用户'),
        // 第二行：显示ID (灰色小字)
        h('div', { style: 'font-size: 12px; color: #999' }, `ID: ${row.target_user_id}`)
      ])
    }
  },
  { 
    title: '涉及商品 (快照)', key: 'product_name', 
    render(row: any) {
      return h('span', { style: 'font-weight: bold; color: #333' }, row.product_name)
    }
  },
  { 
    title: '授权详情', key: 'details',
    render(row: any) {
      if (row.action === 'REVOKE') {
        return h('span', { style: 'color: #ccc' }, '---')
      }
      return h('span', null, [
        `时长: ${row.duration_days}天`,
        h('br'),
        h('span', { style: 'color: #999; font-size: 12px' }, `至: ${format(new Date(row.expire_at), 'yyyy-MM-dd')}`)
      ])
    }
  }
]

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/admin/auth-logs', {
      params: { 
        page: pagination.page, 
        page_size: pagination.pageSize,
        operator_id: filter.operatorId,
        target_id: filter.targetId
      }
    })
    list.value = res.data || []
    pagination.itemCount = res.total || 0
  } finally { loading.value = false }
}

const handleSearch = () => { pagination.page = 1; fetchData() }
const handlePageChange = (page: number) => { pagination.page = page; fetchData() }

onMounted(fetchData)
</script>

<template>
  <div class="page-container">
    <n-page-header title="🛡️ 授权审计日志" subtitle="所有权限变更操作均在此留痕，支持追溯责任人" style="margin-bottom: 20px;">
      <template #avatar><n-icon size="24"><ShieldCheckmarkOutline /></n-icon></template>
    </n-page-header>
    
    <n-card>
      <div class="toolbar">
        <n-input v-model:value="filter.operatorId" placeholder="搜索操作员ID..." style="width: 200px" clearable />
        <n-input v-model:value="filter.targetId" placeholder="搜索客户ID..." style="width: 200px" clearable />
        <n-button type="primary" @click="handleSearch">
          <template #icon><n-icon><SearchOutline /></n-icon></template> 
          查询日志
        </n-button>
      </div>
      
      <n-data-table 
        remote 
        striped
        :columns="columns" 
        :data="list" 
        :loading="loading" 
        :pagination="pagination" 
        @update:page="handlePageChange" 
        style="margin-top: 16px;" 
      />
    </n-card>
  </div>
</template>

<style scoped>
.page-container { padding: 24px; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>