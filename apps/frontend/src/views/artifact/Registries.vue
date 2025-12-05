<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="registries-container">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>制品注册表管理</span>
              <div class="toolbar">
                <el-input
                  v-model="searchForm.name"
                  placeholder="按注册表名称搜索"
                  clearable
                  class="toolbar-input"
                  @keyup.enter="handleSearch"
                />
                <el-select
                  v-model="searchForm.artifactType"
                  placeholder="类型(全部)"
                  clearable
                  class="toolbar-select-narrow"
                  @change="handleSearch"
                >
                  <el-option v-for="opt in ARTIFACT_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
                <el-select
                  v-model="searchForm.artifactSource"
                  placeholder="来源(全部)"
                  clearable
                  class="toolbar-select"
                  @change="handleSearch"
                >
                  <el-option v-for="opt in ARTIFACT_SOURCE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
                <el-button type="primary" @click="handleSearch" class="toolbar-btn">
                  <el-icon><Search /></el-icon>搜索
                </el-button>
                <el-button @click="resetSearch" class="toolbar-btn">
                  <el-icon><Refresh /></el-icon>重置
                </el-button>
                <el-button type="primary" plain @click="refresh" class="toolbar-btn">
                  <el-icon><Refresh /></el-icon>刷新
                </el-button>
                <el-button type="primary" @click="handleAdd" class="toolbar-btn">
                  <el-icon><Plus /></el-icon>新增注册表
                </el-button>
                <el-button type="success" @click="handleAddNamespace" class="toolbar-btn">
                  <el-icon><Plus /></el-icon>新建制品库
                </el-button>
              </div>
            </div>
          </template>

          

          <el-table v-loading="loading" :data="registryList" border style="width:100%">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="url" label="地址" show-overflow-tooltip />
            <el-table-column prop="artifact_type" label="类型" :formatter="formatArtifactType" />
            <el-table-column prop="artifact_source" label="来源" :formatter="formatArtifactSource" />
            <el-table-column prop="is_public" label="公开" :formatter="formatBool" width="100" />
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

  <!-- 注册表新增/编辑 -->
  <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增注册表' : '编辑注册表'" width="600px" @close="handleDialogClose">
    <el-form ref="formRef" :model="form" :rules="formRules" label-width="120px">
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入名称" />
      </el-form-item>
      <el-form-item label="地址" prop="url">
        <el-input v-model="form.url" placeholder="请输入URL" />
      </el-form-item>
      <el-form-item label="描述" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="3" />
      </el-form-item>
      <el-form-item label="所属项目" prop="project_id">
        <el-select v-model="form.project_id" placeholder="请选择项目" style="width:100%">
          <el-option v-for="project in projectOptions" :key="project.id" :label="project.name" :value="project.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="类型" prop="artifact_type">
        <el-select v-model="form.artifact_type" placeholder="请选择类型" style="width:100%">
          <el-option v-for="opt in ARTIFACT_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="来源" prop="artifact_source">
        <el-select v-model="form.artifact_source" placeholder="请选择来源" style="width:100%">
          <el-option v-for="opt in ARTIFACT_SOURCE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="公开" prop="is_public">
        <el-switch v-model="form.is_public" />
      </el-form-item>
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" placeholder="认证用户名(可选)" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input v-model="form.password" type="password" placeholder="认证密码(可选)" />
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </span>
    </template>
  </el-dialog>

  <!-- 新建制品库（命名空间） -->
  <el-dialog v-model="namespaceDialogVisible" title="新建制品库" width="600px" @close="handleNamespaceDialogClose">
    <div style="margin-bottom: 12px; display:flex; align-items:center; gap:8px;">
      <el-tag type="success" effect="dark">🐳 Docker 制品库</el-tag>
    </div>
    <el-form ref="namespaceFormRef" :model="namespaceForm" :rules="namespaceFormRules" label-width="120px">
      <el-form-item label="名称" prop="name">
        <el-input v-model="namespaceForm.name" placeholder="请输入制品库名称" />
      </el-form-item>
      <el-form-item label="描述" prop="description">
        <el-input v-model="namespaceForm.description" type="textarea" :rows="3" />
      </el-form-item>
      <el-form-item label="注册表" prop="registry_id">
        <el-select v-model="namespaceForm.registry_id" placeholder="请选择注册表" style="width:100%" filterable>
          <el-option v-for="reg in filteredRegistries" :key="reg.id" :label="reg.name + ' / ' + formatArtifactSource(null,null,reg.artifact_source)" :value="reg.id" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="namespaceDialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="submittingNamespace" @click="handleSubmitNamespace">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listRegistries, createRegistry, updateRegistry, deleteRegistry } from '@/api/artifact/registry'
import { createNamespace } from '@/api/artifact/namespace'
import { getProjectList } from '@/api/project'
import { useProjectStore } from '@/stores/project'
import ProjectTabs from '@/components/ProjectTabs.vue'

const ARTIFACT_TYPE = { UNSPECIFIED: 0, DOCKER: 1, GENERIC_FILE: 2 }
const ARTIFACT_TYPE_OPTIONS = [
  { label: 'Docker镜像', value: ARTIFACT_TYPE.DOCKER },
  { label: '泛型文件', value: ARTIFACT_TYPE.GENERIC_FILE }
]
const ARTIFACT_SOURCE = { UNSPECIFIED: 0, XCODING_REGISTRY: 1, ALI_REGISTRY: 2, SMB: 10, FTP: 11 }
const ARTIFACT_SOURCE_OPTIONS = [
  { label: 'XCoding Registry', value: ARTIFACT_SOURCE.XCODING_REGISTRY },
  { label: '阿里云 Registry', value: ARTIFACT_SOURCE.ALI_REGISTRY },
  { label: 'SMB', value: ARTIFACT_SOURCE.SMB },
  { label: 'FTP', value: ARTIFACT_SOURCE.FTP }
]

const loading = ref(false)
const submitting = ref(false)

const searchForm = reactive({ name: '', artifactType: '', artifactSource: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const registryList = ref([])
const projectOptions = ref([])
const projectStore = useProjectStore()

const filteredRegistries = computed(() => {
  let list = registryList.value || []
  if (searchForm.artifactType !== '' && searchForm.artifactType !== null && searchForm.artifactType !== undefined) {
    list = list.filter(r => r.artifact_type === searchForm.artifactType)
  }
  if (searchForm.artifactSource !== '' && searchForm.artifactSource !== null && searchForm.artifactSource !== undefined) {
    list = list.filter(r => r.artifact_source === searchForm.artifactSource)
  }
  return list
})

const dialogVisible = ref(false)
const dialogType = ref('add')
const formRef = ref(null)
const form = reactive({
  id: '', name: '', url: '', description: '', project_id: '',
  artifact_type: ARTIFACT_TYPE.DOCKER, artifact_source: ARTIFACT_SOURCE.XCODING_REGISTRY,
  is_public: false, username: '', password: ''
})

const formRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入URL', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

const formatBool = (_row, _col, val) => (val ? '是' : '否')
const formatArtifactType = (_row, _col, val) => {
  const opt = ARTIFACT_TYPE_OPTIONS.find(o => o.value === val)
  return opt ? opt.label : '未知'
}
const formatArtifactSource = (_row, _col, val) => {
  const opt = ARTIFACT_SOURCE_OPTIONS.find(o => o.value === val)
  return opt ? opt.label : '未知'
}

const fetchProjectOptions = async () => {
  try {
    if (!projectStore.projectOptions.length) {
      const res = await getProjectList({ page: 1, page_size: 100 })
      projectStore.projectOptions = res.data || []
    }
    projectOptions.value = projectStore.projectOptions
    // 保留项目选项以供对话框使用，不在筛选中使用项目ID
  } catch (e) { console.error('获取项目列表失败:', e) }
}

const fetchRegistries = async () => {
  loading.value = true
  try {
    const params = { page: pagination.page, page_size: pagination.pageSize, name: searchForm.name }
    const res = await listRegistries(params)
    let data = res.data || []
    if (searchForm.artifactType !== '' && searchForm.artifactType !== null && searchForm.artifactType !== undefined) {
      data = data.filter(r => r.artifact_type === searchForm.artifactType)
    }
    if (searchForm.artifactSource !== '' && searchForm.artifactSource !== null && searchForm.artifactSource !== undefined) {
      data = data.filter(r => r.artifact_source === searchForm.artifactSource)
    }
    registryList.value = data
    pagination.total = res.pagination?.total_items || 0
  } catch (e) {
    console.error('获取注册表失败:', e)
    ElMessage.error('获取注册表失败')
  } finally { loading.value = false }
}

const handleSearch = () => { pagination.page = 1; fetchRegistries() }
const resetSearch = () => { searchForm.name=''; searchForm.artifactType=''; searchForm.artifactSource=''; handleSearch() }
const refresh = async () => { await fetchRegistries() }
const handleSizeChange = (size) => { pagination.pageSize = size; fetchRegistries() }
const handleCurrentChange = (page) => { pagination.page = page; fetchRegistries() }

const handleAdd = () => { dialogType.value='add'; dialogVisible.value=true }
const handleEdit = (row) => {
  dialogType.value='edit'; dialogVisible.value=true
  Object.assign(form, { id: row.id, name: row.name, url: row.url, description: row.description, project_id: row.project_id, artifact_type: row.artifact_type, artifact_source: row.artifact_source, is_public: row.is_public, username: row.username, password: '' })
}
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除注册表 "${row.name}"?`, '提示', { type: 'warning' })
    .then(async ()=>{ await deleteRegistry(row.id); ElMessage.success('删除成功'); fetchRegistries() })
    .catch(()=>{})
}
const handleSubmit = () => {
  formRef.value.validate(async (valid)=>{
    if (!valid) return
    submitting.value = true
    try {
      if (dialogType.value==='add') {
        const payload = { name: form.name, url: form.url, description: form.description, is_public: form.is_public, username: form.username, password: form.password, project_id: form.project_id, artifact_type: form.artifact_type, artifact_source: form.artifact_source }
        await createRegistry(payload); ElMessage.success('新增成功')
      } else {
        const { id, ...payload } = form; await updateRegistry(id, payload); ElMessage.success('编辑成功')
      }
      dialogVisible.value=false; fetchRegistries()
    } catch(e){ console.error('提交失败:',e); ElMessage.error('提交失败') } finally { submitting.value=false }
  })
}
const handleDialogClose = ()=>{ formRef.value?.resetFields(); Object.assign(form, { id:'', name:'', url:'', description:'', project_id:'', artifact_type: ARTIFACT_TYPE.DOCKER, artifact_source: ARTIFACT_SOURCE.XCODING_REGISTRY, is_public:false, username:'', password:'' }) }

// 新建制品库（命名空间）
const namespaceDialogVisible = ref(false)
const namespaceFormRef = ref(null)
const submittingNamespace = ref(false)
const namespaceForm = reactive({ name: '', description: '', registry_id: '' })
const namespaceFormRules = {
  name: [{ required: true, message: '请输入制品库名称', trigger: 'blur' }],
  registry_id: [{ required: true, message: '请选择注册表', trigger: 'change' }]
}
const handleAddNamespace = ()=>{ namespaceDialogVisible.value = true }
const handleNamespaceDialogClose = ()=>{ namespaceFormRef.value?.resetFields(); Object.assign(namespaceForm, { name:'', description:'', registry_id:'' }) }
const handleSubmitNamespace = ()=>{
  namespaceFormRef.value.validate(async (valid)=>{
    if (!valid) return
    submittingNamespace.value = true
    try {
      await createNamespace({ name: namespaceForm.name, description: namespaceForm.description, registry_id: namespaceForm.registry_id })
      ElMessage.success('制品库创建成功'); namespaceDialogVisible.value=false
    } catch(e){ console.error('创建制品库失败:', e); ElMessage.error('创建制品库失败') } finally { submittingNamespace.value=false }
  })
}

onMounted(async () => { await projectStore.loadPersisted(); await fetchProjectOptions(); fetchRegistries() })
</script>

<style scoped>
.project-section-layout { min-height: calc(100vh - 60px); }
.project-section-main { padding: 0; }
.registries-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; gap:12px; }
.toolbar { display:flex; align-items:center; gap:8px; flex-wrap: wrap; }
.toolbar-input { width: 220px; max-width: 220px; }
.toolbar-select { width: 220px; }
.toolbar-select-narrow { width: 180px; }
.toolbar-btn :deep(.el-icon) { margin-right: 4px; }
.pagination-container { margin-top: 20px; display: flex; justify-content: flex-end; }
</style>