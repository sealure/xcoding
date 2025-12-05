# 项目服务权限模型

本文档描述项目服务的角色来源、管理员判定与各接口的权限规则，便于开发与测试保持一致。

## 角色来源与传递
- 角色字符串来源于 HTTP 网关从请求头注入的 `X-User-Role`，服务端在 gRPC 元数据中读取为 `x-user-role`。
- 用户服务生成的 JWT 仅包含 `user_id`，角色信息由用户服务在鉴权后通过网关头部传递。
- 项目服务会将 `X-User-ID`、`X-Username`、`X-User-Role` 通过网关回显，便于客户端与测试校验。

## 管理员判定（全局）
- 统一使用 `apps/project/internal/service/roles.go` 的 `isGlobalAdmin(ctx)`：
  - 从 gRPC 元数据读取 `x-user-role`，进行大小写与空白归一化。
  - 接受的管理员字符串（不区分大小写）：`ADMIN`、`USER_ROLE_ADMIN`、`PROJECT_MEMBER_ROLE_ADMIN`。
- 代码入口统一在 `isOwnerOrAdmin(...)` 先行判定全局管理员，随后再检查项目所有者与成员角色。

## 成员兜底与所有者/管理员
- 当全局管理员不满足时，按项目成员身份兜底：`OWNER` 或 `ADMIN` 角色成员拥有管理员能力。
- `isOwnerOrAdmin(projectID, actorID)`：
  - 全局管理员：放行。
  - 项目所有者（`projects.owner_id == actorID`）：放行。
  - 项目成员（`project_members.role in {OWNER, ADMIN}`）：放行。

## 接口权限规则
- 创建项目 `CreateProject`：
  - 任何已认证用户；项目所有者为调用者。
- 更新项目 `UpdateProject`：
  - 所有者或管理员（全局管理员或项目级 ADMIN/OWNER）。
  - 可更新字段：`name`、`description`、`language`、`framework`、`is_public`、`status`。
- 删除项目 `DeleteProject`：
  - 所有者或管理员。
- 成员管理 `AddMember` / `UpdateMember` / `RemoveMember`：
  - 所有者或管理员。
- 列表项目 `ListProjects`：
  - `all=false`：返回调用者拥有或参与的项目；默认分页 `page=1`、`page_size=10`。
  - `all=true`：仅管理员允许（全局管理员或项目级 OWNER/ADMIN 兜底）。
- 列表成员 `ListProjectMembers`：
  - 当前实现未强制权限限制；主要用于同步与展示，网关回显用户头信息。

## 审计日志
- 主要审计事件：
  - `create_project`：创建项目。
  - `list_projects_all`：`all=true` 的允许/拒绝。
  - `update_project`：更新项目结果。
  - `delete_project`：删除项目结果。
  - `add_member` / `update_member` / `remove_member`：成员管理。
- 日志示例：
  - `audit: action=list_projects_all result=allowed actor_id=123 role=ADMIN`
  - `audit: action=update_project result=ok project_id=10 actor_id=123`

## 相关实现位置
- 全局管理员判定：`apps/project/internal/service/roles.go`（`isGlobalAdmin`、`normalizeRole`）。
- 统一授权入口：`apps/project/internal/service/project_service.go`（`isOwnerOrAdmin` 与各服务方法）。

## 测试与验证
- 端到端测试：
  - `python3 tests/project/test_project_api_extended_e2e.py`
  - `python3 tests/project/test_project_members_update_delete_flow.py`
- 两套测试覆盖：管理员 `all=true`、成员可见性、管理员更新/删除、成员增改删与回显头部校验。