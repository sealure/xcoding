<template>
  <el-form label-width="120px" class="basic-info-form">
    <el-form-item label="name">
      <el-input v-model="localName" placeholder="工作流名称（可选）" clearable />
    </el-form-item>
    <el-form-item label="run-name">
      <el-input v-model="localRunName" :placeholder="runNamePlaceholder" clearable />
    </el-form-item>
  </el-form>
  
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({
  name: { type: String, default: '' },
  runName: { type: String, default: '' },
  runNamePlaceholder: { type: String, default: '' }
})
const emit = defineEmits(['update:name', 'update:runName'])
const localName = computed({
  get: () => props.name,
  set: (v) => emit('update:name', v)
})
const localRunName = computed({
  get: () => props.runName,
  set: (v) => emit('update:runName', v)
})
</script>

<style scoped>
.basic-info-form { width: 100%; }
.basic-info-form :deep(.el-form-item__content) { flex: 1; min-width: 0; }
.basic-info-form :deep(.el-input) { width: 100%; }
</style>