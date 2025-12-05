# Actions 支持规划与当前进度

本文记录在 `apps/ci/executor_service` 中为工作流 `steps.uses` 引入支持的目标、架构设计、接入点与后续计划，便于后续迭代。

## 背景与目标
- 背景：CI 执行器通过 K8s Job + 单容器脚本运行步骤，日志标记驱动 Step 状态，gRPC + Gateway 暴露能力，Helm 管理部署。
- 目标：基于现有架构，为 `steps.uses` 引入最小可用的 Actions 机制，支持示例：
  ```yaml
  steps:
    - name: hello action test
      uses: actions/hello@v4
      with:
        env_name: hello
  ```
- 范围：先落地 `uses` 的解析与脚本生成（内置 `actions/hello`），保留后续扩展到远端/容器型/复合 Action。

## 已支持
- 支持 `steps.uses` 引用解析：`owner/name@version`、`owner/name/path@version`
- 自动下载 Actions 包并解析 `action.yml` 元数据
- 将 `with` 输入映射为 `INPUT_*` 环境变量并在脚本中可用
- 按 `runs.using` 生成脚本并注入 `BuildScript(job)` 执行
- 已支持类型：`composite`（展开 `run` 子步骤）、`node`（镜像包含 Node 运行时）
- 暂不支持：`docker`（以日志提示方式告知）
- 详细实现与约束见 [executor_service/README.md：数据流与状态 → 脚本与 Actions](../executor_service/README.md#数据流与状态)

## 现状概览（代码参照）
- 解析器：支持 `steps.run`/`steps.uses`，尚无 `with` 字段。
  - `apps/ci/executor_service/internal/parser/workflow_parser.go:24-29`
- 脚本生成：仅处理 `run` 步骤；`uses` 目前不产生脚本。
  - `apps/ci/executor_service/internal/executor/script_builder.go:26-52`
- 单步包装与错误策略：
  - `apps/ci/executor_service/internal/executor/step_runner.go:10-37`
- Job 构建入口：
  - `apps/ci/executor_service/internal/executor/job_builder.go:15-25`
- DAG 执行与单 Job 调度：
  - `apps/ci/executor_service/internal/executor/dag_engine.go:66-101`
  - `apps/ci/executor_service/internal/executor/dag_scheduler.go:48-76`、`77-103`
- 资源与扩展：
  - `apps/ci/executor_service/internal/executor/resources_injector.go:13-29`
  - `apps/ci/executor_service/internal/executor/podspec_extensions.go:10-39`
- 日志标记与格式化、WS 推送：
  - `apps/ci/executor_service/internal/executor/log_processor.go:68-102`
  - `apps/ci/executor_service/internal/executor/log_formatter.go:1-33`
  - `apps/ci/executor_service/internal/ws/handler.go:263-306`

## 方案设计
- 新增模块：`apps/ci/executor_service/internal/executor/actions`
  - 作用：解析 `uses` 引用并生成对应脚本片段，注入到现有 `BuildScript(job)` 流程。
- 引用规范：`<owner>/<name>@<version>`；本期使用内置 registry，不做远端拉取。
- `with` 约定：转换为步骤输入环境变量 `INPUT_<UPPER_SNAKE_CASE_KEY>`，与 GitHub Actions 行为对齐。

### 模块文件拆分
- `actions/types.go`
  - `type Action interface { Build(step parser.Step, job parser.Job) (string, error) }`
  - `type ParsedRef struct { Owner, Name, Version string }`
- `actions/registry.go`
  - `func Register(name string, a Action)`、`func Get(name string) (Action, bool)`、`func RegisterBuiltins()`
- `actions/resolver.go`
  - `func ParseUsesRef(uses string) (ParsedRef, error)`
  - `func BuildUsesScript(step parser.Step, job parser.Job) (string, error)`：生成脚本并导出 `INPUT_*`
- `actions/builtin/hello.go`
  - `HelloAction` 从 `with.env_name` 读取并输出 `echo "hello action test: $INPUT_ENV_NAME"`

### 解析与脚本接入点
- 解析器变更：为 `Step` 增加 `With map[string]string` 字段。
  - 修改位置：`apps/ci/executor_service/internal/parser/workflow_parser.go`
- 脚本生成：`BuildScript(job)` 中每个步骤：
  - 输出 `✔️__step_begin__ <name>` 标记。
  - 若 `st.Uses` 非空：调用 `actions.BuildUsesScript(st, job)` 获取脚本片段并写入。
  - 否则沿用现有 `BuildStepCommand(st)` 逻辑。
  - 输出 `__step_end__ <name>` 标记。
  - 修改位置：`apps/ci/executor_service/internal/executor/script_builder.go`
- 注册内置 Actions：在服务初始化处执行 `actions.RegisterBuiltins()`。
  - 建议位置：`apps/ci/executor_service/cmd/main.go`

## 数据流不变部分
- 触发 → 解析 → DAG → K8s Job → 流日志 → 状态落库 → WS 推送：
  - 解析快照：`apps/ci/executor_service/internal/consumer/queue_consumer.go:87-90`
  - 初始化 Job/Step：`apps/ci/executor_service/internal/consumer/queue_consumer.go:92-107`
  - 并发执行：`apps/ci/executor_service/internal/executor/dag_engine.go:66-101`
  - Job 生命周期与日志：`apps/ci/executor_service/internal/executor/dag_scheduler.go:48-103`
  - 状态收敛与日志入库：`apps/ci/executor_service/internal/executor/log_processor.go:68-102`
  - 前端 WS：`apps/ci/executor_service/internal/ws/handler.go:263-306`

## 示例与约定
- 工作流示例路径：`apps/frontend/public/workflows/example.yml`
- `with` 转环境规则：`with.env_name: hello` → 导出 `INPUT_ENV_NAME=hello`，脚本中使用 `$INPUT_ENV_NAME`。
- 错误策略：继续错误由 `XC_CONTINUE_ON_ERROR` 控制，沿用现有包装：`apps/ci/executor_service/internal/executor/step_runner.go:19-36`
- 资源/超时/节点选择：保持现有注入约定（`XC_RESOURCE_*`、`XC_JOB_TIMEOUT_SECONDS`、`XC_NODE_SELECTOR_*`）。

## 当前进度
- 架构与模块划分已确定，明确接入点与最小可用能力。
- 拟新增的 Go 文件与函数接口已定义（见“模块文件拆分”）。
- 需要实施的改动点：
  - 解析器：`Step.With` 字段解析
  - 脚本生成：`uses` 分支接入 `actions.BuildUsesScript`
  - main 初始化：注册内置 `actions`
  - 新增 `actions` 目录与内置 `hello` action

## 下一步计划
- 实施解析器与脚本改动，落地 `actions/hello@v4`。
- 写一个使用 `uses` 的工作流示例，验证日志输出与 Step 状态。
- 增加最小单元测试（resolver/registry/BuildScript）。

## 扩展路线（后续迭代）
- 远端/复合/容器型 Actions：解析 `action.yml`，支持 `runs.using` 为 `composite/docker/node`。
- 镜像型 Action：允许 `Action.Build` 返回镜像与命令覆盖，调整 `BuildJobSpec`。
- 缓存与工件：为 Action 提供共享工作空间与缓存策略。
- Secrets/权限：统一 `with` 敏感输入的处理与验证策略。

## 校验与风险
- 校验：
  - E2E 通过现有 DAG/日志通道，断言 Step 状态与日志包含 `hello action test: hello`。
  - 单测覆盖 `ParseUsesRef`、`BuildUsesScript` 与 `BuildScript` 分支。
- 风险：
  - `with` 注入需避免与 Job/Step env 冲突；采用 `INPUT_*` 前缀规避。
  - 后续容器型 Action 需要仔细处理镜像安全与拉取策略。
