// 节点样式：颜色与基本 attrs 提取到独立文件，便于复用
export const PIPELINE_STYLE_MAP: Record<string, { stroke: string; fill: string; accent: string }> = {
  stage: { stroke: '#8c8c8c', fill: '#f5f7fa', accent: '#8c8c8c' },
  step: { stroke: 'transparent', fill: '#f5f7fa', accent: '#8c8c8c' },
  job: { stroke: 'transparent', fill: '#ffffff', accent: '#52c41a' },
  success: { stroke: '#67C23A', fill: '#F6FFED', accent: '#67C23A' },
  failed: { stroke: '#F56C6C', fill: '#FEF0F0', accent: '#F56C6C' },
  running: { stroke: 'var(--el-color-primary)', fill: '#ECF5FF', accent: 'var(--el-color-primary)' },
  conditional: { stroke: '#E6A23C', fill: '#FFF7E6', accent: '#E6A23C' },
  always: { stroke: '#909399', fill: '#F2F3F5', accent: '#909399' },
  header: { stroke: 'transparent', fill: '#52c41a', accent: '#52c41a' },
  headerConditional: { stroke: 'transparent', fill: '#52c41a', accent: '#52c41a' },
  headerAlways: { stroke: 'transparent', fill: '#52c41a', accent: '#52c41a' },
  service: { stroke: '#8c8c8c', fill: '#FAFAFA', accent: '#8c8c8c' },
  start: { stroke: '#67C23A', fill: '#FFFFFF', accent: '#67C23A' },
  end: { stroke: '#F56C6C', fill: '#FFFFFF', accent: '#F56C6C' },
  tag: { stroke: '#ebeef5', fill: '#eef2f7', accent: '#8c8c8c' },
  accent: { stroke: '#52c41a', fill: '#52c41a', accent: '#52c41a' },
  default: { stroke: '#8c8c8c', fill: '#fff', accent: '#8c8c8c' },
}

export const pipelineNodeAttrs = (type: string) => {
  const s = PIPELINE_STYLE_MAP[type] || PIPELINE_STYLE_MAP.default
  return {
    body: {
      stroke: s.stroke,
      fill: s.fill,
      rx: type === 'header' ? 4 : 4,
      ry: type === 'header' ? 4 : 4,
      filter: type === 'step' ? 'none' : (type === 'accent' ? 'none' : 'drop-shadow(0 2px 6px rgba(0,0,0,0.08))'),
      ...(type === 'service' ? { strokeDasharray: '4,4' } : {}),
      ...(type === 'accent' ? { rx: 0, ry: 0 } : {}),
      ...(type === 'step' ? { rx: 3, ry: 3 } : {}),
    },
    label:
      (type === 'header' || type === 'headerConditional' || type === 'headerAlways')
        ? { fill: '#ffffff', fontSize: 14, fontWeight: 700, refX: 12, refY: 22 }
        : (type === 'job')
          ? { fill: '#303133', fontSize: 14, fontWeight: 700, refX: 12, refY: 22 }
        : (type === 'start' || type === 'end')
          ? { fill: '#333', fontSize: 14, fontWeight: 600, refX: 12, refY: 22 }
        : (type === 'step')
            ? { fill: '#606266', fontSize: 12, fontWeight: 500, refX: 16, refY: 20, textWrap: { width: 220, height: 20, ellipsis: true } }
            : (type === 'tag')
              ? { fill: '#606266', fontSize: 12, fontWeight: 500, refX: 10, refY: 12 }
            : { fill: '#333', fontSize: 14, fontWeight: 400, refX: 12, refY: 22 },
  }
}
