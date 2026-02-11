# 模型路由策略分析

## 概述

CLIProxyAPI Plus 实现了一个复杂的多层路由策略，用于将客户端请求路由到正确的 AI 模型和提供商。路由系统支持多种认证方式、负载均衡策略和模型映射功能。

## 路由架构层次

```
┌─────────────────────────────────────────────────────────────────┐
│                      客户端请求                                    │
│              (OpenAI / Claude / Gemini 格式)                      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    1. 请求解析层                                   │
│  - 解析模型名称 (包括 thinking 后缀)                               │
│  - 识别客户端格式 (openai/claude/gemini)                           │
│  - 提取请求元数据 (request-id, idempotency-key)                   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    2. 提供商识别层                                   │
│  - 通过 ModelRegistry 查询可用提供商                               │
│  - 解析模型别名 (oauth-model-alias, models[].alias)               │
│  - 处理 Amp 模型映射                                              │
│  - 处理 "auto" 模型自动解析                                        │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    3. 访问控制层                                    │
│  - AccessManager 验证 API 密钥                                    │
│  - 支持多种认证提供者                                             │
│  - 返回认证结果 (Provider, Principal, Metadata)                  │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    4. 认证管理层                                    │
│  - AuthManager 管理认证客户端                                     │
│  - 选择匹配的认证客户端 (OAuth/API Key)                            │
│  - 执行请求并处理响应                                             │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    5. 负载均衡层                                    │
│  - Round-Robin: 轮询选择客户端                                    │
│  - Fill-First: 优先使用未使用的客户端                              │
│  - 优先级支持 (Priority 字段)                                     │
│  - 配额超限处理 (QuotaExceeded)                                   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    6. 协议转换层                                    │
│  - OpenAI ↔ Claude ↔ Gemini 格式转换                             │
│  - Thinking 后缀处理                                              │
│  - Payload 规则应用                                               │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    7. 上游提供商                                    │
│  - Gemini / Claude / Codex / Kiro / OpenAI Compatibility        │
│  - OAuth 认证 / API Key 认证                                     │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件详解

### 1. ModelRegistry (模型注册表)

**文件**: `internal/registry/model_registry.go`

模型注册表是路由系统的核心，负责：

```go
type ModelRegistry struct {
    models           map[string]*ModelRegistration  // 模型ID -> 注册信息
    clientModels     map[string][]string           // 客户端ID -> 模型列表
    clientModelInfos map[string]map[string]*ModelInfo // 客户端ID -> 模型信息
    clientProviders  map[string]string             // 客户端ID -> 提供商标识
    mutex            *sync.RWMutex
    hook             ModelRegistryHook              // 注册变更回调
}
```

**关键功能**:

| 功能 | 方法 | 说明 |
|------|------|------|
| 注册客户端 | `RegisterClient()` | 注册客户端及其支持的模型 |
| 注销客户端 | `UnregisterClient()` | 移除客户端并递减模型计数 |
| 查询模型 | `GetModelInfo()` | 获取模型信息（优先提供商特定） |
| 获取可用模型 | `GetAvailableModels()` | 获取有可用客户端的模型 |
| 模型计数 | `GetModelCount()` | 获取模型的可用客户端数量 |
| 提供商列表 | `GetModelProviders()` | 获取支持该模型的提供商列表 |

#### GetAvailableModels() 详解

**文件**: `internal/registry/model_registry.go:709`

**函数签名**:
```go
func (r *ModelRegistry) GetAvailableModels(handlerType string) []map[string]any
```

**功能概述**:
返回当前可用的模型列表，考虑客户端数量、配额超限状态和暂停状态。

**可用性判断逻辑**:

```go
// 配额冷却期：5 分钟
quotaExpiredDuration := 5 * time.Minute

// 计算有效客户端数
effectiveClients := availableClients - expiredClients - otherSuspended

// 模型包含条件：
// 1. effectiveClients > 0 - 有立即可用的客户端
// 2. 或：availableClients > 0 && (expiredClients > 0 || cooldownSuspended > 0) && otherSuspended == 0
//    - 有客户端但都在冷却期（配额超限），且没有其他暂停原因
```

**状态分类**:

| 状态 | 说明 | 处理方式 |
|------|------|----------|
| `expiredClients` | 配额超限且在冷却期（<5分钟） | 从有效客户端中扣除 |
| `cooldownSuspended` | 原因为 "quota" 的暂停 | 从有效客户端中扣除，但模型仍显示 |
| `otherSuspended` | 其他原因的暂停 | 从有效客户端中扣除，模型隐藏 |

**格式转换**:

函数根据 `handlerType` 参数返回不同格式的模型信息：

| handlerType | 格式 | 特殊字段 |
|-------------|------|----------|
| `openai` | OpenAI 兼容 | `id`, `object`, `owned_by`, `created`, `type`, `context_length` |
| `claude`/`kiro`/`antigravity` | Claude Code 兼容 | `thinking`, `extended_thinking` (支持 thinking 功能) |
| `gemini` | Gemini 格式 | `name`, `displayName`, `inputTokenLimit`, `outputTokenLimit` |
| 默认 | 通用格式 | `id`, `object`, `owned_by`, `created`, `type` |

#### GetAvailableModels() 详解

**文件**: `internal/registry/model_registry.go:709`

**函数签名**:
```go
func (r *ModelRegistry) GetAvailableModels(handlerType string) []map[string]any
```

**功能概述**:
返回当前可用的模型列表，考虑客户端数量、配额超限状态和暂停状态。

**可用性判断逻辑**:

```go
// 配额冷却期：5 分钟
quotaExpiredDuration := 5 * time.Minute

// 计算有效客户端数
effectiveClients := availableClients - expiredClients - otherSuspended

// 模型包含条件：
// 1. effectiveClients > 0 - 有立即可用的客户端
// 2. 或：availableClients > 0 && (expiredClients > 0 || cooldownSuspended > 0) && otherSuspended == 0
//    - 有客户端但都在冷却期（配额超限），且没有其他暂停原因
```

**状态分类**:

| 状态 | 说明 | 处理方式 |
|------|------|----------|
| `expiredClients` | 配额超限且在冷却期（<5分钟） | 从有效客户端中扣除 |
| `cooldownSuspended` | 原因为 "quota" 的暂停 | 从有效客户端中扣除，但模型仍显示 |
| `otherSuspended` | 其他原因的暂停 | 从有效客户端中扣除，模型隐藏 |

**格式转换**:

函数根据 `handlerType` 参数返回不同格式的模型信息：

| handlerType | 格式 | 特殊字段 |
|-------------|------|----------|
| `openai` | OpenAI 兼容 | `id`, `object`, `owned_by`, `created`, `type`, `context_length` |
| `claude`/`kiro`/`antigravity` | Claude Code 兼容 | `thinking`, `extended_thinking` (支持 thinking 功能) |
| `gemini` | Gemini 格式 | `name`, `displayName`, `inputTokenLimit`, `outputTokenLimit` |
| 默认 | 通用格式 | `id`, `object`, `owned_by`, `created`, `type` |

**配额管理**:

```go
// 标记模型配额超限
SetModelQuotaExceeded(clientID, modelID string)

// 清除配额超限状态
ClearModelQuotaExceeded(clientID, modelID string)

// 暂停客户端模型
SuspendClientModel(clientID, modelID, reason string)

// 恢复客户端模型
ResumeClientModel(clientID, modelID string)
```

### 2. 提供商识别

**文件**: `internal/util/provider.go`

```go
func GetProviderName(modelName string) []string
```

**识别流程**:

1. 首先查询 `ModelRegistry` 获取注册的提供商
2. 如果没有找到，使用启发式规则推断
3. 支持的提供商包括：
   - `gemini` - Google Gemini
   - `claude` - Anthropic Claude
   - `codex` - OpenAI Codex
   - `qwen` - Alibaba Qwen
   - `kiro` - AWS CodeWhisperer
   - `antigravity` - Antigravity
   - `vertex` - Google Vertex AI
   - `iflow` - iFlow
   - `kimi` - Moonshot Kimi
   - `github-copilot` - GitHub Copilot

### 3. 路由策略配置

**文件**: `internal/config/config.go`

```yaml
routing:
  strategy: "round-robin"  # 或 "fill-first"

quota-exceeded:
  switch-project: true       # 配额超限时自动切换项目
  switch-preview-model: true # 配额超限时切换预览模型
```

**策略说明**:

| 策略 | 行为 | 适用场景 |
|------|------|----------|
| `round-robin` | 轮询选择可用客户端 | 均匀分布负载 |
| `fill-first` | 优先使用未使用的客户端 | 最小化客户端使用数 |

### 4. 访问控制

**文件**: `sdk/access/manager.go`

```go
type Manager struct {
    mu        sync.RWMutex
    providers []Provider
}

func (m *Manager) Authenticate(ctx context.Context, r *http.Request) (*Result, *AuthError)
```

**认证流程**:

```
请求 → Provider1 → 失败
     → Provider2 → 成功 → 返回 Result
     → ...
```

**错误处理优先级**:

1. `NotHandled` - 继续尝试下一个提供者
2. `NoCredentials` - 记录缺失，继续尝试
3. `InvalidCredential` - 记录无效，继续尝试
4. 其他错误 - 直接返回

### 5. 认证管理器

**文件**: `sdk/auth/manager.go`

认证管理器负责协调 OAuth 和 API Key 认证：

```go
type Manager struct {
    mu              sync.RWMutex
    clients         []Client
    selector        Selector
    tokenStore      TokenStore
    cooldownManager *CooldownManager
}
```

**核心方法**:

```go
// 执行非流式请求
Execute(ctx context.Context, providers []string, req Request, opts Options) (*Response, error)

// 执行流式请求
ExecuteStream(ctx context.Context, providers []string, req Request, opts Options) (<-chan StreamChunk, error)

// 执行计数请求
ExecuteCount(ctx context.Context, providers []string, req Request, opts Options) (*Response, error)
```

### 6. 负载均衡选择器

**文件**: `sdk/cliproxy/auth/selector.go`

```go
type Selector interface {
    Select(model string, candidates []Client) Client
}
```

**选择策略**:

| 策略 | 实现类 | 行为 |
|------|--------|------|
| RoundRobin | `roundRobinSelector` | 轮询选择，跳过暂停/配额超限的客户端 |
| FillFirst | `fillFirstSelector` | 优先选择未使用的客户端 |
| Priority | `prioritySelector` | 按优先级排序后选择 |

**选择算法**:

```go
// Round-Robin 示例
func (s *roundRobinSelector) Select(model string, candidates []Client) Client {
    valid := make([]Client, 0)
    for _, c := range candidates {
        if !s.isSuspended(c, model) && !s.isQuotaExceeded(c, model) {
            valid = append(valid, c)
        }
    }
    if len(valid) == 0 {
        return nil
    }
    // 使用原子计数器轮询
    index := atomic.AddInt64(&s.counter, 1) % int64(len(valid))
    return valid[index]
}
```

## 模型别名系统

### 1. OAuth 模型别名

**配置**: `config.yaml`

```yaml
oauth-model-alias:
  kiro:
    - name: "kiro-claude-opus-4-5"
      alias: "op45"
    - name: "claude-sonnet-4-5"
      alias: "cs4.5"
```

**文件**: `internal/config/config.go`

```go
type OAuthModelAlias struct {
    Name  string `yaml:"name"`   // 上游模型名称
    Alias string `yaml:"alias"`  // 客户端可见别名
    Fork  bool   `yaml:"fork"`   // 是否保留原模型
}
```

### 2. 模型排除

```yaml
oauth-excluded-models:
  kiro:
    - "kiro-claude-haiku-4-5"     # 精确匹配
    - "kiro-claude-3-*"             # 前缀通配符
    - "*-preview"                   # 后缀通配符
```

### 3. Amp 模型映射

**文件**: `internal/api/modules/amp/model_mapping.go`

```yaml
ampcode:
  model-mappings:
    - from: "claude-opus-4-5-20251101"
      to: "gemini-claude-opus-4-5-thinking"
      regex: false  # 精确匹配
    - from: "claude-sonnet-.*"
      to: "gemini-claude-sonnet-4-5"
      regex: true   # 正则匹配
```

**映射逻辑**:

```go
func (m *DefaultModelMapper) MapModel(requestedModel string) string {
    // 1. 解析 thinking 后缀
    requestResult := thinking.ParseSuffix(requestedModel)
    baseModel := requestResult.ModelName

    // 2. 检查精确映射
    if target, exists := m.mappings[normalizedBase]; exists {
        // 3. 验证目标模型有可用提供商
        if providers := util.GetProviderName(target); len(providers) > 0 {
            // 4. 处理后缀（配置优先或保留用户后缀）
            return applySuffix(target, requestResult)
        }
    }

    // 5. 尝试正则映射
    for _, rm := range m.regexps {
        if rm.re.MatchString(baseModel) {
            return applySuffix(rm.to, requestResult)
        }
    }

    return ""
}
```

## Thinking 后缀处理

**文件**: `internal/thinking/suffix.go`

```go
type ParseResult struct {
    ModelName string  // 基础模型名称
    RawSuffix string  // 原始后缀内容
    HasSuffix  bool   // 是否有后缀
}

func ParseSuffix(model string) ParseResult
```

**支持格式**:

| 格式 | 示例 | 说明 |
|------|------|------|
| `模型(数字)` | `gemini-2.5-pro(8192)` | 指定思考预算 |
| `模型(auto)` | `claude-sonnet-4-5(auto)` | 自动预算 |
| `模型(级别)` | `gpt-5(high)` | 级别模式 |

**后缀传播规则**:

1. **Amp 映射**: 配置中的 `to` 字段后缀优先
2. **别名传播**: 保持原请求的后缀
3. **自动解析**: `auto` 模型会自动解析为可用模型

## Payload 规则系统

**文件**: `internal/runtime/executor/payload_helpers.go`

```yaml
payload:
  default:  # 仅当参数缺失时设置
    - models:
        - name: "gemini-2.5-pro"
          protocol: "gemini"
      params:
        "generationConfig.thinkingConfig.thinkingBudget": 32768

  default-raw:  # 使用原始 JSON 值
    - models:
        - name: "gemini-2.5-pro"
      params:
        "generationConfig.responseJsonSchema": "{\"type\":\"object\"}"

  override:  # 始终覆盖现有值
    - models:
        - name: "gpt-*"
      params:
        "reasoning.effort": "high"

  filter:  # 删除指定参数
    - models:
        - name: "gemini-2.5-pro"
      params:
        - "generationConfig.thinkingConfig.thinkingBudget"
```

**规则匹配**:

```go
func matchModelPattern(pattern, model string) bool {
    // 支持 * 通配符
    // "gpt-*" 匹配 "gpt-5" 和 "gpt-4"
    // "*-5" 匹配 "gpt-5"
    // "gemini-*-pro" 匹配 "gemini-2.5-pro"
}
```

## 配额超限处理

**文件**: `sdk/cliproxy/auth/selector.go`

**处理流程**:

```
请求 → 选择客户端 → 执行请求
                    ↓
              检查响应状态
                    ↓
         ┌──────────┴──────────┐
         │                     │
      配额超限                正常
         │                     │
         ↓                     ↓
  标记客户端超限          返回成功
         │
         ↓
  检查 switch-project
         │
    ┌────┴────┐
    │         │
   是         否
    │         │
    ↓         ↓
切换项目    检查 switch-preview-model
           │
      ┌────┴────┐
      │         │
     是         否
      │         │
      ↓         ↓
 切换预览模型  返回错误
```

## 完整路由流程示例

### 示例 1: 简单请求

```
客户端请求: POST /v1/chat/completions
Body: { "model": "claude-sonnet-4-5", ... }

1. 解析模型名: "claude-sonnet-4-5"
2. 查询提供商: ["claude"]
3. 访问控制: 验证 API 密钥
4. 选择客户端: round-robin 选择 claude 客户端
5. 协议转换: OpenAI → Claude 格式
6. 执行请求: 发送到 Anthropic API
7. 返回响应: Claude → OpenAI 格式
```

### 示例 2: 带 Thinking 后缀

```
客户端请求: POST /v1/chat/completions
Body: { "model": "gemini-2.5-pro(8192)", ... }

1. 解析模型名: base="gemini-2.5-pro", suffix="8192"
2. 查询提供商: ["gemini"]
3. 选择客户端: 轮询选择 gemini 客户端
4. Payload 规则: 应用 default 规则设置 thinkingBudget
5. 协议转换: OpenAI → Gemini 格式
6. 执行请求: 发送到 Gemini API
7. 返回响应: Gemini → OpenAI 格式
```

### 示例 3: Amp 模型映射

```
客户端请求: POST /v1/chat/completions
Body: { "model": "claude-opus-4-5", ... }

配置:
  model-mappings:
    - from: "claude-opus-4-5"
      to: "gemini-claude-opus-4-5-thinking"

1. 解析模型名: "claude-opus-4-5"
2. Amp 映射: 检测到映射，转换为 "gemini-claude-opus-4-5-thinking"
3. 查询提供商: ["gemini"]
4. 选择客户端: 轮询选择 gemini 客户端
5. 执行请求: 发送到 Gemini API
6. 返回响应: Gemini → OpenAI 格式
```

### 示例 4: 配额超限恢复

```
客户端请求: POST /v1/chat/completions
Body: { "model": "kiro-claude-sonnet-4-5", ... }

1. 解析模型名: "kiro-claude-sonnet-4-5"
2. 查询提供商: ["kiro"]
3. 选择客户端: client-1
4. 执行请求: 返回 429 (配额超限)
5. 标记超限: client-1 标记为 quota-exceeded
6. 重试请求: 选择 client-2
7. 执行请求: 返回成功
8. 冷却管理: 5 分钟后自动恢复 client-1
```

## 性能优化

### 1. 缓存策略

- **模型信息缓存**: ModelRegistry 使用内存缓存
- **提供商识别缓存**: GetProviderName 结果缓存
- **认证令牌缓存**: TokenStore 持久化 OAuth 令牌

### 2. 并发控制

```go
// 使用读写锁保护共享数据
type ModelRegistry struct {
    mutex *sync.RWMutex
    models map[string]*ModelRegistration
}

// 原子计数器实现无锁轮询
type roundRobinSelector struct {
    counter int64  // atomic.AddInt64
}
```

### 3. 流式响应优化

```go
// Bootstrap 重试机制
StreamingBootstrapRetries(cfg *config.SDKConfig) int

// 在发送任何字节前重试，允许认证轮换
if !sentPayload && bootstrapRetries < maxBootstrapRetries {
    retryChunks, retryErr := h.AuthManager.ExecuteStream(...)
    if retryErr == nil {
        chunks = retryChunks
        continue outer
    }
}
```

## 监控和调试

### 1. 请求日志

```
格式: [timestamp] [level] | request_id | status | latency | client_ip | method | path | provider | model | account

示例: [2025-01-28 04:00:00] [info ] | a1b2c3d4 | 200 | 23.559s | 192.168.1.100 | POST | /v1/chat/completions | provider=gemini | model=gemini-pro | account=oauth:user@example.com
```

### 2. 配额监控

```go
// 获取可用客户端数
GetModelCount(modelID string) int

// 获取模型提供商列表
GetModelProviders(modelID string) []string

// 获取客户端支持的模型
GetModelsForClient(clientID string) []*ModelInfo
```

### 3. 管理端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/v0/management/models` | GET | 获取所有可用模型 |
| `/v0/management/clients` | GET | 获取所有客户端状态 |
| `/v0/management/quota` | GET | 获取配额状态 |
| `/v0/management/usage` | GET | 获取使用统计 |

## 配置示例

### 完整路由配置示例

```yaml
# 路由策略
routing:
  strategy: "round-robin"  # 或 "fill-first"

# 配额超限处理
quota-exceeded:
  switch-project: true
  switch-preview-model: true

# OAuth 模型别名
oauth-model-alias:
  kiro:
    - name: "kiro-claude-opus-4-5"
      alias: "op45"
    - name: "claude-sonnet-4-5"
      alias: "cs4.5"

# OAuth 模型排除
oauth-excluded-models:
  kiro:
    - "kiro-claude-haiku-4-5"

# Payload 规则
payload:
  default:
    - models:
        - name: "gemini-2.5-pro"
          protocol: "gemini"
      params:
        "generationConfig.thinkingConfig.thinkingBudget": 32768

# Amp 模型映射
ampcode:
  model-mappings:
    - from: "claude-opus-4-5-20251101"
      to: "gemini-claude-opus-4-5-thinking"
```

## 总结

CLIProxyAPI Plus 的路由系统具有以下特点：

1. **多层架构**: 从请求解析到上游调用的完整分层设计
2. **灵活策略**: 支持 round-robin 和 fill-first 负载均衡
3. **模型映射**: 支持别名、排除和 Amp 映射
4. **配额管理**: 自动处理配额超限和冷却恢复
5. **协议转换**: 无缝支持多种 API 格式
6. **高可用**: Bootstrap 重试和客户端故障转移
7. **可扩展**: 支持自定义认证提供者和路由规则