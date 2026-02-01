<script setup lang="ts">
import { ref, onMounted, computed, reactive, nextTick, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  NCard, NButton, NIcon, NSpin, NTag, useMessage, NEmpty, NInput, NAvatar, 
  NDivider, NImage, NPopover, NUpload, NTooltip, NTime, NPopconfirm, NDropdown,
  NModal, NRadioGroup, NRadio, NSpace
} from 'naive-ui'
import { 
  ArrowBackOutline, ChatboxEllipsesOutline, Send, AddOutline, HappyOutline,
  GlobeOutline, LockClosedOutline, CloseOutline, CloseCircleOutline,
  FlameOutline, TimeOutline, ThumbsUpOutline, ThumbsUp, StarOutline, Star,
  ChatbubbleOutline, TrashOutline, EllipsisVertical, WarningOutline, CreateOutline
} from '@vicons/ionicons5'
import request from '../utils/request'
import { useUserStore } from '../stores/user'
import QuestionItem from '../components/QuestionItem.vue' 

const route = useRoute()
const router = useRouter()
const message = useMessage()
const userStore = useUserStore()

const questionId = route.params.id as string
const loading = ref(true)
const question = ref<any>(null)

// =======================
// 1. 获取题目详情
// =======================
const fetchDetail = async () => {
  loading.value = true
  try {
    const res: any = await request.get(`/questions/${questionId}`)
    question.value = res.data
    // 加载评论
    fetchNotes()
  } catch (e) {
    message.error('题目加载失败')
  } finally {
    loading.value = false
  }
}

// =======================
// 2. 题目渲染辅助逻辑 (🔥 参考 QuestionCard)
// =======================
const hasChild = computed(() => question.value?.children?.length > 0)

// 解析 B1 题型的共用选项
const sharedOpts = computed(() => {
    if (!hasChild.value || !question.value.options) return null
    return typeof question.value.options === 'string' ? JSON.parse(question.value.options) : question.value.options
})

// 解析文本 (Markdown图片等)
const parseText = (text: string) => {
    if (!text) return ''
    return text.replace(/!\[(.*?)\]\((.*?)\)/g, (match, alt, url) => {
        const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`
        return `<img src="${fullUrl}" style="max-width:100%;vertical-align:bottom;border-radius:4px;margin:8px 0;cursor:pointer" onclick="window.open(this.src)" />`
    })
}

// 格式化题干 (去除【共用题干】等标记)
const fmtStem = (s: string) => {
    let txt = s ? s.replace(/【(共用主干|共用题干|案例描述)】/g, '').trim() : ''
    return parseText(txt)
}

const tagType = (t: string) => !t ? 'default' : (t.includes('A3')||t.includes('A4')||t.includes('案例') ? 'info' : (t.includes('B1') ? 'warning' : 'success'))


// =======================
// 3. 评论区核心逻辑
// =======================
const notesList = ref<any[]>([])
const loadingNotes = ref(false)
const notePage = ref(1)
const hasMoreNotes = ref(false)
const noteSort = ref('hot') 

const fetchNotes = async (isLoadMore = false) => {
  if (loadingNotes.value) return 
  loadingNotes.value = true
  if (!isLoadMore) { notePage.value = 1; notesList.value = [] }

  try {
    const res: any = await request.get('/notes', { 
        params: { question_id: parseInt(questionId), page: notePage.value, page_size: 10, sort: noteSort.value } 
    })
    if (isLoadMore) { notesList.value = [...notesList.value, ...res.data] } else { notesList.value = res.data || [] }
    hasMoreNotes.value = res.has_more || false
    if (hasMoreNotes.value) notePage.value++
  } catch (e) { console.error(e) } finally { loadingNotes.value = false }
}

const toggleSort = () => { noteSort.value = noteSort.value === 'hot' ? 'time' : 'hot'; fetchNotes(false) }

// 编辑器状态
const editorState = reactive({ id: 0, content: '', isPublic: true, parentId: 0, replyToUser: '', images: [] as string[] })
const savingNote = ref(false)
const showEmojiPicker = ref(false)
const emojiList = ['😀','😃','😄','😁','😆','😅','😂','🤣','😊','😇','🙂','🙃','😉','😌','😍','🥰','😘','😗','😙','😚','😋','😛','😝','😜','🤪','🤨','🧐','🤓','😎','🤩','🥳','😏','😒','😞','😔','😟','😕','🙁','☹️','😣','😖','😫','😩','🥺','😢','😭','😤','😠','😡','🤬','🤯','😳','🥵','🥶','😱','😨','😰','😥','😓','🤗','🤔','🤭','🤫','🤥','😶','😐','😑','😬','🙄','😯','😦','😧','😮','😲','😴','🤤','😪','😵','🤐','🥴','🤢','🤮','🤧','😷','🤒','🤕','🤑','🤠','😈','👿','👹','👺','🤡','💩','👻','💀','☠️','👽','👾','🤖','🎃','😺','😸','😹','😻','😼','😽','🙀','😿','😾']

const resetEditor = () => Object.assign(editorState, { id: 0, content: '', isPublic: true, parentId: 0, replyToUser: '', images: [] })
const insertEmoji = (emoji: string) => {
  if (editorState.content.length + emoji.length > 200) return message.warning('字数已达上限')
  editorState.content += emoji; showEmojiPicker.value = false
}
const handleUpload = async ({ file }: { file: any }) => {
  if (editorState.images.length >= 5) return message.warning('最多上传5张图片')
  const form = new FormData(); form.append('file', file.file)
  try {
    const res: any = await request.post('/notes/upload', form, { headers: { 'Content-Type': 'multipart/form-data' } })
    editorState.images.push(res.url)
  } catch (e: any) { message.error('上传失败') }
}
const saveNote = async () => {
  if (!editorState.content.trim() && !editorState.images.length) return message.warning('内容不能为空')
  if (editorState.content.length > 200) return message.warning('字数不能超过200字')
  savingNote.value = true
  try {
    await request.post('/notes', {
      id: editorState.id, question_id: parseInt(questionId), content: editorState.content,
      is_public: editorState.isPublic, parent_id: editorState.parentId || null, images: editorState.images 
    })
    message.success(editorState.id ? '修改成功' : '发布成功'); resetEditor(); fetchNotes(false)
  } catch (e: any) { message.error(e.response?.data?.error || '操作失败') } finally { savingNote.value = false }
}

const renderIcon = (icon: any) => () => h(NIcon, null, { default: () => h(icon) })
const isNoteOwner = (note: any) => { const myId = String(userStore.id || ''), uid = String(note.user_id || note.user?.id || ''); return myId === uid }
const canDelete = (note: any) => isNoteOwner(note) || userStore.role === 'admin'
const handleLike = async (n: any) => { const old = n.is_liked; n.is_liked = !n.is_liked; n.like_count += n.is_liked ? 1 : -1; try { await request.post(`/notes/${n.id}/like`) } catch { n.is_liked = old; n.like_count -= n.is_liked ? 1 : -1; message.error('失败') } }
const handleCollect = async (n: any) => { const old = n.is_collected; n.is_collected = !n.is_collected; try { await request.post(`/notes/${n.id}/collect`); message.success(n.is_collected ? '已收藏' : '已取消') } catch { n.is_collected = old; message.error('失败') } }
const reply = (n: any) => { Object.assign(editorState, { id: 0, content: '', isPublic: true, parentId: n.id, replyToUser: n.user?.nickname || n.user?.username, images: [] }); nextTick(() => document.querySelector('.input-area')?.scrollIntoView({ behavior: 'smooth' })) }
const getOpts = (n: any) => isNoteOwner(n) ? [{ label: '编辑', key: 'edit', icon: renderIcon(CreateOutline) }, { label: n.is_public ? '设私密' : '设公开', key: 'privacy', icon: renderIcon(n.is_public ? LockClosedOutline : GlobeOutline) }] : [{ label: '举报', key: 'report', icon: renderIcon(WarningOutline), style: { color: '#d03050' } }]
const onAction = async (key: string, n: any) => {
  if (key === 'edit') { Object.assign(editorState, { id: n.id, content: n.content, isPublic: n.is_public, parentId: n.parent_id || 0, replyToUser: '', images: n.images ? [...n.images] : [] }) }
  else if (key === 'privacy') { try { await request.post('/notes', { ...n, is_public: !n.is_public, question_id: parseInt(questionId) }); message.success('已更新'); fetchNotes(false) } catch { message.error('失败') } }
  else if (key === 'delete') { try { await request.delete(`/notes/${n.id}`); message.success('已删除'); if (editorState.id === n.id) resetEditor(); fetchNotes(false) } catch { message.error('失败') } }
  else if (key === 'report') openRpt(n.id)
}

const rpt = reactive({ show: false, noteId: 0, type: '营销广告', desc: '', loading: false })
const rptTypes = ['言语辱骂', '虚假消息', '营销广告', '涉政有害', '违法违规', '垃圾消息','其他']
const openRpt = (id: number) => { rpt.noteId = id; rpt.type = '营销广告'; rpt.desc = ''; rpt.show = true }
const submitRpt = async () => { rpt.loading = true; try { await request.post(`/notes/${rpt.noteId}/report`, { reason: `${rpt.type}${rpt.desc ? '：'+rpt.desc : ''}` }); message.success('举报已提交'); rpt.show = false } catch (e: any) { message.error(e.response?.data?.error || '提交失败') } finally { rpt.loading = false } }

const fb = reactive({ show: false, type: '答案错误', content: '', loading: false })
const fbTypes = ['答案错误', '错别字/排版', '解析错误', '图片无法显示', '其他']
const openFb = () => { fb.type = '答案错误'; fb.content = ''; fb.show = true }
const submitFb = async () => { if (!fb.content.trim()) return message.warning('请填写具体描述'); fb.loading = true; try { await request.post('/feedback', { question_id: parseInt(questionId), type: fb.type, content: fb.content }); message.success('反馈成功'); fb.show = false } catch(e:any) { message.error('提交失败') } finally { fb.loading = false } }

const getImages = (n: any) => n.images?.length ? n.images : [...n.content.matchAll(/\[图片:(.*?)\]/g)].map(m => m[1])
const cleanTxt = (s: string) => s ? s.replace(/\[图片:.*?\]/g, '').trim() : ''

onMounted(() => {
  fetchDetail()
})
</script>

<template>
  <div class="detail-page">
    
    <div class="page-header">
      <n-button text style="font-size: 16px;" @click="router.back()">
        <template #icon><n-icon><ArrowBackOutline /></n-icon></template>
        返回
      </n-button>
      <span class="header-title">题目详情 & 讨论</span>
    </div>

    <div v-if="loading" class="loading-box">
      <n-spin size="large" />
    </div>

    <div v-else-if="question" class="content-grid">
      
      <div class="left-col">
        <n-card :bordered="false" class="q-wrapper" content-style="padding-bottom: 24px;">
            <template #header>
               <div class="head">
                  <n-tag :type="tagType(question.type)" size="small" strong round>{{ question.type || '题型' }}</n-tag>
                  <span v-if="hasChild" class="sub">(共 {{ question.children.length }} 小题)</span>
                  
                  <div style="flex:1"></div>
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

            <div v-if="hasChild">
                <div v-if="question.stem && !question.type.includes('B1')" class="ctx">
                   <div class="ctx-h">（主题干）</div> 
                   <div class="ctx-c" v-html="fmtStem(question.stem)"></div>
                </div>
                
                <div v-if="question.type.includes('B1') && sharedOpts" class="b1-ctx">
                   <div class="b1-h">（共用备选答案）</div>
                   <div class="b1-list">
                       <div v-for="(t, k) in sharedOpts" :key="k" class="b1-row">
                           <span class="b1-k">{{ k }}.</span> 
                           <span v-html="parseText(t)"></span>
                       </div>
                   </div>
                </div>

                <div class="child-list">
                   <div v-for="(child, i) in question.children" :key="child.id" class="child-item">
                      <div class="idx-mark">Topic {{ child.displayIndex || (i+1) }}</div>
                      <QuestionItem 
                          :question="child" 
                          :shared-options="sharedOpts" 
                          :index="i+1" 
                          is-child 
                          :show-shared-header="false" 
                          :show-analysis="true"
                          :readonly="true"
                      />
                      <n-divider v-if="i < question.children.length - 1" dashed style="margin: 20px 0;" />
                   </div>
                </div>
            </div>

            <div v-else>
                <QuestionItem 
                    :question="question" 
                    :index="1" 
                    :is-child="false" 
                    :show-analysis="true"
                    :readonly="true"
                />
            </div>
        </n-card>
      </div>

      <div class="right-col">
        <n-card :bordered="false" class="discuss-card" title="学习笔记 & 讨论" content-style="padding: 0; display: flex; flex-direction: column; height: 100%;">
           <template #header-extra>
              <div class="sort-bar">
                  <span class="sort-item" :class="{active: noteSort === 'hot'}" @click="toggleSort">
                      <n-icon class="ico"><FlameOutline/></n-icon> 最热
                  </span>
                  <span class="divider">|</span>
                  <span class="sort-item" :class="{active: noteSort === 'time'}" @click="toggleSort">
                      <n-icon class="ico"><TimeOutline/></n-icon> 最新
                  </span>
              </div>
           </template>

           <div class="notes-scroll">
              <n-spin :show="loadingNotes">
                  <div v-if="!notesList.length" class="empty-state">
                      <n-icon size="48" color="#e0e0e0"><ChatboxEllipsesOutline/></n-icon>
                      <p>暂无笔记，抢沙发！</p>
                  </div>
                  <div v-else class="feed">
                      <div class="item" v-for="n in notesList" :key="n.id">
                          <div class="avatar"><n-avatar round size="small" :src="n.user?.avatar ? `http://localhost:8080${n.user.avatar}` : undefined" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"/></div>
                          <div class="body">
                              <div class="head-row">
                                  <div class="info">
                                      <span class="name">{{ n.user?.nickname || n.user?.username }}</span>
                                      <n-tag v-if="isNoteOwner(n)" type="success" size="tiny" bordered style="transform:scale(0.9)">我</n-tag>
                                      <span class="time"><n-time :time="new Date(n.created_at)" type="relative"/></span>
                                      <div v-if="!n.is_public" class="priv"><n-icon><LockClosedOutline/></n-icon></div>
                                  </div>
                                  <div class="acts">
                                      <n-button text size="tiny" class="ibtn like" :class="{liked:n.is_liked}" @click="handleLike(n)"><template #icon><n-icon><ThumbsUp v-if="n.is_liked"/><ThumbsUpOutline v-else/></n-icon></template>{{ n.like_count||'' }}</n-button>
                                      <n-button text size="tiny" class="ibtn rep" @click="reply(n)"><template #icon><n-icon><ChatbubbleOutline/></n-icon></template></n-button>
                                      <template v-if="canDelete(n)">
                                          <n-popconfirm @positive-click="onAction('delete', n)"><template #trigger><n-button text size="tiny" class="ibtn del"><template #icon><n-icon><TrashOutline/></n-icon></template></n-button></template>删除</n-popconfirm>
                                      </template>
                                      <n-dropdown trigger="click" :options="getOpts(n)" @select="(k) => onAction(k, n)"><n-button text size="tiny" class="ibtn more"><template #icon><n-icon><EllipsisVertical/></n-icon></template></n-button></n-dropdown>
                                  </div>
                              </div>
                              <div v-if="n.parent" class="quote">
                                  <div class="q-head">回复 <span class="q-user">@{{ n.parent.user?.nickname || n.parent.user?.username }}</span> :</div>
                                  <div class="q-txt">{{ cleanTxt(n.parent.content) }}</div>
                              </div>
                              <div class="txt">{{ cleanTxt(n.content) }}</div>
                              <div v-if="getImages(n).length" class="imgs">
                                 <n-image v-for="(u, i) in getImages(n)" :key="i" :src="`http://localhost:8080${u}`" class="img" object-fit="cover" />
                              </div>
                          </div>
                      </div>
                      <div v-if="hasMoreNotes" class="load-more">
                          <n-button text type="primary" @click="fetchNotes(true)" :loading="loadingNotes">更多评论</n-button>
                      </div>
                  </div>
              </n-spin>
           </div>

           <div class="input-wrapper">
              <div v-if="editorState.parentId" class="reply-tag">
                 <span>回复 <strong style="color:#18a058">@{{ editorState.replyToUser }}</strong></span>
                 <n-button size="tiny" circle quaternary @click="resetEditor"><template #icon><n-icon><CloseCircleOutline/></n-icon></template></n-button>
              </div>
              
              <div class="input-area">
                 <n-input 
                    v-model:value="editorState.content" 
                    type="textarea" 
                    :placeholder="editorState.parentId ? '友善回复...' : '写下你的解题思路...'" 
                    :autosize="{minRows: 2, maxRows: 6}" 
                    maxlength="200" 
                    show-count
                    class="editor-input"
                 />
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
                        <n-tooltip trigger="hover"><template #trigger>
                            <n-button size="small" circle secondary class="tool-btn" :disabled="editorState.images.length >= 5">
                                <template #icon><n-icon size="20"><AddOutline/></n-icon></template>
                            </n-button>
                        </template>图片</n-tooltip>
                    </n-upload>
                    <n-popover trigger="click" :show="showEmojiPicker" @update:show="v=>showEmojiPicker=v" placement="top-start" :show-arrow="false" raw>
                        <template #trigger><n-button size="small" circle secondary class="tool-btn" @click="showEmojiPicker=!showEmojiPicker"><template #icon><n-icon size="20"><HappyOutline/></n-icon></template></n-button></template>
                        <div class="emojis"><span v-for="e in emojiList" :key="e" class="emoji" @click="insertEmoji(e)">{{e}}</span></div>
                    </n-popover>
                 </div>
                 <div class="right">
                    <n-switch v-model:value="editorState.isPublic" size="small">
                        <template #checked-icon><n-icon><GlobeOutline/></n-icon></template>
                        <template #unchecked-icon><n-icon><LockClosedOutline/></n-icon></template>
                    </n-switch>
                    <n-button type="primary" round class="send-btn" :loading="savingNote" @click="saveNote">
                        <template #icon><n-icon><Send/></n-icon></template> 发送
                    </n-button>
                 </div>
              </div>
           </div>
        </n-card>
      </div>

    </div>

    <n-modal v-model:show="rpt.show" preset="dialog" title="🚨 举报违规">
        <div style="padding:10px 0">
            <n-radio-group v-model:value="rpt.type"><n-space vertical><n-radio v-for="r in rptTypes" :key="r" :value="r">{{r}}</n-radio></n-space></n-radio-group>
            <n-input v-model:value="rpt.desc" type="textarea" placeholder="细节描述..." style="margin-top:10px"/>
        </div>
        <template #action><n-button @click="rpt.show=false">取消</n-button><n-button type="error" :loading="rpt.loading" @click="submitRpt">提交</n-button></template>
    </n-modal>

    <n-modal v-model:show="fb.show" preset="dialog" title="🛠️ 题目纠错">
        <div style="padding:10px 0">
            <n-radio-group v-model:value="fb.type"><n-space><n-radio v-for="t in fbTypes" :key="t" :value="t">{{t}}</n-radio></n-space></n-radio-group>
            <n-input v-model:value="fb.content" type="textarea" placeholder="请详细描述..." style="margin-top:10px"/>
        </div>
        <template #action><n-button @click="fb.show=false">取消</n-button><n-button type="primary" :loading="fb.loading" @click="submitFb">提交</n-button></template>
    </n-modal>
  </div>
</template>

<style scoped>
.detail-page { padding: 24px; max-width: 1400px; margin: 0 auto; height: 100vh; display: flex; flex-direction: column; box-sizing: border-box; }
.page-header { flex-shrink: 0; margin-bottom: 16px; display: flex; align-items: center; gap: 16px; }
.header-title { font-size: 18px; font-weight: 700; color: #1f2937; }
.loading-box { flex: 1; display: flex; align-items: center; justify-content: center; }

/* 布局 */
.content-grid { flex: 1; display: grid; grid-template-columns: 1fr 400px; gap: 24px; overflow: hidden; }
.left-col { overflow-y: auto; padding-right: 8px; }
.q-wrapper { border-radius: 12px; height: auto; }

/* 右侧评论区 */
.right-col { height: 100%; }
.discuss-card { height: 100%; display: flex; flex-direction: column; border-radius: 12px; border: 1px solid #f3f4f6; box-shadow: 0 4px 12px rgba(0,0,0,0.03); }
.notes-scroll { flex: 1; overflow-y: auto; padding: 16px; background: #f9fafb; }

/* 评论列表项样式 (保持与 QuestionCard 一致) */
.item { display: flex; gap: 12px; margin-bottom: 16px; }
.avatar { flex-shrink: 0; padding-top: 2px; }
.body { flex: 1; background: #fff; border: 1px solid #eee; border-radius: 8px; padding: 10px 14px; position: relative; }
.body::before { content:''; position: absolute; left: -6px; top: 12px; width: 10px; height: 10px; background: #fff; border-left: 1px solid #eee; border-bottom: 1px solid #eee; transform: rotate(45deg); }
.head-row { display: flex; justify-content: space-between; margin-bottom: 8px; }
.info { display: flex; align-items: center; gap: 8px; font-size: 12px; }
.name { font-weight: bold; color: #333; }
.time { color: #ccc; }
.priv { color: #f0a020; }
.acts { display: flex; gap: 8px; }
.ibtn { font-size: 16px; color: #999; padding: 2px; }
.ibtn:hover { color: #333; }
.ibtn.like.liked { color: #d03050; }
.quote { background: #f7f9fc; border-left: 3px solid #d0d0d0; padding: 8px 12px; margin-bottom: 10px; border-radius: 0 4px 4px 0; font-size: 12px; color: #666; }
.q-user { color: #18a058; font-weight: bold; }
.txt { font-size: 14px; color: #444; line-height: 1.6; white-space: pre-wrap; word-break: break-all; }
.imgs { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; }
.img { width: 80px; height: 80px; border-radius: 6px; border: 1px solid #eee; }

.empty-state { text-align: center; padding: 40px 0; color: #aaa; display: flex; flex-direction: column; align-items: center; gap: 10px; }
.load-more { text-align: center; margin-top: 10px; }

/* 底部输入框 */
.input-wrapper { padding: 12px; background: #fff; border-top: 1px solid #e5e7eb; }
.reply-tag { display: flex; justify-content: space-between; align-items: center; background: #eff6ff; color: #2563eb; padding: 4px 8px; font-size: 12px; border-radius: 4px; margin-bottom: 8px; }
.editor-input { background: #f9fafb; }
.toolbar { margin-top: 8px; display: flex; justify-content: space-between; align-items: center; }
.left { display: flex; gap: 8px; }
.right { display: flex; align-items: center; gap: 12px; }
.tool-btn { color: #666; }
.send-btn { width: 80px; }

.prev-scroll { display: flex; gap: 8px; overflow-x: auto; padding: 8px 0; }
.prev-item { position: relative; width: 60px; height: 60px; flex-shrink: 0; border-radius: 6px; overflow: hidden; border: 1px solid #eee; }
.del-btn { position: absolute; top: 2px; right: 2px; background: rgba(0,0,0,0.6); color: #fff; border-radius: 50%; width: 16px; height: 16px; display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 12px; }

.sort-bar { display: flex; gap: 8px; font-size: 12px; color: #999; cursor: pointer; }
.sort-item.active { color: #2080f0; font-weight: bold; }
.divider { color: #eee; }

.emojis { display: grid; grid-template-columns: repeat(8, 1fr); gap: 4px; padding: 8px; max-height: 200px; overflow-y: auto; width: 280px; }
.emoji { font-size: 20px; cursor: pointer; padding: 4px; border-radius: 4px; text-align: center; }
.emoji:hover { background: #f0f0f0; }

/* 题目样式复刻 */
.head { display:flex; align-items:center; gap:10px; padding-bottom:12px; border-bottom:1px solid #eee; margin-bottom:16px; }
.sub { font-size:12px; color:#999; }
.ctx, .b1-ctx { background:#fdfdfd; border:1px solid #e0e0e0; border-left:4px solid #2080f0; border-radius:4px; padding:12px 16px; margin-bottom:20px; }
.b1-ctx { border-left-color:#18a058; background:#fcfcfc; }
.ctx-h, .b1-h { display:flex; gap:8px; margin-bottom:8px; font-size:14px; font-weight:bold; color:#18a058; }
.ctx-c { font-size:15px; color:#444; line-height:1.6; }
.child-list { display:flex; flex-direction:column; }
.b1-list { display:flex; flex-direction:column; gap:8px; }
.b1-row { font-size:14px; color:#333; line-height:1.5; }
.b1-k { font-weight:bold; color:#18a058; margin-right:6px; }
.idx-mark { display:inline-block; font-size:12px; color:#2080f0; background:#eef6fc; padding:2px 8px; border-radius:4px; margin-bottom:6px; font-weight:500; user-select:none; }
.act-btn { font-size:14px; color:#666; transition:all .2s; }
.act-btn:hover { color:#2080f0; }

@media (max-width: 900px) {
  .detail-page { height: auto; overflow: auto; }
  .content-grid { grid-template-columns: 1fr; overflow: visible; }
  .left-col { overflow: visible; padding-right: 0; margin-bottom: 20px; }
  .discuss-card { height: 600px; }
}
</style>