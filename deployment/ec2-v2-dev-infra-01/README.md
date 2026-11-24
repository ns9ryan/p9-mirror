
# ec2-v2-dev-infra-01

基础设施服务器配置文件管理。

## 服务器临时目录

```text
/tmp/p9-deployment/ec2-v2-dev-infra-01/
```

## 配置文件映射

| 本地文件 | 服务器实际路径 |
| --- | --- |
| `opt/crm-data/docker-compose.yml` | `/opt/crm-data/docker-compose.yml` |
| `data/doris-fe/conf/fe.conf` | `/data/doris-fe/conf/fe.conf` |
| `data/doris-be/conf/be.conf` | `/data/doris-be/conf/be.conf` |

## 从服务器下载配置文件

先通过 JumpServer SFTP 在服务器 `/tmp/p9-deployment/ec2-v2-dev-infra-01/` 下创建对应目录，然后 SSH 执行：

```bash
BASE=/tmp/p9-deployment/ec2-v2-dev-infra-01

cp -a /opt/crm-data/docker-compose.yml "$BASE/opt/crm-data/"
cp -a /data/doris-fe/conf/fe.conf "$BASE/data/doris-fe/conf/"
cp -a /data/doris-be/conf/be.conf "$BASE/data/doris-be/conf/"
```

复制完成后，通过 JumpServer SFTP 下载到本地 `deployment/ec2-v2-dev-infra-01/` 对应目录。

## 将本地配置文件上传到服务器

先通过 JumpServer SFTP 将本地修改后的文件上传到服务器临时目录，保持原有目录结构。

然后 SSH 执行以下命令。

### 1. 备份服务器原配置

```bash
TS=$(date +%Y%m%d-%H%M%S)

cp -a /opt/crm-data/docker-compose.yml "/var/backups/docker-compose.yml.$TS"
cp -a /data/doris-fe/conf/fe.conf "/var/backups/fe.conf.$TS"
cp -a /data/doris-be/conf/be.conf "/var/backups/be.conf.$TS"
```

### 2. 替换服务器配置

```bash
BASE=/tmp/p9-deployment/ec2-v2-dev-infra-01

cp -a "$BASE/opt/crm-data/docker-compose.yml" /opt/crm-data/
cp -a "$BASE/data/doris-fe/conf/fe.conf" /data/doris-fe/conf/
cp -a "$BASE/data/doris-be/conf/be.conf" /data/doris-be/conf/
```

### 3. 检查文件是否一致

```bash
BASE=/tmp/p9-deployment/ec2-v2-dev-infra-01

cmp -s "$BASE/opt/crm-data/docker-compose.yml" /opt/crm-data/docker-compose.yml && echo "[OK] docker-compose.yml" || echo "[FAILED] docker-compose.yml"

cmp -s "$BASE/data/doris-fe/conf/fe.conf" /data/doris-fe/conf/fe.conf && echo "[OK] fe.conf" || echo "[FAILED] fe.conf"

cmp -s "$BASE/data/doris-be/conf/be.conf" /data/doris-be/conf/be.conf && echo "[OK] be.conf" || echo "[FAILED] be.conf"
```

## 注意事项

- 只同步实际需要修改的文件，不必每次全部同步。
- 替换配置前必须备份原文件。
- 配置替换后检查文件内容及权限。
- 配置文件替换不会自动重启服务，需要根据实际变更决定是否重启。
- 不要将密码、密钥等敏感信息提交到 Git。
