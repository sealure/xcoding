<template>
  <div class="build-detail-tabs">
    <el-card shadow="never" class="full-card">
      <template #header>
        <div class="card-header">
          <div class="title-area">
            <span>持续集成 · 运行详情</span>
            <el-tag v-if="build?.id" type="info" size="small">ID: {{ build.id }}</el-tag>
            <el-tag v-if="build?.status" :type="statusType(build.status)" effect="light" round>{{ statusText(build.status) }}</el-tag>
          </div>
          <div class="actions">
            <el-button size="small" type="text" @click="goBuildList">返回构建列表</el-button>
            <el-button size="small" type="text" v-if="build?.pipeline_id" @click="goPipeline(build.pipeline_id)">返回流水线</el-button>
          </div>
        </div>
      </template>
      <el-tabs v-model="activeTab" class="clean-tabs fill-tabs">
        <el-tab-pane label="基本信息" name="basic">
          <BuildBasicInfo :detail="build" />
        </el-tab-pane>
        <el-tab-pane label="构建过程" name="process">
          <BuildProcess :snapshot="build.snapshot" :logs="logs" :dag-data="dagData" />
        </el-tab-pane>
        <el-tab-pane label="K8s 状态" name="k8s">
          <BuildK8sStatus :build-id="build?.id || ''" />
        </el-tab-pane>
        <el-tab-pane label="结构化日志" name="structured">
          <StructuredLogs :structured-logs="structuredLogs" :dag-data="dagData" />
        </el-tab-pane>
        <el-tab-pane label="构建快照" name="snapshot">
          <BuildSnapshot :snapshot="build.snapshot" :pipeline-id="build.pipeline_id" :active="activeTab==='snapshot'" />
        </el-tab-pane>
        <el-tab-pane label="构建制品" name="artifacts">
          <BuildArtifacts />
        </el-tab-pane>
        <!-- <el-tab-pane label="旧视图" name="legacy">
          <Tab1 :build="build" :logs="logs" :dag-data="dagData" />
        </el-tab-pane>
         -->
        
      </el-tabs>

    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import BuildProcess from './BuildProcess.vue'
import StructuredLogs from './StructuredLogs.vue'
import BuildSnapshot from './BuildSnapshot.vue'
import BuildArtifacts from './BuildArtifacts.vue'
import Tab1 from './Tab1.vue'
import BuildBasicInfo from './BuildBasicInfo.vue'
import BuildK8sStatus from './BuildK8sStatus.vue'

const props = defineProps({
  build: {
    type: Object,
    default: () => ({})
  },
  logs: {
    type: Array,
    default: () => []
  },
  structuredLogs: {
    type: Array,
    default: () => []
  },
  dagData: {
    type: Object,
    default: () => null
  }
})

onMounted(() => {
 console.log('BuildDetailTabs - build object:', props.build)
  console.log('BuildDetailTabs - build.snapshot:', props.build.snapshot)
  console.log('BuildDetailTabs - logs:', props.logs)
  console.log('BuildDetailTabs - dagData:', props.dagData)
})

const activeTab = ref('process')
const router = useRouter()

const goBuildList = () => {
  try {
    const pid = props.build?.pipeline_id
    if (pid) router.push(`/ci/pipeline/${pid}/builds`)
    else router.push('/ci/builds')
  } catch (_) {}
}
const goPipeline = (pid) => { try { if (pid) router.push(`/ci/pipeline/${pid}`) } catch (_) {} }

const statusText = (val) => {
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
const statusType = (val) => {
  const map = {
    BUILD_STATUS_PENDING: 'info',
    BUILD_STATUS_QUEUED: 'warning',
    BUILD_STATUS_RUNNING: 'primary',
    BUILD_STATUS_SUCCEEDED: 'success',
    BUILD_STATUS_FAILED: 'danger',
    BUILD_STATUS_CANCELLED: 'info'
  }
  return map[val] || 'info'
}
</script>

<style scoped>
.build-detail-tabs { margin-top: 0; display: flex; flex-direction: column; height: 100%; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.title-area { display:flex; align-items:center; gap:8px; }
.actions { display:flex; align-items:center; gap:8px; }
.clean-tabs :deep(.el-tabs__header) { margin: 0; }
.clean-tabs :deep(.el-tabs__nav-wrap::after) { height: 0; }
.clean-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  color: var(--el-text-color-regular);
  height: 44px;
  line-height: 44px;
}
.clean-tabs :deep(.el-tabs__item.is-active) {
  color: var(--el-color-primary);
  font-weight: 600;
}
.full-card { display: flex; flex-direction: column; height: 100%; }
.full-card :deep(.el-card__body) { flex: 1 1 auto; display: flex; flex-direction: column; min-height: 0; }
.fill-tabs { flex: 1 1 auto; display: flex; flex-direction: column; min-width: 0; }
.fill-tabs :deep(.el-tabs__content) { flex: 1 1 auto; display: flex; min-height: 0; }
.fill-tabs :deep(.el-tab-pane) { flex: 1 1 auto; display: flex; min-height: 0; }
</style>
