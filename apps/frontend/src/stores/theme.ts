import { defineStore } from 'pinia'

export type ThemeVars = {
  colorPrimary: string
  sidebarBg: string
  sidebarBorder: string
  sidebarText: string
  sidebarActiveText: string
  headerBg: string
  headerText: string
  headerShadow: string
  mainBg: string
  // Element Plus 状态色
  colorSuccess?: string
  colorWarning?: string
  colorDanger?: string
  colorInfo?: string
}

const DEFAULT_THEME: ThemeVars = {
  colorPrimary: '#2563eb', // Modern Inter Blue
  sidebarBg: '#ffffff',    // Light sidebar
  sidebarBorder: '#e5e7eb',
  sidebarText: '#4b5563',
  sidebarActiveText: '#2563eb',
  headerBg: '#ffffff',
  headerText: '#1f2937',
  headerShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  mainBg: '#f9fafb',       // Very light gray background
  // Modern status colors (Tailwind-inspired)
  colorSuccess: '#10b981',
  colorWarning: '#f59e0b',
  colorDanger: '#ef4444',
  colorInfo: '#6b7280'
}

const STORAGE_KEY = 'xcoding.theme'

function clamp(num: number, min = 0, max = 255) { return Math.min(Math.max(num, min), max) }
function hexToRgb(hex: string) {
  const sanitized = hex.replace('#', '')
  const full = sanitized.length === 3 ? sanitized.split('').map(c => c + c).join('') : sanitized
  const bigint = parseInt(full, 16)
  const r = (bigint >> 16) & 255
  const g = (bigint >> 8) & 255
  const b = bigint & 255
  return { r, g, b }
}
function parseColor(input: string) {
  if (!input) return { r: 64, g: 158, b: 255 } // default Element Plus primary
  const trimmed = input.trim()
  if (trimmed.startsWith('#')) return hexToRgb(trimmed)
  const m = trimmed.match(/rgba?\((\s*\d+\s*),(\s*\d+\s*),(\s*\d+\s*)(?:,\s*(\d*\.?\d+)\s*)?\)/i)
  if (m) {
    return { r: Number(m[1]), g: Number(m[2]), b: Number(m[3]) }
  }
  // Fallback to default
  return { r: 64, g: 158, b: 255 }
}
function rgbToHex(r: number, g: number, b: number) {
  return '#' + [r, g, b].map(v => clamp(Math.round(v)).toString(16).padStart(2, '0')).join('')
}
function lighten(color: string, ratio: number) {
  const { r, g, b } = parseColor(color)
  return rgbToHex(r + (255 - r) * ratio, g + (255 - g) * ratio, b + (255 - b) * ratio)
}
function darken(color: string, ratio: number) {
  const { r, g, b } = parseColor(color)
  return rgbToHex(r * (1 - ratio), g * (1 - ratio), b * (1 - ratio))
}

function applyTheme(vars: ThemeVars) {
  const root = document.documentElement
  // 应用自定义主题变量
  root.style.setProperty('--color-primary', vars.colorPrimary)
  root.style.setProperty('--sidebar-bg', vars.sidebarBg)
  root.style.setProperty('--sidebar-border', vars.sidebarBorder)
  root.style.setProperty('--sidebar-text', vars.sidebarText)
  root.style.setProperty('--sidebar-active-text', vars.sidebarActiveText)
  root.style.setProperty('--header-bg', vars.headerBg)
  root.style.setProperty('--header-text', vars.headerText)
  root.style.setProperty('--header-shadow', vars.headerShadow)
  root.style.setProperty('--main-bg', vars.mainBg)

  // 让 Element Plus 组件主色也随主题变化
  root.style.setProperty('--el-color-primary', vars.colorPrimary)
  root.style.setProperty('--el-color-primary-light-3', lighten(vars.colorPrimary, 0.3))
  root.style.setProperty('--el-color-primary-light-5', lighten(vars.colorPrimary, 0.5))
  root.style.setProperty('--el-color-primary-light-7', lighten(vars.colorPrimary, 0.7))
  root.style.setProperty('--el-color-primary-dark-2', darken(vars.colorPrimary, 0.2))

  // EP 状态色随主题变量应用（如需自定义可由 vars 控制）
  const success = vars.colorSuccess || DEFAULT_THEME.colorSuccess!
  const warning = vars.colorWarning || DEFAULT_THEME.colorWarning!
  const danger = vars.colorDanger || DEFAULT_THEME.colorDanger!
  const info = vars.colorInfo || DEFAULT_THEME.colorInfo!

  // success
  root.style.setProperty('--el-color-success', success)
  root.style.setProperty('--el-color-success-light-3', lighten(success, 0.3))
  root.style.setProperty('--el-color-success-light-5', lighten(success, 0.5))
  root.style.setProperty('--el-color-success-light-7', lighten(success, 0.7))
  root.style.setProperty('--el-color-success-dark-2', darken(success, 0.2))
  // warning
  root.style.setProperty('--el-color-warning', warning)
  root.style.setProperty('--el-color-warning-light-3', lighten(warning, 0.3))
  root.style.setProperty('--el-color-warning-light-5', lighten(warning, 0.5))
  root.style.setProperty('--el-color-warning-light-7', lighten(warning, 0.7))
  root.style.setProperty('--el-color-warning-dark-2', darken(warning, 0.2))
  // danger/error（保持一致）
  root.style.setProperty('--el-color-danger', danger)
  root.style.setProperty('--el-color-danger-light-3', lighten(danger, 0.3))
  root.style.setProperty('--el-color-danger-light-5', lighten(danger, 0.5))
  root.style.setProperty('--el-color-danger-light-7', lighten(danger, 0.7))
  root.style.setProperty('--el-color-danger-dark-2', darken(danger, 0.2))
  root.style.setProperty('--el-color-error', danger)
  root.style.setProperty('--el-color-error-light-3', lighten(danger, 0.3))
  root.style.setProperty('--el-color-error-light-5', lighten(danger, 0.5))
  root.style.setProperty('--el-color-error-light-7', lighten(danger, 0.7))
  root.style.setProperty('--el-color-error-dark-2', darken(danger, 0.2))
  // info
  root.style.setProperty('--el-color-info', info)
  root.style.setProperty('--el-color-info-light-3', lighten(info, 0.3))
  root.style.setProperty('--el-color-info-light-5', lighten(info, 0.5))
  root.style.setProperty('--el-color-info-light-7', lighten(info, 0.7))
  root.style.setProperty('--el-color-info-dark-2', darken(info, 0.2))
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    vars: { ...DEFAULT_THEME } as ThemeVars
  }),
  actions: {
    load() {
      // 首次进入使用默认经典蓝灰；之后按存储恢复
      try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (raw) {
          this.vars = { ...DEFAULT_THEME, ...JSON.parse(raw) }
        } else {
          this.vars = { ...DEFAULT_THEME }
        }
      } catch (_) {
        this.vars = { ...DEFAULT_THEME }
      }
      applyTheme(this.vars)
    },
    set(vars: Partial<ThemeVars>) {
      this.vars = { ...this.vars, ...vars }
      localStorage.setItem(STORAGE_KEY, JSON.stringify(this.vars))
      applyTheme(this.vars)
    },
    reset() {
      this.vars = { ...DEFAULT_THEME }
      localStorage.setItem(STORAGE_KEY, JSON.stringify(this.vars))
      applyTheme(this.vars)
    }
  }
})