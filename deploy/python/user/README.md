# 用户服务部署脚本

这个Python脚本用于管理用户服务的部署，包括读取YAML配置、修改镜像标签、构建镜像、推送和部署到Kubernetes。

## 功能

- 读取和修改values.yaml中的镜像标签
- 构建Docker镜像
- 推送镜像到仓库
- 使用Helm部署或升级服务到Kubernetes
- 检查部署状态和获取日志

## 安装依赖

```bash
pip install -r requirements.txt
```

## 使用方法

### 基本用法

```bash
# 部署指定标签的镜像（包括构建、推送和部署）
python deploy.py 13

# 部署指定标签的镜像（仅构建和部署，跳过推送）
python deploy.py 13 --no-push

# 指定项目根目录
python deploy.py 13 --project-root /path/to/project
```

### 参数说明

- `tag`: 必需参数，指定镜像标签
- `--no-push`: 可选参数，跳过镜像推送步骤（用于本地部署）
- `--project-root`: 可选参数，指定项目根目录路径，默认为/home/hr/xcoding

## 示例

```bash
# 完整部署流程（构建、推送、部署）
python deploy.py 14

# 本地部署流程（构建、部署，不推送）
python deploy.py 14 --no-push
```

## 工作流程

1. 更新values.yaml中的用户服务镜像标签
2. 构建Docker镜像
3. 推送镜像到仓库（除非指定--no-push）
4. 使用Helm部署或升级服务到Kubernetes
5. 检查部署状态
6. 获取用户服务日志

## 注意事项

- 确保已安装Docker、kubectl和Helm
- 确保当前用户有权限执行Docker和kubectl命令
- 确保Kubernetes集群可访问且已配置好kubectl