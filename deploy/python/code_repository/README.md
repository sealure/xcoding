# 代码仓库服务部署脚本

该 Python 脚本用于管理代码仓库服务的部署，包括读取 Deployment YAML、修改镜像标签、构建镜像、推送以及使用 Helm 部署到 Kubernetes。脚本风格对齐 `deploy/python/user/deploy.py`。

## 功能

- 读取和修改 `templates/services/code_repository/deployment.yaml` 中的镜像标签
- 构建 Docker 镜像（使用 `apps/code_repository/Dockerfile`）
- 推送镜像到私有仓库 `localhost:31500`
- 使用 Helm 安装或升级到 Kubernetes 集群
- 检查部署状态并获取服务日志

## 安装依赖

```bash
pip install -r requirements.txt
```

## 使用方法

### 基本用法

```bash
# 构建、推送、部署指定标签的镜像
python deploy.py --tag 13

# 本地部署（构建与部署，跳过推送）
python deploy.py --tag 13  # 当前版本未提供 --no-push，可按需扩展

# 指定项目根目录
python deploy.py --tag 13 --project-root /path/to/project  # 当前版本未提供该参数，可按需扩展
```

### 参数说明

- `--tag`: 可选参数，指定镜像标签；未提供时自动生成随机字符串

## 工作流程

1. 更新 `deployment.yaml` 中 `code-repository` 容器的镜像标签为 `localhost:31500/code-repository-service:<tag>`
2. 构建镜像
3. 推送镜像到仓库
4. 使用 Helm 安装或升级到 Kubernetes
5. 检查 Pod/Service 状态
6. 获取服务日志

## 注意事项

- 确保已安装 Docker、kubectl 和 Helm，并已配置到 PATH
- 确保当前用户有权限执行 Docker 与 kubectl 命令
- 确保 Kubernetes 集群可访问且 `kubectl` 已配置好上下文
- 需要在 `deploy/xcoding/templates/services` 下提供 `code_repository/deployment.yaml` 与 `code_repository/service.yaml`，容器名约定为 `code-repository`