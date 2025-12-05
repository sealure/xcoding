<template>
  <div class="pipeline-job-container">
    <div class="pipeline-job-main">
      

      <el-tabs v-model="activeMainTab">
        <el-tab-pane label="Pipeline编辑视图" name="container">
          <PipelineEdit
            v-model:collapsed="collapsed"
            :isLoading="isLoading"
            :loadError="loadError"
            :renderError="renderError"
            :doc="lastDoc || {}"
            :collapsedJobs="collapsedJobs"
            :layout="computedLayout"
            @show-job="onShowJob"
            @show-step="onShowStep"
            @create-job="onCreateJobFromEdge"
            @step-actions="showStepMenu"
            @delete-step="onDeleteStepFromGraph"
            @delete-job="onDeleteJobFromGraph"
            @update:collapsedJobs="(v) => (collapsedJobs.value = v)"
          />
        </el-tab-pane>

        <el-tab-pane label="流水线全局配置" name="edit">
          <PipelineGlobalConfig :doc="lastDoc || {}" ref="globalConfigRef" @patch="onGlobalPatched" />
        </el-tab-pane>

        <el-tab-pane label="YAML视图" name="yaml-view">
          <YamlViewPane
            :originalText="originalYamlText"
            v-model:doc="lastDoc"
          />
        </el-tab-pane>
      </el-tabs>

      <StepDetailDrawer
        v-model="drawerVisible"
        :step="drawerStep"
        :doc="lastDoc || {}"
        :jobId="drawerStepJobId"
        :stepIndex="drawerStepIndex"
        @save="onStepSaved"
      />
  <JobDetialDrawer
    v-model="jobDrawerVisible"
    :jobId="jobDrawerId"
    :doc="lastDoc || {}"
    :strategyPlaceholder="strategyPlaceholder"
    @save="onJobSaved"
    @delete="onJobDeleted"
  />
      <AddJobModeMenu
        v-model="addJobDialogVisible"
        :x="addMenuX"
        :y="addMenuY"
        @confirm="onAddJobModeChosen"
        @cancel="() => (addJobDialogVisible = false)"
      />
      <StepActionMenu
        v-model="stepMenuVisible"
        :x="stepMenuX"
        :y="stepMenuY"
        :jobId="stepMenuJobId"
        :stepIndex="stepMenuStepIndex"
        :step="stepMenuStep"
        @insert="onInsertStep"
        @edit="onEditStep"
        @delete="onDeleteStep"
        @cancel="() => (stepMenuVisible = false)"
      />
    </div>
  </div>
  
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'
import PipelineEdit from './components/PipelineView/PipelineEdit.vue'
import PipelineGlobalConfig from './components/global/components/PipelineGlobalConfig.vue'
import StepDetailDrawer from './components/PipelineView/step/shell/StepDetailDrawer.vue'
import JobDetialDrawer from './components/PipelineView/job/JobDetialDrawer/JobDetialDrawer.vue'
import AddJobModeMenu from './components/AddJobModeMenu.vue'
import StepActionMenu from './components/StepActionMenu.vue'
import { ElMessage } from 'element-plus'
import YamlViewPane from './components/YamlView/YamlViewPane.vue'
import { createPipeline, updatePipeline } from '@/api/ci/pipeline'
import { useProjectStore } from '@/stores/project'

const pjLog = (...args) => { try { console.log('[PipelineJob]', ...args) } catch (_) {} }
const pjWarn = (...args) => { try { console.warn('[PipelineJob]', ...args) } catch (_) {} }

const projectStore = useProjectStore()

// 将全局配置键置顶；并在每个 job 中将 steps 放到最后
const GLOBAL_FIRST_KEYS = ['name', 'run-name', 'on', 'env', 'permissions', 'concurrency', 'defaults']
const reorderYamlDoc = (doc) => {
  try {
    if (!doc || typeof doc !== 'object' || Array.isArray(doc)) return doc
    const next = {}
    const keys = Object.keys(doc)
    // 顶层：全局配置置顶
    GLOBAL_FIRST_KEYS.forEach((k) => { if (k in doc) next[k] = doc[k] })
    // 顶层：jobs 排在全局配置之后，并重排内部 steps 顺序
    if ('jobs' in doc && doc.jobs && typeof doc.jobs === 'object') {
      const jobsNext = {}
      Object.keys(doc.jobs).forEach((jid) => {
        const job = doc.jobs[jid]
        if (!job || typeof job !== 'object') { jobsNext[jid] = job; return }
        const jobNext = {}
        const jKeys = Object.keys(job)
        // 先放除 steps 外的键，保持原有相对顺序
        jKeys.forEach((jk) => { if (jk !== 'steps') jobNext[jk] = job[jk] })
        // 最后追加 steps
        if ('steps' in job) jobNext.steps = job.steps
        jobsNext[jid] = jobNext
      })
      next.jobs = jobsNext
    }
    // 顶层：其余未放入的键按原顺序追加
    keys.forEach((k) => { if (!(k in next)) next[k] = doc[k] })
    return next
  } catch (_) { return doc }
}

// 接收父组件传入的后端 YAML 文本与流水线ID（详情页编辑/保存）
const props = defineProps({
  serverYaml: { type: String, default: '' },
  pipelineId: { type: [String, Number], default: '' },
  // 由详情页传入的流水线名称，用于同步 YAML 顶层 name
  pipelineName: { type: String, default: '' }
})

const selectedFile = ref('')
const workflowOptions = ref([])
const activeMainTab = ref('container')
const collapsed = ref(false)
const isLoading = ref(false)
const loadError = ref('')
const renderError = ref('')
const lastDoc = ref(null)
let collapsedJobs = ref({})
const drawerVisible = ref(false)
const drawerStep = ref(null)
const drawerStepJobId = ref('')
const drawerStepIndex = ref(-1)

const jobDrawerVisible = ref(false)
const jobDrawerId = ref('')
const strategyPlaceholder = '例如：{\n  "matrix": { \n    "node-version": ["16.x", "18.x"] \n  }\n}'

// “+” 选择串并行弹窗
const addJobDialogVisible = ref(false)
const addMenuX = ref(0)
const addMenuY = ref(0)

// 画布布局参数：增大列间距与左侧留白，使“开始”更远，边更整齐
const computedLayout = computed(() => ({
  // 更宽的列间距与更大的基础 X，扩大各列与“开始”的间距
  columnSpacing: 560,
  baseX: 320,
  baseY: 160,
  // 增大同列上下 job 的间距
  columnRowGap: 80,
  // 显著增加左侧留白，使“开始”离得更远
  startOffset: 340,
  // 右侧留白也适当加大，避免“结束”挤在右侧
  endOffset: 8,
  startEndHeight: 64,
  // 默认开启自动布局
  autoOrder: true,
}))
const pendingAddAnchor = ref('after-start')
const pendingPrevJobId = ref('')
const pendingNextJobId = ref('')
// 步骤右键菜单状态
const stepMenuVisible = ref(false)
const stepMenuX = ref(0)
const stepMenuY = ref(0)
const stepMenuJobId = ref('')
const stepMenuStepIndex = ref(-1)
const stepMenuStep = ref({})

const globalConfigRef = ref(null)
// 应用来自 GlobalConfigTabs 的增量补丁到 YAML 文档
const onGlobalPatched = (payload) => {
  try {
    if (!payload || typeof payload !== 'object') return
    const base = lastDoc.value || {}
    const next = { ...base }
    let changed = false
    const setOrDelete = (key, val) => {
      if (val === undefined) return
      if (val) {
        const prev = next[key]
        const same = (() => {
          try { return JSON.stringify(prev) === JSON.stringify(val) } catch { return prev === val }
        })()
        if (!same) { next[key] = val; changed = true }
      } else {
        if (key in next) { delete next[key]; changed = true }
      }
    }
    // 基础字段
    setOrDelete('name', payload.name)
    setOrDelete('run-name', payload.runName)
    // 结构化字段
    setOrDelete('on', payload.on)
    setOrDelete('env', payload.env)
    setOrDelete('permissions', payload.permissions)
    setOrDelete('concurrency', payload.concurrency)
    setOrDelete('defaults', payload.defaults)
    if (changed) lastDoc.value = next
  } catch (_) { /* ignore */ }
}
const applyGlobalPayload = (payload) => {
  try {
    if (!lastDoc.value || typeof lastDoc.value !== 'object') { ElMessage.error('当前没有已加载的 YAML'); return false }
    const base = lastDoc.value
    const next = { ...base }
    if (payload.name !== undefined) { if (payload.name) next.name = payload.name; else delete next.name }
    if (payload.runName !== undefined) { if (payload.runName) next['run-name'] = payload.runName; else delete next['run-name'] }
    if (payload.on !== undefined) { if (payload.on) next.on = payload.on; else delete next.on }
    if (payload.env !== undefined) { if (payload.env) next.env = payload.env; else delete next.env }
    if (payload.permissions !== undefined) { if (payload.permissions) next.permissions = payload.permissions; else delete next.permissions }
    if (payload.concurrency !== undefined) { if (payload.concurrency) next.concurrency = payload.concurrency; else delete next.concurrency }
    if (payload.defaults !== undefined) { if (payload.defaults) next.defaults = payload.defaults; else delete next.defaults }
    // 强制替换引用以触发所有依赖 doc 的深度监听（如 YAML 视图）
    lastDoc.value = next
    return true
  } catch (e) {
    console.error('应用全局配置失败', e)
    return false
  }
}

const onShowJob = ({ jobId }) => {
  try {
    jobDrawerId.value = jobId || ''
    jobDrawerVisible.value = !!jobDrawerId.value
  } catch (e) {
    pjWarn('打开 Job 抽屉失败', e)
  }
}
const onShowStep = ({ step, jobId, stepIndex }) => {
  try {
    drawerStep.value = step || null
    drawerStepJobId.value = jobId || ''
    drawerStepIndex.value = (typeof stepIndex === 'number' ? stepIndex : -1)
    drawerVisible.value = !!drawerStep.value
  } catch (e) {
    pjWarn('打开 Step 抽屉失败', e)
  }
}
const onStepSaved = (nextDocOrStep) => {
  try {
    // 如果是完整文档则更新
    if (nextDocOrStep && typeof nextDocOrStep === 'object' && nextDocOrStep.jobs) {
      lastDoc.value = nextDocOrStep
    } else {
      // 否则在当前上下文替换 step
      const jid = drawerStepJobId.value
      const idx = drawerStepIndex.value
      const next = { ...(lastDoc.value || {}) }
      if (jid && typeof idx === 'number' && idx >= 0) {
        if (!next.jobs) next.jobs = {}
        const job = { ...(next.jobs[jid] || {}) }
        const steps = Array.isArray(job.steps) ? [...job.steps] : []
        if (idx < steps.length) steps[idx] = nextDocOrStep; else steps.push(nextDocOrStep)
        job.steps = steps
        next.jobs[jid] = job
        lastDoc.value = next
      }
    }
    drawerVisible.value = false
    ElMessage.success('步骤配置已保存')
  } catch (e) {
    pjWarn('保存 Step 返回处理失败', e)
  }
}
const onJobSaved = (nextDoc) => {
  try {
    lastDoc.value = nextDoc || lastDoc.value
    jobDrawerVisible.value = false
    ElMessage.success('Job 配置已保存')
  } catch (e) {
    pjWarn('保存 Job 返回处理失败', e)
  }
}

// 删除 Job：接收抽屉返回的 nextDoc 并刷新视图
const onJobDeleted = (nextDoc) => {
  try {
    lastDoc.value = nextDoc || lastDoc.value
    jobDrawerVisible.value = false
    // 同步 collapsedJobs：移除被删除的键，避免残留状态
    try {
      const cur = { ...(collapsedJobs.value || {}) }
      if (jobDrawerId.value && jobDrawerId.value in cur) delete cur[jobDrawerId.value]
      collapsedJobs.value = cur
    } catch (_) {}
    ElMessage.success('Job 已删除')
  } catch (e) {
    pjWarn('删除 Job 返回处理失败', e)
    ElMessage.error('删除 Job 失败')
  }
}

// 点击默认边上的“+”：弹出串并行选择
const onCreateJobFromEdge = (payload) => {
  try {
    pendingAddAnchor.value = payload?.anchor || 'after-start'
    pendingPrevJobId.value = payload?.prevJobId || ''
    pendingNextJobId.value = payload?.nextJobId || ''
    addMenuX.value = Number(payload?.x || 0)
    addMenuY.value = Number(payload?.y || 0)
    addJobDialogVisible.value = true
  } catch (e) {
    pjWarn('打开添加 Job 模式选择失败', e)
  }
}

const showStepMenu = (payload) => {
  try {
    stepMenuX.value = Number(payload?.x || 0)
    stepMenuY.value = Number(payload?.y || 0)
    stepMenuJobId.value = String(payload?.jobId || '')
    stepMenuStepIndex.value = Number(payload?.stepIndex ?? -1)
    stepMenuStep.value = payload?.step || {}
    stepMenuVisible.value = true
  } catch (e) { pjWarn('显示步骤菜单失败', e) }
}

// 根据选择创建 Job：serial（插入串行依赖）或 parallel（保持并行结构）
const onAddJobModeChosen = ({ mode }) => {
  try {
    const base = lastDoc.value || {}
    const next = { ...base }
    if (!next.jobs) next.jobs = {}

    // 生成唯一 Job ID
    let idx = 1
    let jid = `job-${idx}`
    while (next.jobs[jid]) { idx++; jid = `job-${idx}` }

    // 新建 Job 默认包含 runs-on 与 container，保证 YAML 初始即有这两个字段
    const newJob = { name: `Job ${idx}`, 'runs-on': 'ubuntu-latest', container: 'ubuntu:latest', steps: [{ name: 'echo hello', run: 'echo "Hello"' }] }

    // 根据 anchor 决定插入位置与依赖调整
    const prevId = pendingPrevJobId.value || ''
    const nextId = pendingNextJobId.value || ''

    const toList = (raw) => Array.isArray(raw) ? raw : (raw ? [raw] : [])
    const toPref = (list) => (list.length <= 1 ? (list[0] || undefined) : list)

    if (pendingAddAnchor.value === 'between') {
      if (nextId) {
        // 在 prevId -> nextId 之间插入
        if (mode === 'serial') {
          // 新 Job 依赖 prevId（若存在），nextId 改为依赖新 Job
          if (prevId) newJob.needs = prevId
          const nextJob = { ...(next.jobs[nextId] || {}) }
          const needsList = toList(nextJob.needs)
          let replaced = []
          if (prevId && needsList.includes(prevId)) {
            replaced = needsList.map((n) => (n === prevId ? jid : n))
          } else if (needsList.length > 0) {
            replaced = [...needsList, jid]
          } else {
            replaced = [jid]
          }
          nextJob.needs = toPref(replaced) || undefined
          next.jobs[nextId] = nextJob
        } else {
          // 并行：保持 nextId 的依赖不变，新 Job 与 nextId 并行（依赖相同前驱）
          const nextJob = { ...(next.jobs[nextId] || {}) }
          const nextPrev = toList(nextJob.needs)
          if (prevId) {
            newJob.needs = prevId
          } else if (nextPrev.length > 0) {
            newJob.needs = toPref(nextPrev)
          }
          // 不修改 nextId 的依赖
        }
      } else if (prevId) {
        // 边：出口 Job -> End 的“+”
        if (mode === 'serial') {
          // 串行：新 Job 依赖于该出口 Job
          newJob.needs = prevId
        } else {
          // 并行：新 Job 与该出口 Job 并行（复制其前驱）
          const prevJob = { ...(next.jobs[prevId] || {}) }
          const prevNeeds = toList(prevJob.needs)
          if (prevNeeds.length > 0) newJob.needs = toPref(prevNeeds)
        }
      } else {
        // between 但没有明确前后：视为新入口 Job（无依赖）
        // newJob.needs 保持未定义
      }
    } else {
      // after-start：为空管道的开始之后追加
      // 空管道场景下，新 Job 作为入口 Job（无依赖）
      if (mode === 'serial') {
        // 空管道时与并行一致：无依赖
      } else {
        // 并行：同样无依赖
      }
    }

    next.jobs[jid] = newJob
    lastDoc.value = next
    ElMessage.success(mode === 'serial' ? '已创建串行 Job' : '已创建并行 Job')
  } catch (e) {
    pjWarn('根据选择创建 Job 失败', e)
    ElMessage.error('创建 Job 失败')
  } finally {
    addJobDialogVisible.value = false
  }
}

const pluginToStep = (plugin) => {
  const key = plugin?.key
  switch (key) {
    case 'shell':
      return { name: '执行 Shell 脚本', run: 'echo "Hello"' }
    case 'echo':
      return { name: '打印消息', run: 'echo "Message"' }
    case 'sleep':
      return { name: '睡眠', run: 'sleep 5' }
    case 'upload-artifact':
      return { name: '上传制品', uses: 'actions/upload-artifact@v3', with: { name: 'artifact', path: 'dist/' } }
    case 'download-artifact':
      return { name: '下载制品', uses: 'actions/download-artifact@v3', with: { name: 'artifact' } }
    default:
      return { name: plugin?.name || '新步骤', run: 'echo "New Step"' }
  }
}

const onInsertStep = ({ position, plugin, jobId, stepIndex }) => {
  try {
    const base = lastDoc.value || {}
    const next = { ...base }
    if (!next.jobs) next.jobs = {}
    const job = { ...(next.jobs[jobId] || {}) }
    const steps = Array.isArray(job.steps) ? [...job.steps] : []
    const newStep = pluginToStep(plugin)
    const idx = Math.max(0, Number(stepIndex))
    const insertAt = position === 'insert-above' ? idx : (idx + 1)
    steps.splice(insertAt, 0, newStep)
    job.steps = steps
    next.jobs[jobId] = job
    lastDoc.value = next
    stepMenuVisible.value = false
    ElMessage.success('已插入步骤')
  } catch (e) { pjWarn('插入步骤失败', e) }
}

const onEditStep = ({ jobId, stepIndex, step }) => {
  try {
    drawerStep.value = step || {}
    drawerStepJobId.value = jobId
    drawerStepIndex.value = Number(stepIndex)
    drawerVisible.value = true
    stepMenuVisible.value = false
  } catch (e) { pjWarn('打开步骤编辑失败', e) }
}

const onDeleteStep = ({ jobId, stepIndex }) => {
  try {
    const base = lastDoc.value || {}
    const next = { ...base }
    if (!next.jobs) next.jobs = {}
    const job = { ...(next.jobs[jobId] || {}) }
    const steps = Array.isArray(job.steps) ? [...job.steps] : []
    const idx = Math.max(0, Number(stepIndex))
    if (idx >= 0 && idx < steps.length) {
      steps.splice(idx, 1)
      job.steps = steps
      next.jobs[jobId] = job
      lastDoc.value = next
      ElMessage.success('已删除步骤')
    }
    stepMenuVisible.value = false
  } catch (e) { pjWarn('删除步骤失败', e) }
}

// 来自图的删除步骤事件（DEL 标签），复用逻辑
const onDeleteStepFromGraph = ({ jobId, stepIndex }) => {
  try {
    onDeleteStep({ jobId, stepIndex })
  } catch (e) { pjWarn('来自图的删除步骤失败', e) }
}

// 来自图的删除 Job 事件（DEL 标签），委托给抽屉的删除逻辑以保持一致的清理 needs
const onDeleteJobFromGraph = ({ jobId }) => {
  try {
    const jid = String(jobId || '')
    if (!jid) return
    const base = lastDoc.value || {}
    const next = { ...base }
    const jobs = { ...(next.jobs || {}) }
    // 删除目标 Job
    delete jobs[jid]
    // 清理所有引用该 Job 的 needs（与 JobDetialDrawer.onDelete 保持一致）
    Object.keys(jobs).forEach((k) => {
      const job = { ...(jobs[k] || {}) }
      const raw = job.needs
      const list = Array.isArray(raw) ? raw : (raw ? [raw] : [])
      const filtered = list.filter((n) => n !== jid)
      if (filtered.length === 0) {
        delete job.needs
      } else if (filtered.length === 1) {
        job.needs = filtered[0]
      } else {
        job.needs = filtered
      }
      jobs[k] = job
    })
    next.jobs = jobs
    lastDoc.value = next
    // 同步折叠状态
    try {
      const cur = { ...(collapsedJobs.value || {}) }
      if (jid && jid in cur) delete cur[jid]
      collapsedJobs.value = cur
    } catch (_) {}
    ElMessage.success('Job 已删除')
  } catch (e) { pjWarn('来自图的删除 Job 失败', e); ElMessage.error('删除 Job 失败') }
}

const exportYaml = () => {
  try {
    if (!lastDoc.value || typeof lastDoc.value !== 'object') {
      ElMessage.error('没有可导出的 YAML 文档')
      return
    }
    const sorted = reorderYamlDoc(lastDoc.value)
    const yamlText = yamlDump(sorted, { lineWidth: 120, noRefs: true })
    const blob = new Blob([yamlText], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const base = (selectedFile.value || 'pipeline.yml').split('/').pop() || 'pipeline.yml'
    const name = base.replace(/\.ya?ml$/i, '') + '.export.yml'
    a.href = url
    a.download = name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success(`已导出：${name}`)
  } catch (e) {
    console.error('导出 YAML 失败', e)
    ElMessage.error('导出失败，请重试')
  }
}

// 重复定义移除，统一保留下方基于 proto 的实现

const loadYaml = async () => {
  isLoading.value = true
  loadError.value = ''
  try {
    pjLog('yaml:load-start', { file: selectedFile.value })
    const res = await fetch(selectedFile.value)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const text = await res.text()
    const doc = yamlLoad(text) || {}
    if (!doc || typeof doc !== 'object') throw new Error('YAML 解析结果为空或格式不正确')
    lastDoc.value = doc
    originalYamlText.value = yamlDump(reorderYamlDoc(doc), { lineWidth: 120, noRefs: true })
    const jobsObj = doc.jobs || {}
    const ids = Object.keys(jobsObj)
    pjLog('yaml:parsed', { jobs: ids.length })
    const nextCollapsed = { ...(collapsedJobs.value || {}) }
    ids.forEach((jid) => { if (!(jid in nextCollapsed)) nextCollapsed[jid] = false })
    collapsedJobs.value = nextCollapsed
  } catch (e) {
    console.error('加载 YAML 失败', e)
    loadError.value = e.message || String(e)
    ElMessage.error(`加载 YAML 失败：${e.message || e}`)
  } finally {
    isLoading.value = false
  }
}

// 新建流水线：将当前 YAML 保存到后端
const onCreatePipeline = async () => {
  try {
    const pid = projectStore.selectedProject?.id
    if (!pid) { ElMessage.warning('请先选择项目'); return }
    // 新建流水线：默认空 YAML，名称自动生成，避免与现有名称冲突
    const name = `pipeline_${Date.now()}`
    const yamlText = ''
    await createPipeline({ projectId: pid, name, workflow_yaml: yamlText, is_active: true })
    ElMessage.success('流水线已创建')
  } catch (e) {
    ElMessage.error(`创建流水线失败：${e?.message || e}`)
  }
}

const ensureNameSynced = (doc) => {
  try {
    const d = { ...(doc || {}) }
    const pn = (props.pipelineName || '').trim()
    if (pn) d.name = pn
    return d
  } catch (_) { return doc }
}

const onSavePipeline = async () => {
  try {
    if (!props.pipelineId) { ElMessage.warning('缺少流水线ID'); return }
    // 合并全局配置到文档
    try {
      const gc = globalConfigRef.value
      if (gc && typeof gc.collectPayload === 'function') {
        const payload = gc.collectPayload()
        if (payload) {
          const ok = applyGlobalPayload(payload)
          if (!ok) { ElMessage.error('全局配置合并失败'); return }
        }
      }
    } catch (_) {}
    let doc = reorderYamlDoc(lastDoc.value || {})
    // 保存前强制同步顶层 name 与流水线名称
    doc = ensureNameSynced(doc)
    const yamlText = yamlDump(doc, { lineWidth: 120, noRefs: true })
    await updatePipeline(String(props.pipelineId), { workflow_yaml: yamlText })
    // 同步原始文本，便于 YAML 视图对比
    originalYamlText.value = yamlText
    ElMessage.success('流水线已保存（含全局配置）')
  } catch (e) {
    ElMessage.error(`保存失败：${e?.message || e}`)
  }
}

onMounted(async () => {
  // 详情页传入的后端 YAML 优先使用；若为空，则保持空白视图
  if (props.serverYaml !== undefined) {
    const t = (props.serverYaml || '').trim()
    if (t) {
      try {
        const doc = yamlLoad(t) || {}
        if (!doc || typeof doc !== 'object') throw new Error('后端 YAML 解析为空或不合法')
        const parsed = ensureNameSynced(doc)
        const jobsObj = parsed.jobs || {}
        let nextDoc = { ...parsed }
        if (!Object.keys(jobsObj).length) {
          const jid = 'job-1'
          const defaultJob = { name: 'Job 1', 'runs-on': 'ubuntu-latest', container: 'ubuntu:latest', steps: [{ name: 'echo hello', run: 'echo "Hello"' }] }
          nextDoc.jobs = { [jid]: defaultJob }
          const nextCollapsed = { ...(collapsedJobs.value || {}) }
          nextCollapsed[jid] = false
          collapsedJobs.value = nextCollapsed
        } else {
          const ids = Object.keys(jobsObj)
          const nextCollapsed = { ...(collapsedJobs.value || {}) }
          ids.forEach((jid) => { if (!(jid in nextCollapsed)) nextCollapsed[jid] = false })
          collapsedJobs.value = nextCollapsed
        }
        lastDoc.value = nextDoc
        originalYamlText.value = yamlDump(reorderYamlDoc(doc), { lineWidth: 120, noRefs: true })
        // 加载列表用于切换示例，但不覆盖当前文档
        await loadWorkflowList(false)
      } catch (e) {
        pjWarn('解析后端 YAML 失败，回退到本地示例', e)
        await loadWorkflowList(true)
      }
    } else {
      // 空 YAML：保持编辑视图为空，不自动加载示例
      // 若空，初始化一个包含顶层 name 的最小文档
      const pn = (props.pipelineName || '').trim()
      lastDoc.value = pn ? { name: pn } : null
      originalYamlText.value = ''
      await loadWorkflowList(false)
    }
  } else {
    // 无后端传入：作为编辑页面，加载示例以供选择
    await loadWorkflowList(true)
  }
})

watch(() => props.serverYaml, async (t) => {
  try {
    const s = String(t || '').trim()
    if (s) {
      const doc = yamlLoad(s) || {}
      if (!doc || typeof doc !== 'object') return
      const parsed = ensureNameSynced(doc)
      const jobsObj = parsed.jobs || {}
      let nextDoc = { ...parsed }
      if (!Object.keys(jobsObj).length) {
        const jid = 'job-1'
        const defaultJob = { name: 'Job 1', 'runs-on': 'ubuntu-latest', container: 'ubuntu:latest', steps: [{ name: 'echo hello', run: 'echo "Hello"' }] }
        nextDoc.jobs = { [jid]: defaultJob }
        const nextCollapsed = { ...(collapsedJobs.value || {}) }
        nextCollapsed[jid] = false
        collapsedJobs.value = nextCollapsed
      } else {
        const ids = Object.keys(jobsObj)
        const nextCollapsed = { ...(collapsedJobs.value || {}) }
        ids.forEach((jid) => { if (!(jid in nextCollapsed)) nextCollapsed[jid] = false })
        collapsedJobs.value = nextCollapsed
      }
      lastDoc.value = nextDoc
      originalYamlText.value = yamlDump(reorderYamlDoc(doc), { lineWidth: 120, noRefs: true })
    } else {
      const pn = (props.pipelineName || '').trim()
      const jid = 'job-1'
      const defaultJob = { name: 'Job 1', 'runs-on': 'ubuntu-latest', container: 'ubuntu:latest', steps: [{ name: 'echo hello', run: 'echo "Hello"' }] }
      lastDoc.value = pn ? { name: pn, jobs: { [jid]: defaultJob } } : { jobs: { [jid]: defaultJob } }
      const nextCollapsed = { ...(collapsedJobs.value || {}) }
      nextCollapsed[jid] = false
      collapsedJobs.value = nextCollapsed
      originalYamlText.value = ''
    }
  } catch (_) {}
})

// YAML 原始文本（供子组件对比使用）
const originalYamlText = ref('')

// 加载可选的工作流 YAML 列表
const loadWorkflowList = async (autoLoad = true) => {
  try {
    const res = await fetch('/workflows/index.json')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const list = await res.json()
    if (Array.isArray(list)) {
      workflowOptions.value = list
    }
  } catch (e) {
    // 回退到内置列表
    workflowOptions.value = [
      { label: 'example_pipeline.yml', path: '/workflows/example_pipeline.yml' },
      { label: 'example.yml', path: '/workflows/example.yml' }
    ]
  } finally {
    // 设置默认选项并加载
    const first = workflowOptions.value[0]
    if (first && !selectedFile.value) selectedFile.value = first.path
    if (autoLoad) await loadYaml()
  }
}

// 暴露给父组件调用的操作方法（用于外层卡片 header 上的按钮）
defineExpose({
  createPipeline: onCreatePipeline,
  savePipeline: onSavePipeline,
  exportYaml
})
</script>

<style scoped>
 .pipeline-job-container { padding: 5px; box-sizing: border-box; width: 100%; height: 100%; display: flex; flex-direction: column; flex: 1; min-height: 0; }
 .pipeline-job-main { display: flex; flex-direction: column; width: 100%; height: 100%; flex: 1; box-sizing: border-box; min-height: 0; }
 .drawer-content { padding: 8px; }
 .run-block { white-space: pre-wrap; font-family: var(--el-font-family); background: #f8f8f8; padding: 8px; border-radius: 4px; }
 .env-pair-row { margin-bottom: 8px; display: flex; align-items: center; }

 /* Tabs 充满剩余空间，并去除多余 padding */
 .pipeline-job-main :deep(.el-tabs) { display: flex; flex-direction: column; flex: 1; height: 100%; min-width: 0; min-height: 0; }
 .pipeline-job-main :deep(.el-tabs__content) { flex: 1; display: flex; height: 100%; padding: 0; box-sizing: border-box; min-width: 0; min-height: 0; }
 .pipeline-job-main :deep(.el-tab-pane) { flex: 1; display: flex; height: 100%; padding: 0; box-sizing: border-box; min-width: 0; min-height: 0; }

/* YAML 视图样式由子组件提供 */
</style>
