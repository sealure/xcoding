import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface NamespacePayload {
  name?: string
  description?: string
  [key: string]: any
}

// 列出命名空间
export function listNamespaces(params: PaginationParams = {}) {
  return request({
    url: 'artifact_service/api/v1/namespaces',
    method: 'get',
    params
  })
}

// 创建命名空间
export function createNamespace(data: NamespacePayload) {
  return request({
    url: 'artifact_service/api/v1/namespaces',
    method: 'post',
    data
  })
}

// 获取命名空间详情
export function getNamespace(id: string | number) {
  return request({
    url: `artifact_service/api/v1/namespaces/${id}`,
    method: 'get'
  })
}

// 更新命名空间
export function updateNamespace(id: string | number, data: NamespacePayload) {
  return request({
    url: `artifact_service/api/v1/namespaces/${id}`,
    method: 'put',
    data
  })
}

// 删除命名空间
export function deleteNamespace(id: string | number) {
  return request({
    url: `artifact_service/api/v1/namespaces/${id}`,
    method: 'delete'
  })
}