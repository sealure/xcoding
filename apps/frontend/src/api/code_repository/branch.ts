import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

function normalizePageParams(params: PaginationParams = {}) {
  const query: PaginationParams = { ...params }
  if (query.pageSize && !query.page_size) {
    query.page_size = query.pageSize
    delete query.pageSize
  }
  return query
}

export interface BranchPayload {
  projectId?: string | number
  project_id?: string | number
  isDefault?: boolean
  is_default?: boolean
  repository_id?: string | number
  branch_id?: string | number
  name?: string
  [key: string]: any
}

// 保留并使用上方带类型的 normalizePageParams

// 列出分支（实体版）：GET /code_repository_service/api/v1/repositories/{repository_id}/branches/entities
export function listBranches(repositoryId: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches/entities`,
    method: 'get',
    params: query
  })
}

// 创建分支（实体版）：POST /code_repository_service/api/v1/repositories/{repository_id}/branches/entities
// 需要在 body 携带 project_id, name, 可选 is_default
export function createBranch(repositoryId: string | number, data: BranchPayload = {}) {
  const payload: BranchPayload = { ...data }
  // 字段规范：确保使用 snake_case
  if (payload.projectId && !payload.project_id) {
    payload.project_id = payload.projectId
    delete payload.projectId
  }
  if (payload.isDefault !== undefined && !payload.is_default) {
    payload.is_default = payload.isDefault
    delete payload.isDefault
  }
  payload.repository_id = repositoryId

  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches/entities`,
    method: 'post',
    data: payload
  })
}

// 获取单个分支：GET /code_repository_service/api/v1/repositories/{repository_id}/branches/entities/{branch_id}
export function getBranch(repositoryId: string | number, branchId: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches/entities/${branchId}`,
    method: 'get',
    params: query
  })
}

// 更新分支（如设置默认）：PUT /code_repository_service/api/v1/repositories/{repository_id}/branches/entities/{branch_id}
export function updateBranch(repositoryId: string | number, branchId: string | number, data: BranchPayload = {}) {
  const payload: BranchPayload = { ...data }
  if (payload.projectId && !payload.project_id) {
    payload.project_id = payload.projectId
    delete payload.projectId
  }
  if (payload.isDefault !== undefined && !payload.is_default) {
    payload.is_default = payload.isDefault
    delete payload.isDefault
  }
  payload.repository_id = repositoryId
  payload.branch_id = branchId

  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches/entities/${branchId}`,
    method: 'put',
    data: payload
  })
}

// 删除分支：DELETE /code_repository_service/api/v1/repositories/{repository_id}/branches/entities/{branch_id}
export function deleteBranch(repositoryId: string | number, branchId: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches/entities/${branchId}`,
    method: 'delete',
    params: query
  })
}