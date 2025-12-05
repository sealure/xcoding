# 项目服务文档

本文档概述 Project 微服务的职责、接口与与网关/权限的关系。权限细则请参阅同目录下的 `project-permissions.md`。

## 服务概述
- 服务职责：项目的创建/更新/删除/查询，项目成员管理（添加/列表/更新角色/移除）、权限同步等。
- 权限边界：以项目为单位，支持 Owner/Admin/Member/Guest 分级；超级管理员具备平台级绕过能力。
- 网关对接：通过 APISIX Ingress 暴露 HTTP 路径，使用 forward-auth 将认证委托给 User 服务；服务内部通过 gRPC 读取用户头实现授权。

## Proto 参考
- 目录：`proto/project/v1/`
  - `project.proto`：`ProjectService` RPC 声明与 HTTP 映射注解

## HTTP 网关路径示例
- `POST /project_service/api/v1/projects`：创建项目
- `GET /project_service/api/v1/projects/{project_id}`：项目详情
- `GET /project_service/api/v1/projects`：项目列表（分页）
- `PUT /project_service/api/v1/projects/{project_id}`：更新项目
- `DELETE /project_service/api/v1/projects/{project_id}`：删除项目
- `POST /project_service/api/v1/projects/{project_id}/members`：添加成员
- `GET /project_service/api/v1/projects/{project_id}/members`：成员列表
- `PUT /project_service/api/v1/projects/{project_id}/members/{user_id}`：更新成员角色
- `DELETE /project_service/api/v1/projects/{project_id}/members/{user_id}`：移除成员

## 依赖与上下游
- 依赖 User 服务：通过网关注入的 `X-User-ID`、`X-User-Role`、`X-Scopes` 等头部承载身份与角色。
- 下游服务：Artifact、Code Repository、CI 等按项目维度实施权限与资源归属。

## 相关实现位置
- gRPC/HTTP 启动与网关注册：`apps/project/cmd/main.go`
- 网关注册与健康检查：`apps/project/internal/gateway/`
- 业务逻辑：`apps/project/internal/service/`
- 模型与持久化：`apps/project/internal/models/`

## 测试建议
- 覆盖成员增删改、权限同步与分页查询。
- 端到端测试建议通过 APISIX 入口验证 forward-auth 头传播与授权判定一致性。

## 权限模型
- 详见：[项目权限模型](./project-permissions.md)
