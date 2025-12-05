<template>
  <div class="project-tabs">
    <el-tabs
      v-model="activeTab"
      class="clean-tabs"
      @tab-click="onTabClick"
    >
      <el-tab-pane label="概览" name="overview" />
      <el-tab-pane label="代码仓库" name="repositories" />
      <el-tab-pane label="持续集成" name="ci" />
      <el-tab-pane label="制品注册表" name="registries" />
      <el-tab-pane label="项目设置" name="users" />
    </el-tabs>
  </div>
  
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const pathToName = (p) => {
  if (!p) return 'overview'
  if (p.startsWith('/projects/overview')) return 'overview'
  if (p.startsWith('/projects/repositories')) return 'repositories'
  if (p.startsWith('/ci/builds') || p.startsWith('/ci/pipeline-job') || p.startsWith('/ci/pipeline')) return 'ci'
  if (p.startsWith('/projects/artifact')) return 'registries'
  if (p.startsWith('/projects/users')) return 'users'
  return 'overview'
}

const nameToPath = (n) => {
  switch (n) {
    case 'overview': return '/projects/overview'
    case 'repositories': return '/projects/repositories'
    case 'ci': return '/ci/pipeline'
    case 'registries': return '/projects/artifact/registries'
    case 'users': return '/projects/users'
    default: return '/projects/overview'
  }
}

const activeTab = computed({
  get() {
    return pathToName(route.path || '')
  },
  set(val) {
    const target = nameToPath(val)
    if ((route.path || '') !== target) router.push(target)
  }
})

const onTabClick = (tab) => {
  const name = tab.paneName
  const target = nameToPath(name)
  if ((route.path || '') !== target) router.push(target)
}
</script>

<style scoped>
.project-tabs {
  background-color: var(--header-bg);
  border-bottom: 1px solid var(--el-border-color-light);
  margin-bottom: 20px;
  padding: 0 20px;
}
/* 移除默认的底部灰色条，使用自定义的边框 */
.clean-tabs :deep(.el-tabs__header) { margin: 0; }
.clean-tabs :deep(.el-tabs__nav-wrap::after) { height: 0; }
.clean-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  color: var(--el-text-color-regular);
  height: 48px;
  line-height: 48px;
}
.clean-tabs :deep(.el-tabs__item.is-active) {
  color: var(--el-color-primary);
  font-weight: 600;
}
</style>