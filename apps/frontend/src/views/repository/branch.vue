<template>
  <div class="branch-cell">
    <span class="branch-name">{{ repository.branch || '—' }}</span>

    <el-popover placement="bottom" trigger="hover" width="240" @show="loadBranches">
      <template #reference>
        <el-link class="select-branch-link" type="primary">选择分支</el-link>
      </template>
      <div v-if="branchLoading" class="branch-popover-loading">加载中…</div>
      <div v-else>
        <div v-if="!(branches && branches.length)" class="branch-empty">暂无分支</div>
        <ul v-else class="branch-list">
          <li
            v-for="b in branches"
            :key="b.id"
            class="branch-item"
            @click="handleSelectBranch(b)"
          >
            <span class="branch-item-name">{{ b.name }}</span>
            <el-tag v-if="b.is_default" type="success" size="small">默认</el-tag>
          </li>
        </ul>
        <div class="branch-actions">
          <el-link type="primary" @click="openCreateBranch">新建分支</el-link>
        </div>
      </div>
    </el-popover>

    <!-- 新建分支对话框 -->
    <el-dialog
      v-model="createBranchDialogVisible"
      title="新建分支"
      width="420px"
      @close="handleCreateBranchDialogClose"
    >
      <el-form
        ref="createBranchFormRef"
        :model="createBranchForm"
        :rules="createBranchFormRules"
        label-width="100px"
      >
        <el-form-item label="分支名称" prop="name">
          <el-input v-model="createBranchForm.name" placeholder="如：feature/login" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="createBranchForm.isDefault" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createBranchDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreateBranchSubmit" :loading="createBranchSubmitting">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
  
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { listBranches, createBranch, updateBranch } from '@/api/code_repository/branch'

const props = defineProps({
  repository: { type: Object, required: true }
})
const emit = defineEmits(['branch-updated'])

// 分支列表状态（当前仓库）
const branches = ref([])
const branchLoading = ref(false)

// 新建分支对话框与表单
const createBranchDialogVisible = ref(false)
const createBranchSubmitting = ref(false)
const createBranchFormRef = ref(null)
const createBranchForm = reactive({ name: '', isDefault: false })
const createBranchFormRules = {
  name: [
    { required: true, message: '请输入分支名称', trigger: 'blur' }
  ]
}

const loadBranches = async () => {
  branchLoading.value = true
  try {
    const res = await listBranches(props.repository.id, { page: 1, page_size: 100, project_id: props.repository.project_id })
    branches.value = res.data || []
  } catch (e) {
    branches.value = []
  } finally {
    branchLoading.value = false
  }
}

const handleSelectBranch = async (branch) => {
  if (branch?.is_default) {
    emit('branch-updated', branch.name)
    ElMessage.info('该分支已是默认分支')
    return
  }
  try {
    await updateBranch(props.repository.id, branch.id, { projectId: props.repository.project_id, isDefault: true })
    ElMessage.success(`默认分支已切换为 ${branch.name}`)
    emit('branch-updated', branch.name)
    // 更新列表标记
    branches.value = (branches.value || []).map(b => ({ ...b, is_default: b.id === branch.id }))
  } catch (e) {
    console.error('切换默认分支失败:', e)
    ElMessage.error(e.message || '切换默认分支失败')
  }
}

const openCreateBranch = () => {
  createBranchForm.name = ''
  createBranchForm.isDefault = false
  createBranchDialogVisible.value = true
}

const handleCreateBranchSubmit = () => {
  createBranchFormRef.value?.validate(async (valid) => {
    if (!valid) return
    createBranchSubmitting.value = true
    try {
      const payload = {
        name: createBranchForm.name,
        projectId: props.repository.project_id,
        isDefault: createBranchForm.isDefault
      }
      await createBranch(props.repository.id, payload)
      ElMessage.success('分支创建成功')
      createBranchDialogVisible.value = false
      await loadBranches()
      if (createBranchForm.isDefault) {
        emit('branch-updated', createBranchForm.name)
      }
    } catch (error) {
      console.error('创建分支失败:', error)
      ElMessage.error(error.message || '创建分支失败')
    } finally {
      createBranchSubmitting.value = false
    }
  })
}

const handleCreateBranchDialogClose = () => {
  createBranchFormRef.value?.resetFields()
  createBranchDialogVisible.value = false
}
</script>

<style scoped>
.branch-cell {
  display: flex;
  align-items: center;
}
.branch-name { flex: none; }

.select-branch-link {
  margin-left: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}
.branch-cell:hover .select-branch-link { opacity: 1; }

.branch-popover-loading { padding: 8px; color: #909399; }
.branch-empty { padding: 8px; color: #909399; }
.branch-list { list-style: none; padding: 0; margin: 0; max-height: 200px; overflow-y: auto; }
.branch-item { display: flex; align-items: center; justify-content: space-between; padding: 6px 8px; cursor: pointer; border-radius: 4px; }
.branch-item:hover { background: #f5f7fa; }
.branch-item-name { font-size: 13px; }
.branch-actions { margin-top: 8px; text-align: right; }
</style>