<template>
  <div class="wd-inputs-editor" ref="rootRef">
    <el-divider content-position="left">workflow_dispatch · inputs（可选）</el-divider>

    <!-- 新增输入项 -->
    <el-form label-width="120px" class="add-input-form">
      <el-form-item>
        <template #label>
          <label :for="nameId">输入名称</label>
        </template>
        <el-input v-model="newItem.name" :id="nameId" placeholder="例如：environment / debug / tags" />
      </el-form-item>
      <el-form-item>
        <template #label>
          <label :for="typeId">类型</label>
        </template>
        <el-select v-model="newItem.type" :id="typeId" placeholder="选择类型" style="width: 240px">
          <el-option label="choice" value="choice" />
          <el-option label="boolean" value="boolean" />
          <el-option label="string" value="string" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="addInputItem">新增输入项</el-button>
      </el-form-item>
    </el-form>

    <!-- 已有输入项列表 -->
    <div v-if="items.length" class="items-list">
      <el-card v-for="(item, idx) in items" :key="`wd-item-${idx}`" class="item-card">
        <template #header>
          <div class="card-header">
            <div class="card-title">
              <strong>{{ item.name }}</strong>
              <span class="type-tag">{{ item.type }}</span>
            </div>
            <div>
              <el-button type="danger" plain size="small" @click="removeItem(idx)">删除</el-button>
            </div>
          </div>
        </template>

        <!-- 公共字段 -->
        <el-form label-width="120px" class="item-form">
          <el-form-item>
            <template #label>
              <label :for="idFor('desc', idx)">描述</label>
            </template>
            <el-input v-model="item.description" :id="idFor('desc', idx)" placeholder="输入项的用途描述" />
          </el-form-item>
          <el-form-item>
            <template #label>
              <label :for="idFor('required', idx)">必填</label>
            </template>
            <el-switch v-model="item.required" :id="idFor('required', idx)" />
          </el-form-item>
        </el-form>

        <!-- 类型专用字段 -->
        <component
          :is="typeComponent(item.type)"
          v-model="items[idx]"
        />
      </el-card>
    </div>

    <div v-else class="empty-hint">
      <el-alert
        title="未添加任何 inputs；保存时将写入 workflow_dispatch: {}"
        type="info"
        :closable="false"
        show-icon
      />
    </div>
  </div>
</template>

<script setup>
import { reactive, watch, computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { dump as yamlDump, load as yamlLoad } from 'js-yaml'
import InputChoice from './InputChoice.vue'
import InputBoolean from './InputBoolean.vue'
import InputString from './InputString.vue'

// v-model: YAML 文本（仅维护 workflow_dispatch 片段）
// - 若 items 为空，则输出 `{ workflow_dispatch: {} }`
// - 若有 items，则输出 `{ workflow_dispatch: { inputs: { ... } } }`
const props = defineProps({
  modelValue: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])

const localText = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const items = reactive([])
/**
 * items 结构示例：
 * {
 *   name: 'environment',
 *   type: 'choice' | 'boolean' | 'string',
 *   description: '',
 *   required: false,
 *   // choice:
 *   options?: string[], default?: string
 *   // boolean:
 *   default?: boolean
 *   // string:
 *   default?: string
 * }
 */

const newItem = reactive({ name: '', type: '' })
const nameId = 'wd-new-name'
const typeId = 'wd-new-type'
const idFor = (field, idx) => `wd-item-${field}-${idx}`
const typeComponent = (t) => {
  if (t === 'choice') return InputChoice
  if (t === 'boolean') return InputBoolean
  return InputString
}

const sanitizeName = (s) => String(s || '').trim()
const sanitizeStr = (s) => String(s || '').trim()
const sanitizeArr = (arr) => (arr || []).map((x) => String(x || '').trim()).filter((x) => x.length)

const addInputItem = () => {
  const name = sanitizeName(newItem.name)
  const type = sanitizeName(newItem.type)
  if (!name || !type || !['choice', 'boolean', 'string'].includes(type)) return
  // 名称唯一性简单检查
  if (items.some((it) => it.name === name)) return
  if (type === 'choice') {
    items.push({ name, type, description: '', required: false, options: [], default: '' })
  } else if (type === 'boolean') {
    items.push({ name, type, description: '', required: false, default: false })
  } else {
    items.push({ name, type, description: '', required: false, default: '' })
  }
  newItem.name = ''
  newItem.type = ''
}
const removeItem = (idx) => { items.splice(idx, 1) }

let updatingFromForm = false

const updateYamlFromForms = () => {
  try {
    updatingFromForm = true
    // 组装 inputs
    const inputs = {}
    for (const it of items) {
      const name = sanitizeName(it.name)
      if (!name) continue
      const common = {
        description: sanitizeStr(it.description),
        required: !!it.required,
        type: it.type
      }
      if (it.type === 'choice') {
        const opts = sanitizeArr(it.options)
        const def = sanitizeStr(it.default)
        inputs[name] = {
          ...common,
          options: opts,
          ...(def && opts.includes(def) ? { default: def } : {})
        }
      } else if (it.type === 'boolean') {
        const def = it.default === true || it.default === false ? it.default : undefined
        inputs[name] = {
          ...common,
          ...(def !== undefined ? { default: def } : {})
        }
      } else { // string
        const def = sanitizeStr(it.default)
        inputs[name] = {
          ...common,
          ...(def ? { default: def } : {})
        }
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
    const wd = obj?.workflow_dispatch ?? obj
    const inps = wd?.inputs ?? {}
    const next = []
    for (const [name, val] of Object.entries(inps)) {
      const type = sanitizeName(val?.type)
      const base = {
        name,
        type: ['choice', 'boolean', 'string'].includes(type) ? type : 'string',
        description: sanitizeStr(val?.description ?? ''),
        required: !!val?.required
      }
      if (base.type === 'choice') {
        next.push({
          ...base,
          options: Array.isArray(val?.options) ? sanitizeArr(val.options) : [],
          default: sanitizeStr(val?.default ?? '')
        })
      } else if (base.type === 'boolean') {
        const defRaw = val?.default
        next.push({
          ...base,
          default: defRaw === true || defRaw === false ? defRaw : false
        })
      } else {
        next.push({
          ...base,
          default: sanitizeStr(val?.default ?? '')
        })
      }
    }
    items.splice(0, items.length, ...next)
  } catch (_) { /* ignore */ }
}

watch(() => props.modelValue, (t) => parseYamlToForms(t), { immediate: true })
watch(items, updateYamlFromForms, { deep: true })

// 捕获阶段阻断扩展脚本在该区域的 focusin 处理
const rootRef = ref(null)
const focusinHandler = (e) => {
  // 若事件目标在本组件根内，则阻止冒泡，以避免外部 content_script 读取
  if (rootRef.value && rootRef.value.contains(e.target)) {
    e.stopPropagation()
  }
}
onMounted(() => {
  document.addEventListener('focusin', focusinHandler, { capture: true })
})
onBeforeUnmount(() => {
  document.removeEventListener('focusin', focusinHandler, { capture: true })
})
</script>

<style scoped>
.wd-inputs-editor { width: 100%; }
.add-input-form { margin-bottom: 12px; }
.items-list { display: flex; flex-direction: column; gap: 12px; }
.item-card { width: 100%; }
.card-header { display: flex; align-items: center; justify-content: space-between; }
.card-title { display: flex; gap: 8px; align-items: center; }
.type-tag { font-size: 12px; color: #888; }
.item-form :deep(.el-form-item__content) { flex: 1; min-width: 0; }
.empty-hint { margin-top: 8px; }
</style>