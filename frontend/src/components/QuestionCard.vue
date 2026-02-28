<script setup lang="ts">
import { ref, computed, watch, h, onMounted, nextTick, reactive } from 'vue'
import { 
  NCard, NTag, NIcon, useMessage, NDivider, NButton, 
  NCollapseTransition, NInput, NSpin, NAvatar, NTime,
  NTooltip, NDropdown, NPopconfirm, NUpload, NPopover, NImage,
  NModal, NRadioGroup, NRadio, NSpace
} from 'naive-ui'
import { 
  StarOutline, Star, ChatboxEllipsesOutline, ChatboxEllipses, 
  GlobeOutline, LockClosedOutline, Send, EllipsisVertical, CreateOutline, TrashOutline,
  ChatbubbleOutline, ThumbsUpOutline, ThumbsUp, HappyOutline, CloseCircleOutline,
  AddOutline, CloseOutline, WarningOutline, ConstructOutline, FlameOutline, TimeOutline,
  SchoolOutline // 🔥 新增学校图标
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user' 
import QuestionItem from './QuestionItem.vue'

const props = defineProps<{ question: any, serialNumber: number, initShowNotes?: boolean }>()
const emit = defineEmits(['answer-result'])
const message = useMessage()
const userStore = useUserStore() 

const renderIcon = (icon: any) => () => h(NIcon, null, { default: () => h(icon) })
const isAdmin = computed(() => userStore.role === 'admin')
const isNoteOwner = (note: any) => {
  const myId = String(userStore.id || ''), uid = String(note.user_id || note.user?.id || '')
  if (myId && uid && myId !== '0' && uid !== '0') return myId === uid
  return userStore.username && note.user?.username && userStore.username === note.user.username
}
const canDelete = (note: any) => isNoteOwner(note) || isAdmin.value

/* 🌟 收藏逻辑 */
const isFavorited = ref(false), favLoading = ref(false)
watch(() => props.question, v => isFavorited.value = !!(v?.is_favorite || v?.IsFavorite), { immediate: true, deep: true })
const toggleFavorite = async () => { 
  if (favLoading.value) return
  favLoading.value = true
  try {
    const res: any = await request.post(`/favorites/${props.question.id}`)
    isFavorited.value = res.is_favorite; message.success(res.message || (res.is_favorite ? '收藏成功' : '已取消'))
  } catch (e) { console.error(e) } finally { favLoading.value = false }
}

/* 📝 笔记列表 (分页 + 排序) */
const showNotes = ref(false), notesList = ref<any[]>([]), loadingNotes = ref(false)
const displayCount = computed(() => notesList.value.length || props.question.note_count || 0)

const notePage = ref(1)
const hasMoreNotes = ref(false)
const noteSort = ref('hot') 

const fetchNotes = async (isLoadMore = false) => {
  if (loadingNotes.value) return 
  loadingNotes.value = true
  if (!isLoadMore) { notePage.value = 1; notesList.value = [] }

  try {
    const res: any = await request.get('/notes', { 
        params: { question_id: Number(props.question.id), page: notePage.value, page_size: 5, sort: noteSort.value } 
    })
    if (isLoadMore) { notesList.value = [...notesList.value, ...res.data] } else { notesList.value = res.data || [] }
    hasMoreNotes.value = res.has_more || false
    if (hasMoreNotes.value) notePage.value++
  } catch (e) { console.error(e) } finally { loadingNotes.value = false }
}

const toggleSort = () => { noteSort.value = noteSort.value === 'hot' ? 'time' : 'hot'; fetchNotes(false) }
watch(showNotes, v => { if (v && notesList.value.length === 0) fetchNotes() })

/* ✏️ 编辑器核心 */
const editorState = reactive({ id: 0, content: '', isPublic: true, parentId: 0, replyToUser: '', images: [] as string[] })
const savingNote = ref(false), showEmojiPicker = ref(false)
const emojiList = ['😀','😃','😄','😁','😆','😅','😂','🤣','😊','😇','🙂','🙃','😉','😌','😍','🥰','😘','😗','😙','😚','😋','😛','😝','😜','🤪','🤨','🧐','🤓','😎','🤩','🥳','😏','😒','😞','😔','😟','😕','🙁','☹️','😣','😖','😫','😩','🥺','😢','😭','😤','😠','😡','🤬','🤯','😳','🥵','🥶','😱','😨','😰','😥','😓','🤗','🤔','🤭','🤫','🤥','😶','😐','😑','😬','🙄','😯','😦','😧','😮','😲','😴','🤤','😪','😵','🤐','🥴','🤢','🤮','🤧','😷','🤒','🤕','🤑','🤠','😈','👿','👹','👺','🤡','💩','👻','💀','☠️','👽','👾','🤖','🎃','😺','😸','😹','😻','😼','😽','🙀','😿','😾']

const resetEditor = () => Object.assign(editorState, { id: 0, content: '', isPublic: true, parentId: 0, replyToUser: '', images: [] })
const insertEmoji = (emoji: string) => {
  if (editorState.content.length + emoji.length > 200) return message.warning('字数已达上限')
  const el = document.getElementById(`note-input-${props.question.id}`) as HTMLTextAreaElement
  if (el) {
    const s = el.selectionStart, e = el.selectionEnd, txt = editorState.content
    editorState.content = txt.substring(0, s) + emoji + txt.substring(e)
    nextTick(() => { el.focus(); el.selectionStart = el.selectionEnd = s + emoji.length })
  } else editorState.content += emoji
  showEmojiPicker.value = false
}
const handleUpload = async ({ file }: { file: any }) => {
  if (editorState.images.length >= 5) return message.warning('最多上传5张图片')
  const form = new FormData(); form.append('file', file.file)
  try {
    const res: any = await request.post('/notes/upload', form, { headers: { 'Content-Type': 'multipart/form-data' } })
    editorState.images.push(res.url)
  } catch (e: any) { message.error('上传失败') }
}
const handlePaste = (e: ClipboardEvent) => {
  const items = e.clipboardData?.items; if (!items) return
  for (let i = 0; i < items.length; i++) {
    if (items[i]?.type?.includes('image')) {
      e.preventDefault(); 
      if (editorState.images.length >= 5) { message.warning('最多上传5张图片'); return }
      const file = items[i]?.getAsFile?.() ?? null
      if (file) { handleUpload({ file: { file } }); message.info('正在上传粘贴图片...') }
      break 
    }
  }
}
const saveNote = async () => {
  if (!editorState.content.trim() && !editorState.images.length) return message.warning('内容不能为空')
  if (editorState.content.length > 200) return message.warning('字数不能超过200字')
  savingNote.value = true
  try {
    await request.post('/notes', {
      id: editorState.id, question_id: Number(props.question.id), content: editorState.content,
      is_public: editorState.isPublic, parent_id: editorState.parentId || null, images: editorState.images 
    })
    message.success(editorState.id ? '修改成功' : '发布成功'); resetEditor(); fetchNotes(false)
  } catch (e: any) { message.error(e.response?.data?.error || '操作失败') } finally { savingNote.value = false }
}

// 🔥 核心修复逻辑：监听题目 ID 变化，彻底解决“旧的不去新的不来”
watch(() => props.question.id, (newId) => {
  if (!newId) return
  
  // 1. 自动收起评论区
  // 这样用户切换到下一题时，界面会恢复整洁，而不是留着上一题庞大的评论列表
  showNotes.value = false 
  
  // 2. 彻底清空上一题的评论缓存
  // 即使随后再次点开，也会因为列表为空而重新触发 fetchNotes() 获取新数据
  notesList.value = []
  
  // 3. 重置页码和加载状态
  notePage.value = 1
  hasMoreNotes.value = false
  
  // 4. 重置编辑器草稿
  // 防止把给 A 题写的回复草稿误发到了 B 题下面
  resetEditor()
  
  // 5. 更新收藏状态
  isFavorited.value = !!(props.question.is_favorite || props.question.IsFavorite)
}, { immediate: true })

/* 🚨 举报与纠错逻辑 */
const rpt = reactive({ show: false, noteId: 0, type: '营销广告', desc: '', loading: false })
const fb = reactive({ show: false, type: '答案错误', content: '', loading: false }) 
const rptTypes = ['言语辱骂', '虚假消息', '营销广告', '涉政有害', '违法违规', '垃圾消息','其他']
const fbTypes = ['答案错误', '错别字/排版', '解析错误', '图片无法显示', '其他']

const openRpt = (id: number) => { rpt.noteId = id; rpt.type = '营销广告'; rpt.desc = ''; rpt.show = true }
const submitRpt = async () => {
    rpt.loading = true
    try { await request.post(`/notes/${rpt.noteId}/report`, { reason: `${rpt.type}${rpt.desc ? '：'+rpt.desc : ''}` }); message.success('举报已提交'); rpt.show = false } catch (e: any) { message.error(e.response?.data?.error || '提交失败') } finally { rpt.loading = false }
}
const openFb = () => { fb.type = '答案错误'; fb.content = ''; fb.show = true }
const submitFb = async () => {
    if (!fb.content.trim()) return message.warning('请填写具体描述')
    fb.loading = true
    try { await request.post('/feedback', { question_id: props.question.id, type: fb.type, content: fb.content }); message.success('反馈成功，感谢纠错！'); fb.show = false } catch(e:any) { message.error('提交失败') } finally { fb.loading = false }
}

/* 🎮 交互动作 */
const handleLike = async (n: any) => {
  const old = n.is_liked; n.is_liked = !n.is_liked; n.like_count += n.is_liked ? 1 : -1
  try { await request.post(`/notes/${n.id}/like`) } catch { n.is_liked = old; n.like_count -= n.is_liked ? 1 : -1; message.error('失败') }
}
const handleCollect = async (n: any) => {
  const old = n.is_collected; n.is_collected = !n.is_collected
  try { await request.post(`/notes/${n.id}/collect`); message.success(n.is_collected ? '已收藏' : '已取消') } catch { n.is_collected = old; message.error('失败') }
}
const getOpts = (n: any) => isNoteOwner(n) 
  ? [{ label: '编辑', key: 'edit', icon: renderIcon(CreateOutline) }, { label: n.is_public ? '设私密' : '设公开', key: 'privacy', icon: renderIcon(n.is_public ? LockClosedOutline : GlobeOutline) }] 
  : [{ label: '举报', key: 'report', icon: renderIcon(WarningOutline), style: { color: '#d03050' } }]

const onAction = async (key: string, n: any) => {
  if (key === 'edit') {
    Object.assign(editorState, { id: n.id, content: n.content, isPublic: n.is_public, parentId: n.parent_id || 0, replyToUser: '', images: n.images ? [...n.images] : [] })
    nextTick(() => document.getElementById(`note-input-${props.question.id}`)?.focus())
  } else if (key === 'privacy') {
    try { await request.post('/notes', { ...n, is_public: !n.is_public, question_id: Number(props.question.id) }); message.success('已更新'); fetchNotes(false) } catch { message.error('失败') }
  } else if (key === 'delete') {
    try { await request.delete(`/notes/${n.id}`); message.success('已删除'); if (editorState.id === n.id) resetEditor(); fetchNotes(false) } catch { message.error('失败') }
  } else if (key === 'report') openRpt(n.id)
}
const reply = (n: any) => {
  Object.assign(editorState, { id: 0, content: '', isPublic: true, parentId: n.id, replyToUser: n.user?.username, images: [] })
  nextTick(() => document.getElementById(`note-input-${props.question.id}`)?.focus())
}

/* 🖼️ 渲染辅助 */
const getImages = (n: any) => n.images?.length ? n.images : [...n.content.matchAll(/\[图片:(.*?)\]/g)].map(m => m[1])
const cleanTxt = (s: string) => s ? s.replace(/\[图片:.*?\]/g, '').trim() : ''

const parseText = (text: string) => {
    if (!text) return ''
    return text.replace(/!\[(.*?)\]\((.*?)\)/g, (match, alt, url) => {
        const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`
        return `<img src="${fullUrl}" style="max-width:100%;vertical-align:bottom;border-radius:4px;margin:8px 0;cursor:pointer" onclick="window.open(this.src)" />`
    })
}

const parsedStem = computed(() => {
    let txt = props.question.stem ? props.question.stem.replace(/【(共用主干|共用题干|案例描述)】/g, '').trim() : ''
    return parseText(txt)
})

const hasChild = computed(() => props.question.children?.length > 0)
const opts = computed(() => { 
    if (!hasChild.value || !props.question.options) return null
    return typeof props.question.options === 'string' ? JSON.parse(props.question.options) : props.question.options 
})

const parsedB1Opts = computed(() => {
  const o = opts.value
  if (!o) return []
  return Object.keys(o).sort().map(k => ({
    key: k,
    parsedValue: parseText(o[k])
  }))
})

const tagType = (t: string) => !t ? 'default' : (t.includes('A3')||t.includes('A4')||t.includes('案例') ? 'info' : (t.includes('B1') ? 'warning' : 'success'))

onMounted(() => { if (props.initShowNotes) showNotes.value = true })
</script>

<template>
  <div :id="`question-anchor-${question.id}`" class="scroll-anchor">
    <n-card class="q-card" hoverable content-style="padding: 0;">
      
      <div v-if="hasChild" class="q-main-content">
          <div class="head" style="margin-bottom: 12px;">
             <n-tag :type="tagType(question.type)" size="small" strong round>{{ question.type || '题型' }}</n-tag>
             <span class="sub">(共 {{ question.children.length }} 小题)</span>
          </div>
          <div v-if="question.stem && !question.type.includes('B1')" class="ctx">
             <div class="ctx-h">（主题干）</div> 
             <div class="ctx-c" v-html="parsedStem"></div>
          </div>
          <div v-if="question.type.includes('B1') && parsedB1Opts.length" class="b1-ctx">
             <div class="b1-h">（共用备选答案）</div>
             <div class="b1-list">
                 <div v-for="t in parsedB1Opts" :key="t.key" class="b1-row">
                     <span class="b1-k">{{ t.key }}.</span> 
                     <span v-html="t.parsedValue"></span>
                 </div>
             </div>
          </div>
          <div class="child-list">
             <div v-for="(c, i) in question.children" :key="c.id" :id="`question-anchor-${c.id}`" class="scroll-anchor" style="position: relative;">
                <div class="idx-mark">题目 {{ c.displayIndex }}</div>
                <QuestionItem :question="c" :shared-options="opts" :index="(i as number) + 1" is-child :show-shared-header="false" @answer-result="emit('answer-result', $event)" />
                <n-divider v-if="(i as number) < question.children.length - 1" dashed style="margin: 0;" />
             </div>
          </div>
      </div>
      <div v-else class="q-main-content">
        <QuestionItem :question="question" :index="serialNumber" :is-child="false" :show-type-tag="true" @answer-result="emit('answer-result', $event)" />
      </div>

      <template #action>
        <div class="act-bar">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button text class="act-btn star" :class="{active: isFavorited}" @click="toggleFavorite">
                <template #icon><n-icon size="18" style="top:-1px"><Star v-if="isFavorited"/><StarOutline v-else/></n-icon></template> {{ isFavorited?'已收藏':'收藏' }}
              </n-button>
            </template>
            {{ isFavorited?'点击取消':'加入收藏' }}
          </n-tooltip>
          
          <n-button text class="act-btn" :class="{active: showNotes}" @click="showNotes = !showNotes">
            <template #icon><n-icon size="18"><ChatboxEllipses v-if="showNotes"/><ChatboxEllipsesOutline v-else/></n-icon></template> {{ showNotes?'收起':`讨论 (${displayCount})` }}
          </n-button>

          <n-tooltip trigger="hover">
            <template #trigger>
                <n-button text class="act-btn" @click="openFb">
                    <template #icon><n-icon size="18"><ConstructOutline/></n-icon></template> 纠错
                </n-button>
            </template>
            题目有误？反馈给管理员
          </n-tooltip>
        </div>
      </template>

      <n-collapse-transition :show="showNotes">
        <div class="notes-wrap">
          <div class="editor" :class="{ reply: editorState.parentId > 0 }">
             <div v-if="editorState.parentId" class="reply-tag">
                <span>回复 <strong style="color:#18a058">@{{ editorState.replyToUser }}</strong></span>
                <n-button size="tiny" circle quaternary @click="resetEditor"><template #icon><n-icon><CloseCircleOutline/></n-icon></template></n-button>
             </div>
             
             <div class="input-area">
                 <n-input 
                    v-model:value="editorState.content" 
                    type="textarea" 
                    :placeholder="editorState.parentId?'友善回复...':'输入评论... (支持粘贴图片)'" 
                    :autosize="{minRows:2,maxRows:8}" 
                    class="txt-area" 
                    :id="`note-input-${question.id}`" 
                    @paste="handlePaste"
                    maxlength="200" 
                 />
                 <span class="count-tip">{{ editorState.content.length }}/200</span>
             </div>

             <div v-if="editorState.images.length" class="prev-scroll">
                <div v-for="(url, i) in editorState.images" :key="i" class="prev-item">
                    <n-image :src="`http://localhost:8080${url}`" object-fit="cover" />
                    <div class="del-btn" @click.stop="editorState.images.splice(i,1)"><n-icon><CloseOutline/></n-icon></div>
                </div>
             </div>

             <div class="toolbar">
               <div class="left">
                  <n-upload :show-file-list="false" @change="handleUpload" accept="image/*" :disabled="editorState.images.length >= 5">
                      <n-tooltip trigger="hover">
                          <template #trigger>
                            <n-button size="small" circle secondary class="tool-btn" :disabled="editorState.images.length >= 5">
                                <template #icon><n-icon size="20"><AddOutline/></n-icon></template>
                            </n-button>
                          </template>
                          上传图片 (最多5张)
                      </n-tooltip>
                  </n-upload>

                  <n-popover trigger="click" :show="showEmojiPicker" @update:show="v=>showEmojiPicker=v" placement="bottom-start" :show-arrow="false" raw>
                      <template #trigger><n-button size="small" circle secondary class="tool-btn" style="margin-left:8px" @click="showEmojiPicker=!showEmojiPicker"><template #icon><n-icon size="20"><HappyOutline/></n-icon></template></n-button></template>
                      <div class="emojis"><span v-for="e in emojiList" :key="e" class="emoji" @click="insertEmoji(e)">{{e}}</span></div>
                  </n-popover>
               </div>
               <div class="right">
                  <div class="privacy-toggle">
                      <n-button size="small" round :type="editorState.isPublic ? 'primary' : 'default'" :secondary="!editorState.isPublic" @click="editorState.isPublic = true">
                          <template #icon><n-icon><GlobeOutline/></n-icon></template> 公开
                      </n-button>
                      <n-button size="small" round :type="!editorState.isPublic ? 'warning' : 'default'" :secondary="editorState.isPublic" @click="editorState.isPublic = false">
                          <template #icon><n-icon><LockClosedOutline/></n-icon></template> 私密
                      </n-button>
                  </div>
                  <n-button type="primary" round class="send" :loading="savingNote" :disabled="!editorState.content.trim() && !editorState.images.length" @click="saveNote">
                      <template #icon><n-icon><Send/></n-icon></template>
                  </n-button>
               </div>
             </div>
          </div>
          
          <div class="sort-bar" v-if="notesList.length > 0">
              <span class="sort-item" :class="{active: noteSort === 'hot'}" @click="noteSort !== 'hot' && toggleSort()">
                  <n-icon class="ico"><FlameOutline/></n-icon> 最热
              </span>
              <span class="divider">|</span>
              <span class="sort-item" :class="{active: noteSort === 'time'}" @click="noteSort !== 'time' && toggleSort()">
                  <n-icon class="ico"><TimeOutline/></n-icon> 最新
              </span>
          </div>

          <div class="list">
              <n-spin :show="loadingNotes">
                  <div v-if="!notesList.length" class="empty">
                      <n-icon size="40" color="#e0e0e0"><ChatboxEllipsesOutline/></n-icon><p>暂无评论，抢沙发！</p>
                  </div>
                  <div v-else class="feed">
                      <div class="item" v-for="n in notesList" :key="n.id">
                          <div class="avatar"><n-avatar round size="small" :src="n.user?.avatar ? `http://localhost:8080${n.user.avatar}` : undefined" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" style="border:1px solid #eee"/></div>
                          <div class="body">
                              <div class="head-row">
                                  <div class="info">
                                      <span class="name">{{ n.user?.nickname || n.user?.username }}</span>
                                      
                                      <span v-if="n.user?.school" class="school-badge">
                                          <n-icon size="12" style="margin-right:2px; vertical-align: -2px;"><SchoolOutline /></n-icon>
                                          {{ n.user.school }}
                                      </span>

                                      <n-tag v-if="isNoteOwner(n)" type="success" size="tiny" bordered style="transform:scale(0.9)">我</n-tag>
                                      <n-tag v-else-if="n.user?.role==='admin'" type="error" size="tiny" style="transform:scale(0.9)">老师</n-tag>
                                      <span class="time"><n-time :time="new Date(n.created_at)" type="relative"/></span>
                                      <div v-if="!n.is_public" class="priv" title="仅自己可见"><n-icon><LockClosedOutline/></n-icon></div>
                                  </div>
                              </div>
                              <div v-if="n.parent" class="quote">
                                  <div class="q-head">回复 <span class="q-user">@{{ n.parent.user?.nickname || n.parent.user?.username }}</span> :</div>
                                  <div class="q-txt">{{ cleanTxt(n.parent.content) }} <span v-if="getImages(n.parent).length" style="color:#18a058;margin-left:4px">[图片]</span></div>
                              </div>
                              <div class="txt">{{ cleanTxt(n.content) }}</div>
                              <div v-if="getImages(n).length" class="imgs">
                                 <n-image v-for="(u, i) in getImages(n)" :key="i" :src="`http://localhost:8080${u}`" class="img" object-fit="cover" />
                              </div>
                              
                              <div class="acts">
                                  <n-button text size="tiny" class="ibtn like" :class="{liked:n.is_liked}" @click="handleLike(n)"><template #icon><n-icon><ThumbsUp v-if="n.is_liked"/><ThumbsUpOutline v-else/></n-icon></template>{{ n.like_count||'' }}</n-button>
                                  <n-button text size="tiny" class="ibtn col" :class="{coled:n.is_collected}" @click="handleCollect(n)"><template #icon><n-icon><Star v-if="n.is_collected"/><StarOutline v-else/></n-icon></template></n-button>
                                  <n-button text size="tiny" class="ibtn rep" @click="reply(n)"><template #icon><n-icon><ChatbubbleOutline/></n-icon></template></n-button>
                                  <template v-if="canDelete(n)">
                                      <n-popconfirm @positive-click="onAction('delete', n)"><template #trigger><n-button text size="tiny" class="ibtn del"><template #icon><n-icon><TrashOutline/></n-icon></template></n-button></template>确定删除？</n-popconfirm>
                                  </template>
                                  <n-dropdown trigger="click" :options="getOpts(n)" @select="(k) => onAction(k, n)"><n-button text size="tiny" class="ibtn more"><template #icon><n-icon><EllipsisVertical/></n-icon></template></n-button></n-dropdown>
                              </div>
                          </div>
                      </div>

                      <div v-if="hasMoreNotes" class="load-more">
                          <n-button text type="primary" @click="fetchNotes(true)" :loading="loadingNotes">查看更多评论 <n-icon><EllipsisVertical/></n-icon></n-button>
                      </div>
                  </div>
              </n-spin>
          </div>
        </div>
      </n-collapse-transition>
    </n-card>

    <n-modal v-model:show="rpt.show" preset="dialog" title="🚨 举报违规">
        <div style="padding:10px 0">
            <div style="margin-bottom:8px;font-weight:bold;color:#666">理由：</div>
            <n-radio-group v-model:value="rpt.type"><n-space vertical><n-radio v-for="r in rptTypes" :key="r" :value="r">{{r}}</n-radio></n-space></n-radio-group>
            <div style="margin-top:16px;font-weight:bold;color:#666">说明：</div>
            <n-input v-model:value="rpt.desc" type="textarea" placeholder="细节描述..." :autosize="{minRows:2,maxRows:4}"/>
        </div>
        <template #action><n-button @click="rpt.show=false">取消</n-button><n-button type="error" :loading="rpt.loading" @click="submitRpt">提交</n-button></template>
    </n-modal>

    <n-modal v-model:show="fb.show" preset="dialog" title="🛠️ 题目纠错">
        <div style="padding:10px 0">
            <div style="margin-bottom:8px;font-weight:bold;color:#666">错误类型：</div>
            <n-radio-group v-model:value="fb.type"><n-space><n-radio v-for="t in fbTypes" :key="t" :value="t">{{t}}</n-radio></n-space></n-radio-group>
            <div style="margin-top:16px;margin-bottom:8px;font-weight:bold;color:#666">问题描述：</div>
            <n-input v-model:value="fb.content" type="textarea" placeholder="请详细描述，如果是答案错误，请提供依据..." :autosize="{minRows:3}"/>
        </div>
        <template #action><n-button @click="fb.show=false">取消</n-button><n-button type="primary" :loading="fb.loading" @click="submitFb">提交反馈</n-button></template>
    </n-modal>
  </div>
</template>

<style scoped>
/* 🌟 基础卡片美化 */
.scroll-anchor { scroll-margin-top: 80px; }

.q-card {
  margin-bottom: 24px;
  border-radius: 20px; 
  border: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  background: #fff;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.q-card:hover { 
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08); 
  transform: translateY(-2px);
}

.head { display: flex; align-items: center; gap: 12px; }
.sub { font-size: 13px; color: #94a3b8; font-weight: 500; }

/* 📚 题干区域 */
.ctx, .b1-ctx {
  background-color: #f8fafc;
  border-left: 4px solid #3b82f6;
  border-radius: 16px;
  padding: 20px 24px;
  margin-bottom: 24px;
  position: relative;
}
.b1-ctx { border-left-color: #10b981; background-color: #f0fdf4; }

.ctx-h, .b1-h { display: flex; align-items: center; gap: 6px; margin-bottom: 12px; font-size: 14px; font-weight: 700; color: #3b82f6; letter-spacing: 0.5px; }
.b1-h { color: #10b981; }

.ctx-c { font-size: 16px; color: #334155; line-height: 1.75; font-weight: 500; }

.b1-list { display: flex; flex-direction: column; gap: 10px; }
.b1-row { display: flex; align-items: baseline; font-size: 15px; color: #475569; line-height: 1.6; }
.b1-k { font-weight: 700; color: #10b981; margin-right: 8px; min-width: 24px; }

/* 👶 子题目列表 */
.child-list { display: flex; flex-direction: column; }
.idx-mark { display: inline-flex; align-items: center; justify-content: center; font-size: 12px; color: #fff; background: linear-gradient(135deg, #60a5fa, #3b82f6); padding: 4px 12px; border-radius: 20px; margin-bottom: 12px; font-weight: 600; box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2); user-select: none; }

/* 🎮 操作栏 */
.act-bar { display: flex; align-items: center; gap: 8px; padding-top: 0; }
.act-btn { font-size: 14px; color: #64748b; font-weight: 500; padding: 8px 16px; border-radius: 12px; transition: all 0.2s ease; }
.act-btn:hover { background-color: #f1f5f9; color: #334155; }
.act-btn.active { color: #3b82f6; background-color: #eff6ff; font-weight: 600; }
.act-btn.star.active { color: #f59e0b !important; background-color: #fffbeb !important; }

/* 📝 评论区容器 */
.notes-wrap { background-color: #f8fafc; border-top: 1px solid #e2e8f0; padding: 24px 32px; border-bottom-left-radius: 20px; border-bottom-right-radius: 20px; }

/* ✏️ 编辑器优化 */
.editor { background: #fff; border: 1px solid #e2e8f0; border-radius: 16px; padding: 16px; transition: all 0.2s ease; box-shadow: 0 1px 2px rgba(0,0,0,0.02); }
.editor:focus-within { border-color: #94a3b8; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
.editor.reply { background-color: #fcfcfc; border-color: #10b981; }

.reply-tag { display: flex; justify-content: space-between; align-items: center; font-size: 13px; color: #64748b; background-color: #f0fdf4; padding: 6px 12px; border-radius: 8px; margin-bottom: 12px; border: 1px dashed #86efac; }
.input-area { position: relative; }
.txt-area { background: transparent; padding: 0; font-size: 15px; line-height: 1.6; color: #334155; --n-border: none !important; --n-box-shadow-focus: none !important; padding-bottom: 20px; }
.count-tip { position: absolute; bottom: 8px; right: 12px; font-size: 12px; color: #cbd5e1; pointer-events: none; }

.toolbar { margin-top: 12px; display: flex; justify-content: space-between; align-items: center; }
.left { display: flex; gap: 8px; }
.right { display: flex; align-items: center; gap: 12px; }
.tool-btn { color: #64748b; transition: all 0.2s; }
.tool-btn:hover { background-color: #e2e8f0; color: #0f172a; }
.privacy-toggle { display: flex; gap: 8px; margin-right: 12px; }

/* 💬 评论列表美化 */
.list { margin-top: 24px; }
.sort-bar { display: flex; align-items: center; justify-content: flex-end; margin-bottom: 16px; padding-bottom: 8px; border-bottom: 1px solid #eee; }
.sort-item { font-size: 13px; color: #64748b; cursor: pointer; display: flex; align-items: center; gap: 4px; padding: 4px 8px; border-radius: 8px; transition: all 0.2s; }
.sort-item:hover { background-color: #f1f5f9; color: #334155; }
.sort-item.active { color: #3b82f6; font-weight: 600; background-color: #eff6ff; }
.divider { margin: 0 6px; color: #e2e8f0; }

.feed { display: flex; flex-direction: column; gap: 20px; }
.item { display: flex; gap: 16px; align-items: flex-start; }
.avatar { flex-shrink: 0; padding-top: 4px; }

.body { flex: 1; background: #fff; border: 1px solid #f1f5f9; border-radius: 16px; padding: 20px 24px; position: relative; box-shadow: 0 1px 3px rgba(0,0,0,0.02); transition: all 0.2s ease; }
.body:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.05); border-color: #e2e8f0; }
.body::before { content: ''; position: absolute; left: -6px; top: 20px; width: 10px; height: 10px; background: #fff; border-left: 1px solid #f1f5f9; border-bottom: 1px solid #f1f5f9; transform: rotate(45deg); border-radius: 0 0 0 3px; }
.body:hover::before { border-color: #e2e8f0; }

.head-row { display: flex; align-items: center; margin-bottom: 8px; width: 100%; }
.info { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.name { font-weight: 600; color: #1e293b; font-size: 14px; }

/* 🔥 学校徽章样式设计 */
.school-badge {
    font-size: 11px;
    color: #0ea5e9; /* 清新的天蓝色 */
    background-color: #e0f2fe;
    padding: 2px 8px;
    border-radius: 100px;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
}

.time { color: #94a3b8; font-size: 12px; margin-left: 4px; }
.priv { color: #f59e0b; font-size: 14px; display: flex; align-items: center; }

.acts { display: flex; align-items: center; justify-content: flex-end; gap: 12px; margin-top: 12px; }
.ibtn { font-size: 14px; color: #94a3b8; padding: 4px 8px; border-radius: 8px; transition: all 0.2s; display: flex; align-items: center; gap: 4px; }
.ibtn:hover { background-color: #f1f5f9; color: #475569; }
.ibtn.like.liked { color: #f43f5e; } 
.ibtn.col.coled { color: #f59e0b; }
.ibtn.rep:hover { color: #3b82f6; }
.ibtn.del:hover { color: #ef4444; background-color: #fef2f2; }

.txt { font-size: 15px; color: #334155; line-height: 1.6; white-space: pre-wrap; word-break: break-all; margin-top: 4px; }
.quote { background-color: #f8fafc; border-left: 3px solid #cbd5e1; padding: 12px 16px; margin-bottom: 12px; border-radius: 8px; font-size: 13px; color: #64748b; }
.q-user { color: #3b82f6; font-weight: 600; margin: 0 4px; }

.imgs { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
.img { width: 100px; height: 100px; border-radius: 12px; overflow: hidden; border: 1px solid #e2e8f0; cursor: zoom-in; object-fit: cover; }
.empty { text-align: center; padding: 40px 0; color: #cbd5e1; display: flex; flex-direction: column; align-items: center; gap: 12px; font-size: 14px; }

.prev-scroll { display: flex; gap: 12px; overflow-x: auto; padding: 12px 0; margin-top: 8px; }
.prev-item { position: relative; width: 80px; height: 80px; flex-shrink: 0; border-radius: 12px; overflow: hidden; border: 1px solid #e2e8f0; }
.del-btn { position: absolute; top: 4px; right: 4px; background: rgba(0,0,0,0.5); color: #fff; width: 20px; height: 20px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; cursor: pointer; transition: 0.2s; }
.del-btn:hover { background: rgba(0,0,0,0.7); }

.emojis { display: grid; grid-template-columns: repeat(8, 1fr); gap: 4px; padding: 12px; background: #fff; border: 1px solid #e2e8f0; border-radius: 16px; box-shadow: 0 10px 25px rgba(0,0,0,0.1); max-height: 200px; overflow-y: auto; width: 300px; }
.emoji { font-size: 22px; cursor: pointer; padding: 6px; text-align: center; border-radius: 8px; transition: 0.2s; }
.emoji:hover { background-color: #f1f5f9; transform: scale(1.1); }

.load-more { text-align: center; margin-top: 24px; padding-bottom: 12px; }
.q-main-content { padding: 20px 24px; }

@media (max-width: 768px) {
    .scroll-anchor { scroll-margin-top: 60px; }
    .q-main-content { padding: 16px; }
    .q-card { margin-bottom: 16px; border-radius: 12px; }
    .ctx, .b1-ctx { padding: 16px; border-left-width: 3px; }
    .notes-wrap { padding: 16px; }
    .list { margin-top: 16px; }
    .item { gap: 10px; }
    .avatar { padding-top: 0; }
    .body { padding: 12px; border-radius: 12px; }
    .body::before { display: none; }
    .img { width: 80px; height: 80px; }
    .emojis { width: 100%; max-width: 300px; }
    .act-bar { flex-wrap: wrap; }
    .toolbar { flex-direction: column; align-items: flex-start; gap: 12px; }
    .toolbar .left, .toolbar .right { width: 100%; justify-content: space-between; }
    .toolbar .right { margin-top: 4px; }
    .privacy-toggle { flex: 1; justify-content: flex-start; } 
}
</style>