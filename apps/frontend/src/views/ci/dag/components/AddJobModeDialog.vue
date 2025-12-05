<template>
  <el-dialog
    v-model="visible"
    :close-on-click-modal="false"
    width="420px"
    title="选择添加 Job 方式"
  >
    <div class="dialog-body">
      <el-radio-group v-model="mode">
        <el-radio label="serial">串行（依赖上一个 Job，生成 needs）</el-radio>
        <el-radio label="parallel">并行（不依赖上一个 Job）</el-radio>
      </el-radio-group>
      <div class="hint">
        串行将为新 Job 设置 <code>needs</code> 指向前一个 Job；并行则不设置依赖。
      </div>
    </div>
    <template #footer>
      <div class="footer-actions">
        <el-button @click="onCancel">取消</el-button>
        <el-button type="primary" @click="onConfirm">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const mode = ref('serial')

watch(() => props.modelValue, (v) => { if (v) mode.value = 'serial' })

const onCancel = () => { emit('cancel'); emit('update:modelValue', false) }
const onConfirm = () => { emit('confirm', { mode: mode.value }); emit('update:modelValue', false) }
</script>

<style scoped>
.dialog-body { display: flex; flex-direction: column; gap: 12px; }
.hint { font-size: 12px; color: #666; }
.footer-actions { display: flex; justify-content: flex-end; gap: 8px; }
</style>