# Code Repository 服务权限模型

本文件定义 Code Repository 微服务的权限模型、来源与传递、在各 RPC 接口上的具体规则，以及错误与审计建议。该模型与 Project 服务的权限模型保持一致，并以项目为边界实施。

## 概述
- 权限边界：所有仓库资源都隶属某个 `project_id`，权限判断以项目为边界。
- 角色继承：仓库的访问权限来源于项目成员角色（参见 `docs/project/project-permissions.md`）。
- 全局管理员：具备平台级管理员身份的用户在权限判断时拥有最高优先级，可越过项目内限制（用于运维和排障）。
- 统一认证：通过网关统一认证，将用户身份与上下文头部转为 gRPC `metadata` 注入服务处理逻辑。

## 身份与角色来源与传递
- 网关从 HTTP 头读取并注入到 gRPC `metadata` 的字段：
  - `authorization`
  - `x-real-ip`
  - `x-forwarded-for`
  - `user-agent`
  - `x-user-id`
  - `x-username`
  - `x-user-role`
  - `x-user-email`
  - `x-scopes`
- 服务端在处理器中读取示例：
  - Go 示例：
    ```go
    md, _ := metadata.FromIncomingContext(ctx)
    userID := first(md.Get("x-user-id"))
    userRole := first(md.Get("x-user-role"))
    authz := first(md.Get("authorization"))
    // first 是一个返回切片第一个元素的辅助方法
    ```
- 角色判定：
  - 项目内角色以 Project 服务的成员关系与角色为准（Owner、Admin、Member）。
  - 全局管理员由统一认证层或用户服务提供的角色/权限标识判定。

## RPC 接口权限规则
以下规则以 `project_id` 为边界进行判定：
- `ListRepositories`：项目成员及以上可访问（Member+）。
- `GetRepository`：项目成员及以上可访问（Member+）。
- `CreateRepository`：仅项目 Owner/Admin 可创建。
- `UpdateRepository`：仅项目 Owner/Admin 可更新。
- `DeleteRepository`：仅项目 Owner/Admin 可删除。
- `TestRepositoryConnection`：仅项目 Owner/Admin 可执行（涉及敏感凭据校验）。
- `GetRepositoryBranches`：项目成员及以上可访问（Member+）。

备注：如后续引入更细粒度权限（如只读成员、维护者），可在此矩阵上扩展。

## 项目校验与成员关系
- 在处理任意与仓库相关的请求前，需校验：
  - `project_id` 是否存在且有效；
  - 当前用户是否为该项目成员；
  - 是否具备执行当前操作所需的最小角色（Owner/Admin/Member）。
- 校验实现建议：
  - 通过调用 Project 服务的 gRPC 接口查询成员关系（推荐）；
  - 或在本服务维护只读的项目成员缓存并定期与 Project 服务同步（需要一致性策略）。

## 认证拦截器与公共方法
- 统一在 gRPC 服务端配置认证拦截器，拦截并验证 `authorization` 与用户头部。
- 公共（无需认证）方法建议仅保留健康检查：`grpc.health.v1.Health.Check` 与 HTTP `GET /healthz`。
- 所有 Code Repository 业务方法默认受保护。

## 错误码与处理
- `codes.Unauthenticated`：缺失或非法的认证信息。
- `codes.PermissionDenied`：角色不足或跨项目访问。
- `codes.NotFound`：目标项目或仓库不存在（注意与权限拒绝区分，避免信息泄露）。
- `codes.InvalidArgument`：参数不合法（如缺少 `project_id`）。
- `codes.AlreadyExists`：资源唯一性冲突（如仓库名在项目内唯一）。

## 审计与日志
- 记录关键操作的审计日志：创建、更新、删除、测试连接、拉取分支。
- 审计字段建议：`time`、`user_id`、`project_id`、`repository_id`、`action`、`ip`、`user_agent`、`result`。
- 敏感信息（密码/Token）绝不可写入日志或审计记录。

## 处理器示例（伪代码）
```go
func (s *Server) CreateRepository(ctx context.Context, req *coderepositoryv1.CreateRepositoryRequest) (*coderepositoryv1.Repository, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    userID := first(md.Get("x-user-id"))
    if userID == "" { return nil, status.Error(codes.Unauthenticated, "missing user") }

    // 1) 校验项目存在与成员关系
    if !isProjectMember(ctx, req.ProjectId, userID) {
        return nil, status.Error(codes.PermissionDenied, "not a project member")
    }
    if !isProjectOwnerOrAdmin(ctx, req.ProjectId, userID) {
        return nil, status.Error(codes.PermissionDenied, "insufficient role")
    }

    // 2) 唯一性检查（项目内仓库名唯一）
    if existsByName(ctx, req.ProjectId, req.Name) {
        return nil, status.Error(codes.AlreadyExists, "repository name exists in project")
    }

    // 3) 创建并返回
    repo := toModel(req)
    if err := s.db.Create(&repo).Error; err != nil { return nil, dbErrToStatus(err) }
    return toProto(repo), nil
}
```

## 测试建议
- 复用现有测试工具：`tests/header/check_forward_auth_headers.py` 与 `tests/user/test_auth_flow.py`。
- 编写仓库接口的权限测试用例：
  - 非成员访问被拒；成员只读接口可访问；Owner/Admin 可进行写操作；
  - 全局管理员可绕过项目内限制（如平台调试场景）。

## 与项目权限的对齐
- 本文与 Project 权限模型保持一致：`docs/project/project-permissions.md`。
- 当 Project 权限规则变更时，应同步更新本文件与相关实现。