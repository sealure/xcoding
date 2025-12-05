<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="build-detail-wrapper">
        <BuildDetailTabs v-if="build" :build="build" :logs="logs" :structured-logs="structuredLogs" :dag-data="dagData" />
        <div v-else class="loading-state">
          <el-skeleton :rows="10" animated />
        </div>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProjectTabs from '@/components/ProjectTabs.vue'
import BuildDetailTabs from './components/BuildDetailTabs.vue'
import { getExecutorBuild } from '@/api/ci/builds'

const route = useRoute()
const buildId = route.params.id

const build = ref(null)
const logs = ref([])
const structuredLogs = ref([])
const dagData = ref(null)
let ws = null
let dagReceived = false // 添加标志来跟踪是否已接收DAG数据

// 数据缓存机制
const dataCache = {
  lastStatus: null,
  lastK8sStatus: null,
  lastDAGData: null,
  processedLogs: new Set() // 使用Set来跟踪已处理的日志
}

const fetchBuild = async () => {
  try {
    const res = await getExecutorBuild(String(buildId))
    build.value = res?.build || res?.data || res || null
    dagReceived = false // 重置标志
    initWebSocket()
  } catch (error) {
    console.error('Failed to fetch build:', error)
    ElMessage.error('Failed to load build details')
  }
}

const initWebSocket = () => {
  if (ws) ws.close()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${window.location.host}/ci_service/api/v1/executor/ws/builds/${buildId}?offset=0`
  
  const socket = new WebSocket(url)
  socket.onopen = () => { console.log('WS connected:', url) }
  socket.onerror = (e) => { console.error('WS error:', e) }
  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'build_status' && msg.data) {
        const data = msg.data
        if (data.build) {
          build.value = { ...(build.value || {}), ...(data.build || {}) }
        }
        structuredLogs.value = data
        const jobs = Array.isArray(data.jobs) ? data.jobs : []
        const edges = Array.isArray(data.edges) ? data.edges : []
        const jobsMapped = jobs.map(j => ({ Name: j?.name || '', Status: j?.status || '' }))
        const stepsMapped = []
        jobs.forEach(j => {
          const jn = j?.name || ''
          const steps = Array.isArray(j?.step) ? j.step : []
          steps.forEach(st => {
            stepsMapped.push({ JobName: jn, Name: st?.name || '', Status: st?.status || '' })
            const arr = Array.isArray(st?.logs) ? st.logs : []
            arr.forEach(li => {
              const s = String(li?.content || '').trim()
              if (!s) return
              const text = `[${jn}] [${st?.name || ''}] ${s}`
              if (!dataCache.processedLogs.has(text)) {
                dataCache.processedLogs.add(text)
                logs.value.push(text)
              }
            })
          })
        })
        dagData.value = { jobs: jobsMapped, steps: stepsMapped, edges }
        const st = String(data?.build?.status || '')
        if (st === 'BUILD_STATUS_SUCCEEDED' || st === 'BUILD_STATUS_FAILED' || st === 'BUILD_STATUS_CANCELLED') {
          socket.close()
          ws = null
        }
      }
    } catch (e) {
      console.error('WS message error', e)
    }
  }
  socket.onclose = (e) => { console.log('WS closed:', e.code, e.reason) }
    ws = socket
}

const stopWebSocket = () => {
  if (ws) {
    ws.close()
    ws = null
  }
}

onMounted(async () => { 
  await fetchBuild()
})
onBeforeUnmount(() => { stopWebSocket() })
</script>

<style scoped>
.project-section-layout { height: 100%; display: flex; flex-direction: column; }
.project-section-main { padding: 0; display: flex; flex-direction: column; height: 100%; }
.build-detail-wrapper { padding: 0; flex: 1; display: flex; flex-direction: column; min-height: 0; }
.loading-state { padding: 20px; background: #fff; border-radius: 8px; }
</style>
