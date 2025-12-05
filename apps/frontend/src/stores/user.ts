import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as userLogin, register as userRegister, getUserInfo } from '@/api/user'

type LoginCredentials = { username?: string; email?: string; password?: string; [k: string]: any }
type RegisterData = Record<string, any>
type UserInfo = Record<string, any>
type AuthResponse = { data?: { token?: string; authenticated?: boolean; user?: UserInfo }; token?: string; authenticated?: boolean; user?: UserInfo }

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo>({})
  const isAuthenticated = computed<boolean>(() => !!token.value)

  // 登录
  async function login(credentials: LoginCredentials) {
    try {
      const response = await userLogin(credentials) as AuthResponse
      console.log('登录响应:', response)
      
      // 处理不同的响应格式
      const tokenValue = response.data?.token || (response as any).token
      if (!tokenValue) {
        throw new Error('登录响应中没有找到 token')
      }
      
      token.value = tokenValue
      localStorage.setItem('token', token.value)
      console.log('保存的 token:', token.value)
      
      // 登录后拉取用户信息（通过Auth）
      await fetchUserInfo()
      return response
    } catch (error) {
      console.error('登录过程中的错误:', error)
      throw error
    }
  }

  // 注册
  async function register(userData: RegisterData) {
    try {
      const response = await userRegister(userData)
      return response
    } catch (error) {
      throw error
    }
  }

  // 获取用户信息（Auth返回结构：authenticated、user、headers...）
  async function fetchUserInfo() {
    try {
      console.log('开始获取用户信息，当前 token:', token.value)
      const response = await getUserInfo() as AuthResponse
      console.log('获取用户信息响应:', response)
      
      if (response.data?.authenticated && response.data?.user) {
        userInfo.value = response.data.user
      } else if ((response as any).authenticated && (response as any).user) {
        userInfo.value = (response as any).user
      } else {
        console.warn('用户信息响应格式异常:', response)
        userInfo.value = {}
      }
      return response
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，清除 token
      if ((error as any).response?.status === 401) {
        console.log('认证失败，清除 token')
        token.value = ''
        localStorage.removeItem('token')
      }
      throw error
    }
  }

  // 登出
  function logout() {
    token.value = ''
    userInfo.value = {}
    localStorage.removeItem('token')
  }

  return {
    token,
    userInfo,
    isAuthenticated,
    login,
    register,
    fetchUserInfo,
    logout
  }
})