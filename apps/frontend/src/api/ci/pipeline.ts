import request from '@/utils/request'

// 轻量类型：与后端 proto 命名保持 snake_case，兼容部分 camelCase 输入
export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface PipelineCreatePayload {
  projectId?: string | number
  project_id?: string | number
  name?: string
  description?: string
  // 兼容旧字段：yaml -> workflow_yaml
  yaml?: string
  workflow_yaml?: string
  is_active?: boolean
  // 早期实验字段（后端未使用），保留以避免破坏调用方
  repository_id?: string | number
  branch?: string
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

// 约定接口前缀：ci_service（与其它模块一致风格）。若后端变更为 pipeline_service，可在此集中切换。
const CI_PREFIX = 'ci_service/api/v1'

// 列出流水线：GET /ci_service/api/v1/pipelines
export function listPipelines(projectId?: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  if (projectId) (query as any).project_id = projectId
  return request({
    url: `${CI_PREFIX}/pipelines`,
    method: 'get',
    params: query
  })
}

// 新建流水线：POST /ci_service/api/v1/pipelines
export function createPipeline(data: PipelineCreatePayload = {}) {
  const payload: PipelineCreatePayload = { ...data }
  // 字段归一：projectId -> project_id
  if (payload.projectId && !payload.project_id) {
    payload.project_id = payload.projectId
    delete payload.projectId
  }
  // 映射 YAML 字段到 proto：workflow_yaml
  if (payload.yaml && !payload.workflow_yaml) {
    payload.workflow_yaml = payload.yaml
    delete payload.yaml
  }
  // 默认启用流水线
  if (payload.is_active === undefined) payload.is_active = true
  return request({
    url: `${CI_PREFIX}/pipelines`,
    method: 'post',
    data: payload
  })
}

// 获取流水线详情：GET /ci_service/api/v1/pipelines/{pipeline_id}
export function getPipeline(pipelineId: string | number) {
  return request({
    url: `${CI_PREFIX}/pipelines/${pipelineId}`,
    method: 'get'
  })
}

// 更新流水线：PUT /ci_service/api/v1/pipelines/{pipeline_id}
export function updatePipeline(pipelineId: string | number, data: PipelineCreatePayload = {}) {
  const payload: PipelineCreatePayload = { ...data }
  if (payload.projectId && !payload.project_id) {
    payload.project_id = payload.projectId
    delete payload.projectId
  }
  if (payload.yaml && !payload.workflow_yaml) {
    payload.workflow_yaml = payload.yaml
    delete payload.yaml
  }
  return request({
    url: `${CI_PREFIX}/pipelines/${pipelineId}`,
    method: 'put',
    data: payload
  })
}

// 删除流水线：DELETE /ci_service/api/v1/pipelines/{pipeline_id}
export function deletePipeline(pipelineId: string | number) {
  return request({
    url: `${CI_PREFIX}/pipelines/${pipelineId}`,
    method: 'delete'
  })
}

// 触发运行：POST /ci_service/api/v1/pipelines/{pipeline_id}/runs
// 触发构建：POST /ci_service/api/v1/pipelines/{pipeline_id}/builds
export interface StartBuildPayload {
  // 可选扩展参数，如分支、提交等，后端可忽略
  branch?: string
  commit?: string
  params?: Record<string, any>
}

export function startPipelineBuild(pipelineId: string | number, data: StartBuildPayload = {}) {
  return request({
    url: `${CI_PREFIX}/pipelines/${pipelineId}/builds`,
    method: 'post',
    data
  })
}