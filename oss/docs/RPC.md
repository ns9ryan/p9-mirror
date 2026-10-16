# 多租户初始化（gRPC）

外面新建一个租户时调这条。**不要**走管理后台的创建/更新服务商。

`rpc InitTenant(InitTenantReq) returns (Provider)`

请求只传 `operator_code`（body 字段，不是登录账号里的）。服务端把 `rpc/etc/init_default_settings.yaml`（oss 连接和桶）写成该租户一条默认服务商。已经有了再调，原样返回，不覆盖。scene 目录和上传大小限制看 `api/etc/oss.yaml`，不进这份模板。

本机先起依赖和 RPC（HTTP `go run ./api` 调不到这条）：

```powershell
cd oss
docker compose up -d
go run ./rpc
```

地址：`127.0.0.1:8021`（`rpc.listen_on`）。未配库：`503` `database not configured`。

| | |
| --- | --- |
| 请求 | `{ "operator_code": "op-xxx" }` |
| 返回 | `Provider`（刚写入或已有的默认配置）。`secret_key` 不返回 |
| 失败 | 没传：`InvalidArgument` `operator_code is required`。未配库：`503` `database not configured` |

### 请求 `InitTenantReq`

| 字段 | 必填 | 类型 | 说明 |
| --- | --- | --- | --- |
| `operator_code` | 是 | string | 要初始化的租户编码 |

```json
{ "operator_code": "op-xxx" }
```

`operator_code` 只放请求体，不必再挂 `x-operator-code` metadata。

### 返回

字段来自 `init_default_settings.yaml`，对接方不用传存储配置。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | uint64 | 本租户这条服务商 id |
| `name` | string | 固定 `default` |
| `is_default` | bool | `true` |
| `driver` | string | 如 `s3` |
| `endpoint` | string | 模板里的 S3 API |
| `file_url` | string | 公开读前缀 |
| `access_key` | string | 模板访问密钥 |
| `secret_key` | string | 空（脱敏） |
| `path_style` | bool | MinIO 一般为 `true` |
| `region` | string | 如 `default` |
| `buckets` | map | 桶名 → `{ mode, sign_ttl }` |

```json
{
  "id": 1,
  "name": "default",
  "is_default": true,
  "driver": "s3",
  "endpoint": "http://127.0.0.1:9200",
  "file_url": "http://127.0.0.1:9200",
  "access_key": "minioadmin",
  "secret_key": "",
  "path_style": true,
  "region": "default",
  "buckets": {
    "oss-bucket": { "mode": "public", "sign_ttl": 0 }
  }
}
```

### Demo（grpcurl）

本机 `Mode=dev`，开了 reflection。先装客户端（只需一次）：

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

可执行文件在 `%USERPROFILE%\go\bin\grpcurl.exe`。若提示找不到命令，把该目录加入 PATH，或写全路径。

PowerShell 会拆掉 JSON 引号，**不要**把 `-d '{...}'` 写在同一行。从 stdin 传入：

```powershell
'{"operator_code":"op-zm"}' | grpcurl -plaintext -d '@' 127.0.0.1:8021 oss.Oss/InitTenant
```

同样 `operator_code` 再调一次不会覆盖，返回同一条。

### Demo（Go）

`ossclient` 是本仓库生成的 gRPC 客户端：`oa.98ent.com/p9/oss/rpc/ossclient`（源：`rpc/oss.proto`）。其它服务在 `go.mod` 里：

```text
require oa.98ent.com/p9/oss v0.0.0
```

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"oa.98ent.com/p9/oss/rpc/ossclient"

	"github.com/zeromicro/go-zero/zrpc"
)

func main() {
	cli := ossclient.NewOss(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{"127.0.0.1:8021"},
		NonBlock:  true,
		Timeout:   10000, // 毫秒
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := cli.InitTenant(ctx, &ossclient.InitTenantReq{
		OperatorCode: "op-xxx",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("id=%d name=%s endpoint=%s buckets=%v\n",
		out.GetId(), out.GetName(), out.GetEndpoint(), out.GetBuckets())
}
```
