#!/usr/bin/env bash
set -Eeuo pipefail

# ec2-v2-dev-infra-01：将服务器配置导出到当前项目目录。
readonly SERVER_NAME="ec2-v2-dev-infra-01"
readonly STAGING_ROOT="/tmp/p9-deployment"
readonly EXPECTED_PROJECT_DIR="${STAGING_ROOT}/${SERVER_NAME}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd -P)"
TEMP_FILE=""

# 需要同步的服务器配置文件。
FILES=(
    "opt/crm-data/docker-compose.yml"
    "data/doris-fe/conf/fe.conf"
    "data/doris-be/conf/be.conf"
)

# 输出错误并退出。
fail() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

# 清理未完成的临时文件。
cleanup() {
    if [[ -n "${TEMP_FILE}" && -e "${TEMP_FILE}" ]]; then
        rm -f -- "${TEMP_FILE}"
    fi
}
trap cleanup EXIT

# ---- 基础检查 ----

[[ "${EUID}" -eq 0 ]] || fail "请使用 root 用户执行。"

[[ "${PROJECT_DIR}" == "${EXPECTED_PROJECT_DIR}" ]] || \
    fail "项目目录不正确，当前目录：${PROJECT_DIR}"

# 先检查全部源文件，避免同步到一半才发现文件缺失。
for relative_path in "${FILES[@]}"; do
    source_path="/${relative_path}"

    [[ -f "${source_path}" && ! -L "${source_path}" ]] || \
        fail "源文件不存在或不是普通文件：${source_path}"
done

# ---- 确认同步 ----

printf '即将同步以下服务器配置文件：\n'

for relative_path in "${FILES[@]}"; do
    printf '  %s\n' "/${relative_path}"
done

printf '\n以下项目中的文件将被覆盖：\n'

for relative_path in "${FILES[@]}"; do
    printf '  %s\n' "${PROJECT_DIR}/${relative_path}"
done

printf '\n'
read -r -p "确认从服务器拉取配置？[y/N] " answer

[[ "${answer}" =~ ^[Yy]$ ]] || {
    printf '已取消同步。\n'
    exit 0
}

# ---- 执行同步 ----

for relative_path in "${FILES[@]}"; do
    source_path="/${relative_path}"
    destination_path="${PROJECT_DIR}/${relative_path}"
    destination_dir="$(dirname -- "${destination_path}")"

    # 创建目标目录。
    install -d -m 0700 -- "${destination_dir}"

    # 先复制到临时文件，再原子替换目标文件。
    TEMP_FILE="$(mktemp "${destination_path}.tmp.XXXXXX")"

    cp --preserve=mode,ownership,timestamps \
        -- "${source_path}" "${TEMP_FILE}"

    mv -f -- "${TEMP_FILE}" "${destination_path}"
    TEMP_FILE=""

    printf '已同步：%s\n' "${source_path}"
done

printf '\n配置拉取完成。\n'
printf '项目目录：%s\n' "${PROJECT_DIR}"