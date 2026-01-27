#!/usr/bin/env bash
#
# build.sh - 自动构建脚本
# 检查git更新 -> 停止服务 -> 拉取代码 -> 编译 -> 启动服务
# 使用 -f 参数强制构建，跳过更新检查

set -euo pipefail  # 严格模式：出错即止

# ------------------- 配置 -------------------
PROXY_CHAINS_CMD="proxychains"  # proxychains命令，可根据需要改为 proxychains4
SERVICE_NAME="cli-proxy-api:cli-proxy-api_00"
OUTPUT_NAME="cli-proxy-api"
OUTPUT_DIR="."

# ------------------- 函数 -------------------

# 检查git是否有新提交
check_git_updates() {
    echo "[INFO] 检查git更新..."
    ${PROXY_CHAINS_CMD} git fetch origin

    LOCAL_COMMIT=$(${PROXY_CHAINS_CMD} git rev-parse HEAD)
    REMOTE_COMMIT=$(${PROXY_CHAINS_CMD} git rev-parse @{u})

    if [[ "${LOCAL_COMMIT}" == "${REMOTE_COMMIT}" ]]; then
        echo "[INFO] 没有新的提交，无需构建"
        echo "[INFO] 本地提交: ${LOCAL_COMMIT}"
        exit 0
    else
        echo "[INFO] 检测到新的提交"
        echo "[INFO] 本地提交: ${LOCAL_COMMIT}"
        echo "[INFO] 远程提交: ${REMOTE_COMMIT}"
        echo "[INFO] 提交差异:"
        ${PROXY_CHAINS_CMD} git log --oneline HEAD..@{u}
    fi
}

# 停止服务
stop_service() {
    echo "[INFO] 停止服务: ${SERVICE_NAME}"
    if command -v supervisorctl &> /dev/null; then
        supervisorctl stop "${SERVICE_NAME}"
        echo "[INFO] 服务已停止"
    else
        echo "[WARN] supervisorctl 未找到，跳过停止服务步骤"
    fi
}

# 拉取代码
pull_code() {
    echo "[INFO] 拉取最新代码..."
    ${PROXY_CHAINS_CMD} git pull origin HEAD
    echo "[INFO] 代码已更新"
}

# 编译
build_binary() {
    echo "[INFO] 开始编译..."

    # ------------------- 读取 Git 信息 -------------------
    VERSION="$(${PROXY_CHAINS_CMD} git describe --tags --always --dirty)"
    COMMIT="$(${PROXY_CHAINS_CMD} git rev-parse --short HEAD)"
    BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    # ------------------- LDFLAGS（注入版本信息） -------------------
    LDFLAGS="-s -w \
      -X main.Version=${VERSION} \
      -X main.Commit=${COMMIT} \
      -X main.BuildDate=${BUILD_DATE}"

    # ------------------- 编译 -------------------
    echo "[INFO] 编译目标平台: linux/amd64"
    echo "[INFO] 版本信息: Version=${VERSION} Commit=${COMMIT} BuildDate=${BUILD_DATE}"
    echo "[INFO] 输出文件: ${OUTPUT_DIR}/${OUTPUT_NAME}"

    # 编译Linux版本
    env GOOS=linux GOARCH=amd64 \
        go build -ldflags="${LDFLAGS}" -trimpath -o "${OUTPUT_DIR}/${OUTPUT_NAME}" ./cmd/server

    echo "[INFO] 编译完成 🎉"
    echo "可执行文件已生成于: ${OUTPUT_DIR}/${OUTPUT_NAME}"
}

# 启动服务
start_service() {
    echo "[INFO] 启动服务: ${SERVICE_NAME}"
    if command -v supervisorctl &> /dev/null; then
        supervisorctl start "${SERVICE_NAME}"
        echo "[INFO] 服务已启动"
        echo "[INFO] 查看服务状态: supervisorctl status ${SERVICE_NAME}"
        echo "[INFO] 查看服务日志: supervisorctl tail -f ${SERVICE_NAME}"
    else
        echo "[WARN] supervisorctl 未找到，跳过启动服务步骤"
    fi
}

# ------------------- 主流程 -------------------
main() {
    # 检查命令行参数
    FORCE_BUILD=false
    if [[ "${1:-}" == "-f" ]]; then
        FORCE_BUILD=true
        echo "[INFO] 强制构建模式 (-f)，跳过更新检查"
    fi

    echo "========================================"
    echo "  CLIProxyAPI 自动构建脚本"
    echo "========================================"
    echo ""

    # 1. 检查git是否有新提交（除非强制构建）
    if [[ "${FORCE_BUILD}" == false ]]; then
        check_git_updates
    else
        echo "[INFO] 跳过更新检查，直接构建"
    fi

    # 2. 停止服务
    stop_service

    # 3. 拉取代码
    pull_code

    # 4. 编译
    build_binary

    # 5. 启动服务
    start_service

    echo ""
    echo "========================================"
    echo "  构建和部署完成！"
    echo "========================================"
}

# 执行主流程
main