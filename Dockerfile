# ============ Tortoise Server Dockerfile ============
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates

# 复制 go mod 文件
COPY server/go.mod server/go.sum ./
RUN go mod download

# 复制源代码
COPY server ./server

# 构建
WORKDIR /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tortoise-server ./cmd/api

# ============ 生产镜像 ============
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates curl

# 复制二进制文件
COPY --from=builder /app/server/tortoise-server .

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 18792

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:18792/api/v1/health || exit 1

# 运行
CMD ["./tortoise-server"]
