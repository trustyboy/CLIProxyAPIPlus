#!/usr/bin/env bash
#
# build.sh - 自动构建脚本
# 检查git更新 -> 停止服务 -> 拉取代码 -> 构建Web前端 -> 编译 -> 启动服务
# 使用 -f 参数强制构建，跳过更新检查

set -euo pipefail  # 严格模式：出错即止

# ------------------- 配置 -------------------
SERVICE_NAME="cli-proxy-api:cli-proxy-api_00"
OUTPUT_NAME="cli-proxy-api"
OUTPUT_DIR="."

# ------------------- 函数 -------------------

# 检查git是否有新提交
check_git_updates() {
    echo "[INFO] 检查git更新..."
    git fetch origin

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

# 拉取代码 - 强制以远程为准，但保留未跟踪文件
pull_code() {
    echo "[INFO] 拉取最新代码..."

    # 先检查是否有新的提交
    LOCAL_COMMIT=$(git rev-parse HEAD)
    REMOTE_COMMIT=$(git rev-parse @{u})

    if [[ "${LOCAL_COMMIT}" == "${REMOTE_COMMIT}" ]]; then
        echo "[INFO] 本地已是最新，无需拉取代码"
        return
    fi

    echo "[INFO] 检测到远程有新的提交，强制以远程代码为准，放弃本地修改和提交，但保留未跟踪文件"

    # 获取远程最新状态
    git fetch origin    

    # 强制重置到当前分支的远程状态（保留未跟踪文件）
    git reset --hard "@{u}"
    echo "[INFO] 已强制重置到远程分支状态，未跟踪文件已保留"
	
	# 先同步子模块到当前提交版本，避免 reset 时出现目录非空警告
    echo "[INFO] 同步子模块..."
    git submodule update --init --recursive --force
    echo "[INFO] 子模块已更新"
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

# 部署可执行文件（停止服务后调用）
deploy_binary() {
    if [[ "${UPDATE_MODE}" != true ]]; then
        return  # 非更新模式不需要部署
    fi

    echo "[INFO] 部署可执行文件..."

    # 检查可执行文件是否存在
    if [[ ! -f "${OUTPUT_DIR}/${OUTPUT_NAME}" ]]; then
        echo "[ERROR] 可执行文件不存在: ${OUTPUT_DIR}/${OUTPUT_NAME}"
        exit 1
    fi

    # 备份旧文件
    if [[ -f "${OUTPUT_DIR}/${OUTPUT_NAME}" ]]; then
        BACKUP_NAME="${OUTPUT_DIR}/${OUTPUT_NAME}.backup.$(date +%Y%m%d%H%M%S)"
        cp "${OUTPUT_DIR}/${OUTPUT_NAME}" "${BACKUP_NAME}"
        echo "[INFO] 已备份旧文件到: ${BACKUP_NAME}"
    fi

    # 新文件已经在目标位置，无需移动
    echo "[INFO] 可执行文件已就绪: ${OUTPUT_DIR}/${OUTPUT_NAME}"
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

        # 2. 拉取代码
        pull_code
    else
        echo "[INFO] 构建模式，跳过更新检查"
    fi

    # 3. 构建Web前端
    build_web

    # 4. 编译
    build_binary

    # 5. 构建成功，检查是否需要更新服务
    if [[ "${UPDATE_MODE}" == true ]]; then
        # 5.1 停止旧服务（确保构建成功后再停服务）
        stop_service

        # 5.2 部署新可执行文件（停止服务后覆盖，安全）
        deploy_binary

        # 5.3 启动新服务
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