<template>
  <div class="theme-pane">
    <div class="presets">
      <div class="title">选择主题预设</div>
      <el-radio-group v-model="selectedPreset">
        <el-radio-button label="classic">经典蓝灰</el-radio-button>
        <el-radio-button label="graphite">石墨深色</el-radio-button>
        <el-radio-button label="light">亮白清爽</el-radio-button>
        <el-radio-button label="green">绿意盎然</el-radio-button>
      </el-radio-group>
      <el-button class="apply-btn" type="primary" @click="applyPreset">应用预设</el-button>
    </div>

  <div class="customize">
    <div class="title">自定义颜色</div>
    <div class="grid">
      <label>主色</label><el-color-picker v-model="form.colorPrimary" show-alpha />
      <label>侧边栏背景</label><el-color-picker v-model="form.sidebarBg" show-alpha />
      <label>侧边栏文字</label><el-color-picker v-model="form.sidebarText" show-alpha />
      <label>头部背景</label><el-color-picker v-model="form.headerBg" show-alpha />
      <label>头部文字</label><el-color-picker v-model="form.headerText" show-alpha />
      <label>主内容背景</label><el-color-picker v-model="form.mainBg" show-alpha />
      <label>成功色</label><el-color-picker v-model="form.colorSuccess" show-alpha />
      <label>警告色</label><el-color-picker v-model="form.colorWarning" show-alpha />
      <label>危险色</label><el-color-picker v-model="form.colorDanger" show-alpha />
      <label>信息色</label><el-color-picker v-model="form.colorInfo" show-alpha />
    </div>
    <div class="actions">
      <el-button @click="onReset">恢复默认</el-button>
      <el-button type="primary" @click="onSave">保存并应用</el-button>
    </div>
  </div>

  <div class="preview">
    <div class="title">状态色预览</div>
    <div class="preview-section">
      <div class="section-title">Buttons</div>
      <div class="preview-row">
        <el-button type="primary">Primary</el-button>
        <el-button type="success">Success</el-button>
        <el-button type="warning">Warning</el-button>
        <el-button type="danger">Danger</el-button>
        <el-button type="info">Info</el-button>
      </div>
    </div>
    <div class="preview-section">
      <div class="section-title">Tags</div>
      <div class="preview-row">
        <el-tag type="success">成功</el-tag>
        <el-tag type="warning">警告</el-tag>
        <el-tag type="danger">危险</el-tag>
        <el-tag type="info">信息</el-tag>
      </div>
    </div>
    <div class="preview-section">
      <div class="section-title">Alerts</div>
      <div class="preview-grid">
        <el-alert title="操作成功" type="success" :closable="false" effect="light" />
        <el-alert title="请注意风险" type="warning" :closable="false" effect="light" />
        <el-alert title="发生错误" type="error" :closable="false" effect="light" />
        <el-alert title="提示信息" type="info" :closable="false" effect="light" />
      </div>
    </div>
  </div>
</div>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()

const selectedPreset = ref('classic')

const PRESETS = {
  classic: {
    // 现代商务：深邃蓝灰侧边栏 + Inter Blue
    colorPrimary: '#2563eb',
    sidebarBg: '#1e293b',
    sidebarBorder: '#334155',
    sidebarText: '#94a3b8',
    sidebarActiveText: '#ffffff',
    headerBg: '#ffffff',
    headerText: '#1e293b',
    headerShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    mainBg: '#f1f5f9',
    colorSuccess: '#10b981',
    colorWarning: '#f59e0b',
    colorDanger: '#ef4444',
    colorInfo: '#64748b'
  },
  graphite: {
    // 极客暗黑：全深色界面 + Indigo 提亮
    colorPrimary: '#6366f1',
    sidebarBg: '#0f172a',
    sidebarBorder: '#1e293b',
    sidebarText: '#94a3b8',
    sidebarActiveText: '#818cf8',
    headerBg: '#0f172a',
    headerText: '#e2e8f0',
    headerShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
    mainBg: '#020617',
    colorSuccess: '#22c55e',
    colorWarning: '#f59e0b',
    colorDanger: '#ef4444',
    colorInfo: '#94a3b8'
  },
  light: {
    // 亮白清爽：默认的极简白风格
    colorPrimary: '#2563eb',
    sidebarBg: '#ffffff',
    sidebarBorder: '#e5e7eb',
    sidebarText: '#4b5563',
    sidebarActiveText: '#2563eb',
    headerBg: '#ffffff',
    headerText: '#1f2937',
    headerShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    mainBg: '#f9fafb',
    colorSuccess: '#10b981',
    colorWarning: '#f59e0b',
    colorDanger: '#ef4444',
    colorInfo: '#6b7280'
  },
  green: {
    // 自然清新：浅色基调 + Emerald 强调
    colorPrimary: '#059669',
    sidebarBg: '#ffffff',
    sidebarBorder: '#e5e7eb',
    sidebarText: '#065f46',
    sidebarActiveText: '#059669',
    headerBg: '#ffffff',
    headerText: '#064e3b',
    headerShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    mainBg: '#f0fdf4',
    colorSuccess: '#16a34a',
    colorWarning: '#d97706',
    colorDanger: '#dc2626',
    colorInfo: '#0d9488'
  }
}

const form = reactive({ ...theme.vars })

// 当选择预设时，将预设值同步到表单，保证“保存并应用”生效
watch(selectedPreset, (val) => {
  const p = PRESETS[val]
  if (p) Object.assign(form, p)
}, { immediate: true })

function applyPreset() {
  const p = PRESETS[selectedPreset.value]
  Object.assign(form, p)
  theme.set(p)
}
function onSave() { theme.set(form) }
function onReset() { theme.reset() }

// 根据当前主题识别预设（用于进入时自动选中对应预设）
const presetEntries = Object.entries(PRESETS)
for (const [key, val] of presetEntries) {
  if (
    theme.vars.colorPrimary === val.colorPrimary &&
    theme.vars.sidebarBg === val.sidebarBg &&
    theme.vars.headerBg === val.headerBg &&
    theme.vars.mainBg === val.mainBg
  ) {
    selectedPreset.value = key
    break
  }
}
</script>

<style scoped>
.theme-pane { display: grid; gap: 16px; }
.presets { display: grid; gap: 12px; }
.customize { display: grid; gap: 12px; }
.title { font-weight: 600; color: var(--header-text); }
.grid { display: grid; grid-template-columns: 140px 1fr; gap: 8px 16px; align-items: center; }
.actions { display: flex; gap: 8px; justify-content: flex-end; }
.apply-btn { justify-self: start; }

/* 预览区域样式 */
.preview { display: grid; gap: 12px; }
.preview-section { display: grid; gap: 8px; }
.section-title { font-weight: 500; color: var(--header-text); opacity: 0.9; }
.preview-row { display: flex; gap: 8px; flex-wrap: wrap; }
.preview-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
</style>