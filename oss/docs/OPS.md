# oss 运维与部署

HTTP 见 [API.md](API.md)，其它服务 gRPC 见 [RPC.md](RPC.md)。总览见 [../README.md](../README.md)。

---

## 环境变量

本机开发可 `go run ./rpc` 与 `go run ./api`，不读这些变量。容器不传 `-f` 且已设 `ETCD_CONFIG_KEY` 时读 etcd。

| 变量 | 示例 | 说明 |
| --- | --- | --- |
| `ETCD_ENDPOINTS` | `127.0.0.1:2379` | Etcd 地址，逗号分隔 |
| `ETCD_CONFIG_KEY` | `/oss` | 配置前缀，实际读取 `{key}/service` |
| `ETCD_USERNAME` / `ETCD_PASSWORD` | | 可选 |
| `POSTGRES_DSN` | `host=127.0.0.1 port=5433 user=root password=root dbname=oss sslmode=disable` | 可选，覆盖 yaml `database.dsn` |
| `OSS_ENDPOINT` | `http://host.docker.internal:9200` | 可选，覆盖 yaml `oss.endpoint`（服务端连存储） |
| `OSS_FILE_URL` | `http://127.0.0.1:9200` | 可选，覆盖 yaml `oss.file_url`（浏览器打开文件） |

---

## 配置（yaml / etcd `/oss/service`）

整份配置分五块：HTTP 端口、可选数据库、上传规则、存储连接、可选本服务 gRPC。字段用 **snake_case**。旧版小驼峰（`fileUrl`、`pathstyle` 等）仍能解析。

优先级：**库里 `isDefault=true` 的服务商 > etcd** `${ETCD_CONFIG_KEY}/service` **>** 本地 yaml。etcd `PUT` 后 watch 热更新 `upload` / `oss`；`port` / `rpc.listen_on` **不会改监听**，改端口必须重启。校验失败保留上一份可用配置。

没有默认服务商时，上传和 `GET /provider/active` 都走 API 进程里的 `oss`（`api/etc/oss.yaml` 或 etcd）。

HTTP 用 `api/etc/oss.yaml`（`port` / `upload` / `oss` / `oss_rpc`）。RPC 用 `rpc/etc/oss.yaml`（只用 `database` / `rpc.listen_on`）；默认驱动在 [`rpc/etc/init_default_settings.yaml`](../rpc/etc/init_default_settings.yaml)，`InitTenant` 时写入库。etcd 仍是一份：两个进程各读自己用到的字段。

```yaml
port: "8020"
cros_conf:
  address: '*'
# database:
#   dsn: "host=127.0.0.1 port=5433 user=root password=root dbname=oss sslmode=disable"
upload:
  max_size_mb: 32
  multipart_memory_mb: 32
  default_scene: public
  scenes:
    public:
      visibility: public
oss:
  driver: s3
  endpoint: "http://127.0.0.1:9200"
  file_url: "http://127.0.0.1:9200"
  buckets:
    oss-bucket:
      mode: public
  access_key: minioadmin
  secret_key: minioadmin
  path_style: true
rpc:
  listen_on: 0.0.0.0:8021
```

| 字段 | 默认 | 说明 |
| --- | --- | --- |
| `port` | | HTTP 监听。容器内一般 `80`，本机 `go run` 建议 `8020` |
| `cros_conf.address` | `*` | 浏览器 CORS Origin。`*` 允许全部。旧 key `CROSConf` 仍能解析 |
| `rpc.listen_on` | 空=不启动 | 对外 gRPC。本机建议 `0.0.0.0:8021`。改完需重启 |
| `database.dsn` | 空 | 可选。配了才写档案、才能管服务商。可被 `POSTGRES_DSN` 覆盖 |
| `upload.max_size_mb` | `32` | `POST /file/upload` 单文件上限（MB） |
| `upload.multipart_memory_mb` | `32` | multipart 内存阈值，超出落临时文件 |
| `upload.presign_max_size_mb` | `2048` | 预签名直传上限（MB） |
| `upload.presign_ttl` | `600` | 预签名 PUT 有效秒数 |
| `upload.default_scene` | `public` | 请求未带 `scene` 时。敏感上传不要依赖此默认值 |
| `upload.scenes` | 空=自动 `public`（public 桶） | **scene → `{ visibility }`**。未列入的请求失败。目前默认只配 `public`；可加 `idcard: { visibility: private }` |
| `upload.allowed_extensions` | 空=全部允许 | 可写 `.jpg` 或 `jpg` |
| `oss.driver` | `s3` | `s3` / `minio` |
| `oss.endpoint` | | S3 API。AWS 官方 S3 可留空 |
| `oss.file_url` | | 公开读 URL 前缀（CDN）。和 `path_style` 一起决定浏览器打开的地址 |
| `oss.buckets.<name>.mode` | 必填 | `public` / `private`。目前默认只配公共桶 |
| `oss.buckets.<name>.sign_ttl` | `600` | 仅 private：预签名 GET 秒数 |
| `oss.access_key` / `oss.secret_key` | | 存储凭证 |
| `oss.path_style` | `false` | API 与公开链。MinIO 设 `true`（`{file_url}/{bucket}/{key}`）；CDN 通常 `false`（`{file_url}/{key}`） |
| `oss.region` | `default` | AWS 需真实 region |

`upload` 可整段省略。未配 `scenes` 时自动 `public`，对齐到第一个 `mode: public` 的桶。建桶 / ACL / CDN 由运维配置，服务不改存储策略。MinIO 建用户与桶见 [`../common/oss/oss命令.md`](../common/oss/oss命令.md)。

`oss_provider` 只存驱动（连接 + 桶）。scene 目录不进库；启用服务商后按桶 `mode` 重新对齐。etcd 若仍是旧写法「桶名 → scene 列表」，需改成上面的 `visibility`。

### 云 + CDN

```yaml
port: "80"
oss:
  endpoint: "https://oss-cn-hongkong.aliyuncs.com"
  file_url: "https://cdn.example.com"
  region: "oss-cn-hongkong"
  path_style: false
```

`path_style: false` 时公开链是 `{file_url}/{key}`（CDN 绑在桶上）。MinIO 直出则 `path_style: true`，链是 `{file_url}/{bucket}/{key}`。`mode: private` 时 `Put` / `URL` 自动预签名 GET。

---

## 本地开发

- Go 1.25+，或只用 Docker
- 本地 MinIO（S3 API）
- 读 yaml 时不需要 Etcd
- 档案 / 服务商：可选 PostgreSQL

### Docker + 本机 Go（推荐，不需要 Harbor）

Docker 只起 PostgreSQL 和 MinIO；服务用本机 `go run`：

```powershell
cd oss
docker compose -f deploy/docker-compose.yml up -d
go run ./rpc
go run ./api
```

| 地址 | 说明 |
| --- | --- |
| http://127.0.0.1:8020/p9/oss/file/upload | 上传 |
| `127.0.0.1:8021` | gRPC `InitTenant`（新建租户套默认模板） |
| http://127.0.0.1:9201 | MinIO 控制台（`minioadmin` / `minioadmin`） |
| `127.0.0.1:5433` | PostgreSQL（`root` / `root`，库名 `oss`） |

配置在 `api/etc/oss.yaml`、`rpc/etc/oss.yaml`。停掉依赖：`docker compose down`；清数据：`docker compose down -v`。

### 启动 MinIO（仅 `go run` 时需要）

```powershell
docker rm -f minio 2>$null
docker run -d --name minio -p 9200:9000 -p 9201:9001 `
  -e MINIO_ROOT_USER=minioadmin `
  -e MINIO_ROOT_PASSWORD=minioadmin `
  minio/minio server /data --console-address ":9001"
```

控制台：[http://127.0.0.1:9201](http://127.0.0.1:9201)（`minioadmin` / `minioadmin`）。当前命令 **未挂数据卷**。先在控制台创建配置里的 bucket（服务不会自动建桶）。

写入 etcd（可选，走 etcd 启动时才需要）：

```powershell
@"
port: `"8020`"
upload:
  default_scene: public
  scenes:
    public:
      visibility: public
oss:
  endpoint: `"http://127.0.0.1:9200`"
  file_url: `"http://127.0.0.1:9200`"
  buckets:
    oss-bucket:
      mode: public
  access_key: minioadmin
  secret_key: minioadmin
  path_style: true
"@ | docker exec -i company-etcd-1 etcdctl put /oss/service
```

### 启动服务

**本地文件（推荐）：** 改 yaml 后需重启。

```powershell
cd oss
go run ./rpc
go run ./api
```

**etcd：** 设置 `ETCD_CONFIG_KEY` 后走 etcd（优先于本地 yaml），读 `${ETCD_CONFIG_KEY}/service`，配置变更会热更新。两个进程都要起。

```powershell
cd oss
$env:ETCD_ENDPOINTS="127.0.0.1:2379"
$env:ETCD_CONFIG_KEY="/oss"
go run ./rpc
go run ./api
```

---

## 上线

日常发版走 Jenkins（`Jenkinsfile`）：选分支构建，推两个镜像并触发部署任务。

| | |
| --- | --- |
| 镜像 | `harbor.ecosyssoft.com/p9/oss-api`、`.../oss-rpc`，同一 tag（构建号） |
| 编排 | [`deploy/docker-compose.alpha.yml`](../deploy/docker-compose.alpha.yml) |
| 配置 | etcd `${ETCD_CONFIG_KEY}/service`（compose 已设 `ETCD_ENDPOINTS`、`ETCD_CONFIG_KEY=/oss`） |
| HTTP | 宿主机 `8020` → 容器 `80`（etcd 里 `port: "80"`） |
| gRPC | 宿主机 `8021` → rpc 容器 `8021` |

etcd 里 API 访问本服务 RPC 不要写 `127.0.0.1`（已拆成两个容器），写成：

```yaml
port: "80"
oss_rpc:
  endpoints:
    - p9-oss-rpc:8021
rpc:
  listen_on: 0.0.0.0:8021
```

部署任务注入 `IMAGE_TAG_OSS` 后，在目标机：

```bash
docker login harbor.ecosyssoft.com
IMAGE_TAG_OSS=<构建号> docker compose -f deploy/docker-compose.alpha.yml up -d
```

本机不要用这套 compose。本地只起 Postgres/MinIO：`docker compose -f deploy/docker-compose.yml up -d`，再 `go run ./rpc`、`go run ./api`。

---

## 桶 CORS

浏览器预签名 PUT 打到存储主机，必须在**桶**上配 CORS（本服务的 CORS 管不到 MinIO / 云域名）。

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "PUT", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3000
    }
  ]
}
```

```bash
mc cors set myminio/oss-bucket cors.json
```

---

## 排错

| 现象 | 原因 |
| --- | --- |
| `ETCD_CONFIG_KEY 环境变量未设置` | 空 `-f` 或配置文件不存在，且未导出 etcd 变量 |
| `配置文件不存在` | `-f` 路径不对；在 oss 目录执行 |
| `etcd 中未找到配置` | 未 put `/oss/service` |
| `Forbidden` / 403 | AccessKey 与 MinIO 不一致，或 `path_style` 未开 |
| `invalid reference format` | compose 未设镜像 tag（已默认 `latest`） |
| 拉镜像 `no basic auth credentials` | 未登录 Harbor |
| 端口已被占用 | MinIO 避开 `9000` / `9001` / `9100` |
| 上传 404 / NoSuchBucket | 运维未预先创建 `upload.scenes` 映射到的桶 |
| `invalid scene` | 前端传了未配置的枚举，或传了桶名 |
| `/file` 或 `/provider` 返回 503 | 未配 `database.dsn` / `POSTGRES_DSN` |
| 上传成功但列表没有记录 | 未配库；或大文件只 PUT 了存储、没调 complete |
| 业务表只存了访问 `url`，换桶或私有链过期后打不开 | 应存 `key`，展示时 `GET /url/detail?key=` |
| 地址栏打开 `putUrl` 得到 `AccessDenied` | 地址栏是 GET，签名是 PUT |
| complete 报 `object not found` | 还没 PUT，或 complete 的 `scene` 与第 1 步不一致 |
| `SignatureDoesNotMatch` | `Content-Type` / `Content-Length` 与签发不一致，或 `size` 不是实际文件大小 |
| 网页 PUT 被浏览器拦 | 桶未配 CORS |
