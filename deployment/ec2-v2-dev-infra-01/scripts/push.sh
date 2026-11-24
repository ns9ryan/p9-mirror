#!/usr/bin/env bash
set -Eeuo pipefail

# ec2-v2-dev-infra-01：备份并更新服务器配置文件。
readonly SERVER_NAME="ec2-v2-dev-infra-01"
readonly STAGING_ROOT="/tmp/p9-deployment"
readonly EXPECTED_PROJECT_DIR="${STAGING_ROOT}/${SERVER_NAME}"
readonly BACKUP_BASE="/var/backups/p9-deployment/${SERVER_NAME}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd -P)"

TEMP_FILE=""
BACKUP_DIR=""
PUSH_STARTED=0
PUSH_COMPLETED=0

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

# 恢复服务器原有配置。
restore_backups() {
    local failed=0

    for relative_path in "${FILES[@]}"; do
        local backup_path="${BACKUP_DIR}/${relative_path}"
        local target_path="/${relative_path}"
        local target_dir
        local target_name
        local restore_file

        target_dir="$(dirname -- "${target_path}")"
        target_name="$(basename -- "${target_path}")"

        if [[ ! -f "${backup_path}" ]]; then
            printf 'ERROR: 备份文件不存在：%s\n' "${backup_path}" >&2
            failed=1
            continue
        fi

        restore_file="$(mktemp "${target_dir}/.${target_name}.rollback.XXXXXX")" || {
            printf 'ERROR: 无法创建恢复临时文件：%s\n' "${target_path}" >&2
            failed=1
            continue
        }

        if ! cp -a -- "${backup_path}" "${restore_file}"; then
            rm -f -- "${restore_file}"
            printf 'ERROR: 恢复备份失败：%s\n' "${target_path}" >&2
            failed=1
            continue
        fi

        if ! mv -f -- "${restore_file}" "${target_path}"; then
            rm -f -- "${restore_file}"
            printf 'ERROR: 替换恢复文件失败：%s\n' "${target_path}" >&2
            failed=1
            continue
        fi

        printf '已恢复：%s\n' "${target_path}"
    done

    return "${failed}"
}

# 清理临时文件，异常退出时恢复服务器配置。
cleanup() {
    local exit_code=$?

    trap - EXIT

    if [[ -n "${TEMP_FILE}" && -e "${TEMP_FILE}" ]]; then
        rm -f -- "${TEMP_FILE}" || true
    fi

    if [[ "${PUSH_STARTED}" -eq 1 && "${PUSH_COMPLETED}" -eq 0 ]]; then
        printf '\n检测到同步异常，正在恢复服务器原有配置...\n'

        if restore_backups; then
            printf '原有配置已恢复。\n'
        else
            printf 'ERROR: 自动恢复未能全部成功，请使用备份目录手动恢复：%s\n' \
                "${BACKUP_DIR}" >&2
        fi
    fi

    exit "${exit_code}"
}
trap cleanup EXIT

# ---- 基础检查 ----

[[ "${EUID}" -eq 0 ]] || fail "请使用 root 用户执行。"

[[ "${PROJECT_DIR}" == "${EXPECTED_PROJECT_DIR}" ]] || \
    fail "项目目录不正确，当前目录：${PROJECT_DIR}"

# 检查本地文件和服务器目标文件是否全部存在。
for relative_path in "${FILES[@]}"; do
    source_path="${PROJECT_DIR}/${relative_path}"
    target_path="/${relative_path}"

    [[ -f "${source_path}" && ! -L "${source_path}" ]] || \
        fail "项目配置文件不存在或不是普通文件：${source_path}"

    [[ -f "${target_path}" && ! -L "${target_path}" ]] || \
        fail "服务器目标文件不存在或不是普通文件：${target_path}"
done

# ---- 确认同步 ----

printf '即将更新以下服务器配置文件：\n'

for relative_path in "${FILES[@]}"; do
    printf '  %s\n' "/${relative_path}"
done

printf '\n备份目录：%s\n' "${BACKUP_BASE}"

printf '\n请输入服务器名称确认本次操作：\n'
read -r -p "> " answer

[[ "${answer}" == "${SERVER_NAME}" ]] || {
    printf '已取消同步。\n'
    exit 0
}

# ---- 备份原有配置 ----

install -d -m 0700 -- "${BACKUP_BASE}"

BACKUP_DIR="$(mktemp -d \
    "${BACKUP_BASE}/$(date -u +%Y%m%dT%H%M%SZ)-XXXXXX")"

printf '\n开始备份服务器原有配置...\n'

# 必须先备份全部文件，再开始替换。
for relative_path in "${FILES[@]}"; do
    target_path="/${relative_path}"
    backup_path="${BACKUP_DIR}/${relative_path}"

    install -d -m 0700 -- "$(dirname -- "${backup_path}")"

    cp -a -- "${target_path}" "${backup_path}"

    printf '已备份：%s\n' "${target_path}"
done

# ---- 更新服务器配置 ----

PUSH_STARTED=1

printf '\n开始更新服务器配置...\n'

for relative_path in "${FILES[@]}"; do
    source_path="${PROJECT_DIR}/${relative_path}"
    target_path="/${relative_path}"

    target_dir="$(dirname -- "${target_path}")"
    target_name="$(basename -- "${target_path}")"

    # 保留服务器原文件的用户、用户组和权限。
    read -r owner group mode < <(
        stat -c '%u %g %a' -- "${target_path}"
    )

    # 在目标目录创建临时文件，确保后续替换在同一文件系统内。
    TEMP_FILE="$(mktemp \
        "${target_dir}/.${target_name}.p9-deployment.XXXXXX")"

    install \
        -o "${owner}" \
        -g "${group}" \
        -m "${mode}" \
        -- "${source_path}" "${TEMP_FILE}"

    # 原子替换目标文件。
    mv -f -- "${TEMP_FILE}" "${target_path}"
    TEMP_FILE=""

    # 检查服务器文件是否与上传文件一致。
    cmp -s -- "${source_path}" "${target_path}" || \
        fail "文件校验失败：${target_path}"

    printf '已更新：%s\n' "${target_path}"
done

PUSH_COMPLETED=1

printf '\n配置同步完成。\n'
printf '备份目录：%s\n' "${BACKUP_DIR}"