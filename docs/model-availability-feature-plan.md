# 模型可用性功能实施计划

## 概述

在 Dashboard 页面上新增"模型可用性"功能模块，用于展示当前处于不可用状态的模型列表，并提供手动重置功能。

## 需求详情

1. **展示当前处于不可用状态的模型列表**，包括：
   - 模型名称（Model Name）
   - 模型供应商（Provider）
   - 凭证信息（Client/Credential）
   - 不可用原因（quota exceeded / suspended / cooldown 等）

2. **提供重置功能**：每行提供一个"重置"按钮，手动重置该模型的不可用状态，使其恢复可用。

## 架构分析

### 后端现有结构

**ModelRegistry**（`internal/registry/model_registry.go`）：
- `ModelRegistration` 结构体已包含不可用状态追踪字段：
  - `QuotaExceededClients map[string]*time.Time` - 记录超出配额的客户端
  - `SuspendedClients map[string]string` - 记录被暂停的客户端及原因
- 已有相关方法：`SetModelQuotaExceeded`, `ClearModelQuotaExceeded`, `SuspendClientModel`, `ResumeClientModel`

**管理端点**（`internal/api/handlers/management/`）：
- 管理 Handler 位于 `handler.go`
- 新端点需要在 `server.go` 的 `registerManagementRoutes()` 中注册

### 前端现有结构

**Dashboard 页面**（`web/src/pages/DashboardPage.tsx`）：
- 当前展示连接状态、快速统计、配置信息等
- 需要在合适位置添加"模型可用性"卡片/表格

**API 服务**（`web/src/services/api/`）：
- 需要新增模型可用性相关的 API 调用

## 实施阶段

### Phase 1: 后端 API 开发

#### 步骤 1.1: 扩展 ModelRegistry 接口（如需要）

**文件**: `internal/registry/model_registry.go`

**检查点**:
- [ ] 确认现有 `QuotaExceededClients` 和 `SuspendedClients` 数据结构是否满足需求
- [ ] 确认是否需要添加获取所有不可用模型的方法

**预估复杂度**: 低（现有数据结构可能已满足）

#### 步骤 1.2: 创建模型可用性管理端点

**文件**: 新建 `internal/api/handlers/management/model_availability.go`

**实现内容**:

```go
// GET /v0/management/model-availability
// 返回所有不可用模型列表
type UnavailableModelInfo struct {
    ModelID    string    `json:"model_id"`
    ModelName  string    `json:"model_name"`
    Provider   string    `json:"provider"`
    ClientID   string    `json:"client_id"`
    Reason     string    `json:"reason"`      // "quota_exceeded", "suspended", "cooldown"
    ReasonText string    `json:"reason_text"` // 详细原因描述
    Since      time.Time `json:"since"`       // 不可用开始时间
}

// POST /v0/management/model-availability/:model_id/reset
// 重置指定模型的不可用状态
// 请求体: { "client_id": "xxx" }
```

**预估复杂度**: 中

#### 步骤 1.3: 注册管理路由

**文件**: `internal/api/server.go`

**修改位置**: `registerManagementRoutes()` 方法

**添加路由**:

```go
mgmt.GET("/model-availability", s.mgmt.GetUnavailableModels)
mgmt.POST("/model-availability/:model_id/reset", s.mgmt.ResetModelAvailability)
```

**预估复杂度**: 低

### Phase 2: 前端 API 服务层

#### 步骤 2.1: 创建模型可用性 API 服务

**文件**: 新建 `web/src/services/api/modelAvailability.ts`

**实现内容**:

```typescript
export interface UnavailableModel {
  model_id: string;
  model_name: string;
  provider: string;
  client_id: string;
  reason: 'quota_exceeded' | 'suspended' | 'cooldown';
  reason_text: string;
  since: string;
}

export const modelAvailabilityApi = {
  // 获取不可用模型列表
  async getUnavailableModels(): Promise<UnavailableModel[]>;

  // 重置模型可用性
  async resetModelAvailability(modelId: string, clientId: string): Promise<void>;
};
```

**预估复杂度**: 低

#### 步骤 2.2: 导出 API 服务

**文件**: `web/src/services/api/index.ts`

**修改**: 添加 `modelAvailabilityApi` 导出

**预估复杂度**: 低

### Phase 3: 前端 UI 组件开发

#### 步骤 3.1: 创建模型可用性组件

**文件**: 新建 `web/src/components/modelAvailability/ModelAvailabilitySection.tsx`

**实现内容**:
- 表格展示不可用模型列表
- 列：模型名称、Provider、Client ID、原因、操作（重置按钮）
- 空状态提示（无 unavailable 模型时）
- 加载状态
- 错误处理

**预估复杂度**: 中

#### 步骤 3.2: 创建组件样式

**文件**: 新建 `web/src/components/modelAvailability/ModelAvailabilitySection.module.scss`

**样式要求**:
- 与现有 Dashboard 卡片样式一致
- 表格响应式设计
- 重置按钮样式（参考现有按钮样式）

**预估复杂度**: 低

#### 步骤 3.3: 创建组件入口

**文件**: 新建 `web/src/components/modelAvailability/index.ts`

**预估复杂度**: 低

### Phase 4: 集成到 Dashboard

#### 步骤 4.1: 修改 Dashboard 页面

**文件**: `web/src/pages/DashboardPage.tsx`

**修改内容**:
- 导入 `ModelAvailabilitySection` 组件
- 在合适位置（建议在"可用模型"统计卡片下方）添加模型可用性展示区域
- 添加刷新逻辑（与现有统计数据一起刷新）

**预估复杂度**: 低

### Phase 5: 国际化支持

#### 步骤 5.1: 添加翻译键

**文件**: `web/src/i18n/locales/zh.json` 和 `web/src/i18n/locales/en.json`

**需要添加的键**:

```json
{
  "model_availability": {
    "title": "模型可用性",
    "subtitle": "当前不可用模型列表",
    "model_name": "模型名称",
    "provider": "供应商",
    "client": "凭证",
    "reason": "原因",
    "actions": "操作",
    "reset": "重置",
    "reset_success": "已重置模型可用性",
    "reset_error": "重置失败",
    "no_unavailable": "所有模型均可用",
    "reason_quota_exceeded": "配额超限",
    "reason_suspended": "已暂停",
    "reason_cooldown": "冷却中"
  }
}
```

**预估复杂度**: 低

## 涉及的关键文件汇总

### 后端文件
| 文件路径 | 操作 | 说明 |
|---------|------|------|
| `internal/api/handlers/management/model_availability.go` | 新建 | 模型可用性管理端点 |
| `internal/api/server.go` | 修改 | 注册新路由 |
| `internal/registry/model_registry.go` | 可能修改 | 如需要添加批量查询方法 |

### 前端文件
| 文件路径 | 操作 | 说明 |
|---------|------|------|
| `web/src/services/api/modelAvailability.ts` | 新建 | API 服务 |
| `web/src/services/api/index.ts` | 修改 | 导出新增服务 |
| `web/src/components/modelAvailability/ModelAvailabilitySection.tsx` | 新建 | 主组件 |
| `web/src/components/modelAvailability/ModelAvailabilitySection.module.scss` | 新建 | 组件样式 |
| `web/src/components/modelAvailability/index.ts` | 新建 | 组件入口 |
| `web/src/pages/DashboardPage.tsx` | 修改 | 集成到 Dashboard |
| `web/src/i18n/locales/zh.json` | 修改 | 中文翻译 |
| `web/src/i18n/locales/en.json` | 修改 | 英文翻译 |

## 风险评估与缓解

| 风险 | 级别 | 描述 | 缓解措施 |
|------|------|------|---------|
| ModelRegistry 并发访问 | 中 | 读取不可用模型列表时需要加锁 | 使用 RLock 确保线程安全 |
| 重置操作误触 | 低 | 用户可能误点击重置按钮 | 添加确认对话框或二次确认 |
| API 性能问题 | 低 | 模型数量过多时查询慢 | 实现分页或限制返回数量 |
| 前端状态同步 | 低 | 重置后列表未及时更新 | 重置成功后立即刷新列表 |
| 权限控制 | 低 | 重置功能需要管理权限 | 复用现有 management 中间件 |

## 预计复杂度

| 阶段 | 预计工时 | 复杂度 |
|------|---------|--------|
| Phase 1: 后端 API | 2-3 小时 | 中 |
| Phase 2: 前端 API 层 | 1 小时 | 低 |
| Phase 3: 前端 UI 组件 | 3-4 小时 | 中 |
| Phase 4: Dashboard 集成 | 1 小时 | 低 |
| Phase 5: 国际化 | 30 分钟 | 低 |
| **总计** | **7.5-9.5 小时** | **中** |

## 依赖关系图

```
Phase 1 (后端 API)
    │
    ▼
Phase 2 (前端 API 层) ─────┐
    │                      │
    ▼                      │
Phase 3 (UI 组件)          │
    │                      │
    ▼                      │
Phase 4 (Dashboard 集成) ◄─┘
    │
    ▼
Phase 5 (国际化)
```

## 成功标准

- [ ] Dashboard 页面正确展示不可用模型列表
- [ ] 列表包含模型名称、Provider、Client ID、原因
- [ ] 点击重置按钮后模型从列表中移除
- [ ] 重置操作调用后端 API 成功
- [ ] 空状态正确显示（当所有模型都可用时）
- [ ] 国际化支持完整（中英文）
- [ ] 代码通过 review，符合项目规范

---

**创建日期**: 2026-02-12
**状态**: 待实施
