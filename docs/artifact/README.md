# Artifact 服务文档

本文档概述 Artifact 微服务的职责、核心资源、接口与依赖，以及与项目/用户权限的关系。权限细则请参阅同目录下的 `permissions.md`。

## 服务概述
- 服务职责：管理构建制品仓库，包括 `Registry`、`Namespace`、`Repository`、`Tag` 等实体的创建、查询、更新、删除与列表。
- 权限边界：以 `project_id` 为主（来源于 Registry）。Repository/Tag 通过所属的 Namespace → Registry 传递到项目维度。
- 公开/私有：`Registry.IsPublic` 与 `Repository.IsPublic` 控制非管理员的可见性；列表/读取接口会基于公开性与项目成员关系进行过滤/校验。

## 核心资源
- Registry：制品源/仓库的注册信息，包含 URL、认证信息、是否公开、归属 `project_id`，以及 `ArtifactType`、`ArtifactSource`。
- Namespace：在某个 Registry 下的命名空间，用于组织 Repository。
- Repository：具体制品仓库，支持公开/私有、路径信息。
- Tag：制品标签，包含摘要（digest）、大小、是否最新等。

## 接口与 Proto 参考
- Proto 目录：`proto/artifact/v1/`
  - [`artifact.proto`](../../proto/artifact/v1/artifact.proto)：`ArtifactService` 服务声明与 HTTP 映射注解
  - [`registry.proto`](../../proto/artifact/v1/registry.proto)：Registry 消息与请求/响应
  - [`namespace.proto`](../../proto/artifact/v1/namespace.proto)：Namespace 消息与请求/响应
  - [`repository.proto`](../../proto/artifact/v1/repository.proto)：Repository 消息与请求/响应
  - [`tag.proto`](../../proto/artifact/v1/tag.proto)：Tag 消息与请求/响应

## 依赖与上下游
- 依赖 Project 服务：用于项目成员/角色判定（Owner/Admin/Member）。
- 依赖 User 服务：通过网关转发的头部（`X-User-ID`、`X-User-Role` 等）承载身份与角色信息。
- 网关/拦截器：统一认证与日志、监控拦截器，确保上下文带有用户信息与角色标识。

## 相关实现位置
- Service 接口定义：`apps/artifact/internal/service/artifact_service.go`
- Registry 处理：`apps/artifact/internal/service/registry_service.go`
- Namespace 处理：`apps/artifact/internal/service/namespace_service.go`
- Repository 处理：`apps/artifact/internal/service/repository_service.go`
- Tag 处理：`apps/artifact/internal/service/tag_service.go`

## 测试建议
- 建议编写覆盖公开/私有可见性、成员角色（Member+/Owner/Admin）、管理员绕过的用例。
- 端到端测试应包含列表过滤（公开性）、读取权限（私有资源需成员+）、创建/更新/删除的角色限制。
