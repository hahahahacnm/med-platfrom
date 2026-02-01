<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NGrid, NGi, NCard, NList, NListItem, NIcon } from 'naive-ui'
// 🔥 修正：将 Fire 改为 Flame
import { Flame, EyeOutline } from '@vicons/ionicons5' 
import request from '../../utils/request'

const router = useRouter()

// 状态
const boards = ref<any[]>([])
const hotPosts = ref<any[]>([])

// 辅助函数
const processUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `http://localhost:8080${url}`
}
const getBoardIcon = (url: string) => {
  return url ? processUrl(url) : 'https://cdn-icons-png.flaticon.com/512/2659/2659360.png'
}

// 1. 获取板块
const fetchBoards = async () => {
  const res: any = await request.get('/forum/boards')
  if (res.data) boards.value = res.data
}

// 2. 获取热门帖子 (侧边栏用)
const fetchHotPosts = async () => {
  const res: any = await request.get('/forum/posts', { params: { page: 1, page_size: 10 } })
  if (res.data) {
    hotPosts.value = res.data.sort((a: any, b: any) => b.view_count - a.view_count).slice(0, 8)
  }
}

// 🔥 核心交互：点击跳转到详情页
const handleBoardClick = (boardId: number) => {
  router.push(`/forum/board/${boardId}`)
}

onMounted(() => {
  fetchBoards()
  fetchHotPosts()
})
</script>

<template>
  <div class="forum-home">
    
    <div class="banner-area">
       <div class="banner-content">
          <h1>医学交流社区</h1>
          <p>分享临床经验，探讨前沿知识，共同进步</p>
       </div>
    </div>

    <div class="main-content-wrapper">
      <n-grid x-gap="24" cols="1 900:3" item-responsive>
        
        <n-gi span="2">
          <div class="section-title-bar">
             <span class="title">全部板块</span>
             <span class="subtitle">点击图标进入专属讨论区</span>
          </div>

          <div class="board-grid">
             <div 
                v-for="board in boards" 
                :key="board.id" 
                class="board-card" 
                @click="handleBoardClick(board.id)"
             >
                <div class="icon-box">
                   <img :src="getBoardIcon(board.icon)" alt="icon" />
                </div>
                <div class="info-box">
                   <div class="b-name">{{ board.name }}</div>
                   <div class="b-desc" :title="board.description">{{ board.description || '暂无描述' }}</div>
                   <div class="b-stats">
                      <span>帖子: {{ board.post_count || 0 }}</span>
                      <span class="divider">|</span>
                      <span>评论: {{ board.comment_count || 0 }}</span>
                   </div>
                </div>
             </div>
          </div>
        </n-gi>

        <n-gi span="1">
           <n-card :bordered="false" class="sidebar-widget">
              <template #header>
                 <div class="widget-header">
                    <n-icon color="#d03050" size="20"><Flame /></n-icon>
                    <span>全站热点</span>
                 </div>
              </template>
              <n-list hoverable clickable>
                 <n-list-item v-for="(post, index) in hotPosts" :key="post.id" @click="router.push(`/post/${post.id}`)">
                    <div class="hot-item">
                       <div class="rank-num" :class="{'top': index < 3}">{{ index + 1 }}</div>
                       <div class="hi-content">
                          <div class="hi-title">{{ post.title }}</div>
                          <div class="hi-meta">
                             <n-icon><EyeOutline /></n-icon> {{ post.view_count }}
                          </div>
                       </div>
                    </div>
                 </n-list-item>
              </n-list>
           </n-card>

           <div class="ad-widget">
              <h4>🏥 官方交流群</h4>
              <p>加入微信群，获取最新资讯。</p>
              <div style="background:#fff; height:100px; line-height:100px; color:#ddd; font-size:12px;">二维码占位</div>
           </div>
        </n-gi>

      </n-grid>
    </div>

  </div>
</template>

<style scoped>
.forum-home { background-color: #f7f9fa; min-height: 100vh; padding-bottom: 60px; }

/* Banner */
.banner-area {
  background: linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%);
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: -50px; 
  box-shadow: 0 4px 15px rgba(0,0,0,0.05);
}
.banner-content { text-align: center; color: #fff; text-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.banner-content h1 { font-size: 36px; margin: 0; font-weight: 700; letter-spacing: 2px; }
.banner-content p { font-size: 16px; margin-top: 10px; opacity: 0.95; }

.main-content-wrapper { max-width: 1200px; margin: 0 auto; padding: 0 20px; position: relative; z-index: 10; }

.section-title-bar { margin-bottom: 20px; display: flex; align-items: baseline; gap: 12px; }
.section-title-bar .title { font-size: 22px; font-weight: bold; color: #333; }
.section-title-bar .subtitle { font-size: 14px; color: #999; }

/* Grid */
.board-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin-bottom: 30px; }
@media (max-width: 600px) { .board-grid { grid-template-columns: 1fr; } }

.board-card {
  background: #fff; border-radius: 12px; padding: 24px;
  display: flex; align-items: center; gap: 20px;
  cursor: pointer; transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(0,0,0,0.03);
  border: 1px solid #f0f0f0;
}
.board-card:hover { transform: translateY(-5px); box-shadow: 0 10px 25px rgba(0,0,0,0.08); border-color: #d1c4e9; }

.icon-box { width: 64px; height: 64px; flex-shrink: 0; }
.icon-box img { width: 100%; height: 100%; object-fit: contain; }

.info-box { flex: 1; overflow: hidden; }
.b-name { font-size: 18px; font-weight: bold; color: #333; margin-bottom: 6px; }
.b-desc { font-size: 13px; color: #888; margin-bottom: 10px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.b-stats { font-size: 12px; color: #aaa; }
.b-stats .divider { margin: 0 8px; color: #eee; }

/* Sidebar */
.sidebar-widget { border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.03); margin-bottom: 24px; }
.widget-header { display: flex; align-items: center; gap: 8px; font-weight: bold; font-size: 16px; color: #333; }

.hot-item { display: flex; gap: 12px; align-items: flex-start; padding: 6px 0; }
.rank-num { width: 20px; height: 20px; background: #eee; color: #999; font-weight: bold; font-size: 12px; text-align: center; line-height: 20px; border-radius: 4px; flex-shrink: 0; }
.rank-num.top { background: #ff7875; color: #fff; }
.hi-content { flex: 1; overflow: hidden; }
/* 🔥 修正：添加 line-clamp 标准属性以消除警告 */
.hi-title { 
  font-size: 14px; color: #333; margin-bottom: 4px; 
  display: -webkit-box; 
  -webkit-line-clamp: 2; 
  line-clamp: 2; /* 兼容性写法 */
  -webkit-box-orient: vertical; 
  overflow: hidden; 
  line-height: 1.5; 
}
.hi-meta { font-size: 12px; color: #aaa; display: flex; align-items: center; gap: 4px; }

.ad-widget { background: linear-gradient(135deg, #fff 0%, #f0f7ff 100%); border-radius: 12px; padding: 24px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); text-align: center; border: 1px solid #e6f7ff; }
.ad-widget h4 { margin: 0 0 8px 0; color: #1890ff; font-size: 16px; }
.ad-widget p { font-size: 13px; color: #666; margin: 0 0 16px 0; }
</style>