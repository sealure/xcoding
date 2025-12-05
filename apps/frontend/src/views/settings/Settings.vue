<template>
  <div class="settings-page">
    <el-card>
      <el-tabs v-model="active">
        <el-tab-pane label="个人设置" name="profile">
          <div class="placeholder">个人设置内容（可后续补充）</div>
        </el-tab-pane>
        <el-tab-pane label="主题设置" name="theme">
          <ThemeSettingsPane />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ThemeSettingsPane from './ThemeSettingsPane.vue'

const active = ref('theme')
</script>

<style scoped>
.settings-page { padding: 12px; }
.placeholder { color: #909399; padding: 12px; }
</style>