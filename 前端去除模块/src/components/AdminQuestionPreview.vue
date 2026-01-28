<script setup lang="ts">
import { computed } from 'vue'
import { NTag, NDivider } from 'naive-ui'

const props = defineProps<{
  question: any
}>()

// 辅助：安全解析选项 JSON
const parseOptions = (opts: any) => {
  if (!opts) return {}
  if (typeof opts === 'string') {
    try { return JSON.parse(opts) } catch { return {} }
  }
  return opts
}

// 辅助：文本格式化 (处理图片等)
const formatText = (text: string) => {
  if (!text) return ''
  return text.replace(/\[图片:(.*?)\]/g, '<img src="$1" class="preview-img" />')
}

// 类型判断
const isGroup = computed(() => props.question.children && props.question.children.length > 0)
const isB1 = computed(() => (props.question.type || '').toUpperCase().includes('B1'))
const isCase = computed(() => ['A3', 'A4', '案例'].some(t => (props.question.type || '').toUpperCase().includes(t)))

// B1 共用选项
const sharedOptions = computed(() => {
  if (isB1.value) return parseOptions(props.question.options)
  return null
})

// A3/A4 共用题干
const mainStem = computed(() => {
  if (isCase.value) return formatText(props.question.stem)
  return ''
})
</script>

<template>
  <div class="admin-preview-container">
    
    <div class="meta-header">
      <NTag type="success" size="small" bordered>{{ question.type || '题型未知' }}</NTag>
      <span class="meta-text">ID: {{ question.id }}</span>
      <span class="meta-text">难度: {{ question.difficulty || '-' }}</span>
      <span class="meta-text">来源: {{ question.source }}</span>
      <span class="meta-text">路径: {{ question.category_path }}</span>
    </div>

    <div v-if="isGroup">
      
      <div v-if="isCase && mainStem" class="group-box stem-box">
        <div class="box-label">（主题干）</div>
        <div class="box-content" v-html="mainStem"></div>
      </div>

      <div v-if="isB1 && sharedOptions" class="group-box options-box">
        <div class="box-label">（共用备选答案）</div>
        <div class="options-list">
          <div v-for="(val, key) in sharedOptions" :key="key" class="opt-row static">
            <span class="opt-key">{{ key }}.</span>
            <span class="opt-val" v-html="formatText(val)"></span>
          </div>
        </div>
      </div>

      <div class="children-list">
        <div v-for="(child, idx) in question.children" :key="child.id" class="child-item">
          <div class="child-index">第 {{ idx + 1 }} 小题 (ID: {{ child.id }})</div>
          
          <div class="q-stem" v-html="formatText(child.stem)"></div>

          <div v-if="!isB1" class="q-options">
             <div v-for="(val, key) in parseOptions(child.options)" :key="key" 
                  class="opt-row" 
                  :class="{ 'is-correct': key === child.correct }">
                <span class="opt-key">{{ key }}</span>
                <span class="opt-val" v-html="formatText(val)"></span>
             </div>
          </div>

          <div class="analysis-wrapper">
            <div class="ans-row">
              <span class="label">正确答案：</span>
              <span class="value-green">{{ child.correct }}</span>
              <span v-if="isB1 && sharedOptions && sharedOptions[child.correct]" class="b1-hint">
                {{ sharedOptions[child.correct] }}
              </span>
            </div>
            <div class="ana-row">
              <span class="label">解析：</span>
              <span class="ana-text" v-html="formatText(child.analysis || '暂无解析')"></span>
            </div>
          </div>
          
          <NDivider v-if="idx < question.children.length - 1" dashed style="margin: 20px 0;" />
        </div>
      </div>
    </div>

    <div v-else>
      <div class="q-stem" v-html="formatText(question.stem)"></div>
      
      <div class="q-options">
         <div v-for="(val, key) in parseOptions(question.options)" :key="key" 
              class="opt-row" 
              :class="{ 'is-correct': key === question.correct }">
            <span class="opt-key">{{ key }}</span>
            <span class="opt-val" v-html="formatText(val)"></span>
         </div>
      </div>

      <div class="analysis-wrapper">
        <div class="ans-row">
          <span class="label">正确答案：</span>
          <span class="value-green">{{ question.correct }}</span>
        </div>
        <div class="ana-row">
          <span class="label">解析：</span>
          <span class="ana-text" v-html="formatText(question.analysis || '暂无解析')"></span>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
.admin-preview-container { font-size: 15px; color: #333; line-height: 1.6; }
.meta-header { display: flex; gap: 12px; align-items: center; margin-bottom: 20px; padding-bottom: 12px; border-bottom: 1px dashed #eee; }
.meta-text { font-size: 12px; color: #999; }

.group-box { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 12px 16px; margin-bottom: 20px; }
.box-label { font-weight: bold; color: #0f766e; margin-bottom: 8px; font-size: 14px; }
.box-content { font-size: 15px; color: #334155; }

.options-list, .q-options { display: flex; flex-direction: column; gap: 8px; }
.opt-row { display: flex; align-items: flex-start; padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 4px; background: #fff; }
.opt-row.is-correct { border-color: #22c55e; background-color: #f0fdf4; } /* 绿色高亮 */
.opt-key { font-weight: bold; margin-right: 10px; min-width: 24px; color: #64748b; }
.is-correct .opt-key { color: #15803d; }

.child-item { padding-left: 4px; }
.child-index { font-size: 12px; font-weight: bold; color: #94a3b8; margin-bottom: 8px; background: #f1f5f9; display: inline-block; padding: 2px 8px; border-radius: 4px; }
.q-stem { font-weight: 500; margin-bottom: 12px; font-size: 16px; }

.analysis-wrapper { margin-top: 16px; background: #fffbeb; border: 1px solid #fcd34d; padding: 12px; border-radius: 4px; }
.ans-row { margin-bottom: 8px; display: flex; align-items: baseline; }
.ana-row { display: flex; align-items: baseline; }
.label { font-weight: bold; color: #b45309; margin-right: 6px; min-width: 40px; }
.value-green { color: #16a34a; font-weight: 900; font-size: 16px; font-family: Arial; }
.b1-hint { color: #78716c; margin-left: 8px; font-size: 14px; }
.ana-text { color: #444; font-size: 14px; }

:deep(.preview-img) { max-width: 100%; border-radius: 4px; border: 1px solid #eee; margin-top: 4px; }
</style>