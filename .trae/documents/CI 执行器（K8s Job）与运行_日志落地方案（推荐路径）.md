## 关键结论

* 将运行态数据迁移到执行器微服务更合理：`Build`、`BuildSnapshot`、日志等放在执行器侧持久化；`pipeline_service` 专注流水线定义与触发。

* DAG 的 `job` 与 `step` 必须建表持久化，并维护 `needs` 依赖，支撑并发/顺序、进度、重试、日志归档与检索。

## 职责边界

* `pipeline_service`（控制面）

  * 管理 `Pipeline/Schedule` 与权限（参考 `apps/ci/pipeline_service/cmd/main.go:55-63`）。

  * 触发：`StartPipelineBuild` 改为远程调用执行器 `CreateBuild` 获取 `build_id`，随后向队列发布该 `build_id`。

  * 对外接口仅保留 `pipelines*` 相关；`build*` 接口转由执行器提供或代理。

* `ci-executor`（执行面）

  * 消费 `build_id`，查询本服务数据库中的 `build` 与 `build_snapshot`，解析 YAML，生成 `job/step` 图并执行。

  * 负责运行状态推进与日志写入，提供 `build*` 查询/日志/取消接口。

## 执行器数据模型（新建于执行器微服务）

* `builds`：`id, pipeline_id, triggered_by, status, commit_sha, branch, variables(jsonb), started_at, finished_at, error, created_at`

* `build_snapshots`：`id, build_id(uniq), pipeline_id, workflow_yaml(text), yaml_sha256`

* `build_jobs`：`id, build_id, name, status, started_at, finished_at, index`

* `build_job_edges`：`id, build_id, from_job, to_job`（映射 `needs`）

* `build_steps`：`id, build_id, job_name, index, name, status, started_at, finished_at, exit_code`

* `build_log_chunks`：`id, build_id, seq, job_name(opt), step_index(opt), content(text), created_at`

## 执行器 API（gRPC + Gateway，供前端/网关访问）

* `CreateBuild(pipeline_id, triggered_by, commit_sha, branch, variables)` → 返回 `build_id`，内部同时生成 `build_snapshot`（从 `pipeline_service` 读取当前 YAML）。

* `ListBuilds(pipeline_id, page, size)`、`GetBuild(build_id)` → 基础查询。

* `AppendBuildLogs(build_id, content, seq)`、`GetBuildLogs(build_id, offset, limit)`、`StreamBuildLogs(build_id)` → 日志存取与流式。

* `CancelBuild(build_id)` → 删除 Job 并标记为取消。

## 运行流程（K8s Job）

1. 前端触发编辑页 → `StartPipelineBuild`（经网关 `/ci_service/api/v1/...`）。
2. `pipeline_service`：调用执行器 `CreateBuild` → 获得 `build_id` → 发布到 RabbitMQ（参考队列注入 `apps/ci/pipeline_service/cmd/main.go:68-82`）。
3. 执行器：消费 `build_id` → 查库获取 `workflow_yaml` → 解析 `jobs/steps` 与 `needs` → 生成并提交 K8s Job（一个构建一个 Job）。
4. 执行器：

   * 监听 Job 状态（成功/失败/超时），推进 `build.status/started_at/finished_at`。

   * 读取 Pod 日志（Follow），切片写入 `build_log_chunks`，并按 `job/step` 高亮位置。
5. 前端：

   * 列表页：`/ci/pipelines/:pipeline_id/builds` → 执行器的 `ListBuilds`。

   * 详情页：`/ci/builds/:build_id` → 执行器的 `GetBuild` + 日志接口（分页或流式）。

## 网关路由

* APISIX 保持前缀 `/ci_service/api/v1/*`：

  * `.../pipelines*` → `ci-pipeline:10055`

  * `.../build*` → `ci-executor:10056`

* 也可短期由 `pipeline_service` 代理执行器，前端无感。

## 前端改造（新增文件，划分清晰）

* 页面：

  * `views/ci/builds/BuildList.vue`（运行列表）

  * `views/ci/builds/BuildDetail.vue`（运行详情与日志）

* API：

  * `src/api/ci/builds.ts`（对接执行器的 `List/Get/Logs/Cancel`）

* 触发入口：编辑页增加“保存并运行”，沿用 `StartPipelineBuild`，成功后跳转到详情页。

## 服务结构（新建执行器微服务目录，不与现有混写）

* `apps/ci/executor_service/`

  * `cmd/main.go`（配置、GRPC/Gateway 与 RabbitMQ consumer 启动）

  * `internal/models/*.go`（上述模型）

  * `internal/api/handler/*.go`（gRPC/Gateway 实现）

  * `internal/consumer/queue_consumer.go`（读取 `build_id` 并驱动执行）

  * `internal/k8s/job_runner.go`（生成/提交/监控 Job 与日志）

  * `internal/parser/workflow_parser.go`（解析 `example.yml` 结构，处理 `needs`）

  * `Dockerfile`、Helm 模板与 RBAC（仅操作 Job/Pod/日志）

## 迁移步骤

1. 在执行器实现 `CreateBuild` 与模型迁移；`buf generate` 更新代码。
2. 修改 `pipeline_service` 的 `StartPipelineBuild`：调用执行器 `CreateBuild` 获取 `build_id`，再发布到队列；返回执行器的 `Build` 数据给前端（参考现有返回结构 `apps/ci/pipeline_service/internal/service/build_service.go:93-94`）。
3. 新增前端页面与 API 文件，对接执行器接口。
4. 配置 APISIX 路由分流 `pipelines*` 与 `build*`。

## 取舍与理由

* 迁走运行态数据：减少跨服务交互与耦合，执行器拥有数据与执行逻辑的一致性；`pipeline_service` 保持轻量控制面。

* DAG 持久化：保证依赖图与并发控制、重试/恢复与日志定位，后续支持矩阵、缓存与产物管理的基础。

## 验证路径

* 端口转发：`python deploy/proxy.py`；部署 `ci-pipeline` 与新增 `ci-executor`。

* 前端 `npm run dev`；触发构建，查看列表与详情日志；观察 Job 创建与清理、状态推进与取消行为。

