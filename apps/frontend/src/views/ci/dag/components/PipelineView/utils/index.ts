// 通用工具函数：减少在多个组件中的重复实现

export type IfType = 'always' | 'conditional' | ''

export type MatrixLike = Record<string, (string | number | boolean)[] | unknown>
export type JobLike = { strategy?: { matrix?: MatrixLike }; if?: string }
export type StepLike = { name?: string; uses?: string; run?: unknown; if?: string }

// 计算 matrix 提示，形如 "2×3=6"
export const matrixHint = (job: JobLike | any): string => {
  try {
    const matrix = job?.strategy?.matrix
    if (!matrix || typeof matrix !== 'object') return ''
    const dims = Object.entries(matrix as Record<string, unknown>)
      .filter(([key, val]) => Array.isArray(val) && key !== 'include' && key !== 'exclude' && (val as unknown[]).every((el) => el !== null && typeof el !== 'object'))
      .map(([, arr]) => (arr as unknown[]).length)
    if (!dims.length) return ''
    const product = dims.reduce((a, b) => a * b, 1)
    return `${dims.join('×')}=${product}`
  } catch (_) {
    return ''
  }
}

export type NormalizedStep = { label: string; ifType: IfType; name?: string; uses?: string; run?: unknown }

// 规格化 step：生成 label 与 ifType 等
export const normalizeStep = (step: StepLike | any): NormalizedStep => {
  try {
    const label = step?.name || (step?.uses ? `uses ${step.uses}` : step?.run ? `run: ${String(step.run).slice(0, 64)}${String(step.run).length > 64 ? '...' : ''}` : 'step')
    const cond = step?.if ? String(step.if).toLowerCase() : ''
    const ifType: IfType = cond.includes('always()') ? 'always' : (cond ? 'conditional' : '')
    return { label, ifType, name: step?.name, uses: step?.uses, run: step?.run }
  } catch (_) {
    return { label: 'step', ifType: '', name: step?.name, uses: step?.uses, run: step?.run }
  }
}

// 判断 Job 的 if 类型
export const jobIfType = (job: JobLike | any): IfType => {
  try {
    const cond = job?.if ? String(job.if).toLowerCase() : ''
    if (!cond) return ''
    return cond.includes('always()') ? 'always' : 'conditional'
  } catch (_) {
    return ''
  }
}

export default { matrixHint, normalizeStep, jobIfType }