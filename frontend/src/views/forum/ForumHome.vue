<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NGrid, NGi, NCard, NList, NListItem, NIcon, NSkeleton, NTag } from 'naive-ui'
import { Flame, ChatbubblesOutline, TrendingUpOutline } from '@vicons/ionicons5' 
import request from '../../utils/request'

const router = useRouter()
const loading = ref(true)
const boards = ref<any[]>([])
const hotPosts = ref<any[]>([])

const processUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
const getBoardIcon = (url: string) => {
  return url ? processUrl(url) : 'https://cdn-icons-png.flaticon.com/512/2659/2659360.png'
}

const fetchBoards = async () => {
  try {
    const res: any = await request.get('/forum/boards')
    if (res.data) boards.value = res.data
  } finally {
    loading.value = false
  }
}

const fetchHotPosts = async () => {
  const res: any = await request.get('/forum/posts', { params: { page: 1, page_size: 10 } })
  if (res.data) {
    hotPosts.value = res.data.sort((a: any, b: any) => b.view_count - a.view_count).slice(0, 8)
  }
}

const handleBoardClick = (boardId: number) => {
  router.push(`/forum/board/${boardId}`)
}

onMounted(() => {
  fetchBoards()
  fetchHotPosts()
})
</script>

<template>
  <div class="forum-container">
    
    <div class="forum-header">
       <div class="header-left">
          <h2 class="page-title">
             <n-icon color="#2080f0" style="margin-right: 8px"><ChatbubblesOutline /></n-icon>
             医学交流社区
          </h2>
          <p class="subtitle">临床经验分享 · 疑难病例讨论 · 学习资料互助</p>
       </div>
       <div class="header-right">
          </div>
    </div>

    <n-grid x-gap="20" y-gap="20" cols="1 m:3" responsive="screen">
        
        <n-gi span="2">
           <div class="section-label">全部板块</div>
           
           <div v-if="loading" class="board-grid">
              <n-card v-for="i in 4" :key="i" class="board-card"><n-skeleton text :repeat="3" /></n-card>
           </div>

           <div v-else class="board-grid">
              <div 
                 v-for="board in boards" 
                 :key="board.id" 
                 class="board-card" 
                 @click="handleBoardClick(board.id)"
              >
                 <div class="icon-wrapper">
                    <img :src="getBoardIcon(board.icon)" alt="icon" />
                 </div>
                 <div class="info-wrapper">
                    <div class="b-header">
                       <span class="b-name">{{ board.name }}</span>
                       <n-tag size="tiny" :bordered="false" type="primary" round class="count-tag">
                          {{ board.post_count || 0 }} 帖
                       </n-tag>
                    </div>
                    <div class="b-desc">{{ board.description || '暂无描述' }}</div>
                 </div>
              </div>
           </div>
        </n-gi>

        <n-gi span="1">
           <div class="section-label">
              <n-icon color="#d03050"><TrendingUpOutline /></n-icon> 全站热榜
           </div>
           
           <n-card size="small" :bordered="false" class="sidebar-card">
              <n-list hoverable clickable>
                 <n-list-item v-for="(post, index) in hotPosts" :key="post.id" @click="router.push(`/post/${post.id}`)">
                    <div class="hot-row">
                       <span class="rank-num" :class="{'top-rank': index < 3}">{{ index + 1 }}</span>
                       <span class="hot-title">{{ post.title }}</span>
                    </div>
                 </n-list-item>
              </n-list>
           </n-card>
        </n-gi>

    </n-grid>
  </div>
</template>

<style scoped>
.forum-container { max-width: 1200px; margin: 0 auto; padding: 24px; min-height: 100vh; }

/* 头部样式对齐 Market.vue */
.forum-header { 
  display: flex; justify-content: space-between; align-items: flex-end; 
  margin-bottom: 32px; padding-bottom: 16px; border-bottom: 1px solid #eee;
}
.page-title { margin: 0; font-size: 24px; font-weight: 700; color: #333; display: flex; align-items: center; }
.subtitle { margin: 4px 0 0 0; color: #666; font-size: 14px; }

.section-label { 
  font-size: 16px; font-weight: bold; color: #333; margin-bottom: 12px; 
  display: flex; align-items: center; gap: 6px;
}

/* 板块网格 */
.board-grid { 
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; 
}

/* 响应式：小屏单列 */
@media (max-width: 640px) {
  .board-grid { grid-template-columns: 1fr; } 
  .forum-container { padding: 16px; }
  .forum-header { flex-direction: column; align-items: flex-start; gap: 10px; }
}

/* 板块卡片 */
.board-card {
  background: #fff; border-radius: 12px; padding: 16px;
  display: flex; align-items: center; gap: 16px;
  border: 1px solid #f1f5f9; cursor: pointer; transition: all 0.2s;
}
.board-card:hover { 
  border-color: #2080f0; transform: translateY(-2px); 
  box-shadow: 0 4px 12px rgba(32, 128, 240, 0.08); 
}

.icon-wrapper { width: 56px; height: 56px; background: #f8fafc; border-radius: 12px; padding: 8px; flex-shrink: 0; }
.icon-wrapper img { width: 100%; height: 100%; object-fit: contain; }

.info-wrapper { flex: 1; overflow: hidden; }
.b-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.b-name { font-size: 16px; font-weight: 700; color: #1e293b; }
.b-desc { font-size: 13px; color: #94a3b8; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* 侧边栏 */
.sidebar-card { border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.02); }
.hot-row { display: flex; align-items: center; gap: 10px; }
.rank-num { 
  width: 20px; height: 20px; background: #f1f5f9; color: #94a3b8; 
  font-size: 12px; font-weight: bold; text-align: center; line-height: 20px; 
  border-radius: 4px; flex-shrink: 0; 
}
.top-rank { background: #fee2e2; color: #ef4444; }
.hot-title { font-size: 14px; color: #334155; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>