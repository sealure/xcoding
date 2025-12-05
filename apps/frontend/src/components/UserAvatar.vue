<template>
  <div class="user-avatar-container">
    <el-dropdown @command="handleCommand">
      <div class="avatar-wrapper">
        <el-avatar :size="40" :src="userStore.userInfo?.avatar">
          {{ (userStore.userInfo?.name || userStore.userInfo?.username)?.charAt(0).toUpperCase() || 'U' }}
        </el-avatar>
        <span class="username">{{ userStore.userInfo?.name || userStore.userInfo?.username || '未登录' }}</span>
        <el-icon class="el-icon--right">
          <arrow-down />
        </el-icon>
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="profile">
            <el-icon><User /></el-icon>个人中心
          </el-dropdown-item>
          <el-dropdown-item command="apitoken">
            <el-icon><Key /></el-icon>ApiToken
          </el-dropdown-item>
          <el-dropdown-item command="settings">
            <el-icon><Setting /></el-icon>系统设置
          </el-dropdown-item>
          <el-dropdown-item divided command="logout">
            <el-icon><SwitchButton /></el-icon>退出登录
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>

  <!-- 放置在组件末尾以便通过 ref 控制弹窗 -->
  <!-- 系统设置入口已跳转到 /settings，此处不再挂载弹窗组件 -->
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { ref } from 'vue'

const router = useRouter()
const userStore = useUserStore()

// 处理下拉菜单命令
const handleCommand = (command) => {
  switch (command) {
    case 'profile':
      // 跳转到个人中心
      router.push('/profile')
      break
    case 'apitoken':
      // 跳转到 ApiToken 管理
      router.push('/apitoken')
      break
    case 'settings':
      // 跳转到系统设置页（包含主题选项卡）
      router.push('/settings')
      break
    case 'logout':
      // 退出登录
      handleLogout()
      break
  }
}

// 处理退出登录
const handleLogout = () => {
  ElMessageBox.confirm(
    '确定要退出登录吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    await userStore.logout()
    router.push('/login')
  }).catch(() => {})
}
</script>

<style scoped>
.user-avatar-container {
  display: flex;
  align-items: center;
}

.avatar-wrapper {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.avatar-wrapper:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.username {
  margin: 0 8px;
  font-size: 14px;
  color: #606266;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>