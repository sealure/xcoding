# 用户服务统一服务器

这个服务器将原来的gRPC服务器和HTTP网关服务器合并为一个统一的服务器，可以同时提供gRPC和HTTP接口。

## 功能

- 同时提供gRPC和HTTP接口
- 统一的配置管理
- 统一的优雅关闭处理

## 使用方法

### 编译

```bash
go build -o bin/server ./cmd/server
```

### 运行

```bash
./bin/server
```

### 配置

服务器使用`.env`文件进行配置，配置项包括：

- 数据库配置
- gRPC配置
- HTTP配置
- 日志配置
- JWT配置
- 监控配置

## 端口

- gRPC服务：默认端口50051
- HTTP网关服务：默认端口10051

## API接口

### HTTP接口

- 用户注册：POST /api/v1/users/register
- 用户登录：POST /api/v1/users/login
- 获取用户信息：GET /api/v1/users/{user_id}
- 更新用户信息：PUT /api/v1/users/{user_id}
- 列出用户：GET /api/v1/users
- 删除用户：DELETE /api/v1/users/{id}
- 认证：POST /api/v1/auth
- 创建API令牌：POST /api/v1/tokens
- 列出API令牌：GET /api/v1/tokens
- 删除API令牌：DELETE /api/v1/tokens/{token_id}

### gRPC接口

gRPC接口定义在`proto/user/v1/user.proto`文件中，可以使用gRPC客户端直接调用。

## 健康检查

- gRPC 健康检查：使用标准 `grpc.health.v1.Health` 服务
  - 示例：`grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check`

## 优雅关闭

服务器支持优雅关闭，当收到SIGINT或SIGTERM信号时，会先停止接收新的请求，然后等待现有请求处理完成后关闭。