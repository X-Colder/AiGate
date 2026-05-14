package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/config"
	"github.com/aigate/model"
	"github.com/aigate/provider"
	"github.com/aigate/service"
)

type InferenceHandler struct {
	modelService   *service.ModelService
	billingService *service.BillingService
	usageService   *service.UsageService
	apikeyService  *service.APIKeyService
	registry       *provider.Registry
	db             *gorm.DB
}

func NewInferenceHandler(
	modelService *service.ModelService,
	billingService *service.BillingService,
	usageService *service.UsageService,
	apikeyService *service.APIKeyService,
	registry *provider.Registry,
	db *gorm.DB,
) *InferenceHandler {
	return &InferenceHandler{
		modelService:   modelService,
		billingService: billingService,
		usageService:   usageService,
		apikeyService:  apikeyService,
		registry:       registry,
		db:             db,
	}
}

func (h *InferenceHandler) ChatCompletion(c *gin.Context) {
	var req model.OpenAICompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]interface{}{
			"message": "Invalid request: " + err.Error(),
			"type":    "invalid_request_error",
		}})
		return
	}

	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	apiKeyID := c.GetString("api_key_id")

	// Find model in catalog
	mc, err := h.modelService.GetByName(req.Model)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": map[string]interface{}{
			"message": fmt.Sprintf("model '%s' not found", req.Model),
			"type":    "invalid_request_error",
		}})
		return
	}

	// Check API key model permissions
	apiKeyVal, _ := c.Get("api_key")
	if apiKey, ok := apiKeyVal.(*model.APIKey); ok && apiKey.Models != "" {
		allowed := false
		for _, m := range splitModels(apiKey.Models) {
			if m == mc.ID || m == mc.ModelID || m == mc.Name {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": map[string]interface{}{
				"message": "this api key does not have access to the requested model",
				"type":    "invalid_request_error",
			}})
			return
		}
	}

	// Check balance
	if err := h.billingService.CheckBalance(userID, mc); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "insufficient_quota",
		}})
		return
	}

	// Get provider: prefer gateway-based lookup, fallback to registry
	var p provider.Provider
	if mc.GatewayID != "" {
		var gw model.Gateway
		if err := h.db.Where("id = ? AND status = 1", mc.GatewayID).First(&gw).Error; err == nil {
			// Create an ad-hoc provider from gateway config
			timeout := gw.Timeout
			if timeout <= 0 {
				timeout = 60
			}
			gwProvider := provider.NewOpenAICompatibleProvider(gw.Provider, config.ProviderConfig{
				Enabled: true,
				APIKey:  gw.APIKey,
				BaseURL: gw.BaseURL,
				Timeout: timeout,
			})
			p = gwProvider
		}
	}
	if p == nil {
		var regErr error
		p, regErr = h.registry.Get(mc.Provider)
		if regErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": map[string]interface{}{
				"message": fmt.Sprintf("provider '%s' not available", mc.Provider),
				"type":    "server_error",
			}})
			return
		}
	}

	// Build chat request for provider
	chatReq := &model.ChatRequest{
		Provider: mc.Provider,
		Model:    mc.ModelID,
		Messages: req.Messages,
		Stream:   req.Stream,
	}

	startTime := time.Now()

	// Call provider
	chatResp, err := p.Chat(context.Background(), chatReq)
	latencyMs := time.Since(startTime).Milliseconds()

	statusCode := 200
	var inputTokens, outputTokens, totalTokens int64

	if err != nil {
		statusCode = 500
		// Record failed usage
		h.usageService.Record(&model.UsageRecord{
			UserID:         userID,
			TenantID:       tenantID,
			APIKeyID:       apiKeyID,
			ModelCatalogID: mc.ID,
			ModelName:      mc.Name,
			Provider:       mc.Provider,
			LatencyMs:      latencyMs,
			StatusCode:     statusCode,
			CreatedAt:      time.Now(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "server_error",
		}})
		return
	}

	// Extract token usage
	if chatResp.Usage != nil {
		inputTokens = int64(chatResp.Usage.PromptTokens)
		outputTokens = int64(chatResp.Usage.CompletionTokens)
		totalTokens = int64(chatResp.Usage.TotalTokens)
	}

	// Calculate cost and deduct
	cost := h.billingService.CalculateCost(mc, inputTokens, outputTokens)
	if cost > 0 {
		desc := fmt.Sprintf("调用 %s (%d tokens)", mc.Name, totalTokens)
		h.billingService.Deduct(userID, tenantID, "", cost, desc)
	}

	// Update monthly usage
	if totalTokens > 0 {
		h.billingService.UpdateMonthlyUsage(userID, totalTokens)
	}

	// Record usage
	h.usageService.Record(&model.UsageRecord{
		UserID:         userID,
		TenantID:       tenantID,
		APIKeyID:       apiKeyID,
		ModelCatalogID: mc.ID,
		ModelName:      mc.Name,
		Provider:       mc.Provider,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		TotalTokens:    totalTokens,
		Cost:           cost,
		LatencyMs:      latencyMs,
		StatusCode:     statusCode,
		CreatedAt:      time.Now(),
	})

	// Build OpenAI-compatible response
	resp := model.OpenAICompletionResponse{
		ID:      "chatcmpl-" + uuid.New().String()[:8],
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []model.OpenAIChoice{
			{
				Index:        0,
				Message:      model.Message{Role: "assistant", Content: chatResp.Content},
				FinishReason: "stop",
			},
		},
		Usage: model.OpenAIUsage{
			PromptTokens:     int(inputTokens),
			CompletionTokens: int(outputTokens),
			TotalTokens:      int(totalTokens),
		},
	}

	c.JSON(http.StatusOK, resp)
}

func (h *InferenceHandler) ListModels(c *gin.Context) {
	models, err := h.modelService.ListEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": map[string]interface{}{
			"message": err.Error(),
			"type":    "server_error",
		}})
		return
	}
	items := make([]model.OpenAIModelItem, len(models))
	for i, m := range models {
		items[i] = model.OpenAIModelItem{
			ID:      m.ModelID,
			Object:  "model",
			OwnedBy: m.Provider,
		}
	}
	c.JSON(http.StatusOK, model.OpenAIModelList{
		Object: "list",
		Data:   items,
	})
}

func splitModels(s string) []string {
	var result []string
	for _, m := range splitByComma(s) {
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}

func splitByComma(s string) []string {
	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
