# Project 服务 E2E 测试说明

本目录的 E2E 覆盖了项目的创建、成员管理与更新删除。

## 覆盖范围
- 项目创建：`create_project_spec_test.go`
- 项目成员增删与角色变更：`project_members_spec_test.go`
- 项目更新与删除：`project_update_delete_spec_test.go`

## 测试特性
- 通过 `e2e/helpers/project.go` 进行项目创建与成员权限同步
- 验证成员角色（owner/admin/member）对跨服务资源的影响（如 artifact）
- 结合聚合器运行，确保跨包依赖能正确初始化与访问

## 运行示例
- 单包运行：`go test -tags e2e -v ./e2e/project`
- 聚合器运行：`go test all_e2e_test.go -v`