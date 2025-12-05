# 测试策略与运行指南

本仓库的 Go 测试分为两类：
- 单元测试（默认随 `go test ./...` 与 CI 执行）
- 端到端（E2E）测试（需显式添加 `-tags e2e`，默认不参与构建）

## 单元测试
- 运行：`go test ./... -v`
- 覆盖范围：各微服务内部包（如 `internal/grpc/handler` 等）
- 目的：快速验证业务逻辑、分页校验、接口契约等，不依赖外部网关或环境

## E2E 测试
- 运行条件：
  - 网关与服务可达，默认基址 `http://xcoding.local:31080`
  - 或通过环境变量覆盖：`export XCODING_BASE_URL=http://your-gateway:31080`
  - 建议在部署完成后执行：`helm status xcoding -n xcoding` 与 `kubectl get pods -n xcoding`
- 运行方式：
  - 单服务：`go test -tags e2e ./apps/user/e2e -v`
  - 聚合器（单文件）：`go test all_e2e_test.go -v`
    - 聚合器会扫描 `./apps/*/e2e`，并对子包执行 `go test -tags e2e -v <pkg>`，只运行 e2e 包，不会触发普通单测
  - 全局（包含所有包）：`go test -tags e2e -v ./...`（注意也会遍历到普通包）
- 行为特性：
  - 各服务 e2e 测试文件带 `//go:build e2e` 与 `// +build e2e`
  - 根聚合器 `all_e2e_test.go` 不需要 build tag，可直接单文件运行
  - 使用 `xcoding/pkg/e2ehttp` 统一请求头与不跟随重定向策略，避免误解析前端 HTML

### 各微服务 E2E 说明
- Artifact：`./e2e/artifact/readme.md`
- Code Repository：`./e2e/code_repository/readme.md`
- Project：`./e2e/project/readme.md`
- User：`./e2e/user/readme.md`
 - CI 执行器/流水线：`./e2e/executor_service` 与 `./e2e/ci`

## CI 流水线
- 工作流文件：`.workflows/example_pipeline.yml`
- 作业划分：
  - `unit-tests`：始终运行，执行 `go test ./... -v`
  - `e2e-tests`：条件运行，需 `workflow_dispatch` 传入 `run_e2e=true` 或设置环境变量 `RUN_E2E=true`
- 运行细则：
  - `e2e-tests` 会使用 `-tags e2e` 并依赖机密 `XCODING_BASE_URL`
  - 若 `XCODING_BASE_URL` 未设置，步骤会提示并跳过 e2e
 - 前端 UI 可在 `apps/frontend` 下本地开发：`npm run dev`；部署后通过 APISIX 访问

## 常见问题
- e2e 测试失败且报头不匹配：检查后端 `Auth` 处理是否正确设置响应头，或确认网关转发不丢失头
- 使用 `go test -tags e2e -v ./...` 时仍触发普通包：请改用聚合器单文件运行 `go test all_e2e_test.go -v`

## 约定
- 新增 e2e 测试文件时，必须在文件顶部添加：
  ```go
  //go:build e2e
  // +build e2e
  ```
- CI 默认只跑单元测试；需要跑 e2e 时使用 `workflow_dispatch` 并传入 `run_e2e=true` 或在环境中设置 `RUN_E2E=true`
- 共享 e2e 辅助位于 `pkg/e2ehttp`，请复用其中的请求与路由探测工具。
