<script setup lang="ts">
import { ref, computed, watch, h, onMounted, nextTick } from 'vue'
import { 
  NCard, NTag, NIcon, useMessage, NDivider, NButton, 
  NCollapseTransition, NInput, NSwitch, NSpin, NAvatar, NTime,
  NTooltip, NDropdown, NPopconfirm
} from 'naive-ui'
import { 
  StarOutline, Star, ChatboxEllipsesOutline, ChatboxEllipses, 
  GlobeOutline, LockClosedOutline, Send, EllipsisVertical, CreateOutline, TrashOutline,
  ChatbubbleOutline, ThumbsUpOutline, ThumbsUp, SchoolOutline, BookOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user' 
import QuestionItem from './QuestionItem.vue'

const props = defineProps<{ 
  question: any
  serialNumber: number,
  initShowNotes?: boolean 
}>()

const emit = defineEmits(['answer-result'])
const message = useMessage()
const userStore = useUserStore() 

const renderIcon = (icon: any) => () => h(NIcon, null, { default: () => h(icon) })

// ==========================================
// 1. 身份与权限判定
// ==========================================
const isAdmin = computed(() => userStore.role === 'admin')

const isNoteOwner = (note: any) => {
  const myId = String(userStore.id || '');
  const noteUserId = String(note.user_id || note.user?.id || '');

  if (myId && noteUserId && myId !== '0' && noteUserId !== '0') {
    return myId === noteUserId;
  }
  if (userStore.username && note.user?.username) {
    return userStore.username === note.user.username;
  }
  return false;
}

const canDelete = (note: any) => {
  return isNoteOwner(note) || isAdmin.value;
}

// ==========================================
// 2. 常规业务逻辑
// ==========================================
const isFavorited = ref(false)
const favLoading = ref(false)

watch(() => props.question, (newVal) => {
  if (newVal) isFavorited.value = !!(newVal.is_favorite || newVal.IsFavorite)
}, { immediate: true, deep: true })

const toggleFavorite = async () => { 
  if (favLoading.value) return
  favLoading.value = true
  try {
    const res: any = await request.post(`/favorites/${props.question.id}`)
    isFavorited.value = res.is_favorite
    message.success(res.message || (isFavorited.value ? '收藏成功' : '已取消收藏'))
  } catch (e) { console.error(e) } finally { favLoading.value = false }
}

const showNotes = ref(false)     
const notesList = ref<any[]>([]) 
const loadingNotes = ref(false)  
const hasLoadedNotes = ref(false)

const displayCount = computed(() => {
  if (hasLoadedNotes.value) return notesList.value.length
  return props.question.note_count || 0
})

const fetchNotes = async () => {
  if (loadingNotes.value) return 
  loadingNotes.value = true
  try {
    const res: any = await request.get('/notes', { 
        params: { question_id: Number(props.question.id) } 
    })
    notesList.value = res.data || []
    hasLoadedNotes.value = true 
  } catch (e) { console.error(e) } finally { loadingNotes.value = false }
}

watch(showNotes, (val) => { if (val) fetchNotes() })

// ==========================================
// 3. 编辑器与交互
// ==========================================
const editorState = ref({
  id: 0, content: '', isPublic: true, parentId: 0, replyToUser: ''  
})
const savingNote = ref(false)    

const resetEditor = () => {
  editorState.value = { id: 0, content: '', isPublic: true, parentId: 0, replyToUser: '' }
}

const handleSaveNote = async () => {
  if (!editorState.value.content.trim()) { message.warning('内容不能为空'); return }
  savingNote.value = true
  try {
    await request.post('/notes', {
      id: editorState.value.id, 
      question_id: Number(props.question.id),
      content: editorState.value.content,
      is_public: editorState.value.isPublic,
      parent_id: editorState.value.parentId || null
    })
    message.success(editorState.value.id > 0 ? '修改成功' : '发布成功')
    resetEditor()
    await fetchNotes() 
  } catch (e) { message.error('操作失败') } finally { savingNote.value = false }
}

const handleLike = async (note: any) => {
    const originalLiked = note.is_liked
    note.is_liked = !note.is_liked
    note.like_count += note.is_liked ? 1 : -1
    try {
        await request.post(`/notes/${note.id}/like`)
    } catch (e) {
        note.is_liked = originalLiked
        note.like_count -= note.is_liked ? 1 : -1
        message.error('点赞失败')
    }
}

const handleCollect = async (note: any) => {
    const originalCollected = note.is_collected
    note.is_collected = !note.is_collected
    try {
        await request.post(`/notes/${note.id}/collect`)
        if (note.is_collected) message.success('已收藏到笔记本')
        else message.success('已取消收藏')
    } catch (e) {
        note.is_collected = originalCollected
        message.error('收藏失败')
    }
}

const createNoteOptions = (note: any) => {
  return [
    { label: '编辑内容', key: 'edit', icon: renderIcon(CreateOutline) },
    { label: note.is_public ? '设为私密' : '设为公开', key: 'toggle_privacy', icon: renderIcon(note.is_public ? LockClosedOutline : GlobeOutline) }
  ]
}

const handleNoteAction = async (key: string, note: any) => {
  if (key === 'edit') {
    editorState.value = {
      id: note.id, content: note.content, isPublic: note.is_public, parentId: note.parent_id || 0, replyToUser: '' 
    }
    nextTick(() => document.getElementById(`note-input-${props.question.id}`)?.focus())
  } 
  else if (key === 'toggle_privacy') {
    try {
      await request.post('/notes', {
        id: note.id, question_id: Number(props.question.id), content: note.content, is_public: !note.is_public
      })
      message.success('隐私状态已更新')
      fetchNotes()
    } catch (e) { message.error('操作失败') }
  } 
  else if (key === 'delete') {
    try {
      await request.delete(`/notes/${note.id}`) 
      message.success('已删除')
      if (editorState.value.id === note.id) resetEditor()
      fetchNotes()
    } catch (e) { message.error('删除失败') }
  }
}

const handleReply = (note: any) => {
  editorState.value = {
    id: 0, content: '', isPublic: true, parentId: note.id, replyToUser: note.user?.username
  }
  nextTick(() => document.getElementById(`note-input-${props.question.id}`)?.focus())
}

const handleStatusUpdate = (payload: any) => { emit('answer-result', payload) }
const hasChildren = computed(() => props.question.children && props.question.children.length > 0)
const sharedOptions = computed(() => {
  if (hasChildren.value && props.question.options) {
    if (typeof props.question.options === 'string') {
        try { return JSON.parse(props.question.options) } catch(e) { return null }
    }
    return props.question.options
  }
  return null
})
const getTagType = (type: string) => {
  if (!type) return 'default'
  if (type.includes('A3') || type.includes('A4') || type.includes('案例')) return 'info'
  if (type.includes('B1')) return 'warning'
  return 'success'
}
const formatContent = (content: string) => {
  if (!content) return ''
  let text = content.replace(/【共用主干】/g, '').replace(/【共用题干】/g, '').replace(/【案例描述】/g, '')
  text = text.trim()
  text = text.replace(/\[图片:(.*?)\]/g, '<img src="$1" class="zoom-image" title="点击查看大图" />')
  return text
}

onMounted(() => {
  if (props.initShowNotes) {
    showNotes.value = true
  }
})
</script>

<template>
  <div :id="`question-anchor-${question.id}`" class="scroll-anchor">
    <n-card class="q-card" :bordered="true" hoverable content-style="padding-bottom: 0;">
      <template #header>
        <div class="card-header">
          <n-tag :type="getTagType(question.type)" size="small" strong round>{{ question.type || '未知题型' }}</n-tag>
          <span v-if="hasChildren" class="sub-info">(共 {{ question.children.length }} 小题)</span>
        </div>
      </template>

      <div v-if="hasChildren">
          <div v-if="question.stem && !question.type.includes('B1')" class="clinical-context">
             <div class="context-header">（主题干）</div>
             <div class="context-content" v-html="formatContent(question.stem)"></div>
          </div>

          <div v-if="question.type.includes('B1') && sharedOptions" class="b1-context">
             <div class="b1-header">（共用备选答案）</div>
             <div class="b1-options-list">
                <div v-for="(optText, key) in sharedOptions" :key="key" class="b1-opt-row">
                   <span class="b1-key">{{ key }}.</span> 
                   <span class="b1-text" v-html="optText"></span>
                </div>
             </div>
          </div>

          <div class="children-list">
             <div 
               v-for="(child, idx) in question.children" 
               :key="child.id"
               :id="`question-anchor-${child.id}`" 
               class="scroll-anchor"
               style="position: relative;"
             >
                <div class="global-index-mark">Topic {{ child.displayIndex }}</div>
                <QuestionItem 
                  :question="child" :shared-options="sharedOptions" :index="idx + 1" :is-child="true" 
                  :show-shared-header="false" @answer-result="handleStatusUpdate" 
                />
                <n-divider v-if="idx < question.children.length - 1" dashed style="margin: 0;" />
             </div>
          </div>
      </div>
      <div v-else>
        <QuestionItem :question="question" :index="serialNumber" :is-child="false" @answer-result="handleStatusUpdate" />
      </div>

      <template #action>
        <div class="action-bar">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button text class="action-btn star-btn" :class="{ 'active': isFavorited }" @click="toggleFavorite">
                <template #icon><n-icon size="18" style="position: relative; top: -1px;"><Star v-if="isFavorited" /><StarOutline v-else /></n-icon></template>
                {{ isFavorited ? '已收藏' : '收藏' }}
              </n-button>
            </template>
            {{ isFavorited ? '点击取消收藏' : '添加到收藏夹' }}
          </n-tooltip>

          <n-button text class="action-btn" :class="{ 'active': showNotes }" @click="showNotes = !showNotes">
            <template #icon><n-icon size="18"><ChatboxEllipses v-if="showNotes" /><ChatboxEllipsesOutline v-else /></n-icon></template>
            {{ showNotes ? '收起讨论' : `讨论 (${displayCount})` }}
          </n-button>
        </div>
      </template>

      <n-collapse-transition :show="showNotes">
        <div class="notes-wrapper">
          
          <div class="note-editor-box" :class="{ 'reply-mode': editorState.parentId > 0 }">
             <div v-if="editorState.parentId > 0" class="reply-badge">
                <span style="color: #666;">正在回复</span> 
                <span style="font-weight: bold; color: #18a058; margin: 0 4px;">@{{ editorState.replyToUser }}</span>
                <n-button size="tiny" circle quaternary @click="resetEditor"><template #icon><n-icon><TrashOutline/></n-icon></template></n-button>
             </div>
             
             <n-input
              v-model:value="editorState.content"
              type="textarea"
              :placeholder="editorState.parentId > 0 ? '友善回复，共同进步...' : '写下你的见解或疑问...'"
              :autosize="{ minRows: 2, maxRows: 6 }"
              class="note-textarea"
              :id="`note-input-${question.id}`"
            />
            <div class="editor-toolbar">
               <div class="toolbar-left">
                  <span class="tip-text" v-if="editorState.id > 0">正在修改旧笔记... <n-button size="tiny" type="warning" text @click="resetEditor">取消修改</n-button></span>
                  <span class="tip-text" v-else>分享你的解题思路。</span>
               </div>
               <div class="toolbar-right">
                  <n-switch v-model:value="editorState.isPublic" size="small">
                    <template #checked-icon><n-icon><GlobeOutline /></n-icon></template>
                    <template #unchecked-icon><n-icon><LockClosedOutline /></n-icon></template>
                  </n-switch>
                  <span class="privacy-label" :class="{ 'public': editorState.isPublic }">{{ editorState.isPublic ? '公开' : '私密' }}</span>
                  <n-button type="primary" size="small" class="send-btn" :loading="savingNote" :disabled="!editorState.content.trim()" @click="handleSaveNote">
                    <template #icon><n-icon><Send /></n-icon></template>
                    {{ editorState.id > 0 ? '保存修改' : '发布' }}
                  </n-button>
               </div>
            </div>
          </div>
          
          <div class="notes-list-section">
              <n-spin :show="loadingNotes">
                  <div v-if="notesList.length === 0" class="empty-state">
                      <n-icon size="40" color="#e0e0e0"><ChatboxEllipsesOutline /></n-icon>
                      <p>还没有讨论，发一条占楼吧！</p>
                  </div>
                  <div v-else class="note-feed">
                      <div class="note-item" v-for="note in notesList" :key="note.id">
                          <div class="note-avatar-col">
                              <n-avatar 
                                round size="small" 
                                :src="note.user?.avatar ? `http://localhost:8080${note.user.avatar}` : undefined"
                                fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
                                :style="{ border: '1px solid #eee' }"
                              />
                          </div>
                          
                          <div class="note-body-col">
                              <div class="note-header-row">
                                  <div class="user-info-group">
                                      <span class="username">{{ note.user?.nickname || note.user?.username }}</span>
                                      
                                      <n-tag v-if="isNoteOwner(note)" type="success" size="tiny" bordered class="role-tag">我</n-tag>
                                      <n-tag v-else-if="note.user?.role === 'admin'" type="error" size="tiny" class="role-tag">老师</n-tag>

                                      <n-tag v-if="note.user?.school" size="tiny" :bordered="false" class="school-tag">
                                        {{ note.user.school }}
                                      </n-tag>

                                      <n-tag v-if="note.user?.grade || note.user?.major" size="tiny" :bordered="false" class="major-tag">
                                        {{ note.user?.grade }} {{ note.user?.major }}
                                      </n-tag>

                                      <span class="time"><n-time :time="new Date(note.created_at)" type="relative" /></span>
                                      <div v-if="!note.is_public" class="private-badge" title="私密笔记"><n-icon><LockClosedOutline /></n-icon></div>
                                  </div>

                                  <div class="action-group">
                                      <n-button text size="tiny" class="icon-btn like-btn" :class="{ 'liked': note.is_liked }" @click="handleLike(note)">
                                          <template #icon><n-icon><ThumbsUp v-if="note.is_liked" /><ThumbsUpOutline v-else /></n-icon></template>
                                          {{ note.like_count > 0 ? note.like_count : '' }}
                                      </n-button>

                                      <n-button text size="tiny" class="icon-btn collect-btn" :class="{ 'collected': note.is_collected }" @click="handleCollect(note)">
                                          <template #icon><n-icon><Star v-if="note.is_collected" /><StarOutline v-else /></n-icon></template>
                                      </n-button>

                                      <n-button text size="tiny" class="icon-btn reply-btn" @click="handleReply(note)">
                                          <template #icon><n-icon><ChatbubbleOutline /></n-icon></template>
                                      </n-button>

                                      <template v-if="canDelete(note)">
                                          <n-popconfirm @positive-click="handleNoteAction('delete', note)">
                                              <template #trigger>
                                                  <n-button text size="tiny" class="icon-btn delete-btn" :title="isAdmin && !isNoteOwner(note) ? '管理员删除' : '删除'">
                                                      <template #icon><n-icon><TrashOutline /></n-icon></template>
                                                  </n-button>
                                              </template>
                                              <span v-if="isAdmin && !isNoteOwner(note)">[管理员操作] </span>
                                              确定删除这条笔记？
                                          </n-popconfirm>
                                      </template>

                                      <template v-if="isNoteOwner(note)">
                                          <n-dropdown 
                                            trigger="click" placement="bottom-end"
                                            :options="createNoteOptions(note)"
                                            @select="(key) => handleNoteAction(key, note)"
                                          >
                                            <n-button text size="tiny" class="icon-btn more-btn">
                                                <template #icon><n-icon><EllipsisVertical /></n-icon></template>
                                            </n-button>
                                          </n-dropdown>
                                      </template>
                                  </div>
                              </div>

                              <div v-if="note.parent" class="parent-quote-block">
                                  <div class="quote-header">
                                     回复 <span class="quote-user">@{{ note.parent.user?.nickname || note.parent.user?.username }}</span> :
                                  </div>
                                  <div class="quote-content text-ellipsis">
                                     {{ note.parent.content }}
                                  </div>
                              </div>

                              <div class="note-text">{{ note.content }}</div>
                          </div>
                      </div>
                  </div>
              </n-spin>
          </div>
        </div>
      </n-collapse-transition>
    </n-card>
  </div>
</template>

<style>
.zoom-image { max-width: 100%; max-height: 200px; border-radius: 4px; cursor: zoom-in; margin: 8px 0; display: block; border: 1px solid #eee; }
</style>

<style scoped>
.note-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  width: 100%;
}

.user-info-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap; 
  flex: 1; 
  overflow: hidden;
}

.action-group {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0; 
  margin-left: 8px;
}

.icon-btn { font-size: 16px; color: #999; transition: all 0.2s; padding: 4px; }
.icon-btn:hover { color: #333; background-color: rgba(0,0,0,0.05); border-radius: 4px; }

.like-btn:hover, .like-btn.liked { color: #d03050; }
.collect-btn:hover, .collect-btn.collected { color: #f0a020; }
.reply-btn:hover { color: #18a058; }
.delete-btn:hover { color: #d03050; background-color: #fff0f0; }
.more-btn:hover { color: #2080f0; }

.username { font-weight: bold; color: #333; font-size: 13px; }
.role-tag { transform: scale(0.9); }
.school-tag { background-color: #f0f7ff; color: #2080f0; transform: scale(0.9); transform-origin: left center; }
/* 🔥🔥🔥 年级专业标签样式 🔥🔥🔥 */
.major-tag { background-color: #f5f7fa; color: #666; transform: scale(0.9); transform-origin: left center; }

.time { color: #ccc; font-size: 12px; margin-left: 2px; }
.private-badge { color: #f0a020; font-size: 12px; display: flex; align-items: center; }

.reply-mode { border-color: #18a058; background-color: #f0f9f4; }
.reply-badge { display: flex; align-items: center; font-size: 12px; margin-bottom: 6px; padding-bottom: 6px; border-bottom: 1px dashed #dcdcdc; }
.scroll-anchor { scroll-margin-top: 80px; }
.q-card { margin-bottom: 24px; border-radius: 8px; transition: box-shadow 0.3s; }
.q-card:hover { box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); }
.card-header { display: flex; align-items: center; gap: 10px; }
.sub-info { font-size: 12px; color: #999; font-weight: normal; }
.clinical-context { background: #fdfdfd; border: 1px solid #e0e0e0; border-left: 4px solid #2080f0; border-radius: 4px; padding: 12px 16px; margin-bottom: 20px; }
.context-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; font-size: 14px; font-weight: bold; color: #18a058; }
.context-content { font-size: 15px; color: #444; line-height: 1.6; text-align: justify; }
.children-list { display: flex; flex-direction: column; gap: 0; padding-left: 0; }
.action-bar { display: flex; align-items: center; gap: 24px; }
.action-btn { font-size: 14px; color: #666; transition: all 0.2s; }
.action-btn:hover { color: #2080f0; }
.star-btn:hover, .star-btn.active { color: #f0a020 !important; }
.action-btn.active { color: #2080f0; }
.notes-wrapper { background-color: #fafafa; border-top: 1px solid #eee; padding: 16px 20px; }
.note-editor-box { background-color: #fff; border: 1px solid #e0e0e0; border-radius: 6px; padding: 12px; box-shadow: 0 1px 2px rgba(0,0,0,0.03); transition: border-color 0.2s; }
.note-editor-box:focus-within { border-color: #2080f0; }
.note-textarea { --n-border: none !important; --n-box-shadow-focus: none !important; padding: 0; background-color: transparent; }
.editor-toolbar { display: flex; justify-content: space-between; align-items: center; margin-top: 10px; padding-top: 10px; border-top: 1px dashed #f0f0f0; }
.tip-text { font-size: 12px; color: #bbb; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.privacy-label { font-size: 12px; color: #999; min-width: 28px; }
.privacy-label.public { color: #18a058; }
.send-btn { padding-left: 12px; padding-right: 12px; }
.notes-list-section { margin-top: 24px; }
.empty-state { text-align: center; padding: 30px 0; color: #aaa; display: flex; flex-direction: column; align-items: center; gap: 10px; font-size: 13px; }
.note-feed { display: flex; flex-direction: column; gap: 16px; }
.note-item { display: flex; gap: 12px; align-items: flex-start; }
.note-avatar-col { flex-shrink: 0; padding-top: 2px; }
.note-body-col { flex: 1; background-color: #fff; border: 1px solid #eee; border-radius: 8px; padding: 10px 14px; position: relative; }
.note-body-col::before { content: ''; position: absolute; left: -6px; top: 12px; width: 10px; height: 10px; background: #fff; border-left: 1px solid #eee; border-bottom: 1px solid #eee; transform: rotate(45deg); }
.note-text { font-size: 14px; color: #444; line-height: 1.6; white-space: pre-wrap; word-break: break-all; }
.parent-quote-block {
  background-color: #f7f9fc;
  border-left: 3px solid #d0d0d0;
  padding: 8px 12px;
  margin-bottom: 10px;
  border-radius: 0 4px 4px 0;
  font-size: 12px;
  color: #666;
}
.quote-header { margin-bottom: 4px; color: #999; }
.quote-user { color: #18a058; font-weight: bold; margin-right: 4px; }
.quote-content { color: #555; line-height: 1.4; }
.text-ellipsis { display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 2; overflow: hidden; text-overflow: ellipsis; }
.b1-context {
  background-color: #fcfcfc;
  border: 1px solid #eee;
  border-left: 4px solid #18a058;
  padding: 12px 16px;
  margin-bottom: 20px;
  border-radius: 4px;
}
.b1-header {
  font-size: 13px;
  font-weight: bold;
  color: #18a058;
  margin-bottom: 8px;
}
.b1-options-list {
  display: flex;
  flex-direction: column; 
  gap: 8px;
}
.b1-opt-row {
  font-size: 14px;
  color: #333;
  line-height: 1.5;
}
.b1-key {
  font-weight: bold;
  color: #18a058;
  margin-right: 6px;
}
.global-index-mark {
  display: inline-block;
  font-size: 12px;
  color: #2080f0;
  background-color: #eef6fc;
  padding: 2px 8px;
  border-radius: 4px;
  margin-bottom: 6px;
  font-weight: 500;
  user-select: none;
}
</style>