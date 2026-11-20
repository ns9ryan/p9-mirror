## 规划存储时建议遵守以下三条：

1. **按「项目（或微服务）+ 环境」建桶**，例如 `oss-prod`。
2. **桶名全局唯一且规范化**，建议 `[公司/部门]-[项目]-[环境]`，例如 `mycompany-oss-prod`。
3. **不要把所有东西塞进一个桶**。前端静态页、用户身份证、系统日志混在同一 Bucket 里，后续做 CDN 或安全审计会很痛苦。

若同一服务既有公开资源（如商品图）又有私密文件（如用户合同），应拆成两个桶：`oss-public` 与 `oss-private`。

---

# MinIO 运维命令

为 `oss` 在 MinIO 上创建**专属用户、专属桶、最小权限策略**。完成后该账号只能读写自己的桶，看不到其它桶。

> 密码、策略名、桶名请按环境替换。

## 一次性开通

以下命令均在 MinIO 容器内执行。别名 `myminio` 需已由 `mc alias set` 配好。

### 1. 进入容器

在部署了 MinIO 的服务器上：

```bash
docker exec -it minio /bin/sh
```

### 2. 创建专属用户

账号 `oss-user`，密码请换成更复杂的值：

```bash
mc admin user add myminio oss-user ServicePass123
```

### 3. 创建存储桶

```bash
mc mb myminio/oss-bucket
```

### 4. 编写并导入读写策略

系统自带的 `readwrite` 可读写**全部桶**，权限过大。应定制只允许操作 `oss-bucket` 的策略。

在容器内生成策略文件：

```bash
cat <<EOF > oss-policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "arn:aws:s3:::oss-bucket",
                "arn:aws:s3:::oss-bucket/*"
            ]
        }
    ]
}
EOF
```

`s3:*` 表示对该桶拥有创建、删除、列出对象等全部操作权，但**仅限** `oss-bucket`。

导入策略，系统内命名为 `oss-policy`：

```bash
mc admin policy create myminio oss-policy oss-policy.json
```

### 5. 将策略绑定给用户

```bash
mc admin policy attach myminio oss-policy --user=oss-user
```

---

## 服务配置

开通后把下列凭证填到调用方自己的配置里（oss 的 etcd `{KEY}/service`，或业务服务自定的 oss 配置），再 `oss.New(cfg)`。`oss.buckets` 须标明每个桶的 `mode`（`public` / `private`），这是 URL 策略，不会从 MinIO 自动探测。


| 项          | 值                          | 说明                            |
| ---------- | -------------------------- | ----------------------------- |
| Endpoint   | `http://<MinIO服务器IP>:9200` | 代码连 **9200 数据口**，不要连 9201 控制台 |
| Access Key | `oss-user`         |                               |
| Secret Key | `ServicePass123`           | 请换成实际密码                       |
| Bucket     | `oss-bucket`       |                               |


该账号只能在本桶内上传、下载、删除，无法看到或操作其它桶。