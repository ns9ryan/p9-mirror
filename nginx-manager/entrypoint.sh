#!/bin/sh
set -e

echo "[ENTRYPOINT] Starting nginx-manager..."

# 执行 Go 应用（同时启动 nginx 和 reload API）
exec /usr/local/bin/reload-api
