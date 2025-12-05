<template>
  <div class="structured-logs-root">
    <div class="sidebar">
      <div class="section-title" @click="clearSelection" style="cursor: pointer">Jobs</div>
      <el-tree
        ref="treeRef"
        :data="jobTreeData"
        :props="treeProps"
        node-key="key"
        highlight-current
        :expand-on-click-node="true"
        @node-click="handleNodeClick"
      >
        <template #default="{ data }">
          <div class="tree-node">
            <div class="node-left">
              <el-icon class="status-icon" :style="{ color: statusColor(data.status) }">
                <component :is="statusIconName(data.status)" />
              </el-icon>
              <span class="node-name">{{ data.label }}</span>
              <span class="node-duration" v-if="data.durationText">{{ data.durationText }}</span>
            </div>
            <div class="node-right">
              <span class="node-count" v-if="data.isJob">{{ data.stepCount }}</span>
              <span class="node-count" v-else>{{ data.logCount }}</span>
            </div>
          </div>
        </template>
      </el-tree>
      <div class="section-title" @click="clearStep" style="cursor: pointer">Steps</div>
    </div>
    <div class="logs-panel">
      <div class="panel-header">
        <div class="title"></div>
        <div class="actions">
          <el-button size="small" @click="copyLogs">复制</el-button>
          <el-button size="small" @click="downloadLogs">下载</el-button>
          <el-button size="small" @click="clearSelection">清空选择</el-button>
        </div>
      </div>
      <div class="logs-container" ref="logsRef">
        <div v-if="!visibleLogs.length" class="empty">暂无日志</div>
        <div v-else class="log-list">
          <div v-for="(ln, idx) in visibleLogs" :key="idx" class="log-line">
            <span class="log-time">{{ formatTime(ln.created_at) }}</span>
            <span class="log-text">{{ ln.content }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  structuredLogs: { type: [Array, Object], default: () => [] },
  dagData: { type: Object, default: () => null },
})

const selectedJobId = ref('')
const selectedStepId = ref('')
const logsRef = ref(null)
const treeRef = ref(null)

const jobsArr = computed(() => {
  const d = props.structuredLogs
  if (d && !Array.isArray(d) && Array.isArray(d.jobs)) {
    return d.jobs
  }
  const jobs = {}
  const arr = Array.isArray(d) ? d : []
  for (const b of arr) {
    const j = b?.job || {}
    const st = b?.step || {}
    const key = String(j.id || j.name || '')
    if (!key) continue
    let job = jobs[key]
    if (!job) {
      job = { id: j.id || key, name: j.name || String(j.id || ''), created_at: j.created_at || null, step: [] }
      jobs[key] = job
    }
    const stepKey = String(st.id || st.name || '')
    let stepItem = job.step.find(it => String(it.id || it.name || '') === stepKey)
    if (!stepItem) {
      stepItem = { id: st.id || st.name || '', name: st.name || String(st.id || ''), logs: [] }
      job.step.push(stepItem)
    }
    const logs = Array.isArray(b?.logs) ? b.logs : []
    for (const ln of logs) {
      stepItem.logs.push({ content: ln.content, created_at: ln.created_at })
    }
  }
  return Object.values(jobs)
})

const jobList = computed(() => {
  return jobsArr.value.map(j => ({ id: j.id, name: j.name, stepCount: Array.isArray(j.step) ? j.step.length : 0 }))
})

const stepList = computed(() => {
  const steps = []
  for (const j of jobsArr.value) {
    const key = String(j.id || j.name || '')
    if (!selectedJobId.value || selectedJobId.value === key) {
      for (const st of j.step || []) {
        const logCount = Array.isArray(st.logs) ? st.logs.length : 0
        steps.push({ id: st.id || st.name || '', name: st.name || String(st.id || ''), logCount })
      }
    }
  }
  return steps
})

const jobTreeData = computed(() => {
  const list = []
  for (const j of jobsArr.value) {
    const jobKey = String(j.id || j.name || '')
    const children = []
    const jobStatus = normalizeStatus(j?.status)
    const jobDurationMs = computeJobDurationMs(j)
    const jobDurationText = jobDurationMs > 0 ? formatDuration(jobDurationMs) : ''
    for (const st of j.step || []) {
      const stepKey = String(st.id || st.name || '')
      const logCount = Array.isArray(st.logs) ? st.logs.length : 0
      const stepStatus = normalizeStatus(st?.status)
      const stepDurationMs = computeStepDurationMs(st)
      const stepDurationText = stepDurationMs > 0 ? formatDuration(stepDurationMs) : ''
      children.push({ key: `${jobKey}::${stepKey}`, id: stepKey, label: st.name || stepKey, isStep: true, jobId: jobKey, logCount, status: stepStatus, durationText: stepDurationText })
    }
    list.push({ key: jobKey, id: jobKey, label: j.name || jobKey, isJob: true, stepCount: children.length, children, status: jobStatus, durationText: jobDurationText })
  }
  return list
})

const treeProps = { children: 'children', label: 'label' }

const visibleLogs = computed(() => {
  const list = []
  for (const j of jobsArr.value) {
    const jobKey = String(j.id || j.name || '')
    if (selectedJobId.value && selectedJobId.value !== jobKey) continue
    for (const st of j.step || []) {
      const stepKey = String(st.id || st.name || '')
      if (selectedStepId.value && selectedStepId.value !== stepKey) continue
      const arr = Array.isArray(st.logs) ? st.logs : []
      for (const ln of arr) {
        list.push({ ...ln })
      }
    }
  }
  return list
})

const selectJob = (jid) => { selectedJobId.value = String(jid || '') ; selectedStepId.value = '' }
const selectStep = (sid) => { selectedStepId.value = String(sid || '') }
const handleNodeClick = (data) => {
  if (data && data.isJob) { selectJob(data.id) ; try { treeRef.value?.setCurrentKey(data.key) } catch (_) {} }
  else if (data && data.isStep) { selectedJobId.value = String(data.jobId || '') ; selectStep(data.id) ; try { treeRef.value?.setCurrentKey(data.key) } catch (_) {} }
}
const clearSelection = () => { selectedJobId.value = ''; selectedStepId.value = '' ; try { treeRef.value?.setCurrentKey(null) } catch (_) {} }
const clearStep = () => {
  selectedStepId.value = ''
  try {
    const key = selectedJobId.value || null
    treeRef.value?.setCurrentKey(key)
  } catch (_) {}
}

const formatTime = (t) => { try { return t ? new Date(t).toLocaleTimeString() : '' } catch { return '' } }

const normalizeStatus = (s) => {
  try {
    const v = String(s || '').toLowerCase()
    if (!v) return ''
    if (v.includes('success') || v.includes('succeeded')) return 'success'
    if (v.includes('fail')) return 'failed'
    if (v.includes('cancel')) return 'cancelled'
    if (v.includes('run')) return 'running'
    if (v.includes('queue') || v.includes('pend')) return 'pending'
    return v
  } catch (_) { return '' }
}
const statusIconName = (st) => {
  const v = String(st || '')
  if (v === 'success') return 'SuccessFilled'
  if (v === 'failed') return 'CloseBold'
  if (v === 'cancelled') return 'CircleCloseFilled'
  if (v === 'running') return 'Loading'
  if (v === 'pending') return 'Clock'
  return 'QuestionFilled'
}
const statusColor = (st) => {
  const v = String(st || '')
  if (v === 'success') return 'var(--el-color-success)'
  if (v === 'failed') return 'var(--el-color-danger)'
  if (v === 'cancelled') return 'var(--el-text-color-secondary)'
  if (v === 'running') return 'var(--el-color-primary)'
  if (v === 'pending') return 'var(--el-color-warning)'
  return 'var(--el-text-color-regular)'
}
const pickTime = (obj, names) => {
  for (const n of names) { const v = obj?.[n]; if (v) return v }
  return ''
}
const computeStepDurationMs = (st) => {
  try {
    const start = pickTime(st || {}, ['started_at', 'start_at', 'start_time', 'startTime', 'created_at'])
    const end = pickTime(st || {}, ['finished_at', 'finish_at', 'end_time', 'endTime', 'updated_at'])
    if (start && end) return Math.max(0, new Date(end).getTime() - new Date(start).getTime())
    const logs = Array.isArray(st?.logs) ? st.logs : []
    if (!logs.length) return 0
    const ts = logs.map(l => new Date(l.created_at).getTime()).filter(n => !Number.isNaN(n))
    if (!ts.length) return 0
    const min = Math.min(...ts)
    const max = Math.max(...ts)
    return Math.max(0, max - min)
  } catch (_) { return 0 }
}
const computeJobDurationMs = (j) => {
  try {
    const start = pickTime(j || {}, ['started_at', 'start_at', 'start_time', 'startTime', 'created_at'])
    const end = pickTime(j || {}, ['finished_at', 'finish_at', 'end_time', 'endTime', 'updated_at'])
    if (start && end) return Math.max(0, new Date(end).getTime() - new Date(start).getTime())
    const steps = Array.isArray(j?.step) ? j.step : []
    const ts = []
    for (const st of steps) {
      const logs = Array.isArray(st?.logs) ? st.logs : []
      for (const l of logs) {
        const t = new Date(l.created_at).getTime()
        if (!Number.isNaN(t)) ts.push(t)
      }
    }
    if (!ts.length) return 0
    const min = Math.min(...ts)
    const max = Math.max(...ts)
    return Math.max(0, max - min)
  } catch (_) { return 0 }
}
const formatDuration = (ms) => {
  try {
    const s = Math.floor(ms / 1000)
    const h = Math.floor(s / 3600)
    const m = Math.floor((s % 3600) / 60)
    const sec = s % 60
    if (h) return `${h}h${m}m${sec}s`
    if (m) return `${m}m${sec}s`
    return `${sec}s`
  } catch (_) { return '' }
}

const copyLogs = async () => {
  try {
    const text = (visibleLogs.value || []).map((ln) => ln.content).join('\n')
    await navigator.clipboard.writeText(text)
  } catch (_) {}
}

const downloadLogs = () => {
  try {
    const text = (visibleLogs.value || []).map((ln) => ln.content).join('\n')
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'build_logs.txt'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (_) {}
}
</script>

<style scoped>
.structured-logs-root { display: flex; height: 100%; min-height: 0; width: 100%; }
.sidebar { width: 240px; border-right: 1px solid #ebeef5; padding: 8px; overflow: auto; }
.section-title { font-weight: 600; font-size: 13px; margin: 8px 0; color: #606266; }
.tree-node { display: flex; justify-content: space-between; padding: 6px 8px; border-radius: 6px; }
.node-left { display: flex; align-items: center; gap: 6px; min-width: 0; }
.status-icon { font-size: 16px; }
.node-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.node-duration { margin-left: 6px; color: #909399; font-size: 12px; white-space: nowrap; }
.node-count { color: #909399; font-size: 12px; }
.sidebar :deep(.el-tree-node__content) { border: 1px solid transparent; border-radius: 6px; transition: box-shadow 0.15s ease, border-color 0.15s ease, background-color 0.15s ease; }
.sidebar :deep(.el-tree-node__content:hover) { box-shadow: 0 1px 3px rgba(0,0,0,0.08); border-color: var(--el-border-color-light); background: var(--el-color-primary-light-7); }
.sidebar :deep(.el-tree-node.is-current > .el-tree-node__content) { background: var(--el-color-primary-light-5); border-color: var(--el-color-primary); box-shadow: 0 1px 3px rgba(0,0,0,0.12); }
.logs-panel { flex: 1 1 auto; display: flex; flex-direction: column; min-width: 0; width: 100%; }
.panel-header { display: flex; justify-content: space-between; align-items: center; padding: 8px; border-bottom: 1px solid #ebeef5; }
.title { font-size: 13px; color: #606266; }
.logs-container { flex: 1 1 auto; overflow: auto; background: #0b0d10; color: #d1d5db; padding: 12px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace; }
.log-list { display: flex; flex-direction: column; gap: 2px; }
.log-line { white-space: pre-wrap; word-break: break-word; }
.log-time { color: #9ca3af; margin-right: 8px; }
</style>
