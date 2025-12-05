<template>
  <div class="build-snapshot">
    <div class="toolbar">
      <el-radio-group v-model="mode" size="small">
        <el-radio-button label="graph">图视图</el-radio-button>
        <el-radio-button label="yaml">YAML</el-radio-button>
      </el-radio-group>
      <div class="spacer" />
      <el-button v-if="mode==='graph'" size="small" @click="handleFit">适配视图</el-button>
      <template v-if="mode==='yaml'">
        <el-button size="small" @click="copyYaml">复制</el-button>
        <el-button size="small" @click="downloadYaml">下载</el-button>
        <el-button size="small" type="primary" plain @click="toggleCompare">快照对比</el-button>
      </template>
    </div>

    <div v-show="mode==='graph'" class="graph-container" ref="graphContainer">
      <div v-if="!snapshot" class="empty-wrap"><el-empty description="暂无快照数据" /></div>
    </div>
    <div v-show="mode==='yaml'" class="yaml-pane">
      <div v-if="compare" class="compare-split">
        <div class="col">
          <CodeEditor :modelValue="snapshot" :fit="true" language="yaml" theme="dark" title="构建快照（只读）" :showActions="false" :readOnly="true" />
        </div>
        <div class="col">
          <CodeEditor :modelValue="currentYaml" :fit="true" language="yaml" theme="dark" title="当前流水线 YAML（只读）" :showActions="false" :readOnly="true" />
        </div>
      </div>
      <CodeEditor v-else :modelValue="snapshot" :fit="true" language="yaml" theme="dark" title="构建快照（只读）" :showActions="false" :readOnly="true" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { load as yamlLoad } from 'js-yaml'
import { createPipelineGraph, renderJobsPipeline, unbindPipelineHandlers, fitView } from '@/views/ci/dag/components/PipelineView/job/JobConrainer/js/JobConrainer.ts'
import CodeEditor from '@/views/ci/dag/components/PipelineView/common/CodeEditor.vue'
import { getPipeline } from '@/api/ci/pipeline'

const props = defineProps({ snapshot: { type: String, default: '' }, pipelineId: { type: [String, Number], default: '' }, active: { type: Boolean, default: true } })

const mode = ref('graph')
const graphContainer = ref(null)
let graph
let compare = ref(false)
let currentYaml = ref('')
let ro = null
let renderTimer = null
let hasFitted = false

const renderGraph = async () => {
  try {
    if (!graphContainer.value) return
    const doc = yamlLoad(props.snapshot) || {}
    const rect = graphContainer.value.getBoundingClientRect()
    if ((rect.width || 0) < 10 || (rect.height || 0) < 10) { setTimeout(renderGraph, 50); return }
    const width = Math.max(300, Math.floor(rect.width))
    const height = Math.max(300, Math.floor(rect.height))
    if (!graph) {
      graph = createPipelineGraph({ container: graphContainer.value, width, height })
      hasFitted = false
    } else {
      try { graph.clearCells() } catch (_) {}
    }
    if (props.snapshot) {
      renderJobsPipeline({ graph, doc, collapsed: false, collapsedJobs: {} })
      if (!hasFitted) { fitView(graph); hasFitted = true }
    } else {
      try { graph.clearCells() } catch (_) {}
    }
  } catch (e) {
    console.error('渲染构建快照失败:', e)
  }
}

const handleFit = () => { try { if (graph) fitView(graph) } catch (_) {} }

const copyYaml = async () => { try { await navigator.clipboard.writeText(props.snapshot || '') } catch (_) {} }
const downloadYaml = () => {
  try {
    const blob = new Blob([props.snapshot || ''], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'build-snapshot.yaml'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (_) {}
}

const toggleCompare = async () => {
  try {
    compare.value = !compare.value
    if (compare.value && !currentYaml.value) {
      await loadCurrentPipelineYaml()
    }
  } catch (_) {}
}

const loadCurrentPipelineYaml = async () => {
  try {
    let pid = String(props.pipelineId || '')
    if (!pid) {
      const path = window.location?.pathname || ''
      const m = path.match(/\/pipelines\/(\d+)/)
      pid = m ? m[1] : ''
    }
    if (!pid) { currentYaml.value = ''; return }
    const res = await getPipeline(pid)
    const yaml = res?.pipeline?.workflow_yaml || res?.workflow_yaml || ''
    currentYaml.value = String(yaml || '')
  } catch (_) { currentYaml.value = '' }
}

onMounted(() => { nextTick(renderGraph); try { if (graphContainer.value) { ro = new ResizeObserver(() => { try { const r = graphContainer.value.getBoundingClientRect(); const w = Math.max(300, Math.floor(r.width || 0)); const h = Math.max(300, Math.floor(r.height || 0)); if (graph) { graph.resize(w, h) } } catch (_) {} }); ro.observe(graphContainer.value) } } catch (_) {} })
watch(() => props.active, async (isActive) => { try { if (isActive && mode.value === 'graph') { await nextTick(); if (renderTimer) { clearTimeout(renderTimer) }; renderTimer = setTimeout(() => { renderGraph() }, 50) } } catch (_) {} })
onBeforeUnmount(() => { try { if (graph) unbindPipelineHandlers(graph) } catch (_) {} try { if (ro) ro.disconnect() } catch (_) {} ro = null })
watch(() => props.snapshot, async () => { hasFitted = false; await nextTick(); renderGraph() })
watch(mode, async (m) => { if (props.active && m === 'graph') { hasFitted = false; await nextTick(); renderGraph() } })
</script>

<style scoped>
.build-snapshot { padding: 0; display: flex; flex-direction: column; gap: 8px; flex: 1 1 auto; min-height: 0; width: 100%; }
.toolbar { display: flex; align-items: center; gap: 8px; }
.spacer { flex: 1; }
.graph-container { width: 100%; flex: 1 1 auto; min-height: 0; height: 100%; background: #fff; border: 1px solid #ebeef5; border-radius: 6px; }
.empty-wrap { padding: 40px 0; }
.yaml-pane { display: flex; flex-direction: column; gap: 8px; min-width: 0; min-height: 0; height: auto; flex: 1 1 auto; }
.compare-split { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; height: 100%; }
.col { display: flex; flex-direction: column; min-width: 0; min-height: 0; height: 100%; }
.yaml-content {
  font-family: 'Fira Code', monospace;
  white-space: pre-wrap;
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 1.5;
  color: #333;
}
</style>
