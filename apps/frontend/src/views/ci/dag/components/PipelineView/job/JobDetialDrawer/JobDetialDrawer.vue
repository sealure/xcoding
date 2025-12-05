<template>
  <el-drawer v-model="localVisible" title="Job 配置" direction="rtl" size="35%">
    <template #default>
      <div v-if="jobId" class="drawer-content">
        <el-tabs v-model="activeMainTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-form label-width="120px">
              <el-form-item label="Job ID">
                <el-input v-model="localJobId" disabled />
              </el-form-item>
              <el-form-item label="名称 (name)">
                <el-input v-model="localForm.name" placeholder="留空则使用 Job ID" />
              </el-form-item>
              <el-form-item label="timeout-minutes">
                <el-input-number v-model="localForm.timeoutMinutes" :min="0" :step="1" controls-position="right" />
              </el-form-item>
              <el-form-item label="runs-on">
                <el-select
                  v-model="localForm.runsOn"
                  filterable
                  allow-create
                  default-first-option
                  placeholder="选择或输入 GitHub 虚拟机"
                  style="width: 100%"
                >
                  <el-option label="ubuntu-latest" value="ubuntu-latest" />
                  <el-option label="windows-latest" value="windows-latest" />
                  <el-option label="macos-latest" value="macos-latest" />
                </el-select>
              </el-form-item>
              <el-form-item label="container">
                <el-input v-model="localForm.containerImage" placeholder="镜像，例如：ubuntu:latest" clearable />
              </el-form-item>
              <el-divider content-position="left">env（键值对）</el-divider>
              <div>
                <div v-for="(pair, idx) in localForm.envPairs" :key="`env-${idx}`" class="env-pair-row">
                  <el-input v-model="pair.key" placeholder="KEY" style="width: 40%; margin-right: 8px" />
                  <el-input v-model="pair.value" placeholder="VALUE" style="width: 48%; margin-right: 8px" />
                  <el-button type="danger" plain size="small" @click="removeEnvPair(idx)">删除</el-button>
                </div>
                <el-button type="primary" plain size="small" @click="addEnvPair">新增变量</el-button>
              </div>
            </el-form>
          </el-tab-pane>
          
          <el-tab-pane label="高级选项" name="advanced">
            <el-form label-width="120px">
              <el-divider content-position="left">高级选项(可选)</el-divider>
              <el-tabs v-model="activeAdvancedTab">
                <el-tab-pane label="services（YAML 文本）" name="services">
                  <CodeEditor
                    v-model="localForm.servicesText"
                    :rows="10"
                    :fit="true"
                    language="yaml"
                    title="services"
                    theme="dark"
                    placeholder="以 YAML 对象填写 services，例如:\nredis:\n  image: redis:7\n  ports: ['6379']"
                  />
                </el-tab-pane>
                <el-tab-pane label="strategy（YAML 文本）" name="strategy">
                  <CodeEditor
                    v-model="localForm.strategyText"
                    :rows="10"
                    :fit="true"
                    language="yaml"
                    title="strategy"
                    theme="dark"
                    :placeholder="strategyPlaceholder || 'fail-fast: false\nmatrix:\n  os: [ubuntu-latest, macos-latest]\n  node: [18, 20]'"
                  />
                </el-tab-pane>
              </el-tabs>
            </el-form>
          </el-tab-pane>
        </el-tabs>
        <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 12px">
          <el-popconfirm title="确认删除该 Job？" confirm-button-text="删除" cancel-button-text="取消" @confirm="onDelete">
            <template #reference>
              <el-button type="danger" plain>删除 Job</el-button>
            </template>
          </el-popconfirm>
          <div>
            <el-button @click="localVisible = false">取消</el-button>
            <el-button type="primary" @click="onSave">保存</el-button>
          </div>
        </div>
      </div>
      <el-empty v-else description="请选择一个 Job 头部进行配置" />
    </template>
  </el-drawer>
</template>

<script setup>
import { reactive, computed, watch, ref } from 'vue'
import { load as yamlLoad, dump as yamlDump } from 'js-yaml'
import CodeEditor from '../../common/CodeEditor.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  jobId: { type: String, default: '' },
  doc: { type: Object, default: () => ({}) },
  strategyPlaceholder: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue', 'save', 'delete'])

const localVisible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})
const localJobId = computed({
  get: () => props.jobId,
  set: () => {}
})
const localForm = reactive({ name: '', timeoutMinutes: 30, runsOn: 'ubuntu-latest', containerImage: 'ubuntu:latest', envPairs: [], servicesText: '', strategyText: '' })
const activeAdvancedTab = ref('services')
const activeMainTab = ref('basic')

const envObjToPairs = (envObj) => {
  try {
    if (!envObj || typeof envObj !== 'object') return []
    return Object.keys(envObj).map((k) => ({ key: k, value: String(envObj[k] ?? '') }))
  } catch (_) {
    return []
  }
}
const envPairsToObj = (pairs) => {
  const obj = {}
  try {
    (pairs || []).forEach((p) => {
      const k = String(p?.key || '').trim()
      if (k) obj[k] = p?.value ?? ''
    })
    return Object.keys(obj).length ? obj : null
  } catch (_) {
    return null
  }
}

// hydrate local form from doc and jobId
const syncFormFromDoc = () => {
  const jid = props.jobId
  const job = props.doc?.jobs?.[jid] || {}
  localForm.name = job?.name || ''
  localForm.timeoutMinutes = (job?.['timeout-minutes'] ?? 30)
  // runs-on：默认 ubuntu-latest
  localForm.runsOn = job?.['runs-on'] ?? 'ubuntu-latest'
  // container：字符串镜像表单（若为对象则不在此表单展示）；默认 ubuntu:latest
  localForm.containerImage = typeof job?.container === 'string' ? (job.container || 'ubuntu:latest') : 'ubuntu:latest'
  localForm.envPairs = envObjToPairs(job?.env)
  try { localForm.servicesText = job?.services ? yamlDump(job.services) : '' } catch (_) { localForm.servicesText = '' }
  try { localForm.strategyText = job?.strategy ? yamlDump(job.strategy) : '' } catch (_) { localForm.strategyText = '' }
}

watch(() => [props.doc, props.jobId], () => syncFormFromDoc(), { immediate: true, deep: true })

const addEnvPair = () => { localForm.envPairs.push({ key: '', value: '' }) }
const removeEnvPair = (idx) => { localForm.envPairs.splice(idx, 1) }

const onSave = () => {
  try {
    const jid = props.jobId
    const next = { ...(props.doc || {}) }
    if (!next.jobs) next.jobs = {}
    const job = next.jobs[jid] || {}
    const nameVal = String(localForm.name || '').trim()
    if (nameVal) job.name = nameVal; else delete job.name
    const tm = localForm.timeoutMinutes
    if (tm === 0 || (tm && !Number.isNaN(Number(tm)))) job['timeout-minutes'] = Number(tm); else delete job['timeout-minutes']
    // runs-on：始终写入，默认 ubuntu-latest
    {
      const ro = String(localForm.runsOn || '').trim()
      job['runs-on'] = ro || 'ubuntu-latest'
    }
    // container（字符串镜像）：仅当填写时写入，不填写则保留原值（若存在对象配置）
    {
      const ci = String(localForm.containerImage || '').trim()
      job.container = ci || 'ubuntu:latest'
    }
    const envObj = envPairsToObj(localForm.envPairs)
    if (envObj) job.env = envObj; else delete job.env
    const servicesText = String(localForm.servicesText || '').trim()
    if (servicesText) {
      try {
        const svcObj = yamlLoad(servicesText)
        if (!svcObj || typeof svcObj !== 'object' || Array.isArray(svcObj)) { throw new Error('services 必须是 YAML 对象') }
        job.services = svcObj
      } catch {
        throw new Error('services 不是有效的 YAML')
      }
    } else {
      delete job.services
    }
    const st = String(localForm.strategyText || '').trim()
    if (st) {
      try {
        const stratObj = yamlLoad(st)
        if (!stratObj || typeof stratObj !== 'object' || Array.isArray(stratObj)) { throw new Error('strategy 必须是 YAML 对象') }
        job.strategy = stratObj
      } catch {
        throw new Error('strategy 不是有效的 YAML')
      }
    } else {
      delete job.strategy
    }
    next.jobs[jid] = job
    emit('save', next)
    localVisible.value = false
  } catch (e) {
    // Rely on parent to show message
    console.warn('保存 Job 配置失败', e)
  }
}

// 删除当前 Job：移除 doc.jobs[jobId]，并清理其他 Job 的 needs 引用
const onDelete = () => {
  try {
    const jid = props.jobId
    const base = props.doc || {}
    const next = { ...base }
    const jobs = { ...(next.jobs || {}) }
    // 删除目标 Job
    delete jobs[jid]
    // 清理所有引用该 Job 的 needs
    Object.keys(jobs).forEach((k) => {
      const job = { ...(jobs[k] || {}) }
      const raw = job.needs
      const list = Array.isArray(raw) ? raw : (raw ? [raw] : [])
      const filtered = list.filter((n) => n !== jid)
      if (filtered.length === 0) {
        delete job.needs
      } else if (filtered.length === 1) {
        job.needs = filtered[0]
      } else {
        job.needs = filtered
      }
      jobs[k] = job
    })
    next.jobs = jobs
    emit('delete', next)
    localVisible.value = false
  } catch (e) {
    console.warn('删除 Job 失败', e)
  }
}
</script>

<style scoped>
.env-pair-row { margin-bottom: 8px; display: flex; align-items: center; }

/* 使抽屉主体与 Tabs 成为可伸缩容器，编辑器可自适应填充 */
:deep(.el-drawer__body) { display: flex; flex-direction: column; height: 100%; }
.drawer-content { box-sizing: border-box; display: flex; flex-direction: column; height: 100%; }
.drawer-content :deep(.el-tabs) { flex: 1 1 auto; display: flex; flex-direction: column; min-height: 0; }
.drawer-content :deep(.el-tabs__content) { flex: 1 1 auto; height: 100%; }
.drawer-content :deep(.el-tab-pane) { height: 100%; display: flex; flex-direction: column; }
.drawer-content :deep(.el-form) { display: flex; flex-direction: column; height: 100%; }
</style>