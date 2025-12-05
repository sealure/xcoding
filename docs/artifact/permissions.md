# Artifact 服务权限模型

本文件定义 Artifact 微服务的权限来源、角色判定与在各 RPC 接口上的具体规则，并与 Project 服务的权限模型保持一致。权限边界以项目为单位，公开性规则为非管理员的可见性过滤条件。

## 身份与角色来源与传递
- 网关注入并在 gRPC `metadata` 传递的头部：
  - `x-user-id`、`x-username`、`x-user-role`、`x-scopes`、`authorization`、`x-real-ip`、`x-forwarded-for`、`user-agent`。
- 服务端角色判定：
  - 全局超级管理员：`isUserRoleSuperAdmin(ctx)`（参考 `apps/artifact/internal/service/roles.go`）。
  - 项目成员/角色：通过 Project 服务查询或本服务辅助方法：`isMemberOrHigher(ctx, projectID, actorID)`、`ensureOwnerOrAdmin(ctx, projectID, actorID)`。

## 权限边界与公开性
- 项目边界：Registry 绑定 `project_id`，Namespace/Repository/Tag 权限通过所属 Registry 传递至项目维度。
- 公开性：
  - 非超级管理员访问时，列表/读取接口只允许访问 `is_public=true` 的资源；
  - 对 Repository/Tag，若所属 Namespace/Registry 全部为私有，则需项目成员及以上访问；
  - 超级管理员不受公开性限制。

## RPC 接口权限规则
- Registry：
  - `CreateRegistry`：仅超级管理员。
  - `GetRegistry`：非超级管理员仅能访问公开 Registry；支持按 `project_id` 归属校验。
  - `UpdateRegistry`：仅超级管理员；当提供 `project_id` 时需匹配归属。
  - `ListRegistries`：非超级管理员仅返回公开 Registry；支持 `project_id` 过滤。
  - `DeleteRegistry`：仅超级管理员；当提供 `project_id` 时需匹配归属。
- Namespace：
  - `CreateNamespace`：仅超级管理员；需目标 Registry 存在。
  - `GetNamespace`：非超级管理员仅能访问公开 Registry 下的 Namespace。
  - `UpdateNamespace`：仅超级管理员。
  - `ListNamespaces`：非超级管理员仅返回公开 Registry 下的 Namespace；支持 `registry_id` 过滤。
  - `DeleteNamespace`：仅超级管理员。
- Repository：
  - `CreateRepository`：项目 Owner/Admin（或超级管理员）；需 Namespace 存在。
  - `GetRepository`：非超级管理员访问私有仓库时需项目成员及以上；公开仓库可访问。
  - `UpdateRepository`：需明确项目边界后按 Owner/Admin（或超级管理员）限制；当前实现以资源归属校验为主。
  - `ListRepositories`：非超级管理员仅返回公开仓库或其 Registry 公开的仓库。
  - `DeleteRepository`：项目 Owner/Admin（或超级管理员）。
- Tag：
  - `CreateTag`：项目 Owner/Admin（或超级管理员）。
  - `GetTag`：非超级管理员访问私有 Tag（其 Repository 或 Registry 为私有）时需项目成员及以上；公开 Tag 可访问。
  - `UpdateTag`：项目 Owner/Admin（或超级管理员）。
  - `ListTags`：非超级管理员仅返回公开 Repository 或其 Registry 公开下的 Tag；支持 `repository_id` 过滤。
  - `DeleteTag`：项目 Owner/Admin（或超级管理员）。

## 错误码与处理
- `codes.Unauthenticated`：缺失或非法的认证信息（例如未携带用户头或令牌无效）。
- `codes.PermissionDenied`：角色不足、跨项目访问或非公开资源无成员关系访问。
- `codes.NotFound`：目标资源不存在（注意与权限拒绝区分，避免信息泄露）。
- `codes.InvalidArgument`：参数不合法（如缺少必须的 ID）。
- `codes.AlreadyExists`：资源唯一性冲突（如名称在所属空间内唯一）。

## 审计与日志建议
- 记录重要操作：Registry/Namespace/Repository/Tag 的创建、更新、删除，以及越权拒绝事件。
- 审计字段建议：`time`、`user_id`、`project_id`、`resource_id`、`resource_type`、`action`、`ip`、`user_agent`、`result`。
- 严禁写入敏感信息（用户名/密码/Token 明文）。

## 参考实现位置
- [`apps/artifact/internal/service/registry_service.go`](../../apps/artifact/internal/service/registry_service.go)
- [`apps/artifact/internal/service/namespace_service.go`](../../apps/artifact/internal/service/namespace_service.go)
- [`apps/artifact/internal/service/repository_service.go`](../../apps/artifact/internal/service/repository_service.go)
- [`apps/artifact/internal/service/tag_service.go`](../../apps/artifact/internal/service/tag_service.go)
- 辅助判定：`getUserIDFromCtx`、`isUserRoleSuperAdmin`、`isMemberOrHigher`、`ensureOwnerOrAdmin`（位于 `apps/artifact/internal/service` 目录下）。

## 与项目权限的对齐
- 与 `docs/project/project-permissions.md` 保持一致：项目 Owner/Admin 是写操作的边界，Member+ 是私有资源的读取边界。
- 当项目权限规则调整时，同步更新本文件与实现。