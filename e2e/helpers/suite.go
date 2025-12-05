// Common test suite helpers: base URL and auth header.

package helpers

// DefaultBaseURL 用于在未设置环境变量时的默认网关地址
const DefaultBaseURL = "http://xcoding.local:31080"

// BaseURL 返回统一的网关地址（可被环境变量 XCODING_BASE_URL 覆盖）
func BaseURL() string { return GetBaseURLOrDefault(DefaultBaseURL) }

// AuthHeader 快速生成 Authorization 头
func AuthHeader(token string) map[string]string {
    if token == "" { return nil }
    return map[string]string{"Authorization": "Bearer " + token}
}