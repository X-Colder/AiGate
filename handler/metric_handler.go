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
// admin 可查看所有租户数据并按 tenant_id 参数筛选；普通用户仅查看自己租户
func (h *MetricHandler) GetSummary(c *gin.Context) {
	role := c.GetString("role")
	tenantID := c.GetString("tenant_id")
	if role == "admin" {
		tenantID = "" // admin 不限制租户，由 query.TenantID 筛选
	}

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
// admin 可查看所有租户趋势并按 tenant_id 参数筛选；普通用户仅查看自己租户
func (h *MetricHandler) GetTrend(c *gin.Context) {
	role := c.GetString("role")
	tenantID := c.GetString("tenant_id")
	if role == "admin" {
		tenantID = ""
	}

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
