<template>
  <div class="boolean-editor">
    <el-form label-width="120px">
      <el-form-item label="默认值">
        <el-switch v-model="local.default" />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({
  modelValue: { type: Object, default: () => ({ default: false }) }
})
const emit = defineEmits(['update:modelValue'])
const local = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})
</script>

<style scoped>
.boolean-editor { width: 100%; }
</style>