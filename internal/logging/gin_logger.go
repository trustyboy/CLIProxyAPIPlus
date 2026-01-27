// Package logging provides Gin middleware for HTTP request logging and panic recovery.
// It integrates Gin web framework with logrus for structured logging of HTTP requests,
// responses, and error handling with panic recovery capabilities.
package logging

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// aiAPIPrefixes defines path prefixes for AI API requests that should have request ID tracking.
var aiAPIPrefixes = []string{
	"/v1/chat/completions",
	"/v1/completions",
	"/v1/messages",
	"/v1/responses",
	"/v1beta/models/",
	"/api/provider/",
}

const skipGinLogKey = "__gin_skip_request_logging__"

// GinLogrusLogger returns a Gin middleware handler that logs HTTP requests and responses
// using logrus. It captures request details including method, path, status code, latency,
// client IP, and any error messages. Request ID is only added for AI API requests.
//
// Output format (AI API): [2025-12-23 20:14:10] [info ] | a1b2c3d4 | 200 |       23.559s | ...
// Output format (others): [2025-12-23 20:14:10] [info ] | -------- | 200 |       23.559s | ...
//
// Returns:
//   - gin.HandlerFunc: A middleware handler for request logging
func GinLogrusLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := util.MaskSensitiveQuery(c.Request.URL.RawQuery)

		// Only generate request ID for AI API paths
		var requestID string
		if isAIAPIPath(path) {
			requestID = GenerateRequestID()
			SetGinRequestID(c, requestID)
			ctx := WithRequestID(c.Request.Context(), requestID)
			c.Request = c.Request.WithContext(ctx)
		}

		// Extract model name before processing the request
		model := extractModelFromRequest(c)
		// Get provider name from model
		providers := util.GetProviderName(model)
		provider := "unknown"

		// Try to get more detailed provider/channel information
		if len(providers) > 0 {
			provider = providers[0]
		}

		c.Next()

		if shouldSkipGinRequestLogging(c) {
			return
		}

		if raw != "" {
			path = path + "?" + raw
		}

		latency := time.Since(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		} else {
			latency = latency.Truncate(time.Millisecond)
		}

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Get account info from gin.Context if available
		var accountInfo string
		if accountVal, exists := c.Get("cliproxy.account_info"); exists {
			if accountStr, ok := accountVal.(string); ok {
				accountInfo = accountStr
			}
		}

		if requestID == "" {
			requestID = "--------"
		}

		logLine := fmt.Sprintf("%s | %d | %13v | %s | %s | %s",
			requestID,
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)

		logEntry := log.WithFields(log.Fields{
			"request_id": requestID,
			"status":     statusCode,
			"latency":    latency,
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"provider":   provider,
			"model":      model,
		})

		// Add account info to log if available
		if accountInfo != "" {
			logEntry = logEntry.WithField("account", accountInfo)
			// Also add to log line for better readability
			logLine += " | account=" + accountInfo
		}

		if errorMessage != "" {
			logEntry = logEntry.WithField("error", errorMessage)
			logLine += " | error=" + errorMessage
		}

		if statusCode >= http.StatusInternalServerError {
			logEntry.Error(logLine)
		} else if statusCode >= http.StatusBadRequest {
			logEntry.Warn(logLine)
		} else {
			logEntry.Info(logLine)
		}
	}
}

// isAIAPIPath checks if the given path is an AI API endpoint that should have request ID tracking.
func isAIAPIPath(path string) bool {
	for _, prefix := range aiAPIPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// GinLogrusRecovery returns a Gin middleware handler that recovers from panics and logs
// them using logrus. When a panic occurs, it captures the panic value, stack trace,
// and request path, then returns a 500 Internal Server Error response to the client.
//
// Returns:
//   - gin.HandlerFunc: A middleware handler for panic recovery
func GinLogrusRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
			// Let net/http handle ErrAbortHandler so the connection is aborted without noisy stack logs.
			panic(http.ErrAbortHandler)
		}

		log.WithFields(log.Fields{
			"panic": recovered,
			"stack": string(debug.Stack()),
			"path":  c.Request.URL.Path,
		}).Error("recovered from panic")

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// SkipGinRequestLogging marks the provided Gin context so that GinLogrusLogger
// will skip emitting a log line for the associated request.
func SkipGinRequestLogging(c *gin.Context) {
	if c == nil {
		return
	}
	c.Set(skipGinLogKey, true)
}

// shouldSkipGinRequestLogging checks if the provided Gin context is marked to skip logging.
func shouldSkipGinRequestLogging(c *gin.Context) bool {
	if c == nil {
		return false
	}
	val, exists := c.Get(skipGinLogKey)
	if !exists {
		return false
	}
	flag, ok := val.(bool)
	return ok && flag
}

// extractModelFromRequest attempts to extract the model name from various request formats
func extractModelFromRequest(c *gin.Context) string {
	// First try to parse from JSON body (OpenAI, Claude, etc.)
	// Check common model field names
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		// Reset the body so it can be read again by subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	if result := gjson.GetBytes(body, "model"); result.Exists() && result.Type == gjson.String {
		return result.String()
	}

	// For Gemini requests, model is in the URL path
	// Standard format: /models/{model}:generateContent -> :action parameter
	if action := c.Param("action"); action != "" {
		// Split by colon to get model name (e.g., "gemini-pro:generateContent" -> "gemini-pro")
		parts := strings.Split(action, ":")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// AMP CLI format: /publishers/google/models/{model}:method -> *path parameter
	// Example: /publishers/google/models/gemini-3-pro-preview:streamGenerateContent
	if path := c.Param("path"); path != "" {
		// Look for /models/{model}:method pattern
		if idx := strings.Index(path, "/models/"); idx >= 0 {
			modelPart := path[idx+8:] // Skip "/models/"
			// Split by colon to get model name
			if colonIdx := strings.Index(modelPart, ":"); colonIdx > 0 {
				return modelPart[:colonIdx]
			}
		}
	}

	return ""
}
