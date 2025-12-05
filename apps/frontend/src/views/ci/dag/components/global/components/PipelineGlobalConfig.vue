<template>
  <div class="pipeline-edit">
    <!-- 全局配置编辑（已组件化） -->
    <GlobalConfigTabs :doc="doc" ref="tabsRef" @patch="onPatch" />
  </div>
  
</template>

<script setup>
import { ref, defineExpose } from 'vue'
import GlobalConfigTabs from './GlobalConfigTabs.vue'
const props = defineProps({
  doc: { type: Object, required: true }
})
const emit = defineEmits(['patch'])
const tabsRef = ref(null)
const onPatch = (payload) => { try { emit('patch', payload || {}) } catch (_) {} }
defineExpose({
  collectPayload: () => {
    try {
      const t = tabsRef.value
      if (t && typeof t.collectPayload === 'function') return t.collectPayload()
      return null
    } catch (_) {
      return null
    }
  }
})
</script>

<style scoped>
.pipeline-edit { padding-top: 8px; height: 100%; width: 100%; display: flex; flex-direction: column; }
</style>
