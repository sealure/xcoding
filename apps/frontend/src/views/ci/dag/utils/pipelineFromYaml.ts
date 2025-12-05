// 将 GitHub Actions 风格的 YAML（含 jobs/steps/needs）转换为 X6 图数据
// 生成节点使用 makePipelineNode（垂直 in/out 端口），边使用 makeEdge（黑色）
import { load as yamlLoad } from 'js-yaml'
import { makePipelineNode } from '@/views/ci/dag/utils/pipelineStyles'
import { makeEdge } from '@/views/ci/dag/utils/edgeStyle'

// 简单坐标布局策略：
// - 每个 job 占一列，列间水平间距 columnSpacing
// - 每个 job 内的 step 纵向排列，行间距 rowSpacing
// - Start/End 放置在全局顶部与底部
// - 处理 needs：从被依赖 job 的最后一个 step 指向依赖 job 的第一个 step

// 计算 Job 的拓扑层级（level），level 越小越靠左
const computeLevels = (jobIds: string[], needsMap: Record<string, string[]>) => {
  const level = {}
  jobIds.forEach((jid) => { level[jid] = undefined })
  let changed = true
  // 初始化无依赖的为 0
  jobIds.forEach((jid) => { if ((needsMap[jid] || []).length === 0) level[jid] = 0 })
  while (changed) {
    changed = false
    jobIds.forEach((jid) => {
      const needs = needsMap[jid] || []
      if (needs.length === 0) return
      const knownNeeds = needs.filter((n) => level[n] !== undefined)
      if (knownNeeds.length === needs.length) {
        const lv = Math.max(...knownNeeds.map((n) => level[n])) + 1
        if (level[jid] !== lv) { level[jid] = lv; changed = true }
      }
    })
  }
  // 未能确定的（循环或异常）归为 0
  jobIds.forEach((jid) => { if (level[jid] === undefined) level[jid] = 0 })
  return level
}

const matrixLabel = (job: any) => {
  const matrix = job?.strategy?.matrix
  if (!matrix || typeof matrix !== 'object') return ''
  const dims = Object.values(matrix).filter((v) => Array.isArray(v)).map((arr) => arr.length)
  if (!dims.length) return ''
  const product = dims.reduce((a, b) => a * b, 1)
  return `${dims.join('×')}=${product}`
}

const headerTypeFromJob = (job: any) => {
  const cond = job?.if
  if (!cond) return 'header'
  const val = String(cond).toLowerCase()
  if (val.includes('always()')) return 'headerAlways'
  return 'headerConditional'
}

const stepTypeFromStep = (step: any) => {
  if (!step?.if) return 'stage'
  const val = String(step.if).toLowerCase()
  if (val.includes('always()')) return 'always'
  return 'conditional'
}

export type PipelineGraphData = { nodes: any[]; edges: any[] }
export const pipelineFromYaml = (yamlText: string): PipelineGraphData => {
  const doc = yamlLoad(yamlText) || {}
  const jobs = doc.jobs || {}
  const jobIds = Object.keys(jobs)

  const columnSpacing = 260
  const rowSpacing = 100
  const baseX = 180
  const baseY = 200

  const nodes: any[] = []
  const edges: any[] = []

  // 依赖关系收集
  const needsMap: Record<string, string[]> = {}
  const dependentsMap: Record<string, string[]> = {}
  jobIds.forEach((jid) => {
    let needs = jobs[jid]?.needs ?? []
    if (typeof needs === 'string') needs = [needs]
    if (!Array.isArray(needs)) needs = []
    needsMap[jid] = needs
    needs.forEach((n) => {
      if (!dependentsMap[n]) dependentsMap[n] = []
      dependentsMap[n].push(jid)
    })
  })

  // 计算拓扑层级，并生成按层级排序的 Job 列序
  const levels = computeLevels(jobIds, needsMap)
  const orderedJobIds = [...jobIds].sort((a, b) => {
    const la = levels[a], lb = levels[b]
    if (la !== lb) return la - lb
    return jobIds.indexOf(a) - jobIds.indexOf(b)
  })

  // 移除全局 Start/End 节点：仅展示各 Job 的 Header/Steps/Services

  // 为每个 job 生成 Header、Service、Step 节点与顺序边
  const firstStepOfJob: Record<string, string> = {}
  const lastStepOfJob: Record<string, string> = {}
  const headerOfJob: Record<string, string> = {}

  let edgeCounter = 1
  orderedJobIds.forEach((jid, orderIndex) => {
    const job = jobs[jid] || {}
    const steps = Array.isArray(job.steps) ? job.steps : []
    const jobX = baseX + orderIndex * columnSpacing

    // Header（包含矩阵提示与条件样式）
    const headerId = `${jid}-header`
    const hLabelMatrix = matrixLabel(job)
    const headerLabel = `${job.name || jid}${hLabelMatrix ? ` [${hLabelMatrix}]` : ''}`
    nodes.push(
      makePipelineNode({ id: headerId, x: jobX, y: baseY - 60, label: headerLabel, type: headerTypeFromJob(job) })
    )
    headerOfJob[jid] = headerId

    // Services（显示为虚线边与服务样式）
    const services = job.services || {}
    const serviceNames = Object.keys(services)
    serviceNames.forEach((sName, sIdx) => {
      const sId = `${jid}-svc-${sName}`
      nodes.push(makePipelineNode({ id: sId, x: jobX - 160, y: (baseY - 60) + sIdx * 40, label: sName, type: 'service' }))
      edges.push(makeEdge({ id: `e${edgeCounter++}`, source: { cell: sId }, target: { cell: headerId }, type: 'dashed' }))
    })

    // 如果没有 steps，使用一个占位 step 表示该 job
    const ensuredSteps = steps.length > 0 ? steps : [{ name: job.name || jid, run: 'noop' }]

    let prevStepNodeId = null
    ensuredSteps.forEach((step, sIndex) => {
      const label = step.name || (step.uses ? `uses ${step.uses}` : step.run ? `run: ${String(step.run).slice(0, 32)}...` : `Step ${sIndex + 1}`)
      const nodeId = `${jid}-s${sIndex + 1}`
      const y = baseY + sIndex * rowSpacing

      nodes.push(
        makePipelineNode({ id: nodeId, x: jobX, y, label, type: stepTypeFromStep(step) })
      )

      if (prevStepNodeId) {
        edges.push(
          makeEdge({
            id: `e${edgeCounter++}`,
            source: { cell: prevStepNodeId, port: `${prevStepNodeId}-out` },
            target: { cell: nodeId, port: `${nodeId}-in` },
            type: 'black',
          })
        )
      }

      prevStepNodeId = nodeId
      if (sIndex === 0) firstStepOfJob[jid] = nodeId
      if (sIndex === ensuredSteps.length - 1) lastStepOfJob[jid] = nodeId
    })

    // Job Header -> 第一个 Step（仅在本 Job 内部）
    if (firstStepOfJob[jid]) {
      edges.push(
        makeEdge({
          id: `e${edgeCounter++}`,
          source: { cell: headerId },
          target: { cell: firstStepOfJob[jid], port: `${firstStepOfJob[jid]}-in` },
          type: 'black',
        })
      )
    }
  })

  // 移除从 Start 到无依赖 Job 的 Header 的连接

  // 处理 needs 依赖：Job Header -> Job Header（仅跨 Job）
  jobIds.forEach((jid) => {
    const needs = needsMap[jid] || []
    needs.forEach((nid) => {
      const src = headerOfJob[nid]
      const tgt = headerOfJob[jid]
      if (src && tgt) {
        edges.push(
          makeEdge({
            id: `e${edgeCounter++}`,
            source: { cell: src },
            target: { cell: tgt },
            type: 'black',
          })
        )
      }
    })
  })

  // 移除从各 Job 的 Header 到 End 的连接

  return { nodes, edges }
}

export default pipelineFromYaml