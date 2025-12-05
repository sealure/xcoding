<template>
<el-container class="project-section-layout">
  <el-main class="project-section-main">
      <ProjectTabs />
      <div class="pipeline-detail-container compact-top">
        <el-card shadow="hover" class="pipeline-detail-card">
          <template #header>
            <div class="card-header">
              <div class="header-info">
                <div class="header-title">持续集成 · 流水线详情</div>
              <div class="header-sub">{{ titleText }}</div>
              </div>
              <div class="actions">
                <el-button type="success" size="small" :disabled="!pipelineId" @click="saveAndRun">运行</el-button>
                <el-button type="success" size="small" :disabled="!pipelineId" @click="navRef?.savePipeline?.()">保存流水线</el-button>
                <el-button type="primary" size="small" @click="goBuilds">查看构建</el-button>
                <el-button type="warning" size="small" @click="navRef?.exportYaml?.()">导出YAML</el-button>
                <el-divider direction="vertical" />
                <el-button type="text" @click="goList">返回构建列表</el-button>
              </div>
            </div>
          </template>

          <div v-if="loading" class="loading"><el-skeleton :rows="3" animated /></div>
          <div v-else class="detail-body">
            <div class="nav">
              <!-- 传入后端返回的 workflow_yaml 到编辑器，并传递流水线名称用于同步 -->
              <PipelineNavigation ref="navRef" :serverYaml="serverYamlText" :pipelineId="String(pipelineId)" :pipelineName="pipelineName" />
            </div>
          </div>
        </el-card>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProjectTabs from '@/components/ProjectTabs.vue'
import PipelineNavigation from '@/views/ci/dag/PipelineNavigation.vue'
import { getPipeline } from '@/api/ci/pipeline'
import { startPipelineBuild } from '@/api/ci/pipeline'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref(null)

// 兼容模板表达式：避免在模板中使用可选链，改用计算属性
const titleText = computed(() => {
  const d = detail.value || null
  return (d && d.name) || (d && d.id) || ''
})
const descText = computed(() => {
  const d = detail.value || null
  return (d && d.description) || '—'
})
const serverYamlText = computed(() => {
  const d = detail.value || null
  return (d && d.workflow_yaml) || ''
})

// 供编辑器同步顶层 YAML name
const pipelineName = computed(() => {
  const d = detail.value || null
  return (d && d.name) || ''
})

const pipelineId = route.params.id
const navRef = ref(null)

const fetchDetail = async () => {
  if (!pipelineId) return
  loading.value = true
  try {
    const res = await getPipeline(String(pipelineId))
    // proto: GetPipelineResponse { pipeline: Pipeline }
    detail.value = res?.pipeline || res?.data?.pipeline || res?.data || res || null
  } catch (e) {
    ElMessage.error(`加载详情失败：${e?.message || e}`)
  } finally {
    loading.value = false
  }
}

const goList = () => { router.push('/ci/pipeline') }
const goBuilds = () => { if (pipelineId) router.push(`/ci/pipeline/${pipelineId}/builds`) }

const saveAndRun = async () => {
  try {
    if (!pipelineId) return
    const nav = navRef.value
    if (nav && typeof nav.savePipeline === 'function') {
      await nav.savePipeline()
    }
    const resp = await startPipelineBuild(String(pipelineId), {})
    const build = (resp && (resp.build || (resp.data && (resp.data.build || resp.data)) || resp)) || {}
    const bid = (build && build.id)
    if (bid) { router.push(`/ci/builds/${bid}`) } else { ElMessage.warning('运行已触发，但未返回构建ID') }
  } catch (e) {
    ElMessage.error(`保存并运行失败：${e?.message || e}`)
  }
}

onMounted(() => { fetchDetail() })
</script>

<style scoped>
.project-section-layout { height: 100%; display: flex; flex-direction: column; }
.project-section-main { padding: 0; display: flex; flex-direction: column; height: 100%; }
.pipeline-detail-container { padding: 20px; display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; }
.pipeline-detail-card { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; height: 100%; }
.pipeline-detail-card :deep(.el-card__header) { padding: 8px 12px; }
.pipeline-detail-card :deep(.el-card__body) { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; height: 100%; }
.detail-body { display: flex; flex-direction: column; flex: 1 1 auto; min-height: 0; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap: 12px; }
.actions { display:flex; gap:8px; align-items:center; }
 .header-info { display:flex; flex-direction:row; align-items:baseline; gap:8px; }
 .header-title { font-weight:600; font-size:16px; white-space:nowrap; }
 .header-sub { color:#909399; font-size:13px; white-space:nowrap; }
 /* 取消正文中的标题与描述，缩减容器层级产生的空白 */
.nav { margin-top: 12px; flex: 1 1 auto; display: flex; min-height: 0; }
</style>
<style scoped>
.compact-top { padding-top: 8px; }
</style>