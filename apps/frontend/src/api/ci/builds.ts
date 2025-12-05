import request from '@/utils/request'

export interface PaginationParams {
  page?: number
  pageSize?: number
  page_size?: number
  [key: string]: any
}

export interface CreateBuildPayload {
  pipeline_id?: string | number
  triggered_by?: string
  commit_sha?: string
  branch?: string
  variables?: Record<string, string>
}

const CI_PREFIX = 'ci_service/api/v1'

function normalizePageParams(params: PaginationParams = {}) {
  const query: PaginationParams = { ...params }
  if (query.pageSize && !query.page_size) {
    query.page_size = query.pageSize
    delete query.pageSize
  }
  return query
}

export function createExecutorBuild(data: CreateBuildPayload = {}) {
  return request({ url: `${CI_PREFIX}/executor/builds`, method: 'post', data })
}

export function getExecutorBuild(buildId: string | number) {
  return request({ url: `${CI_PREFIX}/executor/builds/${buildId}`, method: 'get' })
}

export function listExecutorBuilds(pipelineId: string | number, params: PaginationParams = {}) {
  const query = normalizePageParams(params)
  return request({ url: `${CI_PREFIX}/executor/pipelines/${pipelineId}/builds`, method: 'get', params: query })
}

export function getExecutorBuildLogs(buildId: string | number, offset = 0, limit = 200) {
  const params: any = { offset, limit }
  return request({ url: `${CI_PREFIX}/executor/builds/${buildId}/logs`, method: 'get', params })
}

export function getExecutorK8sStatus(buildId: string | number, jobNamePrefix = '', page = 1, pageSize = 20) {
  const params: any = { job_name_prefix: jobNamePrefix, page, page_size: pageSize }
  return request({ url: `${CI_PREFIX}/executor/builds/${buildId}/k8s_status`, method: 'get', params })
}

export function cancelExecutorBuild(buildId: string | number) {
  return request({ url: `${CI_PREFIX}/executor/builds/${buildId}/cancel`, method: 'post' })
}