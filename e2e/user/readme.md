# User 服务 E2E 测试说明

本目录的 E2E 覆盖了用户注册、登录、API Token 管理与用户 CRUD。

## 覆盖范围
- 注册与登录：`register_spec_test.go`、`login_spec_test.go`、`auth_spec_test.go`
- API Token 管理：`apitokens_spec_test.go`
- 用户增删改查：`user_crud_spec_test.go`

## 测试特性
- 通过 `e2e/helpers/auth.go` 提供注册与登录辅助
- 使用统一的 `e2e/helpers/e2e_http.go` 发起请求并断言状态码
- 验证鉴权上下文在跨服务请求中的传播与头一致性

## 运行示例
- 单包运行：`go test -tags e2e -v ./e2e/user`
- 聚合器运行：`go test all_e2e_test.go -v`