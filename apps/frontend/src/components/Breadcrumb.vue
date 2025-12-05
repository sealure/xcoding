<template>
  <el-breadcrumb separator="/">
    <!-- 单次循环保持顺序：项目 / 项目名 / 当前页 -->
    <el-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index" :to="item.path">
      <!-- 项目段：支持悬浮选择项目 -->
      <template v-if="item.type === 'project'">
        <el-popover placement="bottom-start" :width="isSidebarOpen ? 480 : 360" trigger="hover">
          <template #reference>
            <span class="project-switch-ref">
              {{ item.title || '未选择项目' }}
              <el-icon class="caret"><ArrowDown /></el-icon>
            </span>
          </template>
          <div class="project-switch-panel">
            <el-input v-model="projectKeyword" clearable placeholder="搜索项目" />
            <el-scrollbar height="320px" class="project-list">
              <div v-for="p in filteredProjects" :key="p.id" class="project-item" @click="selectProject(p)">
                <span class="name">{{ p.name }}</span>
              </div>
              <div v-if="!filteredProjects.length" class="empty">暂无匹配项目</div>
            </el-scrollbar>
          </div>
        </el-popover>
      </template>

      <!-- 其他段：普通文本 -->
      <template v-else>
        {{ item.title }}
      </template>
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useProjectStore } from '@/stores/project'
import { ArrowDown } from '@element-plus/icons-vue'
import { useRoute as useRouterRoute } from 'vue-router'
import { useRouter } from 'vue-router'
import { useProjectStore as useStoreAgain } from '@/stores/project'

const route = useRoute()
const projectStore = useProjectStore()
const router = useRouter()
const projectKeyword = ref('')
const isSidebarOpen = ref(true) // 控制弹层宽度的简易响应，保持一致视觉

// 根据路由和所选项目生成面包屑
const breadcrumbs = computed(() => {
  const path = route.path || ''

  // 项目二级页面：项目 / {项目名} / {当前页}
  if (path.startsWith('/projects')) {
    const items = [{ title: '项目', path: '/projects', type: 'root' }]
    const projectName = projectStore.selectedProject?.name
    items.push({ title: projectName || '未选择项目', path: '/projects/overview', type: 'project' })
    const sectionMap = {
      '/projects/overview': '概览',
      '/projects/repositories': '代码仓库',
      '/projects/users': '用户管理',
      '/projects/artifact/registries': '制品注册表',
      '/projects/artifact/namespaces': '命名空间',
      '/projects/artifact/repositories': '制品仓库',
      '/projects/artifact/tags': '制品标签'
    }
    const section = sectionMap[path]
    if (section) {
      items.push({ title: section, path, type: 'section' })
    }
    return items
  }

  // CI 页面采用与项目子页一致的面包屑样式：项目 / {项目名} / 持续集成
  if (path.startsWith('/ci/')) {
    const items = [{ title: '项目', path: '/projects', type: 'root' }]
    const projectName = projectStore.selectedProject?.name
    items.push({ title: projectName || '未选择项目', path: '/projects/overview', type: 'project' })
    items.push({ title: '持续集成', path: '/ci/pipeline', type: 'section' })
    return items
  }

  // 其他页面的简单映射
  const miscMap = {
    '/dashboard': '工作台'
  }
  const title = miscMap[path]
  return title ? [{ title, path, type: 'root' }] : []
})

// 项目筛选
const filteredProjects = computed(() => {
  const kw = projectKeyword.value.trim().toLowerCase()
  const list = projectStore.projectOptions || []
  if (!kw) return list
  return list.filter(p => String(p.name).toLowerCase().includes(kw))
})

const selectProject = (p) => {
  projectStore.setSelectedProject(p)
  // 保持当前页面不变；若在项目入口则进入概览
  const path = route.path || ''
  if (path === '/projects') router.push('/projects/overview')
}

onMounted(async () => {
  if (!projectStore.projectOptions?.length) {
    try { await projectStore.fetchProjectOptions() } catch (_) {}
  }
})
</script>

<style scoped>
.el-breadcrumb {
  font-size: 14px;
  line-height: 50px;
  margin-left: 10px;
}

.project-switch-ref {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}
.project-switch-ref .caret { margin-left: 6px; font-size: 14px; }
.project-switch-panel { padding: 8px; width: 100%; }
.project-list { margin-top: 8px; }
.project-item { padding: 8px 6px; cursor: pointer; border-radius: 4px; }
.project-item:hover { background: rgba(64, 158, 255, 0.12); }
.project-item .name { color: #303133; }
.empty { color: #909399; padding: 12px; text-align: center; }
</style>