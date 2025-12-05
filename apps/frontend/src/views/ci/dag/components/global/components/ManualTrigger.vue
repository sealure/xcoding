<template>
  <div class="manual-trigger-root">
    <el-form label-width="120px" class="manual-form">
      <el-divider content-position="left">workflow_dispatch · inputs（可选）</el-divider>

      <el-form-item label="启用 environment">
        <el-switch v-model="enabled.env" />
      </el-form-item>

      <!-- environment: choice -->
      <template v-if="enabled.env">
        <el-form-item label="environment · 描述">
          <el-input v-model="formEnv.description" placeholder="部署环境" />
        </el-form-item>
        <el-form-item label="environment · 必填">
          <el-switch v-model="formEnv.required" />
        </el-form-item>
        <el-form-item label="environment · 选项">
          <div>
            <div v-for="(opt, idx) in formEnv.options" :key="`env-opt-${idx}`" class="list-row">
              <el-input v-model="formEnv.options[idx]" placeholder="例如：staging / production" style="width: 70%; margin-right: 8px" />
              <el-button type="danger" plain size="small" @click="removeOption(idx)">删除</el-button>
            </div>
            <el-button type="primary" plain size="small" @click="addOption">新增选项</el-button>
          </div>
        </el-form-item>
        <el-form-item label="environment · 默认值">
          <el-select v-model="formEnv.default" placeholder="选择默认环境">
            <el-option v-for="opt in sanitizedOptions" :key="opt" :label="opt" :value="opt" />
          </el-select>
        </el-form-item>
      </template>

      <el-form-item label="启用 debug">
        <el-switch v-model="enabled.debug" />
      </el-form-item>
      <!-- debug: boolean -->
      <template v-if="enabled.debug">
        <el-form-item label="debug · 描述">
          <el-input v-model="formDebug.description" placeholder="启用调试模式" />
        </el-form-item>
        <el-form-item label="debug · 必填">
          <el-switch v-model="formDebug.required" />
        </el-form-item>
      </template>

      <el-form-item label="启用 tags">
        <el-switch v-model="enabled.tags" />
      </el-form-item>
      <!-- tags: string -->
      <template v-if="enabled.tags">
        <el-form-item label="tags · 描述">
          <el-input v-model="formTags.description" placeholder="自定义标签" />
        </el-form-item>
        <el-form-item label="tags · 必填">
          <el-switch v-model="formTags.required" />
        </el-form-item>
      </template>
    </el-form>
  </div>
</template>

<script setup>
import { reactive, computed, watch } from 'vue'
import { dump as yamlDump, load as yamlLoad } from 'js-yaml'

// v-model: 维护 workflow_dispatch 的 YAML 片段
const props = defineProps({
  modelValue: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])
const localText = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

// 表单状态
const enabled = reactive({ env: false, debug: false, tags: false })
const formEnv = reactive({ description: '部署环境', required: true, options: ['staging', 'production'], default: 'staging' })
const formDebug = reactive({ description: '启用调试模式', required: false })
const formTags = reactive({ description: '自定义标签', required: false })

const sanitizedOptions = computed(() => (formEnv.options || []).map((x) => String(x || '').trim()).filter((x) => x.length))
const addOption = () => { formEnv.options.push('') }
const removeOption = (idx) => { formEnv.options.splice(idx, 1) }

let updatingFromForm = false

const updateYamlFromForms = () => {
  try {
    updatingFromForm = true
    const inputs = {}
    if (enabled.env) {
      const envOpt = sanitizedOptions.value
      const envDefault = envOpt.includes(formEnv.default) ? formEnv.default : (envOpt[0] || '')
      inputs.environment = {
        description: String(formEnv.description || ''),
        required: !!formEnv.required,
        type: 'choice',
        options: envOpt,
        ...(envDefault ? { default: envDefault } : {})
      }
    }
    if (enabled.debug) {
      inputs.debug = {
        description: String(formDebug.description || ''),
        required: !!formDebug.required,
        type: 'boolean'
      }
    }
    if (enabled.tags) {
      inputs.tags = {
        description: String(formTags.description || ''),
        required: !!formTags.required,
        type: 'string'
      }
    }
    const obj = Object.keys(inputs).length ? { workflow_dispatch: { inputs } } : { workflow_dispatch: {} }
    localText.value = yamlDump(obj, { noRefs: true, lineWidth: 120 })
  } finally {
    updatingFromForm = false
  }
}

const parseYamlToForms = (text) => {
  if (updatingFromForm) return
  try {
    const obj = yamlLoad(String(text || '')) || {}
    const wd = obj?.workflow_dispatch || obj || {}
    const inputs = wd?.inputs || {}
    enabled.env = !!inputs?.environment
    enabled.debug = !!inputs?.debug
    enabled.tags = !!inputs?.tags
    const env = inputs?.environment || {}
    formEnv.description = String(env?.description ?? '部署环境')
    formEnv.required = !!env?.required
    formEnv.options = Array.isArray(env?.options) ? [...env.options] : (env?.options ? [String(env.options)] : ['staging', 'production'])
    formEnv.default = String(env?.default ?? (formEnv.options[0] || 'staging'))
    const dbg = inputs?.debug || {}
    formDebug.description = String(dbg?.description ?? '启用调试模式')
    formDebug.required = !!dbg?.required
    const tgs = inputs?.tags || {}
    formTags.description = String(tgs?.description ?? '自定义标签')
    formTags.required = !!tgs?.required
  } catch (_) { /* ignore parse errors */ }
}

watch(() => props.modelValue, (t) => parseYamlToForms(t), { immediate: true })
watch([enabled, formEnv, formDebug, formTags], updateYamlFromForms, { deep: true })
</script>

<style scoped>
.manual-trigger-root { width: 100%; height: 100%; display: flex; flex-direction: column; }
.manual-form { width: 100%; }
.manual-form :deep(.el-form-item__content) { flex: 1; min-width: 0; }
.list-row { margin-bottom: 8px; display: flex; align-items: center; }
</style>