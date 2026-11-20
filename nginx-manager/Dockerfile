# Stage 1: Build Go application
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o reload-api .

# Stage 2: Runtime
FROM nginx:1.27-alpine

# 安装必要工具（curl 用于健康检查）
RUN apk add --no-cache curl

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf
COPY conf.d/ /etc/nginx/conf.d/

# 复制编译好的 Go 应用
COPY --from=builder /build/reload-api /usr/local/bin/

# 启动脚本
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8888/health || exit 1

ENTRYPOINT ["/entrypoint.sh"]
