<template>
  <div class="artifact-repositories-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>制品仓库管理</span>
          <div class="toolbar">
            <el-input
              v-model="searchForm.name"
              placeholder="按仓库名称搜索"
              clearable
              class="toolbar-input"
              @keyup.enter="handleSearch"
            />
            <el-input
              v-model="searchForm.namespaceId"
              placeholder="命名空间ID(可选)"
              clearable
              class="toolbar-input-narrow"
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
              <el-icon><Plus /></el-icon>新增仓库
            </el-button>
          </div>
        </div>
      </template>


      <el-table v-loading="loading" :data="repositoryList" border style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="namespace_id" label="命名空间ID" />
        <el-table-column prop="path" label="路径" />
        <el-table-column prop="is_public" label="公开" :formatter="formatBool" width="100" />
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

    <el-dialog v-model="dialogVisible" :title="dialogType==='add'?'新增仓库':'编辑仓库'" width="600px" @close="handleDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述" prop="description"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="命名空间ID" prop="namespace_id"><el-input v-model="form.namespace_id" /></el-form-item>
        <el-form-item label="路径" prop="path"><el-input v-model="form.path" /></el-form-item>
        <el-form-item label="公开" prop="is_public"><el-switch v-model="form.is_public" /></el-form-item>
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
import { listRepositories, createRepository, updateRepository, deleteRepository } from '@/api/artifact/repository'

const loading = ref(false)
const submitting = ref(false)
const searchForm = reactive({ name: '', namespaceId: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const repositoryList = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref(null)
const form = reactive({ id: '', name: '', description: '', namespace_id: '', path: '', is_public: false })
const formRules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }], namespace_id: [{ required: true, message: '请输入命名空间ID', trigger: 'blur' }] }

const formatBool = (_r,_c,val)=> (val ? '是' : '否')

const fetchRepositories = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.pageSize, namespace_id: searchForm.namespaceId, name: searchForm.name }
    const res = await listRepositories(params)
    repositoryList.value = res.data || []
    pagination.total = res.pagination?.total_items || 0
  } catch (e) { console.error('获取仓库失败:', e); ElMessage.error('获取仓库失败') } finally { loading.value = false }
}

const handleSearch = ()=>{ pagination.page=1; fetchRepositories() }
const resetSearch = ()=>{ searchForm.name=''; searchForm.namespaceId=''; handleSearch() }
const refresh = async ()=>{ await fetchRepositories() }
const handleSizeChange = (s)=>{ pagination.pageSize=s; fetchRepositories() }
const handleCurrentChange = (p)=>{ pagination.page=p; fetchRepositories() }
const handleAdd = ()=>{ dialogType.value='add'; dialogVisible.value=true }
const handleEdit = (row)=>{ dialogType.value='edit'; dialogVisible.value=true; Object.assign(form, row) }
const handleDelete = (row)=>{
  ElMessageBox.confirm(`确定删除仓库 "${row.name}"?`, '提示', { type: 'warning' })
    .then(async ()=>{ await deleteRepository(row.id); ElMessage.success('删除成功'); fetchRepositories() })
    .catch(()=>{})
}
const handleSubmit = ()=>{
  formRef.value.validate(async (valid)=>{
    if (!valid) return
    submitting.value = true
    try {
      if (dialogType.value==='add') { await createRepository({ name: form.name, description: form.description, namespace_id: form.namespace_id, is_public: form.is_public, path: form.path }); ElMessage.success('新增成功') }
      else { const { id, ...payload } = form; await updateRepository(id, payload); ElMessage.success('编辑成功') }
      dialogVisible.value=false; fetchRepositories()
    } catch(e){ console.error('提交失败:',e); ElMessage.error('提交失败') } finally { submitting.value=false }
  })
}
const handleDialogClose = ()=>{ formRef.value?.resetFields(); Object.assign(form, { id:'', name:'', description:'', namespace_id:'', path:'', is_public:false }) }

onMounted(()=>{ fetchRepositories() })
</script>

<style scoped>
.artifact-repositories-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap: 12px; }
.toolbar { display:flex; align-items:center; gap: 8px; flex-wrap: wrap; }
.toolbar-input { width: 220px; max-width: 220px; }
.toolbar-input-narrow { width: 180px; max-width: 180px; }
.toolbar-btn :deep(.el-icon) { margin-right: 4px; }
.pagination-container { margin-top: 20px; display: flex; justify-content: flex-end; }
</style>