<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NCard, NIcon, NSkeleton, NAvatar, NTag, NTooltip, NCollapse, NCollapseItem, NEmpty,
  NNumberAnimation, NButton, NModal, NSpin, NList, NListItem, NText, useMessage
} from 'naive-ui'
import { 
  ArrowForwardOutline, BookOutline, StarOutline, JournalOutline,
  Flame, BarChartOutline, TrophyOutline, SchoolOutline, 
  TimeOutline, CheckmarkCircleOutline, TrendingUpOutline, RibbonOutline,
  NotificationsOutline, CheckmarkDoneOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const router = useRouter()
const message = useMessage()
const loading = ref(true)

// Dashboard 数据
const stats = ref({ 
    total_count: 0, 
    today_count: 0, 
    accuracy: 0,
    consecutive_days: 0,
    activity_map: [] as any[],
    subject_analysis: [] as any[], 
    rank_list: [] as any[]
})

// 通知数据
const notifications = ref<any[]>([])
const unreadCount = ref(0)
const notifLoading = ref(false)

// ===========================
// 1. Dashboard 数据获取
// ===========================
const fetchStats = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/stats') 
    if (res.data) {
        stats.value = { 
            ...stats.value, 
            ...res.data,
            activity_map: res.data.activity_map || [],
            subject_analysis: res.data.subject_analysis || [],
            rank_list: res.data.rank_list || []
        }
    }
  } catch (e) { 
    console.error("Dashboard Stats Fetch Error:", e) 
  } finally { 
    loading.value = false 
  }
}

// ===========================
// 2. 消息通知逻辑
// ===========================
const fetchNotifications = async () => {
    notifLoading.value = true
    try {
        const res: any = await request.get('/notifications')
        notifications.value = res.data || []
        unreadCount.value = res.unread_count || 0
    } catch (e) {
        console.error(e)
    } finally {
        notifLoading.value = false
    }
}

const handleNotifClick = async (item: any) => {
    // 标记已读
    if (!item.is_read) {
        try {
            await request.put(`/notifications/${item.id}/read`)
            item.is_read = true
            unreadCount.value = Math.max(0, unreadCount.value - 1)
        } catch (e) {}
    }

    // 跳转逻辑
    switch (item.source_type) {
        case 'forum':
            router.push(`/post/${item.source_id}`)
            break
        case 'question':
        case 'note':
            router.push(`/question/${item.source_id}`)
            break
        case 'system':
            message.info('系统通知：' + item.content)
            break
        default:
            message.warning('无法跳转到未知来源')
    }
}

const markAllRead = async () => {
    if (unreadCount.value === 0) return
    try {
        await request.put('/notifications/read-all')
        unreadCount.value = 0
        notifications.value.forEach(n => n.is_read = true)
        message.success('已全部标记为已读')
    } catch (e) {
        message.error('操作失败')
    }
}

// ===========================
// 3. 完整排行榜逻辑 (弹窗)
// ===========================
const showRankModal = ref(false)
const rankListFull = ref<any[]>([])
const rankLoading = ref(false)
const rankPage = ref(1)
const rankHasMore = ref(true)

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
        const res: any = await request.get('/rank/daily', {
            params: { page: rankPage.value, page_size: 20 }
        })
        if (res.data) {
            if (rankPage.value === 1) {
                rankListFull.value = res.data
            } else {
                rankListFull.value = [...rankListFull.value, ...res.data]
            }
            rankHasMore.value = res.has_more
            if (rankHasMore.value) rankPage.value++
        }
    } catch (e) {
        console.error(e)
    } finally {
        rankLoading.value = false
    }
}

const handleRankScroll = (e: Event) => {
    const target = e.target as HTMLElement
    if (target.scrollHeight - target.scrollTop - target.clientHeight < 50) {
        fetchFullRank()
    }
}

// ===========================
// 4. 通用辅助 & 头像处理
// ===========================
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 5) return '夜深了'
  if (hour < 9) return '早安'
  if (hour < 12) return '上午好'
  if (hour < 14) return '午安'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const getAvatar = (path: string) => {
    if (!path) return undefined
    return path.startsWith('http') ? path : `http://localhost:8080${path}`
}

onMounted(() => { 
    fetchStats(); 
    fetchNotifications(); // 🔥 加载通知
})
</script>

<template>
  <div class="dashboard-container">
    <div class="welcome-banner animate-enter" style="animation-delay: 0.1s;">
        <div class="banner-glass">
            <div class="user-welcome">
                <div class="avatar-ring">
                      <n-avatar 
                        round 
                        :size="72" 
                        :src="getAvatar(userStore.avatar)" 
                        fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
                        class="user-avatar"
                      />
                      <div class="status-badge"></div>
                </div>
                <div class="text-content">
                    <h2 class="greet-title">{{ greeting }}，{{ userStore.nickname || userStore.username }}</h2>
                    <p class="greet-sub">
                        <n-icon class="icon-flame"><Flame /></n-icon> 
                        已连续专注学习 <span class="highlight">{{ stats.consecutive_days }}</span> 天
                    </p>
                </div>
            </div>
            
            <div class="header-stats">
                 <div class="stat-item">
                    <div class="stat-icon-wrapper blue-grad">
                        <n-icon><TimeOutline /></n-icon>
                    </div>
                    <div class="stat-meta">
                        <div class="label">今日刷题</div>
                        <div class="value"><n-number-animation :from="0" :to="stats.today_count" /></div>
                    </div>
                 </div>
                 <div class="stat-divider"></div>
                 <div class="stat-item">
                    <div class="stat-icon-wrapper blue-grad">
                        <n-icon><CheckmarkCircleOutline /></n-icon>
                    </div>
                    <div class="stat-meta">
                        <div class="label">正确率</div>
                        <div class="value">
                           {{ stats.accuracy.toFixed(0) }}<span class="unit">%</span>
                        </div>
                    </div>
                 </div>
                 <div class="stat-divider"></div>
                 <div class="stat-item">
                      <div class="stat-icon-wrapper blue-grad">
                          <n-icon><TrendingUpOutline /></n-icon>
                      </div>
                      <div class="stat-meta">
                          <div class="label">累计做题</div>
                          <div class="value"><n-number-animation :from="0" :to="stats.total_count" /></div>
                      </div>
                 </div>
            </div>
        </div>
    </div>

    <div class="main-grid">
      
      <div class="main-column">
          
          <div class="section-actions animate-enter" style="animation-delay: 0.2s;">
              <div class="grid-actions">
                  <div class="action-card" @click="router.push('/quiz')">
                      <div class="ac-content">
                          <div class="ac-icon bg-blue-1"><n-icon><ArrowForwardOutline/></n-icon></div>
                          <div class="ac-info">
                              <h3>开始刷题</h3>
                              <p>自由选择章节练习</p>
                          </div>
                      </div>

                  </div>
                  
                  <div class="action-card" @click="router.push('/mistakes')">
                       <div class="ac-content">
                          <div class="ac-icon bg-blue-2"><n-icon><BookOutline/></n-icon></div>
                          <div class="ac-info">
                              <h3>消灭错题</h3>
                              <p>精准复习薄弱点</p>
                          </div>
                      </div>

                  </div>

                  <div class="action-card" @click="router.push('/favorites')">
                       <div class="ac-content">
                          <div class="ac-icon bg-blue-3"><n-icon><StarOutline/></n-icon></div>
                          <div class="ac-info">
                              <h3>我的收藏</h3>
                              <p>重难点考题回顾</p>
                          </div>
                      </div>

                  </div>

                  <div class="action-card" @click="router.push('/my-notes')">
                       <div class="ac-content">
                          <div class="ac-icon bg-blue-4"><n-icon><JournalOutline/></n-icon></div>
                          <div class="ac-info">
                              <h3>复习笔记</h3>
                              <p>沉淀个人知识库</p>
                          </div>
                      </div>

                  </div>
              </div>
          </div>

          <div class="chart-section animate-enter" style="animation-delay: 0.3s;">
             <n-card :bordered="false" class="panel-card" content-style="padding: 24px;">
                  <template #header>
                      <div class="card-header">
                          <div class="title-with-icon">
                              <div class="icon-box themed-box"><n-icon><BarChartOutline /></n-icon></div>
                              <span>学习热力图</span>
                          </div>
                      </div>
                  </template>
                  <div v-if="loading"><n-skeleton text :repeat="2" /></div>
                  <div v-else-if="stats.activity_map.length > 0" class="heatmap-container">
                      <div class="heatmap-scroll">
                          <n-tooltip trigger="hover" v-for="(day, index) in stats.activity_map" :key="index" placement="top">
                              <template #trigger>
                                  <div class="heat-col">
                                      <div class="heat-track">
                                          <div class="heat-fill" 
                                               :class="`level-${day.level}`" 
                                               :style="{height: (day.level * 20 + 15) + '%'}">
                                          </div>
                                      </div>
                                      <span class="heat-label">{{ day.date.split('-')[2] }}</span>
                                  </div>
                              </template>
                              <div class="heat-tooltip">
                                  <div class="tooltip-date">{{ day.date }}</div>
                                  <div class="tooltip-val">完成 <b>{{ day.count }}</b> 题</div>
                              </div>
                          </n-tooltip>
                      </div>
                  </div>
                  <n-empty v-else description="暂无记录，快去刷题点亮热力图吧！" />
              </n-card>
          </div>

          <div class="analysis-section animate-enter" style="animation-delay: 0.4s;">
              <n-card :bordered="false" class="panel-card" content-style="padding: 0;">
                   <template #header>
                      <div class="card-header">
                          <div class="title-with-icon">
                              <div class="icon-box themed-box"><n-icon><SchoolOutline /></n-icon></div>
                              <span>学科能力分布</span>
                          </div>
                          <n-tag size="small" round :bordered="false" type="primary" class="tag-label">知识点透视</n-tag>
                      </div>
                  </template>
                  
                  <div v-if="loading" style="padding: 20px;"><n-skeleton text :repeat="5" /></div>
                  <div v-else-if="stats.subject_analysis.length > 0">
                      <n-collapse display-directive="show" arrow-placement="right" class="custom-collapse">
                          <n-collapse-item v-for="sub in stats.subject_analysis" :key="sub.name" :name="sub.name">
                              <template #header>
                                  <div class="collapse-trigger">
                                      <span class="trigger-title">{{ sub.name }}</span>
                                      <div class="trigger-meta">
                                          <span class="count-badge">{{ sub.total }}题</span>
                                          <div class="mini-progress">
                                            <div class="mp-bar" :style="{width: sub.accuracy+'%', background: sub.accuracy > 60 ? '#18a058' : '#d03050'}"></div>
                                          </div>
                                          <span class="acc-val">{{ sub.accuracy.toFixed(0) }}%</span>
                                      </div>
                                  </div>
                              </template>
                              
                              <div class="sub-detail-grid">
                                  <div v-for="chap in sub.chapters" :key="chap.name" class="chapter-tile">
                                      <div class="tile-head">
                                          <span class="chap-t" :title="chap.name">{{ chap.name }}</span>
                                          <span class="chap-p">{{ chap.accuracy.toFixed(0) }}%</span>
                                      </div>
                                      <n-progress 
                                        type="line" 
                                        :percentage="Number(chap.accuracy.toFixed(1))" 
                                        :color="chap.accuracy > 80 ? '#18a058' : (chap.accuracy > 60 ? '#2080f0' : '#d03050')"
                                        :height="6"
                                        :border-radius="3"
                                        :show-indicator="false"
                                        rail-color="#f1f5f9"
                                      />
                                      <div class="tile-foot">{{ chap.total }} 道题目</div>
                                  </div>
                              </div>
                          </n-collapse-item>
                      </n-collapse>
                  </div>
                  <n-empty v-else description="暂无数据" style="padding: 40px;">
                    <template #icon><n-icon color="#cbd5e1" size="40"><SchoolOutline /></n-icon></template>
                  </n-empty>
              </n-card>
          </div>
      </div>

      <div class="side-column animate-enter" style="animation-delay: 0.5s;">
          
          <n-card :bordered="false" class="panel-card notif-panel" content-style="padding: 0;">
              <template #header>
                  <div class="card-header center-y">
                       <div class="title-with-icon">
                          <div class="icon-box notif-bg"><n-icon><NotificationsOutline /></n-icon></div>
                          <span>消息中心</span>
                          <n-tag v-if="unreadCount > 0" round type="error" size="small" class="unread-badge">{{ unreadCount }}</n-tag>
                       </div>
                       <n-button v-if="unreadCount > 0" text size="tiny" type="primary" @click="markAllRead">
                           <template #icon><n-icon><CheckmarkDoneOutline /></n-icon></template> 已读
                       </n-button>
                  </div>
              </template>
              
              <div v-if="notifLoading" style="padding: 20px;"><n-skeleton text :repeat="3" /></div>
              <div v-else-if="notifications.length > 0" class="notif-list-container">
                  <div 
                    v-for="item in notifications" 
                    :key="item.id" 
                    class="notif-item" 
                    :class="{ 'is-read': item.is_read }"
                    @click="handleNotifClick(item)"
                  >
                      <div class="notif-left">
                          <n-avatar round size="small" :src="getAvatar(item.sender?.avatar)" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
                      </div>
                      <div class="notif-body">
                          <div class="notif-top">
                              <span class="notif-user">{{ item.sender?.nickname || '匿名' }}</span>
                              <span class="notif-time">{{ new Date(item.created_at).toLocaleDateString() }}</span>
                          </div>
                          <div class="notif-action">
                              {{ item.source_type === 'forum' ? '回复了你的帖子' : '评论了你的笔记' }}
                          </div>
                          <div class="notif-content">{{ item.content }}</div>
                      </div>
                      <div v-if="!item.is_read" class="notif-dot"></div>
                  </div>
              </div>
              <n-empty v-else description="暂无新消息" style="padding: 30px;" />
          </n-card>

          <n-card :bordered="false" class="panel-card rank-panel" content-style="padding: 0;">
              <template #header>
                  <div class="card-header center-y">
                       <div class="title-with-icon">
                          <div class="icon-box themed-box"><n-icon><TrophyOutline /></n-icon></div>
                          <span>今日卷王榜</span>
                       </div>
                       <n-button text size="tiny" type="primary" @click="openRankModal">全部 ></n-button>
                  </div>
              </template>
              
              <div v-if="loading" style="padding: 20px;"><n-skeleton text :repeat="5" /></div>
              <div v-else-if="stats.rank_list.length > 0" class="rank-list">
                  <div v-for="(user, idx) in stats.rank_list" :key="user.username" class="rank-item">
                      <div class="rank-pos">
                           <template v-if="idx === 0">🥇</template>
                           <template v-else-if="idx === 1">🥈</template>
                           <template v-else-if="idx === 2">🥉</template>
                           <span v-else class="rank-num">{{ idx + 1 }}</span>
                      </div>
                      
                      <n-avatar round size="small" :src="getAvatar(user.avatar)" class="rank-avi" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
                      
                      <div class="rank-details">
                          <div class="rd-name">{{ user.username }}</div>
                          <div class="rd-score">已刷 {{ user.count }} 题</div>
                      </div>
                  </div>
              </div>
              <n-empty v-else description="今日虚位以待" style="padding: 20px;" />
              
              <div class="rank-footer">
                  <p>每日凌晨 0:00 更新榜单</p>
              </div>
          </n-card>
          
          <div class="daily-quote-card">
              <div class="quote-content">
                  <div class="quote-icon"><n-icon><RibbonOutline /></n-icon></div>
                  <p>"医者仁心，术业专攻。"</p>
              </div>
          </div>
      </div>

    </div>

    <n-modal v-model:show="showRankModal" preset="card" title="今日卷王总榜 🏆" style="width: 500px; max-width: 90%;" :bordered="false" size="huge">
        <div class="full-rank-container" @scroll="handleRankScroll">
            <div v-for="(user, idx) in rankListFull" :key="user.user_id" class="rank-row animate-in" :style="{animationDelay: idx * 0.05 + 's'}">
                <div class="rank-idx">
                   <span v-if="user.rank === 1">🥇</span>
                   <span v-else-if="user.rank === 2">🥈</span>
                   <span v-else-if="user.rank === 3">🥉</span>
                   <span v-else class="num">{{ user.rank }}</span>
                </div>
                
                <div class="rank-info">
                   <n-avatar round :size="40" :src="getAvatar(user.avatar)" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
                   <div class="info-text">
                      <div class="name-row">
                          <span class="name">{{ user.nickname || user.username }}</span>
                          <n-tag v-if="user.school" size="tiny" :bordered="false" class="school-tag">
                             {{ user.school }}
                          </n-tag>
                      </div>
                      <div class="score">今日刷题 <span class="highlight">{{ user.count }}</span></div>
                   </div>
                </div>
            </div>

            <div class="loading-state">
               <n-spin v-if="rankLoading" size="small" />
               <div v-else-if="!rankHasMore && rankListFull.length > 0" class="end-text">--- 到底了，前100名展示完毕 ---</div>
               <n-empty v-else-if="rankListFull.length === 0" description="今天还没人刷题，快去抢沙发！" />
            </div>
        </div>
    </n-modal>

  </div>
</template>

<style scoped>
/* VARIABLE DEFINITIONS */
.dashboard-container { 
    --primary: #2080f0;
    --text-main: #334155;
    --text-sub: #64748b;
    --radius-box: 16px;
    --radius-item: 12px;
    width: 100%; 
    box-sizing: border-box; 
    margin: 0 auto; 
    padding: 24px 24px 40px 24px; 
}

@keyframes slideInUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
}
.animate-enter {
    animation: slideInUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}

/* 1. Welcome Banner */
.welcome-banner {
    position: relative;
    border-radius: var(--radius-box);
    background: linear-gradient(120deg, #eff6ff 0%, #f8fafc 100%); 
    box-shadow: 0 4px 15px rgba(32, 128, 240, 0.08); 
    overflow: hidden;
    margin-bottom: 24px;
    border: 1px solid #eef2f6;
}
.banner-glass {
    padding: 30px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 24px;
}
.user-welcome {
    display: flex;
    align-items: center;
    gap: 24px;
}
.avatar-ring {
    position: relative;
    padding: 4px;
    background: #fff;
    border-radius: 50%;
    box-shadow: 0 4px 15px rgba(32, 128, 240, 0.15);
}
.status-badge {
    position: absolute;
    bottom: 5px;
    right: 5px;
    width: 14px;
    height: 14px;
    background: #2080f0; 
    border: 2px solid #fff;
    border-radius: 50%;
}
.greet-title {
    margin: 0;
    font-size: 26px;
    font-weight: 800;
    color: var(--text-main);
    letter-spacing: -0.02em;
}
.greet-sub {
    margin: 8px 0 0 0;
    font-size: 15px;
    color: var(--text-sub);
    display: flex;
    align-items: center;
    gap: 6px;
}
.icon-flame { color: #f0a020; font-size: 18px; } 
.highlight { color: #2080f0; font-weight: 800; }

/* Stats in Banner */
.header-stats {
    display: flex;
    align-items: center;
    background: rgba(255,255,255,0.8);
    padding: 16px 24px;
    border-radius: var(--radius-box);
    box-shadow: 0 4px 20px rgba(0,0,0,0.02);
    border: 1px solid #f1f5f9;
}
.stat-item { display: flex; align-items: center; gap: 14px; min-width: 120px; }
.stat-icon-wrapper {
    width: 44px; height: 44px;
    border-radius: var(--radius-item);
    display: flex; align-items: center; justify-content: center;
    font-size: 22px; color: #fff;
    box-shadow: 0 4px 10px rgba(32, 128, 240, 0.2);
}
.blue-grad { background: linear-gradient(135deg, #4299e1, #2b6cb0); }

.stat-meta .label { font-size: 12px; color: var(--text-sub); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600; margin-bottom: 2px; }
.stat-meta .value { font-size: 20px; font-weight: 900; color: #1e293b; line-height: 1; font-feature-settings: "tnum"; }
.stat-meta .unit { font-size: 12px; font-weight: 700; margin-left: 2px; color: var(--text-sub); }
.stat-divider { width: 1px; height: 36px; background: #e2e8f0; margin: 0 16px; }

/* 2. Main Grid Layout */
.main-grid {
    display: grid;
    grid-template-columns: 1fr 360px; 
    gap: 24px;
    align-items: start;
}
.main-column { display: flex; flex-direction: column; gap: 24px; min-width: 0; }
.side-column { display: flex; flex-direction: column; gap: 24px; }

/* 3. Panel Cards */
.panel-card {
    border-radius: var(--radius-box);
    box-shadow: 0 1px 4px rgba(0,0,0,0.03);
    transition: all 0.3s ease;
    border: 1px solid #f1f5f9;
    overflow: hidden;
    background: #fff;
}
.panel-card:hover { box-shadow: 0 8px 24px rgba(32, 128, 240, 0.08); transform: translateY(-2px); }

.card-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 24px; border-bottom: 1px solid #f8fafc; }
.title-with-icon { display: flex; align-items: center; gap: 10px; font-weight: 700; font-size: 16px; color: var(--text-main); }
.icon-box { width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 18px; color: #fff; }
.themed-box { background: var(--primary); box-shadow: 0 4px 10px rgba(32, 128, 240, 0.25); }
.notif-bg { background: #f59e0b; box-shadow: 0 4px 10px rgba(245, 158, 11, 0.25); }

/* 4. Quick Actions Grid */
.grid-actions {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
}
.action-card {
    position: relative;
    padding: 24px;
    border-radius: var(--radius-box);
    cursor: pointer;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    background: #fff;
    border: 2px solid transparent;
    box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.action-card:hover { transform: translateY(-4px); border-color: #bfdbfe; box-shadow: 0 8px 20px rgba(32, 128, 240, 0.1); }

.ac-content { position: relative; z-index: 2; display: flex; flex-direction: column; align-items: flex-start; gap: 16px; }
.ac-icon {
    width: 48px; height: 48px; border-radius: var(--radius-item); display: flex; align-items: center; justify-content: center; font-size: 24px; color: #fff;
    box-shadow: 0 4px 12px rgba(32, 128, 240, 0.15);
}
.bg-blue-1 { background: #3b82f6; } 
.bg-blue-2 { background: #2563eb; } 
.bg-blue-3 { background: #1d4ed8; } 
.bg-blue-4 { background: #1e40af; } 

.ac-info h3 { margin: 0; font-size: 16px; font-weight: 700; color: #1e293b; }
.ac-info p { margin: 4px 0 0 0; font-size: 12px; color: #64748b; }

.ac-bg-shape {
    position: absolute; top: -20px; right: -20px; width: 100px; height: 100px; border-radius: 50%; opacity: 0.05; transition: transform 0.5s;
}
.bg-soft-blue { background: #2080f0; }
.action-card:hover .ac-bg-shape { transform: scale(1.5); opacity: 0.1; }

/* 5. Heatmap */
.heatmap-container { padding: 10px 0; overflow: hidden; }
.heatmap-scroll { 
    display: flex; gap: 6px; align-items: flex-end; justify-content: space-between; 
    height: 140px; 
    padding-bottom: 10px;
    overflow-x: auto; 
    scrollbar-width: thin; 
}
.heat-col { display: flex; flex-direction: column; align-items: center; gap: 8px; flex: 1; min-width: 24px; cursor: pointer; height: 100%; justify-content: flex-end; }
.heat-track { width: 100%; height: 100%; background: #f8fafc; border-radius: 8px; position: relative; display: flex; align-items: flex-end; overflow: hidden; }
.heat-fill { width: 100%; border-radius: 6px; transition: height 0.6s ease; }
.level-0 { background: #e2e8f0; } 
.level-1 { background: #bfdbfe; } 
.level-2 { background: #60a5fa; } 
.level-3 { background: #3b82f6; } 
.level-4 { background: #2563eb; }
.heat-label { font-size: 10px; color: #94a3b8; font-family: monospace; }
.heat-col:hover .heat-fill { background: #1d4ed8; }

/* 6. Subject Analysis */
.custom-collapse :deep(.n-collapse-item__header) { padding: 16px 24px !important; transition: background 0.2s; }
.custom-collapse :deep(.n-collapse-item__header:hover) { background: #f8fafc; }
.collapse-trigger { display: flex; justify-content: space-between; align-items: center; width: 100%; padding-right: 12px; }
.trigger-title { font-weight: 700; color: #334155; }
.trigger-meta { display: flex; align-items: center; gap: 12px; }
.count-badge { font-size: 11px; color: #64748b; background: #f1f5f9; padding: 2px 8px; border-radius: 4px; }
.mini-progress { width: 40px; height: 4px; background: #e2e8f0; border-radius: 2px; overflow: hidden; }
.mp-bar { height: 100%; border-radius: 2px; }
.acc-val { font-size: 13px; font-weight: 700; width: 36px; text-align: right; color: var(--text-main); }

.sub-detail-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; padding: 20px 24px; background: #fafafa; }
.chapter-tile { background: #fff; border: 1px solid #f1f5f9; padding: 12px; border-radius: var(--radius-item); box-shadow: 0 1px 2px rgba(0,0,0,0.02); }
.tile-head { display: flex; justify-content: space-between; margin-bottom: 8px; font-size: 13px; font-weight: 600; color: #475569; }
.chap-t { max-width: 130px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tile-foot { font-size: 10px; color: #cbd5e1; margin-top: 6px; text-align: right; }

/* 7. Notification Panel Styles */
.notif-list-container {
    max-height: 300px;
    overflow-y: auto;
    scrollbar-width: thin;
}
.notif-item {
    display: flex;
    gap: 12px;
    padding: 16px 20px;
    border-bottom: 1px solid #f8fafc;
    cursor: pointer;
    transition: background 0.2s;
    align-items: flex-start;
}
.notif-item:hover { background-color: #fdfcff; }
.notif-item:last-child { border-bottom: none; }
.notif-item.is-read { opacity: 0.6; }
.notif-body { flex: 1; min-width: 0; }
.notif-top { display: flex; justify-content: space-between; margin-bottom: 4px; font-size: 12px; }
.notif-user { font-weight: 700; color: #334155; }
.notif-time { color: #94a3b8; }
.notif-action { font-size: 12px; color: #64748b; margin-bottom: 6px; }
.notif-content { 
    font-size: 13px; color: #1e293b; background: #f1f5f9; padding: 8px; border-radius: 8px; 
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis; 
}
.notif-dot { width: 8px; height: 8px; background: #d03050; border-radius: 50%; margin-top: 6px; flex-shrink: 0; }
.unread-badge { margin-left: 8px; height: 18px; line-height: 18px; padding: 0 6px; }

/* 8. Rank Panel */
.rank-list { display: flex; flex-direction: column; }
.rank-item { display: flex; align-items: center; padding: 12px 24px; border-bottom: 1px solid #f1f5f9; transition: background 0.2s; }
.rank-item:last-child { border-bottom: none; }
.rank-item:hover { background: #fdfcff; }
.rank-pos { width: 30px; font-size: 18px; text-align: center; font-weight: 800; color: #cbd5e1; display: flex; justify-content: center; }
.rank-num { font-size: 14px; color: #94a3b8; }
.rank-avi { margin: 0 12px; border: 2px solid #fff; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
.rank-details { flex: 1; overflow: hidden; }
.rd-name { font-size: 14px; font-weight: 600; color: #334155; }
.rd-score { font-size: 12px; color: var(--primary); font-weight: bold; }
.rank-footer { text-align: center; padding: 12px; font-size: 11px; color: #cbd5e1; border-top: 1px solid #f8fafc; background: #fafafa; }

/* 9. Quote Card */
.daily-quote-card {
    background: #1e293b; 
    color: #fff;
    padding: 24px;
    border-radius: var(--radius-box);
    position: relative;
    overflow: hidden;
    box-shadow: 0 4px 12px rgba(30, 41, 59, 0.2);
}
.quote-icon { font-size: 32px; margin-bottom: 12px; color: #3b82f6; opacity: 1; }
.quote-content p { margin: 0; font-size: 14px; line-height: 1.6; font-style: italic; color: #e2e8f0; position: relative; z-index: 10; }

/* 10. Full Rank Modal Styles */
.full-rank-container {
    height: 60vh;
    overflow-y: auto;
    padding-right: 4px;
}
.rank-row {
    display: flex;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #f1f5f9;
}
.rank-idx { width: 40px; font-size: 18px; font-weight: 800; text-align: center; margin-right: 12px; }
.rank-idx .num { color: #94a3b8; font-size: 16px; }
.rank-info { flex: 1; display: flex; align-items: center; gap: 12px; }
.info-text { display: flex; flex-direction: column; gap: 2px; }
.name-row { display: flex; align-items: center; gap: 8px; }
.name { font-weight: 700; font-size: 14px; color: #334155; }
.school-tag { background: #eff6ff; color: #3b82f6; transform: scale(0.9); transform-origin: left center; }
.score { font-size: 12px; color: #64748b; }
.loading-state { text-align: center; padding: 20px; font-size: 12px; color: #94a3b8; }

/* 11. Responsive Adjustments */
@media (max-width: 900px) {
    .main-grid { grid-template-columns: 1fr; gap: 20px; }
    .side-column { order: 2; }
}

@media (max-width: 650px) {
    .dashboard-container { padding: 0 0 40px 0; }
    .banner-glass { flex-direction: column; align-items: stretch; padding: 20px 16px; gap: 16px; }
    .user-welcome { width: 100%; justify-content: flex-start; flex-direction: row; align-items: center; }
    .user-avatar { width: 48px !important; height: 48px !important; font-size: 18px !important; }
    .avatar-ring { padding: 3px; }
    .greet-title { font-size: 20px; }
    .greet-sub { font-size: 13px; }
    .header-stats { width: 100%; display: grid; grid-template-columns: 1fr 1fr 1fr; padding: 12px 8px; gap: 4px; background: rgba(255,255,255,0.6); }
    .stat-item { flex-direction: column; align-items: center; text-align: center; min-width: auto; gap: 4px; }
    .stat-divider { display: none; }
    .stat-icon-wrapper { width: 32px; height: 32px; font-size: 16px; margin-bottom: 2px; }
    .stat-meta .label { font-size: 9px; transform: scale(0.9); }
    .stat-meta .value { font-size: 16px; font-weight: 800; }
    .grid-actions { grid-template-columns: 1fr 1fr; gap: 12px; }
    .action-card { padding: 16px; }
    .ac-icon { width: 40px; height: 40px; font-size: 20px; }
    .ac-info h3 { font-size: 14px; }
    .ac-info p { font-size: 11px; }
    .custom-collapse :deep(.n-collapse-item__header) { padding: 12px 16px !important; }
    .sub-detail-grid { padding: 12px; gap: 10px; grid-template-columns: repeat(2, 1fr); } 
    .heatmap-scroll { height: 75px; }
    .heat-tooltip { display: none; }
    .card-header { padding: 12px 16px; }
}

@media (max-width: 380px) {
    .sub-detail-grid { grid-template-columns: 1fr; }
    .header-stats { gap: 2px; }
}
</style>