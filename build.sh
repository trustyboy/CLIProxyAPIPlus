#!/usr/bin/env bash
#
# build.sh - 自动构建脚本
# 检查git更新 -> 停止服务 -> 拉取代码 -> 构建Web前端 -> 编译 -> 启动服务
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
    # 只有 fetch 需要网络请求，使用 proxychains
    if command -v ${PROXY_CHAINS_CMD} &> /dev/null; then
        ${PROXY_CHAINS_CMD} git fetch origin
    else
        echo "[WARN] ${PROXY_CHAINS_CMD} 未找到，直接使用git命令"
        git fetch origin
    fi

    # 本地git命令不需要proxychains
    LOCAL_COMMIT=$(git rev-parse HEAD)
    REMOTE_COMMIT=$(git rev-parse @{u})

    if [[ "${LOCAL_COMMIT}" == "${REMOTE_COMMIT}" ]]; then
        echo "[INFO] 没有新的提交，无需构建"
        echo "[INFO] 本地提交: ${LOCAL_COMMIT}"
        exit 0
    else
        echo "[INFO] 检测到新的提交"
        echo "[INFO] 本地提交: ${LOCAL_COMMIT}"
        echo "[INFO] 远程提交: ${REMOTE_COMMIT}"
        echo "[INFO] 提交差异:"
        git --no-pager log --oneline HEAD..@{u}
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
    # pull 需要网络请求，使用 proxychains
    if command -v ${PROXY_CHAINS_CMD} &> /dev/null; then
        ${PROXY_CHAINS_CMD} git pull origin HEAD
    else
        echo "[WARN] ${PROXY_CHAINS_CMD} 未找到，直接使用git命令"
        git pull origin HEAD
    fi
    echo "[INFO] 代码已更新"
}

# 构建Web前端
build_web() {
    echo "[INFO] 开始构建Web前端..."

    # 检查 web 目录是否存在
    if [[ ! -d "web" ]]; then
        echo "[ERROR] web 目录不存在"
        exit 1
    fi

    # 检查 node_modules 是否存在
    if [[ ! -d "web/node_modules" ]]; then
        echo "[INFO] 首次构建，安装依赖..."
        cd web
        npm install
        cd ..
    fi

    # 构建 Web 前端
    cd web
    npm run build
    cd ..

    # 检查构建结果
    if [[ ! -f "web/dist/index.html" ]]; then
        echo "[ERROR] Web 构建失败：web/dist/index.html 不存在"
        exit 1
    fi

    # 复制到嵌入目录
    echo "[INFO] 复制 Web 构建结果到嵌入目录..."
    mkdir -p internal/managementasset/embedded
    cp web/dist/index.html internal/managementasset/embedded/management.html

    # 验证嵌入文件
    if [[ ! -f "internal/managementasset/embedded/management.html" ]]; then
        echo "[ERROR] 嵌入文件复制失败"
        exit 1
    fi

    EMBED_SIZE=$(du -h internal/managementasset/embedded/management.html | cut -f1)
    echo "[INFO] Web 构建完成 🎉"
    echo "[INFO] 嵌入文件大小: ${EMBED_SIZE}"
}

# 编译
build_binary() {
    echo "[INFO] 开始编译..."

    # ------------------- 读取 Git 信息 -------------------
    # 本地git命令不需要proxychains
    VERSION="$(git describe --tags --always --dirty)"
    COMMIT="$(git rev-parse --short HEAD)"
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
    UPDATE_MODE=false
    FORCE_BUILD=false

    # 检查是否包含 update 参数
    for arg in "$@"; do
        if [[ "$arg" == "update" ]]; then
            UPDATE_MODE=true
        elif [[ "$arg" == "-f" ]]; then
            FORCE_BUILD=true
        fi
    done

    if [[ "${UPDATE_MODE}" == true ]]; then
        if [[ "${FORCE_BUILD}" == true ]]; then
            echo "[INFO] 更新模式 + 强制构建 (-f)，跳过更新检查"
        else
            echo "[INFO] 更新模式，执行完整更新流程"
        fi
    else
        echo "[INFO] 构建模式，仅执行构建操作"
    fi

    echo "========================================"
    echo "  CLIProxyAPI 自动构建脚本"
    echo "========================================"
    echo ""

    # 只在更新模式下执行以下步骤
    if [[ "${UPDATE_MODE}" == true ]]; then
        # 1. 检查git是否有新提交（除非强制构建）
        if [[ "${FORCE_BUILD}" == true ]]; then
            echo "[INFO] 强制构建模式，跳过更新检查"
        else
            check_git_updates
        fi

        # 2. 停止服务
        stop_service

        # 3. 拉取代码
        pull_code
    else
        echo "[INFO] 构建模式，跳过更新检查"
    fi

    # 4. 构建Web前端
    build_web

    # 5. 编译
    build_binary

    # 只在更新模式下启动服务
    if [[ "${UPDATE_MODE}" == true ]]; then
        # 6. 启动服务
        start_service
    fi

    echo ""
    echo "========================================"
    echo "  构建和部署完成！"
    echo "========================================"
    echo "[INFO] 二进制文件: ${OUTPUT_DIR}/${OUTPUT_NAME}"
    echo "[INFO] 嵌入Web: internal/managementasset/embedded/management.html"
}

# 执行主流程
main "$@"