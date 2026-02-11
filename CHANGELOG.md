# 更新日志

CLIProxyAPIPlus 项目（gf 分支）的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

## [未发布]

### 新增功能

#### 模型可用性管理页面
- 新增独立的模型可用性管理页面 (`/model-availability`)
- 展示当前处于不可用状态的模型列表，包括：
  - 模型名称和 ID
  - 供应商 (Provider)
  - 凭证 (Client ID)
  - 不可用原因（配额超限/已暂停/冷却中）
  - 不可用开始时间
- 提供手动重置功能，可恢复模型的可用状态
- 支持中英文国际化
- 通过侧边栏导航访问，独立于 Dashboard

**新增 API 端点：**
- `GET /v0/management/model-availability` - 获取不可用模型列表
- `POST /v0/management/model-availability/:model_id/reset` - 重置模型可用性

**涉及文件：**
- 后端：`internal/api/handlers/management/model_availability.go`
- 后端：`internal/registry/model_registry.go` (添加辅助方法)
- 前端：`web/src/pages/ModelAvailabilityPage.tsx`
- 前端：`web/src/services/api/modelAvailability.ts`

#### 增强日志功能
- AI API 请求的请求 ID 追踪（v1/chat/completions、v1/completions、v1/messages、v1/responses）
- 在日志输出中显示提供商、模型和账号信息
- 结构化日志字段，便于日志分析和过滤
- 日志格式：`[时间戳] [级别] | 请求ID | 状态 | 耗时 | 客户端IP | 方法 | 路径 | provider=X | model=Y | account=Z`

#### 嵌入式管理界面
- 管理界面直接嵌入到二进制文件中（embedded/management.html）
- 支持离线部署 - 无需网络下载
- 更快的启动速度，UI 立即可用
- UI 和后端版本一致性

#### 自动化构建脚本
- 自动化构建脚本（build.sh），具有以下功能：
  - 检查 git 更新，如果没有新提交则自动退出
  - 自动停止/启动 supervisor 服务
  - 使用 `-f` 参数强制构建模式
  - 自动构建并嵌入 React 前端
  - 首次构建时自动安装 npm 依赖
  - 将版本信息注入二进制文件

#### Web 子模块
- Web 前端集成为 git 子模块
- Web 开发使用独立仓库
- 自动子模块更新

### 变更

#### 构建流程
- 简化代码拉取逻辑
- 优化子模块更新流程
- 更新时保护未跟踪的配置文件
- 从 build.sh 中移除 proxychains 依赖
- 修复 pull_code 函数使用正确的远程分支

#### 日志功能
- 修复多个渠道使用相同模型时的渠道显示问题
- 仅在模型可用时显示提供商和模型信息
- 从 gin.Context 读取实际提供商，而不是从模型名推断
- 从 gin.Context 读取实际模型（由 auth manager 设置）作为主要来源，fallback 到从请求体提取

#### 认证持久化
- 修复跨重启的禁用状态持久化
- Auth 元数据正确合并到 auth 文件中
- 确保禁用的 auth 在服务重启后保持禁用状态

### 技术细节

#### 日志实现
**修改的文件：**
- `internal/logging/gin_logger.go`：增强以提取和记录提供商、模型和账号信息
- `sdk/cliproxy/auth/conductor.go`：将 routeModel、provider 和账号信息存储到 gin.Context

**关键变更：**
- 在 `executeMixedOnce`、`executeCountMixedOnce` 和 `executeStreamMixedOnce` 中：
  ```go
  if ginCtx := ctx.Value("gin"); ginCtx != nil {
      if c, ok := ginCtx.(*gin.Context); ok {
          c.Set("cliproxy.provider", provider)
          c.Set("cliproxy.model", routeModel)
      }
  }
  ```
- 在 `GinLogrusLogger` 中：
  ```go
  // 首先尝试从 gin.Context 获取模型（由 auth manager 设置）
  model := ""
  if modelVal, exists := c.Get("cliproxy.model"); exists {
      if modelStr, ok := modelVal.(string); ok {
          model = modelStr
      }
  }
  // Fallback 到从请求体提取
  if model == "" {
      model = extractModelFromRequest(c)
  }
  ```

#### 认证持久化实现
**修改的文件：**
- `sdk/auth/filestore.go`：添加 `mergeMetadataIntoFile()` 函数
- `internal/watcher/synthesizer/file.go`：正确处理禁用状态

**关键变更：**
```go
// 保存存储后，将元数据（如禁用状态）合并到文件中
if auth.Metadata != nil && len(auth.Metadata) > 0 {
    if err = s.mergeMetadataIntoFile(path, auth); err != nil {
        return "", fmt.Errorf("auth filestore: merge metadata failed: %w", err)
    }
}
```

#### 构建脚本改进
**修改的文件：**
- `build.sh`：完全重写，具有自动化功能

**关键功能：**
- 构建前检查 git 更新
- 自动服务管理
- Web 前端构建和嵌入
- 版本注入
- 子模块管理

### 破坏性变更

无

### 迁移说明

#### 从主分支迁移到 GF 分支

1. **更新 README**：gf 分支包含主分支中没有的增强功能
2. **构建流程**：使用 `./build.sh` 而不是手动 `go build`
3. **Web 部署**：Web 现在是 git 子模块，使用以下命令更新：
   ```bash
   git submodule update --remote web
   ```
4. **日志格式**：日志现在包含额外字段（provider、model、account）
5. **管理界面**：通过嵌入的 `/management.html` 访问，而不是下载
