package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// MetricHandler 监控指标请求处理器
type MetricHandler struct {
	metricService *service.MetricService
}

// NewMetricHandler 创建监控处理器
func NewMetricHandler(metricService *service.MetricService) *MetricHandler {
	return &MetricHandler{metricService: metricService}
}

// GetSummary 处理 GET /api/v1/metrics/summary
func (h *MetricHandler) GetSummary(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var query model.MetricQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}

	summary, err := h.metricService.GetSummary(tenantID, &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, summary)
}

// GetTrend 处理 GET /api/v1/metrics/trend
func (h *MetricHandler) GetTrend(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var query model.MetricQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query: "+err.Error())
		return
	}

	trend, err := h.metricService.GetTrend(tenantID, &query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, trend)
}
