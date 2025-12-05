<template>
  <div class="code-editor" :class="{ fit }">
    <div class="editor-header" v-if="title || language || showActions">
      <div class="editor-meta">
        <span v-if="title" class="editor-title">{{ title }}</span>
        <span v-if="language" class="editor-lang">{{ (language || '').toUpperCase() }}</span>
      </div>
      <div class="editor-actions" v-if="showActions">
        <el-button size="small" @click="onFormat">格式化</el-button>
      </div>
    </div>
    <div class="editor-container">
      <div ref="editorRef" class="editor-body" :style="fit ? { height: '100%' } : { minHeight: `${rows * 20 + 24}px` }"></div>
      <div v-if="showPlaceholder" class="editor-placeholder">{{ placeholder }}</div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, watch, computed } from 'vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
// Register YAML tokenization
import 'monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution'
// Register Shell & PowerShell tokenization for run scripts
import 'monaco-editor/esm/vs/basic-languages/shell/shell.contribution'
import 'monaco-editor/esm/vs/basic-languages/powershell/powershell.contribution'
// Worker setup for Vite
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
// YAML formatting via js-yaml
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  rows: { type: Number, default: 10 },
  language: { type: String, default: 'yaml' },
  title: { type: String, default: '' },
  showActions: { type: Boolean, default: true },
  theme: { type: String, default: 'light' }, // 'light' | 'dark'
  fit: { type: Boolean, default: false }, // 自适应填充父容器高度
  readOnly: { type: Boolean, default: false }, // 只读显示
})
const emit = defineEmits(['update:modelValue'])

const editorRef = ref(null)
let editor = null
let model = null
let updatingFromOutside = false

// Monaco worker environment for Vite
if (typeof self !== 'undefined' && !self.MonacoEnvironment) {
  self.MonacoEnvironment = {
    getWorker: function () {
      return new EditorWorker()
    }
  }
}

const monacoTheme = computed(() => (props.theme || '').toLowerCase() === 'dark' ? 'vs-dark' : 'vs')
const monacoLanguage = computed(() => {
  const raw = (props.language || 'yaml').toLowerCase()
  if (raw === 'yml') return 'yaml'
  if (raw === 'bash') return 'shell'
  return raw
})
const showPlaceholder = computed(() => !((props.modelValue || '').length))

onMounted(() => {
  try {
    model = monaco.editor.createModel(props.modelValue || '', monacoLanguage.value)
    editor = monaco.editor.create(editorRef.value, {
      model,
      automaticLayout: true,
      theme: monacoTheme.value,
      fontSize: 13,
      lineNumbers: 'on',
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      renderWhitespace: 'selection',
      readOnly: !!props.readOnly,
    })
    editor.onDidChangeModelContent(() => {
      if (updatingFromOutside) return
      try { emit('update:modelValue', editor.getValue()) } catch (_) {}
    })

    // Hotkeys
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      try { emit('update:modelValue', editor.getValue()) } catch (_) {}
    })
    editor.addCommand(monaco.KeyMod.Shift | monaco.KeyMod.Alt | monaco.KeyCode.KeyF, () => {
      onFormat()
    })
  } catch (e) {
    console.warn('Monaco init failed', e)
  }
})

onBeforeUnmount(() => {
  try { if (editor) editor.dispose() } catch (_) {}
  try { if (model) model.dispose() } catch (_) {}
  editor = null; model = null
})

watch(() => props.modelValue, (val) => {
  if (!editor || typeof val !== 'string') return
  const current = editor.getValue()
  if (val !== current) {
    try {
      updatingFromOutside = true
      const viewState = editor.saveViewState?.()
      const selection = editor.getSelection?.()
      editor.setValue(val || '')
      // 恢复视图与光标位置，避免跳到首行
      try { if (viewState) editor.restoreViewState?.(viewState) } catch (_) {}
      try { if (selection) editor.setSelection?.(selection) } catch (_) {}
    } finally {
      updatingFromOutside = false
    }
  }
})

watch(monacoTheme, (t) => {
  try { monaco.editor.setTheme(t) } catch (_) {}
})

watch(monacoLanguage, (lang) => {
  try { if (model) monaco.editor.setModelLanguage(model, lang) } catch (_) {}
})

const onFormat = () => {
  try {
    const lang = (props.language || '').toLowerCase()
    const text = editor ? editor.getValue() : (props.modelValue || '')
    if (lang === 'yaml' || lang === 'yml') {
      const obj = yamlLoad(text)
      if (!obj || typeof obj !== 'object') throw new Error('YAML 需为对象')
      const pretty = yamlDump(obj)
      updatingFromOutside = true
      if (editor) editor.setValue(pretty)
      updatingFromOutside = false
      emit('update:modelValue', pretty)
    }
  } catch (e) {
    console.warn('format failed', e)
  }
}
</script>

<style scoped>
.code-editor { border: 1px solid #ebeef5; border-radius: 6px; background: #fff; width: 100%; box-sizing: border-box; flex: 1 1 auto; display: block; }
.editor-header { display: flex; justify-content: space-between; align-items: center; padding: 6px 8px; border-bottom: 1px solid #f2f2f2; }
.editor-meta { display: flex; gap: 8px; align-items: center; }
.editor-title { font-size: 12px; color: #606266; }
.editor-lang { font-size: 12px; color: #909399; }
.editor-actions { display: flex; gap: 6px; }
.editor-container { position: relative; width: 100%; box-sizing: border-box; }
.editor-body { min-height: 200px; width: 100%; }
.editor-placeholder { position: absolute; top: 8px; left: 12px; color: #c0c4cc; font-size: 13px; pointer-events: none; }

/* fit 模式：组件与编辑器填充父容器剩余高度 */
.code-editor.fit { display: flex; flex-direction: column; height: 100%; }
.code-editor.fit .editor-container { flex: 1 1 auto; height: 100%; }
.code-editor.fit .editor-body { height: 100%; min-height: 0; }
</style>