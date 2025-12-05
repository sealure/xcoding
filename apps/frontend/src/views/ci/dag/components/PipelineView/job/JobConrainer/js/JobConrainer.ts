import { makePipelineNode } from '@/views/ci/dag/utils/pipelineStyles'
import { makeEdge } from '@/views/ci/dag/utils/edgeStyle'
import { PIPELINE_STYLE_MAP } from '@/views/ci/dag/utils/nodeStyles'
import { matrixHint, normalizeStep, jobIfType } from '../../../utils'
import { Graph } from '@antv/x6'

// 辅助：根据 ID 生成稳定且区分度足够的颜色（HSL→Hex）
const GOLDEN_ANGLE = 137.508
const hashStr = (s: string) => { let h = 0; for (let i = 0; i < s.length; i++) { h = (h * 31 + s.charCodeAt(i)) >>> 0 } return h }
const hslToHex = (h: number, s: number, l: number) => {
  const S = s / 100, L = l / 100
  const c = (1 - Math.abs(2 * L - 1)) * S
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1))
  const m = L - c / 2
  let r = 0, g = 0, b = 0
  if (h < 60) { r = c; g = x; b = 0 }
  else if (h < 120) { r = x; g = c; b = 0 }
  else if (h < 180) { r = 0; g = c; b = x }
  else if (h < 240) { r = 0; g = x; b = c }
  else if (h < 300) { r = x; g = 0; b = c }
  else { r = c; g = 0; b = x }
  const toHex = (v: number) => Math.round((v + m) * 255).toString(16).padStart(2, '0')
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}
const colorFromId = (id: string) => {
  const h = Math.floor(((hashStr(id) + 1) * GOLDEN_ANGLE) % 360)
  const s = 72
  const l = 48
  return hslToHex(h, s, l)
}

type RenderJobContainerInput = { graph: Graph; jid: string; job: any; jobX: number; jobY: number; collapsed: boolean; tagText?: string }
export const renderJobContainer = ({ graph, jid, job, jobX, jobY, collapsed, tagText }: RenderJobContainerInput) => {
  const stepsArr = Array.isArray(job?.steps) ? job.steps : []
  const headerText = `${tagText ? `${tagText}  ` : ''}${job?.name || jid}${matrixHint(job) ? ` [${matrixHint(job)}]` : ''}`  // job name拼接 1-1 job name

  const headerH = 36
  const svcH = 0
  const STEP_HEIGHT = 34
  const STEP_GAP = 6
  const stepH = collapsed ? 0 : Math.max(stepsArr.length, 1) * (STEP_HEIGHT + STEP_GAP)
  const jobHeight = headerH + svcH + stepH + 24

  const jobNodeId = `${jid}-container`
  const jobNode = makePipelineNode({ id: jobNodeId, x: jobX, y: jobY, label: '', type: 'job', width: 300, height: jobHeight })
  const jobCell = graph.addNode(jobNode as any)

  const paddingX = 10
  let yCursor = jobY + headerH + 12
  const headerNodeId = `${jid}-header`
  const jobIf = jobIfType(job)
  const headerStyleType = jobIf === 'always' ? 'headerAlways' : (jobIf === 'conditional' ? 'headerConditional' : 'header')
  const headerInset = 0
  const headerWidth = 300
  const headerHtml = `
    <div xmlns="http://www.w3.org/1999/xhtml" style="width: 100%; height: 100%;">
      <div style="padding: 10px; background: #d9d9d9; color: white; font-weight: bold; font-size: 14px; border-radius: 4px 4px 0 0;">
        ${headerText}
      </div>
    </div>
  `
  const headerNode = {
    id: headerNodeId,
    x: jobX,
    y: jobY + headerInset,
    width: headerWidth,
    height: headerH,
    shape: 'rect',
    markup: [ { tagName: 'rect', selector: 'body' }, { tagName: 'foreignObject', selector: 'fo' } ],
    attrs: { body: { fill: 'transparent', stroke: 'transparent' }, fo: { width: headerWidth, height: headerH, x: 0, y: 0, html: headerHtml } },
    data: { type: headerStyleType, jobId: jid, anchorTo: `${jid}-container`, anchorOffsetX: 0, anchorOffsetY: headerInset },
  }
  const headerCell = graph.addNode(headerNode as any)
  jobCell.addChild(headerCell)
  // 移除 Job Header 右侧 DEL 标签
  // Job 编号标签已并入标题文本，不再单独渲染左侧小标签
  

  let edgesCount = 0
  let stepsCount = 0
  let prevStepId: string | null = null
  let firstStepId: string | null = null

  const ensuredSteps = stepsArr.length ? stepsArr : [{ name: job?.name || jid, run: 'noop' }]

  if (!collapsed) {
    ensuredSteps.forEach((step: any, sIdx: number) => {
      const norm = normalizeStep(step)
      const label = norm.label
      // 统一将步骤渲染为灰色“step”类型
      const stepType = 'step'
      const nodeId = `${jid}-s${sIdx + 1}`
      const stepHtml = `
        <div xmlns="http://www.w3.org/1999/xhtml" style="width:100%;height:100%;">
          <div style="padding: 6px 10px; background: #f5f7fa; border-left: 3px solid #d9d9d9; border-radius: 3px; font-size: 12px; line-height: 20px;">
            <div style="display:flex; justify-content:space-between; align-items:center;">
              <span style="color:#606266; font-weight:500; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width: 228px;">${String(label)}</span>
            </div>
          </div>
        </div>
      `
      const stepNode = {
        id: nodeId,
        x: jobX + paddingX,
        y: yCursor,
        width: 280,
        height: STEP_HEIGHT,
        shape: 'rect',
        markup: [ { tagName: 'rect', selector: 'body' }, { tagName: 'foreignObject', selector: 'fo' } ],
        attrs: { body: { fill: 'transparent', stroke: 'transparent' }, fo: { width: 280, height: STEP_HEIGHT, x: 0, y: 0, html: stepHtml } },
        data: { type: 'step', jobId: jid, stepIndex: sIdx, payload: step },
      }
      const stepCell = graph.addNode(stepNode as any)
      jobCell.addChild(stepCell)
      
      // 移除 Step 右侧 DEL 标签
      stepsCount++
      if (prevStepId) {
      }
      prevStepId = nodeId
      if (!firstStepId) firstStepId = nodeId
      yCursor += (STEP_HEIGHT + STEP_GAP)
    })
  }

  // 不再连接 header → firstStep，保持标题纯展示

  // 同步容器底部位置以适配增大的步骤间距
  const jobBottom = jobY + jobHeight

  
  return {
    jobCell,
    headerCell,
    firstStepId,
    lastStepId: prevStepId,
    jobHeight,
    jobBottom,
    stats: { headers: 1, steps: stepsCount, edges: edgesCount },
  }
}

// 连接 Job 容器之间的依赖边
// 为依赖边分配不相交“车道”，通过 vertices 指定路径，避免不同节点的边相交
export const linkJobDependencies = ({ graph, jobs, jobContainerById, levels, layout }: { graph: Graph; jobs: Record<string, any>; jobContainerById: Record<string, any>; levels: Record<string, number>; layout?: { columnSpacing?: number } }) => {
  let edges = 0
  try {
    // 基于“首个父节点”计算分支根，并按根分配颜色，保证整条分支颜色一致
    const ids = Object.keys(jobs || {})
    const needsOf = (jid: string) => {
      const raw = jobs?.[jid]?.needs
      const list = raw ? (Array.isArray(raw) ? raw : [raw]) : []
      return list.filter(Boolean) as string[]
    }
    const branchRootCache: Record<string, string> = {}
    const resolveBranchRoot = (jid: string): string => {
      if (branchRootCache[jid]) return branchRootCache[jid]
      const deps = needsOf(jid)
      if (deps.length === 0) {
        branchRootCache[jid] = jid
        return jid
      }
      const firstParent = deps[0]
      const root = resolveBranchRoot(firstParent)
      branchRootCache[jid] = root
      return root
    }
    ids.forEach((id) => resolveBranchRoot(id))
    const branchColorCache: Record<string, string> = {}
    const colorForBranch = (jid: string) => {
      const root = resolveBranchRoot(jid)
      if (!branchColorCache[root]) branchColorCache[root] = colorFromId(root)
      return branchColorCache[root]
    }
    
    // 按目标列对边分组（源列可不同），每组在两列中间分配若干“车道”
    const groups: Record<number, Array<{ dep: string; jid: string }>> = {}
    ids.forEach((jid) => {
      const needs = jobs[jid]?.needs
      if (!needs) return
      const list = Array.isArray(needs) ? needs : [needs]
      list.forEach((dep) => {
        const lv = Number(levels?.[jid] ?? 0)
        if (!groups[lv]) groups[lv] = []
        groups[lv].push({ dep, jid })
      })
    })

    const { columnSpacing = 560 } = layout || {}
    // 依据布局动态计算安全边距/车道间距/路由 padding，避免硬编码
    const safePad = Math.max(8, Math.floor(columnSpacing * 0.03))
    const laneInterval = Math.max(20, Math.floor(columnSpacing * 0.06))
    const dynamicPadding = Math.max(8, Math.floor(columnSpacing * 0.02))

    Object.keys(groups).map((k) => Number(k)).sort((a, b) => a - b).forEach((targetLv) => {
      const edgesInBand = groups[targetLv]
      if (!edgesInBand || !edgesInBand.length) return

      // 计算该目标列与其左侧列的大致 X 范围：从源 job 容器右侧到目标容器左侧
      // 通过任意一个源/目标节点的 bbox 推断两列的 X，中点附近分布多个车道
      let leftX = Infinity
      let rightX = -Infinity
      const sourceYs: number[] = []
      const targetYs: number[] = []
      edgesInBand.forEach(({ dep, jid }) => {
        const from = jobContainerById[dep]
        const to = jobContainerById[jid]
        if (!from || !to) return
        try {
          const fb = from.getBBox()
          const tb = to.getBBox()
          leftX = Math.min(leftX, fb.x + fb.width)
          rightX = Math.max(rightX, tb.x)
          sourceYs.push(fb.y + fb.height / 2)
          targetYs.push(tb.y + tb.height / 2)
        } catch (_) {}
      })
      if (!Number.isFinite(leftX) || !Number.isFinite(rightX)) return
      const bandLeft = leftX + safePad
      const bandRight = rightX - safePad
      const bandWidth = Math.max(60, bandRight - bandLeft)
      // 改为按 Y 方向分配“车道”，让跨列的长水平段走在不同的 Y 车道上，减少与其它列垂直段的相交
      const bandTop = Math.min(...sourceYs, ...targetYs)
      const bandBottom = Math.max(...sourceYs, ...targetYs)
      const laneCount = Math.max(1, edgesInBand.length)
      const laneYs: number[] = []
      for (let i = 0; i < laneCount; i++) {
        laneYs.push(Math.floor(bandTop + ((bandBottom - bandTop) * (i + 1)) / (laneCount + 1)))
      }

      // 按源中心 Y 排序，为每条边分配一个 lane，保证从左到右的水平段互不相交
      const ordered = [...edgesInBand].sort((a, b) => {
        try {
          const fa = jobContainerById[a.dep]?.getBBox?.()
          const fb = jobContainerById[b.dep]?.getBBox?.()
          const ya = fa ? (fa.y + fa.height / 2) : 0
          const yb = fb ? (fb.y + fb.height / 2) : 0
          return ya - yb
        } catch (_) { return 0 }
      })

      ordered.forEach((item, idx) => {
        const from = jobContainerById[item.dep]
        const to = jobContainerById[item.jid]
        if (!from || !to) return
        try {
          const fb = from.getBBox()
          const tb = to.getBBox()
          const sx = fb.x + fb.width
          const sy = fb.y + fb.height / 2
          const tx = tb.x
          const ty = tb.y + tb.height / 2
          const laneY = laneYs[idx % laneYs.length]
          const edgeId = `dep-edge-${item.dep}__${item.jid}`
          const sourceLv = Number(levels?.[item.dep] ?? targetLv - 1)
          const diffY = Math.abs(sy - ty)
          const isAdjacent = Math.abs(targetLv - sourceLv) <= 1
          // 更激进的直连策略：
          // - 邻近列几乎全部直连
          // - 垂直距离较小或同一带内边不多也直连
          // - 当源在左、目标在右且中间列间距充足（safePad/laneInterval 已保证），优先直连
          const simpleCase = (
            isAdjacent ||
            diffY <= laneInterval ||
            edgesInBand.length <= 3
          )
          const e = graph.addEdge(makeEdge({
            id: edgeId,
            source: { cell: from.id, port: `${from.id}-out` },
            target: { cell: to.id, port: `${to.id}-in` },
            type: 'black',
            routerName: 'manhattan',
            routerArgs: { padding: dynamicPadding },
            connectorName: 'rounded',
            // 复杂场景走“纵向车道”，简单场景不设置 vertices 以获得直连
            ...(simpleCase
              ? {}
              : {
                vertices: [
                  { x: Math.floor(sx + safePad), y: Math.floor(sy) },
                  { x: Math.floor(sx + safePad), y: Math.floor(laneY) },
                  { x: Math.floor(tx - safePad), y: Math.floor(laneY) },
                  { x: Math.floor(tx - safePad), y: Math.floor(ty) },
                ],
              }),
          }))
          edges += 1
        } catch (_) {}
      })
    })
  } catch (_) {}
  return edges
}

// 计算 Job 的层级，便于进行列布局
export const computeLevels = (ids: string[], needsMap: Record<string, string[]>) => {
  const idSet = new Set(ids)
  const level: Record<string, number> = {}
  const temp: Record<string, boolean> = {}
  const dfs = (id: string): number => {
    if (level[id] !== undefined) return level[id]
    if (temp[id]) return 0
    temp[id] = true
    const needs = needsMap[id] || []
    let maxDep = -1
    for (const n of needs) {
      if (!idSet.has(n)) continue
      const lv = dfs(n)
      if (lv > maxDep) maxDep = lv
    }
    temp[id] = false
    level[id] = maxDep + 1
    return level[id]
  }
  ids.forEach((id: string) => { dfs(id) })
  return level
}

// 整体渲染：根据文档绘制所有 Job 容器并连接依赖
export const renderJobsPipeline = ({ graph, doc, collapsed, collapsedJobs, layout }: { graph: Graph; doc: any; collapsed: boolean; collapsedJobs: Record<string, boolean>; layout?: { columnSpacing?: number; baseX?: number; baseY?: number; columnRowGap?: number; startOffset?: number; endOffset?: number; startEndHeight?: number; autoOrder?: boolean } }) => {
  const jobsObj = (doc && doc.jobs) ? doc.jobs : {}
  const ids = Object.keys(jobsObj)
  if (!graph) return { stats: { jobs: 0, headers: 0, services: 0, steps: 0, edges: 0 } }
  try { graph.clearCells() } catch (_) {}
  const stats = { jobs: 0, headers: 0, services: 0, steps: 0, edges: 0 }

  const { columnSpacing = 560, baseX = 600, baseY = 160, columnRowGap = 80, startOffset = 800, endOffset = 420, startEndHeight = 64, autoOrder = true } = layout || {}

  const jobContainerById: Record<string, any> = {}
  let maxBottomY = 0

  const needsMap: Record<string, string[]> = {}
  ids.forEach((jid) => {
    const raw = jobsObj[jid]?.needs
    const list = raw ? (Array.isArray(raw) ? raw : [raw]) : []
    needsMap[jid] = list.filter(Boolean)
  })
  // 分支根与颜色：按“第一个父节点”递归解析分支根，为整条分支分配同一颜色
  const branchRootCache: Record<string, string> = {}
  const resolveBranchRoot = (jid: string): string => {
    if (branchRootCache[jid]) return branchRootCache[jid]
    const deps = needsMap[jid] || []
    if (deps.length === 0) { branchRootCache[jid] = jid; return jid }
    const root = resolveBranchRoot(deps[0])
    branchRootCache[jid] = root
    return root
  }
  const branchColorCache: Record<string, string> = {}
  const colorForBranch = (jid: string) => {
    const root = resolveBranchRoot(jid)
    if (!branchColorCache[root]) branchColorCache[root] = colorFromId(root)
    return branchColorCache[root]
  }
  const levels = computeLevels(ids, needsMap)
  const yamlIndex: Record<string, number> = {}
  ids.forEach((id, i) => (yamlIndex[id] = i))
  const orderedIds = [...ids].sort((a, b) => {
    const la = levels[a] ?? 0
    const lb = levels[b] ?? 0
    if (la !== lb) return la - lb
    return ids.indexOf(a) - ids.indexOf(b)
  })

  const levelToJobs: Record<number, string[]> = {}
  orderedIds.forEach((jid) => {
    const lv = levels[jid] ?? 0
    if (!levelToJobs[lv]) levelToJobs[lv] = []
    levelToJobs[lv].push(jid)
  })
  const levelKeys = Object.keys(levelToJobs).map((n) => Number(n)).sort((a, b) => a - b)

  const columnBounds: Array<{ top: number; bottom: number }> = []

  // 估算 Job 高度，用于计算目标 Y 值（与实际渲染高度一致的近似）
  const estimateJobHeight = (jid: string) => {
    try {
      const j = jobsObj[jid] || {}
      const stepsArr = Array.isArray(j?.steps) ? j.steps : []
      const isJobCollapsed = !!(collapsed || (collapsedJobs && collapsedJobs[jid]))
      const headerH = 36
      const svcH = 0
      const STEP_HEIGHT = 34
      const STEP_GAP = 6
      const stepH = isJobCollapsed ? 0 : Math.max(stepsArr.length, 1) * (STEP_HEIGHT + STEP_GAP)
      return headerH + svcH + stepH + 24
    } catch (_) { return 60 }
  }

  // 各 Job 的近似中心 Y 值（用于后续列的排序）
  const approxCenterY: Record<string, number> = {}

  levelKeys.forEach((lv, colIndex) => {
    // 显式标注类型，避免 union 到 never[]
    let jobsAtLevel: string[] = levelToJobs[lv] ? [...levelToJobs[lv]] : []
    const jobX = baseX + colIndex * columnSpacing
    let columnYCursor = baseY
    let columnBottom = baseY
    let withinIndex = 0
    let lastJobIdInColumn = ''

    // 树枝展开：按前一列父节点分组，组内子节点依次在父节点中心Y下方展开
    if (autoOrder) {
      const branchGap = 24
      type SingleGroup = { kind: 'single'; parentId: string; center: number; jobs: string[] }
      type MultiGroup = { kind: 'multi'; parentKey: string; centers: number[]; top: number; bottom: number; jobs: string[] }
      type Ungrouped = { kind: 'ungrouped'; jobs: string[] }
      const singleMap: Record<string, SingleGroup> = {}
      const multiMap: Record<string, MultiGroup> = {}
      const ungrouped: string[] = []
      // 分组：单父 → 单组；多父 → 其父中心区间的分支组；无父 → 未分组
      jobsAtLevel.forEach((jid) => {
        const deps = needsMap[jid] || []
        const centers = deps.map((d) => approxCenterY[d]).filter((v) => typeof v === 'number') as number[]
        if (centers.length <= 0) {
          ungrouped.push(jid)
          return
        }
        if (centers.length === 1) {
          const parent = deps[0]
          const c = centers[0]
          if (!singleMap[parent]) singleMap[parent] = { kind: 'single', parentId: parent, center: c, jobs: [] }
          singleMap[parent].jobs.push(jid)
          return
        }
        const parentKey = [...deps].sort().join('__')
        const sortedCenters = [...centers].sort((a, b) => a - b)
        const top = sortedCenters[0]
        const bottom = sortedCenters[sortedCenters.length - 1]
        if (!multiMap[parentKey]) multiMap[parentKey] = { kind: 'multi', parentKey, centers: sortedCenters, top, bottom, jobs: [] }
        multiMap[parentKey].jobs.push(jid)
        // 更新该组的区间
        multiMap[parentKey].top = Math.min(multiMap[parentKey].top, top)
        multiMap[parentKey].bottom = Math.max(multiMap[parentKey].bottom, bottom)
      })

      const extraUngrouped: Ungrouped[] = ungrouped
        .sort((a, b) => (yamlIndex[a] ?? 0) - (yamlIndex[b] ?? 0))
        .map((jid): Ungrouped => ({ kind: 'ungrouped', jobs: [jid] }))
      const groups: Array<SingleGroup | MultiGroup | Ungrouped> = [
        ...Object.values(singleMap).sort((a, b) => a.center - b.center),
        ...Object.values(multiMap).sort((a, b) => a.top - b.top || a.bottom - b.bottom),
        ...extraUngrouped,
      ]

      groups.forEach((g) => {
        if (g.kind === 'single') {
          // 单父分支：围绕父中心Y做对称展开，子节点按均匀间距上下分布，显著降低边相交
          const parentCenter = g.center
          const sortedJobs = [...g.jobs].sort((a, b) => {
            const aDeps = (needsMap[a] || []).map((d) => approxCenterY[d]).filter((v) => typeof v === 'number') as number[]
            const bDeps = (needsMap[b] || []).map((d) => approxCenterY[d]).filter((v) => typeof v === 'number') as number[]
            const ay = aDeps.length ? Math.max(...aDeps) : NaN
            const by = bDeps.length ? Math.max(...bDeps) : NaN
            const aHas = !Number.isNaN(ay)
            const bHas = !Number.isNaN(by)
            if (aHas && bHas && ay !== by) return ay - by
            return (yamlIndex[a] ?? 0) - (yamlIndex[b] ?? 0)
          })
          const childEstHeights = sortedJobs.map((jid) => estimateJobHeight(jid))
          const avgH = Math.max(60, Math.floor(childEstHeights.reduce((acc, h) => acc + h, 0) / Math.max(1, childEstHeights.length)))
          const spacing = Math.max(24, Math.floor(avgH + columnRowGap / 2))
          sortedJobs.forEach((jid, idx) => {
            const j = jobsObj[jid] || {}
            const isJobCollapsed = !!(collapsed || (collapsedJobs && collapsedJobs[jid]))
            const estH = estimateJobHeight(jid)
            // 对称序号：-k..0..+k，使子 Job 围绕父中心均匀上下分布
            const balanced = idx - Math.floor((sortedJobs.length - 1) / 2)
            const targetCenter = parentCenter + balanced * spacing
            const jobY = Math.max(baseY, Math.floor(targetCenter - estH / 2))
            const result = renderJobContainer({ graph, jid, job: j, jobX, jobY, collapsed: isJobCollapsed, tagText: `${colIndex + 1}-${withinIndex + 1}` })
            stats.jobs++
            stats.headers += result?.stats?.headers || 0
            stats.steps += result?.stats?.steps || 0
            stats.edges += result?.stats?.edges || 0
            jobContainerById[jid] = result.jobCell
            const finalH = result.jobHeight || estH
            approxCenterY[jid] = jobY + Math.max(estH, finalH) / 2
            const jobBottom = result.jobBottom
            columnBottom = Math.max(columnBottom, jobBottom)
            maxBottomY = Math.max(maxBottomY, jobBottom)
            withinIndex += 1
            lastJobIdInColumn = jid
          })
          // 推进列游标到当前分组实际底部之后，避免对后续分组造成不必要的整体下移。
          columnYCursor = Math.max(columnYCursor, columnBottom + columnRowGap)
          return
        }

        if (g.kind === 'multi') {
          const bandTop = g.top
          const bandBottom = g.bottom
          // 多父分组也改为以父区间 slot center 对齐，不使用 startY 进行下推；
          let startY = columnYCursor
          const jobs = [...g.jobs]
          const m = jobs.length
          const slotCenters: number[] = []
          for (let i = 1; i <= m; i++) {
            slotCenters.push(bandTop + ((bandBottom - bandTop) * i) / (m + 1))
          }
          let groupBottom = startY
          jobs.forEach((jid, idx) => {
            const j = jobsObj[jid] || {}
            const isJobCollapsed = !!(collapsed || (collapsedJobs && collapsedJobs[jid]))
            const estH = estimateJobHeight(jid)
            const targetCenter = slotCenters[idx]
            // 直接对齐父区间 slot 的中心，无额外偏移
            const candidateY = Math.floor(targetCenter - estH / 2)
            const jobY = Math.max(baseY, candidateY)
            const result = renderJobContainer({ graph, jid, job: j, jobX, jobY, collapsed: isJobCollapsed, tagText: `${colIndex + 1}-${withinIndex + 1}` })
            stats.jobs++
            stats.headers += result?.stats?.headers || 0
            stats.steps += result?.stats?.steps || 0
            stats.edges += result?.stats?.edges || 0
            jobContainerById[jid] = result.jobCell
            const estHeight = estimateJobHeight(jid)
            approxCenterY[jid] = jobY + Math.max(estHeight, result.jobHeight || estHeight) / 2
            const jobBottom = result.jobBottom
            columnBottom = Math.max(columnBottom, jobBottom)
            maxBottomY = Math.max(maxBottomY, jobBottom)
            groupBottom = Math.max(groupBottom, jobBottom)
            withinIndex += 1
            lastJobIdInColumn = jid
          })
          // 推进列游标到当前分组实际底部之后，避免对后续分组造成整体下移。
          columnYCursor = Math.max(columnYCursor, groupBottom + columnRowGap)
          return
        }

        // 未分组：按 YAML 顺序堆叠
        if (g.kind === 'ungrouped') {
          g.jobs.forEach((jid) => {
            const j = jobsObj[jid] || {}
            const isJobCollapsed = !!(collapsed || (collapsedJobs && collapsedJobs[jid]))
            const jobY = columnYCursor
            const result = renderJobContainer({ graph, jid, job: j, jobX, jobY, collapsed: isJobCollapsed, tagText: `${colIndex + 1}-${withinIndex + 1}` })
            stats.jobs++
            stats.headers += result?.stats?.headers || 0
            stats.steps += result?.stats?.steps || 0
            stats.edges += result?.stats?.edges || 0
            jobContainerById[jid] = result.jobCell
            const estHeight = estimateJobHeight(jid)
            approxCenterY[jid] = jobY + Math.max(estHeight, result.jobHeight || estHeight) / 2
            const jobBottom = result.jobBottom
            columnBottom = jobBottom
            maxBottomY = Math.max(maxBottomY, jobBottom)
            columnYCursor += result.jobHeight + columnRowGap
            withinIndex += 1
            lastJobIdInColumn = jid
          })
        }
      })
    } else {
      // 非自动布局：按 YAML 顺序从上至下堆叠
      jobsAtLevel.forEach((jid) => {
        const j = jobsObj[jid] || {}
        const isJobCollapsed = !!(collapsed || (collapsedJobs && collapsedJobs[jid]))
        const jobY = columnYCursor
        const result = renderJobContainer({ graph, jid, job: j, jobX, jobY, collapsed: isJobCollapsed, tagText: `${colIndex + 1}-${withinIndex + 1}` })
        stats.jobs++
        stats.headers += result?.stats?.headers || 0
        stats.steps += result?.stats?.steps || 0
        stats.edges += result?.stats?.edges || 0
        jobContainerById[jid] = result.jobCell
        const estHeight = estimateJobHeight(jid)
        approxCenterY[jid] = jobY + Math.max(estHeight, result.jobHeight || estHeight) / 2
        const jobBottom = result.jobBottom
        columnBottom = jobBottom
        maxBottomY = Math.max(maxBottomY, jobBottom)
        columnYCursor += result.jobHeight + columnRowGap
        withinIndex += 1
        lastJobIdInColumn = jid
      })
    }

    // 垂直扫线：对本列已放置的 Job 进行一次下推消除重叠
    try {
      const placedCells = (levelToJobs[lv] || [])
        .map((jid) => jobContainerById[jid])
        .filter((c) => !!c)
      if (placedCells.length) {
        const sortedByY = [...placedCells].sort((a: any, b: any) => {
          try { return (a.getBBox().y || 0) - (b.getBBox().y || 0) } catch (_) { return 0 }
        })
        let sweepCursor = baseY
        sortedByY.forEach((cell: any) => {
          try {
            const bb = cell.getBBox()
            const newY = Math.max(bb.y, sweepCursor)
            if (newY !== bb.y) {
              try { cell.position(bb.x, newY) } catch (_) {}
            }
            const jid = String(cell.id || '').replace(/-container$/, '')
            const estH = estimateJobHeight(jid)
            approxCenterY[jid] = newY + Math.max(estH, bb?.height || estH) / 2
            const bottom = newY + (bb?.height || estH)
            columnBottom = Math.max(columnBottom, bottom)
            maxBottomY = Math.max(maxBottomY, bottom)
            try { reflowStepsInJob(graph, jid) } catch (_) {}
            sweepCursor = bottom + columnRowGap
          } catch (_) {}
        })
      }
    } catch (_) {}

    columnBounds[colIndex] = { top: baseY, bottom: columnBottom }
    // 已移除列底部“+ 增加job”占位，交互统一由右键与依赖边触发
    // 无需记录出口行，依赖对齐与分组展开已使最后一列靠近其前驱
  })

  try { stats.edges += linkJobDependencies({ graph, jobs: jobsObj, jobContainerById, levels, layout: { columnSpacing } }) } catch (_) {}
  // 渲染完成后，进行一次全局约束，确保步骤与服务节点都被夹紧在所属 Job 容器内
  try { applyGlobalConstraints(graph) } catch (_) {}
  return { stats, jobContainerById, columnBounds, maxBottomY }
}

// 约束：限制步骤/服务节点只能位于所属 Job 容器内部
export const clampNodeIntoJob = (graph: Graph, node: any, setLock?: (v: boolean) => void) => {
  try {
    const dt = node?.getData?.() || node?.data || {}
    const t = dt?.type
    const jid = dt?.jobId
    if (!jid) return
    if (!(t === 'stage' || t === 'conditional' || t === 'always' || t === 'service' || t === 'step')) return
    const job = graph.getCellById(`${jid}-container`)
    if (!job) return
    const p = job.getBBox()
    const b = node.getBBox()
    const pad = 10
    const minX = p.x + pad
    const maxX = p.x + p.width - b.width - pad
    const minY = p.y + pad
    const maxY = p.y + p.height - b.height - pad
    let nx = Math.max(minX, Math.min(b.x, maxX))
    let ny = Math.max(minY, Math.min(b.y, maxY))
    if (nx !== b.x || ny !== b.y) {
      if (setLock) setLock(true)
      try { node.position(nx, ny) } catch (_) { try { node.setPosition(nx, ny) } catch (_) {} }
      if (setLock) setLock(false)
    }
  } catch (_) {}
}

export const clampAllStepsInJob = (graph: Graph, jid: string, setLock?: (v: boolean) => void) => {
  try {
    const nodes = graph.getNodes()
    nodes.forEach((n) => {
      const dt = n?.getData?.() || n?.data || {}
      const t = dt?.type
      if ((t === 'stage' || t === 'conditional' || t === 'always' || t === 'service' || t === 'step') && dt?.jobId === jid) {
        clampNodeIntoJob(graph, n, setLock)
      }
    })
  } catch (_) {}
}

export const reflowStepsInJob = (graph: Graph, jid: string, setLock?: (v: boolean) => void) => {
  try {
    const job = graph.getCellById(`${jid}-container`)
    if (!job) return
    const jb = job.getBBox()
    const headerH = 36
    const STEP_HEIGHT = 34
    const STEP_GAP = 6
    const paddingX = 10
    let y = jb.y + headerH + 12
    const nodes = graph.getNodes()
    const list = nodes.filter((n) => {
      const dt = n?.getData?.() || n?.data || {}
      const t = dt?.type
      return dt?.jobId === jid && (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always')
    })
    list.sort((a: any, b: any) => {
      try {
        const da = a?.getData?.() || a?.data || {}
        const db = b?.getData?.() || b?.data || {}
        const ia = Number(da?.stepIndex ?? 0)
        const ib = Number(db?.stepIndex ?? 0)
        if (ia !== ib) return ia - ib
        const aid = String(a?.id || '')
        const bid = String(b?.id || '')
        return aid.localeCompare(bid)
      } catch (_) { return 0 }
    })
    list.forEach((n) => {
      try {
        const nx = jb.x + paddingX
        const ny = y
        const b = n.getBBox()
        if (b.x !== nx || b.y !== ny) {
          if (setLock) setLock(true)
          try { n.position(nx, ny) } catch (_) { try { (n as any).setPosition?.(nx, ny) } catch (_) {} }
          if (setLock) setLock(false)
        }
        y += STEP_HEIGHT + STEP_GAP
      } catch (_) {}
    })
    ensureJobFitsChildren(graph, jid, setLock)
  } catch (_) {}
}

// 渲染后全局约束：保证所有步骤/服务节点位于其所属 Job 容器内
export const applyGlobalConstraints = (graph: Graph) => {
  try {
    const nodes = graph.getNodes()
    nodes.forEach((n) => {
      try {
        const dt = n?.getData?.() || n?.data || {}
        const t = dt?.type
        const jid = dt?.jobId
        if ((t === 'stage' || t === 'conditional' || t === 'always' || t === 'service' || t === 'step') && jid) {
          clampNodeIntoJob(graph, n)
        }
      } catch (_) {}
    })
  } catch (_) {}
}

// 保证 Job 容器高度足以容纳其子节点（含 step），若不足则自动增高
export const ensureJobFitsChildren = (graph: Graph, jid: string, setLock?: (v: boolean) => void) => {
  try {
    const job = graph.getCellById(`${jid}-container`)
    if (!job) return
    const jb = job.getBBox()
    const pad = 24
    let maxBottom = jb.y + jb.height
    try {
      const nodes = graph.getNodes()
      nodes.forEach((n) => {
        try {
          const dt = n?.getData?.() || n?.data || {}
          const t = dt?.type
          if (dt?.jobId === jid && (t === 'stage' || t === 'conditional' || t === 'always' || t === 'service' || t === 'step')) {
            const b = n.getBBox()
            maxBottom = Math.max(maxBottom, b.y + b.height + pad)
          }
        } catch (_) {}
      })
    } catch (_) {}
    const requiredHeight = Math.max(jb.height, maxBottom - jb.y)
    if (requiredHeight > jb.height) {
      if (setLock) setLock(true)
      try { (job as any).resize?.(jb.width, requiredHeight) } catch (_) { try { (job as any).size?.(jb.width, requiredHeight) } catch (_) { try { (job as any).setSize?.(jb.width, requiredHeight) } catch (_) { try { (job as any).prop?.('size', { width: jb.width, height: requiredHeight }) } catch (_) {} } } }
      if (setLock) setLock(false)
      // 扩容后，重新夹紧所有子节点，避免越界
      clampAllStepsInJob(graph, jid, setLock)
      // 同步标题位置到容器顶部（保持锚点偏移）
      try {
        const header = graph.getCellById(`${jid}-header`)
        if (header) {
          const dt = header?.getData?.() || header?.data || {}
          const ox = Number(dt?.anchorOffsetX ?? 10)
          const oy = Number(dt?.anchorOffsetY ?? -20)
          const nx = (job.getBBox().x || jb.x) + ox
          const ny = (job.getBBox().y || jb.y) + oy
          const hb = header.getBBox()
          if (hb.x !== nx || hb.y !== ny) {
            if (setLock) setLock(true)
            try {
              const h: any = header
              if (typeof h?.position === 'function') {
                h.position(nx, ny)
              }
            } catch (_) {}
            if (setLock) setLock(false)
          }
        }
      } catch (_) {}
    }
  } catch (_) {}
}

// 事件绑定：交互点击与双击（页面通过 emit 接收）
export const bindPipelineInteractions = ({ graph, emit, getCollapsedJobs }: { graph: Graph; emit?: Function; getCollapsedJobs?: () => Record<string, boolean> }) => {
  try {
    const onClick = (evt: any) => {
      try {
        const node = evt?.node
        const dt = node?.getData?.() || node?.data || {}
        const t = dt?.type
        // 移除 DEL 标签相关点击行为
        // 已移除列底部“增加job”占位及其点击行为
        if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') {
          const id = node.id || ''
          const jid = id.replace(/-header$/, '')
          if (emit) emit('show-job', { jobId: jid })
          return
        }
        if (t === 'stage' || t === 'conditional' || t === 'always' || t === 'step') {
          const payload = dt?.payload
          const jobId = dt?.jobId
          const stepIndex = dt?.stepIndex
          if (emit) emit('show-step', { step: payload || { name: node?.label || node?.id }, jobId, stepIndex })
          return
        }
      } catch (_) {}
    }
    graph.on('node:click', onClick)

    const onDblClick = (evt: any) => {
      try {
        const node = evt?.node
        const dt = node?.getData?.() || node?.data || {}
        const t = dt?.type
        if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') {
          const id = node.id || ''
          const jid = id.replace(/-header$/, '')
          const current = (getCollapsedJobs && getCollapsedJobs()) || {}
          const next = { ...current }
          next[jid] = !next[jid]
          if (emit) emit('update:collapsedJobs', next)
        }
      } catch (_) {}
    }
    graph.on('node:dblclick', onDblClick)

    // 右键（contextmenu）：在 Start 或 Job 标题处弹出“添加串/并行 Job”选择
    const onContextMenu = (evt: any) => {
      try {
        // 阻止浏览器默认右键菜单，提升交互体验
        try { evt?.e?.preventDefault?.() } catch (_) {}
        const node = evt?.node
        const dt = node?.getData?.() || node?.data || {}
        const t = dt?.type
        const cx = Number(evt?.e?.clientX || 0)
        const cy = Number(evt?.e?.clientY || 0)
        // 已移除列底部“增加job”占位及其右键行为
        if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') {
          // Job 标题：右键后在该 Job 之后追加/并行一个新 Job
          const id = node?.id || ''
          const jid = id.replace(/-header$/, '')
          try { (graph as any).trigger('pipeline:create-job', { anchor: 'between', prevJobId: jid, nextJobId: '', x: cx, y: cy }) } catch (_) {}
          return
        }
        if (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always') {
          // Step 节点：右键弹出步骤操作菜单（插入/编辑/删除）
          const payload = {
            x: cx,
            y: cy,
            jobId: String(dt?.jobId || ''),
            stepIndex: Number(dt?.stepIndex ?? -1),
            step: dt?.payload || {},
            nodeId: node?.id || '',
          }
          try { (graph as any).trigger('pipeline:step-actions', payload) } catch (_) {}
          return
        }
      } catch (_) {}
    }
    try { graph.on('node:contextmenu', onContextMenu) } catch (_) {}

    // 画布空白处右键：默认作为入口 Job（Start 之后），并携带坐标用于就地弹出菜单
    const onBlankContextMenu = (evt: any) => {
      try {
        try { evt?.e?.preventDefault?.() } catch (_) {}
        const cx = Number(evt?.e?.clientX || 0)
        const cy = Number(evt?.e?.clientY || 0)
        try { (graph as any).trigger('pipeline:create-job', { anchor: 'after-start', x: cx, y: cy }) } catch (_) {}
      } catch (_) {}
    }
    try { graph.on('blank:contextmenu', onBlankContextMenu) } catch (_) {}

    const shouldHighlight = (t: string) => (t === 'job' || t === 'header' || t === 'headerConditional' || t === 'headerAlways' || t === 'step' || t === 'stage' || t === 'conditional' || t === 'always')
    const getStrokeForType = (t: string) => (PIPELINE_STYLE_MAP[t]?.stroke || PIPELINE_STYLE_MAP.default.stroke)
    const getAccentForType = (t: string) => (PIPELINE_STYLE_MAP[t]?.accent || PIPELINE_STYLE_MAP.default.accent)
    const onMouseEnter = (evt: any) => {
      try {
        const node = evt?.node
        const dt = node?.getData?.() || node?.data || {}
        const t = String(dt?.type || '')
        if (!shouldHighlight(t)) return
        const color = getAccentForType(t)
        try { node.attr('body/stroke', color) } catch (_) {}
        try { node.attr('body/strokeWidth', 3) } catch (_) {}
      } catch (_) {}
    }
    const onMouseLeave = (evt: any) => {
      try {
        const node = evt?.node
        const dt = node?.getData?.() || node?.data || {}
        const t = String(dt?.type || '')
        if (!shouldHighlight(t)) return
        const color = getStrokeForType(t)
        try { node.attr('body/stroke', color) } catch (_) {}
        try { node.attr('body/strokeWidth', 2) } catch (_) {}
      } catch (_) {}
    }
    try { graph.on('node:mouseenter', onMouseEnter) } catch (_) {}
    try { graph.on('node:mouseleave', onMouseLeave) } catch (_) {}

    // 统一监听边标签点击：根据边 ID 或端点判断插入位置，触发创建 Job
    const onEdgeLabelClick = (evt: any) => {
      try {
        const edge = evt?.edge || evt
        if (!edge) return
        let payload: any = null
        const id: string = String((edge as any)?.id || '')
        // 依赖边：dep-edge-<prev>__<next>
        const m = id.match(/^dep-edge-(.+?)__(.+)$/)
        if (m && m[1] && m[2]) {
          const cx = Number(evt?.e?.clientX || evt?.e?.x || 0)
          const cy = Number(evt?.e?.clientY || evt?.e?.y || 0)
          payload = { anchor: 'between', prevJobId: m[1], nextJobId: m[2], x: cx, y: cy }
        }
        if (payload) {
          try { (graph as any).trigger('pipeline:create-job', payload) } catch (_) {}
        }
      } catch (_) {}
    }
    try { graph.on('edge:label:click', onEdgeLabelClick) } catch (_) {}

    // 统一监听整条边点击：作为兜底，提升可点击性
    const onEdgeClick = (evt: any) => {
      try {
        const edge = evt?.edge
        if (!edge) return
        let payload: any = null
        const id: string = String((edge as any)?.id || '')
        const m = id.match(/^dep-edge-(.+?)__(.+)$/)
        if (m && m[1] && m[2]) {
          const cx = Number(evt?.e?.clientX || evt?.e?.x || 0)
          const cy = Number(evt?.e?.clientY || evt?.e?.y || 0)
          payload = { anchor: 'between', prevJobId: m[1], nextJobId: m[2], x: cx, y: cy }
        }
        if (payload) {
          try { (graph as any).trigger('pipeline:create-job', payload) } catch (_) {}
        }
      } catch (_) {}
    }
    try { graph.on('edge:click', onEdgeClick) } catch (_) {}

    const cleanup = () => {
      try {
        graph.off('node:click', onClick)
        graph.off('node:dblclick', onDblClick)
        try { graph.off('node:contextmenu', onContextMenu) } catch (_) {}
        try { graph.off('blank:contextmenu', onBlankContextMenu) } catch (_) {}
        try { graph.off('edge:label:click', onEdgeLabelClick) } catch (_) {}
        try { graph.off('edge:click', onEdgeClick) } catch (_) {}
        try { graph.off('node:mouseenter', onMouseEnter) } catch (_) {}
        try { graph.off('node:mouseleave', onMouseLeave) } catch (_) {}
      } catch (_) {}
    }
    ;(graph as any).__pluginBindings = (graph as any).__pluginBindings || {}
    ;(graph as any).__pluginBindings.interactionsCleanup = cleanup
    return cleanup
  } catch (_) {}
}

// 事件绑定：移动/位置/尺寸变化约束
export const bindConstraintHandlers = (graph: Graph, setLock?: (v: boolean) => void, isLocked?: () => boolean) => {
  try {
    const locked = () => (isLocked ? !!isLocked() : false)
    // 固定 Job 头部右侧的 DEL 标签到头部位置
    const syncHeaderDelToHeader = (_headerNode: any) => {}
    // 固定 Job 标号标签（如“1-1”）到头部左侧位置
    const syncHeaderTagToHeader = (headerNode: any) => {
      try {
        const dt = headerNode?.getData?.() || headerNode?.data || {}
        const type = dt?.type
        if (!(type === 'header' || type === 'headerConditional' || type === 'headerAlways')) return
        const jid = String(dt?.jobId || (headerNode?.id || '').replace(/-header$/, ''))
        if (!jid) return
        const tagId = `${jid}-tag`
        const tagNode: any = graph.getCellById(tagId)
        if (!tagNode) return
        const hb = headerNode.getBBox()
        const tagW = 36
        const margin = 6
        const nx = hb.x + margin
        const ny = hb.y + 8
        const tb = tagNode.getBBox?.()
        if (!tb || tb.x !== nx || tb.y !== ny) {
          if (setLock) setLock(true)
          try { tagNode.position(nx, ny) } catch (_) { try { tagNode.setPosition(nx, ny) } catch (_) {} }
          if (setLock) setLock(false)
        }
      } catch (_) {}
    }
    // 固定 Step 右侧的 DEL 标签到 Step 卡片位置
    const syncStepDelToStep = (_stepNode: any) => {}
    const syncHeaderToJob = (headerNode: any) => {
      try {
        const dt = headerNode?.getData?.() || headerNode?.data || {}
        const type = dt?.type
        if (!(type === 'header' || type === 'headerConditional' || type === 'headerAlways')) return
        const jid = String(dt?.jobId || (headerNode?.id || '').replace(/-header$/, ''))
        if (!jid) return
        const job = graph.getCellById(`${jid}-container`)
        if (!job) return
        const jb = job.getBBox()
        const ox = Number(dt?.anchorOffsetX ?? 10)
        const oy = Number(dt?.anchorOffsetY ?? -20)
        const nx = jb.x + ox
        const ny = jb.y + oy
        const hb = headerNode.getBBox()
        if (hb.x !== nx || hb.y !== ny) {
          if (setLock) setLock(true)
          try { headerNode.position(nx, ny) } catch (_) { try { headerNode.setPosition(nx, ny) } catch (_) {} }
          if (setLock) setLock(false)
        }
        // 同步头部右侧 DEL 标签与左侧编号标签
        syncHeaderDelToHeader(headerNode)
        syncHeaderTagToHeader(headerNode)
      } catch (_) {}
    }
    const onMoving = (evt: any) => {
      const node = evt?.node
      if (locked()) return
      const dt = node?.getData?.() || node?.data || {}
      const t = dt?.type
      // 禁止拖动 Job 编号标签（如“1-1”）：立刻回到锚点
      if (t === 'tag' && String(node?.id || '').endsWith('-tag')) {
        const jid = String((node?.id || '').replace(/-tag$/, ''))
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderTagToHeader(header)
        return
      }
      // 禁止拖动 DEL 标签：无论拖到哪里，立刻回到锚点位置
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-job') {
        const jid = String(dt?.jobId || '')
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderDelToHeader(header)
        return
      }
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-step') {
        const jid = String(dt?.jobId || '')
        const idx = Number(dt?.stepIndex ?? -1)
        if (jid && idx >= 0) {
          const stepId = `${jid}-s${idx + 1}`
          const step = graph.getCellById(stepId)
          if (step) syncStepDelToStep(step)
        }
        return
      }
      if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') { syncHeaderToJob(node); return }
      if (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always' || t === 'service') {
        const jid = String(dt?.jobId || '')
        if (jid) ensureJobFitsChildren(graph, jid, setLock)
        clampNodeIntoJob(graph, node, setLock)
        
        return
      }
      clampNodeIntoJob(graph, node, setLock)
    }
    const onChangePosition = (evt: any) => {
      if (locked()) return
      const node = evt?.node
      const dt = node?.getData?.() || node?.data || {}
      const t = dt?.type
      // 禁止拖动 Job 编号标签：位置变更时回到锚点
      if (t === 'tag' && String(node?.id || '').endsWith('-tag')) {
        const jid = String((node?.id || '').replace(/-tag$/, ''))
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderTagToHeader(header)
        return
      }
      // 禁止拖动 DEL 标签：位置变更时回到锚点
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-job') {
        const jid = String(dt?.jobId || '')
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderDelToHeader(header)
        return
      }
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-step') {
        const jid = String(dt?.jobId || '')
        const idx = Number(dt?.stepIndex ?? -1)
        if (jid && idx >= 0) {
          const stepId = `${jid}-s${idx + 1}`
          const step = graph.getCellById(stepId)
          if (step) syncStepDelToStep(step)
        }
        return
      }
      if (t === 'job') {
        const jid = String(node?.id || '').replace(/-container$/, '')
        clampAllStepsInJob(graph, jid, setLock)
        // 保持标题与容器绑定：同步位置
        const header = graph.getCellById(`${jid}-header`)
        if (header) { syncHeaderToJob(header); syncHeaderTagToHeader(header) }
        return
      }
      if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') { syncHeaderToJob(node); return }
      if (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always' || t === 'service') {
        const jid = String(dt?.jobId || '')
        if (jid) ensureJobFitsChildren(graph, jid, setLock)
        clampNodeIntoJob(graph, node, setLock)
        
        return
      }
      clampNodeIntoJob(graph, node, setLock)
    }
    const onChangeSize = (evt: any) => {
      if (locked()) return
      const node = evt?.node
      const dt = node?.getData?.() || node?.data || {}
      const t = dt?.type
      if (t === 'job') {
        const jid = String(node?.id || '').replace(/-container$/, '')
        clampAllStepsInJob(graph, jid, setLock)
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderToJob(header)
        try { reflowStepsInJob(graph, jid, setLock) } catch (_) {}
        return
      }
      if (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always' || t === 'service') {
        const jid = String(dt?.jobId || '')
        if (jid) ensureJobFitsChildren(graph, jid, setLock)
        
      }
    }
    const onMoved = (evt: any) => {
      const node = evt?.node
      if (locked()) return
      const dt = node?.getData?.() || node?.data || {}
      const t = dt?.type
      // 禁止拖动 Job 编号标签：拖动结束时回到锚点
      if (t === 'tag' && String(node?.id || '').endsWith('-tag')) {
        const jid = String((node?.id || '').replace(/-tag$/, ''))
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderTagToHeader(header)
        return
      }
      // 禁止拖动 DEL 标签：拖动结束时回到锚点
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-job') {
        const jid = String(dt?.jobId || '')
        const header = graph.getCellById(`${jid}-header`)
        if (header) syncHeaderDelToHeader(header)
        return
      }
      if ((t === 'tag' || t === 'service') && dt?.kind === 'delete-step') {
        const jid = String(dt?.jobId || '')
        const idx = Number(dt?.stepIndex ?? -1)
        if (jid && idx >= 0) {
          const stepId = `${jid}-s${idx + 1}`
          const step = graph.getCellById(stepId)
          if (step) syncStepDelToStep(step)
        }
        return
      }
      if (t === 'header' || t === 'headerConditional' || t === 'headerAlways') { syncHeaderToJob(node); return }
      if (t === 'step' || t === 'stage' || t === 'conditional' || t === 'always' || t === 'service') {
        const jid = String(dt?.jobId || '')
        if (jid) ensureJobFitsChildren(graph, jid, setLock)
        clampNodeIntoJob(graph, node, setLock)
        if (t !== 'service') syncStepDelToStep(node)
        return
      }
      clampNodeIntoJob(graph, node, setLock)
    }

    graph.on('node:moving', onMoving)
    graph.on('node:change:position', onChangePosition)
    graph.on('node:change:size', onChangeSize)
    graph.on('node:moved', onMoved)

    const cleanup = () => {
      try {
        graph.off('node:moving', onMoving)
        graph.off('node:change:position', onChangePosition)
        graph.off('node:change:size', onChangeSize)
        graph.off('node:moved', onMoved)
      } catch (_) {}
    }
    ;(graph as any).__pluginBindings = (graph as any).__pluginBindings || {}
    ;(graph as any).__pluginBindings.constraintsCleanup = cleanup
    return cleanup
  } catch (_) {}
}

// 创建 Graph：统一初始化配置
export const createPipelineGraph = ({ container, width, height }: { container: HTMLElement; width: number; height: number }) => {
  const graph = new Graph({
    container,
    width,
    height,
    grid: { size: 10, visible: false },
    panning: { enabled: true },
    mousewheel: { enabled: true, modifiers: ['ctrl', 'meta'] },
    embedding: { enabled: false },
    connecting: { allowBlank: false, snap: true },
    autoResize: true,
  })
  return graph
}

// 解绑事件：调用已注册的清理方法
export const unbindPipelineHandlers = (graph: Graph) => {
  try {
    const b = graph && (graph as any).__pluginBindings
    if (b?.interactionsCleanup) { try { b.interactionsCleanup() } catch (_) {} }
    if (b?.constraintsCleanup) { try { b.constraintsCleanup() } catch (_) {} }
    if (b?.edgeLabelCleanup) { try { b.edgeLabelCleanup() } catch (_) {} }
    if (b?.edgeLabelCleanupList && Array.isArray(b.edgeLabelCleanupList)) { try { b.edgeLabelCleanupList.forEach((fn: Function) => { try { fn() } catch (_) {} }) } catch (_) {} }
    if (b?.edgeClickCleanup) { try { b.edgeClickCleanup() } catch (_) {} }
  } catch (_) {}
}

// 视图辅助：居中、缩放、重置
export const fitView = (graph: Graph) => { try { if (graph) graph.zoomToFit({ padding: 12, minScale: 0.2, maxScale: 1 }) } catch (_) {} }
export const zoomIn = (graph: Graph) => { try { if (graph) graph.zoom(0.1) } catch (_) {} }
export const zoomOut = (graph: Graph) => { try { if (graph) graph.zoom(-0.1) } catch (_) {} }
export const resetZoom = (graph: Graph) => { try { if (graph) graph.zoomTo(1) } catch (_) {} }
export const resetGraph = ({ graph, doc, collapsed, collapsedJobs, layout }: { graph: Graph; doc: any; collapsed: boolean; collapsedJobs: Record<string, boolean>; layout?: { columnSpacing?: number; baseX?: number; baseY?: number; columnRowGap?: number } }) => {
  try {
    if (!graph || !doc) return
    renderJobsPipeline({ graph, doc, collapsed, collapsedJobs, layout })
    fitView(graph)
  } catch (_) {}
}
