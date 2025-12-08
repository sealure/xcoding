# XCoding 平台

XCoding 是一款基于 Kubernetes 原生架构构建的企业级微服务研发平台。项目采用最新的技术栈组合，包括 gRPC/Buf 接口定义、APISIX 高性能网关以及 Helm 声明式部署管理。平台集成了用户权限管理、项目管理、代码仓库托管、制品管理以及 CI/CD 流水线等核心能力，旨在提供一套现代化、高性能且易于扩展的一站式研发效能解决方案。

## 基本特点
- **标准化接口定义**：基于 Protocol Buffers (Buf) 进行统一的接口定义与管理。利用 gRPC-Gateway 自动生成符合 RESTful 规范的 HTTP 接口，兼顾内部服务间的高性能 gRPC 通信与外部调用的便捷性。
- **统一网关与安全认证**：采用 APISIX 作为统一流量入口，集成 User 服务实现集中式认证与鉴权。所有请求在转发至后端微服务前均经过严格的安全验证，确保系统的安全性与访问控制的一致性。
- **模块化微服务架构**：系统拆分为功能独立的微服务模块，包括用户管理 (User)、项目管理 (Project)、 CI核心引擎 (Pipeline & Executor)、代码仓库 (Code Repository)及制品库 (Artifact)， 各服务职责边界清晰，易于维护与扩展。
- **云原生架构设计**：完全基于 Kubernetes 原生体系构建。各服务支持健康检查与水平自动伸缩；CI 流水线任务通过动态调度 K8s Job 执行，实现计算资源的按需分配与环境隔离。
- **声明式部署管理**：采用 Helm Chart 统一编排所有微服务及其基础设施依赖（PostgreSQL, RabbitMQ, APISIX 等），支持声明式的版本管理与一键化部署升级，简化运维复杂度。

## 界面预览
- 任务执行中：![任务执行中](docs/img/1任务执行中.png)
- Job 运行结果查看：![Job 运行结果查看](docs/img/2Job运行结果查看.png)
- 实时查看 YAML 视图：![实时查看 YAML 视图](docs/img/3实时查看yaml视图.png)
- 新建 API Token：![新建 API Token](docs/img/4新建apitoken.png)
- 设置主题：![设置主题](docs/img/5设置主题.png)

## 部署方式
1. 准备 Kubernetes 集群，并确保本机可用的 `kubectl`、`helm`、`docker`。
2. 进入仓库根目录，按需修改 `deploy/xcoding/values.yaml` 中镜像仓库与端口等配置。

### 一键部署（推荐）
- 2核2G的电脑，确保已配置好 `kubectl`、`helm`、`docker` 环境。
- 使用 `deploy/deploy_all.py` 执行全量构建/推送/部署：
  - 运行：`python deploy/deploy_all.py`
  - （可选）可通过 `python deploy/deploy_all.py -h` 查看可用选项

### 按微服务部署（可选）
- 用户服务：`python deploy/python/user/deploy.py --tag <tag>`
- 项目服务：`python deploy/python/project/deploy.py --tag <tag> --action deploy`
- 代码仓库服务：`python deploy/python/code_repository/deploy.py --tag <tag>`
- CI 执行器：`python deploy/python/executor_service/deploy.py --tag <tag>`
- CI Pipeline：`python deploy/python/ci/deploy.py --tag <tag>`
- 前端：`python deploy/python/frontend/deploy.py --tag <tag>`

### Helm 安装/升级
- `helm upgrade --install xcoding deploy/xcoding -n xcoding`

### 访问入口
- APISIX 通过 NodePort 暴露（默认 `31080`），示例域名 `api.xcoding.local`/`xcoding.local`。

提示：脚本会自动更新对应 Helm 模板中的镜像标签，并在命名空间不存在时创建 `xcoding` 命名空间；安装后可通过脚本提供的状态与日志命令查看部署情况。

## 文档索引
- [用户服务概述](docs/user/README.md)与[权限/令牌范围](docs/user/permissions.md)。
- [项目服务概述与接口](docs/project/README.md)；权限模型参见 [docs/project/project-permissions.md](docs/project/project-permissions.md)。
- [代码仓库服务架构说明](docs/code_repository/feature/README.md)；权限见 [docs/code_repository/permissions.md](docs/code_repository/permissions.md)。
- [制品服务概述](docs/artifact/README.md)；权限见 [docs/artifact/permissions.md](docs/artifact/permissions.md)。
- [单元/E2E 测试策略与运行指南](docs/tests/README.md)。
- [CI Pipeline 服务说明、数据流与架构图](docs/ci/pipeline_service/README.md)。
- [CI 执行器服务说明、数据流与架构图](docs/ci/executor_service/README.md)。
- [CI 插件`steps.uses` 的规划与当前进度](docs/ci/actions/progress.md)。

## 代码生成
- Buf 配置：`buf.yaml`，模块路径 `proto/`。
- 生成命令示例：`buf generate`。

## 许可
本项目源代码遵循BSD-3许可声明
