<template>
  <div class="project-detail-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="title-area">
            <span>项目详情</span>
            <el-tag v-if="projectStore.selectedProject" type="info" size="small">{{ projectStore.selectedProject.name }}</el-tag>
          </div>
          <div class="actions">
            <el-button type="text" @click="go('/projects')">返回项目列表</el-button>
          </div>
        </div>
      </template>

      <div v-if="!projectStore.selectedProject" class="empty">
        <el-empty description="请先在项目列表选择项目" />
      </div>

      <div v-else class="entry-grid">
        <div class="entry-item" @click="go('/projects/overview')">
          <div class="entry-title">项目概览</div>
          <div class="entry-desc">项目统计与功能入口</div>
        </div>

        <div class="entry-item" @click="go('/projects/repositories')">
          <div class="entry-title">代码仓库</div>
          <div class="entry-desc">管理仓库、提交与分支</div>
        </div>

        <div class="entry-item" @click="go('/projects/artifact/registries')">
          <div class="entry-title">制品库</div>
          <div class="entry-desc">注册表、命名空间与仓库</div>
        </div>

        <div class="entry-item" @click="go('/projects/users')">
          <div class="entry-title">用户管理</div>
          <div class="entry-desc">项目成员与权限</div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'

const router = useRouter()
const projectStore = useProjectStore()

const go = (path) => { router.push(path) }
</script>

<style scoped>
.project-detail-container { padding: 20px; }
.card-header { display:flex; justify-content:space-between; align-items:center; }
.title-area { display:flex; align-items:center; gap:8px; }
.empty { padding: 20px 0; }
.entry-grid { display:grid; grid-template-columns: repeat(4, 1fr); gap:16px; }
.entry-item { background:#fff; border:1px solid #ebeef5; border-radius:6px; padding:18px; cursor:pointer; transition: box-shadow .2s ease; }
.entry-item:hover { box-shadow: 0 2px 12px rgba(0,0,0,.08); }
.entry-title { font-weight:600; font-size:16px; color:#303133; }
.entry-desc { margin-top:8px; color:#909399; font-size:12px; }
@media (max-width: 1280px) { .entry-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) { .entry-grid { grid-template-columns: 1fr; } }
</style>