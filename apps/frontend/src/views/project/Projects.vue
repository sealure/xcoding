<template>
  <el-container class="project-section-layout">
    <el-main class="project-section-main">
      <ProjectTabs />
      <div class="projects-overview-container">
        <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="title-area">
            <span>项目概览</span>
            <el-tag v-if="projectStore.selectedProject" type="info" size="small">{{ projectStore.selectedProject.name }}</el-tag>
          </div>
          <el-button type="primary" size="small" @click="go('/projects')">选择项目</el-button>
        </div>
      </template>

      <div v-if="!projectStore.selectedProject" class="empty-overview">
        <el-empty description="请选择项目以查看概览" />
      </div>

      <div v-else class="overview-grid">
        <div class="overview-item" @click="go('/projects/repositories')">
          <div class="overview-item-main">
            <div class="overview-item-title">代码仓库</div>
            <div class="overview-item-count">{{ metrics.repositoriesCount }}</div>
          </div>
          <div class="overview-item-desc">管理项目的代码仓库、提交与分支</div>
        </div>

        <div class="overview-item" @click="go('/projects/users')">
          <div class="overview-item-main">
            <div class="overview-item-title">项目用户</div>
            <div class="overview-item-count">—</div>
          </div>
          <div class="overview-item-desc">查看并管理项目成员与权限</div>
        </div>

        <div class="overview-item" @click="go('/projects/artifact/registries')">
          <div class="overview-item-main">
            <div class="overview-item-title">制品注册表</div>
            <div class="overview-item-count">{{ metrics.registriesCount }}</div>
          </div>
          <div class="overview-item-desc">配置 Docker 等制品源</div>
        </div>

        <div class="overview-item" @click="go('/projects/artifact/namespaces')">
          <div class="overview-item-main">
            <div class="overview-item-title">命名空间</div>
            <div class="overview-item-count">—</div>
          </div>
          <div class="overview-item-desc">按命名空间组织制品</div>
        </div>

        <div class="overview-item" @click="go('/projects/artifact/repositories')">
          <div class="overview-item-main">
            <div class="overview-item-title">制品仓库</div>
            <div class="overview-item-count">—</div>
          </div>
          <div class="overview-item-desc">管理制品仓库与路径</div>
        </div>

        <div class="overview-item" @click="go('/projects/artifact/tags')">
          <div class="overview-item-main">
            <div class="overview-item-title">制品标签</div>
            <div class="overview-item-count">—</div>
          </div>
          <div class="overview-item-desc">查看和维护制品标签</div>
        </div>
      </div>
        </el-card>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { reactive, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { getRepositoryList } from '@/api/code_repository/repository'
import { listRegistries } from '@/api/artifact/registry'
import ProjectTabs from '@/components/ProjectTabs.vue'

const router = useRouter()
const projectStore = useProjectStore()

const metrics = reactive({ repositoriesCount: 0, registriesCount: 0 })

const fetchCounts = async () => {
  const pid = projectStore.selectedProject?.id
  if (!pid) { metrics.repositoriesCount = 0; metrics.registriesCount = 0; return }
  try {
    const [reposRes, regsRes] = await Promise.all([
      getRepositoryList(pid, { page: 1, page_size: 1 }),
      listRegistries({ page: 1, page_size: 1, project_id: pid })
    ])
    metrics.repositoriesCount = reposRes?.pagination?.total_items || 0
    metrics.registriesCount = regsRes?.pagination?.total_items || 0
  } catch (_) {
    metrics.repositoriesCount = 0
    metrics.registriesCount = 0
  }
}

const go = (path) => { router.push(path) }

watch(() => projectStore.selectedProject?.id, () => { fetchCounts() })
onMounted(() => { fetchCounts() })
</script>

<style scoped>
.project-section-layout { min-height: calc(100vh - 60px); }
.project-section-main { padding: 0; }
.projects-overview-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.title-area { display:flex; align-items:center; gap:8px; }
.empty-overview { padding: 20px 0; }
.overview-grid { display:grid; grid-template-columns: repeat(3, 1fr); gap:16px; }
.overview-item { background:#fff; border:1px solid #ebeef5; border-radius:6px; padding:16px; cursor:pointer; transition: box-shadow .2s ease; }
.overview-item:hover { box-shadow: 0 2px 12px rgba(0,0,0,.08); }
.overview-item-main { display:flex; justify-content:space-between; align-items:center; }
.overview-item-title { font-weight:600; font-size:16px; }
.overview-item-count { font-size:20px; color:#409EFF; }
.overview-item-desc { margin-top:8px; color:#909399; font-size:12px; }
</style>