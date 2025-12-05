<template>
  <div class="namespaces-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>命名空间管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增命名空间
          </el-button>
        </div>
      </template>

      <div class="search-container">
        <el-form :inline="true" :model="searchForm" class="search-form">
          <el-form-item label="名称">
            <el-input v-model="searchForm.name" placeholder="请输入名称" clearable />
          </el-form-item>
          <el-form-item label="注册表">
            <el-input v-model="searchForm.registryId" placeholder="注册表ID(可选)" clearable />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch"><el-icon><Search /></el-icon>搜索</el-button>
            <el-button @click="resetSearch"><el-icon><Refresh /></el-icon>重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table v-loading="loading" :data="namespaceList" border style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="registry_id" label="注册表ID" />
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

    <el-dialog v-model="dialogVisible" :title="dialogType==='add'?'新增命名空间':'编辑命名空间'" width="600px" @close="handleDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述" prop="description"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="注册表ID" prop="registry_id"><el-input v-model="form.registry_id" /></el-form-item>
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
import { listNamespaces, createNamespace, updateNamespace, deleteNamespace } from '@/api/artifact/namespace'

const loading = ref(false)
const submitting = ref(false)
const searchForm = reactive({ name: '', registryId: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const namespaceList = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref(null)
const form = reactive({ id: '', name: '', description: '', registry_id: '' })
const formRules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }], registry_id: [{ required: true, message: '请输入注册表ID', trigger: 'blur' }] }

const fetchNamespaces = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.pageSize, registry_id: searchForm.registryId, name: searchForm.name }
    const res = await listNamespaces(params)
    namespaceList.value = res.data || []
    pagination.total = res.pagination?.total_items || 0
  } catch (e) { console.error('获取命名空间失败:', e); ElMessage.error('获取命名空间失败') } finally { loading.value = false }
}

const handleSearch = () => { pagination.page = 1; fetchNamespaces() }
const resetSearch = () => { searchForm.name=''; searchForm.registryId=''; handleSearch() }
const handleSizeChange = (s)=>{ pagination.pageSize=s; fetchNamespaces() }
const handleCurrentChange = (p)=>{ pagination.page=p; fetchNamespaces() }
const handleAdd = ()=>{ dialogType.value='add'; dialogVisible.value=true }
const handleEdit = (row)=>{ dialogType.value='edit'; dialogVisible.value=true; Object.assign(form, row) }
const handleDelete = (row)=>{
  ElMessageBox.confirm(`确定删除命名空间 "${row.name}"?`, '提示', { type: 'warning' })
    .then(async ()=>{ await deleteNamespace(row.id); ElMessage.success('删除成功'); fetchNamespaces() })
    .catch(()=>{})
}
const handleSubmit = ()=>{
  formRef.value.validate(async (valid)=>{
    if (!valid) return
    submitting.value = true
    try {
      if (dialogType.value==='add') { await createNamespace({ name: form.name, description: form.description, registry_id: form.registry_id }); ElMessage.success('新增成功') }
      else { const { id, ...payload } = form; await updateNamespace(id, payload); ElMessage.success('编辑成功') }
      dialogVisible.value=false; fetchNamespaces()
    } catch(e){ console.error('提交失败:',e); ElMessage.error('提交失败') } finally { submitting.value=false }
  })
}
const handleDialogClose = ()=>{ formRef.value?.resetFields(); Object.assign(form, { id:'', name:'', description:'', registry_id:'' }) }

onMounted(()=>{ fetchNamespaces() })
</script>

<style scoped>
.namespaces-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.search-container { margin-bottom:20px; }
.pagination-container { margin-top: 20px; display: flex; justify-content: flex-end; }
</style>