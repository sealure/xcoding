import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface RegistryPayload {
  name?: string
  description?: string
  url?: string
  [key: string]: any
}

// 列出注册表
export function listRegistries(params: PaginationParams = {}) {
  return request({
    url: 'artifact_service/api/v1/registries',
    method: 'get',
    params
  })
}

// 创建注册表
export function createRegistry(data: RegistryPayload) {
  return request({
    url: 'artifact_service/api/v1/registries',
    method: 'post',
    data
  })
}

// 获取注册表详情
export function getRegistry(id: string | number) {
  return request({
    url: `artifact_service/api/v1/registries/${id}`,
    method: 'get'
  })
}

// 更新注册表
export function updateRegistry(id: string | number, data: RegistryPayload) {
  return request({
    url: `artifact_service/api/v1/registries/${id}`,
    method: 'put',
    data
  })
}

// 删除注册表
export function deleteRegistry(id: string | number) {
  return request({
    url: `artifact_service/api/v1/registries/${id}`,
    method: 'delete'
  })
}