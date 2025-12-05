import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface ProjectCreatePayload {
  name: string
  description?: string
  [key: string]: any
}

export interface ProjectUpdatePayload {
  name?: string
  description?: string
  [key: string]: any
}

// 获取项目列表
// 统一分页参数命名：支持 pageSize 自动映射到 page_size
export function getProjectList(params: PaginationParams = {}) {
  const query: PaginationParams = { ...params }
  if (query.pageSize && !query.page_size) {
    query.page_size = query.pageSize
    delete query.pageSize
  }
  return request({
    url: 'project_service/api/v1/projects',
    method: 'get',
    params: query
  })
}

// 创建项目
export function createProject(data: ProjectCreatePayload) {
  return request({
    url: 'project_service/api/v1/projects',
    method: 'post',
    data
  })
}

// 获取项目详情
export function getProjectDetail(id: string | number) {
  return request({
    url: `project_service/api/v1/projects/${id}`,
    method: 'get'
  })
}

// 更新项目
export function updateProject(id: string | number, data: ProjectUpdatePayload) {
  return request({
    url: `project_service/api/v1/projects/${id}`,
    method: 'put',
    data
  })
}

// 删除项目
export function deleteProject(id: string | number) {
  return request({
    url: `project_service/api/v1/projects/${id}`,
    method: 'delete'
  })
}

// -------- ProjectMember 相关接口 --------
// 约定 REST 路径：/projects/{project_id}/members
// 对应 proto：AddProjectMember, ListProjectMembers, UpdateProjectMember, RemoveProjectMember

// 添加项目成员
export function addProjectMember(projectId: string | number, data: Record<string, any> = {}) {
  const payload: Record<string, any> = { ...data }
  // 字段规范：将 role 映射为 snake_case，如有需要
  if (payload.memberRole && !payload.role) {
    payload.role = payload.memberRole
    delete payload.memberRole
  }
  return request({
    url: `project_service/api/v1/projects/${projectId}/members`,
    method: 'post',
    data: payload
  })
}

// 列出项目成员（分页可选）
export function listProjectMembers(projectId: string | number, params: PaginationParams = {}) {
  return request({
    url: `project_service/api/v1/projects/${projectId}/members`,
    method: 'get',
    params
  })
}

// 更新项目成员（例如修改角色）
export function updateProjectMember(projectId: string | number, memberId: string | number, data: Record<string, any> = {}) {
  const payload: Record<string, any> = { ...data }
  if (payload.memberRole && !payload.role) {
    payload.role = payload.memberRole
    delete payload.memberRole
  }
  return request({
    url: `project_service/api/v1/projects/${projectId}/members/${memberId}`,
    method: 'put',
    data: payload
  })
}

// 移除项目成员
export function removeProjectMember(projectId: string | number, memberId: string | number) {
  return request({
    url: `project_service/api/v1/projects/${projectId}/members/${memberId}`,
    method: 'delete'
  })
}