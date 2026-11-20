FROM alpine:3.22

# 镜像作者
ARG AUTHOR="p9.work@gmail.com"

LABEL org.opencontainers.image.authors=${AUTHOR}

WORKDIR /app

# 设置容器时区并安装运行依赖
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata ca-certificates && mkdir -p /app/etc

# 复制节点代理可执行文件
COPY ./node_agent ./
RUN chmod +x ./node_agent

ENTRYPOINT ["./node_agent"]
CMD ["-f", "etc/node_agent.yaml"]