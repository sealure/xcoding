<template>
  <div class="on-config-root">
    <el-tabs v-model="activeTab" class="on-tabs">
      <el-tab-pane label="workflow_dispatch" name="workflow_dispatch">
        <div class="pane">
          <el-form label-width="120px" class="option-form">
            <el-form-item label="是否打开手动触发">
              <el-checkbox v-model="enableManual">启用</el-checkbox>
            </el-form-item>
          </el-form>
          <!-- 使用新的动态 inputs 编辑器 -->
          <WorkflowDispatchInputs v-if="enableManual" v-model="manualText" />
        </div>
      </el-tab-pane>
      <el-tab-pane label="push" name="push">
        <div class="pane">
          <el-form label-width="100px" class="pane-form">
            <el-form-item label="branches">
              <div>
                <div v-for="(b, i) in formPush.branches" :key="`pb-${i}`" class="list-row">
                  <el-input v-model="formPush.branches[i]" placeholder="例如：main 或 develop" style="width: 70%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeItem(formPush.branches, i)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addItem(formPush.branches)">新增分支</el-button>
              </div>
            </el-form-item>
            <el-form-item label="tags">
              <div>
                <div v-for="(t, i) in formPush.tags" :key="`pt-${i}`" class="list-row">
                  <el-input v-model="formPush.tags[i]" placeholder="例如：v* 或 beta-*" style="width: 70%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeItem(formPush.tags, i)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addItem(formPush.tags)">新增标签</el-button>
              </div>
            </el-form-item>
            <el-form-item label="paths">
              <div>
                <div v-for="(p, i) in formPush.paths" :key="`pp-${i}`" class="list-row">
                  <el-input v-model="formPush.paths[i]" placeholder="例如：src/** 或 docs/**" style="width: 70%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeItem(formPush.paths, i)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addItem(formPush.paths)">新增路径</el-button>
              </div>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="pull_request" name="pull_request">
        <div class="pane">
          <el-form label-width="120px" class="pane-form">
            <el-form-item label="branches">
              <div>
                <div v-for="(b, i) in formPR.branches" :key="`prb-${i}`" class="list-row">
                  <el-input v-model="formPR.branches[i]" placeholder="目标分支（如 main）" style="width: 70%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeItem(formPR.branches, i)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addItem(formPR.branches)">新增分支</el-button>
              </div>
            </el-form-item>
            <el-form-item label="types">
              <el-select v-model="formPR.types" multiple placeholder="选择事件类型" style="width: 70%">
                <el-option v-for="t in prTypeOptions" :key="t" :label="t" :value="t" />
              </el-select>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="schedule" name="schedule">
        <div class="pane">
          <el-form label-width="100px" class="pane-form">
            <el-form-item label="cron">
              <div>
                <div v-for="(c, i) in formSchedule.crons" :key="`sc-${i}`" class="list-row">
                  <el-input v-model="formSchedule.crons[i]" placeholder="例如：0 0 * * *" style="width: 70%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeItem(formSchedule.crons, i)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addItem(formSchedule.crons)">新增计划</el-button>
              </div>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="高级编辑" name="advanced">
        <div class="pane advanced-pane">
          <CodeEditor
            v-model="localText"
            :rows="rows"
            :fit="true"
            language="yaml"
            title="on"
            :placeholder="placeholder"
          />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'
import CodeEditor from '../../PipelineView/common/CodeEditor.vue'
import WorkflowDispatchInputs from './workflow_dispatch_inputs/WorkflowDispatchInputs.vue'
const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  rows: { type: Number, default: 8 }
})
const emit = defineEmits(['update:modelValue'])
const localText = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const activeTab = ref('workflow_dispatch')
const prTypeOptions = [
  'opened','closed','synchronize','reopened','labeled','unlabeled','edited','assigned','unassigned','ready_for_review','converted_to_draft'
]
const baseObj = reactive({})
const enableManual = ref(false)
const manualText = ref('')
const formPush = reactive({ branches: [], tags: [], paths: [] })
const formPR = reactive({ branches: [], types: [] })
const formSchedule = reactive({ crons: [] })

let updatingFromForm = false

const sanitize = (arr) => (arr || []).map((x) => String(x || '').trim()).filter((x) => x.length)
const addItem = (arr) => { arr.push('') }
const removeItem = (arr, i) => { arr.splice(i, 1) }

const updateBaseFromForms = () => {
  try {
    updatingFromForm = true
    const next = {}
    const brs = sanitize(formPush.branches)
    const tgs = sanitize(formPush.tags)
    const pts = sanitize(formPush.paths)
    if (brs.length || tgs.length || pts.length) {
      next.push = {}
      if (brs.length) next.push.branches = brs
      if (tgs.length) next.push.tags = tgs
      if (pts.length) next.push.paths = pts
    }
    const prb = sanitize(formPR.branches)
    const prt = sanitize(formPR.types)
    if (prb.length || prt.length) {
      next.pull_request = {}
      if (prb.length) next.pull_request.branches = prb
      if (prt.length) next.pull_request.types = prt
    }
    const cronList = sanitize(formSchedule.crons)
    if (cronList.length) next.schedule = cronList.map((c) => ({ cron: c }))
    // workflow_dispatch（来自 ManualTrigger 组件的 YAML 文本）
    // workflow_dispatch（由启用勾选决定是否写入）
    if (enableManual.value) {
      try {
        const mt = String(manualText.value || '').trim()
        const obj = mt ? (yamlLoad(mt) || {}) : { workflow_dispatch: {} }
        const wd = obj?.workflow_dispatch ?? obj
        next.workflow_dispatch = (wd && typeof wd === 'object') ? wd : {}
      } catch (_) {
        next.workflow_dispatch = {}
      }
    }
    // 合并保留 baseObj 其他键（如 workflow_call 等）
    let merged = { ...baseObj, ...next }
    // 如果未启用，则从合并结果中移除 workflow_dispatch
    if (!enableManual.value && 'workflow_dispatch' in merged) {
      const { workflow_dispatch, ...rest } = merged
      merged = rest
    }
    Object.keys(baseObj).forEach((k) => { delete baseObj[k] })
    Object.assign(baseObj, merged)
    localText.value = yamlDump(baseObj, { noRefs: true, lineWidth: 120 })
  } finally {
    updatingFromForm = false
  }
}

const parseTextToForms = (text) => {
  if (updatingFromForm) return
  try {
    const obj = yamlLoad(String(text || '')) || {}
    if (typeof obj !== 'object') return
    // 重置并写入 baseObj
    Object.keys(baseObj).forEach((k) => { delete baseObj[k] })
    Object.assign(baseObj, obj)
    // push
    const p = obj?.push || {}
    formPush.branches = Array.isArray(p.branches) ? [...p.branches] : []
    formPush.tags = Array.isArray(p.tags) ? [...p.tags] : []
    formPush.paths = Array.isArray(p.paths) ? [...p.paths] : []
    // pull_request
    const pr = obj?.pull_request || {}
    formPR.branches = Array.isArray(pr.branches) ? [...pr.branches] : []
    formPR.types = Array.isArray(pr.types) ? [...pr.types] : []
    // schedule
    const sc = Array.isArray(obj?.schedule) ? obj.schedule : []
    formSchedule.crons = sc.map((x) => String(x?.cron || '')).filter((x) => x.length)
    // workflow_dispatch -> manualText
    const wd = obj?.workflow_dispatch
    enableManual.value = wd !== undefined
    manualText.value = enableManual.value
      ? yamlDump({ workflow_dispatch: wd ?? {} }, { noRefs: true, lineWidth: 120 })
      : ''
  } catch (_) { /* ignore parse errors */ }
}

watch(() => props.modelValue, (t) => parseTextToForms(t), { immediate: true })
watch(formPush, updateBaseFromForms, { deep: true })
watch(formPR, updateBaseFromForms, { deep: true })
watch(formSchedule, updateBaseFromForms, { deep: true })
watch(manualText, updateBaseFromForms)
watch(enableManual, updateBaseFromForms)
</script>

<style scoped>
.on-config-root { height: 100%; width: 100%; display: flex; flex-direction: column; }
.on-tabs { display: flex; flex-direction: column; height: 100%; }
.on-tabs :deep(.el-tabs__content) { flex: 1; display: flex; height: 100%; }
.on-tabs :deep(.el-tab-pane) { flex: 1; display: flex; height: 100%; }
.pane { flex: 1; display: flex; flex-direction: column; height: 100%; }
.pane-form { flex: 1; overflow-y: auto; }
.list-row { margin-bottom: 8px; display: flex; align-items: center; }
.advanced-pane { padding-top: 4px; }
</style>