// 边样式映射与工厂方法（当前提供黑色样式）
export const EDGE_STYLE_MAP: Record<string, { stroke: string; strokeWidth: number; targetMarker: { name: string; size: number }; strokeDasharray?: string }> = {
  black: { stroke: '#bfbfbf', strokeWidth: 2, targetMarker: { name: 'classic', size: 8 } },
  dashed: { stroke: '#c0c4cc', strokeWidth: 2, targetMarker: { name: 'classic', size: 8 }, strokeDasharray: '5,5' },
  step: { stroke: '#bfbfbf', strokeWidth: 2, targetMarker: { name: 'classic', size: 8 } },
  default: { stroke: '#bfbfbf', strokeWidth: 2, targetMarker: { name: 'classic', size: 8 } },
}

export const edgeAttrs = (type: string) => {
  const s = EDGE_STYLE_MAP[type] || EDGE_STYLE_MAP.default
  return {
    line: {
      stroke: s.stroke,
      strokeWidth: s.strokeWidth,
      targetMarker: s.targetMarker,
      ...(s.strokeDasharray ? { strokeDasharray: s.strokeDasharray } : {}),
    },
  }
}

export type EdgeEndpoint = { cell: string; port?: string }
export type EdgeLabel = { position?: number | { distance?: number }; attrs?: Record<string, any>; events?: Record<string, any> }
export type EdgeVertex = { x: number; y: number }
export type EdgeInput = { id?: string; source: EdgeEndpoint; target: EdgeEndpoint; type?: keyof typeof EDGE_STYLE_MAP; routerName?: string; routerArgs?: Record<string, any>; labels?: EdgeLabel[]; vertices?: EdgeVertex[]; stroke?: string; connectorName?: 'normal' | 'smooth' | 'rounded'; connectorArgs?: Record<string, any> }

export const makeEdge = ({ id, source, target, type = 'black', routerName = 'manhattan', routerArgs = { padding: 14 }, labels, vertices, stroke, connectorName = 'rounded', connectorArgs }: EdgeInput) => {
  const base = edgeAttrs(type)
  if (stroke) {
    base.line.stroke = stroke
  }
  return {
    id,
    shape: 'edge',
    source,
    target,
    attrs: base,
    // 允许关闭路由：当 routerName === 'none' 时不设置 router，让连接器直接连线
    ...(routerName && routerName !== 'none' ? { router: { name: routerName, args: routerArgs } } : {}),
    connector: { name: connectorName, ...(connectorArgs ? { args: connectorArgs } : {}) },
    ...(vertices && vertices.length ? { vertices } : {}),
    ...(labels && labels.length
      ? {
          labels: labels.map((l) => ({
            ...l,
            // 标签需要启用指针事件才能响应 edge:label:click
            attrs: {
              ...(l.attrs || {}),
              label: { ...(l.attrs?.label || {}), pointerEvents: 'auto', cursor: 'pointer' },
              body: { ...(l.attrs?.body || {}), pointerEvents: 'auto', cursor: 'pointer' },
            },
          })),
        }
      : {}),
  }
}
