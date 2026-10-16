# oss

对象存储驱动（S3 / MinIO）。只做 `Put` / `Exists` / `URL` / `PresignPut`，不读 etcd、不写档案。

HTTP 上传直接调本包；`oss_file` 入库、服务商配置在 `rpc`。

扩三方：实现 `Store`（浏览器直传再实现 `PresignStore`），在 `NewStore` 增加 `driver` 分支。MinIO 建用户/桶见 [oss命令.md](oss命令.md)。
