import request from '@/utils/request'

export interface APITokenCreateInput {
  name: string
  description?: string
  expires_in: TokenExpiration
  scopes: Array<Scope | number>
}

export const TOKEN_EXPIRATION = {
  UNSPECIFIED: 0,
  NEVER: 1,
  ONE_DAY: 2,
  ONE_WEEK: 3,
  ONE_MONTH: 4,
  THREE_MONTHS: 5,
  ONE_YEAR: 6
} as const

export type TokenExpiration = typeof TOKEN_EXPIRATION[keyof typeof TOKEN_EXPIRATION]

export const TOKEN_EXPIRATION_OPTIONS: Array<{ label: string; value: TokenExpiration }> = [
  { label: '永不过期', value: TOKEN_EXPIRATION.NEVER },
  { label: '1天', value: TOKEN_EXPIRATION.ONE_DAY },
  { label: '7天', value: TOKEN_EXPIRATION.ONE_WEEK },
  { label: '30天', value: TOKEN_EXPIRATION.ONE_MONTH },
  { label: '90天', value: TOKEN_EXPIRATION.THREE_MONTHS },
  { label: '365天', value: TOKEN_EXPIRATION.ONE_YEAR }
]

export const SCOPE = {
  UNSPECIFIED: 'SCOPE_UNSPECIFIED',
  READ: 'SCOPE_READ',
  WRITE: 'SCOPE_WRITE',
  DELETE: 'SCOPE_DELETE',
  ADMIN: 'SCOPE_ADMIN',
  USER_MANAGEMENT: 'SCOPE_USER_MANAGEMENT',
  PROJECT_MANAGEMENT: 'SCOPE_PROJECT_MANAGEMENT',
  REPOSITORY_ACCESS: 'SCOPE_REPOSITORY_ACCESS',
  PIPELINE_MANAGEMENT: 'SCOPE_PIPELINE_MANAGEMENT',
  ARTIFACT_MANAGEMENT: 'SCOPE_ARTIFACT_MANAGEMENT'
} as const

export type Scope = typeof SCOPE[keyof typeof SCOPE]

export const SCOPE_OPTIONS: Array<{ label: string; value: Scope; description: string }> = [
  { label: '读取权限', value: SCOPE.READ, description: '允许获取资源信息' },
  { label: '写入权限', value: SCOPE.WRITE, description: '允许创建和修改资源' },
  { label: '删除权限', value: SCOPE.DELETE, description: '允许删除资源' },
  { label: '管理员权限', value: SCOPE.ADMIN, description: '拥有所有权限' },
  { label: '用户管理权限', value: SCOPE.USER_MANAGEMENT, description: '允许管理用户账户' },
  { label: '项目管理权限', value: SCOPE.PROJECT_MANAGEMENT, description: '允许管理项目' },
  { label: '代码仓库权限', value: SCOPE.REPOSITORY_ACCESS, description: '允许访问代码仓库' },
  { label: '流水线权限', value: SCOPE.PIPELINE_MANAGEMENT, description: '允许管理CI/CD流水线' },
  { label: '制品权限', value: SCOPE.ARTIFACT_MANAGEMENT, description: '允许管理构建制品' }
]

/**
 * 创建 API Token
 * @param {Object} data - 创建 API Token 的数据
 * @param {string} data.name - Token 名称
 * @param {string} data.description - Token 描述
 * @param {number} data.expires_in - 过期时间枚举值
 * @param {Array<number>} data.scopes - 权限范围数组
 * @returns {Promise} API 响应
 */
export const createAPIToken = (data: APITokenCreateInput) => {
  return request({
    url: '/user_service/api/v1/tokens',
    method: 'post',
    data
  })
}

/**
 * 获取 API Token 列表
 * @returns {Promise} API 响应
 */
export const listAPITokens = () => {
  return request({
    url: '/user_service/api/v1/tokens',
    method: 'get'
  })
}

/**
 * 删除 API Token
 * @param {number} tokenId - Token ID
 * @returns {Promise} API 响应
 */
export const deleteAPIToken = (tokenId: number | string) => {
  return request({
    url: `/user_service/api/v1/tokens/${tokenId}`,
    method: 'delete'
  })
}

// Token 过期时间枚举

/**
 * 格式化权限范围显示文本
 * @param {Array<number>} scopes - 权限范围数组
 * @returns {string} 格式化后的权限范围文本
 */
export const formatScopes = (scopes: Array<Scope | number>) => {
  if (!scopes || scopes.length === 0) return '无权限'

  const scopeLabels = scopes.map(scope => {
    const option = SCOPE_OPTIONS.find(opt => opt.value === scope)
    return option ? option.label : `未知权限(${String(scope)})`
  })

  return scopeLabels.join(', ')
}

/**
 * 格式化过期时间显示文本
 * @param {number} expiration - 过期时间枚举值
 * @returns {string} 格式化后的过期时间文本
 */
export const formatExpiration = (expiration: TokenExpiration) => {
  const option = TOKEN_EXPIRATION_OPTIONS.find(opt => opt.value === expiration)
  return option ? option.label : '未知'
}