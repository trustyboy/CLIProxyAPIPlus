# CLIProxyAPI Plus 代码结构文档

## 项目概述

CLIProxyAPI Plus 是 CLIProxyAPI 的增强版本，在主项目基础上增加了对第三方提供商的支持。该项目是一个 AI 代理服务器，提供 OpenAI/Gemini/Claude 兼容的 API 接口，使 CLI 模型可以与标准 AI API 工具和库一起使用。

### 项目特点

- **多提供商支持**：Gemini、Claude、Codex、Kiro (AWS CodeWhisperer)、GitHub Copilot、Vertex AI 等
- **嵌入式管理界面**：直接编译到二进制文件中，无需网络下载
- **OAuth Web 认证**：浏览器界面的 OAuth 流程
- **详细日志记录**：请求 ID、提供商、模型、账号信息
- **自动化构建**：支持自动拉取代码、构建前端、编译二进制

## 目录结构

```
CLIProxyAPIPlus/
├── .claude/                        # CLI 配置和规则
├── .github/                        # GitHub Actions 工作流
│   └── workflows/                  # CI/CD 配置
├── .gitmodules                     # Git 子模块配置
├── .goreleaser.yml                 # GoReleaser 构建配置
├── .sisyphus/                      # Sisyphus 项目管理系统
├── assets/                         # 静态资源
├── auths/                          # 认证文件存储目录
├── cmd/                            # 命令行工具
│   └── server/
│       └── main.go                 # 服务器入口点
├── dist/                           # 构建输出目录
├── docs/                           # 文档目录
│   ├── sdk-access.md               # SDK 访问控制文档
│   ├── sdk-advanced.md             # SDK 高级用法文档
│   ├── sdk-usage.md                # SDK 使用文档
│   ├── sdk-watcher.md              # SDK 监控器文档
│   └── code-structure.md           # 本文档
├── examples/                       # 示例代码
│   ├── custom-provider/            # 自定义提供商示例
│   ├── http-request/               # HTTP 请求示例
│   └── translator/                 # 翻译示例
├── internal/                       # 内部包（核心业务逻辑）
│   ├── access/                     # 访问控制和认证
│   ├── api/                        # API 处理器
│   │   ├── handlers/               # API 处理器
│   │   │   ├── management/         # 管理 API 处理器
│   │   │   └── oauth/              # OAuth 处理器
│   │   ├── middleware/             # 中间件
│   │   ├── modules/                # API 模块
│   │   │   └── amp/                # Amp 模块
│   │   └── server.go               # 服务器实现
│   ├── auth/                       # 认证模块
│   │   └── kiro/                   # Kiro 认证
│   ├── browser/                    # 浏览器自动化
│   ├── buildinfo/                  # 构建信息
│   ├── cache/                      # 缓存管理
│   ├── cmd/                        # 命令处理
│   │   ├── anthropic_login.go      # Anthropic 登录
│   │   ├── antigravity_login.go    # Antigravity 登录
│   │   ├── claude_login.go         # Claude 登录
│   │   ├── codex_login.go          # Codex 登录
│   │   ├── github_copilot_login.go # GitHub Copilot 登录
│   │   ├── iflow_login.go          # iFlow 登录
│   │   ├── kimi_login.go           # Kimi 登录
│   │   ├── kiro_login.go           # Kiro 登录
│   │   ├── login.go                # 通用登录
│   │   ├── openai_login.go         # OpenAI 登录
│   │   ├── qwen_login.go           # Qwen 登录
│   │   ├── run.go                  # 服务运行
│   │   └── vertex_import.go        # Vertex 导入
│   ├── config/                     # 配置管理
│   ├── constant/                   # 常量定义
│   ├── interfaces/                 # 接口定义
│   ├── logging/                    # 日志管理
│   ├── managementasset/            # 管理资产（嵌入式 UI）
│   ├── misc/                       # 杂项工具
│   ├── registry/                   # 模型注册表
│   ├── runtime/                    # 运行时管理
│   ├── store/                      # 数据存储
│   │   ├── gitstore.go             # Git 存储
│   │   ├── objectstore.go          # 对象存储
│   │   └── postgresstore.go        # PostgreSQL 存储
│   ├── thinking/                   # 思考/推理模块
│   ├── translator/                 # 翻译模块
│   ├── usage/                      # 使用统计
│   ├── util/                       # 工具函数
│   ├── watcher/                    # 配置监视器
│   └── wsrelay/                    # WebSocket 代理
├── sdk/                            # 软件开发工具包
│   ├── access/                     # 访问控制 SDK
│   ├── api/                        # API SDK
│   │   └── handlers/               # API 处理器 SDK
│   │       ├── claude/             # Claude 处理器
│   │       ├── gemini/             # Gemini 处理器
│   │       └── openai/             # OpenAI 处理器
│   ├── auth/                       # 认证 SDK
│   ├── cliproxy/                   # CLI 代理 SDK
│   ├── config/                     # 配置 SDK
│   ├── logging/                    # 日志 SDK
│   └── translator/                 # 翻译 SDK
├── skills/                         # 技能定义
│   └── instincts/                  # 模式定义
├── static/                         # 静态文件
├── test/                           # 测试文件
├── web/                            # 前端代码（git submodule）
│   ├── dist/                       # 构建输出
│   ├── node_modules/               # npm 依赖
│   ├── src/                        # 源代码
│   ├── package.json                # Node.js 依赖
│   ├── vite.config.ts              # Vite 配置
│   └── index.html                  # 入口 HTML
├── build.sh                        # 自动化构建脚本
├── config.example.yaml             # 配置示例
├── config.yaml                     # 主配置文件
├── docker-compose.yml              # Docker Compose 配置
├── go.mod                          # Go 模块定义
├── go.sum                          # Go 模块校验和
├── README.md                       # 项目说明文档
├── README_CN.md                    # 中文项目说明文档
└── CHANGELOG.md                    # 更新日志
```

## 核心模块详解

### 1. internal/api/ - API 层

**主要文件：**
- `server.go` - HTTP 服务器实现，包含路由设置、中间件配置
- `handlers/` - 各类 API 请求处理器
- `middleware/` - CORS、认证等中间件
- `modules/amp/` - Amp CLI 集成模块

**功能：**
- 提供 OpenAI 兼容的 API 端点 (`/v1/chat/completions`, `/v1/completions`, `/v1/models`)
- 提供 Gemini 兼容的 API 端点 (`/v1beta/models/{model}:generateContent`)
- 提供 Claude 兼容的 API 端点 (`/v1/messages`)
- OAuth 回调处理
- 管理面板 API

### 2. internal/config/ - 配置管理

**主要类型：**
```go
type Config struct {
    Host string              // 服务器主机
    Port int                 // 服务器端口
    TLS TLSConfig            // TLS 配置
    RemoteManagement         // 远程管理配置
    AuthDir string           // 认证目录
    Debug bool               // 调试模式
    GeminiKey []GeminiKey    // Gemini API 密钥
    KiroKey []KiroKey        // Kiro 配置
    CodexKey []CodexKey      // Codex API 密钥
    ClaudeKey []ClaudeKey    // Claude API 密钥
    OpenAICompatibility []   // OpenAI 兼容提供商
    VertexCompatAPIKey []    // Vertex 兼容密钥
    AmpCode AmpCode          // Amp 集成配置
    OAuthModelAlias          // OAuth 模型别名
    OAuthExcludedModels      // OAuth 排除模型
    Payload PayloadConfig    // 载荷配置
    // ... 更多配置项
}
```

### 3. internal/registry/ - 模型注册表

**主要功能：**
- 集中管理所有 AI 服务提供商的模型
- 动态模型注册和注销
- 引用计数跟踪活跃客户端
- 自动隐藏无客户端或配额超限的模型

**核心结构：**
```go
type ModelRegistry struct {
    models map[string]*ModelRegistration
    clientModels map[string][]string
    clientProviders map[string]string
    mutex *sync.RWMutex
}

type ModelInfo struct {
    ID string
    Type string
    DisplayName string
    ContextLength int
    MaxCompletionTokens int
    SupportedGenerationMethods []string
    Thinking *ThinkingSupport
    // ... 更多字段
}
```

### 4. internal/store/ - 数据存储

支持三种存储后端：

#### PostgreSQL 存储 (`postgresstore.go`)
- 使用 pgx/v5 驱动
- 支持自定义 schema
- 配置和认证文件持久化

#### Git 存储 (`gitstore.go`)
- 使用 go-git 操作 Git 仓库
- 支持远程 Git 仓库同步
- 配置版本控制

#### 对象存储 (`objectstore.go`)
- 兼容 S3 API (使用 MinIO SDK)
- 支持 HTTP/HTTPS
- 配置和认证文件云端存储

### 5. internal/cmd/ - 命令处理

**登录命令：**
- `DoLogin` - Google/Gemini 登录
- `DoCodexLogin` - Codex OAuth 登录
- `DoClaudeLogin` - Claude OAuth 登录
- `DoQwenLogin` - Qwen OAuth 登录
- `DoIFlowLogin` - iFlow OAuth 登录
- `DoIFlowCookieAuth` - iFlow Cookie 认证
- `DoKimiLogin` - Kimi OAuth 登录
- `DoKiroLogin` - Kiro Google OAuth 登录
- `DoKiroGoogleLogin` - Kiro Google OAuth 登录
- `DoKiroAWSLogin` - Kiro AWS Builder ID 设备码登录
- `DoKiroAWSAuthCodeLogin` - Kiro AWS Builder ID 授权码登录
- `DoKiroImport` - 从 Kiro IDE 导入 Token
- `DoGitHubCopilotLogin` - GitHub Copilot 设备码登录

### 6. internal/browser/ - 浏览器自动化

使用 `github.com/pkg/browser` 打开浏览器，支持：
- OAuth 流程的浏览器打开
- 隐私模式/普通模式切换

### 7. internal/auth/kiro/ - Kiro 认证

**主要功能：**
- Kiro Token 后台刷新（过期前 10 分钟自动刷新）
- 多账户支持
- Token 文件持久化
- 冷却管理

### 8. internal/watcher/ - 配置监视器

**主要功能：**
- 监控配置文件变化
- 热重载配置变更
- 客户端状态管理
- 事件分发

### 9. internal/logging/ - 日志管理

**主要组件：**
- `global_logger.go` - 全局日志配置
- `gin_logger.go` - Gin 框架日志中间件
- `request_logger.go` - 请求日志记录器
- `log_dir_cleaner.go` - 日志目录清理
- `requestid.go` - 请求 ID 生成

**日志格式：**
```
[时间戳] [级别] | 请求ID | 状态码 | 延迟 | 客户端IP | 方法 | 路径 | provider=提供商 | model=模型 | account=账号
```

### 10. internal/managementasset/ - 管理资产

**功能：**
- 嵌入式管理 UI
- 从 GitHub Releases 自动更新
- 版本一致性保证
- 离线部署支持

### 11. internal/thinking/ - 思考/推理模块

**主要文件：**
- `types.go` - 思考类型定义
- `convert.go` - 思考内容转换
- `validate.go` - 思考配置验证
- `suffix.go` - 思考后缀处理
- `strip.go` - 思考内容剥离
- `errors.go` - 错误处理

### 12. internal/wsrelay/ - WebSocket 代理

**主要功能：**
- WebSocket 连接管理
- 消息转发
- 会话管理
- 心跳机制

## 配置文件详解

### config.yaml 主要配置项

```yaml
# 服务器配置
host: ""              # 绑定主机，空值表示所有接口
port: 8317            # 服务器端口

# TLS 配置
tls:
  enable: false
  cert: ""
  key: ""

# 管理 API 配置
remote-management:
  allow-remote: false           # 是否允许远程管理
  secret-key: ""                # 管理密钥
  disable-control-panel: false  # 是否禁用控制面板
  panel-github-repository: "https://github.com/router-for-me/Cli-Proxy-API-Management-Center"

# 认证配置
auth-dir: "~/.cli-proxy-api"
api-keys:
  - "your-api-key-1"

# 日志配置
debug: false
logging-to-file: false
logs-max-total-size-mb: 0
error-logs-max-files: 10

# 代理配置
proxy-url: ""

# 路由策略
routing:
  strategy: "round-robin"  # round-robin 或 fill-first

# 请求重试配置
request-retry: 3
max-retry-interval: 30

# 配额超限行为
quota-exceeded:
  switch-project: true
  switch-preview-model: true

# 提供商配置
gemini-api-key: []
claude-api-key: []
codex-api-key: []
kiro: []
openai-compatibility: []
vertex-api-key: []
ampcode: {}

# OAuth 模型别名
oauth-model-alias: {}

# OAuth 排除模型
oauth-excluded-models: {}

# 载荷配置
payload:
  default: []
  default-raw: []
  override: []
  override-raw: []
  filter: []
```

## API 端点

### OpenAI 兼容 API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/chat/completions` | POST | 聊天补全 |
| `/v1/completions` | POST | 文本补全 |
| `/v1/models` | GET | 模型列表 |

### Gemini 兼容 API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1beta/models/{model}:generateContent` | POST | 内容生成 |
| `/v1beta/models` | GET | 模型列表 |

### Claude 兼容 API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/messages` | POST | 消息 API |
| `/v1/models` | GET | 模型列表 |

### OAuth 端点

| 端点 | 描述 |
|------|------|
| `/v0/oauth/google` | Google OAuth 回调 |
| `/v0/oauth/claude` | Claude OAuth 回调 |
| `/v0/oauth/codex` | Codex OAuth 回调 |
| `/v0/oauth/qwen` | Qwen OAuth 回调 |
| `/v0/oauth/iflow` | iFlow OAuth 回调 |
| `/v0/oauth/kimi` | Kimi OAuth 回调 |
| `/v0/oauth/kiro` | Kiro OAuth Web 界面 |
| `/v0/oauth/github-copilot` | GitHub Copilot OAuth 回调 |

### 管理 API

| 端点 | 描述 |
|------|------|
| `/v0/management/*` | 管理面板 API |
| `/management.html` | 管理面板 UI |

### WebSocket API

| 端点 | 描述 |
|------|------|
| `/v1/ws` | WebSocket 代理 |

## 依赖管理

### Go 依赖 (go.mod)

**主要依赖：**
- `github.com/gin-gonic/gin` v1.10.1 - Web 框架
- `github.com/sirupsen/logrus` v1.9.3 - 日志库
- `github.com/google/uuid` v1.6.0 - UUID 生成
- `github.com/joho/godotenv` v1.5.1 - 环境变量加载
- `github.com/go-git/go-git/v6` v6.0.0 - Git 操作
- `github.com/minio/minio-go/v7` v7.0.66 - 对象存储
- `github.com/jackc/pgx/v5` v5.7.6 - PostgreSQL 驱动
- `github.com/gorilla/websocket` v1.5.3 - WebSocket 支持
- `github.com/tidwall/gjson` v1.18.0 - JSON 查询
- `github.com/tiktoken-go/tokenizer` v0.7.0 - Token 计数

### 前端依赖 (web/package.json)

- React 18
- TypeScript
- Vite (构建工具)

## 构建脚本 (build.sh)

**功能：**
1. 检查 git 更新
2. 停止服务
3. 拉取代码
4. 构建 Web 前端 (`npm run build`)
5. 复制到 embed 目录
6. 编译 Go 二进制文件
7. 启动服务

**使用：**
```bash
# 正常模式（检查更新）
./build.sh

# 强制构建模式（跳过更新检查）
./build.sh -f
```

**环境变量：**
- `PROXY_CHAINS_CMD` - proxychains 命令
- `SERVICE_NAME` - supervisor 服务名称
- `OUTPUT_NAME` - 输出文件名
- `OUTPUT_DIR` - 输出目录

## 支持的提供商

| 提供商 | 认证方式 | 说明 |
|--------|----------|------|
| Gemini | API Key / OAuth | Google Gemini 模型 |
| Claude | API Key / OAuth | Anthropic Claude 模型 |
| Codex | API Key / OAuth | Microsoft CodeX 模型 |
| Kiro | OAuth / Token 导入 | AWS CodeWhisperer |
| GitHub Copilot | OAuth | GitHub Copilot |
| Vertex AI | API Key / Service Account | Google Cloud Vertex AI |
| Kimi | OAuth | 月之暗面 Kimi |
| Qwen | OAuth | 通义千问 |
| iFlow | OAuth / Cookie | 智谱 AI |
| Antigravity | OAuth | Antigravity |

## Plus 版本增强功能

### 1. 增强日志记录
- 请求 ID、模型名、提供商信息
- 实际使用的通道账号信息
- 结构化日志输出

### 2. 嵌入式管理 UI
- 直接编译到二进制文件
- 无需网络下载
- 离线部署支持

### 3. OAuth Web 认证
- 浏览器界面的 OAuth 流程
- 支持隐私模式
- Kiro 多账户支持

### 4. 后台 Token 刷新
- Kiro Token 自动刷新
- 过期前 10 分钟触发

### 5. 速率限制
- 内置请求速率限制
- 防止 API 滥用

### 6. 使用统计
- 实时监控
- 配额管理

## 技术栈

### 后端
- **语言**：Go 1.24
- **Web 框架**：gin-gonic/gin
- **日志**：logrus
- **配置**：YAML (gopkg.in/yaml.v3)
- **认证**：OAuth 2.0、API 密钥

### 前端
- **框架**：React 18
- **语言**：TypeScript
- **构建工具**：Vite

### 数据库
- **PostgreSQL**：可选，用于持久化配置和认证文件
- **Git**：可选，用于配置版本控制
- **对象存储**：可选，兼容 S3 API

## 部署方式

### Docker 部署
```bash
docker compose up -d
```

### 源码部署
```bash
./build.sh
./cli-proxy-api
```

### Supervisor 部署
```bash
./build.sh  # 自动处理 supervisor 服务
```

## 环境变量

| 变量名 | 描述 |
|--------|------|
| `PGSTORE_DSN` | PostgreSQL 连接字符串 |
| `PGSTORE_SCHEMA` | PostgreSQL schema |
| `PGSTORE_LOCAL_PATH` | PostgreSQL 本地路径 |
| `GITSTORE_GIT_URL` | Git 仓库 URL |
| `GITSTORE_GIT_USERNAME` | Git 用户名 |
| `GITSTORE_GIT_TOKEN` | Git 访问令牌 |
| `GITSTORE_LOCAL_PATH` | Git 本地路径 |
| `OBJECTSTORE_ENDPOINT` | 对象存储端点 |
| `OBJECTSTORE_ACCESS_KEY` | 对象存储访问密钥 |
| `OBJECTSTORE_SECRET_KEY` | 对象存储密钥 |
| `OBJECTSTORE_BUCKET` | 对象存储桶名 |
| `OBJECTSTORE_LOCAL_PATH` | 对象存储本地路径 |
| `DEPLOY` | 部署模式 (cloud) |

## 版本信息

```bash
./cli-proxy-api -version
```

输出格式：
```
CLIProxyAPI Version: {version}, Commit: {commit}, BuiltAt: {date}
```