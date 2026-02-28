<script setup lang="ts">
import { computed, h } from 'vue'
import { 
    NAvatar, NTag, NTime, NIcon, NImage, NButton, NDropdown, NPopconfirm 
} from 'naive-ui'
import { 
    ThumbsUpOutline, ThumbsUp, StarOutline, Star, ChatbubbleOutline, 
    TrashOutline, EllipsisVertical, WarningOutline, LockClosedOutline 
} from '@vicons/ionicons5'

const props = defineProps<{
  note: any
  isOwner: boolean
  isAdmin?: boolean
}>()

const emit = defineEmits(['like', 'collect', 'reply', 'delete', 'report', 'jump-question'])

const canDelete = computed(() => props.isOwner || props.isAdmin)

// 提取图片数组
const getImages = (n: any) => {
  if (n.images?.length) return n.images
  if (!n.content) return []
  return [...n.content.matchAll(/\[图片:(.*?)\]/g)].map(m => m[1])
}

// 清理纯文本内容 (去除自定义图片标签)
const cleanTxt = (s: string) => s ? s.replace(/\[图片:.*?\]/g, '').trim() : ''

// 清理 HTML 标签 (用于题干预览)
const stripHtml = (html: string) => {
  if (!html) return ''
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  const text = tmp.textContent || tmp.innerText || ''
  return text.length > 50 ? text.substring(0, 50) + '...' : text
}

// 提取题干摘要
const questionSnapshot = computed(() => {
  const q = props.note.question
  if (!q) return null
  let stem = q.stem ? q.stem.replace(/【(共用主干|共用题干|案例描述)】/g, '').trim() : ''
  return {
    id: q.parent_id || q.id, // 如果是子题，跳转时跳到它的父题大类
    type: q.type || '题目',
    preview: stripHtml(stem) || '查看完整题目内容...'
  }
})

// 下拉菜单
const dropdownOptions = computed(() => {
  return props.isOwner 
    ? [] // 作者本人的操作通过外层直接展示
    : [{ label: '举报', key: 'report', icon: () => h(NIcon, null, { default: () => h(WarningOutline) }), style: { color: '#d03050' } }]
})

const handleDropdown = (key: string) => {
  if (key === 'report') emit('report', props.note.id)
}
</script>

<template>
  <div class="feed-card">
    <div class="card-header">
      <n-avatar 
        round 
        size="medium" 
        :src="note.user?.avatar ? `http://localhost:8080${note.user.avatar}` : undefined" 
        fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
        class="avatar"
      />
      <div class="user-meta">
        <div class="name-row">
          <span class="name">{{ note.user?.nickname || note.user?.username || '匿名用户' }}</span>
          <n-tag v-if="isOwner" type="success" size="tiny" bordered class="role-tag">我</n-tag>
          <n-tag v-else-if="note.user?.role === 'admin'" type="error" size="tiny" class="role-tag">老师</n-tag>
          <n-icon v-if="!note.is_public" class="private-icon" size="14"><LockClosedOutline/></n-icon>
        </div>
        <div class="time"><n-time :time="new Date(note.created_at)" type="relative" /></div>
      </div>
    </div>

    <div v-if="note.parent" class="quote-block">
      <span class="q-user">回复 @{{ note.parent.user?.nickname || note.parent.user?.username }}：</span>
      <span class="q-text">{{ cleanTxt(note.parent.content) }}</span>
      <span v-if="getImages(note.parent).length" class="q-img-tag">[图片]</span>
    </div>

    <div class="card-body">
      <div class="text-content" v-if="cleanTxt(note.content)">
        {{ cleanTxt(note.content) }}
      </div>
      
      <div v-if="getImages(note).length > 0" class="image-grid" :class="`grid-${Math.min(getImages(note).length, 3)}`">
        <n-image 
          v-for="(url, idx) in getImages(note)" 
          :key="idx" 
          :src="`http://localhost:8080${url}`" 
          object-fit="cover" 
          class="grid-img"
        />
      </div>
    </div>

    <div v-if="questionSnapshot" class="question-reference" @click="emit('jump-question', questionSnapshot.id)">
      <div class="ref-icon"><n-icon><ChatbubbleOutline /></n-icon></div>
      <div class="ref-content">
        <n-tag type="info" size="tiny" round class="ref-tag">{{ questionSnapshot.type }}</n-tag>
        <span class="ref-text">{{ questionSnapshot.preview }}</span>
      </div>
      <div class="ref-arrow"><n-icon><ChevronForwardOutline /></n-icon></div>
    </div>

    <div class="card-footer">
      <div class="actions-left">
        <n-button text class="act-btn" :class="{ 'is-active': note.is_liked }" @click="emit('like', note)">
          <template #icon><n-icon><ThumbsUp v-if="note.is_liked"/><ThumbsUpOutline v-else/></n-icon></template>
          {{ note.like_count || '点赞' }}
        </n-button>
        
        <n-button text class="act-btn" :class="{ 'is-active-star': note.is_collected }" @click="emit('collect', note)">
          <template #icon><n-icon><Star v-if="note.is_collected"/><StarOutline v-else/></n-icon></template>
          {{ note.is_collected ? '已收藏' : '收藏' }}
        </n-button>
        
        <n-button text class="act-btn" @click="emit('reply', note)">
          <template #icon><n-icon><ChatbubbleOutline/></n-icon></template>回复
        </n-button>
      </div>

      <div class="actions-right">
        <n-popconfirm v-if="canDelete" @positive-click="emit('delete', note.id)">
          <template #trigger>
            <n-button text class="act-btn delete-btn"><template #icon><n-icon><TrashOutline/></n-icon></template></n-button>
          </template>
          确定要删除这条笔记吗？
        </n-popconfirm>

        <n-dropdown v-if="!isOwner" trigger="click" :options="dropdownOptions" @select="handleDropdown">
          <n-button text class="act-btn"><template #icon><n-icon><EllipsisVertical/></n-icon></template></n-button>
        </n-dropdown>
      </div>
    </div>
  </div>
</template>

<style scoped>
.feed-card {
  background: #fff;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #f1f5f9;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  transition: box-shadow 0.3s ease;
}
.feed-card:hover {
  box-shadow: 0 8px 24px rgba(0,0,0,0.06);
}

/* Header */
.card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.avatar { border: 1px solid #f1f5f9; }
.user-meta { display: flex; flex-direction: column; }
.name-row { display: flex; align-items: center; gap: 6px; }
.name { font-size: 15px; font-weight: 600; color: #1e293b; }
.role-tag { transform: scale(0.9); }
.private-icon { color: #94a3b8; }
.time { font-size: 12px; color: #94a3b8; margin-top: 2px; }

/* Quote */
.quote-block {
  background: #f8fafc;
  border-left: 3px solid #cbd5e1;
  padding: 10px 12px;
  border-radius: 6px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
}
.q-user { font-weight: 600; color: #3b82f6; }
.q-img-tag { color: #10b981; margin-left: 4px; }

/* Body */
.card-body { margin-bottom: 16px; }
.text-content {
  font-size: 15px;
  color: #334155;
  line-height: 1.7;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* Image Grid */
.image-grid { display: grid; gap: 8px; margin-top: 12px; }
.grid-1 { grid-template-columns: repeat(1, minmax(0, 200px)); }
.grid-2 { grid-template-columns: repeat(2, minmax(0, 150px)); }
.grid-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.grid-img { width: 100%; aspect-ratio: 1 / 1; border-radius: 8px; border: 1px solid #f1f5f9; cursor: zoom-in; }

/* 🔥 Question Reference (The Game Changer) */
.question-reference {
  display: flex;
  align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px 16px;
  margin-bottom: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.question-reference:hover {
  background: #eff6ff;
  border-color: #bfdbfe;
}
.ref-icon { color: #94a3b8; margin-right: 12px; display: flex; align-items: center; }
.question-reference:hover .ref-icon { color: #3b82f6; }
.ref-content { flex: 1; display: flex; align-items: center; gap: 8px; overflow: hidden; }
.ref-tag { flex-shrink: 0; }
.ref-text {
  font-size: 13px;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 500;
}
.question-reference:hover .ref-text { color: #1e293b; }
.ref-arrow { color: #cbd5e1; }

/* Footer Actions */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px dashed #f1f5f9;
}
.actions-left, .actions-right { display: flex; align-items: center; gap: 16px; }
.act-btn { font-size: 14px; color: #64748b; transition: color 0.2s; }
.act-btn:hover { color: #1e293b; background: transparent; }
.act-btn.is-active { color: #f43f5e; font-weight: 600; }
.act-btn.is-active-star { color: #f59e0b; font-weight: 600; }
.delete-btn:hover { color: #ef4444; }

@media (max-width: 768px) {
  .feed-card { padding: 16px; border-radius: 12px; }
  .grid-3 { grid-template-columns: repeat(3, minmax(0, 80px)); }
  .actions-left { gap: 8px; }
}
</style>