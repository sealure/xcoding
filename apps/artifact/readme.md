# Artifact Service

Artifact Service 是一个用于管理容器镜像仓库、命名空间、仓库和标签的服务。它提供了完整的容器镜像元数据管理功能，包括注册表管理、命名空间管理、仓库管理和标签管理。

## 功能特性

- **注册表管理**：创建、查询、更新和删除容器镜像注册表
- **命名空间管理**：管理注册表下的命名空间
- **仓库管理**：管理命名空间下的镜像仓库
- **标签管理**：管理仓库下的镜像标签
- **镜像操作**：获取镜像清单和镜像层数据
- **RESTful API**：通过HTTP网关提供RESTful API接口
- **gRPC服务**：提供高性能的gRPC接口
- **健康检查**：使用标准 gRPC 健康检查服务

## 技术栈

- Go 1.21
- gRPC & gRPC Gateway
- PostgreSQL
- GORM
- Docker

## 目录结构

```
artifact/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序文件
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   │   └── config.go
│   ├── db/                # 数据库连接
│   │   ├── db.go
│   │   └── gorm.go
│   ├── gateway/           # HTTP网关
│   │   └── artifact_gateway.go
│   ├── grpc/              # gRPC处理
│   │   ├── handler/
│   │   │   └── artifact_handler.go
│   │   └── interceptors/
│   │       └── interceptors.go
│   ├── models/            # 数据模型
│   │   ├── registry.go
│   │   ├── namespace.go
│   │   ├── repository.go
│   │   └── tag.go
│   └── service/           # 业务逻辑
│       └── artifact_service.go
├── Dockerfile             # Docker构建文件
├── .env.example           # 环境变量示例
└── readme.md              # 项目说明
```

## API 接口

### 注册表管理

- `POST /v1/registries` - 创建注册表
- `GET /v1/registries/{id}` - 获取注册表详情
- `PUT /v1/registries/{id}` - 更新注册表
- `DELETE /v1/registries/{id}` - 删除注册表
- `GET /v1/registries` - 列出注册表

### 命名空间管理

- `POST /v1/namespaces` - 创建命名空间
- `GET /v1/namespaces/{id}` - 获取命名空间详情
- `PUT /v1/namespaces/{id}` - 更新命名空间
- `DELETE /v1/namespaces/{id}` - 删除命名空间
- `GET /v1/namespaces` - 列出命名空间

### 仓库管理

- `POST /v1/repositories` - 创建仓库
- `GET /v1/repositories/{id}` - 获取仓库详情
- `PUT /v1/repositories/{id}` - 更新仓库
- `DELETE /v1/repositories/{id}` - 删除仓库
- `GET /v1/repositories` - 列出仓库

### 标签管理

- `POST /v1/tags` - 创建标签
- `GET /v1/tags/{id}` - 获取标签详情
- `PUT /v1/tags/{id}` - 更新标签
- `DELETE /v1/tags/{id}` - 删除标签
- `GET /v1/tags` - 列出标签

### 镜像操作

- `GET /v1/images/{repository}/manifests/{reference}` - 获取镜像清单
- `GET /v1/images/{repository}/blobs/{digest}` - 获取镜像层数据

## 快速开始

### 使用Docker运行

1. 构建Docker镜像：

```bash
docker build -t artifact-service .
```

2. 运行容器：

```bash
docker run -d \
  --name artifact-service \
  -p 8080:8080 \
  -p 9090:9090 \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=your-db-name \
  artifact-service
```

### 本地开发

1. 克隆项目：

```bash
git clone <repository-url>
cd artifact
```

2. 安装依赖：

```bash
go mod download
```

3. 复制环境变量文件：

```bash
cp .env.example .env
```

4. 修改`.env`文件中的配置，特别是数据库连接信息。

5. 运行服务：

```bash
go run cmd/main.go
```

## 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| GRPC_HOST | gRPC服务主机 | 0.0.0.0 |
| GRPC_PORT | gRPC服务端口 | 9090 |
| HTTP_HOST | HTTP服务主机 | 0.0.0.0 |
| HTTP_PORT | HTTP服务端口 | 8080 |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户名 | postgres |
| DB_PASSWORD | 数据库密码 | password |
| DB_NAME | 数据库名称 | artifact_db |
| DB_SSLMODE | 数据库SSL模式 | disable |
| DB_TIMEZONE | 数据库时区 | UTC |
| LOG_LEVEL | 日志级别 | info |

## 健康检查

- gRPC 健康检查：标准 `grpc.health.v1.Health` 服务
  - 示例：`grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check`

## Docker Registry API 示例

以下是一些常用的Docker Registry API示例：

### 获取仓库列表
```bash
curl http://localhost:31500/v2/_catalog
```

### 获取标签列表
```bash
curl http://localhost:31500/v2/testbox/tags/list
```

### 获取镜像的digest
```bash
curl -I -H "Accept: application/vnd.docker.distribution.manifest.v2+json" http://localhost:31500/v2/testbox/manifests/latest
```

### 使用digest删除镜像
```bash
curl -X DELETE http://localhost:31500/v2/testbox/manifests/sha256:ebd0a6156be6cdc24c4bd1ad0e0fcb938cbdd5f8b86b15d18821f280902b6380
```

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。