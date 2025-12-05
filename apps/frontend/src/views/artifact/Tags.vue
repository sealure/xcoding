<template>
  <div class="tags-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>制品标签管理</span>
          <el-button type="primary" @click="handleAdd"><el-icon><Plus /></el-icon>新增标签</el-button>
        </div>
      </template>

      <div class="search-container">
        <el-form :inline="true" :model="searchForm" class="search-form">
          <el-form-item label="名称"><el-input v-model="searchForm.name" placeholder="请输入名称" clearable /></el-form-item>
          <el-form-item label="仓库ID"><el-input v-model="searchForm.repositoryId" placeholder="仓库ID(可选)" clearable /></el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch"><el-icon><Search /></el-icon>搜索</el-button>
            <el-button @click="resetSearch"><el-icon><Refresh /></el-icon>重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="tagList" border style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="digest" label="摘要" />
        <el-table-column prop="repository_id" label="仓库ID" />
        <el-table-column prop="size_bytes" label="大小(字节)" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)"><el-icon><Edit /></el-icon>编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)"><el-icon><Delete /></el-icon>删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogType==='add'?'新增标签':'编辑标签'" width="600px" @close="handleDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="摘要" prop="digest"><el-input v-model="form.digest" /></el-form-item>
        <el-form-item label="仓库ID" prop="repository_id"><el-input v-model="form.repository_id" /></el-form-item>
        <el-form-item label="大小(字节)" prop="size_bytes"><el-input v-model.number="form.size_bytes" /></el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible=false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listTags, createTag, updateTag, deleteTag } from '@/api/artifact/tag'

const loading = ref(false)
const submitting = ref(false)
const searchForm = reactive({ name: '', repositoryId: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const tagList = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref(null)
const form = reactive({ id: '', name: '', digest: '', repository_id: '', size_bytes: 0 })
const formRules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }], repository_id: [{ required: true, message: '请输入仓库ID', trigger: 'blur' }] }

const fetchTags = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.pageSize, repository_id: searchForm.repositoryId, name: searchForm.name }
    const res = await listTags(params)
    tagList.value = res.data || []
    pagination.total = res.pagination?.total_items || 0
  } catch (e) { console.error('获取标签失败:', e); ElMessage.error('获取标签失败') } finally { loading.value = false }
}

const handleSearch = ()=>{ pagination.page=1; fetchTags() }
const resetSearch = ()=>{ searchForm.name=''; searchForm.repositoryId=''; handleSearch() }
const handleSizeChange = (s)=>{ pagination.pageSize=s; fetchTags() }
const handleCurrentChange = (p)=>{ pagination.page=p; fetchTags() }
const handleAdd = ()=>{ dialogType.value='add'; dialogVisible.value=true }
const handleEdit = (row)=>{ dialogType.value='edit'; dialogVisible.value=true; Object.assign(form, row) }
const handleDelete = (row)=>{
  ElMessageBox.confirm(`确定删除标签 "${row.name}"?`, '提示', { type: 'warning' })
    .then(async ()=>{ await deleteTag(row.id); ElMessage.success('删除成功'); fetchTags() })
    .catch(()=>{})
}
const handleSubmit = ()=>{
  formRef.value.validate(async (valid)=>{
    if (!valid) return
    submitting.value = true
    try {
      if (dialogType.value==='add') { await createTag({ name: form.name, digest: form.digest, repository_id: form.repository_id, size_bytes: form.size_bytes }); ElMessage.success('新增成功') }
      else { const { id, ...payload } = form; await updateTag(id, payload); ElMessage.success('编辑成功') }
      dialogVisible.value=false; fetchTags()
    } catch(e){ console.error('提交失败:',e); ElMessage.error('提交失败') } finally { submitting.value=false }
  })
}
const handleDialogClose = ()=>{ formRef.value?.resetFields(); Object.assign(form, { id:'', name:'', digest:'', repository_id:'', size_bytes:0 }) }

onMounted(()=>{ fetchTags() })
</script>

<style scoped>
.tags-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.search-container { margin-bottom:20px; }
.pagination-container { margin-top: 20px; display: flex; justify-content: flex-end; }
</style>