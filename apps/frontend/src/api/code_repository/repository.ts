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

export interface RepositoryCreatePayload {
  projectId?: string | number
  project_id?: string | number
  git_url?: string
  url?: string
  authType?: string
  auth_type?: string
  gitUsername?: string
  git_username?: string
  gitPassword?: string
  git_password?: string
  gitSSHKey?: string
  git_ssh_key?: string
  branch?: string
  [key: string]: any
}

// 保留并使用上方带类型的 normalizePageParams

// 获取代码仓库列表（顶层路径，使用 project_id 作为查询参数）
export function getRepositoryList(projectId: string | number | undefined, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  if (projectId) query.project_id = projectId
  return request({
    url: 'code_repository_service/api/v1/repositories',
    method: 'get',
    params: query
  })
}

// 创建代码仓库（顶层路径，project_id 放入请求体）
export function createRepository(projectId: string | number | undefined, data: RepositoryCreatePayload = {}) {
  const payload: RepositoryCreatePayload = { ...data }
  // 如果未显式传入 projectId，但表单里有，则以表单为准
  if (!projectId && payload.projectId) projectId = payload.projectId
  // 使用蛇形字段，并移除驼峰字段，避免 protojson 解析为重复字段
  if (projectId) payload.project_id = projectId
  if ('projectId' in payload) delete payload.projectId
  // 兼容旧表单字段，将 url 映射到 git_url，并移除 url
  if (payload.url && !payload.git_url) {
    payload.git_url = payload.url
    delete payload.url
  }
  // 认证字段映射：authType -> auth_type，gitUsername -> git_username，gitPassword -> git_password，gitSSHKey -> git_ssh_key
  if (payload.authType && !payload.auth_type) {
    payload.auth_type = payload.authType
    delete payload.authType
  }
  if (payload.gitUsername && !payload.git_username) {
    payload.git_username = payload.gitUsername
    delete payload.gitUsername
  }
  if (payload.gitPassword && !payload.git_password) {
    payload.git_password = payload.gitPassword
    delete payload.gitPassword
  }
  if (payload.gitSSHKey && !payload.git_ssh_key) {
    payload.git_ssh_key = payload.gitSSHKey
    delete payload.gitSSHKey
  }
  // 分支默认值（后端也会兜底为 main，这里显式设置）
  if (!payload.branch) payload.branch = 'main'
  return request({
    url: 'code_repository_service/api/v1/repositories',
    method: 'post',
    data: payload
  })
}

// 获取代码仓库详情（按 ID）
export function getRepositoryById(repositoryId: string | number) {
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}`,
    method: 'get'
  })
}

// 兼容旧命名（保留）
export function getRepositoryDetail(_projectId: any, repositoryId: string | number) {
  return getRepositoryById(repositoryId)
}

// 更新代码仓库（按 ID，project_id 可选加入请求体）
export function updateRepository(repositoryId: string | number, data: RepositoryCreatePayload = {}, projectId?: string | number) {
  const payload: RepositoryCreatePayload = { ...data }
  if (projectId) payload.project_id = projectId
  // 防止同时存在 projectId 与 project_id 导致重复字段
  if ('projectId' in payload) delete payload.projectId
  // 兼容旧表单字段，将 url 映射到 git_url
  if (payload.url && !payload.git_url) {
    payload.git_url = payload.url
    delete payload.url
  }
  // 认证字段映射
  if (payload.authType && !payload.auth_type) {
    payload.auth_type = payload.authType
    delete payload.authType
  }
  if (payload.gitUsername && !payload.git_username) {
    payload.git_username = payload.gitUsername
    delete payload.gitUsername
  }
  if (payload.gitPassword && !payload.git_password) {
    payload.git_password = payload.gitPassword
    delete payload.gitPassword
  }
  if (payload.gitSSHKey && !payload.git_ssh_key) {
    payload.git_ssh_key = payload.gitSSHKey
    delete payload.gitSSHKey
  }
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}`,
    method: 'put',
    data: payload
  })
}

// 删除代码仓库（按 ID，project_id 可选作为查询参数）
export function deleteRepository(repositoryId: string | number, projectId?: string | number) {
  const params: Record<string, any> = {}
  if (projectId) params.project_id = projectId
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}`,
    method: 'delete',
    params
  })
}

// 获取仓库提交列表（分页参数使用 snake_case）
export function getRepositoryCommits(repositoryId: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/commits`,
    method: 'get',
    params: query
  })
}

// 获取仓库分支列表（旧接口，字符串列表）
export function getRepositoryBranches(repositoryId: string | number) {
  return request({
    url: `code_repository_service/api/v1/repositories/${repositoryId}/branches`,
    method: 'get'
  })
}