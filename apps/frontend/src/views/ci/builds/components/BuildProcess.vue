<template>
  <div class="build-process">
  <div class="debug-info" style="padding: 8px; background: #f9fafb; margin-bottom: 12px; border: 1px solid #ebeef5; border-radius: 6px;">
      <div style="font-size:12px;color:#909399">Jobs: {{ dagData?.jobs?.length || 0 }}, Steps: {{ dagData?.steps?.length || 0 }}</div>
    </div>
    
    <div class="dag-view">
      <div v-if="!dagData || !dagData.jobs || dagData.jobs.length === 0" class="empty-overlay">
        <el-empty description="暂无构建过程数据" />
      </div>
      <div ref="graphContainer" class="graph-container"></div>
    </div>

    <el-drawer
      v-model="drawerVisible"
      :title="drawerTitle"
      direction="rtl"
      size="50%"
    >
      <div class="log-container" ref="logContainer">
        <div v-for="(line, index) in currentLogs" :key="index" class="log-line">{{ line }}</div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { Graph } from '@antv/x6'

const props = defineProps({
  snapshot: {
    type: String,
    default: ''
  },
  logs: {
    type: Array,
    default: () => []
  },
  dagData: {
    type: Object,
    default: () => null
  }
})

const graphContainer = ref(null)
const drawerVisible = ref(false)
const selectedJob = ref('')
const selectedStep = ref('')
let graph = null

const drawerTitle = computed(() => {
  if (selectedStep.value) return `Step: ${selectedStep.value}`
  return `Job: ${selectedJob.value}`
})

// 根据状态获取颜色
const getStatusColor = (status) => {
  if (!status) return '#d9d9d9' // pending
  const s = status.toLowerCase()
  if (s.includes('running')) return '#1890ff'  // blue
  if (s.includes('success') || s.includes('succeeded')) return '#52c41a' // green
  if (s.includes('fail')) return '#ff4d4f' // red
  if (s.includes('cancel')) return '#8c8c8c' // gray
  return '#d9d9d9'
}

// 创建 Job 节点内容（包含 Steps）
const createJobNodeMarkup = (job, steps) => {
  const statusColor = getStatusColor(job.Status)
  const stepItems = steps.map(step => {
    const stepColor = getStatusColor(step.Status)
    return `
      <div class="step-item" data-job="${job.Name}" data-step="${step.Name}" style="border-left: 3px solid ${stepColor};">
        <span class="step-name">${step.Name}</span>
        <span class="step-status" style="color: ${stepColor};">${step.Status || 'pending'}</span>
      </div>
    `
  }).join('')
  
  return `
    <div class="job-node">
      <div class="job-header" style="background-color: ${statusColor};">
        <span class="job-name">${job.Name}</span>
      </div>
      <div class="job-steps">
        ${stepItems || '<div class="no-steps">No steps</div>'}
      </div>
    </div>
  `
}

// 初始化图
const initGraph = () => {
  if (!graphContainer.value) return
  
  graph = new Graph({
    container: graphContainer.value,
    autoResize: true,
    panning: { enabled: true },
    mousewheel: { enabled: true, modifiers: ['ctrl', 'meta'] },
    connecting: { enabled: false },
  })
  
  // 监听节点点击和 step 点击
  graph.on('node:click', ({ node, e }) => {
    console.log('Node clicked:', node.id, 'event target:', e.target)
    
    const target = e.target
    let stepElement = target
    
    // 向上查找 data-step 属性
    while (stepElement && stepElement !== e.currentTarget) {
      if (stepElement.dataset?.step) {
        const jobName = stepElement.dataset.job
        const stepName = stepElement.dataset.step
        console.log('Step clicked:', jobName, stepName)
        selectedJob.value = jobName
        selectedStep.value = stepName
        drawerVisible.value = true
        return
      }
      stepElement = stepElement.parentElement
    }
    
    // 如果没有找到 step，就是点击了 job
    console.log('Job clicked:', node.id)
    selectedJob.value = node.id
    selectedStep.value = ''
    drawerVisible.value = true
  })
}

// 渲染 DAG
const renderDAG = () => {
  console.log('renderDAG called', graph, props.dagData)
  if (!graph || !props.dagData) {
    console.warn('Cannot render: graph or dagData missing')
    return
  }
  
  graph.clearCells()
  
  const jobs = props.dagData.jobs || []
  const steps = props.dagData.steps || []
  const edges = props.dagData.edges || []
  
  console.log('Rendering jobs:', jobs.length, 'steps:', steps.length, 'edges:', edges.length)
  
  if (jobs.length === 0) return
  
  // 计算层级布局
  const jobLevels = new Map() // job name -> level
  const inDegree = new Map() // job name -> 入度
  
  // 初始化入度
  jobs.forEach(job => {
    inDegree.set(job.Name, 0)
  })
  
  // 计算入度
  edges.forEach(edge => {
    const current = inDegree.get(edge.ToJob) || 0
    inDegree.set(edge.ToJob, current + 1)
  })
  
  // 层级分配（使用拓扑排序）
  let currentLevel = 0
  let remaining = new Set(jobs.map(j => j.Name))
  
  while (remaining.size > 0) {
    // 找出当前层的节点（入度为0的节点）
    const currentLevelJobs = []
    remaining.forEach(jobName => {
      if (inDegree.get(jobName) === 0) {
        currentLevelJobs.push(jobName)
      }
    })
    
    if (currentLevelJobs.length === 0) break // 避免死循环
    
    // 分配层级
    currentLevelJobs.forEach(jobName => {
      jobLevels.set(jobName, currentLevel)
      remaining.delete(jobName)
      
      // 减少后继节点的入度
      edges.forEach(edge => {
        if (edge.FromJob === jobName) {
          inDegree.set(edge.ToJob, inDegree.get(edge.ToJob) - 1)
        }
      })
    })
    
    currentLevel++
  }
  
  // 按层级分组
  const levelGroups = new Map()
  jobs.forEach(job => {
    const level = jobLevels.get(job.Name) || 0
    if (!levelGroups.has(level)) {
      levelGroups.set(level, [])
    }
    levelGroups.get(level).push(job)
  })
  
  const levelSpacingX = 520
  const nodeSpacingY = 200
  const posX = new Map()
  const posY = new Map()
  const jobMap = new Map(jobs.map(j => [j.Name, j]))
  const parentsMap = new Map()
  edges.forEach(e => {
    const from = e?.FromJob
    const to = e?.ToJob
    if (!from || !to) return
    if (!parentsMap.has(to)) parentsMap.set(to, [])
    parentsMap.get(to).push(from)
  })
  const stepsByJob = new Map()
  jobs.forEach(j => { stepsByJob.set(j.Name, steps.filter(s => s.JobName === j.Name)) })
  const heights = new Map()
  jobs.forEach(j => { const arr = stepsByJob.get(j.Name) || []; heights.set(j.Name, Math.max(120, 70 + arr.length * 40)) })
  const positionsY = new Map()
  const centersY = new Map()
  const sortedLevels = Array.from(levelGroups.keys()).sort((a, b) => a - b)
  const rowGap = nodeSpacingY
  sortedLevels.forEach(lv => {
    const names = (levelGroups.get(lv) || []).map(j => j.Name)
    if (lv === 0) {
      const ordered = names.slice().sort((a, b) => {
        const ma = String(a || '').match(/^(.*?)-(\d+)$/)
        const mb = String(b || '').match(/^(.*?)-(\d+)$/)
        if (ma && mb && ma[1] === mb[1]) return Number(ma[2]) - Number(mb[2])
        return String(a || '').localeCompare(String(b || ''))
      })
      let sweep = 80
      ordered.forEach(n => {
        const h = heights.get(n) || 120
        const y = sweep
        positionsY.set(n, y)
        centersY.set(n, y + Math.floor(h / 2))
        sweep = y + h + rowGap
      })
      return
    }
    const targets = names.map(n => {
      const ps = parentsMap.get(n) || []
      const vals = ps.map(p => centersY.get(p)).filter(v => typeof v === 'number')
      const avg = vals.length ? Math.floor(vals.reduce((a, b) => a + b, 0) / vals.length) : NaN
      return { n, t: avg }
    }).sort((a, b) => {
      const ta = Number.isNaN(a.t) ? Number.POSITIVE_INFINITY : a.t
      const tb = Number.isNaN(b.t) ? Number.POSITIVE_INFINITY : b.t
      return ta - tb
    })
    let sweep = 80
    targets.forEach(({ n, t }) => {
      const h = heights.get(n) || 120
      const desiredTop = Number.isNaN(t) ? sweep : Math.floor(t - Math.floor(h / 2))
      const y = Math.max(desiredTop, sweep)
      positionsY.set(n, y)
      centersY.set(n, y + Math.floor(h / 2))
      sweep = y + h + rowGap
    })
  })
  
  sortedLevels.forEach(lv => {
    const names = (levelGroups.get(lv) || []).map(j => j.Name).slice().sort((a, b) => (positionsY.get(a) || 0) - (positionsY.get(b) || 0))
    names.forEach(name => {
      const job = jobMap.get(name)
      const jobSteps = stepsByJob.get(name) || []
      const statusColor = getStatusColor(job.Status)
      const x = lv * levelSpacingX + 80
      const y = positionsY.get(name) || 80
      posX.set(name, x)
      posY.set(name, y)
      const stepHtml = jobSteps.map((step) => {
        const stepColor = getStatusColor(step.Status)
        return `
          <div style="padding: 4px 8px; margin: 2px 0; background: #f5f7fa; border-left: 3px solid ${stepColor}; border-radius: 3px; cursor: pointer; font-size: 12px;"
               data-job="${job.Name}" data-step="${step.Name}">
            <div style="display: flex; justify-content: space-between;">
              <span>${step.Name}</span>
              <span style="color: ${stepColor}; font-size: 10px;">${step.Status || 'pending'}</span>
            </div>
          </div>
        `
      }).join('')
      const nodeHtml = `
        <div xmlns="http://www.w3.org/1999/xhtml" style="width: 100%; height: 100%; background: white; border-radius: 4px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden;">
          <div style="padding: 10px; background: ${statusColor}; color: white; font-weight: bold; font-size: 14px;">
            ${job.Name}
          </div>
          <div style="padding: 8px; max-height: 200px; overflow-y: auto;">
            ${stepHtml || '<div style="color: #999; text-align: center; padding: 10px; font-size: 12px;">No steps</div>'}
          </div>
        </div>
      `
      const nodeHeight = heights.get(name) || Math.max(120, 70 + jobSteps.length * 40)
      graph.addNode({
        id: job.Name,
        x,
        y,
        width: 280,
        height: nodeHeight,
        shape: 'rect',
        markup: [ { tagName: 'rect', selector: 'body' }, { tagName: 'foreignObject', selector: 'fo' } ],
        attrs: { body: { fill: 'transparent', stroke: 'transparent' }, fo: { width: 280, height: nodeHeight, x: 0, y: 0, html: nodeHtml } },
        ports: { groups: { left: { position: 'left', attrs: { circle: { r: 4, magnet: true, stroke: '#8c8c8c', strokeWidth: 1, fill: '#fff' } } }, right: { position: 'right', attrs: { circle: { r: 4, magnet: true, stroke: '#8c8c8c', strokeWidth: 1, fill: '#fff' } } } }, items: [ { id: 'left', group: 'left' }, { id: 'right', group: 'right' } ] },
      })
    })
  })
  
  edges.forEach(edge => {
    if (!edge.FromJob || !edge.ToJob) return
    const lf = jobLevels.get(edge.FromJob) || 0
    const lt = jobLevels.get(edge.ToJob) || 0
    const xf = posX.get(edge.FromJob) || 0
    const xt = posX.get(edge.ToJob) || 0
    const forward = lt > lf || xt > xf
    const sourcePort = forward ? 'right' : 'left'
    const targetPort = forward ? 'left'  : 'right'
    graph.addEdge({
      source: { cell: edge.FromJob, port: sourcePort },
      target: { cell: edge.ToJob, port: targetPort },
      attrs: { line: { stroke: '#8c8c8c', strokeWidth: 2, targetMarker: { name: 'classic', size: 8 } } },
      router: { name: 'manhattan', args: { padding: 14 } },
      connector: { name: 'rounded' },
    })
  })
  
  // 自动适配视图
  setTimeout(() => {
    graph.zoomToFit({ padding: 80, maxScale: 1 })
  }, 100)
  
  console.log('DAG rendered successfully')
}

// 获取当前日志
const currentLogs = computed(() => {
  console.log('currentLogs - selectedJob:', selectedJob.value, 'selectedStep:', selectedStep.value)
  console.log('currentLogs - total logs:', props.logs.length)
  
  if (!selectedJob.value) {
    console.log('No job selected')
    return ['No job selected']
  }
  
  // 过滤日志：查找包含 job 名称的行
  const jobLogs = props.logs.filter(line => {
    // 匹配格式如: 📦 Job [job-1] Started 或 🔹 Step [M1bash] Running
    return line.includes(`[${selectedJob.value}]`) || line.includes(selectedJob.value)
  })
  
  console.log('currentLogs - jobLogs count:', jobLogs.length)
  
  if (selectedStep.value) {
    const stepLogs = jobLogs.filter(line => line.includes(`[${selectedStep.value}]`) || line.includes(selectedStep.value))
    console.log('currentLogs - stepLogs count:', stepLogs.length)
    return stepLogs.length > 0 ? stepLogs : [`No logs found for step: ${selectedStep.value}`]
  }
  
  return jobLogs.length > 0 ? jobLogs : [`No logs found for job: ${selectedJob.value}`]
})

// 监听 dagData 变化
watch(() => props.dagData, () => {
  console.log('dagData changed, graph:', graph, 'container:', graphContainer.value)
  if (!graph && graphContainer.value) {
    console.log('Initializing graph because dagData changed')
    initGraph()
  }
  nextTick(() => {
    renderDAG()
  })
}, { deep: true })

onMounted(() => {
  console.log('BuildProcess mounted, container:', graphContainer.value)
  nextTick(() => {
    initGraph()
    if (props.dagData) {
      renderDAG()
    }
  })
})
</script>

<style scoped>
.build-process { padding: 0; height: 100%; display: flex; flex-direction: column; width: 100%; }
.dag-view { flex: 1 1 auto; display: flex; flex-direction: column; position: relative; min-width: 0; }
.empty-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  z-index: 10;
}
.graph-container { flex: 1 1 auto; min-height: 0; border: 1px solid #e4e7ed; border-radius: 4px; background: #fafafa; width: 100%; }

/* Job Node Styles */
:deep(.job-node) {
  background: white;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  overflow: hidden;
  height: 100%;
}
:deep(.job-header) {
  padding: 10px 15px;
  color: white;
  font-weight: bold;
  font-size: 14px;
}
:deep(.job-name) {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
:deep(.job-steps) {
  padding: 8px;
  max-height: calc(100% - 40px);
  overflow-y: auto;
}
:deep(.step-item) {
  padding: 6px 10px;
  margin-bottom: 4px;
  background: #f5f7fa;
  border-radius: 3px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: background 0.2s;
}
:deep(.step-item:hover) {
  background: #e6f7ff;
}
:deep(.step-name) {
  font-size: 12px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
:deep(.step-status) {
  font-size: 11px;
  margin-left: 8px;
  font-weight: 500;
}
:deep(.no-steps) {
  color: #999;
  font-size: 12px;
  text-align: center;
  padding: 10px;
}

.log-container {
  height: 100%;
  overflow-y: auto;
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 10px;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  line-height: 1.5;
}
.log-line {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
