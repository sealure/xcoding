# e2ehttp: Go E2E HTTP Helpers

共享的 Go E2E 测试 HTTP 辅助包，统一方法/状态码常量、路由探测与请求行为（禁用重定向、默认 JSON/UA）。已在 `user`、`code_repository`、`project` 三个服务接入。

## 特性
- 方法与状态码常量重导出：`MethodGet/MethodPost/MethodPut/MethodDelete` 与常见 HTTP 状态码。
- `DoRequest`：禁用重定向；默认 `Accept: application/json` 与 `User-Agent: xcoding-e2e-go`；覆盖常见状态码范围以返回响应体与状态码。
- 路由探测：`RouteExists`（GET）与 `RouteExistsWithMethod`（按方法），自动验证 `Content-Type` 为 JSON；`RequireRouteOrSkip`/`RequireMethodRouteOrSkip` 在不可达时跳过测试。

## 快速开始（在各服务 e2e/helpers 中接入）
```go
package e2e

import (
    "os"
    "testing"
    "xcoding/pkg/e2ehttp"
)

// BaseURL 可通过环境变量覆盖
var BaseURL = "http://xcoding.local:31080"
func init() {
    if v := os.Getenv("XCODING_BASE_URL"); v != "" { BaseURL = v }
}

// 方法与状态码常量（从共享包重导出）
const (
    MethodGet    = e2ehttp.MethodGet
    MethodPost   = e2ehttp.MethodPost
    MethodPut    = e2ehttp.MethodPut
    MethodDelete = e2ehttp.MethodDelete

    StatusOK                  = e2ehttp.StatusOK
    StatusCreated             = e2ehttp.StatusCreated
    StatusNoContent           = e2ehttp.StatusNoContent
    StatusBadRequest          = e2ehttp.StatusBadRequest
    StatusUnauthorized        = e2ehttp.StatusUnauthorized
    StatusForbidden           = e2ehttp.StatusForbidden
    StatusNotFound            = e2ehttp.StatusNotFound
    StatusMethodNotAllowed    = e2ehttp.StatusMethodNotAllowed
)

// 统一封装请求与探测
func doRequest(t *testing.T, method, endpoint string, body any, headers map[string]string) (int, []byte) {
    return e2ehttp.DoRequest(t, BaseURL, method, endpoint, body, headers)
}
func routeExistsWithMethod(t *testing.T, method, endpoint string) bool {
    return e2ehttp.RouteExistsWithMethod(t, BaseURL, method, endpoint)
}
func requireMethodRouteOrSkip(t *testing.T, method, endpoint string) {
    e2ehttp.RequireMethodRouteOrSkip(t, BaseURL, method, endpoint)
}
```

## 用例示例
```go
headers := map[string]string{"Authorization": "Bearer " + token}
status, body := doRequest(t, MethodGet, "/user_service/api/v1/auth", nil, headers)
if status != StatusOK { t.Fatalf("auth status=%d body=%s", status, string(body)) }
```

## 行为说明
- 重定向：所有请求禁用跟随重定向，避免捕获到前端 HTML；便于准确断言 30x。
- 默认头：未显式传入时自动设置 `Accept: application/json` 与 `User-Agent: xcoding-e2e-go`，可通过 `headers` 覆盖。
- 探测策略：`RouteExists*` 会验证返回 `Content-Type` 包含 JSON 且状态码不是 404。
- BaseURL：各服务的 e2e/helpers 自行维护 `BaseURL`，通过 `XCODING_BASE_URL` 环境变量覆盖。

## 运行
```bash
# 在仓库根目录
export XCODING_BASE_URL="http://xcoding.local:31080"

go test ./apps/user/e2e -v
go test ./apps/code_repository/e2e -v
go test ./apps/project/e2e -v
```

## 约定
- 新增 e2e 场景时，复用该包以保持重定向与头部策略一致。
- 若后端返回结构变更，请在各服务 `helpers_test.go` 更新轻量模型与解析逻辑。