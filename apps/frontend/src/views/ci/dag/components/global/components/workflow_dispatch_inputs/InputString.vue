<template>
  <div class="string-editor">
    <el-form label-width="120px">
      <el-form-item label="默认值">
        <el-input v-model="local.default" placeholder="例如：feature-x" />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({
  modelValue: { type: Object, default: () => ({ default: '' }) }
})
const emit = defineEmits(['update:modelValue'])
const local = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})
</script>

<style scoped>
.string-editor { width: 100%; }
</style>