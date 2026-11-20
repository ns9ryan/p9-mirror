# Nginx Manager 项目

## 项目结构

```
nginx-manager/
├── Dockerfile              # Docker 构建文件
├── Makefile               # 构建和发布脚本
├── main.go                # 主程序（启动 nginx + reload API）
├── go.mod                 # Go 模块定义
├── entrypoint.sh          # 容器启动脚本
├── nginx.conf             # Nginx 主配置文件
├── docker-compose.yml     # 本地测试 docker-compose
├── conf.d/                # Nginx 配置目录
│   ├── node01/            # Node-Agent-01 生成的配置
│   ├── node02/            # Node-Agent-02 生成的配置
│   └── node03/            # Node-Agent-03 生成的配置
└── internal/
    └── api/
        └── reload.go      # Reload API 处理器
```

## 功能特性

✅ **Nginx + Go 集成**
- 在单个容器中运行 nginx 和 reload API
- 无需安装额外依赖

✅ **Reload API 接口**
- `GET /internal/nginx/reload` - 验证配置并重载 nginx
- `GET /health` - 健康检查端点

✅ **配置管理**
- 支持按节点目录组织配置（node01、node02、node03）
- Node-Agent 自动生成 `node{XX}/domain.conf`

✅ **健康检查**
- 内置 Docker healthcheck
- 每 30 秒检查一次 API 端点

## 使用方法

### 本地构建

```bash
cd nginx-manager

# 构建镜像
make docker

# 本地测试
make test-local

# 清理镜像
make clean
```

### 发布到仓库

```bash
# 发布到内部 Docker 仓库
make publish-docker
```

### 服务器部署

编辑服务器上的 `docker-compose.yml`：

```yaml
services:
  nginx:
    image: 192.168.0.15:5000/p9/nginx-manager:1.0.0
    container_name: p9-nginx
    restart: always
    networks:
      - p9-network
    ports:
      - "80:80"
      - "8888:8888"  # Reload API
    volumes:
      - ./conf.d:/etc/nginx/conf.d:ro
      - /home/alan/headquarter:/var/www/headquarter:ro
      - /home/alan/outstation:/var/www/outstation:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8888/health"]
      interval: 30s
      timeout: 5s
      retries: 3
```

然后启动：

```bash
docker pull 192.168.0.15:5000/p9/nginx-manager:1.0.0
docker-compose up -d
```

## API 接口

### 验证并重载 Nginx

```bash
curl -X GET http://localhost:8888/internal/nginx/reload
```

**响应示例：**

```json
{
  "success": true,
  "validate": "nginx: the configuration file /etc/nginx/nginx.conf syntax is ok\nnginx: configuration file /etc/nginx/nginx.conf test is successful",
  "reload": ""
}
```

### 健康检查

```bash
curl http://localhost:8888/health
```

**响应示例：**

```json
{
  "status": "ok"
}
```

## Node-Agent 集成

Node-Agent 生成的配置文件结构：

```
conf.d/
└── node01/
    ├── test01.com.conf
    ├── test02.com.conf
    └── ...
```

通过调用以下接口重载 nginx：

```go
url := "http://nginx-manager:8888/internal/nginx/reload"
resp, err := http.Get(url)
```

## 调试

### 进入容器

```bash
docker exec -it nginx-manager sh
```

### 查看日志

```bash
# 容器日志
docker logs nginx-manager

# Nginx 错误日志
docker exec nginx-manager cat /var/log/nginx/error.log
```

### 测试 API

```bash
docker exec nginx-manager curl http://localhost:8888/health
docker exec nginx-manager curl http://localhost:8888/internal/nginx/reload
```

### 验证 Nginx 配置

```bash
docker exec nginx-manager nginx -t
```

## 故障排查

| 问题 | 解决方案 |
|------|--------|
| Reload API 无法连接 | 检查 8888 端口是否暴露，容器是否正常运行 |
| Nginx 配置验证失败 | 检查 conf.d 目录下的配置文件语法 |
| Healthcheck 失败 | 检查容器内 curl 和 API 服务是否正常 |
| 配置未生效 | 确认配置文件在 /etc/nginx/conf.d/ 下，然后调用 reload API |

## 许可

oa.98ent.com/p9/nginx-manager
