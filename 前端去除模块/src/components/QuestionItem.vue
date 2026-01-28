<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NButton, useMessage, NTag, NIcon, NModal } from 'naive-ui'
import request from '../utils/request'

const props = defineProps<{
  question: any       
  sharedOptions?: any 
  index?: number    
  isChild?: boolean 
  showSharedHeader?: boolean 
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

// 🔥🔥🔥 修复点：统一样式，不再区别对待子题括号 🔥🔥🔥
const indexLabel = computed(() => {
  if (!props.index) return ''
  
  // 🔥🔥🔥 核心修改：如果是子题目，显示 (1)，否则显示 1. 🔥🔥🔥
  if (props.isChild) {
    return `(${props.index})`
  }
  
  return `${props.index}.`
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

const formatText = (text: string) => {
  if (!text) return ''
  return text.replace(/\[图片:(.*?)\]/g, '<img src="$1" class="zoom-image" title="点击查看大图" />')
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
      <span class="q-index">{{ indexLabel }}</span>
      <n-tag v-if="isMultiChoice" type="warning" size="small" style="margin-right: 6px; vertical-align: text-bottom;">多选</n-tag>
      <span class="q-text" v-html="formatText(question.stem)"></span>
      <a v-if="result" class="redo-link" @click.stop="handleRedo">重做</a>
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
.zoom-image { max-width: 120px; max-height: 120px; border: 1px solid #e0e0e0; margin: 5px 0; cursor: zoom-in; display: block; border-radius: 4px; }
</style>

<style scoped>
.exam-item { padding: 24px 0; border-bottom: 1px solid #eee; position: relative; }

.shared-options-box { background-color: #f0f9ff; border: 1px solid #e0f2fe; border-radius: 6px; padding: 12px 16px; margin-bottom: 20px; }
.shared-title { font-weight: bold; color: #0c4a6e; margin-bottom: 8px; font-size: 14px; }
.shared-list { display: flex; flex-direction: column; gap: 6px; }
.shared-item { font-size: 15px; color: #333; line-height: 1.5; display: flex; }
.shared-key { font-weight: bold; margin-right: 8px; min-width: 20px; color: #0ea5e9; }

.q-stem { font-size: 16px; line-height: 1.6; color: #2c3e50; margin-bottom: 16px; padding-right: 10px; }
.q-index { font-weight: bold; margin-right: 5px; color: #000; }
.redo-link { float: right; font-size: 13px; color: #999; cursor: pointer; text-decoration: underline; margin-left: 10px; }
.redo-link:hover { color: #f0a020; }

.b1-selector-row { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 10px; margin-left: 20px; }
.opt-btn-large { width: 36px; height: 36px; font-size: 15px; font-weight: 600; }

.q-options { display: flex; flex-direction: column; gap: 8px; }
.opt-row { display: flex; align-items: flex-start; padding: 10px 12px; cursor: pointer; border-radius: 6px; transition: all 0.1s; border: 1px solid transparent; }
.opt-row:hover { background-color: #f5f7fa; }
.opt-circle { width: 28px; height: 28px; border: 1px solid #dcdfe6; border-radius: 50%; text-align: center; line-height: 26px; font-size: 14px; color: #606266; margin-right: 12px; flex-shrink: 0; background-color: #fff; font-weight: 500; }
.opt-content { font-size: 15px; color: #333; line-height: 1.8; margin-top: -2px; }

.opt-selected .opt-circle { border-color: #18a058; color: #18a058; background-color: #eafbf2; }
.opt-selected { background-color: #f0fdf4; border-color: #bbf7d0; }

.opt-wrong .opt-circle { border-color: #d03050; color: #d03050; background-color: #fef0f0; }
.opt-wrong .opt-content { color: #d03050; }
.opt-correct .opt-circle { background-color: #18a058; border-color: #18a058; color: #fff; }
.opt-correct .opt-content { color: #18a058; font-weight: 600; }

.analysis-panel { margin-top: 20px; background-color: #fff; border-radius: 8px; border: 1px solid #ebebeb; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.02); }
.result-bar { display: flex; gap: 20px; padding: 12px 16px; border-bottom: 1px solid #f0f0f0; }
.bar-success { background-color: #f0fdf4; border-bottom-color: #dcfce7; }
.bar-error { background-color: #fef2f2; border-bottom-color: #fee2e2; }
.res-item { display: flex; align-items: baseline; gap: 6px; font-size: 15px; }
.res-label { color: #666; font-weight: bold; }
.res-val { font-weight: 900; font-family: Arial; font-size: 16px; }
.res-val.green { color: #18a058; }
.res-val.red { color: #d03050; }
.res-text-hint { font-size: 14px; color: #666; font-weight: normal; margin-left: 4px; }

.analysis-body { padding: 16px; color: #333; }
.label-heading { font-weight: bold; font-size: 14px; color: #333; margin-bottom: 8px; border-left: 3px solid #2080f0; padding-left: 8px; line-height: 1; }
.subjective-ref { margin-bottom: 16px; }
.text-content { font-size: 14px; line-height: 1.7; color: #444; text-align: justify; }

.meta-footer { display: flex; justify-content: space-between; align-items: center; padding: 8px 16px; background-color: #fafafa; border-top: 1px solid #eee; font-size: 12px; color: #999; }
.meta-group { display: flex; align-items: center; gap: 8px; }

.img-preview-box { display: flex; justify-content: center; align-items: center; cursor: zoom-out; }
.img-preview-box img { max-width: 90vw; max-height: 90vh; box-shadow: 0 5px 20px rgba(0,0,0,0.5); border-radius: 4px; }
</style>