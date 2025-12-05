import request from '@/utils/request'

export interface LoginPayload {
  username: string
  password: string
}

export interface RegisterPayload {
  username: string
  password: string
  email?: string
}

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface UpdateUserPayload {
  username?: string
  email?: string
  password?: string
  [key: string]: any
}

// 用户登录
export function login(data: LoginPayload) {
  return request({
    url: 'user_service/api/v1/users/login',
    method: 'post',
    data
  })
}

// 用户注册
export function register(data: RegisterPayload) {
  return request({
    url: 'user_service/api/v1/users/register',
    method: 'post',
    data
  })
}

// 获取当前登录用户信息（通过 Auth 接口）
export function getUserInfo() {
  return request({
    url: 'user_service/api/v1/auth',
    method: 'get'
  })
}

// 根据ID获取用户信息
export function getUserById(id: string | number) {
  return request({
    url: `user_service/api/v1/users/${id}`,
    method: 'get'
  })
}

// 获取用户列表
// 统一分页参数命名：支持 pageSize 自动映射到 page_size
export function getUserList(params: PaginationParams = {}) {
  const query: PaginationParams = { ...params }
  if (query.pageSize && !query.page_size) {
    query.page_size = query.pageSize
    delete query.pageSize
  }
  return request({
    url: 'user_service/api/v1/users',
    method: 'get',
    params: query
  })
}

// 创建用户
export function createUser(data: RegisterPayload) {
  return request({
    url: 'user_service/api/v1/users/register',
    method: 'post',
    data
  })
}

// 更新用户
export function updateUser(id: string | number, data: UpdateUserPayload) {
  return request({
    url: `user_service/api/v1/users/${id}`,
    method: 'put',
    data
  })
}

// 删除用户
export function deleteUser(id: string | number) {
  return request({
    url: `user_service/api/v1/users/${id}`,
    method: 'delete'
  })
}

// 创建API令牌（符合proto：POST /user_service/api/v1/tokens）
export function createApiToken(data: any) {
  return request({
    url: 'user_service/api/v1/tokens',
    method: 'post',
    data
  })
}

// 获取当前用户的API令牌列表（符合proto：GET /user_service/api/v1/tokens）
export function getUserApiTokens() {
  return request({
    url: 'user_service/api/v1/tokens',
    method: 'get'
  })
}

// 删除API令牌（符合proto：DELETE /user_service/api/v1/tokens/{token_id}）
export function deleteApiToken(tokenId: string | number) {
  return request({
    url: `user_service/api/v1/tokens/${tokenId}`,
    method: 'delete'
  })
}