package management

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/registry"
)

// UnavailableModelInfo 表示一个不可用的模型信息
type UnavailableModelInfo struct {
	ModelID    string    `json:"model_id"`
	ModelName  string    `json:"model_name"`
	Provider   string    `json:"provider"`
	ClientID   string    `json:"client_id"`
	Reason     string    `json:"reason"`      // "quota_exceeded", "suspended", "cooldown"
	ReasonText string    `json:"reason_text"` // 详细原因描述
	Since      time.Time `json:"since"`       // 不可用开始时间
}

// GetUnavailableModels 返回当前所有不可用的模型列表
// GET /v0/management/model-availability
func (h *Handler) GetUnavailableModels(c *gin.Context) {
	reg := registry.GetGlobalRegistry()
	unavailableModels := make([]UnavailableModelInfo, 0)

	// 获取所有不可用模型信息
	models := reg.GetAllModels()
	now := time.Now()
	quotaExpiredDuration := 5 * time.Minute

	for modelID, registration := range models {
		if registration == nil {
			continue
		}

		// 检查配额超限的客户端
		if registration.QuotaExceededClients != nil {
			for clientID, quotaTime := range registration.QuotaExceededClients {
				if quotaTime != nil && now.Sub(*quotaTime) < quotaExpiredDuration {
					info := UnavailableModelInfo{
						ModelID:    modelID,
						ModelName:  registration.Info.DisplayName,
						Provider:   reg.GetClientProvider(clientID),
						ClientID:   clientID,
						Reason:     "cooldown",
						ReasonText: "配额超限冷却中",
						Since:      *quotaTime,
					}
					if info.ModelName == "" && registration.Info != nil {
						info.ModelName = registration.Info.ID
					}
					unavailableModels = append(unavailableModels, info)
				}
			}
		}

		// 检查被暂停的客户端
		if registration.SuspendedClients != nil {
			for clientID, reason := range registration.SuspendedClients {
				suspendedReason := "suspended"
				reasonText := "已暂停"
				if reason != "" {
					reasonText = reason
					if reason == "quota" {
						suspendedReason = "quota_exceeded"
						reasonText = "配额超限暂停"
					}
				}

				info := UnavailableModelInfo{
					ModelID:    modelID,
					ModelName:  registration.Info.DisplayName,
					Provider:   reg.GetClientProvider(clientID),
					ClientID:   clientID,
					Reason:     suspendedReason,
					ReasonText: reasonText,
					Since:      registration.LastUpdated,
				}
				if info.ModelName == "" && registration.Info != nil {
					info.ModelName = registration.Info.ID
				}
				unavailableModels = append(unavailableModels, info)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"models": unavailableModels,
		"count":  len(unavailableModels),
	})
}

// ResetModelAvailabilityRequest 重置模型可用性请求
type ResetModelAvailabilityRequest struct {
	ClientID string `json:"client_id" binding:"required"`
}

// ResetModelAvailability 重置指定模型的不可用状态
// POST /v0/management/model-availability/:model_id/reset
func (h *Handler) ResetModelAvailability(c *gin.Context) {
	modelID := c.Param("model_id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model_id is required"})
		return
	}

	var req ResetModelAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	reg := registry.GetGlobalRegistry()

	// 清除配额超限状态
	reg.ClearModelQuotaExceeded(req.ClientID, modelID)

	// 恢复暂停状态
	reg.ResumeClientModel(req.ClientID, modelID)

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "模型可用性已重置",
		"model_id":  modelID,
		"client_id": req.ClientID,
	})
}
