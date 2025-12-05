# Artifact 服务 E2E 测试说明

本目录的 E2E 覆盖了制品库（artifact）的核心能力与权限场景。

## 覆盖范围
- 命名空间（Namespace）增删改查与枚举校验：`namespace_crud_spec_test.go`
- 注册表（Registry）增删改查与枚举校验：`registry_crud_spec_test.go`、`registry_enums_spec_test.go`
- 仓库（Repository）创建、路径合法性与读取：`repository_crud_spec_test.go`、`repository_path_spec_test.go`
- 标签（Tag）创建、读取、更新、删除：`tag_crud_spec_test.go`
- 权限场景：
  - 空用户（未登录）行为：`permissions_blank_user_spec_test.go`
  - 项目成员可读写范围：`permissions_member_spec_test.go`
  - 项目所有者/管理员特权：`permissions_owner_admin_spec_test.go`
  - 公开 Registry 读权限与写限制：`permissions_public_registry_spec_test.go`
  - 公开 Repository 读权限与写限制：`permissions_public_repository_spec_test.go`

## 测试特性
- 统一通过 `e2e/helpers` 构建项目、注册表、仓库、标签等资源
- 读取权限遵循：仓库公开或所属注册表公开即可读取，写入仅限成员/管理者
- 网关要求认证头，公开读取在测试中使用“非成员已登录用户”模拟

## 运行示例
- 单包运行：`go test -tags e2e -v ./e2e/artifact`
- 聚合器运行：`go test all_e2e_test.go -v`