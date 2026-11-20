#!/bin/sh
set -eu

# 业务 Handler、Logic 和路由沿用手写实现，只更新 API 描述生成的类型。
repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"
generated_dir=$(mktemp -d "$repo_dir/api/.goctl-gen-XXXXXX")
case "$generated_dir" in
    "$repo_dir"/api/.goctl-gen-*) ;;
    *) exit 1 ;;
esac
trap 'rm -rf -- "$generated_dir"' EXIT

goctl api go --api ./api/desc/all.api --dir "$generated_dir" --style=go_zero
cp "$generated_dir/internal/types/types.go" api/internal/types/types.go
