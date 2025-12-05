<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="pipelines-container compact-top">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>持续集成 · 流水线列表</span>
              <div class="toolbar">
                <el-input
                  v-model="searchForm.name"
                  placeholder="按流水线名称搜索"
                  clearable
                  class="toolbar-input"
                  @keyup.enter="handleSearch"
                />
                <el-button type="primary" @click="handleSearch" class="toolbar-btn">
                  <el-icon><Search /></el-icon>搜索
                </el-button>
                <el-button @click="resetSearch" class="toolbar-btn">
                  <el-icon><Refresh /></el-icon>重置
                </el-button>
                <el-button type="primary" plain @click="refresh" class="toolbar-btn">
                  <el-icon><Refresh /></el-icon>刷新
                </el-button>
                <el-button type="success" @click="handleAdd" class="toolbar-btn">
                  <el-icon><Plus /></el-icon>新建流水线
                </el-button>
              </div>
            </div>
          </template>

          <div v-if="!projectStore.selectedProject" class="empty">
            <el-empty description="请选择项目以管理流水线" />
          </div>

          <div v-else>
            <el-table :data="items" v-loading="loading" border style="width: 100%">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="name" label="名称" min-width="200" />
              <el-table-column prop="description" label="描述" min-width="300" />
              <el-table-column label="操作" width="330">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="goDetail(row)">
                    <el-icon><Edit /></el-icon>编辑
                  </el-button>
                  <el-button type="success" link size="small" @click="runPipeline(row)">运行</el-button>
                  <el-button type="primary" link size="small" @click="goBuilds(row)">运行列表</el-button>
                  <el-button type="danger" link size="small" @click="handleDelete(row)">
                    <el-icon><Delete /></el-icon>删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div class="pagination-container">
              <el-pagination
                background
                layout="prev, pager, next"
                :total="pagination.total"
                :page-size="pagination.pageSize"
                :current-page="pagination.page"
                @current-change="handlePageChange"
              />
            </div>
          </div>
        </el-card>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProjectStore } from '@/stores/project'
import { listPipelines, createPipeline, deletePipeline } from '@/api/ci/pipeline'
import { startPipelineBuild } from '@/api/ci/pipeline'
import ProjectTabs from '@/components/ProjectTabs.vue'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'

const router = useRouter()
const projectStore = useProjectStore()

const loading = ref(false)
const items = ref([])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '' })

const fetchList = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { items.value = []; pagination.total = 0; return }
  loading.value = true
  try {
    const res = await listPipelines(pid, { page: pagination.page, page_size: pagination.pageSize })
    let list = res?.items || res?.data || res?.list || res?.pipelines || []
    list = Array.isArray(list) ? list : []
    if (searchForm.name) {
      const kw = String(searchForm.name).toLowerCase()
      list = list.filter(p => String(p.name || '').toLowerCase().includes(kw))
    }
    items.value = list
    const total = res?.pagination?.total_items || list.length
    pagination.total = Number(total) || 0
  } catch (e) {
    ElMessage.error(`加载流水线失败：${e?.message || e}`)
  } finally {
    loading.value = false
  }
}

const handlePageChange = async (p) => { pagination.page = p; await fetchList() }

const handleSearch = () => { pagination.page = 1; fetchList() }
const resetSearch = () => { searchForm.name = ''; handleSearch() }
const refresh = async () => { await fetchList() }

const handleAdd = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { ElMessage.warning('请先选择项目'); return }
  try {
    const { value } = await ElMessageBox.prompt('请输入流水线名称（可留空自动生成）', '新建流水线', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：构建与测试'
    })
    const inputName = (value || '').trim()
    const name = inputName || `pipeline_${Date.now()}`
    // 创建时初始化 YAML 顶层 name 与流水线名一致
    const res = await createPipeline({ projectId: pid, name, workflow_yaml: `name: ${name}\n` })
    const id = res?.pipeline?.id || res?.data?.pipeline?.id || res?.data?.id || res?.id
    ElMessage.success('流水线创建成功')
    if (id) {
      router.push(`/ci/pipeline/${id}`)
    } else {
      await fetchList()
    }
  } catch (e) {
    if (e === 'cancel') return
    ElMessage.error(`创建失败：${e?.message || e}`)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除流水线「${row?.name || row?.id}」？`, '提示', { type: 'warning' })
    await deletePipeline(row.id)
    ElMessage.success('已删除')
    await fetchList()
  } catch (e) {
    if (String(e).includes('cancel')) return
    ElMessage.error(`删除失败：${e?.message || e}`)
  }
}

const goDetail = (row) => { router.push(`/ci/pipeline/${row.id}`) }
const goBuilds = (row) => { router.push(`/ci/pipeline/${row.id}/builds`) }
const runPipeline = async (row) => {
  try {
    const resp = await startPipelineBuild(String(row.id), {})
    const build = (resp && (resp.build || (resp.data && (resp.data.build || resp.data)) || resp)) || {}
    const bid = (build && build.id)
    if (bid) { router.push(`/ci/builds/${bid}`) } else { ElMessage.warning('运行已触发，但未返回构建ID') }
  } catch (e) { ElMessage.error(`运行失败：${e?.message || e}`) }
}

// 监听项目变化自动刷新（包括初始加载）
watch(() => projectStore.selectedProject, (newProject) => {
  if (newProject?.id) {
    fetchList()
  }
}, { immediate: true })
</script>

<style scoped>
.project-section-layout { min-height: calc(100vh - 60px); }
.project-section-main { padding: 0; }
.pipelines-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap: 12px; }
.toolbar { display:flex; align-items:center; gap: 8px; flex-wrap: nowrap; }
.toolbar-input { width: 220px; max-width: 220px; }
.toolbar-btn :deep(.el-icon) { margin-right: 4px; }
.empty { padding: 20px 0; }
.pagination-container { margin-top: 16px; display:flex; justify-content:flex-end; }
</style>
<style scoped>
.compact-top { padding-top: 8px; }
</style>