# 后端 Dockerfile - 多阶段构建
# 阶段 1: 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 设置 Go 代理（加速模块下载）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# 复制依赖文件（优先复制，利用 Docker 层缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
# CGO_ENABLED=0 禁用 CGO，生成静态二进制（不依赖 glibc）
# GOOS=linux 指定目标平台
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags '-w -s' \
    -o bin/server \
    ./cmd/server

# 阶段 2: 运行阶段（使用最小化镜像）
FROM alpine:3.18

# 安装 ca-certificates（HTTPS 请求需要）和 tzdata（时区支持）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为中国标准时间
ENV TZ=Asia/Shanghai

# 创建非 root 用户（安全最佳实践）
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/server .

# 更改文件所有者
RUN chown appuser:appgroup /app/server

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# 启动命令
CMD ["./server"]
