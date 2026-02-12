// Package usage provides usage tracking and storage abstractions.
package usage

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/cache"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"

	coreusage "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
	log "github.com/sirupsen/logrus"
)

// StatsStorage defines the interface for usage statistics storage.
type StatsStorage interface {
	// Record records a usage record.
	Record(ctx context.Context, record coreusage.Record)

	// Snapshot returns a copy of the aggregated metrics.
	Snapshot() StatisticsSnapshot

	// MergeSnapshot merges an exported statistics snapshot into the current store.
	MergeSnapshot(snapshot StatisticsSnapshot) MergeResult
}

// NewStatsStorage creates a new stats storage based on configuration.
func NewStatsStorage(cfg config.RedisCacheConfig) StatsStorage {
	if cfg.Enable {
		return &redisStatsStorage{
			config: cfg,
		}
	}
	return &memoryStatsStorage{
		stats: NewRequestStatistics(),
	}
}

var defaultStatsStorage StatsStorage

// InitStatsStorage initializes the global stats storage with the given configuration.
func InitStatsStorage(cfg config.RedisCacheConfig) {
	defaultStatsStorage = NewStatsStorage(cfg)
}

// GetStatsStorage returns the global stats storage instance.
func GetStatsStorage() StatsStorage {
	if defaultStatsStorage == nil {
		// Fallback to memory storage if not initialized
		return &memoryStatsStorage{
			stats: GetRequestStatistics(),
		}
	}
	return defaultStatsStorage
}

// memoryStatsStorage implements StatsStorage using in-memory storage.
type memoryStatsStorage struct {
	stats *RequestStatistics
}

func (s *memoryStatsStorage) Record(ctx context.Context, record coreusage.Record) {
	if s.stats != nil {
		s.stats.Record(ctx, record)
	}
}

func (s *memoryStatsStorage) Snapshot() StatisticsSnapshot {
	if s.stats == nil {
		return StatisticsSnapshot{}
	}
	return s.stats.Snapshot()
}

func (s *memoryStatsStorage) MergeSnapshot(snapshot StatisticsSnapshot) MergeResult {
	if s.stats == nil {
		return MergeResult{}
	}
	return s.stats.MergeSnapshot(snapshot)
}

// redisStatsStorage implements StatsStorage using Redis.
type redisStatsStorage struct {
	config config.RedisCacheConfig
	mu     sync.RWMutex
}

const (
	statsTotalKey       = "total"
	statsAPIsKey        = "apis"
	statsRequestsByDay  = "requests_by_day"
	statsRequestsByHour = "requests_by_hour"
	statsTokensByDay    = "tokens_by_day"
	statsTokensByHour   = "tokens_by_hour"
)

func (s *redisStatsStorage) key(prefix string) string {
	return s.config.KeyPrefix + prefix
}

func (s *redisStatsStorage) Record(ctx context.Context, record coreusage.Record) {
	client := cache.GetClient()
	if client == nil {
		return
	}

	// Use background context to avoid context cancellation issues
	// The request context may be canceled before Redis operations complete
	bgCtx := context.Background()

	// Get current snapshot, update it, and write back
	// This is a simplified approach - in production, consider using Lua scripts for atomicity
	snapshot := s.Snapshot()

	// Convert record to detail
	timestamp := record.RequestedAt
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	detail := normalizeRecordDetail(record)
	totalTokens := detail.TotalTokens

	statsKey := record.APIKey
	if statsKey == "" {
		statsKey = "unknown"
	}
	failed := record.Failed
	success := !failed
	modelName := record.Model
	if modelName == "" {
		modelName = "unknown"
	}

	dayKey := timestamp.Format("2006-01-02")
	hourKey := timestamp.Hour()

	// Update snapshot
	snapshot.TotalRequests++
	if success {
		snapshot.SuccessCount++
	} else {
		snapshot.FailureCount++
	}
	snapshot.TotalTokens += totalTokens

	// Update API stats
	if snapshot.APIs == nil {
		snapshot.APIs = make(map[string]APISnapshot)
	}
	apiSnapshot, ok := snapshot.APIs[statsKey]
	if !ok {
		apiSnapshot = APISnapshot{Models: make(map[string]ModelSnapshot)}
	}
	apiSnapshot.TotalRequests++
	apiSnapshot.TotalTokens += totalTokens

	if apiSnapshot.Models == nil {
		apiSnapshot.Models = make(map[string]ModelSnapshot)
	}
	modelSnapshot, ok := apiSnapshot.Models[modelName]
	if !ok {
		modelSnapshot = ModelSnapshot{}
	}
	modelSnapshot.TotalRequests++
	modelSnapshot.TotalTokens += totalTokens
	modelSnapshot.Details = append(modelSnapshot.Details, RequestDetail{
		Timestamp: timestamp,
		Source:    record.Source,
		AuthIndex: record.AuthIndex,
		Tokens:    detail,
		Failed:    failed,
	})
	apiSnapshot.Models[modelName] = modelSnapshot
	snapshot.APIs[statsKey] = apiSnapshot

	// Update time-based stats
	if snapshot.RequestsByDay == nil {
		snapshot.RequestsByDay = make(map[string]int64)
	}
	snapshot.RequestsByDay[dayKey]++

	if snapshot.RequestsByHour == nil {
		snapshot.RequestsByHour = make(map[string]int64)
	}
	snapshot.RequestsByHour[formatHour(hourKey)]++

	if snapshot.TokensByDay == nil {
		snapshot.TokensByDay = make(map[string]int64)
	}
	snapshot.TokensByDay[dayKey] += totalTokens

	if snapshot.TokensByHour == nil {
		snapshot.TokensByHour = make(map[string]int64)
	}
	snapshot.TokensByHour[formatHour(hourKey)] += totalTokens

	// Write back to Redis
	s.saveSnapshot(bgCtx, snapshot)
}

func (s *redisStatsStorage) Snapshot() StatisticsSnapshot {
	client := cache.GetClient()
	if client == nil {
		return StatisticsSnapshot{}
	}

	ctx := context.Background()
	snapshot := StatisticsSnapshot{}

	// Load total stats
	totalData, err := client.Get(ctx, s.key(statsTotalKey)).Result()
	if err == nil {
		var total struct {
			TotalRequests int64 `json:"total_requests"`
			SuccessCount  int64 `json:"success_count"`
			FailureCount  int64 `json:"failure_count"`
			TotalTokens   int64 `json:"total_tokens"`
		}
		if json.Unmarshal([]byte(totalData), &total) == nil {
			snapshot.TotalRequests = total.TotalRequests
			snapshot.SuccessCount = total.SuccessCount
			snapshot.FailureCount = total.FailureCount
			snapshot.TotalTokens = total.TotalTokens
		}
	}

	// Load APIs stats
	apisData, err := client.Get(ctx, s.key(statsAPIsKey)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(apisData), &snapshot.APIs); err == nil {
			// Ensure maps are initialized
			if snapshot.APIs == nil {
				snapshot.APIs = make(map[string]APISnapshot)
			}
		}
	}

	// Load requests by day
	requestsByDayData, err := client.Get(ctx, s.key(statsRequestsByDay)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(requestsByDayData), &snapshot.RequestsByDay); err == nil {
			if snapshot.RequestsByDay == nil {
				snapshot.RequestsByDay = make(map[string]int64)
			}
		}
	}

	// Load requests by hour
	requestsByHourData, err := client.Get(ctx, s.key(statsRequestsByHour)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(requestsByHourData), &snapshot.RequestsByHour); err == nil {
			if snapshot.RequestsByHour == nil {
				snapshot.RequestsByHour = make(map[string]int64)
			}
		}
	}

	// Load tokens by day
	tokensByDayData, err := client.Get(ctx, s.key(statsTokensByDay)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(tokensByDayData), &snapshot.TokensByDay); err == nil {
			if snapshot.TokensByDay == nil {
				snapshot.TokensByDay = make(map[string]int64)
			}
		}
	}

	// Load tokens by hour
	tokensByHourData, err := client.Get(ctx, s.key(statsTokensByHour)).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(tokensByHourData), &snapshot.TokensByHour); err == nil {
			if snapshot.TokensByHour == nil {
				snapshot.TokensByHour = make(map[string]int64)
			}
		}
	}

	return snapshot
}

func (s *redisStatsStorage) MergeSnapshot(snapshot StatisticsSnapshot) MergeResult {
	bgCtx := context.Background()
	// For Redis storage, we merge by loading current snapshot, merging, and saving
	current := s.Snapshot()
	result := s.mergeSnapshots(&current, snapshot)
	s.saveSnapshot(bgCtx, current)
	return result
}

func (s *redisStatsStorage) mergeSnapshots(target *StatisticsSnapshot, source StatisticsSnapshot) MergeResult {
	result := MergeResult{}

	seen := make(map[string]struct{})
	if target.APIs != nil {
		for apiName, stats := range target.APIs {
			if stats.Models == nil {
				continue
			}
			for modelName, modelStatsValue := range stats.Models {
				for _, detail := range modelStatsValue.Details {
					seen[dedupKey(apiName, modelName, detail)] = struct{}{}
				}
			}
		}
	}

	if target.APIs == nil {
		target.APIs = make(map[string]APISnapshot)
	}

	for apiName, apiSnapshot := range source.APIs {
		if apiName = normalizeAPIKey(apiName); apiName == "" {
			continue
		}
		stats, ok := target.APIs[apiName]
		if !ok {
			stats = APISnapshot{Models: make(map[string]ModelSnapshot)}
		}
		if stats.Models == nil {
			stats.Models = make(map[string]ModelSnapshot)
		}
		for modelName, modelSnapshot := range apiSnapshot.Models {
			if modelName = normalizeModelName(modelName); modelName == "" {
				modelName = "unknown"
			}
			for _, detail := range modelSnapshot.Details {
				detail.Tokens = normaliseTokenStats(detail.Tokens)
				if detail.Timestamp.IsZero() {
					detail.Timestamp = time.Now()
				}
				key := dedupKey(apiName, modelName, detail)
				if _, exists := seen[key]; exists {
					result.Skipped++
					continue
				}
				seen[key] = struct{}{}
				s.recordImported(target, apiName, modelName, &stats, detail)
				result.Added++
			}
		}
		target.APIs[apiName] = stats
	}

	return result
}

func (s *redisStatsStorage) recordImported(snapshot *StatisticsSnapshot, apiName, modelName string, stats *APISnapshot, detail RequestDetail) {
	totalTokens := detail.Tokens.TotalTokens
	if totalTokens < 0 {
		totalTokens = 0
	}

	snapshot.TotalRequests++
	if detail.Failed {
		snapshot.FailureCount++
	} else {
		snapshot.SuccessCount++
	}
	snapshot.TotalTokens += totalTokens

	stats.TotalRequests++
	stats.TotalTokens += totalTokens

	if stats.Models == nil {
		stats.Models = make(map[string]ModelSnapshot)
	}
	modelStatsValue, ok := stats.Models[modelName]
	if !ok {
		modelStatsValue = ModelSnapshot{}
	}
	modelStatsValue.TotalRequests++
	modelStatsValue.TotalTokens += totalTokens
	modelStatsValue.Details = append(modelStatsValue.Details, detail)
	stats.Models[modelName] = modelStatsValue

	dayKey := detail.Timestamp.Format("2006-01-02")
	hourKey := detail.Timestamp.Hour()

	if snapshot.RequestsByDay == nil {
		snapshot.RequestsByDay = make(map[string]int64)
	}
	snapshot.RequestsByDay[dayKey]++

	if snapshot.RequestsByHour == nil {
		snapshot.RequestsByHour = make(map[string]int64)
	}
	snapshot.RequestsByHour[formatHour(hourKey)]++

	if snapshot.TokensByDay == nil {
		snapshot.TokensByDay = make(map[string]int64)
	}
	snapshot.TokensByDay[dayKey] += totalTokens

	if snapshot.TokensByHour == nil {
		snapshot.TokensByHour = make(map[string]int64)
	}
	snapshot.TokensByHour[formatHour(hourKey)] += totalTokens
}

func (s *redisStatsStorage) saveSnapshot(ctx context.Context, snapshot StatisticsSnapshot) {
	client := cache.GetClient()
	if client == nil {
		return
	}

	ttl := time.Duration(s.config.TTL) * time.Second
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	// Save total stats
	totalData, _ := json.Marshal(map[string]int64{
		"total_requests": snapshot.TotalRequests,
		"success_count":  snapshot.SuccessCount,
		"failure_count":  snapshot.FailureCount,
		"total_tokens":   snapshot.TotalTokens,
	})

	err := client.Set(ctx, s.key(statsTotalKey), totalData, ttl).Err()
	if err != nil {
		log.Errorf("Redis saveSnapshot failed: %v", err)
		return
	}

	// Save APIs stats
	if snapshot.APIs != nil {
		apisData, _ := json.Marshal(snapshot.APIs)
		client.Set(ctx, s.key(statsAPIsKey), apisData, ttl)
	}

	// Save requests by day
	if snapshot.RequestsByDay != nil {
		requestsByDayData, _ := json.Marshal(snapshot.RequestsByDay)
		client.Set(ctx, s.key(statsRequestsByDay), requestsByDayData, ttl)
	}

	// Save requests by hour
	if snapshot.RequestsByHour != nil {
		requestsByHourData, _ := json.Marshal(snapshot.RequestsByHour)
		client.Set(ctx, s.key(statsRequestsByHour), requestsByHourData, ttl)
	}

	// Save tokens by day
	if snapshot.TokensByDay != nil {
		tokensByDayData, _ := json.Marshal(snapshot.TokensByDay)
		client.Set(ctx, s.key(statsTokensByDay), tokensByDayData, ttl)
	}

	// Save tokens by hour
	if snapshot.TokensByHour != nil {
		tokensByHourData, _ := json.Marshal(snapshot.TokensByHour)
		client.Set(ctx, s.key(statsTokensByHour), tokensByHourData, ttl)
	}
}

func normalizeRecordDetail(record coreusage.Record) TokenStats {
	tokens := TokenStats{
		InputTokens:     record.Detail.InputTokens,
		OutputTokens:    record.Detail.OutputTokens,
		ReasoningTokens: record.Detail.ReasoningTokens,
		CachedTokens:    record.Detail.CachedTokens,
		TotalTokens:     record.Detail.TotalTokens,
	}
	if tokens.TotalTokens == 0 {
		tokens.TotalTokens = tokens.InputTokens + tokens.OutputTokens + tokens.ReasoningTokens
	}
	if tokens.TotalTokens == 0 {
		tokens.TotalTokens = tokens.InputTokens + tokens.OutputTokens + tokens.ReasoningTokens + tokens.CachedTokens
	}
	return tokens
}

func normalizeAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	return apiKey
}

func normalizeModelName(modelName string) string {
	if modelName == "" {
		return ""
	}
	return modelName
}
