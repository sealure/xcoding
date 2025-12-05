<template>
  <teleport to="body">
    <div v-show="visible" class="step-action-menu" :style="menuStyle" @mousedown.stop>
      <div class="menu-header">
        <el-input v-model="keyword" placeholder="快速筛选" size="small" clearable />
        <div class="tabs">
          <el-radio-group v-model="sourceTab" size="small">
            <el-radio-button label="全部" />
            <el-radio-button label="官方插件" />
            <!-- <el-radio-button label="团队插件" /> -->
          </el-radio-group>
        </div>
      </div>
      <div class="menu-body">
        <div class="categories">
          <div
            v-for="cat in categories"
            :key="cat.key"
            :class="['cat-item', { active: cat.key === activeCat }]"
            @click="activeCat = cat.key"
          >
            {{ cat.label }}
          </div>
        </div>
        <div class="plugins">
          <div class="plugins-title">{{ activeTitle }}</div>
          <div class="plugin-list">
            <el-scrollbar height="260px">
              <el-empty v-if="filteredPlugins.length === 0" description="无匹配项" />
              <div v-for="p in filteredPlugins" :key="p.key" class="plugin-item">
                <div class="plugin-main">
                  <div class="plugin-name">{{ p.name }}</div>
                  <div class="plugin-desc">{{ p.desc }}</div>
                </div>
                <div class="plugin-actions">
                  <el-button size="small" type="primary" link @click="choose('insert-below', p)">插入到下方</el-button>
                  <el-button size="small" link @click="choose('insert-above', p)">插入到上方</el-button>
                </div>
              </div>
            </el-scrollbar>
          </div>
          <div class="more-actions">
            <el-divider />
            <el-button size="small" type="primary" link @click="emitEdit">编辑当前步骤</el-button>
            <el-button size="small" type="danger" link @click="emitDelete">删除当前步骤</el-button>
          </div>
        </div>
      </div>
      <div class="menu-footer">
        <el-button size="small" link @click="onCancel">关闭</el-button>
      </div>
    </div>
  </teleport>
  </template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  x: { type: Number, default: 0 },
  y: { type: Number, default: 0 },
  jobId: { type: String, default: '' },
  stepIndex: { type: Number, default: -1 },
  step: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['update:modelValue', 'insert', 'edit', 'delete', 'cancel'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const menuStyle = computed(() => ({
  position: 'fixed',
  left: `${Math.max(8, Math.min(window.innerWidth - 520, props.x))}px`,
  top: `${Math.max(8, Math.min(window.innerHeight - 420, props.y))}px`,
}))

const sourceTab = ref('全部')
const keyword = ref('')
const activeCat = ref('command')

const categories = [
  { key: 'command', label: '命令' },
  // { key: 'code', label: '代码管理' },
  // { key: 'file', label: '文件操作' },
  // { key: 'artifact', label: '制品库' },
  // { key: 'report', label: '收集报告' },
  // { key: 'flow', label: '流程控制' },
  // { key: 'security', label: '安全' },
  // { key: 'deploy', label: '持续部署' },
  // { key: 'quality', label: '质量管理' },
  // { key: 'build', label: '编译' },
  // { key: 'release', label: '发布部署' },
  // { key: 'notify', label: '消息通知' },
  // { key: 'cloud', label: '腾讯云插件' },
]

const allPlugins = [
  { key: 'shell', name: '执行 Shell 脚本', desc: '运行任意 Shell 命令', cat: 'command', source: '官方' },
  { key: 'echo', name: '打印消息', desc: '输出调试或提示信息', cat: 'command', source: '官方' },
  { key: 'pipeline-script', name: '执行 Pipeline 脚本', desc: '运行内置脚本', cat: 'flow', source: '官方' },
  { key: 'error', name: '错误信号', desc: '故意触发错误，测试处理逻辑', cat: 'flow', source: '官方' },
  { key: 'sleep', name: '睡眠', desc: '等待指定秒数', cat: 'flow', source: '官方' },
  { key: 'upload-artifact', name: '上传制品', desc: '上传构建产物到制品库', cat: 'artifact', source: '官方' },
  { key: 'download-artifact', name: '下载制品', desc: '从制品库拉取产物', cat: 'artifact', source: '官方' },
  { key: 'notify-slack', name: 'Slack 通知', desc: '向 Slack 频道发送消息', cat: 'notify', source: '团队' },
]

const activeTitle = computed(() => {
  const c = categories.find((c) => c.key === activeCat.value)
  return c ? `${c.label}` : '全部'
})

const filteredPlugins = computed(() => {
  const tab = sourceTab.value
  const kw = keyword.value.trim().toLowerCase()
  return allPlugins.filter((p) => {
    const byCat = activeCat.value ? p.cat === activeCat.value : true
    const byTab = tab === '全部' ? true : (tab === '官方插件' ? p.source === '官方' : p.source !== '官方')
    const byKw = kw ? (p.name.toLowerCase().includes(kw) || p.desc.toLowerCase().includes(kw) || p.key.toLowerCase().includes(kw)) : true
    return byCat && byTab && byKw
  })
})

const choose = (position, plugin) => {
  emit('insert', { position, plugin, jobId: props.jobId, stepIndex: props.stepIndex })
  emit('update:modelValue', false)
}
const onCancel = () => { emit('cancel'); emit('update:modelValue', false) }
const emitEdit = () => { emit('edit', { jobId: props.jobId, stepIndex: props.stepIndex, step: props.step }); emit('update:modelValue', false) }
const emitDelete = () => { emit('delete', { jobId: props.jobId, stepIndex: props.stepIndex }); emit('update:modelValue', false) }

const handleDocClick = (e) => {
  if (!visible.value) return
  const target = e.target
  const panel = document.querySelector('.step-action-menu')
  if (panel && !panel.contains(target)) emit('update:modelValue', false)
}

onMounted(() => { document.addEventListener('mousedown', handleDocClick) })
onBeforeUnmount(() => { document.removeEventListener('mousedown', handleDocClick) })
</script>

<style scoped>
.step-action-menu {
  width: 520px;
  padding: 12px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  box-shadow: var(--el-box-shadow-light);
  z-index: 9999;
}
.menu-header { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; flex-wrap: nowrap; }
.menu-header :deep(.el-input) { flex: 1; }
.menu-header .tabs { flex: none; display: flex; align-items: center; }
.menu-header .tabs :deep(.el-radio-group) { display: inline-flex; flex-direction: row; flex-wrap: nowrap; }
.menu-header .tabs :deep(.el-radio-button) { margin-right: 4px; }
.menu-body { display: flex; gap: 12px; }
.categories { width: 168px; border-right: 1px solid var(--el-border-color); padding-right: 8px; display: flex; flex-direction: column; gap: 6px; }
.cat-item { padding: 6px 8px; border-radius: 6px; cursor: pointer; font-size: 13px; }
.cat-item.active { background: var(--el-color-primary-light-9); color: var(--el-color-primary); }
.plugins { flex: 1; display: flex; flex-direction: column; }
.plugins-title { font-size: 13px; color: var(--el-text-color-secondary); margin: 2px 0 6px 0; }
.plugin-list { display: flex; flex-direction: column; gap: 6px; }
.plugin-item { display: flex; justify-content: space-between; align-items: flex-start; padding: 10px; border: 1px solid var(--el-border-color); border-radius: 6px; margin-bottom: 6px; }
.plugin-name { font-weight: 600; color: var(--el-text-color-primary); }
.plugin-desc { font-size: 12px; color: var(--el-text-color-secondary); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.plugin-actions { display: flex; gap: 8px; flex-shrink: 0; align-items: center; }
.more-actions { margin-top: 6px; display: flex; gap: 8px; align-items: center; }
.menu-footer { margin-top: 8px; display: flex; justify-content: flex-end; }
</style>
