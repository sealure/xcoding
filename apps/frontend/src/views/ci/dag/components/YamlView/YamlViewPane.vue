<template>
  <div class="yaml-view-root">
    <div class="yaml-view-toolbar">
      <el-tag :type="yamlConsistent ? 'success' : 'warning'" size="small">
        {{ yamlConsistent ? '与原始 YAML 一致' : '与原始 YAML 存在差异' }}
      </el-tag>
      <el-divider direction="vertical" />
      <el-button size="small" @click="copyCurrentYaml">复制当前 YAML</el-button>
    </div>
    <el-alert v-if="yamlEditError" :title="yamlEditError" type="error" show-icon class="mb8" />
    <div class="yaml-view-split">
      <div class="yaml-column">
        <CodeEditor
          :modelValue="localOriginalText"
          :fit="true"
          language="yaml"
          theme="dark"
          title="原始 YAML（只读）"
          :showActions="false"
          :readOnly="true"
        />
      </div>
      <div class="yaml-column">
        <CodeEditor
          v-model="currentYamlText"
          :fit="true"
          language="yaml"
          theme="dark"
          title="当前 YAML（可编辑，自动同步画布）"
          :showActions="true"
          :readOnly="false"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'
import CodeEditor from '../PipelineView/common/CodeEditor.vue'

const props = defineProps({
  originalText: { type: String, default: '' },
  doc: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['update:doc'])

// 与编辑/保存保持一致的键排序：全局配置置顶，其次 jobs，再其余
const GLOBAL_FIRST_KEYS = ['name', 'run-name', 'on', 'env', 'permissions', 'concurrency', 'defaults']
const reorderYamlDoc = (doc) => {
  try {
    if (!doc || typeof doc !== 'object' || Array.isArray(doc)) return doc
    const next = {}
    const keys = Object.keys(doc)
    // 顶层：全局配置置顶
    GLOBAL_FIRST_KEYS.forEach((k) => { if (k in doc) next[k] = doc[k] })
    // 顶层：jobs 置于全局配置之后，并重排内部 steps 至最后
    if ('jobs' in doc && doc.jobs && typeof doc.jobs === 'object') {
      const jobsNext = {}
      Object.keys(doc.jobs).forEach((jid) => {
        const job = doc.jobs[jid]
        if (!job || typeof job !== 'object') { jobsNext[jid] = job; return }
        const jobNext = {}
        const jKeys = Object.keys(job)
        jKeys.forEach((jk) => { if (jk !== 'steps') jobNext[jk] = job[jk] })
        if ('steps' in job) jobNext.steps = job.steps
        jobsNext[jid] = jobNext
      })
      next.jobs = jobsNext
    }
    // 顶层：其余键按原顺序追加
    keys.forEach((k) => { if (!(k in next)) next[k] = doc[k] })
    return next
  } catch (_) { return doc }
}

const localOriginalText = ref(props.originalText || '')
watch(() => props.originalText, (t) => { localOriginalText.value = t || '' })

const currentYamlText = ref('')
const yamlEditError = ref('')
let updatingFromDoc = false
let yamlEditTimer = null
let suppressDocToTextOnce = false

// 由 doc -> 当前 YAML 文本
watch(() => props.doc, (doc) => {
  try {
    if (suppressDocToTextOnce) { suppressDocToTextOnce = false; return }
    if (!doc || typeof doc !== 'object') { updatingFromDoc = true; currentYamlText.value = ''; updatingFromDoc = false; return }
    updatingFromDoc = true
    currentYamlText.value = yamlDump(reorderYamlDoc(doc), { lineWidth: 120, noRefs: true })
    updatingFromDoc = false
  } catch (_) {
    currentYamlText.value = ''
  }
}, { deep: true })

// 由当前 YAML 文本 -> doc（防抖与错误提示）
watch(currentYamlText, (text) => {
  try {
    if (updatingFromDoc) return
    if (yamlEditTimer) { clearTimeout(yamlEditTimer); yamlEditTimer = null }
    yamlEditTimer = setTimeout(() => {
      try {
        const trimmed = String(text || '').trim()
        if (!trimmed) { yamlEditError.value = ''; emit('update:doc', {}); return }
        const obj = yamlLoad(trimmed)
        if (!obj || typeof obj !== 'object' || Array.isArray(obj)) { yamlEditError.value = 'YAML 根节点必须为对象'; return }
        yamlEditError.value = ''
        suppressDocToTextOnce = true
        emit('update:doc', obj)
      } catch (e) {
        yamlEditError.value = `YAML 解析失败：${e.message || e}`
      }
    }, 300)
  } catch (_) {}
})

const yamlConsistent = computed(() => {
  try {
    const a = (props.originalText || '').trim()
    if (!a && !currentYamlText.value.trim()) return true
    const objA = yamlLoad(a)
    const objB = yamlLoad(currentYamlText.value || '')
    return JSON.stringify(objA) === JSON.stringify(objB)
  } catch {
    return false
  }
})

const copyCurrentYaml = async () => {
  try {
    await navigator.clipboard.writeText(currentYamlText.value || '')
    // 提示交由父级或忽略
  } catch (_) {}
}
</script>

<style scoped>
.yaml-view-root { display: flex; flex-direction: column; width: 100%; height: 100%; min-width: 0; min-height: 0; }
.yaml-view-toolbar { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.yaml-view-split { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; flex: 1 1 auto; min-height: 0; height: 100%; }
.yaml-column { display: flex; flex-direction: column; min-width: 0; min-height: 0; height: 100%; }
</style>