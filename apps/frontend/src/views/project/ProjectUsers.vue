<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="project-users-container">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="title-area">
                <span>项目用户</span>
                <el-tag v-if="projectStore.selectedProject" type="info" size="small">{{ projectStore.selectedProject.name }}</el-tag>
              </div>
              <div class="actions">
                <el-button type="primary" size="small" @click="handleOpenAddDialog">
                  <el-icon><Plus /></el-icon>
                  添加成员
                </el-button>
              </div>
            </div>
          </template>

          <div v-if="!projectStore.selectedProject" class="empty">
            <el-empty description="请选择项目以查看成员">
              <template #description>
                请选择项目以查看成员或前往项目列表选择项目
              </template>
              <el-button type="primary" size="small" @click="go('/projects')">前往项目列表</el-button>
            </el-empty>
          </div>

          <div v-else>
            <div class="search-container">
              <el-form :inline="true" :model="searchForm" class="search-form">
                <el-form-item label="关键词">
                  <el-input v-model="searchForm.keyword" placeholder="按用户名/邮箱过滤" clearable />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="handleSearch">
                    <el-icon><Search /></el-icon>搜索
                  </el-button>
                  <el-button @click="resetSearch">
                    <el-icon><Refresh /></el-icon>重置
                  </el-button>
                </el-form-item>
              </el-form>
            </div>

            <el-table v-loading="loading" :data="filteredMembers" border style="width:100%">
              <el-table-column prop="user_id" label="用户ID" width="120" />
              <el-table-column label="用户名" min-width="160">
                <template #default="{ row }">
                  {{ row.user?.username || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="邮箱" min-width="200">
                <template #default="{ row }">
                  {{ row.user?.email || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="角色" width="220">
                <template #default="{ row }">
                  <el-select v-model="row.role" size="small" style="width: 180px" @change="val => handleRoleChange(row, val)">
                    <el-option v-for="opt in ROLE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
                  </el-select>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="160">
                <template #default="{ row }">
                  <el-button type="danger" size="small" @click="handleRemove(row)">
                    <el-icon><Delete /></el-icon>移除
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
          </div>
        </el-card>
      </div>
    </el-main>
  </el-container>

  <!-- 添加成员对话框 -->
  <el-dialog v-model="addDialogVisible" title="添加项目成员" width="500px" @close="handleAddDialogClose">
    <el-form ref="addFormRef" :model="addForm" :rules="addFormRules" label-width="100px">
      <el-form-item label="选择用户" prop="userId">
        <el-select v-model="addForm.userId" placeholder="请选择用户" filterable :disabled="addSubmitting" style="width:100%">
          <el-option v-for="u in userOptions" :key="u.id" :label="formatUserLabel(u)" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="角色" prop="role">
        <el-select v-model="addForm.role" placeholder="请选择角色" :disabled="addSubmitting" style="width:100%">
          <el-option v-for="opt in ROLE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSubmit" :loading="addSubmitting">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProjectTabs from '@/components/ProjectTabs.vue'
import { useProjectStore } from '@/stores/project'
import { listProjectMembers, addProjectMember, updateProjectMember, removeProjectMember } from '@/api/project'
import { getUserList } from '@/api/user'

const router = useRouter()
const projectStore = useProjectStore()

// 角色选项（与后端 proto 一致）
const ROLE_OPTIONS = [
  { label: '所有者', value: 'PROJECT_MEMBER_ROLE_OWNER' },
  { label: '管理员', value: 'PROJECT_MEMBER_ROLE_ADMIN' },
  { label: '成员', value: 'PROJECT_MEMBER_ROLE_MEMBER' },
  { label: '访客', value: 'PROJECT_MEMBER_ROLE_GUEST' }
]

// 列表与分页
const loading = ref(false)
const members = ref([])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ keyword: '' })
const filteredMembers = computed(() => {
  if (!searchForm.keyword) return members.value
  const k = searchForm.keyword.toLowerCase()
  return members.value.filter(m => {
    const uname = String(m.user?.username || '').toLowerCase()
    const email = String(m.user?.email || '').toLowerCase()
    return uname.includes(k) || email.includes(k) || String(m.user_id).includes(k)
  })
})

// 添加成员对话框
const addDialogVisible = ref(false)
const addSubmitting = ref(false)
const addFormRef = ref(null)
const addForm = reactive({ userId: '', role: 'PROJECT_MEMBER_ROLE_MEMBER' })
const addFormRules = {
  userId: [{ required: true, message: '请选择用户', trigger: 'change' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}
const userOptions = ref([])
const formatUserLabel = (u) => `${u.username || '用户'}${u.email ? '（' + u.email + '）' : ''}`

const go = (path) => { router.push(path) }

const fetchMembers = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { members.value = []; pagination.total = 0; return }
  loading.value = true
  try {
    const res = await listProjectMembers(pid, { page: pagination.page, page_size: pagination.pageSize })
    members.value = res.data || []
    pagination.total = res.pagination?.total_items || 0
  } catch (e) {
    console.error('获取项目成员失败:', e)
    ElMessage.error('获取项目成员失败')
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const res = await getUserList({ page: 1, page_size: 100 })
    userOptions.value = res.data || []
  } catch (e) {
    console.error('获取用户列表失败:', e)
    userOptions.value = []
  }
}

const handleSearch = () => { /* 前端过滤，无需请求 */ }
const resetSearch = () => { searchForm.keyword = '' }

const handleSizeChange = (size) => { pagination.pageSize = size; fetchMembers() }
const handleCurrentChange = (page) => { pagination.page = page; fetchMembers() }

const handleOpenAddDialog = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { ElMessage.warning('请先选择项目'); return }
  addDialogVisible.value = true
  // 加载用户选项
  if (!userOptions.value.length) await fetchUsers()
}

const handleAddDialogClose = () => {
  addForm.userId = ''
  addForm.role = 'PROJECT_MEMBER_ROLE_MEMBER'
}

const handleAddSubmit = () => {
  addFormRef.value?.validate(async (valid) => {
    if (!valid) return
    const pid = projectStore.selectedProject?.id
    if (!pid) { ElMessage.warning('请先选择项目'); return }
    addSubmitting.value = true
    try {
      await addProjectMember(pid, { user_id: addForm.userId, role: addForm.role })
      ElMessage.success('添加成功')
      addDialogVisible.value = false
      fetchMembers()
    } catch (e) {
      console.error('添加成员失败:', e)
      ElMessage.error('添加成员失败')
    } finally {
      addSubmitting.value = false
    }
  })
}

const handleRoleChange = async (row, newRole) => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { ElMessage.warning('请先选择项目'); return }
  const prev = row.role
  row.role = newRole
  try {
    await updateProjectMember(pid, row.user_id, { role: newRole })
    ElMessage.success('角色已更新')
  } catch (e) {
    console.error('更新成员角色失败:', e)
    ElMessage.error('更新成员角色失败')
    row.role = prev
  }
}

const handleRemove = (row) => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { ElMessage.warning('请先选择项目'); return }
  ElMessageBox.confirm(`确定移除成员 ${row.user?.username || row.user_id}？`, '提示', { type: 'warning' })
    .then(async () => {
      try {
        await removeProjectMember(pid, row.user_id)
        ElMessage.success('移除成功')
        fetchMembers()
      } catch (e) {
        console.error('移除成员失败:', e)
        ElMessage.error('移除成员失败')
      }
    })
    .catch(() => {})
}

watch(() => projectStore.selectedProject?.id, () => {
  pagination.page = 1
  fetchMembers()
})
onMounted(() => { fetchMembers() })
</script>

<style scoped>
.project-section-layout { min-height: calc(100vh - 60px); }
.project-section-main { padding: 0; }
.project-users-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.title-area { display:flex; align-items:center; gap:8px; }
.empty { padding: 20px 0; }
.search-container { margin-bottom: 12px; }
.pagination-container { margin-top: 16px; display:flex; justify-content:flex-end; }
</style>