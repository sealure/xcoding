# 前端服务部署脚本

这个Python脚本用于管理前端服务的部署，包括读取Helm模板、修改镜像标签、构建镜像、推送和部署到Kubernetes。

## 功能

- 修改 Helm 模板中的镜像标签
- 构建 Docker 镜像
- 推送镜像到私有仓库
- 使用 Helm 部署或升级到 Kubernetes
- 检查部署状态与获取日志

## 安装依赖

```bash
pip install -r requirements.txt
```

## 使用方法

```bash
# 构建、推送并部署
python deploy.py --tag v1

# 仅构建和部署（跳过推送）
python deploy.py --tag v1 --no-push
```

## 注意事项

- 需要本地安装 Docker、kubectl、Helm
- APISIX 网关已通过 NodePort 暴露：`31080`
- 前端通过 Nginx 监听 `80` 端口（容器内），Service 映射见 Helm 模板