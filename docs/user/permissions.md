# 用户服务权限与令牌范围

本文档说明用户服务的权限边界、管理员判定、API 令牌的范围（Scopes）与有效期模型，以及 `Auth` 接口返回的头部约定，便于下游服务实施一致的授权策略。

## 角色与管理员判定
- 用户角色枚举：`UserRole`（`USER_ROLE_USER`、`USER_ROLE_SUPER_ADMIN`），定义于 [`proto/user/v1/user.proto`](../../proto/user/v1/user.proto)。
- 管理员判定：`isUserRoleSuperAdmin(ctx)`（参考 [`apps/user/internal/service/roles.go`](../../apps/user/internal/service/roles.go)）依据网关注入的 `x-user-role` 判定，大小写不敏感接受 `SUPER_ADMIN`。

## API 令牌范围（Scopes）
- 枚举定义：[`proto/user/v1/apitoken.proto`](../../proto/user/v1/apitoken.proto) 的 `Scope` 枚举，包含但不限于：
  - `SCOPE_READ`（读取）、`SCOPE_WRITE`（写入）、`SCOPE_DELETE`（删除）、`SCOPE_ADMIN`（管理员）
  - `SCOPE_USER_MANAGEMENT`（用户管理）、`SCOPE_PROJECT_MANAGEMENT`（项目管理）
  - `SCOPE_REPOSITORY_ACCESS`（代码仓库访问）、`SCOPE_PIPELINE_MANAGEMENT`（流水线管理）、`SCOPE_ARTIFACT_MANAGEMENT`（制品管理）
- 令牌生效：`Auth` 会把范围以 `X-Scopes` 逗号分隔字符串回显，下游服务据此实施接口级范围校验（推荐）。

## API 令牌有效期
- 枚举定义：`TokenExpiration`（`NEVER`、`ONE_DAY`、`ONE_WEEK`、`ONE_MONTH`、`THREE_MONTHS`、`ONE_YEAR`），定义于 [`proto/user/v1/apitoken.proto`](../../proto/user/v1/apitoken.proto)。
- 生效逻辑：创建时转换为 `expires_at`；`Auth` 验证时如过期则返回未认证原因；`ListAPITokens` 返回 `expires_at` 供 UI 显示。

## 用户服务 RPC 权限规则
- `Register`：公开接口；创建后默认角色为 `USER_ROLE_USER`。
- `Login`：公开接口；返回用户 JWT。
- `GetUser`：已认证用户可获取自己的资料；超级管理员可通过 `user_id` 获取任意用户资料。
- `UpdateUser`：
  - 超级管理员可更新敏感字段（`role`、`is_active`）并可更新其他用户；
  - 非管理员只能更新自己的基础信息（`username`、`email`、`avatar`）。
- `ListUsers`：当前实现为开放列表（注意生产环境建议加权限控制或仅管理员可见）。
- `DeleteUser`：当前实现为开放删除（注意生产环境建议加权限控制或仅管理员可执行）。
- `Auth`：验证 `JWT` 与 `API` 令牌，返回用户信息、过期时间与范围字符串；并生成 `X-User-ID`、`X-Username`、`X-User-Role`、`X-Scopes`（如有）。
- `CreateAPIToken`：要求调用者已认证（从上下文解析用户），持久化令牌哈希与所选范围、有效期；返回一次性明文令牌（前缀 `tk_`）。
- `ListAPITokens`：返回调用者拥有的令牌列表（不包含明文令牌），附带范围与创建/过期时间。
- `DeleteAPIToken`：删除指定令牌 ID（不检查所有权的实现可能需强化，建议后续按 `user_id` 归属校验）。

## 与下游服务的对接建议
- 网关在调用 `Auth` 后将返回的 `Headers` 注入到下游服务请求头；下游服务从 gRPC `metadata` 读取：
  - 基础身份：`x-user-id`、`x-username`、`x-user-role`；范围字符串：`x-scopes`。
- 下游服务在接口上实施范围与角色联合判定：
  - 例如制品管理操作需包含 `SCOPE_ARTIFACT_MANAGEMENT` 或管理员角色；
  - 项目写操作需 `SCOPE_PROJECT_MANAGEMENT` 或项目 Owner/Admin；
  - 代码仓库读写需 `SCOPE_REPOSITORY_ACCESS`，写操作可能同时需要 `SCOPE_WRITE`。

## 错误码约定
- `codes.Unauthenticated`：令牌缺失或非法（JWT 无效、API 令牌不存在或过期）。
- `codes.PermissionDenied`：角色或范围不足（建议由下游服务在执行具体操作时返回）。
- `codes.NotFound`：目标用户或令牌不存在。
- `codes.InvalidArgument`：参数非法。
- `codes.AlreadyExists`：用户名/邮箱唯一性冲突（注册场景）。

## 参考实现位置
- 服务实现：[`apps/user/internal/service/user_service.go`](../../apps/user/internal/service/user_service.go)
- 角色判定：[`apps/user/internal/service/roles.go`](../../apps/user/internal/service/roles.go)
- Proto 枚举与消息：[`proto/user/v1/user.proto`](../../proto/user/v1/user.proto)、[`proto/user/v1/apitoken.proto`](../../proto/user/v1/apitoken.proto)