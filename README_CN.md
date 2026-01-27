# CLIProxyAPI Plus

[English](README.md) | 中文

这是 [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 的 Plus 版本，在原有基础上增加了第三方供应商的支持。

所有的第三方供应商支持都由第三方社区维护者提供，CLIProxyAPI 不提供技术支持。如需取得支持，请与对应的社区维护者联系。

该 Plus 版本的主线功能与主线功能强制同步。

## 与主线版本版本差异

- 新增 GitHub Copilot 支持（OAuth 登录），由[em4go](https://github.com/em4go/CLIProxyAPI/tree/feature/github-copilot-auth)提供
- 新增 Kiro (AWS CodeWhisperer) 支持 (OAuth 登录), 由[fuko2935](https://github.com/fuko2935/CLIProxyAPI/tree/feature/kiro-integration)、[Ravens2121](https://github.com/Ravens2121/CLIProxyAPIPlus/)提供

## 新增功能 (Plus 增强版)

### 1. 日志增强

**详细的请求日志记录**
- 在日志中显示请求 ID、模型名称、提供商信息
- 记录实际使用的渠道商账号信息
- 支持结构化日志输出，便于日志分析

**日志格式示例**
```
[2025-01-28 04:00:00] [info ] | a1b2c3d4 | 200 |       23.559s | 192.168.1.100 | POST | /v1/chat/completions | provider=gemini | model=gemini-pro | account=oauth:user@example.com
```

**关键改进**
- 修复多个渠道有相同模型时的渠道显示问题
- 当无法获取模型时，不显示提供商和模型信息
- 从 gin.Context 读取实际使用的提供商，而非从模型名推断

### 2. 自动构建脚本

**build.sh 功能**
- 检查 git 更新，无新提交时自动退出
- 自动停止/启动 supervisor 服务
- 支持 proxychains 代理拉取代码
- 注入版本信息到二进制文件
- 支持 `-f` 参数强制构建

**使用方法**
```bash
# 正常模式（检查更新）
./build.sh

# 强制构建模式（跳过更新检查）
./build.sh -f
```

### 3. 其他增强功能

- **OAuth Web 认证**: 基于浏览器的 Kiro OAuth 登录，提供美观的 Web UI
- **请求限流器**: 内置请求限流，防止 API 滥用
- **后台令牌刷新**: 过期前 10 分钟自动刷新令牌
- **监控指标**: 请求指标收集，用于监控和调试
- **设备指纹**: 设备指纹生成，增强安全性
- **冷却管理**: 智能冷却机制，应对 API 速率限制
- **用量检查器**: 实时用量监控和配额管理
- **模型转换器**: 跨供应商的统一模型名称转换
- **UTF-8 流处理**: 改进的流式响应处理

## Kiro 认证

### 网页端 OAuth 登录

访问 Kiro OAuth 网页认证界面：

```
http://your-server:8080/v0/oauth/kiro
```

提供基于浏览器的 Kiro (AWS CodeWhisperer) OAuth 认证流程，支持：
- AWS Builder ID 登录
- AWS Identity Center (IDC) 登录
- 从 Kiro IDE 导入令牌

## 快速开始

### 使用 Docker 部署

```bash
# 创建部署目录
mkdir -p ~/cli-proxy && cd ~/cli-proxy

# 创建 docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  cli-proxy-api:
    image: 17600006524/cli-proxy-api-plus:latest
    container_name: cli-proxy-api-plus
    ports:
      - "8317:8317"
    volumes:
      - ./config.yaml:/CLIProxyAPI/config.yaml
      - ./auths:/root/.cli-proxy-api
      - ./logs:/CLIProxyAPI/logs
    restart: unless-stopped
EOF

# 下载示例配置
curl -o config.yaml https://raw.githubusercontent.com/linlang781/CLIProxyAPIPlus/main/config.example.yaml

# 启动服务
docker compose up -d
```

### 使用源码部署

```bash
# 克隆仓库
git clone https://github.com/trustyboy/CLIProxyAPIPlus.git
cd CLIProxyAPIPlus

# 切换到 gf 分支（增强功能分支）
git checkout gf

# 使用自动构建脚本
./build.sh

# 或手动编译
go build -o cli-proxy-api ./cmd/server

# 配置 config.yaml
cp config.example.yaml config.yaml
vim config.yaml

# 启动服务
./cli-proxy-api
```

### 配置说明

启动前请编辑 `config.yaml`：

```yaml
# 基本配置示例
server:
  port: 8317

# 在此添加你的供应商配置
```

### 更新到最新版本

```bash
cd ~/cli-proxy
docker compose pull && docker compose up -d
```

## API 接口

### OpenAI 兼容接口

- `POST /v1/chat/completions` - 聊天补全
- `POST /v1/completions` - 文本补全
- `GET /v1/models` - 模型列表

### Gemini 兼容接口

- `POST /v1beta/models/{model}:generateContent` - 内容生成
- `GET /v1beta/models` - 模型列表

### Claude 兼容接口

- `POST /v1/messages` - 消息接口
- `GET /v1/models` - 模型列表

## 日志说明

### 日志级别

- `INFO` - 正常请求
- `WARN` - 4xx 错误
- `ERROR` - 5xx 错误

### 日志字段

- `request_id` - 请求 ID（AI API 请求）
- `status` - HTTP 状态码
- `latency` - 请求耗时
- `client_ip` - 客户端 IP
- `method` - HTTP 方法
- `path` - 请求路径
- `provider` - 提供商（有模型时显示）
- `model` - 模型名称（有模型时显示）
- `account` - 渠道商账号（有账号信息时显示）
- `error` - 错误信息（有错误时显示）

## 版本信息

查看版本：
```bash
./cli-proxy-api -version
```

## 贡献

该项目仅接受第三方供应商支持的 Pull Request。任何非第三方供应商支持的 Pull Request 都将被拒绝。

如果需要提交任何非第三方供应商支持的 Pull Request，请提交到主线版本。

## 许可证

此项目根据 MIT 许可证授权 - 有关详细信息，请参阅 [LICENSE](LICENSE) 文件。