<template>
  <div class="apitoken-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>API Token 管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>创建 Token
          </el-button>
        </div>
      </template>
      
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>
      
      <div v-else>
        <!-- 筛选工具栏 -->
        <div class="toolbar">
          <el-input
            v-model="searchForm.name"
            placeholder="按名称搜索"
            clearable
            class="toolbar-input"
            @keyup.enter="handleSearch"
          />
          <el-select
            v-model="searchForm.scope"
            placeholder="权限(全部)"
            clearable
            class="toolbar-select"
            @change="handleSearch"
          >
            <el-option
              v-for="opt in SCOPE_OPTIONS"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <el-select
            v-model="searchForm.status"
            placeholder="状态(全部)"
            clearable
            class="toolbar-select-narrow"
            @change="handleSearch"
          >
            <el-option label="有效" value="valid" />
            <el-option label="已过期" value="expired" />
          </el-select>
          <el-date-picker
            v-model="searchForm.createdRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            class="toolbar-date"
            :shortcuts="dateShortcuts"
            @change="handleSearch"
          />
          <el-button type="primary" @click="handleSearch" class="toolbar-btn">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button @click="resetSearch" class="toolbar-btn">
            <el-icon><Refresh /></el-icon>重置
          </el-button>
        </div>

        <!-- Token 列表 -->
        <el-table :data="filteredTokens" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="150">
            <template #default="{ row }">
              <div class="token-name">
                <strong>{{ row.name }}</strong>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="description" label="描述" min-width="220">
            <template #default="{ row }">
              {{ row.description || '—' }}
            </template>
          </el-table-column>
          
          <el-table-column label="权限范围" min-width="200">
            <template #default="{ row }">
              <div class="scopes-container">
                <el-tag
                  v-for="scope in row.scopes"
                  :key="scope"
                  size="small"
                  class="scope-tag"
                >
                  {{ getScopeLabel(scope) }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="created_at" label="创建时间" width="180" sortable :sort-method="sortCreatedAt">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          
          <el-table-column prop="expires_at" label="过期时间" width="180" sortable :sort-method="sortExpiresAt">
            <template #default="{ row }">
              <span :class="{ 'expired': isExpired(row.expires_at) }">
                {{ formatExpiresAt(row.expires_at) }}
              </span>
            </template>
          </el-table-column>
          
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="isExpired(row.expires_at) ? 'danger' : 'success'" size="small">
                {{ isExpired(row.expires_at) ? '已过期' : '有效' }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button
                type="danger"
                size="small"
                @click="handleDelete(row)"
                :loading="deletingIds.includes(row.id)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <!-- 空状态 -->
        <el-empty v-if="!loading && tokens.length === 0" description="暂无 API Token">
          <el-button type="primary" @click="showCreateDialog = true">创建第一个 Token</el-button>
        </el-empty>
      </div>
    </el-card>

    <!-- 创建 Token 对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建 API Token" width="600px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item label="Token 名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入 Token 名称" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入 Token 描述（可选）"
          />
        </el-form-item>
        
        <el-form-item label="过期时间" prop="expires_in">
          <el-select v-model="createForm.expires_in" placeholder="请选择过期时间" style="width: 100%">
            <el-option
              v-for="option in TOKEN_EXPIRATION_OPTIONS"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="权限范围" prop="scopes">
          <el-checkbox-group v-model="createForm.scopes">
            <div class="scopes-grid">
              <el-checkbox
                v-for="option in SCOPE_OPTIONS"
                :key="option.value"
                :label="option.value"
                class="scope-checkbox"
              >
                <div class="scope-option">
                  <div class="scope-label">{{ option.label }}</div>
                  <div class="scope-description">{{ option.description }}</div>
                </div>
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :loading="creating">
            创建
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Token 创建成功对话框 -->
    <el-dialog v-model="showTokenDialog" title="Token 创建成功" width="500px" :close-on-click-modal="false">
      <div class="token-success">
        <el-alert
          title="请妥善保存您的 Token"
          description="Token 只会显示一次，请立即复制并保存到安全的地方。"
          type="warning"
          :closable="false"
          show-icon
        />
        
        <div class="token-display">
          <el-input
            v-model="newToken"
            readonly
            type="textarea"
            :rows="3"
            class="token-input"
          />
          <el-button type="primary" @click="copyToken" class="copy-button">
            <el-icon><DocumentCopy /></el-icon>复制 Token
          </el-button>
        </div>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button type="primary" @click="showTokenDialog = false">我已保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, DocumentCopy, Search, Refresh } from '@element-plus/icons-vue'
import {
  createAPIToken,
  listAPITokens,
  deleteAPIToken,
  TOKEN_EXPIRATION_OPTIONS,
  SCOPE_OPTIONS,
  formatScopes
} from '@/api/user'

// 响应式数据
const loading = ref(false)
const creating = ref(false)
const deletingIds = ref([])
const tokens = ref([])
const showCreateDialog = ref(false)
const showTokenDialog = ref(false)
const newToken = ref('')

// 表单引用
const createFormRef = ref()

// 创建表单
const createForm = reactive({
  name: '',
  description: '',
  expires_in: 4, // 默认30天
  scopes: [SCOPE_OPTIONS.find(opt => opt.value)?.value || 'SCOPE_READ']
})

// 表单验证规则
const createRules = {
  name: [
    { required: true, message: '请输入 Token 名称', trigger: 'blur' },
    { min: 3, max: 50, message: 'Token 名称长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  expires_in: [
    { required: true, message: '请选择过期时间', trigger: 'change' }
  ],
  scopes: [
    { type: 'array', min: 1, message: '请至少选择一个权限范围', trigger: 'change' }
  ]
}

// 获取权限范围标签
const getScopeLabel = (scope) => {
  const option = SCOPE_OPTIONS.find(opt => opt.value === scope)
  return option ? option.label : `未知权限(${scope})`
}

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 格式化过期时间
const formatExpiresAt = (expiresAt) => {
  if (!expiresAt) return '永不过期'
  return new Date(expiresAt).toLocaleString('zh-CN')
}

// 检查是否过期
const isExpired = (expiresAt) => {
  if (!expiresAt) return false
  return new Date(expiresAt) < new Date()
}

// 获取 Token 列表
const fetchTokens = async () => {
  loading.value = true
  try {
    const res = await listAPITokens()
    // 兼容多种返回结构：{ tokens: [...] } 或 { data: { tokens: [...] } }
    tokens.value = (res && res.tokens) || (res && res.data && res.data.tokens) || []
  } catch (error) {
    console.error('获取 Token 列表失败:', error)
    ElMessage.error('获取 Token 列表失败')
  } finally {
    loading.value = false
  }
}

// 创建 Token
const handleCreate = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    creating.value = true
    
    const res = await createAPIToken(createForm)
    // 创建接口通常直接返回 { token: '...' }；同时兼容 { data: { token: '...' } }
    newToken.value = (res && res.token) || (res && res.data && res.data.token) || ''
    
    ElMessage.success('Token 创建成功')
    showCreateDialog.value = false
    showTokenDialog.value = true
    
    // 重置表单
    createForm.name = ''
    createForm.description = ''
    createForm.expires_in = 4
    createForm.scopes = [SCOPE_OPTIONS.find(opt => opt.value)?.value || 'SCOPE_READ']
    
    // 刷新列表
    await fetchTokens()
  } catch (error) {
    console.error('创建 Token 失败:', error)
    ElMessage.error('创建 Token 失败')
  } finally {
    creating.value = false
  }
}

// 删除 Token
const handleDelete = (token) => {
  ElMessageBox.confirm(
    `确定要删除 Token "${token.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    deletingIds.value.push(token.id)
    try {
      await deleteAPIToken(token.id)
      ElMessage.success('Token 删除成功')
      await fetchTokens()
    } catch (error) {
      console.error('删除 Token 失败:', error)
      ElMessage.error('删除 Token 失败')
    } finally {
      deletingIds.value = deletingIds.value.filter(id => id !== token.id)
    }
  })
}

// 复制 Token
const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(newToken.value)
    ElMessage.success('Token 已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}

// 搜索与过滤
const searchForm = reactive({ name: '', scope: '', status: '', createdRange: null })

// 日期快捷选项
const dateShortcuts = [
  {
    text: '今天',
    value: () => {
      const now = new Date()
      return [now, now]
    }
  },
  {
    text: '近7天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 6)
      return [start, end]
    }
  },
  {
    text: '近30天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 29)
      return [start, end]
    }
  },
  {
    text: '近90天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 89)
      return [start, end]
    }
  },
  {
    text: '近半年',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 179)
      return [start, end]
    }
  },
  {
    text: '近一年',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setDate(start.getDate() - 364)
      return [start, end]
    }
  }
]

const filteredTokens = computed(() => {
  let list = tokens.value || []
  const kw = String(searchForm.name || '').trim().toLowerCase()
  if (kw) {
    list = list.filter(t => String(t.name || '').toLowerCase().includes(kw))
  }
  if (searchForm.scope) {
    const scopeVal = searchForm.scope
    list = list.filter(t => Array.isArray(t.scopes) && t.scopes.includes(scopeVal))
  }
  if (searchForm.status) {
    if (searchForm.status === 'valid') {
      list = list.filter(t => !isExpired(t.expires_at))
    } else if (searchForm.status === 'expired') {
      list = list.filter(t => isExpired(t.expires_at))
    }
  }
  if (searchForm.createdRange && Array.isArray(searchForm.createdRange) && searchForm.createdRange.length === 2) {
    const [start, end] = searchForm.createdRange
    const startDate = start ? new Date(`${start}T00:00:00`) : null
    const endDate = end ? new Date(`${end}T23:59:59`) : null
    list = list.filter(t => {
      const created = t.created_at ? new Date(t.created_at) : null
      if (!created) return false
      if (startDate && created < startDate) return false
      if (endDate && created > endDate) return false
      return true
    })
  }
  return list
})

// 表格排序方法
const sortCreatedAt = (a, b) => {
  const ta = a?.created_at ? new Date(a.created_at).getTime() : 0
  const tb = b?.created_at ? new Date(b.created_at).getTime() : 0
  return ta - tb
}

const sortExpiresAt = (a, b) => {
  const ta = a?.expires_at ? new Date(a.expires_at).getTime() : Number.POSITIVE_INFINITY
  const tb = b?.expires_at ? new Date(b.expires_at).getTime() : Number.POSITIVE_INFINITY
  return ta - tb
}

const handleSearch = () => {
  // 由于使用 computed 过滤，这里无需额外操作
}

const resetSearch = () => {
  searchForm.name = ''
  searchForm.scope = ''
  searchForm.status = ''
  searchForm.createdRange = null
}

// 组件挂载时获取数据
onMounted(() => {
  fetchTokens()
})
</script>

<style scoped>
.apitoken-container {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: bold;
}

.loading-container {
  padding: 20px;
}

.toolbar {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 12px;
}

.toolbar-input {
  width: 280px;
}

.toolbar-select {
  width: 220px;
}

.toolbar-select-narrow {
  width: 160px;
}

.toolbar-date {
  width: 320px;
}

.toolbar-btn {
  margin-left: 0;
}

.token-name {
  line-height: 1.4;
}

.token-description {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.scopes-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.scope-tag {
  margin: 2px 0;
}

.expired {
  color: #f56c6c;
}

.scopes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 12px;
  width: 100%;
}

.scope-checkbox {
  margin: 0;
  width: 100%;
}

.scope-option {
  padding: 8px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  transition: all 0.3s;
}

.scope-option:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.scope-label {
  font-weight: 500;
  color: #303133;
}

.scope-description {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.token-success {
  padding: 20px 0;
}

.token-display {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.token-input {
  font-family: 'Courier New', monospace;
}

.copy-button {
  align-self: flex-end;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .apitoken-container {
    margin: 0 10px;
  }
  
  .scopes-grid {
    grid-template-columns: 1fr;
  }
  
  .card-header {
    flex-direction: column;
    gap: 10px;
    align-items: stretch;
  }
}
</style>