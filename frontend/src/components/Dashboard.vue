<script setup lang="ts">
import { ref, onMounted, onUnmounted, shallowRef, computed } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NCard, NIcon, NSkeleton, NAvatar, NTooltip, NEmpty,
  NNumberAnimation, NButton, NModal, NSpin, useMessage, NTag, NGradientText,
  useNotification 
} from 'naive-ui'
import { 
  ArrowForwardOutline, BookOutline, StarOutline, JournalOutline,
  Flame, BarChartOutline, TrophyOutline, 
  NotificationsOutline, CheckmarkDoneOutline, SparklesOutline, CalendarOutline,
  ChevronForwardOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const router = useRouter()
const message = useMessage()
const notification = useNotification()

const loading = ref(true)

// 核心数据模型
const stats = ref({ 
    total_count: 0, today_count: 0, accuracy: 0,
    consecutive_days: 0, activity_map: [] as any[], rank_list: [] as any[]
})

const notifications = ref<any[]>([])
const unreadCount = ref(0)
const notifLoading = ref(false)

// 排行榜弹窗逻辑
const showRankModal = ref(false)
const rankListFull = ref<any[]>([])
const rankLoading = ref(false)
const rankPage = ref(1)
const rankHasMore = ref(true)

// WebSocket 实例
let ws: WebSocket | null = null

const fetchStats = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/stats') 
    if (res.data) stats.value = { ...res.data }
  } catch (e) { console.error(e) } finally { loading.value = false }
}

const fetchNotifications = async () => {
    notifLoading.value = true
    try {
        const res: any = await request.get('/notifications')
        notifications.value = res.data || []
        unreadCount.value = res.unread_count || 0
    } catch (e) { console.error(e) } finally { notifLoading.value = false }
}

// 🔥 全部标记为已读 (已修复路径，移除 /forum)
const handleReadAll = async () => {
    if (unreadCount.value === 0) return
    try {
        await request.put('/notifications/read-all')
        // 前端状态同步
        notifications.value.forEach(n => n.is_read = true)
        unreadCount.value = 0
        message.success('所有消息已标记为已读')
    } catch (e) { message.error('操作失败') }
}

// 🔥 单条消息点击事件 (已修复路径，状态即时同步)
const handleNotifClick = async (item: any) => {
    const targetPath = item.source_type === 'forum' ? `/post/${item.source_id}` : '/quiz'
    
    // 如果已读，直接跳转即可
    if (item.is_read) {
        router.push(targetPath)
        return
    }

    try {
        // 调用后端已读接口
        await request.put(`/notifications/${item.id}/read`)
        
        // 状态即时闭环，让红点和数字立刻消失
        item.is_read = true 
        if (unreadCount.value > 0) unreadCount.value--
        
        router.push(targetPath)
    } catch (error) {
        // 即使接口调用失败，为了防止死锁，也允许用户跳转
        router.push(targetPath)
    }
}

// WebSocket 实时通知接收逻辑
const initWebSocket = () => {
    const uid = userStore.id 
    if (!uid) return
    const wsUrl = `ws://localhost:8080/ws?uid=${uid}`
    ws = new WebSocket(wsUrl)
    
    ws.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data)
            if (msg.type === 'new_notification') {
                const data = msg.data
                
                // 右上角系统通知弹窗
                notification.info({
                    title: `💬 新动态: ${data.title}`,
                    content: data.content,
                    duration: 5000
                })
                
                // 推入本地通知列表顶部
                const newNotif = { 
                    ...data, 
                    is_read: false, 
                    sender: data.sender || { avatar: null, nickname: '系统' } 
                }
                notifications.value.unshift(newNotif)
                if (notifications.value.length > 15) notifications.value.pop() 
                unreadCount.value++
            }
        } catch (e) {}
    }
    
    ws.onclose = () => setTimeout(initWebSocket, 5000)
}

// 排行榜相关逻辑
const openRankModal = () => {
    showRankModal.value = true
    rankPage.value = 1
    rankListFull.value = []
    rankHasMore.value = true
    fetchFullRank()
}

const fetchFullRank = async () => {
    if (rankLoading.value || !rankHasMore.value) return
    rankLoading.value = true
    try {
        const res: any = await request.get('/rank/daily', { params: { page: rankPage.value, page_size: 20 } })
        if (res.data) {
            rankListFull.value = rankPage.value === 1 ? res.data : [...rankListFull.value, ...res.data]
            rankHasMore.value = res.has_more
            if (rankHasMore.value) rankPage.value++
        }
    } catch (e) { console.error(e) } finally { rankLoading.value = false }
}

const handleRankScroll = (e: Event) => {
    const target = e.target as HTMLElement
    if (target.scrollHeight - target.scrollTop - target.clientHeight < 50) fetchFullRank()
}

const getAvatar = (path: string | undefined) => {
    if (!path) return undefined
    return path.startsWith('http') ? path : `http://localhost:8080${path}`
}

onMounted(() => { 
    fetchStats()
    fetchNotifications()
    initWebSocket()
})

onUnmounted(() => { if (ws) ws.close() })
</script>

<template>
  <div class="db-container">
    <div class="welcome-hero animate-in">
        <div class="hero-left">
            <div class="avatar-ring">
                <n-avatar round :size="84" :src="getAvatar(userStore.avatar)" fallback-src="https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png" />
            </div>
            <div class="welcome-text">
                <h1 class="welcome-title">早安，{{ userStore.nickname || userStore.username }}</h1>
                <p class="welcome-subtitle">今天也要保持专注，离梦想更近一步。</p>
                <div class="badge-row">
                    <div class="stat-pill">
                        <n-icon color="#f59e0b"><Flame /></n-icon>
                        <span>专注打卡 <b>{{ stats.consecutive_days }}</b> 天</span>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="hero-stats">
             <div class="hero-stat-item">
                <span class="label">今日进度</span>
                <span class="value"><n-number-animation :to="stats.today_count" /> <small>题</small></span>
             </div>
             <div class="stat-v-line"></div>
             <div class="hero-stat-item">
                <span class="label">当前正确率</span>
                <span class="value">{{ stats.accuracy.toFixed(0) }}<small>%</small></span>
             </div>
        </div>
    </div>

    <div class="dashboard-grid">
        <div class="main-column">
            <div class="action-grid animate-in" style="animation-delay: 0.1s">
                <div class="action-tile t-blue" @click="router.push('/quiz')">
                    <div class="icon-wrap"><n-icon><ArrowForwardOutline /></n-icon></div>
                    <div class="tile-content"><h3>科学刷题</h3><p>系统化章节练习</p></div>
                    <n-icon class="arrow-hint"><ChevronForwardOutline /></n-icon>
                </div>
                <div class="action-tile t-red" @click="router.push('/mistakes')">
                    <div class="icon-wrap"><n-icon><BookOutline /></n-icon></div>
                    <div class="tile-content"><h3>错题清零</h3><p>攻克薄弱环节</p></div>
                    <n-icon class="arrow-hint"><ChevronForwardOutline /></n-icon>
                </div>
                <div class="action-tile t-amber" @click="router.push('/favorites')">
                    <div class="icon-wrap"><n-icon><StarOutline /></n-icon></div>
                    <div class="tile-content"><h3>考题收藏</h3><p>核心考点复盘</p></div>
                    <n-icon class="arrow-hint"><ChevronForwardOutline /></n-icon>
                </div>
                <div class="action-tile t-indigo" @click="router.push('/my-notes')">
                    <div class="icon-wrap"><n-icon><JournalOutline /></n-icon></div>
                    <div class="tile-content"><h3>复习笔记</h3><p>思维精华沉淀</p></div>
                    <n-icon class="arrow-hint"><ChevronForwardOutline /></n-icon>
                </div>
            </div>

            <n-card class="data-card animate-in" style="animation-delay: 0.2s">
                <template #header>
                    <div class="card-header-with-icon">
                        <n-icon color="#2563eb"><CalendarOutline /></n-icon>
                        <span>学习活力追踪 (14天)</span>
                    </div>
                </template>
                <div class="heatmap-container" v-if="!loading">
                    <div class="heatmap-flex">
                        <n-tooltip v-for="(day, i) in stats.activity_map" :key="i" trigger="hover">
                            <template #trigger>
                                <div class="heat-col">
                                    <div class="heat-track">
                                        <div class="heat-fill" :class="`lvl-${day.level}`" :style="{height: (day.level*20 + 15)+'%'}"></div>
                                    </div>
                                    <span class="heat-label">{{ day.date.split('-')[2] }}</span>
                                </div>
                            </template>
                            {{ day.date }}：完成 {{ day.count }} 题
                        </n-tooltip>
                    </div>
                </div>
                <n-empty v-else description="暂无数据" />
            </n-card>
        </div>

        <div class="side-column">
            <n-card class="data-card side-widget animate-in" style="animation-delay: 0.3s" content-style="padding: 0;">
                <template #header>
                    <div class="card-header-with-icon" style="padding: 16px 20px 0 20px">
                        <n-icon color="#f59e0b"><NotificationsOutline /></n-icon>
                        <span>动态通知</span>
                        <n-tag v-if="unreadCount > 0" type="error" size="tiny" round :bordered="false" class="badge-count">{{ unreadCount }}</n-tag>
                    </div>
                </template>
                <template #header-extra>
                   <n-tooltip trigger="hover">
                      <template #trigger>
                        <n-button text @click="handleReadAll" style="font-size: 18px; margin: 16px 20px 0 0">
                           <n-icon><CheckmarkDoneOutline /></n-icon>
                        </n-button>
                      </template>
                      全部标记为已读
                   </n-tooltip>
                </template>

                <div class="notif-scroll-area">
                    <div v-if="notifications.length > 0" class="notif-list">
                        <div 
                          v-for="n in notifications" :key="n.id" 
                          class="notif-card" 
                          :class="{ 'unread': !n.is_read }"
                          @click="handleNotifClick(n)"
                        >
                            <div v-if="!n.is_read" class="unread-dot"></div>
                            
                            <n-avatar round size="small" :src="getAvatar(n.sender?.avatar)" fallback-src="https://img.icons8.com/clouds/100/000000/user.png" />
                            <div class="notif-content">
                                <div class="notif-title-row">
                                   <span class="notif-user">{{ n.sender?.nickname || n.sender?.username || '系统通知' }}</span>
                                   <span class="notif-time">{{ new Date(n.created_at).toLocaleDateString() }}</span>
                                </div>
                                <p class="notif-text">{{ n.content }}</p>
                            </div>
                        </div>
                    </div>
                    <n-empty v-else description="暂无新动态" size="small" style="padding: 40px 0;" />
                </div>
            </n-card>

            <n-card class="data-card side-widget animate-in" style="animation-delay: 0.4s">
                <template #header>
                    <div class="card-header-with-icon">
                        <n-icon color="#d03050"><TrophyOutline /></n-icon>
                        <span>今日卷王榜</span>
                    </div>
                </template>
                <template #header-extra>
                    <n-button text type="primary" size="tiny" @click="openRankModal">全部 ></n-button>
                </template>
                <div class="compact-rank">
                    <div v-for="(u, i) in stats.rank_list" :key="i" class="mini-rank-row">
                        <span class="rank-idx" :class="{'top-3': i < 3}">{{ i+1 }}</span>
                        <n-avatar round :size="24" :src="getAvatar(u.avatar)" />
                        <span class="rank-name">{{ u.nickname || u.username }}</span>
                        <span class="rank-val">{{ u.count }} <small>题</small></span>
                    </div>
                    <n-empty v-if="stats.rank_list.length === 0" description="虚位以待" size="small" />
                </div>
            </n-card>
        </div>
    </div>

    <n-modal v-model:show="showRankModal" preset="card" title="今日卷王总榜 🏆" style="width: 500px; border-radius: 20px;" :segmented="{ content: true }">
        <div class="full-rank-list" @scroll="handleRankScroll">
            <div v-for="user in rankListFull" :key="user.user_id" class="full-rank-row">
                <div class="fr-idx">#{{ user.rank }}</div>
                <n-avatar round :size="40" :src="getAvatar(user.avatar)" />
                <div class="fr-info">
                    <span class="fr-name">{{ user.nickname || user.username }}</span>
                    <span class="fr-school" v-if="user.school">{{ user.school }}</span>
                </div>
                <div class="fr-score">今日 <strong>{{ user.count }}</strong> 题</div>
            </div>
            <div class="modal-footer">
                <n-spin v-if="rankLoading" size="small" />
                <span v-else-if="!rankHasMore" class="end-msg">仅展示前100名活跃学霸</span>
            </div>
        </div>
    </n-modal>
  </div>
</template>

<style scoped>
.db-container { max-width: 1200px; margin: 0 auto; padding: 32px 24px; box-sizing: border-box; }

@keyframes fadeInUp {
    from { opacity: 0; transform: translateY(24px); }
    to { opacity: 1; transform: translateY(0); }
}
.animate-in { animation: fadeInUp 0.7s cubic-bezier(0.16, 1, 0.3, 1) backwards; }

/* 欢迎横幅 */
.welcome-hero { 
    background: #fff; border: 1px solid #e2e8f0; border-radius: 28px; padding: 40px;
    display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px;
    box-shadow: 0 4px 20px -6px rgba(0,0,0,0.02);
}
.hero-left { display: flex; align-items: center; gap: 32px; }
.avatar-ring { border: 4px solid #f1f5f9; border-radius: 50%; padding: 2px; }
.welcome-title { margin: 0; font-size: 32px; font-weight: 900; color: #0f172a; letter-spacing: -0.04em; }
.welcome-subtitle { margin: 8px 0 16px 0; color: #64748b; font-size: 16px; }
.stat-pill { background: #fffbeb; border: 1px solid #fef3c7; color: #d97706; padding: 6px 16px; border-radius: 100px; display: flex; align-items: center; gap: 6px; font-weight: 800; font-size: 14px; }

.hero-stats { display: flex; align-items: center; gap: 48px; }
.hero-stat-item { display: flex; flex-direction: column; align-items: flex-end; }
.hero-stat-item .label { font-size: 12px; font-weight: 800; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.1em; }
.hero-stat-item .value { font-size: 32px; font-weight: 900; color: #1e293b; font-family: 'JetBrains Mono', monospace; }
.hero-stat-item .value small { font-size: 14px; margin-left: 2px; opacity: 0.4; }
.stat-v-line { width: 1px; height: 40px; background: #e2e8f0; }

.dashboard-grid { display: grid; grid-template-columns: 1fr 340px; gap: 32px; }
.main-column, .side-column { display: flex; flex-direction: column; gap: 32px; }

/* 快捷入口区块 */
.action-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.action-tile { 
    background: #fff; border: 1px solid #f1f5f9; border-radius: 24px; padding: 28px;
    cursor: pointer; display: flex; align-items: center; gap: 20px; position: relative;
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.action-tile:hover { transform: translateY(-8px); box-shadow: 0 12px 30px -10px rgba(0,0,0,0.06); }
.icon-wrap { width: 56px; height: 56px; border-radius: 18px; display: flex; align-items: center; justify-content: center; font-size: 26px; color: #fff; }
.tile-content h3 { margin: 0; font-size: 18px; font-weight: 800; color: #1e293b; }
.tile-content p { margin: 4px 0 0 0; font-size: 13px; color: #94a3b8; }
.arrow-hint { position: absolute; right: 24px; opacity: 0; transition: 0.3s; color: #cbd5e1; }
.action-tile:hover .arrow-hint { opacity: 1; transform: translateX(4px); }

.t-blue .icon-wrap { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.t-red .icon-wrap { background: linear-gradient(135deg, #ef4444, #dc2626); }
.t-amber .icon-wrap { background: linear-gradient(135deg, #f59e0b, #d97706); }
.t-indigo .icon-wrap { background: linear-gradient(135deg, #6366f1, #4f46e5); }

/* 通用卡片样式 */
.data-card { border-radius: 24px; border: 1px solid #f1f5f9; box-shadow: 0 1px 3px rgba(0,0,0,0.02); }
.card-header-with-icon { display: flex; align-items: center; gap: 10px; font-weight: 800; font-size: 16px; color: #334155; }

/* 热力图 */
.heatmap-container { padding: 12px 0; }
.heatmap-flex { display: flex; justify-content: space-between; align-items: flex-end; height: 140px; }
.heat-col { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.heat-track { width: 12px; height: 100px; background: #f8fafc; border-radius: 100px; position: relative; display: flex; align-items: flex-end; overflow: hidden; }
.heat-fill { width: 100%; border-radius: 100px; transition: height 0.8s cubic-bezier(0.16, 1, 0.3, 1); }
.lvl-0 { background: #f1f5f9; } .lvl-1 { background: #bfdbfe; } .lvl-2 { background: #60a5fa; } .lvl-3 { background: #2563eb; }
.heat-label { font-size: 11px; font-weight: 800; color: #cbd5e1; font-family: monospace; }

/* 🌟 优化后的通知中心列表 */
.notif-scroll-area { max-height: 400px; overflow-y: auto; padding: 12px; }
.notif-list { display: flex; flex-direction: column; gap: 8px; }

.notif-card {
    position: relative;
    padding: 14px;
    border-radius: 16px;
    display: flex;
    gap: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    border: 1px solid transparent;
}

.notif-card:hover { background: #f8fafc; transform: translateX(4px); }

/* 未读状态专属高亮样式 */
.notif-card.unread { background: #f0f7ff; border-color: #e0efff; }

/* 左上角未读红点指示器 */
.unread-dot {
    position: absolute;
    top: 14px;
    left: 8px;
    width: 6px;
    height: 6px;
    background: #ef4444;
    border-radius: 50%;
    box-shadow: 0 0 0 2px #fff;
}

.notif-content { flex: 1; min-width: 0; }
.notif-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.notif-user { font-size: 13px; font-weight: 800; color: #1e293b; }
.notif-time { font-size: 11px; color: #94a3b8; }
.notif-text { 
    margin: 0; 
    font-size: 13px; 
    color: #64748b; 
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.badge-count { margin-left: auto; background: #ef4444; font-weight: 900; }

/* 卷王榜与全量榜单 */
.compact-rank { display: flex; flex-direction: column; gap: 10px; }
.mini-rank-row { display: flex; align-items: center; gap: 14px; font-size: 14px; padding: 4px 0; }
.rank-idx { width: 20px; font-weight: 900; color: #cbd5e1; font-style: italic; }
.rank-idx.top-3 { color: #3b82f6; }
.rank-name { flex: 1; font-weight: 700; color: #475569; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rank-val { font-weight: 800; color: #1e293b; font-family: monospace; white-space: nowrap; }

.full-rank-list { max-height: 500px; overflow-y: auto; padding-right: 8px; }
.full-rank-row { display: flex; align-items: center; gap: 16px; padding: 16px 0; border-bottom: 1px solid #f1f5f9; }
.fr-idx { width: 30px; font-weight: 900; color: #cbd5e1; font-size: 16px; }
.fr-info { flex: 1; display: flex; flex-direction: column; }
.fr-name { font-weight: 800; font-size: 15px; color: #1e293b; }
.fr-school { font-size: 11px; color: #94a3b8; margin-top: 2px; }
.fr-score { font-size: 13px; color: #64748b; }
.fr-score strong { color: #2563eb; font-size: 18px; margin: 0 4px; }
.modal-footer { padding-top: 20px; text-align: center; }
.end-msg { font-size: 12px; color: #cbd5e1; }

@media (max-width: 1000px) {
    .dashboard-grid { grid-template-columns: 1fr; }
    .welcome-hero { flex-direction: column; gap: 32px; text-align: center; }
    .hero-left { flex-direction: column; }
    .hero-stats { width: 100%; justify-content: space-around; }
}
@media (max-width: 600px) {
    .action-grid { grid-template-columns: 1fr; }
    .hero-stat-item .value { font-size: 26px; }
    .welcome-title { font-size: 26px; }
}
</style>