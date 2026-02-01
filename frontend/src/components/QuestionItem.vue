<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NButton, useMessage, NTag, NIcon, NModal, NDivider } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import request from '../utils/request'

const props = defineProps<{
  question: any       
  sharedOptions?: any 
  index?: number    
  isChild?: boolean 
  showSharedHeader?: boolean 
  showTypeTag?: boolean
}>()

const emit = defineEmits(['answer-result'])
const message = useMessage()

const selectedOption = ref('')    
const multiSelection = ref<string[]>([]) 
const result = ref<any>(null)     
const submitting = ref(false)     
const showAnswer = ref(false)
const showPreview = ref(false)
const previewImageUrl = ref('')

// 判断多选
const isMultiChoice = computed(() => {
  const t = (props.question.type || '').toUpperCase()
  return t.includes('X') || t.includes('多选')
})

const initData = () => {
  if (props.question.user_record) {
    const record = props.question.user_record
    if (isMultiChoice.value && record.choice) {
      multiSelection.value = record.choice.split('')
      selectedOption.value = record.choice 
    } else {
      selectedOption.value = record.choice || ''
    }

    if (typeof record.is_correct === 'boolean') {
      result.value = {
        is_correct: record.is_correct,
        correct_answer: props.question.correct, 
        analysis: props.question.analysis
      }
    }
  } else {
    selectedOption.value = ''
    multiSelection.value = []
    result.value = null
  }
}

watch(() => props.question, () => { initData() }, { immediate: true })

// 判断主观题
const isSubjective = computed(() => {
  const t = (props.question.type || '').trim()
  const explicitTypes = ['简答', '论述', '名词解释', '案例分析', '问答']
  if (explicitTypes.some(type => t.includes(type))) return true
  const hasSelfOpts = props.question.options && Object.keys(props.question.options).length > 0
  const hasSharedOpts = props.sharedOptions && Object.keys(props.sharedOptions).length > 0
  return !hasSelfOpts && !hasSharedOpts
})

const isB1Child = computed(() => !!props.sharedOptions)

const displayOptions = computed(() => {
  const opts = props.question.options || props.sharedOptions
  if (!opts) return []
  return Object.keys(opts).sort().map(key => ({ key: key, value: opts[key] }))
})

const sharedOptionsList = computed(() => {
  if (!props.sharedOptions) return []
  return Object.keys(props.sharedOptions).sort().map(key => ({ key: key, value: props.sharedOptions[key] }))
})

const indexLabel = computed(() => {
  if (!props.index) return ''
  if (props.isChild) {
    return `(${props.index})`
  }
  return `${props.index}.`
})

const tagTypeComputed = computed(() => {
    const t = props.question.type || ''
    if (!t) return 'default'
    if (t.includes('A3') || t.includes('A4') || t.includes('案例')) return 'info'
    if (t.includes('B1')) return 'warning'
    return 'success' 
})

const handleRedo = async () => {
  try {
    await request.delete(`/questions/${props.question.id}/reset`)
    message.success('已重置')
    selectedOption.value = ''
    multiSelection.value = []
    result.value = null
    showAnswer.value = false
  } catch (e) { console.error(e) }
}

const handleOptionClick = (key: string) => {
  if (result.value) return 
  if (isMultiChoice.value) {
    const idx = multiSelection.value.indexOf(key)
    if (idx > -1) multiSelection.value.splice(idx, 1)
    else multiSelection.value.push(key)
  } else {
    submitAnswer(key)
  }
}

const submitMultiChoice = () => {
  if (multiSelection.value.length === 0) { message.warning('请至少选择一个选项'); return }
  const finalAnswer = multiSelection.value.sort().join('')
  submitAnswer(finalAnswer)
}

const submitAnswer = async (answerStr: string) => {
  selectedOption.value = answerStr
  submitting.value = true
  try {
    const res: any = await request.post(`/questions/${props.question.id}/submit`, { choice: answerStr })
    result.value = res 
    emit('answer-result', { id: props.question.id, isCorrect: res.is_correct })
  } catch (e) { console.error(e) } finally { submitting.value = false }
}

const getOptionClass = (key: string) => {
  const isSelected = isMultiChoice.value ? multiSelection.value.includes(key) : selectedOption.value === key
  if (!result.value) return isSelected ? 'opt-selected' : ''
  const isKeyInCorrect = result.value.correct_answer.includes(key)
  if (isKeyInCorrect) return 'opt-correct'
  if (isSelected && !isKeyInCorrect) return 'opt-wrong'
  return '' 
}

const getBtnType = (key: string) => { 
  if (!result.value) return selectedOption.value === key ? 'primary' : 'default'
  if (key === result.value.correct_answer) return 'success' 
  if (key === selectedOption.value && !result.value.is_correct) return 'error' 
  return 'default'
}

// 🔥🔥🔥 核心修复：升级版文本格式化器 🔥🔥🔥
const formatText = (text: string) => {
  if (!text) return ''
  
  // 1. 处理后端返回的标准 Markdown 图片: ![图片](/uploads/...)
  let res = text.replace(/!\[(.*?)\]\((.*?)\)/g, (match, alt, url) => {
      const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`
      return `<img src="${fullUrl}" class="zoom-image" title="点击查看大图" />`
  })

  // 2. 兼容旧有的自定义格式: [图片:...] (双保险)
  res = res.replace(/\[图片:(.*?)\]/g, (match, url) => {
      const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`
      return `<img src="${fullUrl}" class="zoom-image" title="点击查看大图" />`
  })

  return res
}

const handleContentClick = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (target.tagName === 'IMG' && target.classList.contains('zoom-image')) {
    previewImageUrl.value = (target as HTMLImageElement).src
    showPreview.value = true
    e.stopPropagation() 
  }
}
const shouldShowHeader = computed(() => props.showSharedHeader !== false)
</script>

<template>
  <div class="exam-item" @click="handleContentClick">
    
    <div v-if="isB1Child && shouldShowHeader" class="shared-options-box">
      <div class="shared-title">共用备选答案：</div>
      <div class="shared-list">
        <div v-for="opt in sharedOptionsList" :key="opt.key" class="shared-item">
          <span class="shared-key">{{ opt.key }}.</span>
          <span class="shared-value" v-html="formatText(opt.value)"></span>
        </div>
      </div>
    </div>

    <div class="q-stem">
      <!-- 浮动在右侧的重做按钮 -->
      <div v-if="result" class="redo-btn" @click.stop="handleRedo">
          <n-icon size="14"><RefreshOutline /></n-icon> 重做
      </div>

      <!-- 题型标签 & 序号 -->
      <n-tag v-if="showTypeTag" :type="tagTypeComputed" size="small" round strong style="margin-right: 6px; vertical-align: text-bottom;">{{ question.type || '题型' }}</n-tag>

      <span class="q-index">{{ indexLabel }}</span>
      <n-tag v-if="isMultiChoice" type="warning" size="small" style="margin-right: 6px; vertical-align: text-bottom;">多选</n-tag>
      
      <span class="q-text" v-html="formatText(question.stem)"></span>
    </div>

    <div v-if="isSubjective" style="margin: 10px 0 0 24px;">
      <n-button size="small" secondary @click="showAnswer = !showAnswer">
        {{ showAnswer ? '收起答案' : '查看参考答案' }}
      </n-button>
    </div>

    <div v-else-if="isB1Child" class="b1-selector-row">
       <div v-for="opt in displayOptions" :key="opt.key" class="b1-selector-item" @click="handleOptionClick(opt.key)">
        <n-button circle :type="getBtnType(opt.key)" :ghost="selectedOption !== opt.key && !result" class="opt-btn-large">{{ opt.key }}</n-button>
      </div>
    </div>

    <div v-else class="q-options">
      <div 
        v-for="opt in displayOptions" 
        :key="opt.key" 
        class="opt-row" 
        :class="getOptionClass(opt.key)"
        @click="handleOptionClick(opt.key)"
      >
        <div class="opt-circle">{{ opt.key }}</div>
        <div class="opt-content" v-html="formatText(opt.value)"></div>
      </div>

      <div v-if="isMultiChoice && !result" style="margin-top: 10px; margin-left: 10px;">
        <n-button type="primary" :disabled="multiSelection.length === 0" :loading="submitting" @click.stop="submitMultiChoice">
          确认答案
        </n-button>
      </div>
    </div>

    <div v-if="result || (isSubjective && showAnswer)" class="analysis-panel">
      <div class="result-bar" :class="result?.is_correct ? 'bar-success' : 'bar-error'">
        <div class="res-item">
          <span class="res-label">正确答案：</span>
          <span class="res-val green">{{ result?.correct_answer }}</span>
          <span v-if="isB1Child && result?.correct_answer" class="res-text-hint">
             {{ sharedOptions[result.correct_answer] }}
          </span>
        </div>
        <div class="res-item" v-if="!isSubjective">
          <span class="res-label">我的答案：</span>
          <span class="res-val" :class="result?.is_correct ? 'green' : 'red'">
            {{ selectedOption }}
          </span>
        </div>
      </div>
      <div class="analysis-body">
        <div v-if="isSubjective" class="subjective-ref">
           <div class="label-heading">【参考答案】</div>
           <div class="text-content" v-html="formatText(question.correct || '略')"></div>
        </div>
        <div class="label-heading">【解析】</div>
        <div class="text-content" v-html="formatText(result?.analysis || question.analysis || '暂无解析')"></div>
      </div>
      
      <div class="meta-footer">
        <div class="meta-group">
          <span v-if="question.difficulty">难度: {{ question.difficulty }}</span>
          <n-divider vertical v-if="question.difficulty" />
          <span v-if="question.diff_value">系数: {{ question.diff_value }}</span>
        </div>
        <div class="meta-group">
          <span v-if="question.syllabus" style="color: #2080f0; font-weight: 500;">
             {{ question.syllabus }}
          </span>
          <n-divider vertical v-if="question.syllabus && question.cognitive_level" />
          <span v-if="question.cognitive_level">{{ question.cognitive_level }}</span>
        </div>
      </div>
    </div>

    <n-modal v-model:show="showPreview" preset="card" style="width:auto; background:transparent; border:none; box-shadow:none;">
      <div class="img-preview-box" @click="showPreview = false">
        <img :src="previewImageUrl" />
      </div>
    </n-modal>
  </div>
</template>

<style>
/* 🔥 确保图片有合适的样式 */
.zoom-image { 
    max-width: 150px; 
    max-height: 150px; 
    border: 1px solid #e2e8f0; 
    margin: 8px 0; 
    cursor: zoom-in; 
    display: block; 
    border-radius: 8px; /* 图片圆角 */
    vertical-align: bottom; 
    transition: transform 0.2s;
}
.zoom-image:hover { transform: scale(1.05); }
</style>

<style scoped>
.exam-item { 
    padding: 24px 0; 
    border-bottom: 1px dashed #e2e8f0; 
    position: relative; 
}
.exam-item:last-child { border-bottom: none; }

/* 🔗 B1 共用选项 */
.shared-options-box { 
    background-color: #f8fafc; 
    border: 1px solid #e2e8f0; 
    border-radius: 16px; 
    padding: 16px 20px; 
    margin-bottom: 24px; 
}
.shared-title { 
    font-weight: 700; 
    color: #0f172a; 
    margin-bottom: 12px; 
    font-size: 14px; 
}
.shared-list { display: flex; flex-direction: column; gap: 8px; }
.shared-item { 
    font-size: 15px; 
    color: #334155; 
    line-height: 1.6; 
    display: flex; 
    align-items: baseline;
}
.shared-key { 
    font-weight: 700; 
    margin-right: 8px; 
    min-width: 24px; 
    color: #3b82f6; 
}
.shared-value { flex: 1; }

/* 📝 题干 */
.q-stem { 
    font-size: 16px; 
    line-height: 1.75; 
    color: #1e293b; 
    margin-bottom: 20px; 
    padding-right: 12px; 
    font-weight: 500;
}
.q-index { 
    font-weight: 700; 
    margin-right: 8px; 
    color: #3b82f6; 
    font-size: 18px;
}
.redo-btn { 
    float: right; 
    font-size: 12px; 
    color: #64748b; 
    cursor: pointer; 
    margin-left: 12px; 
    transition: all 0.2s;
    background-color: #f1f5f9;
    padding: 4px 10px;
    border-radius: 20px;
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 600;
}
.redo-btn:hover { 
    color: #fff; 
    background-color: #f59e0b; 
    transform: translateY(-1px);
    box-shadow: 0 2px 5px rgba(245, 158, 11, 0.3);
}

/* 🔘 B1 选择器 */
.b1-selector-row { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 16px; margin-left: 24px; }
.opt-btn-large { width: 40px; height: 40px; font-size: 16px; font-weight: 600; }

/* ✅ 选项列表 */
.q-options { display: flex; flex-direction: column; gap: 12px; }
.opt-row { 
    display: flex; 
    align-items: flex-start; 
    padding: 12px 16px; 
    cursor: pointer; 
    border-radius: 12px; 
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); 
    border: 1px solid transparent; 
    background-color: #fff;
}
.opt-row:hover { 
    background-color: #f8fafc; 
    border-color: #e2e8f0;
    transform: translateX(4px);
}
.opt-circle { 
    width: 32px; 
    height: 32px; 
    border: 2px solid #cbd5e1; 
    border-radius: 50%; 
    text-align: center; 
    line-height: 28px; 
    font-size: 15px; 
    color: #64748b; 
    margin-right: 16px; 
    flex-shrink: 0; 
    background-color: #fff; 
    font-weight: 600; 
    transition: all 0.2s;
}
.opt-content { 
    font-size: 15px; 
    color: #334155; 
    line-height: 1.8; 
    margin-top: 2px; 
    flex: 1;
}

/* 选中状态 */
.opt-selected { background-color: #eff6ff; border-color: #bfdbfe; }
.opt-selected .opt-circle { 
    border-color: #3b82f6; 
    color: #fff; 
    background-color: #3b82f6; 
    box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3);
}
.opt-selected .opt-content { color: #1e3a8a; font-weight: 500; }

/* 错误状态 */
.opt-wrong { background-color: #fef2f2; border-color: #fecaca; }
.opt-wrong .opt-circle { 
    border-color: #ef4444; 
    color: #fff; 
    background-color: #ef4444; 
    box-shadow: 0 2px 6px rgba(239, 68, 68, 0.3);
}
.opt-wrong .opt-content { color: #991b1b; }

/* 正确状态 */
.opt-correct { background-color: #f0fdf4; border-color: #bbf7d0; }
.opt-correct .opt-circle { 
    background-color: #10b981; 
    border-color: #10b981; 
    color: #fff; 
    box-shadow: 0 2px 6px rgba(16, 185, 129, 0.3);
}
.opt-correct .opt-content { color: #065f46; font-weight: 600; }

/* 💡 解析面板 */
.analysis-panel { 
    margin-top: 24px; 
    background-color: #fff; 
    border-radius: 16px; 
    border: 1px solid #f1f5f9; 
    overflow: hidden; 
    box-shadow: 0 4px 12px rgba(0,0,0,0.03); 
    transition: all 0.3s;
}
.analysis-panel:hover { box-shadow: 0 8px 24px rgba(0,0,0,0.06); }

.result-bar { 
    display: flex; 
    gap: 32px; 
    padding: 16px 24px; 
    border-bottom: 1px solid #f1f5f9; 
    align-items: center;
}
.bar-success { background: linear-gradient(to right, #f0fdf4, #fff); border-left: 6px solid #10b981; }
.bar-error { background: linear-gradient(to right, #fef2f2, #fff); border-left: 6px solid #ef4444; }

.res-item { display: flex; align-items: baseline; gap: 8px; font-size: 15px; }
.res-label { color: #64748b; font-weight: 600; }
.res-val { font-weight: 800; font-family: 'Roboto Mono', monospace; font-size: 18px; }
.res-val.green { color: #10b981; text-shadow: 0 1px 2px rgba(16, 185, 129, 0.1); }
.res-val.red { color: #ef4444; text-shadow: 0 1px 2px rgba(239, 68, 68, 0.1); }
.res-text-hint { font-size: 14px; color: #475569; font-weight: 500; }

.analysis-body { padding: 24px; color: #334155; }
.label-heading { 
    font-weight: 700; 
    font-size: 15px; 
    color: #1e293b; 
    margin-bottom: 12px; 
    display: flex;
    align-items: center;
}
.label-heading::before {
    content: '';
    display: inline-block;
    width: 4px;
    height: 16px;
    background-color: #3b82f6;
    margin-right: 8px;
    border-radius: 2px;
}
.subjective-ref { margin-bottom: 24px; }
.text-content { 
    font-size: 15px; 
    line-height: 1.8; 
    color: #475569; 
    text-align: justify; 
}

.meta-footer { 
    display: flex; 
    justify-content: space-between; 
    align-items: center; 
    padding: 12px 24px; 
    background-color: #f8fafc; 
    border-top: 1px solid #f1f5f9; 
    font-size: 13px; 
    color: #94a3b8; 
}
.meta-group { display: flex; align-items: center; gap: 12px; }

.img-preview-box { display: flex; justify-content: center; align-items: center; cursor: zoom-out; }
.img-preview-box img { 
    max-width: 90vw; 
    max-height: 90vh; 
    box-shadow: 0 10px 40px rgba(0,0,0,0.3); 
    border-radius: 12px; 
}
</style>