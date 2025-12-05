# 多阶段构建Dockerfile
# 第一阶段：构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
# RUN apk add --no-cache git ca-certificates tzdata

# 复制整个项目的go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制整个仓库（提高一致性，包含共享包与生成代码）
COPY . .

# 构建code_repository服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/code-repository-service ./apps/code_repository/cmd

# ----------------------------------------------------------------------------------------------------
# 第二阶段：运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
# RUN apk --no-cache add ca-certificates tzdata wget curl

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/code-repository-service .

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口（HTTP网关和gRPC）
EXPOSE 50053 10053

# 启动服务
ENTRYPOINT ["./code-repository-service"]