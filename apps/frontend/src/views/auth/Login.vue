<template>
  <div class="login-container">
    <div class="login-content">
      <div class="login-header">
        <h2 class="app-title">XCoding</h2>
        <p class="app-subtitle">极简高效的Devops集成平台</p>
      </div>
      
      <el-card class="login-card" shadow="hover">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          size="large"
          @keyup.enter="handleSubmit"
        >
          <el-form-item prop="username">
            <el-input 
              v-model="form.username" 
              placeholder="用户名" 
              prefix-icon="User" 
              clearable
            />
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input 
              v-model="form.password" 
              type="password" 
              placeholder="密码" 
              prefix-icon="Lock" 
              show-password
            />
          </el-form-item>
          
          <el-form-item>
            <el-button 
              type="primary" 
              class="submit-btn" 
              :loading="loading" 
              @click="handleSubmit"
            >
              登 录
            </el-button>
          </el-form-item>

          <div class="form-footer">
            <el-link type="info" :underline="false" @click="$router.push('/register')">注册账号</el-link>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = reactive<FormRules>({
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
})

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await userStore.login(form)
        ElMessage.success('欢迎回来')
        router.push('/dashboard')
      } catch (error: any) {
        ElMessage.error(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
}

.login-content {
  width: 100%;
  max-width: 380px;
  padding: 20px;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.app-title {
  font-size: 28px;
  color: #303133;
  margin: 0 0 10px;
  font-weight: 600;
}

.app-subtitle {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.login-card {
  border-radius: 8px;
  border: none;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
}

.submit-btn {
  width: 100%;
  font-weight: 500;
  letter-spacing: 1px;
}

.form-footer {
  display: flex;
  justify-content: center;
  margin-top: -10px;
}

:deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #dcdfe6 inset;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}
</style>