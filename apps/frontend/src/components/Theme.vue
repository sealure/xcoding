<template>
  <el-dialog v-model="visible" title="主题配置" width="560px">
    <div class="section">
      <div class="row">
        <el-color-picker v-model="form.colorPrimary" show-alpha />
        <span>主色</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.sidebarBg" show-alpha />
        <span>侧边栏背景</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.sidebarText" show-alpha />
        <span>侧边栏文字</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.headerBg" show-alpha />
        <span>头部背景</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.headerText" show-alpha />
        <span>头部文字</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.mainBg" show-alpha />
        <span>主内容背景</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.colorSuccess" show-alpha />
        <span>成功色</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.colorWarning" show-alpha />
        <span>警告色</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.colorDanger" show-alpha />
        <span>危险/错误色</span>
      </div>
      <div class="row">
        <el-color-picker v-model="form.colorInfo" show-alpha />
        <span>信息色</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="onReset">恢复默认</el-button>
      <el-button type="primary" @click="onSave">保存并应用</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, watch, ref, defineExpose } from 'vue'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
const visible = ref(false)

const form = reactive({
  colorPrimary: theme.vars.colorPrimary,
  sidebarBg: theme.vars.sidebarBg,
  sidebarBorder: theme.vars.sidebarBorder,
  sidebarText: theme.vars.sidebarText,
  sidebarActiveText: theme.vars.sidebarActiveText,
  headerBg: theme.vars.headerBg,
  headerText: theme.vars.headerText,
  headerShadow: theme.vars.headerShadow,
  mainBg: theme.vars.mainBg,
  colorSuccess: theme.vars.colorSuccess,
  colorWarning: theme.vars.colorWarning,
  colorDanger: theme.vars.colorDanger,
  colorInfo: theme.vars.colorInfo
})

watch(() => visible.value, (v) => {
  if (v) {
    Object.assign(form, theme.vars)
  }
})

function open() { visible.value = true }
function close() { visible.value = false }
function onSave() { theme.set(form); close() }
function onReset() { theme.reset(); close() }

defineExpose({ open, close })
</script>

<style scoped>
.section { display: grid; grid-template-columns: 1fr; gap: 12px; }
.row { display: flex; align-items: center; gap: 12px; }
.row span { min-width: 100px; color: var(--header-text); }
</style>