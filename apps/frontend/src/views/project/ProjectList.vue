<template>
  <div class="project-list-container">
    <el-card shadow="hover" class="project-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">项目管理</span>
          <div class="toolbar">
            <el-input
              v-model="searchForm.name"
              placeholder="按项目名称搜索"
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
              <el-icon><Plus /></el-icon>新增项目
            </el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="projectList" border style="width: 100%">
        <el-table-column prop="id" label="ID" width="120" />
        <el-table-column prop="name" label="项目名称" min-width="200" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="enter(row)">
              <el-icon><View /></el-icon>进入
            </el-button>
            <el-button type="success" size="small" @click="handleEdit(row)" plain>
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)" plain>
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑项目 -->
    <el-dialog v-model="dialogVisible" :title="dialogType==='add'?'新增项目':'编辑项目'" width="560px" @close="handleDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px" class="project-form">
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入项目名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="简要描述该项目" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible=false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useProjectStore } from '@/stores/project'
import { getProjectList, createProject, updateProject, deleteProject } from '@/api/project'

const router = useRouter()
const projectStore = useProjectStore()

// 加载状态与数据
const loading = ref(false)
const projectList = ref([])
const searchForm = reactive({ name: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

// 对话框
const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref()
const form = reactive({ id: '', name: '', description: '' })
const formRules = { name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }] }
const submitting = ref(false)

const fetchProjects = async () => {
  loading.value = true
  try {
    const res = await getProjectList({ page: pagination.page, page_size: pagination.pageSize })
    let data = res.data || []
    if (searchForm.name) {
      const kw = String(searchForm.name).toLowerCase()
      data = data.filter(p => String(p.name || '').toLowerCase().includes(kw))
    }
    projectList.value = data
    pagination.total = res.pagination?.total_items || data.length || 0
  } catch (e) {
    console.error('获取项目失败:', e)
    ElMessage.error('获取项目失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => { pagination.page = 1; fetchProjects() }
const resetSearch = () => { searchForm.name = ''; handleSearch() }
const handleSizeChange = (s) => { pagination.pageSize = s; fetchProjects() }
const handleCurrentChange = (p) => { pagination.page = p; fetchProjects() }

const refresh = async () => { await fetchProjects(); try { await projectStore.fetchProjectOptions() } catch (_) {} }
const handleAdd = () => { dialogType.value = 'add'; dialogVisible.value = true; Object.assign(form, { id: '', name: '', description: '' }) }
const handleEdit = (row) => { dialogType.value = 'edit'; dialogVisible.value = true; Object.assign(form, { id: row.id, name: row.name, description: row.description || '' }) }

const handleSubmit = () => {
  formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (dialogType.value === 'add') {
        await createProject({ name: form.name, description: form.description })
        ElMessage.success('新增成功')
      } else {
        await updateProject(form.id, { name: form.name, description: form.description })
        ElMessage.success('编辑成功')
      }
      dialogVisible.value = false
      await fetchProjects()
      try { await projectStore.fetchProjectOptions() } catch (_) {}
    } catch (e) {
      console.error('提交失败:', e)
      ElMessage.error('提交失败')
    } finally { submitting.value = false }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  Object.assign(form, { id: '', name: '', description: '' })
}

const enter = (p) => {
  projectStore.setSelectedProject({ id: p.id, name: p.name })
  router.push('/projects/overview')
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除项目「${row.name}」？此操作不可撤销。`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch (_) { return }
  try {
    await deleteProject(row.id)
    ElMessage.success('删除成功')
    await fetchProjects()
    try { await projectStore.fetchProjectOptions() } catch (_) {}
  } catch (e) {
    console.error('删除失败:', e)
    ElMessage.error('删除失败')
  }
}

onMounted(async () => {
  await fetchProjects()
  try { await projectStore.loadPersisted() } catch (_) {}
})
</script>

<style scoped>
.project-list-container { padding: 20px; }
.project-card { border-radius: 8px; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap: 12px; }
.card-title { font-weight: 600; font-size: 16px; }
.toolbar { display:flex; align-items:center; gap: 8px; flex-wrap: nowrap; }
.toolbar-input { width: 220px; max-width: 220px; }
.toolbar-btn :deep(.el-icon) { margin-right: 4px; }
.pagination-container { margin-top: 12px; display:flex; justify-content:flex-end; }
.project-form { padding-top: 8px; }
</style>