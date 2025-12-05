<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isSidebarOpen ? '240px' : '64px'" :class="['sidebar', { collapsed: !isSidebarOpen }]">
      <div class="logo-container">
        <h1 v-show="isSidebarOpen">XCoding</h1>
        <el-button type="text" class="collapse-btn" @click="toggleSidebar">
          <el-icon><Fold v-if="isSidebarOpen" /><Expand v-else /></el-icon>
        </el-button>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        background-color="var(--sidebar-bg)"
        text-color="var(--sidebar-text)"
        active-text-color="var(--sidebar-active-text)"
        :collapse="!isSidebarOpen"
        router
      >
        <!-- 一级：工作台 -->
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>工作台</span>
        </el-menu-item>

        <!-- 一级：项目（点击后进入项目二级布局） -->
        <el-menu-item index="/projects">
          <el-icon><FolderOpened /></el-icon>
          <span>项目</span>
        </el-menu-item>

      </el-menu>
      <div class="sidebar-footer">
        <UserAvatar />
      </div>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="header">
        <div class="header-left">
          <breadcrumb />
        </div>
        <div class="header-right"></div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useProjectStore } from '@/stores/project'
import Breadcrumb from '@/components/Breadcrumb.vue'
import UserAvatar from '@/components/UserAvatar.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const projectStore = useProjectStore()

// 侧边栏状态
const isSidebarOpen = ref(true)

// 当前激活的菜单项
const activeMenu = computed(() => route.path)

// 判断用户是否为超级管理员
const isSuperAdmin = computed(() => {
  return userStore.userInfo?.role === 'USER_ROLE_SUPER_ADMIN'
})

// 切换侧边栏
const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value
}

// 根据路由自动折叠：进入项目二级页面（/projects/... 且不等于 /projects）时折叠一级菜单
const shouldCollapseByRoute = computed(() => {
  const p = route.path || ''
  return p.startsWith('/projects/') && p !== '/projects'
})

// 仅在进入项目子页时自动折叠；其他路由保持用户当前状态
watch(() => route.path, () => {
  if (shouldCollapseByRoute.value) {
    isSidebarOpen.value = false
  }
}, { immediate: true })

// 组件挂载时获取用户信息
onMounted(async () => {
  if (userStore.isAuthenticated) {
    try {
      await userStore.fetchUserInfo()
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }
  // 恢复已选项目，供全局使用
  try { await projectStore.loadPersisted() } catch (_) {}
  // 进入 /projects 直接展示列表，无需弹窗逻辑
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

/* 让内层 el-container 承接满高，便于 main 内容区计算剩余高度 */
.layout-container > .el-container { height: 100%; }

.sidebar {
  background-color: var(--sidebar-bg);
  border-right: 1px solid var(--sidebar-border);
  transition: width 0.3s;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--sidebar-active-text);
  font-size: 18px;
  font-weight: bold;
  padding: 0 20px;
  /* border-bottom: 1px solid var(--sidebar-border); */
}

.sidebar-menu {
  border-right: none;
  flex: 1 1 auto;
}

.collapse-btn {
  color: var(--sidebar-text);
}

.sidebar.collapsed .logo-container {
  justify-content: center;
  padding: 0;
}

.header {
  background-color: var(--header-bg);
  color: var(--header-text);
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--sidebar-border);
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.main-content {
  background-color: var(--main-bg);
  padding: 20px;
  overflow-y: auto;
  /* Header 默认高度约 60px，这里将内容区设为剩余高度，并启用 flex 以让子页填满 */
  height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
}
/* 让当前路由渲染的根元素自动填满剩余高度，承接子页的满高布局 */
.main-content > * { flex: 1 1 auto; min-height: 0; display: flex; flex-direction: column; }
</style>
.sidebar-footer {
  border-top: 1px solid #1f2d3d;
  padding: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sidebar.collapsed .sidebar-footer {
  padding: 8px;
}