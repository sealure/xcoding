<template>
  <div class="choice-editor">
    <el-form label-width="120px">
      <el-form-item label="选项">
        <div>
          <div v-for="(opt, i) in local.options" :key="`opt-${i}`" class="opt-row">
            <el-input v-model="local.options[i]" placeholder="例如：staging / production" style="width: 70%; margin-right: 8px" />
            <el-button type="danger" plain size="small" @click="removeOption(i)">删除</el-button>
          </div>
          <el-button type="primary" plain size="small" @click="addOption">新增选项</el-button>
        </div>
      </el-form-item>
      <el-form-item label="默认值">
        <el-select v-model="local.default" placeholder="选择默认值" style="width: 240px">
          <el-option v-for="opt in sanitizedOptions" :key="opt" :label="opt" :value="opt" />
        </el-select>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: Object, default: () => ({ options: [], default: '' }) }
})
const emit = defineEmits(['update:modelValue'])

const local = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})
const sanitizedOptions = computed(() => (local.value.options || []).map((x) => String(x || '').trim()).filter((x) => x.length))
const addOption = () => { local.value.options.push('') }
const removeOption = (i) => { local.value.options.splice(i, 1) }
const defaultId = 'wd-choice-default'
const optsId = 'wd-choice-opts'
const idFor = (i) => `wd-choice-opt-${i}`
</script>

<style scoped>
.choice-editor { width: 100%; }
.opt-row { margin-bottom: 8px; display: flex; align-items: center; }
</style>