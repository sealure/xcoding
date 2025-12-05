<template>
  <div class="dashboard-container">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>用户总数</span>
              <el-icon><User /></el-icon>
            </div>
          </template>
          <div class="card-value">{{ statistics.userCount }}</div>
          <div class="card-footer">
            <span>较昨日</span>
            <span :class="statistics.userGrowth >= 0 ? 'text-success' : 'text-danger'">
              {{ statistics.userGrowth >= 0 ? '+' : '' }}{{ statistics.userGrowth }}%
            </span>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>项目总数</span>
              <el-icon><FolderOpened /></el-icon>
            </div>
          </template>
          <div class="card-value">{{ statistics.projectCount }}</div>
          <div class="card-footer">
            <span>较昨日</span>
            <span :class="statistics.projectGrowth >= 0 ? 'text-success' : 'text-danger'">
              {{ statistics.projectGrowth >= 0 ? '+' : '' }}{{ statistics.projectGrowth }}%
            </span>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>代码仓库总数</span>
              <el-icon><Document /></el-icon>
            </div>
          </template>
          <div class="card-value">{{ statistics.repositoryCount }}</div>
          <div class="card-footer">
            <span>较昨日</span>
            <span :class="statistics.repositoryGrowth >= 0 ? 'text-success' : 'text-danger'">
              {{ statistics.repositoryGrowth >= 0 ? '+' : '' }}{{ statistics.repositoryGrowth }}%
            </span>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="dashboard-card">
          <template #header>
            <div class="card-header">
              <span>活跃用户</span>
              <el-icon><UserFilled /></el-icon>
            </div>
          </template>
          <div class="card-value">{{ statistics.activeUserCount }}</div>
          <div class="card-footer">
            <span>较昨日</span>
            <span :class="statistics.activeUserGrowth >= 0 ? 'text-success' : 'text-danger'">
              {{ statistics.activeUserGrowth >= 0 ? '+' : '' }}{{ statistics.activeUserGrowth }}%
            </span>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="mt-20">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近创建的项目</span>
              <el-button type="text" @click="$router.push('/projects')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentProjects" style="width: 100%" v-loading="loading">
            <el-table-column prop="name" label="项目名称" />
            <el-table-column prop="creator" label="创建者" />
            <el-table-column prop="createTime" label="创建时间" width="180" />
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近创建的代码仓库</span>
              <el-button type="text" @click="$router.push('/projects/repositories')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentRepositories" style="width: 100%" v-loading="loading">
            <el-table-column prop="name" label="仓库名称" />
            <el-table-column prop="projectName" label="所属项目" />
            <el-table-column prop="createTime" label="创建时间" width="180" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { getProjectList } from '@/api/project'
import { getRepositoryList } from '@/api/code_repository/repository'
import { getUserList } from '@/api/user'
import { useProjectStore } from '@/stores/project'

// 加载状态
const loading = ref(false)

// 统计数据
const statistics = reactive({
  userCount: 0,
  userGrowth: 0,
  projectCount: 0,
  projectGrowth: 0,
  repositoryCount: 0,
  repositoryGrowth: 0,
  activeUserCount: 0,
  activeUserGrowth: 0
})

// 最近的项目
const recentProjects = ref([])

// 最近的代码仓库
const recentRepositories = ref([])

// 项目存储（用于选择项目后再拉取仓库相关统计）
const projectStore = useProjectStore()

// 获取统计数据
const fetchStatistics = async () => {
  loading.value = true
  try {
    // 获取用户总数
    const userRes = await getUserList({ page: 1, page_size: 1 })
    statistics.userCount = userRes.pagination?.total_items || 0
    statistics.userGrowth = Math.floor(Math.random() * 20) - 5 // 模拟增长率
    
    // 获取项目总数
    const projectRes = await getProjectList({ page: 1, page_size: 1 })
    statistics.projectCount = projectRes.pagination?.total_items || 0
    statistics.projectGrowth = Math.floor(Math.random() * 20) - 5 // 模拟增长率
    
    // 获取代码仓库总数（需选定项目，否则跳过以避免 400）
    const pid = projectStore.selectedProject?.id
    if (pid) {
      const repositoryRes = await getRepositoryList(pid, { page: 1, page_size: 1 })
      statistics.repositoryCount = repositoryRes.pagination?.total_items || 0
    } else {
      statistics.repositoryCount = 0
    }
    statistics.repositoryGrowth = Math.floor(Math.random() * 20) - 5 // 模拟增长率
    
    // 模拟活跃用户数
    statistics.activeUserCount = Math.floor(statistics.userCount * 0.7)
    statistics.activeUserGrowth = Math.floor(Math.random() * 20) - 5 // 模拟增长率
    
    // 获取最近的项目
    const recentProjectRes = await getProjectList({ page: 1, page_size: 5 })
    recentProjects.value = recentProjectRes.data || []
    
    // 获取最近的代码仓库（需选定项目，否则为空列表）
    if (pid) {
      const recentRepositoryRes = await getRepositoryList(pid, { page: 1, page_size: 5 })
      recentRepositories.value = recentRepositoryRes.data || []
    } else {
      recentRepositories.value = []
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取数据
onMounted(async () => {
  try { await projectStore.loadPersisted() } catch (_) {}
  fetchStatistics()
})

// 监听项目切换，重新拉取统计数据
watch(() => projectStore.selectedProject?.id, () => { fetchStatistics() })
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

.mt-20 {
  margin-top: 20px;
}

.dashboard-card {
  height: 180px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-value {
  font-size: 28px;
  font-weight: bold;
  margin: 20px 0;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #909399;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>