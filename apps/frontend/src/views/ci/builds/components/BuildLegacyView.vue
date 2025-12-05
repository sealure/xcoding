<template>
  <div class="legacy-view">
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
    </div>

    <el-divider content-position="left">日志</el-divider>
    <div class="logs-container" ref="logsRef">
      <pre class="log-lines">{{ logText }}</pre>
    </div>
    <div class="log-actions">
      <el-button size="small" @click="copyLogs">复制</el-button>
      <el-button size="small" @click="downloadLogs">下载</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const props = defineProps<{
  build: any
  logs: string[]
}>()

const logText = computed(() => props.logs.join(''))
const logsRef = ref<HTMLElement | null>(null)

// K8s Status Logic
const k8sJobs = ref<any[]>([])
const k8sOnlyAbnormal = ref(false)

const k8sCondType = (c: any) => {
  if (!c) return ''
  const t = String(c.type || '').toLowerCase()
  if (t.includes('failed')) return 'danger'
  if (t.includes('complete') || t.includes('succeeded')) return 'success'
  return 'info'
}

const fetchK8sStatus = async () => {
  try {
    // Re-implement fetching K8s status or use the data from WebSocket if available
    // Since the old view fetched it separately, we'll try to fetch it here
    const res = await axios.get(`/ci_service/api/v1/executor/builds/${props.build.id}/k8s_status`)
    const all = res.data.jobs || res.data.data || []
    
    const abnormal = (all || []).filter((j: any) => {
      const hasUnsched = (j.conditions || []).some((c: any) => String(c.reason).toLowerCase() === 'unschedulable')
      const hasFailed = (j.failed && j.failed > 0) || (j.conditions || []).some((c: any) => String(c.type).toLowerCase().includes('failed'))
      const podsFailed = (j.pods || []).some((p: any) => String(p.phase).toLowerCase() === 'failed')
      return hasUnsched || hasFailed || podsFailed
    })
    
    k8sJobs.value = k8sOnlyAbnormal.value ? abnormal : all
  } catch (e) {
    console.error('获取K8s状态失败:', e)
  }
}

const copyLogs = async () => {
  try {
    await navigator.clipboard.writeText(logText.value || '')
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

const downloadLogs = () => {
  try {
    const blob = new Blob([logText.value || ''], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `build_${props.build.id}_logs.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch {
    ElMessage.error('下载失败')
  }
}

onMounted(() => {
  fetchK8sStatus()
})
</script>

<style scoped>
.legacy-view {
  padding: 20px;
}
.k8s-status {
  margin-bottom: 20px;
}
.k8s-actions {
  margin-bottom: 10px;
  display: flex;
  gap: 10px;
  align-items: center;
}
.logs-container {
  background: #0b0d10;
  color: #d1d5db;
  border-radius: 6px;
  padding: 12px;
  height: 360px;
  overflow: auto;
  font-family: monospace;
}
.log-lines {
  white-space: pre-wrap;
  margin: 0;
}
.log-actions {
  margin-top: 10px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
