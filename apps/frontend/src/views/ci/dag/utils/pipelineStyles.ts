// 管道（Pipeline）节点样式与工厂：Job 使用左右端口（左入、右出），其它保持垂直（上入、下出）
// 样式相关从 nodeStyles.js 抽取
import { PIPELINE_STYLE_MAP, pipelineNodeAttrs } from '@/views/ci/dag/utils/nodeStyles'

export type PipelineNodeType = 'stage' | 'job' | 'success' | 'failed' | 'running' | 'conditional' | 'always' | 'header' | 'headerConditional' | 'headerAlways' | 'service' | 'start' | 'end' | 'default' | 'tag'
// 扩展一个用于内嵌色带的类型，沿用 rect 基本样式但颜色取 accent
export type AccentNodeType = PipelineNodeType | 'accent'
export const pipelineNodePorts = (type: AccentNodeType | string, id: string) => {
  const useLR = (type === 'job' || type === 'start' || type === 'end')
  const portStroke = '#8c8c8c'
  const groups = {
    in: {
      position: useLR ? 'left' : 'top',
      attrs: { circle: { r: 4, magnet: true, stroke: portStroke, strokeWidth: 1, fill: '#fff' } },
    },
    out: {
      position: useLR ? 'right' : 'bottom',
      attrs: { circle: { r: 4, magnet: true, stroke: portStroke, strokeWidth: 1, fill: '#fff' } },
    },
  }
  if (
    type === 'header' ||
    type === 'headerConditional' ||
    type === 'headerAlways' ||
    type === 'service' ||
    type === 'tag' ||
    type === 'accent' ||
    type === 'step' ||
    type === 'stage' ||
    type === 'conditional' ||
    type === 'always'
  ) {
    return { groups, items: [] }
  }
  return {
    groups,
    items: [
      { id: `${id}-in`, group: 'in' },
      { id: `${id}-out`, group: 'out' },
    ],
  }
}

// 计算不同类型节点的默认尺寸，避免内联复杂三元表达式造成语法问题
const defaultWidthForType = (type: PipelineNodeType | string) => {
  switch (type) {
    case 'header':
    case 'headerConditional':
    case 'headerAlways':
      return 220
    case 'service':
      return 160
    case 'job':
      return 320
    case 'start':
    case 'end':
      return 160
    default:
      return 220
  }
}

const defaultHeightForType = (type: PipelineNodeType | string) => {
  switch (type) {
    case 'header':
    case 'headerConditional':
    case 'headerAlways':
      return 36
    case 'service':
      return 36
    case 'job':
      return 200
    case 'start':
    case 'end':
      return 48
    default:
      return 60
  }
}

export type PipelineNodeInput = { id: string; x: number; y: number; label: string; type?: AccentNodeType | string; width?: number; height?: number; data?: any; payload?: any }
export const makePipelineNode = ({ id, x, y, label, type = 'stage', width: customWidth, height: customHeight, data, payload }: PipelineNodeInput) => ({
  id,
  x,
  y,
  width: customWidth ?? defaultWidthForType(type),
  height: customHeight ?? defaultHeightForType(type),
  shape: 'rect',
  label,
  attrs: pipelineNodeAttrs(type),
  ports: pipelineNodePorts(type, id),
  data: { type, ...(data || {}), ...(payload ? { payload } : {}) },
})
