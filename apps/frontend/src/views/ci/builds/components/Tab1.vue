<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="build-detail-container compact-top">
        <el-card shadow="hover" class="build-detail-card">
          <template #header>
            <div class="card-header">
              <div class="header-info">
                <div class="header-title">持续集成 · 运行详情</div>
                <div class="header-sub">ID: {{ buildId }}</div>
              </div>
              <div class="actions">
                <el-button type="primary" size="small" @click="refresh">刷新</el-button>
                <el-button type="danger" size="small" :disabled="!canCancel(detail)" @click="cancel"><el-icon><CloseBold /></el-icon>取消</el-button>
                <el-divider direction="vertical" />
                <el-button type="text" @click="goList">返回构建列表</el-button>
                <el-button type="text" v-if="detail?.pipeline_id" @click="goPipeline(detail?.pipeline_id)">返回流水线</el-button>
              </div>
            </div>
          </template>

          <div v-if="loading" class="loading"><el-skeleton :rows="3" animated /></div>
          <div v-else class="detail-body">
            <el-descriptions title="基本信息" :column="3" border>
              <el-descriptions-item label="状态">{{ formatStatus(detail?.status) }}</el-descriptions-item>
              <el-descriptions-item label="流水线ID">{{ detail?.pipeline_id || '—' }}</el-descriptions-item>
              <el-descriptions-item label="触发者">{{ detail?.triggered_by || '—' }}</el-descriptions-item>
              <el-descriptions-item label="分支">{{ detail?.branch || '—' }}</el-descriptions-item>
              <el-descriptions-item label="提交">{{ detail?.commit_sha || '—' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatDate(detail?.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="开始时间">{{ formatDate(detail?.started_at) }}</el-descriptions-item>
              <el-descriptions-item label="结束时间">{{ formatDate(detail?.finished_at) }}</el-descriptions-item>
            </el-descriptions>

            <el-divider content-position="left">K8s 状态</el-divider>
            <div class="k8s-status">
              <div class="k8s-actions">
                <el-switch v-model="k8sOnlyAbnormal" active-text="只看异常" inactive-text="显示全部" />
                <el-button size="small" @click="fetchK8sStatus">刷新</el-button>
              </div>
              <el-table :data="k8sJobs" border style="width: 100%">
                <el-table-column prop="job_name" label="Job" width="260" />
                <el-table-column label="条件" width="320">
                  <template #default="{ row }">
                    <div v-for="c in row.conditions" :key="c.type + c.reason">
                      <el-tag size="small" :type="k8sCondType(c)">{{ c.type }}</el-tag>
                      <span style="margin-left:6px">{{ c.reason }}</span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="Pods" min-width="320">
                  <template #default="{ row }">
                    <div v-for="p in row.pods" :key="p.name">
                      <el-tag size="small">{{ p.name }}</el-tag>
                      <span style="margin-left:6px">{{ p.phase }}</span>
                      <span style="margin-left:6px">{{ p.node || '—' }}</span>
                      <el-tag v-if="p.reason" size="small" type="danger" style="margin-left:6px">{{ p.reason }}</el-tag>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
              <div class="pagination-container">
                <el-pagination
                  v-model:current-page="k8sPage"
                  v-model:page-size="k8sPageSize"
                  :page-sizes="[10, 20, 50]"
                  :total="k8sTotal"
                  layout="total, sizes, prev, pager, next, jumper"
                  @size-change="handleK8sSizeChange"
                  @current-change="handleK8sPageChange"
                />
              </div>
            </div>

            <el-divider content-position="left">日志</el-divider>
            <div class="logs-container" ref="logsRef">
              <pre class="log-lines">{{ logText }}</pre>
            </div>
            <div class="log-actions">
              <el-switch v-model="autoFollow" active-text="自动跟随" inactive-text="暂停跟随" @change="toggleFollow" />
              <el-button size="small" @click="copyLogs">复制</el-button>
              <el-button size="small" @click="downloadLogs">下载</el-button>
              <el-button size="small" @click="clearLogs">清空</el-button>
              <el-button size="small" @click="loadMore">加载更多</el-button>
            </div>
          </div>
        </el-card>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProjectTabs from '@/components/ProjectTabs.vue'
import { getExecutorBuild, getExecutorBuildLogs, cancelExecutorBuild, getExecutorK8sStatus } from '@/api/ci/builds'

const route = useRoute()
const router = useRouter()
const buildId = route.params.id

// 接收来自父组件的props
const props = defineProps({
  build: {
    type: Object,
    default: () => ({})
  },
  logs: {
    type: Array,
    default: () => []
  },
  dagData: {
    type: Object,
    default: () => null
  }
})

const loading = ref(false)
const detail = ref(null)
const logsRef = ref(null)
const logOffset = ref(0)
const logText = ref('')
const autoFollow = ref(true)
let followTimer = null

const formatStatus = (val) => {
  const map = {
    BUILD_STATUS_PENDING: '待下发',
    BUILD_STATUS_QUEUED: '队列中',
    BUILD_STATUS_RUNNING: '运行中',
    BUILD_STATUS_SUCCEEDED: '成功',
    BUILD_STATUS_FAILED: '失败',
    BUILD_STATUS_CANCELLED: '已取消'
  }
  return map[val] || val || '—'
}
const formatDate = (ts) => { try { return ts ? new Date(ts).toLocaleString('zh-CN') : '—' } catch { return '—' } }
const canCancel = (row) => ['BUILD_STATUS_PENDING','BUILD_STATUS_QUEUED','BUILD_STATUS_RUNNING'].includes(row?.status)

const k8sJobs = ref([])
const k8sPage = ref(1)
const k8sPageSize = ref(10)
const k8sTotal = ref(0)
const k8sOnlyAbnormal = ref(false)
const k8sCondType = (c) => {
  if (!c) return ''
  const t = String(c.type || '').toLowerCase()
  if (t.includes('failed')) return 'danger'
  if (t.includes('complete') || t.includes('succeeded')) return 'success'
  return 'info'
}
const fetchK8sStatus = async () => {
  try {
    const res = await getExecutorK8sStatus(String(buildId), '', k8sPage.value, k8sPageSize.value)
    const all = res?.jobs || res?.data || []
    const abnormal = (all || []).filter(j => {
      const hasUnsched = (j.conditions || []).some(c => String(c.reason).toLowerCase() === 'unschedulable')
      const hasFailed = (j.failed && j.failed > 0) || (j.conditions || []).some(c => String(c.type).toLowerCase().includes('failed'))
      const podsFailed = (j.pods || []).some(p => String(p.phase).toLowerCase() === 'failed')
      return hasUnsched || hasFailed || podsFailed
    })
    const data = k8sOnlyAbnormal.value ? abnormal : all
    k8sJobs.value = data
    const totalItems = (k8sOnlyAbnormal.value ? abnormal.length : (res?.pagination?.total_items || (all.length || 0)))
    k8sTotal.value = totalItems
  } catch (e) { console.error('获取K8s状态失败:', e) }
}
const handleK8sSizeChange = (s) => { k8sPageSize.value = s; fetchK8sStatus() }
const handleK8sPageChange = (p) => { k8sPage.value = p; fetchK8sStatus() }

const refresh = async () => {
  loading.value = true
  try {
    const res = await getExecutorBuild(String(buildId))
    detail.value = res?.build || res?.data || res || null
  } catch (e) {
    console.error('加载详情失败:', e)
    ElMessage.error(e?.message || '加载详情失败')
  } finally { loading.value = false }
}

const loadLogs = async () => {
  try {
    const res = await getExecutorBuildLogs(String(buildId), logOffset.value, 500)
    const lines = res?.lines || []
    if (lines.length) {
      logText.value += (logText.value ? '\n' : '') + lines.join('\n')
      logOffset.value = Number(res?.next_offset || (logOffset.value + lines.length))
    }
    if (autoFollow.value && logsRef.value) logsRef.value.scrollTop = logsRef.value.scrollHeight
  } catch (e) {
    console.error('加载日志失败:', e)
  }
}
const loadMore = async () => { await loadLogs() }

// 移除Tab1中的WebSocket连接，避免与BuildDetail.vue重复
// 数据将通过props从父组件BuildDetail.vue传递

const cancel = async () => {
  try {
    await cancelExecutorBuild(String(buildId))
    ElMessage.success('取消成功')
    // WS will update status
  } catch (e) {
    ElMessage.error(e?.message || '取消失败')
  }
}

const goList = () => {
  const pid = detail.value?.pipeline_id
  if (pid) {
    router.push(`/ci/pipeline/${pid}/builds`)
  } else {
    router.push('/ci/pipelines')
  }
}
const goPipeline = (pid) => { if (pid) { router.push(`/ci/pipeline/${pid}`) } }
const toggleFollow = () => { 
  autoFollow.value = !autoFollow.value
  if (autoFollow.value && logsRef.value) logsRef.value.scrollTop = logsRef.value.scrollHeight
}
const copyLogs = async () => { try { await navigator.clipboard.writeText(logText.value || ''); ElMessage.success('已复制') } catch { ElMessage.error('复制失败') } }
const downloadLogs = () => {
  try {
    const blob = new Blob([logText.value || ''], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `build_${buildId}_logs.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch { ElMessage.error('下载失败') }
}
const clearLogs = () => { logText.value = ''; logOffset.value = 0 }

// 监听来自父组件的数据变化
watch(() => props.build, (newBuild) => {
  if (newBuild) {
    detail.value = newBuild
  }
}, { immediate: true, deep: true })

watch(() => props.logs, (newLogs) => {
  if (newLogs && newLogs.length > 0) {
    // 将日志数组转换为字符串
    logText.value = newLogs.join('\n')
    if (autoFollow.value && logsRef.value) {
      // 使用nextTick确保DOM更新后再滚动
      nextTick(() => {
        logsRef.value.scrollTop = logsRef.value.scrollHeight
      })
    }
  }
}, { immediate: true, deep: true })

onMounted(async () => { 
  // 如果没有从父组件接收到数据，则初始化
  if (!props.build) {
    await refresh()
  }
  if (!props.logs || props.logs.length === 0) {
    await loadLogs()
  }
  if (!props.build) {
    await fetchK8sStatus()
  }
  // WebSocket连接已移除，数据将通过props从父组件传递
})
onBeforeUnmount(() => { 
  // WebSocket清理已移除，由父组件BuildDetail.vue管理
})
</script>

<style scoped>
.project-section-layout { height: 100%; display: flex; flex-direction: column; }
.project-section-main { padding: 0; display: flex; flex-direction: column; height: 100%; }
.build-detail-container { padding: 20px; display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; }
.build-detail-card { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; height: 100%; }
.build-detail-card :deep(.el-card__header) { padding: 8px 12px; }
.build-detail-card :deep(.el-card__body) { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; height: 100%; }
.detail-body { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; gap: 12px; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap: 12px; }
.actions { display:flex; gap:8px; align-items:center; }
.header-info { display:flex; flex-direction:row; align-items:baseline; gap:8px; }
.header-title { font-weight:600; font-size:16px; white-space:nowrap; }
.header-sub { color:#909399; font-size:13px; white-space:nowrap; }
.logs-container { background:#0b0d10; color:#d1d5db; border-radius:6px; padding:12px; height: 360px; overflow:auto; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
.log-lines { white-space: pre-wrap; margin: 0; }
.log-actions { display:flex; justify-content:flex-end; gap:8px; }
</style>
<style scoped>
.compact-top { padding-top: 8px; }
</style>
