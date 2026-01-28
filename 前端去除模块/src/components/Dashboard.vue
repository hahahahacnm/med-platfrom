<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NCard, NGrid, NGridItem, NIcon, NProgress, NSkeleton, NSpace, 
  NButton, NAvatar, NList, NListItem, NTag, NTooltip, NCollapse, NCollapseItem, NEmpty
} from 'naive-ui'
import { 
  ArrowForwardOutline, BookOutline, StarOutline, JournalOutline,
  FlameOutline, PodiumOutline, SchoolOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()
const router = useRouter()
const loading = ref(true)

const stats = ref({ 
    total_count: 0, 
    today_count: 0, 
    accuracy: 0,
    consecutive_days: 0,
    activity_map: [] as any[],
    subject_analysis: [] as any[], 
    rank_list: [] as any[]
})

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

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 5) return '夜深了'
  if (hour < 9) return '早安'
  if (hour < 12) return '上午好'
  if (hour < 14) return '午安'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const scrollToLibrary = () => {
    const container = document.querySelector('#question-scroll-container')
    if (container) container.scrollTo({ top: 0, behavior: 'smooth' })
    else window.scrollTo({ top: 500, behavior: 'smooth' })
}

const getAvatar = (path: string) => {
    if (!path) return undefined
    return path.startsWith('http') ? path : `http://localhost:8080${path}`
}

onMounted(() => { fetchStats() })
</script>

<template>
  <div class="dashboard-container">
    <div class="welcome-section">
        <div class="banner-content">
            <div class="text-group">
                <h2 class="greet-title">{{ greeting }}，{{ userStore.nickname || userStore.username }}</h2>
                <p class="greet-sub">
                    <n-icon style="vertical-align: middle; margin-right: 4px; color: #f0a020;"><FlameOutline /></n-icon>
                    已连续学习 <b style="color: #f0a020;">{{ stats.consecutive_days }}</b> 天
                </p>
            </div>
            <div class="quick-stats">
                <div class="qs-item"><div class="qs-val">{{ stats.today_count }}</div><div class="qs-label">今日刷题</div></div>
                <div class="qs-divider"></div>
                <div class="qs-item"><div class="qs-val">{{ stats.total_count }}</div><div class="qs-label">累计做题</div></div>
                <div class="qs-divider"></div>
                <div class="qs-item">
                    <div class="qs-val" :class="stats.accuracy > 60 ? 'good' : 'bad'">{{ stats.accuracy.toFixed(0) }}<span class="pct">%</span></div>
                    <div class="qs-label">正确率</div>
                </div>
            </div>
        </div>
    </div>

    <n-grid x-gap="16" y-gap="16" cols="1 800:3" style="margin-top: 20px;">
      <n-grid-item span="2">
          <n-space vertical :size="20">
              <n-card :bordered="false" class="panel-card" title="📅 学习强度">
                  <div v-if="loading"><n-skeleton text :repeat="2" /></div>
                  <div v-else-if="stats.activity_map.length > 0" class="heatmap-container">
                      <n-tooltip trigger="hover" v-for="(day, index) in stats.activity_map" :key="index">
                          <template #trigger>
                              <div class="heat-column">
                                  <div class="heat-block" :class="`level-${day.level}`" :style="{ height: day.level * 20 + 20 + '%' }"></div>
                                  <span class="heat-date">{{ day.date }}</span>
                              </div>
                          </template>
                          {{ day.date }}: {{ day.count }} 题
                      </n-tooltip>
                  </div>
                  <n-empty v-else description="暂无记录" />
              </n-card>

              <n-card :bordered="false" class="panel-card" title="📊 学科能力透视">
                  <template #header-extra><span style="font-size: 12px; color: #999;">按学科 > 大章节统计</span></template>
                  
                  <div v-if="loading">
                      <n-skeleton text :repeat="5" style="margin-bottom: 10px;" />
                  </div>
                  <div v-else-if="stats.subject_analysis.length > 0">
                      <n-collapse accordion display-directive="show">
                          <n-collapse-item v-for="sub in stats.subject_analysis" :key="sub.name" :name="sub.name">
                              <template #header>
                                  <div class="collapse-header">
                                      <span class="sub-title">{{ sub.name }}</span>
                                      <div class="sub-meta">
                                          <n-tag size="small" :bordered="false" type="default" style="margin-right: 8px">{{ sub.total }} 题</n-tag>
                                          <span :class="sub.accuracy > 60 ? 'text-good' : 'text-bad'" style="font-size: 13px; font-weight: bold;">
                                              {{ sub.accuracy.toFixed(0) }}%
                                          </span>
                                      </div>
                                  </div>
                              </template>
                              
                              <div class="chapter-grid">
                                  <div v-for="chap in sub.chapters" :key="chap.name" class="chapter-card">
                                      <div class="chap-info">
                                          <span class="chap-name" :title="chap.name">{{ chap.name }}</span>
                                          <span class="chap-stat">{{ chap.total }} 题</span>
                                      </div>
                                      <n-progress 
                                        type="line" 
                                        :percentage="Number(chap.accuracy.toFixed(1))" 
                                        :color="chap.accuracy > 80 ? '#18a058' : (chap.accuracy > 60 ? '#2080f0' : '#d03050')"
                                        :height="6"
                                        :border-radius="3"
                                        :show-indicator="false"
                                      />
                                      <div class="chap-acc">正确率 {{ chap.accuracy.toFixed(0) }}%</div>
                                  </div>
                              </div>
                          </n-collapse-item>
                      </n-collapse>
                  </div>
                  <n-empty v-else description="做题太少，暂无法分析">
                    <template #icon><n-icon><SchoolOutline /></n-icon></template>
                  </n-empty>
              </n-card>

              <n-grid x-gap="12" y-gap="12" cols="2 600:4">
                <n-grid-item v-for="(item, idx) in [
                    {t:'开始刷题', d:'选择章节', i:ArrowForwardOutline, c:'primary', f:scrollToLibrary},
                    {t:'消灭错题', d:'精准复习', i:BookOutline, c:'error', f:()=>router.push('/mistakes')},
                    {t:'我的收藏', d:'重点关注', i:StarOutline, c:'warning', f:()=>router.push('/favorites')},
                    {t:'复习笔记', d:'学习心得', i:JournalOutline, c:'info', f:()=>router.push('/my-notes')}
                ]" :key="idx">
                    <div class="action-btn-box" @click="item.f">
                        <div class="icon-circle" :class="item.c"><n-icon><component :is="item.i"/></n-icon></div>
                        <div class="action-info"><div class="title">{{item.t}}</div><div class="desc">{{item.d}}</div></div>
                    </div>
                </n-grid-item>
              </n-grid>
          </n-space>
      </n-grid-item>

      <n-grid-item>
          <n-card :bordered="false" class="panel-card" content-style="padding: 0;">
              <template #header><div style="display: flex; align-items: center; gap: 8px;"><n-icon color="#f0a020"><PodiumOutline /></n-icon> 卷王榜</div></template>
              <div v-if="loading" style="padding: 20px;"><n-skeleton text :repeat="5" /></div>
              <n-list v-else-if="stats.rank_list.length > 0" hoverable>
                  <n-list-item v-for="(user, idx) in stats.rank_list" :key="user.username">
                      <div class="rank-item">
                          <div class="rank-num" :class="`top-${idx+1}`">{{ idx + 1 }}</div>
                          <n-avatar round size="small" :src="getAvatar(user.avatar)" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
                          <div class="rank-user">{{ user.username }}</div>
                          <div class="rank-score">{{ user.count }} 题</div>
                      </div>
                  </n-list-item>
              </n-list>
              <n-empty v-else description="暂无排名" style="padding: 20px;" />
          </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>

<style scoped>
.dashboard-container { padding: 0 12px 40px 12px; max-width: 1200px; margin: 0 auto; }
.welcome-section { background: linear-gradient(120deg, #e3f2fd 0%, #f0f9eb 100%); border-radius: 12px; padding: 20px 24px; border: 1px solid #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.03); }
.banner-content { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 20px; }
.greet-title { margin: 0; font-size: 20px; font-weight: 800; color: #333; }
.greet-sub { margin: 4px 0 0 0; color: #666; font-size: 14px; }
.quick-stats { display: flex; align-items: center; background: rgba(255,255,255,0.7); padding: 8px 16px; border-radius: 8px; backdrop-filter: blur(4px); }
.qs-item { text-align: center; min-width: 60px; }
.qs-val { font-size: 18px; font-weight: 800; color: #333; }
.qs-val.good { color: #18a058; } .qs-val.bad { color: #d03050; }
.qs-val .pct { font-size: 12px; font-weight: normal; color: #999; }
.qs-label { font-size: 11px; color: #999; }
.qs-divider { width: 1px; height: 20px; background: #ddd; margin: 0 12px; }

.panel-card { border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.02); transition: all 0.3s; }
.panel-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,0.06); }

.heatmap-container { display: flex; justify-content: space-between; align-items: flex-end; height: 100px; padding: 10px 0; }
.heat-column { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; height: 100%; gap: 6px; cursor: pointer; }
.heat-block { width: 60%; background: #eee; border-radius: 3px 3px 0 0; transition: all 0.3s; min-height: 4px; }
.heat-date { font-size: 9px; color: #999; }
.level-0 { background: #f0f0f0; } .level-1 { background: #bcf0da; } .level-2 { background: #6fcf97; } .level-3 { background: #27ae60; } .level-4 { background: #219653; }

.collapse-header { display: flex; justify-content: space-between; align-items: center; width: 100%; padding-right: 10px; }
.sub-title { font-weight: bold; font-size: 14px; color: #333; }
.sub-meta { display: flex; align-items: center; }
.text-good { color: #18a058; } .text-bad { color: #d03050; }

.chapter-grid { 
    display: grid; 
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); 
    gap: 12px; 
    padding: 12px; 
    background-color: #fafafa; 
    border-radius: 4px; 
}
.chapter-card { 
    background: #fff; 
    border: 1px solid #eee; 
    border-radius: 6px; 
    padding: 10px; 
}
.chap-info { display: flex; justify-content: space-between; margin-bottom: 6px; font-size: 13px; }
.chap-name { font-weight: bold; color: #555; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 160px; }
.chap-stat { color: #999; font-size: 11px; }
.chap-acc { font-size: 11px; color: #999; margin-top: 4px; text-align: right; }

.action-btn-box { display: flex; align-items: center; padding: 12px; border-radius: 10px; background: #fff; cursor: pointer; transition: all 0.2s; border: 1px solid #f5f5f5; height: 100%; }
.action-btn-box:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.06); transform: translateY(-2px); border-color: #eee; }
.icon-circle { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 18px; color: #fff; flex-shrink: 0; margin-right: 10px; }
.icon-circle.primary { background: #2080f0; } .icon-circle.error { background: #d03050; } .icon-circle.warning { background: #f0a020; } .icon-circle.info { background: #18a058; }
.action-info .title { font-weight: bold; font-size: 13px; color: #333; } .action-info .desc { font-size: 11px; color: #999; }

.rank-item { display: flex; align-items: center; padding: 8px 16px; gap: 10px; }
.rank-num { width: 20px; font-weight: bold; color: #999; font-size: 13px; text-align: center; }
.rank-num.top-1 { color: #f0a020; font-size: 16px; } .rank-num.top-2 { color: #999999; font-size: 15px; } .rank-num.top-3 { color: #b87333; font-size: 15px; }
.rank-user { flex: 1; font-weight: 500; color: #333; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rank-score { font-size: 12px; color: #18a058; font-weight: bold; }

@media (max-width: 600px) {
    .chapter-grid { grid-template-columns: 1fr; }
}
</style>