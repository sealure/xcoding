<template>
  <el-drawer v-model="localVisible" title="步骤详情" direction="rtl" size="30%">
    
    <template #default>
      <div v-if="step" class="drawer-content">
        <div class="drawer-subtitle">
          <span class="label">步骤名称:</span>
          <span class="value">{{ stepDisplayName }}</span>
        </div>
        
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本配置" name="run">
            <CodeEditor
              v-model="runText"
              :rows="20"
              :fit="true"
              language="shell"
              title="编写"
              theme="dark"
              :showActions="false"
              placeholder="当前步骤的 run 内容预览"
            />
          </el-tab-pane>
          <el-tab-pane label="高级选项" name="advanced">
            <el-form label-width="120px" class="advanced-form">
              <el-form-item label="名称 (name)">
                <el-input v-model="localStep.name" placeholder="可选，步骤显示名称" />
              </el-form-item>
              <el-form-item label="类型">
                <el-input :model-value="stepType || '-'" disabled />
              </el-form-item>
              <el-form-item label="uses">
                <el-input v-model="localStep.uses" placeholder="如 actions/checkout@v4" />
              </el-form-item>
              <el-form-item label="env（键值对）">
                <div>
                  <div v-for="(pair, idx) in localStep.envPairs" :key="`env-${idx}`" class="env-pair-row">
                    <el-input v-model="pair.key" placeholder="KEY" style="width: 40%; margin-right: 8px" />
                    <el-input v-model="pair.value" placeholder="VALUE" style="width: 48%; margin-right: 8px" />
                    <el-button type="danger" plain size="small" @click="removeEnvPair(idx)">删除</el-button>
                  </div>
                  <el-button type="primary" plain size="small" @click="addEnvPair">新增变量</el-button>
                </div>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
        <div style="text-align: right; margin-top: 12px">
          <el-button @click="localVisible = false">取消</el-button>
          <el-button type="primary" @click="onSave">保存</el-button>
        </div>
      </div>
      <el-empty v-else description="请选择一个步骤节点查看详情" />
    </template>
  </el-drawer>
</template>

<script setup>
import { computed, ref, watch, reactive } from 'vue'
import CodeEditor from '../../common/CodeEditor.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  step: { type: Object, default: null },
  doc: { type: Object, default: () => ({}) },
  jobId: { type: String, default: '' },
  stepIndex: { type: Number, default: -1 }
})
const emit = defineEmits(['update:modelValue', 'save'])

const localVisible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const activeTab = ref('run')
const runText = ref('')
const localStep = reactive({ name: '', uses: '', envPairs: [] })
watch(
  () => props.step,
  (s) => {
    try {
      runText.value = String(s?.run || '')
    } catch (_) { runText.value = '' }
    try {
      localStep.name = String(s?.name || s?.label || '')
    } catch (_) { localStep.name = '' }
    try {
      localStep.uses = String(s?.uses || '')
    } catch (_) { localStep.uses = '' }
    try {
      localStep.envPairs = envObjToPairs(s?.env)
    } catch (_) { localStep.envPairs = [] }
  },
  { immediate: true, deep: true }
)

const stepDisplayName = computed(() => {
  try {
    const s = props.step
    return s?.name || s?.label || '-'
  } catch (_) {
    return '-'
  }
})

const stepType = computed(() => {
  try {
    const s = props.step
    if (!s) return '-'
    const cond = s?.if ? String(s.if).toLowerCase() : ''
    if (!cond) return 'stage'
    return cond.includes('always()') ? 'always' : 'conditional'
  } catch (_) {
    return '-'
  }
})

const formatEnv = (envObj) => {
  try {
    if (!envObj || typeof envObj !== 'object') return '-'
    const keys = Object.keys(envObj)
    if (!keys.length) return '-'
    return keys.map((k) => `${k}=${String(envObj[k])}`).join('\n')
  } catch (_) {
    return '-'
  }
}

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

const addEnvPair = () => { localStep.envPairs.push({ key: '', value: '' }) }
const removeEnvPair = (idx) => { localStep.envPairs.splice(idx, 1) }

const onSave = () => {
  try {
    const updated = { ...(props.step || {}) }
    // run
    const rt = String(runText.value || '')
    if (rt) updated.run = rt; else delete updated.run
    // name
    const nameVal = String(localStep.name || '').trim()
    if (nameVal) updated.name = nameVal; else delete updated.name
    // uses
    const usesVal = String(localStep.uses || '').trim()
    if (usesVal) updated.uses = usesVal; else delete updated.uses
    // env
    const envObj = envPairsToObj(localStep.envPairs)
    if (envObj) updated.env = envObj; else delete updated.env

    // 尝试直接写入 doc
    const jid = props.jobId
    const idx = props.stepIndex
    const next = { ...(props.doc || {}) }
    try {
      if (jid && typeof idx === 'number' && idx >= 0) {
        if (!next.jobs) next.jobs = {}
        const job = { ...(next.jobs[jid] || {}) }
        const steps = Array.isArray(job.steps) ? [...job.steps] : []
        if (idx < steps.length) steps[idx] = updated; else steps.push(updated)
        job.steps = steps
        next.jobs[jid] = job
        emit('save', next)
      } else {
        // 无上下文时，仅回传更新后的 step
        emit('save', updated)
      }
    } catch (_) {
      emit('save', updated)
    }
    localVisible.value = false
  } catch (e) {
    console.warn('保存步骤失败', e)
  }
}
</script>

<style scoped>
.run-block { white-space: pre-wrap; font-family: Menlo, Monaco, Consolas, 'Courier New', monospace; }
.drawer-content :deep(.el-tabs) { width: 100%; flex: 1 1 auto; display: flex; flex-direction: column; min-height: 0; }
.drawer-content { box-sizing: border-box; display: flex; flex-direction: column; height: 100%; }
.drawer-subtitle { display: flex; align-items: center; gap: 6px; margin: 4px 0 6px; font-size: 13px; }
.drawer-subtitle .label { color: #909399; }
.drawer-subtitle .value { color: #606266; font-weight: 500; }
/* 收紧抽屉与标签整体间距 */
:deep(.el-drawer__body) { padding: 10px 12px; }
:deep(.el-tabs__header) { margin-bottom: 6px; }
:deep(.el-tabs__content) { padding: 0; flex: 1 1 auto; height: 100%; }

/* Drawer 主体设为可伸缩容器，子内容可填充剩余空间 */
:deep(.el-drawer__body) { display: flex; flex-direction: column; height: 100%; }

/* 让 TabPane 填充，以便编辑器 fit 占满 */
.drawer-content :deep(.el-tab-pane) { height: 100%; display: flex; flex-direction: column; }
.drawer-content :deep(.el-form) { display: flex; flex-direction: column; height: 100%; }
.env-pair-row { margin-bottom: 8px; display: flex; align-items: center; }
</style>