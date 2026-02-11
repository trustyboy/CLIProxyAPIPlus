# 模型可用性功能实现总结

## 实现概述

创建了独立的"模型可用性"页面，用于展示当前处于不可用状态的模型列表，并提供手动重置功能。该页面可通过侧边栏导航访问。

## 实现内容

### 1. 后端 API (Go)

**文件**: `internal/api/handlers/management/model_availability.go`

新增两个管理端点：

- **GET /v0/management/model-availability** - 获取所有不可用模型列表
  - 返回模型名称、供应商、凭证、不可用原因、开始时间等信息
  - 支持三种不可用原因：quota_exceeded（配额超限）、suspended（已暂停）、cooldown（冷却中）

- **POST /v0/management/model-availability/:model_id/reset** - 重置指定模型的不可用状态
  - 清除配额超限状态
  - 恢复暂停状态

**文件**: `internal/registry/model_registry.go`

新增两个辅助方法：
- `GetAllModels()` - 获取所有注册的模型
- `GetClientProvider()` - 获取客户端的 provider 标识

**文件**: `internal/api/server.go`

注册新的管理路由：
```go
mgmt.GET("/model-availability", s.mgmt.GetUnavailableModels)
mgmt.POST("/model-availability/:model_id/reset", s.mgmt.ResetModelAvailability)
```

### 2. 前端 API 服务 (TypeScript)

**文件**: `web/src/services/api/modelAvailability.ts`

创建 API 服务：
```typescript
export const modelAvailabilityApi = {
  async getUnavailableModels(): Promise<UnavailableModelsResponse>,
  async resetModelAvailability(modelId: string, clientId: string): Promise<ResetModelAvailabilityResponse>
};
```

**文件**: `web/src/services/api/index.ts`

导出新增的 API 服务。

### 3. 前端独立页面

**文件**: `web/src/pages/ModelAvailabilityPage.tsx`

功能特性：
- 独立页面展示不可用模型列表
- 显示模型名称、Provider、Client ID、原因、开始时间
- 每行提供"重置"按钮
- 空状态提示（当所有模型都可用时）
- 刷新按钮
- 加载状态
- 操作成功/失败通知
- 页面标题显示不可用模型数量

**文件**: `web/src/pages/ModelAvailabilityPage.module.scss`

样式特点：
- 与现有页面风格一致
- 响应式表格设计
- 不同原因使用不同颜色标识（配额超限-黄色、已暂停-红色）

### 4. 路由与导航集成

**文件**: `web/src/router/MainRoutes.tsx`

添加独立页面路由：
```typescript
{ path: '/model-availability', element: <ModelAvailabilityPage /> }
```

**文件**: `web/src/components/layout/MainLayout.tsx`

在侧边栏导航中添加菜单项：
```typescript
{ path: '/model-availability', label: t('nav.model_availability'), icon: sidebarIcons.modelAvailability }
```

### 5. 国际化支持

**文件**: `web/src/i18n/locales/zh-CN.json` 和 `en.json`

添加翻译键：
- `model_availability.title` - 标题
- `model_availability.model_name` - 模型名称
- `model_availability.provider` - 供应商
- `model_availability.client` - 凭证
- `model_availability.reason` - 原因
- `model_availability.since` - 开始时间
- `model_availability.actions` - 操作
- `model_availability.reset` - 重置按钮
- `model_availability.reset_success` - 重置成功提示
- `model_availability.reset_error` - 重置失败提示
- `model_availability.no_unavailable` - 空状态标题
- `model_availability.no_unavailable_desc` - 空状态描述
- `model_availability.reason_quota_exceeded` - 配额超限
- `model_availability.reason_suspended` - 已暂停
- `model_availability.reason_cooldown` - 冷却中

## 关键代码片段

### 后端 - 获取不可用模型
```go
func (h *Handler) GetUnavailableModels(c *gin.Context) {
    reg := registry.GetGlobalRegistry()
    unavailableModels := make([]UnavailableModelInfo, 0)

    models := reg.GetAllModels()
    now := time.Now()
    quotaExpiredDuration := 5 * time.Minute

    for modelID, registration := range models {
        // 检查配额超限的客户端
        if registration.QuotaExceededClients != nil {
            for clientID, quotaTime := range registration.QuotaExceededClients {
                if quotaTime != nil && now.Sub(*quotaTime) < quotaExpiredDuration {
                    // 添加到不可用列表
                }
            }
        }
        // 检查被暂停的客户端
        if registration.SuspendedClients != nil {
            for clientID, reason := range registration.SuspendedClients {
                // 添加到不可用列表
            }
        }
    }
}
```

### 前端 - 重置模型可用性
```typescript
const handleReset = async (model: UnavailableModel) => {
  try {
    await modelAvailabilityApi.resetModelAvailability(model.model_id, model.client_id);
    showNotification(t('model_availability.reset_success'), 'success');
    await fetchUnavailableModels();
  } catch (error) {
    showNotification(t('model_availability.reset_error'), 'error');
  }
};
```

## 测试验证

1. **后端编译**: `go build ./...` - 通过
2. **前端编译**: `npm run build` - 通过

## 涉及文件清单

### 后端
- `internal/api/handlers/management/model_availability.go` (新建)
- `internal/registry/model_registry.go` (修改 - 添加辅助方法)
- `internal/api/server.go` (修改 - 注册路由)

### 前端
- `web/src/services/api/modelAvailability.ts` (新建)
- `web/src/services/api/index.ts` (修改 - 导出)
- `web/src/pages/ModelAvailabilityPage.tsx` (新建 - 独立页面)
- `web/src/pages/ModelAvailabilityPage.module.scss` (新建 - 页面样式)
- `web/src/router/MainRoutes.tsx` (修改 - 添加路由)
- `web/src/components/layout/MainLayout.tsx` (修改 - 添加导航菜单)
- `web/src/pages/DashboardPage.tsx` (修改 - 移除旧组件引用)
- `web/src/i18n/locales/zh-CN.json` (修改 - 中文翻译)
- `web/src/i18n/locales/en.json` (修改 - 英文翻译)

## 后续建议

1. **权限控制**: 当前复用了现有的 management 中间件，如有需要可添加更细粒度的权限控制
2. **批量重置**: 可考虑添加批量重置功能（选择多个模型一键重置）
3. **自动刷新**: 可考虑添加定时自动刷新功能
4. **历史记录**: 可考虑添加重置操作的历史记录

---

**实现日期**: 2026-02-12
**状态**: 已完成
