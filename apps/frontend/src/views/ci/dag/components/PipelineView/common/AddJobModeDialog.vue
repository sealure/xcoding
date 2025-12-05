<template>
  <el-dialog v-model="visible" title="添加 Job" width="360px">
    <div class="content">
      <p class="desc">选择添加方式：</p>
      <div class="btns">
        <el-button type="primary" @click="choose('serial')">串行（依赖上一个）</el-button>
        <el-button type="success" @click="choose('parallel')">并行（无依赖）</el-button>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible=false">取消</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'choose'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const choose = (mode) => {
  emit('choose', mode)
  visible.value = false
}
</script>

<style scoped>
.content { padding: 8px 0; }
.desc { margin: 0 0 12px 0; color: #666; }
.btns { display: flex; gap: 8px; }
</style>