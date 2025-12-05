# 用户服务文档

本文档概述 User 微服务的职责、接口、令牌类型与与网关的对接方式。权限细则与 API 令牌范围请参阅同目录下的 `permissions.md`。

## 服务概述
- 服务职责：用户注册/登录、用户资料管理、用户列表、用户删除、认证鉴权（`Auth`），以及 API 令牌的创建/列表/删除。
- 令牌类型：
  - 用户登录令牌（JWT，`TokenType=USER`）：用于用户交互场景，`Auth` 会解码并返回用户信息与过期时间。
  - API 令牌（前缀 `tk_`，`TokenType=API`）：用于服务与自动化调用，`Auth` 会返回绑定用户与范围（Scopes）。
- 网关对接：`Auth` 返回结构化头信息，供 APISIX forward-auth 或上游代理注入下游服务。

## Proto 参考
- [`proto/user/v1/user.proto`](../../proto/user/v1/user.proto)：`UserService` RPC 声明，包含用户操作与令牌接口；与 HTTP 网关路径映射。
- [`proto/user/v1/apitoken.proto`](../../proto/user/v1/apitoken.proto)：`Scope`（范围枚举）与 `TokenExpiration`（有效期枚举），以及令牌的请求/响应消息。

## 认证头部与上下文传递
- `Auth` 返回的 `Headers` 会包含：
  - `X-User-ID`、`X-Username`、`X-User-Role`；当为 API 令牌且有范围时还包含 `X-Scopes`（逗号分隔）。
- 下游服务通过 gRPC `metadata` 读取以上头部进行权限判定，示例参见 `docs/code_repository/permissions.md` 与 `docs/project/project-permissions.md`。

## 相关实现位置
- 服务实现：`apps/user/internal/service/user_service.go`
- 角色辅助：`apps/user/internal/service/roles.go`
- gRPC/HTTP 启动与拦截器：`apps/user/cmd/main.go`

## 测试建议
- 覆盖注册/登录流程、`Auth` 对 JWT 与 API 令牌的判定与回显头、令牌创建/列表/删除与过期处理。
- 建议在端到端测试中验证下游服务读取头部并进行权限判断的行为。
