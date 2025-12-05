<template>
  <el-card class="global-config-card" shadow="never" ref="rootRef">
    <div class="global-config-header">
      <span>工作流全局配置</span>
    </div>
  <el-tabs v-model="activeTab">
      <el-tab-pane label="基本信息" name="basic">
        <div ref="basicRef" class="basic-scope">
          <BasicInfoConfig
            v-model:name="globalForm.name"
            v-model:runName="globalForm.runName"
            :runNamePlaceholder="runNamePlaceholder"
          />
        </div>
      </el-tab-pane>
      <el-tab-pane label="触发事件配置" name="on">
        <OnConfig v-model="globalForm.onText" :placeholder="onPlaceholder" :rows="8" />
      </el-tab-pane>
      <el-tab-pane label="环境变量配置" name="env">
        <EnvConfig v-model:pairs="globalForm.envPairs" />
      </el-tab-pane>
      <el-tab-pane label="工作流权限" name="permissions">
        <PermissionsConfig v-model:pairs="globalForm.permissionsPairs" />
      </el-tab-pane>
      <el-tab-pane label="并发控制" name="concurrency">
        <ConcurrencyConfig v-model="globalForm.concurrencyText" :placeholder="concurrencyPlaceholder" :rows="6" />
      </el-tab-pane>
      <el-tab-pane label="默认值 defaults" name="defaults">
        <DefaultsConfig v-model="globalForm.defaultsText" :placeholder="defaultsPlaceholder" :rows="6" />
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'
import { ElMessage } from 'element-plus'
import BasicInfoConfig from './BasicInfoConfig.vue'
import OnConfig from './OnConfig.vue'
import EnvConfig from './EnvConfig.vue'
import PermissionsConfig from './PermissionsConfig.vue'
import ConcurrencyConfig from './ConcurrencyConfig.vue'
import DefaultsConfig from './DefaultsConfig.vue'
// 实时补丁事件：将表单改动以增量方式同步到父级（YAML 文档）
const emit = defineEmits(['patch'])

const props = defineProps({
  doc: { type: Object, required: true }
})

const activeTab = ref('on')
const globalForm = ref({
  name: '',
  runName: '',
  onText: '',
  envPairs: [],
  permissionsPairs: [],
  concurrencyText: '',
  defaultsText: ''
})

// 初始化/应用外部 doc 时不触发写回，避免循环
const isApplyingDoc = ref(false)
// 文本字段防抖（on/concurrency/defaults）
const debTimers = {}

const onPlaceholder = '示例 YAML:\n  push:\n    branches: [main]\n  pull_request: {}'
const concurrencyPlaceholder = '示例 YAML:\n  group: build-main\n  cancel-in-progress: true'
const runNamePlaceholder = '示例：使用表达式 ${{ github.run_id }} 或自定义文本'
const defaultsPlaceholder = '示例 YAML:\n  run:\n    shell: bash\n    working-directory: ./src'

const envObjToPairs = (envObj) => {
  try {
    if (!envObj || typeof envObj !== 'object') return []
    return Object.keys(envObj).map((k) => ({ key: k, value: String(envObj[k] ?? '') }))
  } catch (_) {
    return []
  }
}
const envPairsToObj = (pairs) => {
  const obj = {}
  try {
    (pairs || []).forEach((p) => {
      const k = String(p?.key || '').trim()
      if (k) obj[k] = p?.value ?? ''
    })
    return Object.keys(obj).length ? obj : null
  } catch (_) {
    return null
  }
}

const initFromDoc = (doc) => {
  isApplyingDoc.value = true
  try {
    globalForm.value.name = String(doc?.name ?? '')
  } catch (_) { globalForm.value.name = '' }
  try {
    globalForm.value.runName = String(doc?.['run-name'] ?? '')
  } catch (_) { globalForm.value.runName = '' }
  try {
    const onVal = doc?.on
    globalForm.value.onText = onVal === undefined ? '' : yamlDump(onVal, { noRefs: true, lineWidth: 120 })
  } catch (_) { globalForm.value.onText = '' }
  try {
    // workflow_dispatch 现已整合到 OnConfig 的 onText 中
  } catch (_) { /* ignore */ }
  try {
    globalForm.value.envPairs = envObjToPairs(doc?.env)
  } catch (_) { globalForm.value.envPairs = [] }
  try {
    globalForm.value.permissionsPairs = envObjToPairs(doc?.permissions)
  } catch (_) { globalForm.value.permissionsPairs = [] }
  try {
    const cVal = doc?.concurrency
    globalForm.value.concurrencyText = cVal === undefined ? '' : yamlDump(cVal, { noRefs: true, lineWidth: 120 })
  } catch (_) { globalForm.value.concurrencyText = '' }
  try {
    const dVal = doc?.defaults
    globalForm.value.defaultsText = dVal === undefined ? '' : yamlDump(dVal, { noRefs: true, lineWidth: 120 })
  } catch (_) { globalForm.value.defaultsText = '' }
  isApplyingDoc.value = false
}

watch(() => props.doc, (d) => { if (d) initFromDoc(d) }, { immediate: true, deep: true })

const collectPayload = () => {
  try {
    const payload = {}
    payload.name = String(globalForm.value.name || '').trim()
    payload.runName = String(globalForm.value.runName || '').trim()
    const onText = String(globalForm.value.onText || '').trim()
    if (onText) {
      try { payload.on = yamlLoad(onText) } catch (e) { ElMessage.error('on 配置不是有效的 YAML'); return null }
    } else { payload.on = null }
    payload.env = envPairsToObj(globalForm.value.envPairs)
    payload.permissions = envPairsToObj(globalForm.value.permissionsPairs)
    const cText = String(globalForm.value.concurrencyText || '').trim()
    if (cText) {
      try { payload.concurrency = yamlLoad(cText) } catch (e) { ElMessage.error('concurrency 配置不是有效的 YAML'); return null }
    } else { payload.concurrency = null }
    const dText = String(globalForm.value.defaultsText || '').trim()
    if (dText) {
      try { payload.defaults = yamlLoad(dText) } catch (e) { ElMessage.error('defaults 配置不是有效的 YAML'); return null }
    } else { payload.defaults = null }
    return payload
  } catch (e) {
    console.error('收集全局配置失败', e)
    ElMessage.error('收集全局配置失败')
    return null
  }
}

// 生成实时补丁，不弹出错误，仅在解析成功时携带结构化字段
const makeRealtimePatch = () => {
  try {
    const payload = {}
    payload.name = String(globalForm.value.name || '').trim()
    payload.runName = String(globalForm.value.runName || '').trim()
    // on（文本为空表示删除；非空解析成功才写入）
    const onText = String(globalForm.value.onText || '').trim()
    if (!onText) payload.on = null
    else {
      try { const obj = yamlLoad(onText); if (obj && typeof obj === 'object') payload.on = obj } catch (_) {}
    }
    // env / permissions
    payload.env = envPairsToObj(globalForm.value.envPairs)
    payload.permissions = envPairsToObj(globalForm.value.permissionsPairs)
    // concurrency / defaults（同 on 的策略）
    const cText = String(globalForm.value.concurrencyText || '').trim()
    if (!cText) payload.concurrency = null
    else { try { const cObj = yamlLoad(cText); if (cObj && typeof cObj === 'object') payload.concurrency = cObj } catch (_) {} }
    const dText = String(globalForm.value.defaultsText || '').trim()
    if (!dText) payload.defaults = null
    else { try { const dObj = yamlLoad(dText); if (dObj && typeof dObj === 'object') payload.defaults = dObj } catch (_) {} }
    return payload
  } catch (_) { return {} }
}
const emitRealtimePatch = () => {
  if (isApplyingDoc.value) return
  try { const payload = makeRealtimePatch(); emit('patch', payload) } catch (_) {}
}

// 全局配置容器范围的事件拦截（window 捕获级别，优先于 document）
const rootRef = ref(null)
let removeIntercept = () => {}
onMounted(() => {
  const handler = (e) => {
    try {
      const el = rootRef.value
      if (!el) return
      const t = e.target
      if (t && el.contains(t)) {
        if (typeof e.stopImmediatePropagation === 'function') e.stopImmediatePropagation()
        e.stopPropagation()
      }
    } catch (_) {}
  }
  const types = ['focusin', 'focus', 'click', 'mousedown']
  types.forEach((tp) => window.addEventListener(tp, handler, true))
  removeIntercept = () => types.forEach((tp) => window.removeEventListener(tp, handler, true))
})
onBeforeUnmount(() => { try { removeIntercept() } catch (_) {} })

// 基本信息页签局部拦截（可独立关闭或扩展）
const basicRef = ref(null)
let removeBasicIntercept = () => {}
onMounted(() => {
  const handler = (e) => {
    try {
      const el = basicRef.value
      if (!el) return
      const t = e.target
      if (t && el.contains(t)) {
        if (typeof e.stopImmediatePropagation === 'function') e.stopImmediatePropagation()
        e.stopPropagation()
      }
    } catch (_) {}
  }
  const types = ['focusin', 'focus', 'click', 'mousedown']
  types.forEach((tp) => window.addEventListener(tp, handler, true))
  removeBasicIntercept = () => types.forEach((tp) => window.removeEventListener(tp, handler, true))
})
onBeforeUnmount(() => { try { removeBasicIntercept() } catch (_) {} })

// 双向实时同步：表单 -> YAML（父组件）
watch(() => globalForm.value.name, () => { if (!isApplyingDoc.value) emit('patch', { name: String(globalForm.value.name || '').trim() }) })
watch(() => globalForm.value.runName, () => { if (!isApplyingDoc.value) emit('patch', { runName: String(globalForm.value.runName || '').trim() }) })
watch(() => globalForm.value.onText, () => {
  if (isApplyingDoc.value) return
  if (debTimers.on) { clearTimeout(debTimers.on); debTimers.on = null }
  debTimers.on = setTimeout(() => emitRealtimePatch(), 300)
})
watch(() => globalForm.value.envPairs, () => emitRealtimePatch(), { deep: true })
watch(() => globalForm.value.permissionsPairs, () => emitRealtimePatch(), { deep: true })
watch(() => globalForm.value.concurrencyText, () => {
  if (isApplyingDoc.value) return
  if (debTimers.concurrency) { clearTimeout(debTimers.concurrency); debTimers.concurrency = null }
  debTimers.concurrency = setTimeout(() => emitRealtimePatch(), 300)
})
watch(() => globalForm.value.defaultsText, () => {
  if (isApplyingDoc.value) return
  if (debTimers.defaults) { clearTimeout(debTimers.defaults); debTimers.defaults = null }
  debTimers.defaults = setTimeout(() => emitRealtimePatch(), 300)
})

defineExpose({ collectPayload })
</script>

<style scoped>
.global-config-card { margin-bottom: 12px; height: 100%; width: 100%; display: flex; flex-direction: column; box-sizing: border-box; }
.global-config-card :deep(.el-card__body) { display: flex; flex-direction: column; flex: 1; height: 100%; padding: 8px; box-sizing: border-box; min-width: 0; min-height: 0; }
.global-config-card :deep(.el-tabs) { display: flex; flex-direction: column; flex: 1; height: 100%; min-width: 0; min-height: 0; }
.global-config-card :deep(.el-tabs__content) { flex: 1; display: flex; height: 100%; min-width: 0; min-height: 0; }
.global-config-card :deep(.el-tab-pane) { flex: 1; display: flex; height: 100%; min-width: 0; min-height: 0; }
.global-config-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.env-pair-row { margin-bottom: 8px; display: flex; align-items: center; }
</style>
