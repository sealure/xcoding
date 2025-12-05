<template>
  <div class="repository-detail-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>代码仓库详情</span>
          <el-button @click="goBack">
            <el-icon><Back /></el-icon>返回
          </el-button>
        </div>
      </template>
      
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="10" animated />
      </div>
      
      <div v-else class="repository-detail-content">
        <div class="repository-info-section">
          <h3>基本信息</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="ID">{{ repositoryInfo.id }}</el-descriptions-item>
            <el-descriptions-item label="仓库名称">{{ repositoryInfo.name }}</el-descriptions-item>
            <el-descriptions-item label="仓库地址">{{ repositoryInfo.url }}</el-descriptions-item>
            <el-descriptions-item label="所属项目">{{ repositoryInfo.projectName }}</el-descriptions-item>
            <el-descriptions-item label="仓库描述">{{ repositoryInfo.description }}</el-descriptions-item>
            <el-descriptions-item label="创建者">{{ repositoryInfo.creator }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ repositoryInfo.createdAt }}</el-descriptions-item>
            <el-descriptions-item label="最后更新">{{ repositoryInfo.updatedAt }}</el-descriptions-item>
          </el-descriptions>
          
          <div class="action-buttons">
            <el-button type="primary" @click="handleEdit">
              <el-icon><Edit /></el-icon>编辑仓库
            </el-button>
            <el-button type="danger" @click="handleDelete">
              <el-icon><Delete /></el-icon>删除仓库
            </el-button>
          </div>
        </div>
        
        <div class="repository-commits-section">
          <div class="section-header">
            <h3>最近提交</h3>
            <el-button type="primary" size="small" @click="refreshCommits">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
          </div>
          <el-table :data="recentCommits" border style="width: 100%">
            <el-table-column prop="hash" label="提交哈希" width="100" />
            <el-table-column prop="message" label="提交信息" show-overflow-tooltip />
            <el-table-column prop="author" label="作者" width="150" />
            <el-table-column prop="createdAt" label="提交时间" width="180" />
          </el-table>
        </div>
        
        <div class="repository-branches-section">
          <div class="section-header">
            <h3>分支列表</h3>
            <el-button type="primary" size="small" @click="refreshBranches">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
          </div>
          <el-table :data="repositoryBranches" border style="width: 100%">
            <el-table-column prop="name" label="分支名称" />
            <el-table-column prop="lastCommit" label="最新提交" show-overflow-tooltip />
            <el-table-column prop="lastCommitAuthor" label="提交者" width="150" />
            <el-table-column prop="lastCommitAt" label="提交时间" width="180" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="viewBranch(row)">
                  查看
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-card>
    
    <!-- 编辑仓库对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑代码仓库"
      width="500px"
      @close="handleEditDialogClose"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="100px"
      >
        <el-form-item label="仓库名称" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入仓库名称" />
        </el-form-item>
        <el-form-item label="仓库地址" prop="url">
          <el-input v-model="editForm.url" placeholder="请输入仓库地址" />
        </el-form-item>
        <el-form-item label="所属项目" prop="projectId">
          <el-select v-model="editForm.projectId" placeholder="请选择项目" style="width: 100%">
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
            v-model="editForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入仓库描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleEditSubmit" :loading="submitting">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRepositoryById, updateRepository, deleteRepository, getRepositoryCommits, getRepositoryBranches } from '@/api/repository'
import { getProjectList } from '@/api/project'

const route = useRoute()
const router = useRouter()

// 加载状态
const loading = ref(false)
const submitting = ref(false)

// 仓库信息
const repositoryInfo = reactive({
  id: '',
  name: '',
  url: '',
  projectId: '',
  projectName: '',
  description: '',
  creator: '',
  createdAt: '',
  updatedAt: ''
})

// 项目选项
const projectOptions = ref([])

// 最近提交
const recentCommits = ref([])

// 分支列表
const repositoryBranches = ref([])

// 编辑对话框
const editDialogVisible = ref(false)
const editFormRef = ref(null)
const editForm = reactive({
  name: '',
  url: '',
  projectId: '',
  description: ''
})

// 编辑表单验证规则
const editFormRules = {
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
  ]
}

// 获取项目列表（用于下拉选择）
const fetchProjectOptions = async () => {
  try {
    const res = await getProjectList({ page: 1, page_size: 1000 })
    projectOptions.value = res.data || []
  } catch (error) {
    console.error('获取项目列表失败:', error)
    ElMessage.error('获取项目列表失败')
  }
}

// 获取仓库详情
const fetchRepositoryDetail = async () => {
  loading.value = true
  try {
    const repositoryId = route.params.id
    const res = await getRepositoryById(repositoryId)
    
    // 填充仓库信息
    Object.assign(repositoryInfo, res.data)
    
    // 获取仓库提交和分支
    const [commitsRes, branchesRes] = await Promise.all([
      getRepositoryCommits(repositoryId, { page: 1, page_size: 10 }),
      getRepositoryBranches(repositoryId)
    ])
    
    recentCommits.value = commitsRes.data || []
    repositoryBranches.value = branchesRes.data || []
  } catch (error) {
    console.error('获取仓库详情失败:', error)
    ElMessage.error('获取仓库详情失败')
  } finally {
    loading.value = false
  }
}

// 返回上一页
const goBack = () => {
  router.back()
}

// 编辑仓库
const handleEdit = () => {
  // 填充编辑表单
  editForm.name = repositoryInfo.name
  editForm.url = repositoryInfo.url
  editForm.projectId = repositoryInfo.projectId
  editForm.description = repositoryInfo.description
  
  editDialogVisible.value = true
}

// 提交编辑
const handleEditSubmit = () => {
  editFormRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        await updateRepository(repositoryInfo.id, editForm)
        ElMessage.success('编辑仓库成功')
        editDialogVisible.value = false
        fetchRepositoryDetail() // 刷新数据
      } catch (error) {
        console.error('编辑仓库失败:', error)
        ElMessage.error('编辑仓库失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

// 关闭编辑对话框
const handleEditDialogClose = () => {
  editFormRef.value?.resetFields()
}

// 删除仓库
const handleDelete = () => {
  ElMessageBox.confirm(
    `确定要删除代码仓库 "${repositoryInfo.name}" 吗？此操作不可恢复！`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await deleteRepository(repositoryInfo.id)
      ElMessage.success('删除仓库成功')
      router.push('/repositories') // 返回仓库列表
    } catch (error) {
      console.error('删除仓库失败:', error)
      ElMessage.error('删除仓库失败')
    }
  }).catch(() => {})
}

// 刷新提交记录
const refreshCommits = async () => {
  try {
    const res = await getRepositoryCommits(repositoryInfo.id, { page: 1, page_size: 10 })
    recentCommits.value = res.data || []
    ElMessage.success('刷新提交记录成功')
  } catch (error) {
    console.error('刷新提交记录失败:', error)
    ElMessage.error('刷新提交记录失败')
  }
}

// 刷新分支列表
const refreshBranches = async () => {
  try {
    const res = await getRepositoryBranches(repositoryInfo.id)
    repositoryBranches.value = res.data || []
    ElMessage.success('刷新分支列表成功')
  } catch (error) {
    console.error('刷新分支列表失败:', error)
    ElMessage.error('刷新分支列表失败')
  }
}

// 查看分支详情
const viewBranch = (branch) => {
  // 这里可以跳转到分支详情页面或打开分支详情对话框
  ElMessage.info(`查看分支: ${branch.name}`)
}

// 组件挂载时获取数据
onMounted(async () => {
  await fetchProjectOptions()
  fetchRepositoryDetail()
})
</script>

<style scoped>
.repository-detail-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.loading-container {
  padding: 20px 0;
}

.repository-detail-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.repository-info-section, .repository-commits-section, .repository-branches-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

h3 {
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}
</style>