<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="repositories-container">
        <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>代码仓库管理</span>
          <div class="toolbar">
            <el-input
              v-model="searchForm.name"
              placeholder="按仓库名称搜索"
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
              <el-icon><Plus /></el-icon>新增代码仓库
            </el-button>
          </div>
        </div>
      </template>
      
      
      
      <!-- 表格区域 -->
      <el-table
        v-loading="loading"
        :data="repositoryList"
        style="width: 100%"
        border
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="仓库名称" />
        <el-table-column prop="git_url" label="仓库地址" show-overflow-tooltip />
        <el-table-column label="所属项目">
          <template #default="{ row }">{{ getProjectName(row.project_id) }}</template>
        </el-table-column>
        <el-table-column label="认证方式" width="160">
          <template #default="{ row }">{{ formatAuth(row.auth_type) }}</template>
        </el-table-column>
        <el-table-column label="默认分支" width="160">
          <template #default="{ row }">
            <Branch :repository="row" @branch-updated="(name) => row.branch = name" />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页区域 -->
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
      </div>
    </el-main>
  </el-container>

  <!-- 代码仓库表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增代码仓库' : '编辑代码仓库'"
      width="500px"
      @close="handleDialogClose"
    >
      <el-form
        ref="repositoryFormRef"
        :model="repositoryForm"
        :rules="repositoryFormRules"
        label-width="100px"
      >
        <el-form-item label="仓库名称" prop="name">
          <el-input v-model="repositoryForm.name" placeholder="请输入仓库名称" />
        </el-form-item>
        <el-form-item label="仓库地址" prop="url">
          <el-input v-model="repositoryForm.url" placeholder="请输入仓库地址" />
        </el-form-item>
        <el-form-item label="认证方式" prop="authType">
          <el-select v-model="repositoryForm.authType" placeholder="请选择认证方式" style="width: 100%">
            <el-option label="无认证" value="REPOSITORY_AUTH_TYPE_NONE" />
            <el-option label="用户名密码" value="REPOSITORY_AUTH_TYPE_PASSWORD" />
            <el-option label="SSH密钥" value="REPOSITORY_AUTH_TYPE_SSH" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_PASSWORD'" label="Git用户名" prop="gitUsername">
          <el-input v-model="repositoryForm.gitUsername" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_PASSWORD'" label="Git密码" prop="gitPassword">
          <el-input v-model="repositoryForm.gitPassword" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item v-if="repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_SSH'" label="SSH私钥" prop="gitSSHKey">
          <el-input v-model="repositoryForm.gitSSHKey" type="textarea" :rows="4" placeholder="粘贴私钥内容（PEM）" />
        </el-form-item>
        <el-form-item label="默认分支" prop="branch">
          <el-input v-model="repositoryForm.branch" placeholder="默认 main，可修改" />
        </el-form-item>
        <!-- 新增模式下项目固定为当前项目，隐藏选择；编辑模式保留显示 -->
        <el-form-item v-if="dialogType === 'edit'" label="所属项目" prop="projectId">
          <el-select v-model="repositoryForm.projectId" placeholder="请选择项目" style="width: 100%">
            <el-option
              v-for="project in projectOptions"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="仓库描述" prop="description">
          <el-input
            v-model="repositoryForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入仓库描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
    
  </template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRepositoryList, createRepository, updateRepository, deleteRepository } from '@/api/code_repository/repository'
import Branch from './branch.vue'
import { getProjectList } from '@/api/project'
import { useProjectStore } from '@/stores/project'
import ProjectTabs from '@/components/ProjectTabs.vue'

// 加载状态
const loading = ref(false)
const submitting = ref(false)

// 搜索表单
const searchForm = reactive({
  name: ''
})

// 分页参数
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 代码仓库列表
const repositoryList = ref([])

// 项目选项
const projectOptions = ref([])
const projectStore = useProjectStore()

// 对话框状态
const dialogVisible = ref(false)
const dialogType = ref('add') // 'add' 或 'edit'

// 代码仓库表单
const repositoryFormRef = ref(null)
const repositoryForm = reactive({
  id: '',
  name: '',
  url: '',
  projectId: '',
  description: '',
  authType: 'REPOSITORY_AUTH_TYPE_NONE',
  gitUsername: '',
  gitPassword: '',
  gitSSHKey: '',
  branch: 'main'
})

// 代码仓库表单验证规则
const repositoryFormRules = {
  name: [
    { required: true, message: '请输入仓库名称', trigger: 'blur' },
    { min: 2, max: 50, message: '仓库名称长度在2到50个字符之间', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入仓库地址', trigger: 'blur' },
    { type: 'url', message: '请输入正确的URL格式', trigger: 'blur' }
  ],
  projectId: [
    { required: true, message: '请选择所属项目', trigger: 'change' }
  ],
  description: [
    { max: 500, message: '仓库描述不能超过500个字符', trigger: 'blur' }
  ],
  authType: [
    { required: true, message: '请选择认证方式', trigger: 'change' }
  ],
  gitUsername: [
    { required: () => repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_PASSWORD', message: '请输入用户名', trigger: 'blur' }
  ],
  gitPassword: [
    { required: () => repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_PASSWORD', message: '请输入密码', trigger: 'blur' }
  ],
  gitSSHKey: [
    { required: () => repositoryForm.authType === 'REPOSITORY_AUTH_TYPE_SSH', message: '请输入私钥', trigger: 'blur' }
  ],
  branch: [
    { required: true, message: '请输入默认分支', trigger: 'blur' }
  ]
}

// 获取项目列表（用于下拉选择）
const fetchProjectOptions = async () => {
  try {
    if (!projectStore.projectOptions.length) {
      const res = await getProjectList({ page: 1, page_size: 100 })
      projectStore.projectOptions = res.data || []
    }
    projectOptions.value = projectStore.projectOptions
  } catch (error) {
    console.error('获取项目列表失败:', error)
    ElMessage.error('获取项目列表失败')
  }
}

// 获取代码仓库列表
const fetchRepositoryList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      name: searchForm.name
    }
    const pid = projectStore.selectedProject?.id
    if (!pid) {
      // 未选择项目时保持空列表，提示用户选择项目
      repositoryList.value = []
      pagination.total = 0
      return
    }
    const res = await getRepositoryList(pid, params)
    repositoryList.value = res.data || res.items || []
    pagination.total = res.pagination?.total_items || res.total || repositoryList.value.length || 0
  } catch (error) {
    console.error('获取代码仓库列表失败:', error)
    ElMessage.error('获取代码仓库列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchRepositoryList()
}

// 重置搜索
const resetSearch = () => {
  searchForm.name = ''
  handleSearch()
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.pageSize = size
  fetchRepositoryList()
}

// 当前页改变
const handleCurrentChange = (page) => {
  pagination.page = page
  fetchRepositoryList()
}

// 刷新列表
const refresh = async () => {
  await fetchRepositoryList()
}

// 新增代码仓库
const handleAdd = () => {
  dialogType.value = 'add'
  dialogVisible.value = true
  // 自动填入当前已选择项目
  const current = projectStore.selectedProject?.id
  repositoryForm.projectId = current || ''
  if (!current) {
    ElMessage.warning('请先选择当前项目，再新增仓库')
  }
}

// 编辑代码仓库
const handleEdit = (row) => {
  dialogType.value = 'edit'
  dialogVisible.value = true
  
  // 填充表单数据
  repositoryForm.id = row.id
  repositoryForm.name = row.name
  repositoryForm.url = row.git_url
  repositoryForm.projectId = row.project_id
  repositoryForm.description = row.description
  repositoryForm.authType = row.auth_type || 'REPOSITORY_AUTH_TYPE_NONE'
  repositoryForm.gitUsername = row.git_username || ''
  repositoryForm.branch = row.branch || 'main'
}

// 删除代码仓库
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除代码仓库 "${row.name}" 吗？`,
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await deleteRepository(row.id, row.project_id)
      ElMessage.success(`代码仓库 "${row.name}" 删除成功`)
      fetchRepositoryList()
    } catch (error) {
      console.error('删除代码仓库失败:', error)
      ElMessage.error('删除代码仓库失败')
    }
  }).catch(() => {})
}

// 提交表单
const handleSubmit = () => {
  repositoryFormRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (dialogType.value === 'add') {
          // 新增代码仓库
          await createRepository(repositoryForm.projectId, repositoryForm)
          ElMessage.success(`代码仓库 "${repositoryForm.name}" 新增成功`)
        } else {
          // 编辑代码仓库
          const { id, projectId, ...data } = repositoryForm
          await updateRepository(id, data, projectId)
          ElMessage.success(`代码仓库 "${repositoryForm.name}" 编辑成功`)
        }
        
        dialogVisible.value = false
        fetchRepositoryList()
      } catch (error) {
        console.error('提交失败:', error)
        ElMessage.error(error.message || '提交失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

// 关闭对话框
const handleDialogClose = () => {
  // 重置表单
  repositoryFormRef.value?.resetFields()
  Object.assign(repositoryForm, {
    id: '',
    name: '',
    url: '',
    projectId: '',
    description: '',
    authType: 'REPOSITORY_AUTH_TYPE_NONE',
    gitUsername: '',
    gitPassword: '',
    gitSSHKey: '',
    branch: 'main'
  })
}

// 组件挂载时获取数据
onMounted(async () => {
  await projectStore.loadPersisted()
  await fetchProjectOptions()
  fetchRepositoryList()
})

// 辅助：项目名称、认证方式、时间格式化
const getProjectName = (pid) => {
  const p = projectOptions.value?.find((x) => String(x.id) === String(pid))
  return p ? p.name : pid
}
const formatAuth = (auth) => {
  switch (auth) {
    case 'REPOSITORY_AUTH_TYPE_NONE': return '无认证'
    case 'REPOSITORY_AUTH_TYPE_PASSWORD': return '用户名密码'
    case 'REPOSITORY_AUTH_TYPE_SSH': return 'SSH密钥'
    default: return '未指定'
  }
}
const formatDate = (iso) => {
  if (!iso) return '—'
  try { return new Date(iso).toLocaleString() } catch { return iso }
}

</script>

<style scoped>
.project-section-layout { min-height: calc(100vh - 60px); }
.project-section-main { padding: 0; }
.repositories-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar { display:flex; align-items:center; gap: 8px; flex-wrap: nowrap; }
.toolbar-input { width: 220px; max-width: 220px; }
.toolbar-btn :deep(.el-icon) { margin-right: 4px; }

/* 旧搜索区域样式移除，统一由 toolbar 控制 */

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 默认分支列悬停显示“选择分支”链接 */
.branch-cell{display:flex;align-items:center}
.branch-name{flex:none}
</style>