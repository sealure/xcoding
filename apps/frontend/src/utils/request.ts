import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

// 创建axios实例
const service: AxiosInstance = axios.create({
  baseURL: '/', // 改为后端真实前缀
  timeout: 15000 // 请求超时时间
})

// 请求拦截器
service.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // 从localStorage获取token
    const token = localStorage.getItem('token')
    if (token) {
      // 设置请求头中的Authorization
      ;(config.headers = config.headers || {})
      ;(config.headers as Record<string, string>)['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    const res = response.data as any
    
    // 如果返回的状态码不是0，说明接口出错了
    if (res.code !== undefined && res.code !== 0) {
      // 处理错误情况
      return Promise.reject(new Error(res.message || '请求失败'))
    } else {
      // 直接返回原始响应体，便于按 { data, pagination, ... } 结构访问
      return res
    }
  },
  error => {
    console.error('响应错误:', error)
    
    // 处理401未授权错误
    if (error.response && error.response.status === 401) {
      // 清除token并跳转到登录页
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    
    return Promise.reject(error)
  }
)

export default service