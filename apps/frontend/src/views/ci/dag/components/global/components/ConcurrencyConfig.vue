<template>
  <div class="concurrency-config-root">
    <CodeEditor
      v-model="localText"
      :rows="rows"
      :fit="true"
      language="yaml"
      title="concurrency"
      :placeholder="placeholder"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import CodeEditor from '../../PipelineView/common/CodeEditor.vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  rows: { type: Number, default: 6 }
})
const emit = defineEmits(['update:modelValue'])
const localText = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})
</script>

<style scoped>
.concurrency-config-root { height: 100%; width: 100%; display: flex; flex-direction: column; }
</style>