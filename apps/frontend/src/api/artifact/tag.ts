import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface TagPayload {
  name?: string
  description?: string
  [key: string]: any
}

// 列出标签
export function listTags(params: PaginationParams = {}) {
  return request({
    url: 'artifact_service/api/v1/tags',
    method: 'get',
    params
  })
}

// 创建标签
export function createTag(data: TagPayload) {
  return request({
    url: 'artifact_service/api/v1/tags',
    method: 'post',
    data
  })
}

// 获取标签详情
export function getTag(id: string | number) {
  return request({
    url: `artifact_service/api/v1/tags/${id}`,
    method: 'get'
  })
}

// 更新标签
export function updateTag(id: string | number, data: TagPayload) {
  return request({
    url: `artifact_service/api/v1/tags/${id}`,
    method: 'put',
    data
  })
}

// 删除标签
export function deleteTag(id: string | number) {
  return request({
    url: `artifact_service/api/v1/tags/${id}`,
    method: 'delete'
  })
}