<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="build-list-container">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="title-area">
                <span>持续集成 · 运行列表</span>
                <el-tag v-if="projectStore.selectedProject" type="info" size="small">{{
                  projectStore.selectedProject.name }}</el-tag>
              </div>
              <div class="actions">
                <el-input v-model="searchForm.buildName" placeholder="搜索构建名称" style="width: 200px" clearable
                  @input="handleSearch">
                  <template #prefix><el-icon>
                      <Search />
                    </el-icon></template>
                </el-input>


                <el-select v-model="searchForm.pipelineId" placeholder="搜索流水线" style="width: 240px" clearable filterable
                  @change="handleSearch" class="with-prefix">
                  <el-option v-for="p in pipelines" :key="p.id" :label="p.name" :value="p.id" />
                  <template #prefix><el-icon>
                      <Search />
                    </el-icon></template>
                </el-select>

                <el-select v-model="selectedStatus" placeholder="状态" style="width: 160px" clearable>
                  <el-option label="全部" value="" />
                  <el-option label="待下发" value="BUILD_STATUS_PENDING" />
                  <el-option label="队列中" value="BUILD_STATUS_QUEUED" />
                  <el-option label="运行中" value="BUILD_STATUS_RUNNING" />
                  <el-option label="成功" value="BUILD_STATUS_SUCCEEDED" />
                  <el-option label="失败" value="BUILD_STATUS_FAILED" />
                  <el-option label="已取消" value="BUILD_STATUS_CANCELLED" />
                </el-select>
                <el-button @click="resetSearch"><el-icon>
                    <Refresh />
                  </el-icon>重置</el-button>
                <el-switch v-model="autoRefresh" active-text="自动刷新" inactive-text="手动" @change="toggleAutoRefresh" />
                <el-button type="danger" :disabled="selectedRows.length === 0" @click="batchCancel">批量取消</el-button>
                <el-button v-if="usingRoutePipelineId" @click="goPipeline(route.params.id)">返回流水线</el-button>
              </div>
            </div>
          </template>

          <div v-if="!projectStore.selectedProject" class="empty-overview">
            <el-empty description="请选择项目以查看运行列表" />
          </div>

          <div v-else>
            <el-table :data="filteredItems" v-loading="loading" border style="width: 100%"
              @selection-change="onSelectionChange">
              <el-table-column type="selection" width="50" />
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column label="流水线" min-width="150">
                <template #default="{ row }">
                  <el-link type="primary" :underline="false" @click="goPipeline(row.pipeline_id)">
                    {{ pipelineMap[row.pipeline_id] || row.pipeline_id }}
                  </el-link>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="140">
                <template #default="{ row }">
                  <el-tag :type="getStatusType(row.status)" effect="light" round>
                    {{ formatStatus(null, null, row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="triggered_by" label="触发者" width="140" />
              <el-table-column prop="branch" label="分支" min-width="150" />
              <el-table-column prop="commit_sha" label="提交" min-width="120" show-overflow-tooltip />
              <el-table-column label="时间" width="240">
                <template #default="{ row }">
                  <div>创建：{{ formatDate(row.created_at) }}</div>
                  <div>开始：{{ formatDate(row.started_at) }}</div>
                  <div>结束：{{ formatDate(row.finished_at) }}</div>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="goDetail(row)"><el-icon>
                      <View />
                    </el-icon>详情</el-button>
                  <el-button type="danger" link size="small" :disabled="!canCancel(row)"
                    @click="handleCancel(row)"><el-icon>
                      <CloseBold />
                    </el-icon>取消</el-button>
                </template>
              </el-table-column>
            </el-table>

            <div class="pagination-container">
              <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
                :page-sizes="[10, 20, 50, 100]" :total="pagination.total"
                layout="total, sizes, prev, pager, next, jumper" @size-change="handleSizeChange"
                @current-change="handleCurrentChange" />
            </div>
          </div>
        </el-card>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProjectTabs from '@/components/ProjectTabs.vue'
import { useProjectStore } from '@/stores/project'
import { listExecutorBuilds, cancelExecutorBuild } from '@/api/ci/builds'
import { listPipelines } from '@/api/ci/pipeline'

const router = useRouter()
const route = useRoute()
const projectStore = useProjectStore()

const loading = ref(false)
const items = ref([])
const pipelines = ref([])
const pipelineMap = computed(() => {
  const map = {}
  pipelines.value.forEach(p => { map[p.id] = p.name })
  return map
})
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ pipelineId: '', buildName: '' })
const selectedStatus = ref('')
const autoRefresh = ref(true)
let refreshTimer = null
const selectedRows = ref([])
const usingRoutePipelineId = computed(() => {
  const id = route.params.id
  return !!id && String(id).length > 0
})
const filteredItems = computed(() => {
  const s = selectedStatus.value
  const buildName = searchForm.buildName?.toLowerCase() || ''
  const pipelineIdSel = String(searchForm.pipelineId || '')
  return items.value.filter(item => {
    const statusMatch = !s || String(item?.status) === s
    const nameMatch = !buildName || (item.name && item.name.toLowerCase().includes(buildName))
    const pipelineMatch = !pipelineIdSel || String(item?.pipeline_id) === pipelineIdSel
    return statusMatch && nameMatch && pipelineMatch
  })
})

const formatStatus = (_row, _col, val) => {
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
const getStatusType = (val) => {
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
const formatDate = (ts) => {
  try { return ts ? new Date(ts).toLocaleString('zh-CN') : '—' } catch { return '—' }
}
const canCancel = (row) => {
  return ['BUILD_STATUS_PENDING', 'BUILD_STATUS_QUEUED', 'BUILD_STATUS_RUNNING'].includes(row?.status)
}

const fetchList = async () => {
  loading.value = true
  try {
    const pid = projectStore.selectedProject?.id
    if (!pid) { items.value = []; pagination.total = 0; return }
    const pipelineId = (usingRoutePipelineId.value ? String(route.params.id) : (searchForm.pipelineId || ''))
    if (!pipelineId) { items.value = []; pagination.total = 0; return }
    const res = await listExecutorBuilds(String(pipelineId), { page: pagination.page, page_size: pagination.pageSize })
    const data = res.data || res.items || []
    items.value = data
    pagination.total = res.pagination?.total_items || data.length || 0
  } catch (e) {
    console.error('获取运行列表失败:', e)
    ElMessage.error(e?.message || '获取运行列表失败')
  } finally { loading.value = false }
}

const handleSearch = () => {
  pagination.page = 1
  const sid = String(searchForm.pipelineId || '')
  if (usingRoutePipelineId.value && sid && sid !== String(route.params.id)) {
    router.push(`/ci/pipeline/${sid}/builds`)
    return
  }
  fetchList()
}
const resetSearch = () => { searchForm.pipelineId = ''; searchForm.buildName = ''; selectedStatus.value = ''; handleSearch() }
const handleSizeChange = (s) => { pagination.pageSize = s; fetchList() }
const handleCurrentChange = (p) => { pagination.page = p; fetchList() }

const goDetail = (row) => {
  router.push(`/ci/builds/${row.id}`)
}

const goPipeline = (id) => {
  if (id) router.push(`/ci/pipeline/${id}`)
}

const handleCancel = (row) => {
  ElMessageBox.confirm(`确定取消构建 ${row.id}？`, '提示', { type: 'warning' }).then(async () => {
    try { await cancelExecutorBuild(row.id); ElMessage.success('取消成功'); fetchList() } catch (e) { ElMessage.error(e?.message || '取消失败') }
  }).catch(() => { })
}
const onSelectionChange = (rows) => { selectedRows.value = rows || [] }
const batchCancel = () => {
  if (!selectedRows.value.length) return
  ElMessageBox.confirm(`确定取消选中的 ${selectedRows.value.length} 个构建？`, '提示', { type: 'warning' }).then(async () => {
    try {
      for (const r of selectedRows.value) {
        if (['BUILD_STATUS_PENDING', 'BUILD_STATUS_QUEUED', 'BUILD_STATUS_RUNNING'].includes(r?.status)) {
          await cancelExecutorBuild(r.id)
        }
      }
      ElMessage.success('批量取消完成')
      selectedRows.value = []
      fetchList()
    } catch (e) { ElMessage.error(e?.message || '批量取消失败') }
  }).catch(() => { })
}

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    if (!refreshTimer) { refreshTimer = setInterval(() => { fetchList() }, 2000) }
  } else {
    if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null }
  }
}

onMounted(() => {
  try {
    const id = route.params.id
    if (id) searchForm.pipelineId = String(id)
  } catch (_) { }
  // 确保buildName初始化为空字符串
  searchForm.buildName = ''
  // 移除手动 fetchList，完全依赖 watch
  toggleAutoRefresh()
})

const fetchPipelines = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) return
  try {
    const res = await listPipelines(pid, { page: 1, page_size: 100 })
    pipelines.value = res.data || res.pipelines || []
  } catch (e) { console.error(e) }
}

// 监听项目选择变化，自动刷新列表（包括页面刷新后的初始加载）
watch(() => projectStore.selectedProject, (newProject) => {
  if (newProject?.id) {
    fetchList()
    fetchPipelines()
  }
}, { immediate: true })

onBeforeUnmount(() => { if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null } })
</script>

<style scoped>
.project-section-layout {
  height: 100%;
}

.project-section-main {
  padding: 0;
}

.build-list-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.select-with-icon {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.select-with-icon .select-prefix {
  position: absolute;
  left: 8px;
  z-index: 2;
  color: var(--el-text-color-placeholder);
}

.select-with-icon .with-prefix {
  padding-left: 26px;
}

.pagination-container {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.empty-overview {
  padding: 20px 0;
}
</style>
