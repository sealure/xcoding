<template>
  <div class="edit-root" ref="rootElRef">
    <div class="toolbar">
      <el-switch v-model="localCollapsed" size="small" inline-prompt active-text="折叠" inactive-text="展开" />
      <el-button size="small" type="primary" @click="resetGraph">重置</el-button>
      <el-button size="small" @click="autoLayout">自动布局</el-button>
      <el-button size="small" @click="fitView">适配视图</el-button>
      <el-divider direction="vertical" />
      <el-button size="small" @click="zoomOut">缩小</el-button>
      <el-button size="small" @click="zoomIn">放大</el-button>
      <el-button size="small" @click="resetZoom">100%</el-button>
      <el-divider direction="vertical" />
      <el-popover placement="bottom" trigger="click" width="360">
        <template #reference>
          <el-button size="small" type="warning" >使用提示</el-button>
        </template>
        <div class="help-popover">
          <div class="help-item"><span class="dot">•</span> 右键 Job 标题添加 Job（串/并行）</div>
          <div class="help-item"><span class="dot">•</span> 右键步骤节点插入/编辑/删除步骤</div>
          <div class="help-item"><span class="dot">•</span> 右键空白位置添加入口 Job</div>
          <div class="help-item"><span class="dot">•</span> 点击依赖边的“+”在两 Job 之间插入</div>
        </div>
      </el-popover>
    </div>

    <div class="status-area">
      <el-alert v-if="loadError" :title="`加载失败：${loadError}`" type="error" show-icon closable class="mb8" />
      <el-alert v-if="renderError" :title="`渲染失败：${renderError}`" type="error" show-icon closable class="mb8" />
      <el-alert v-if="!isLoading && !loadError" :title="`已解析 jobs：${parsedJobCount}`" type="success" show-icon class="mb8" />
      <el-skeleton v-if="isLoading" :rows="3" animated class="mb8" />
    </div>

    <div ref="containerElRef" class="x6-container"></div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { renderJobsPipeline, bindPipelineInteractions, bindConstraintHandlers, createPipelineGraph, unbindPipelineHandlers, fitView as pluginFitView, zoomIn as pluginZoomIn, zoomOut as pluginZoomOut, resetZoom as pluginResetZoom, resetGraph as pluginResetGraph } from './job/JobConrainer/js/JobConrainer.ts'

const peLog = (...args) => { try { console.log('[PipelineEdit]', ...args) } catch (_) {} }
const peWarn = (...args) => { try { console.warn('[PipelineEdit]', ...args) } catch (_) {} }

const props = defineProps({
  collapsed: { type: Boolean, default: false },
  isLoading: { type: Boolean, default: false },
  loadError: { type: String, default: '' },
  renderError: { type: String, default: '' },
  doc: { type: Object, default: () => ({}) },
  collapsedJobs: { type: Object, default: () => ({}) },
  layout: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['update:collapsed', 'show-job', 'show-step', 'update:collapsedJobs', 'create-job', 'step-actions'])

const localCollapsed = computed({
  get: () => props.collapsed,
  set: (v) => emit('update:collapsed', v)
})

const parsedJobCount = computed(() => {
  try { return Object.keys(props.doc?.jobs || {}).length } catch (_) { return 0 }
})

const rootElRef = ref(null)
const containerElRef = ref(null)
let graph
const DEFAULT_WIDTH = 3000
const DEFAULT_HEIGHT = 1366
const resizeObs = ref(null)
let clampLock = false
let renderScheduled = false
let pendingDoc = null
let isRendering = false

// 统一渲染调度，合并来自多个 watcher 的请求
const scheduleRender = (doc) => {
  try {
    pendingDoc = doc || props.doc || {}
    if (!graph) return
    if (renderScheduled) { peLog('render:schedule-coalesce'); return }
    renderScheduled = true
    nextTick(() => {
      try {
        isRendering = true
        renderJobsAsContainers(pendingDoc || {})
        setGraphSize()
        try { graph.centerContent() } catch (_) {}
      } finally {
        isRendering = false
        renderScheduled = false
        pendingDoc = null
      }
    })
  } catch (_) { renderScheduled = false }
}

// helpers
const measureContainer = () => {
  const el = containerElRef.value
  if (!el) return { width: DEFAULT_WIDTH, height: DEFAULT_HEIGHT }
  const rect = el.getBoundingClientRect()
  const width = Math.max(800, Math.floor(rect.width || 0))
  const height = Math.max(480, Math.floor(rect.height || 0))
  return { width, height }
}

// 计算应当占用的可用尺寸：以父 pane 的尺寸减去工具栏、状态区高度
const computeAvailableSize = () => {
  try {
    const root = rootElRef.value
    const container = containerElRef.value
    if (!root || !container) return measureContainer()
    const pane = root.parentElement || root
    const pr = pane.getBoundingClientRect()
    const paneStyle = window.getComputedStyle(pane)
    const vPad = (parseFloat(paneStyle.paddingTop || '0') || 0) + (parseFloat(paneStyle.paddingBottom || '0') || 0)
    const hPad = (parseFloat(paneStyle.paddingLeft || '0') || 0) + (parseFloat(paneStyle.paddingRight || '0') || 0)
    const toolbarH = root.querySelector('.toolbar')?.getBoundingClientRect()?.height || 0
    const statusH = root.querySelector('.status-area')?.getBoundingClientRect()?.height || 0
    const width = Math.max(800, Math.floor(pr.width - hPad))
    const height = Math.max(480, Math.floor(pr.height - vPad - toolbarH - statusH))
    return { width, height }
  } catch (e) {
    return measureContainer()
  }
}

// 设置容器的样式尺寸，使其与父 div 保持一致（扣除顶部区域）
const setContainerStyleSize = () => {
  try {
    const el = containerElRef.value
    if (!el) return
    const { width, height } = computeAvailableSize()
    el.style.height = `${height}px`
    el.style.width = '100%'
  } catch (e) {
    peWarn('设置容器样式尺寸失败', e)
  }
}

const setGraphSize = () => {
  try {
    if (!graph) return
    // 先同步容器样式尺寸，再取实际尺寸用于图缩放
    setContainerStyleSize()
    const { width, height } = measureContainer()
    graph.resize(width, height)
  } catch (e) {
    peWarn('调整图尺寸失败', e)
  }
}

const setupResizeObserver = () => {
  try {
    const el = containerElRef.value
    const root = rootElRef.value
    if (!el) return
    if (typeof ResizeObserver === 'undefined') {
      window.addEventListener('resize', setGraphSize)
      return
    }
    const ro = new ResizeObserver(() => { setGraphSize() })
    ro.observe(el)
    if (root) {
      try { ro.observe(root) } catch (_) {}
      try { if (root.parentElement) ro.observe(root.parentElement) } catch (_) {}
    }
    resizeObs.value = ro
  } catch (e) {
    peWarn('初始化 ResizeObserver 失败', e)
  }
}

const teardownResizeObserver = () => {
  try {
    if (resizeObs.value) {
      resizeObs.value.disconnect()
      resizeObs.value = null
    }
    window.removeEventListener('resize', setGraphSize)
  } catch (e) {
    peWarn('清理 ResizeObserver 失败', e)
  }
}

// normalizeStep / matrixHint / jobIfType 已抽取到 utils

// computeLevels 已抽取到插件入口

const renderJobsAsContainers = (doc) => {
  try {
    const { stats } = renderJobsPipeline({ graph, doc, collapsed: props.collapsed, collapsedJobs: props.collapsedJobs, layout: props.layout })
    try { graph.centerContent() } catch (e) { peWarn('居中内容失败', e) }
    peLog('render:done', { stats, nodes: graph.getNodes().length, edges: graph.getEdges().length })
  } catch (e) {
    peWarn('渲染失败', e)
  } finally {
    isRendering = false
  }
}

const initGraph = () => {
  try {
    const el = containerElRef.value
    if (!el || graph) return
    const { width, height } = measureContainer()
    graph = createPipelineGraph({ container: el, width, height })
    peLog('graph:init', { width, height, hasContainer: !!el })

    bindPipelineInteractions({ graph, emit, getCollapsedJobs: () => props.collapsedJobs })

    // 监听由 JobConrainer 插入的默认边上的“+”事件
    try {
          graph.on('pipeline:create-job', (payload) => {
        try { if (emit) emit('create-job', payload || { anchor: 'after-start' }) } catch (_) {}
      })
    } catch (e) { peWarn('绑定 create-job 事件失败', e) }

    // 监听步骤右键动作事件并上报给父组件
    try {
      graph.on('pipeline:step-actions', (payload) => {
        try { if (emit) emit('step-actions', payload || {}) } catch (_) {}
      })
    } catch (e) { peWarn('绑定 create-job 事件失败', e) }

    // 移除 DEL 标签事件绑定

    try {
      // no minimap
    } catch (e) {
      peWarn('初始化缩略图失败', e)
    }

    // 已移除 Start/End 节点添加事件与去重钩子

    // 限制步骤节点移动范围：只能在所属 Job 容器内
    const setLock = (v) => { clampLock = !!v }
    bindConstraintHandlers(graph, setLock, () => clampLock)
  } catch (e) {
    console.error('图初始化失败', e)
  }
}

const fitView = () => { try { if (graph) pluginFitView(graph) } catch (e) { console.warn('适配视图失败', e) } }
const zoomIn = () => { try { if (graph) pluginZoomIn(graph) } catch (e) { console.warn('放大失败', e) } }
const zoomOut = () => { try { if (graph) pluginZoomOut(graph) } catch (e) { console.warn('缩小失败', e) } }
const resetZoom = () => { try { if (graph) pluginResetZoom(graph) } catch (e) { console.warn('重置缩放失败', e) } }
const resetGraph = () => { try { if (graph && props.doc) pluginResetGraph({ graph, doc: props.doc, collapsed: props.collapsed, collapsedJobs: props.collapsedJobs, layout: props.layout }) } catch (_) {} }

const autoLayout = () => {
  try {
    if (!graph || !props.doc) return
    const layout = { ...(props.layout || {}), autoOrder: true }
    pluginResetGraph({ graph, doc: props.doc, collapsed: props.collapsed, collapsedJobs: props.collapsedJobs, layout })
    pluginFitView(graph)
  } catch (e) {
    peWarn('自动布局失败', e)
  }
}

onMounted(async () => {
  await nextTick()
  peLog('mounted', { hasContainerEl: !!containerElRef.value })
  initGraph()
  setupResizeObserver()
  setGraphSize()
  if (props.doc) scheduleRender(props.doc)
  // 渲染后根据状态区与缩略图高度再适配一次
  await nextTick()
  setGraphSize()
})

watch(() => props.doc, () => { if (graph) { scheduleRender(props.doc) } }, { deep: true })
watch(() => props.collapsed, () => { if (graph && props.doc) { scheduleRender(props.doc) } })
watch(() => props.collapsedJobs, () => { if (graph && props.doc) { scheduleRender(props.doc) } }, { deep: true })
watch(() => props.layout, () => { if (graph && props.doc) { scheduleRender(props.doc) } }, { deep: true })

onBeforeUnmount(() => { try { if (graph) unbindPipelineHandlers(graph) } catch (_) {} teardownResizeObserver() })
</script>

<style scoped>
.edit-root { display: flex; flex-direction: column; height: 100%; width: 100%; flex: 1; box-sizing: border-box; min-width: 0; }
.x6-container {
  flex: 1;
  width: 100%;
  min-height: 480px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  position: relative;
  /* 避免浏览器将 touchstart 视为滚动阻塞事件 */
  touch-action: none;
  overscroll-behavior: contain;
  user-select: none;
  overflow: hidden;
}

.help-popover { font-size: 12px; color: var(--el-text-color-secondary); display: flex; flex-direction: column; gap: 6px; }
.help-item { display: flex; align-items: flex-start; gap: 6px; }
.help-item .dot { color: var(--el-text-color-placeholder); }

</style>
