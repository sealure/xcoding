# Code Repository 服务 E2E 测试说明

本目录的 E2E 覆盖了代码仓库的核心 CRUD 与权限行为。

## 覆盖范围
- 仓库创建、列表与获取：`repository_create_list_get_spec_test.go`
- 仓库更新：`repository_update_spec_test.go`
- 仓库删除：`repository_delete_spec_test.go`
- 分支关联/连接能力验证：`repository_connect_branches_spec_test.go`
- 超管绕过与特权验证：`super_admin_user1_bypass_spec_test.go`

## 测试特性
- 通过 `e2e/helpers` 的 HTTP 与鉴权工具统一请求流程
- 使用 `e2e/code_repository/helpers.go` 提供本服务专属构造与断言
- 权限模型覆盖普通用户与超级管理员，验证绕过逻辑仅适用于超管

## 运行示例
- 单包运行：`go test -tags e2e -v ./e2e/code_repository`
- 聚合器运行：`go test all_e2e_test.go -v`