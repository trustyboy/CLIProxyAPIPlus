# Redis 统计缓存功能说明

## 概述

Redis 统计缓存功能允许将使用统计数据持久化到 Redis，支持服务重启后数据不丢失。

## 功能特性

- **数据持久化**：统计数据存储在 Redis 中，服务重启后自动恢复
- **灵活配置**：支持自定义 Redis 连接、密码、数据库、key 前缀和 TTL
- **自动回退**：未启用 Redis 时自动使用内存存储
- **前端可视化配置**：通过管理界面配置 Redis 连接参数

## 配置说明

### 后端配置 (config.yaml)

```yaml
usage-statistics-enabled: true
usage-statistics-cache:
  enable: true
  addr: localhost:6379
  password: ""  # 可选，留空表示无密码
  db: 0
  key-prefix: "cliproxy:usage:"
  ttl: 86400
```

### 配置参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| enable | bool | false | 是否启用 Redis 缓存 |
| addr | string | "" | Redis 地址，如 `localhost:6379` |
| password | string | "" | Redis 密码，可留空 |
| db | int | 0 | Redis 数据库编号 |
| key-prefix | string | "cliproxy:usage:" | Redis key 前缀 |
| ttl | int | 86400 | 缓存过期时间（秒），默认 1 天 |

## 前端配置

在系统设置页面中：
1. 启用"使用统计"
2. 启用"Redis 缓存"
3. 填写 Redis 连接信息

## 架构设计

### 关键文件

- `internal/config/config.go` - 配置结构定义
- `internal/cache/redis.go` - Redis 客户端封装
- `internal/usage/stats.go` - 统计存储接口和实现
- `internal/api/server.go` - 服务初始化
- `internal/api/handlers/management/handler.go` - 管理接口
- `internal/usage/logger_plugin.go` - 统计记录插件

### 核心接口

```go
// StatsStorage 定义统计存储接口
type StatsStorage interface {
    Record(ctx context.Context, record coreusage.Record)
    Snapshot() StatisticsSnapshot
    MergeSnapshot(snapshot StatisticsSnapshot) MergeResult
}
```

### 存储实现

1. **memoryStatsStorage** - 内存存储（默认）
2. **redisStatsStorage** - Redis 存储

## 使用示例

### 配置文件示例

```yaml
usage-statistics-enabled: true
usage-statistics-cache:
  enable: true
  addr: redis.example.com:6379
  password: mypassword
  db: 0
  key-prefix: "myapp:usage:"
  ttl: 86400
```

### 验证方法

1. 启动服务后发起 API 请求
2. 查看 Redis 中的 key：`redis-cli KEYS cliproxy:usage:*`
3. 查看统计数据：`redis-cli GET cliproxy:usage:total`
4. 重启服务，验证数据是否恢复

## 注意事项

- Redis 密码不会暴露在 JSON 配置中（使用 `json:"-"` 标签）
- 使用独立的 background context 避免请求取消导致的写入失败
- TTL 过期后统计数据会自动清除
