# 用户服务

用户服务是一个基于gRPC和HTTP的微服务，提供完整的用户管理功能，包括用户注册、登录、信息管理、权限控制等。

## 功能特性

- 用户注册和登录
- 用户信息管理
- 基于角色的访问控制
- JWT令牌认证
- API令牌管理
- gRPC和HTTP双协议支持
- 完整的单元测试和集成测试
- 结构化日志和监控指标
- GORM ORM支持
- Atlas数据库迁移管理

## 技术栈

- Go 1.19+
- gRPC
- gRPC-Gateway
- PostgreSQL
- GORM (ORM)
- Atlas (数据库迁移)
- JWT认证
- Redis (缓存)
- Docker
- Prometheus (监控)
- Grafana (可视化)

## 项目结构

```
apps/user/
├── cmd/server/           # 主程序入口
│   └── main.go
├── internal/             # 内部实现
│   ├── config/          # 配置管理
│   ├── db/              # 数据库连接
│   │   ├── db.go        # 原始数据库连接
│   │   └── gorm.go      # GORM数据库连接
│   ├── gateway/         # HTTP网关
│   ├── middleware/      # 中间件
│   ├── repository/      # 数据访问层
│   │   ├── user_repository.go       # 用户仓储接口
│   │   └── user_repository_gorm.go  # GORM用户仓储实现
│   ├── server/          # gRPC服务器
│   └── service/         # 业务逻辑层
├── models/              # GORM数据模型
│   └── user.go
├── pkg/                 # 公共包
│   ├── auth/            # 认证工具
│   └── validator/       # 验证工具
├── tests/               # 测试
│   ├── integration/     # 集成测试
│   └── unit/            # 单元测试
├── scripts/             # 脚本
│   └── init.sql         # 数据库初始化脚本
├── docs/                # 文档
│   └── api.md
├── .env.example         # 环境变量示例
├── .gitignore
├── atlas.hcl            # Atlas数据库迁移配置
├── docker-compose.yml   # Docker Compose配置
├── Dockerfile
├── go.mod
├── Makefile             # 构建和开发脚本
└── README.md
```

## 快速开始

### 环境要求

- Go 1.19+
- PostgreSQL 12+
- Redis 6+
- Protocol Buffers compiler (protoc)
- gRPC plugins for protoc
- Atlas CLI (用于数据库迁移)
- Docker & Docker Compose

### 安装依赖

```bash
# 安装Go依赖
make install

# 或者手动安装
go mod download

# 安装protoc插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.11.0
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.11.0

# 安装Atlas CLI (如果尚未安装)
curl -sSf https://atlasgo.io/install.sh | sh
```

### 配置环境变量

```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑 .env 文件，设置数据库连接和其他配置
# 主要配置项包括：
# - GRPC_ADDRESS: gRPC服务器地址
# - HTTP_ADDRESS: HTTP服务器地址
# - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME: 数据库配置
# - REDIS_URL: Redis连接URL
# - JWT_SECRET: JWT签名密钥
# - LOG_LEVEL: 日志级别
```

### 数据库设置

1. 启动数据库服务：
```bash
# 使用Docker Compose启动PostgreSQL和Redis
make db-up

# 或者手动启动PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_DB=xcoding \
  -e POSTGRES_USER=xcoding \
  -e POSTGRES_PASSWORD=xcoding \
  -p 5432:5432 \
  postgres:13
```

2. 运行数据库迁移：
```bash
# 应用数据库迁移
make db-migrate-up

# 或者使用Atlas直接应用
atlas migrate apply --env local
```

### 运行服务

```bash
# 生成protobuf代码
make proto

# 运行服务
make run

# 或者直接运行
go run cmd/server/main.go
```

服务将在以下端口启动：
- gRPC: `:9090`
- HTTP: `:8080`
- 健康检查: `:8080/health`
- 指标: `:8080/metrics`

## API文档

详细的API文档请参考 [API文档](docs/api.md)。

## 测试

### 运行单元测试

```bash
go test ./tests/unit/...
```

### 运行集成测试

```bash
go test ./tests/integration/...
```

### 运行所有测试

```bash
go test ./...
```

## Docker部署

### 构建镜像

```bash
docker build -t user-service .
```

### 运行容器

```bash
docker run -d \
  --name user-service \
  -p 50051:50051 \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://user:password@localhost/user_service \
  -e JWT_SECRET=your-secret-key \
  user-service
```

### Docker Compose

使用Docker Compose启动所有服务：

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:13
    container_name: postgres
    environment:
      POSTGRES_DB: xcoding
      POSTGRES_USER: xcoding
      POSTGRES_PASSWORD: xcoding
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U xcoding"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:6-alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  user-service:
    build: .
    container_name: user-service
    ports:
      - "50051:50051"   # gRPC
      - "10051:10051"   # HTTP Gateway
    environment:
      - GRPC_ADDRESS=0.0.0.0
      - GRPC_PORT=50051
      - HTTP_ADDRESS=0.0.0.0
      - HTTP_PORT=10051
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=xcoding
      - DB_PASSWORD=xcoding
      - DB_NAME=xcoding
      - DB_SSLMODE=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=your-secret-key
      - LOG_LEVEL=info
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "grpcurl", "-plaintext", "localhost:50051", "grpc.health.v1.Health/Check"]
      interval: 30s
      timeout: 10s
      retries: 3

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    depends_on:
      - user-service

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    depends_on:
      - prometheus

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

启动所有服务：

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f user-service

# 停止所有服务
docker-compose down
```

## 开发指南

### 代码生成

```bash
# 生成protobuf代码
make proto

# 格式化代码
make fmt

# 运行linter
make lint

# 运行测试
make test
```

### 数据库迁移

使用Atlas进行数据库迁移：

```bash
# 创建新的迁移文件
make db-migrate

# 应用迁移
make db-migrate-up

# 回滚迁移
make db-migrate-down

# 重置数据库
make db-reset

# 检查数据库模式
make db-inspect

# 比较模式差异
make db-diff
```

Atlas配置文件 `atlas.hcl` 定义了数据库连接和迁移设置：

```hcl
env "local" {
  url = "postgres://xcoding:xcoding@localhost:5432/xcoding?sslmode=disable"
  src = "file://models"
  dev = "docker://postgres/15"
}
```

### GORM模型

GORM模型定义在 `models/` 目录中，例如 `models/user.go`：

```go
package models

import (
    "time"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// User 用户模型
type User struct {
    ID        uint64         `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"uniqueIndex;not null" json:"username"`
    Email     string         `gorm:"uniqueIndex;not null" json:"email"`
    Avatar    string         `json:"avatar"`
    Role      UserRole       `gorm:"type:varchar(50);not null;default:'USER_ROLE_USER'" json:"role"`
    IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}

// APIToken API令牌模型
type APIToken struct {
    ID          uint64      `gorm:"primaryKey" json:"id"`
    UserID      uint64      `gorm:"not null;index" json:"user_id"`
    Name        string      `gorm:"not null" json:"name"`
    Token       string      `gorm:"uniqueIndex;not null" json:"token"`
    Description string      `json:"description"`
    Scopes      []Scope     `gorm:"type:text[]" json:"scopes"`
    ExpiresAt   *time.Time  `json:"expires_at"`
    CreatedAt   time.Time   `json:"created_at"`
    
    // 关联
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

### 添加新API

1. 在`proto/user/v1/user.proto`中定义新的服务和方法
2. 运行`make proto`生成代码
3. 在`models/`中定义GORM模型（如果需要新的数据模型）
4. 在`internal/repository/`中实现数据访问层
5. 在`internal/service/`中实现业务逻辑
6. 在`internal/server/`中实现新的gRPC方法
7. 添加相应的测试

### 添加新中间件

1. 在`internal/middleware/`中实现新的中间件函数
2. 在`cmd/server/main.go`中注册新的中间件

### 环境变量

主要环境变量：

```bash
# 服务器配置
GRPC_ADDRESS=:9090
HTTP_ADDRESS=:8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=xcoding
DB_PASSWORD=xcoding
DB_NAME=xcoding
DB_SSLMODE=disable

# Redis配置
REDIS_URL=redis://localhost:6379

# JWT配置
JWT_SECRET=your-secret-key

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# 限流配置
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# CORS配置
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*

# 监控配置
METRICS_ENABLED=true
METRICS_PATH=/metrics

# 健康检查配置
HEALTH_CHECK_ENABLED=true
HEALTH_CHECK_PATH=/health
```

## 监控和日志

服务支持以下监控和日志功能：

- 结构化日志输出
- gRPC拦截器记录请求和响应
- Prometheus指标收集
- 健康检查端点

### 健康检查

```bash
# gRPC健康检查
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# HTTP健康检查
curl http://localhost:8080/healthz
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发送邮件至 [your-email@example.com]

## 更新日志

### v1.0.0 (2023-01-01)

- 初始版本发布
- 实现基本的用户管理功能
- 支持gRPC和HTTP双协议
- 添加JWT认证
- 完整的测试覆盖

### v1.1.0 (2023-02-01)

- 添加API令牌管理功能
- 增强权限控制
- 改进日志记录
- 添加Prometheus指标

### v1.2.0 (2023-03-01)

- 增强认证接口，支持APISIX forward-auth插件
- 添加批量操作支持
- 优化性能
- 修复已知问题