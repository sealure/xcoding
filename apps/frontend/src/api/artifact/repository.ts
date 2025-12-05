import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface RepositoryPayload {
  name?: string
  description?: string
  url?: string
  [key: string]: any
}

// 列出仓库
export function listRepositories(params: PaginationParams = {}) {
  return request({
    url: 'artifact_service/api/v1/repositories',
    method: 'get',
    params
  })
}

// 创建仓库
export function createRepository(data: RepositoryPayload) {
  return request({
    url: 'artifact_service/api/v1/repositories',
    method: 'post',
    data
  })
}

// 获取仓库详情
export function getRepository(id: string | number) {
  return request({
    url: `artifact_service/api/v1/repositories/${id}`,
    method: 'get'
  })
}

// 更新仓库
export function updateRepository(id: string | number, data: RepositoryPayload) {
  return request({
    url: `artifact_service/api/v1/repositories/${id}`,
    method: 'put',
    data
  })
}

// 删除仓库
export function deleteRepository(id: string | number) {
  return request({
    url: `artifact_service/api/v1/repositories/${id}`,
    method: 'delete'
  })
}