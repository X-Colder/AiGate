package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

type DeveloperHandler struct {
	modelService   *service.ModelService
	billingService *service.BillingService
	usageService   *service.UsageService
}

func NewDeveloperHandler(modelService *service.ModelService, billingService *service.BillingService, usageService *service.UsageService) *DeveloperHandler {
	return &DeveloperHandler{modelService: modelService, billingService: billingService, usageService: usageService}
}

func (h *DeveloperHandler) ListModels(c *gin.Context) {
	models, err := h.modelService.ListEnabled()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, models)
}

func (h *DeveloperHandler) GetModelDoc(c *gin.Context) {
	id := c.Param("id")
	m, err := h.modelService.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "model not found")
		return
	}
	response.Success(c, gin.H{
		"id":          m.ID,
		"name":        m.Name,
		"provider":    m.Provider,
		"model_id":    m.ModelID,
		"description": m.Description,
		"doc_content": m.DocContent,
	})
}

func (h *DeveloperHandler) GetBalance(c *gin.Context) {
	userID := c.GetString("user_id")
	balance, err := h.billingService.GetBalance(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, balance)
}

func (h *DeveloperHandler) GetTransactions(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	txns, total, err := h.billingService.GetTransactions(userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"list": txns, "total": total, "page": page, "page_size": pageSize})
}

func (h *DeveloperHandler) GetUsageSummary(c *gin.Context) {
	userID := c.GetString("user_id")
	var query model.UsageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}
	summary, err := h.usageService.GetSummary(userID, &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, summary)
}

func (h *DeveloperHandler) GetUsageTrend(c *gin.Context) {
	userID := c.GetString("user_id")
	var query model.UsageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}
	trend, err := h.usageService.GetTrend(userID, &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, trend)
}

func (h *DeveloperHandler) GetUsageRecords(c *gin.Context) {
	userID := c.GetString("user_id")
	var query model.UsageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	records, total, err := h.usageService.GetRecords(userID, &query, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"list": records, "total": total, "page": page, "page_size": pageSize})
}

func (h *DeveloperHandler) PurchaseSubscription(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	var req model.PurchaseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	result, err := h.billingService.PurchaseSubscription(userID, tenantID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *DeveloperHandler) ListSubscriptions(c *gin.Context) {
	userID := c.GetString("user_id")
	subs, err := h.billingService.ListSubscriptions(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, subs)
}
