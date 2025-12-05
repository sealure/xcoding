<template>
  <teleport to="body">
    <div v-show="visible" class="add-job-menu" :style="menuStyle" @mousedown.stop>
      <div class="title">添加 Job 方式</div>
      <div class="actions">
        <el-button size="small" type="primary" @click="onChoose('serial')">串行</el-button>
        <el-button size="small" type="success" @click="onChoose('parallel')">并行</el-button>
      </div>
      <el-button class="close-btn" size="small" text @click="onCancel">关闭</el-button>
    </div>
  </teleport>
</template>

<script setup>
import { computed, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  x: { type: Number, default: 0 },
  y: { type: Number, default: 0 },
})
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const menuStyle = computed(() => ({
  position: 'fixed',
  left: `${Math.max(8, Math.min(window.innerWidth - 216, props.x))}px`,
  top: `${Math.max(8, Math.min(window.innerHeight - 120, props.y))}px`,
}))

const onChoose = (mode) => {
  emit('confirm', { mode })
  emit('update:modelValue', false)
}
const onCancel = () => {
  emit('cancel')
  emit('update:modelValue', false)
}

const handleDocClick = (e) => {
  if (!visible.value) return
  const target = e.target
  const panel = document.querySelector('.add-job-menu')
  if (panel && !panel.contains(target)) emit('update:modelValue', false)
}

onMounted(() => { document.addEventListener('mousedown', handleDocClick) })
onBeforeUnmount(() => { document.removeEventListener('mousedown', handleDocClick) })
</script>

<style scoped>
.add-job-menu {
  width: 208px;
  padding: 12px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  z-index: 9999;
}
.title {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
}
.actions { display: flex; gap: 8px; }
.close-btn { margin-top: 8px; }
</style>