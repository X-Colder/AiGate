package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// GatewayHandler 网关管理请求处理器
type GatewayHandler struct {
	gatewayService *service.GatewayService
}

// NewGatewayHandler 创建网关管理处理器
func NewGatewayHandler(gatewayService *service.GatewayService) *GatewayHandler {
	return &GatewayHandler{gatewayService: gatewayService}
}

// Create 处理 POST /api/v1/gateways
func (h *GatewayHandler) Create(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var req model.CreateGatewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	gw, err := h.gatewayService.Create(tenantID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gw)
}

// List 处理 GET /api/v1/gateways
// admin 查看所有网关；普通用户仅查看自己租户的网关
func (h *GatewayHandler) List(c *gin.Context) {
	role := c.GetString("role")
	tenantID := c.GetString("tenant_id")
	if role == "admin" {
		tenantID = "" // admin 不限制租户
	}

	gateways, err := h.gatewayService.List(tenantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gateways)
}

// GetByID 处理 GET /api/v1/gateways/:id
func (h *GatewayHandler) GetByID(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	gatewayID := c.Param("id")

	gw, err := h.gatewayService.GetByID(tenantID, gatewayID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, gw)
}

// Update 处理 PUT /api/v1/gateways/:id
func (h *GatewayHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	gatewayID := c.Param("id")

	var req model.UpdateGatewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	gw, err := h.gatewayService.Update(tenantID, gatewayID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gw)
}

// Delete 处理 DELETE /api/v1/gateways/:id
func (h *GatewayHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	gatewayID := c.Param("id")

	if err := h.gatewayService.Delete(tenantID, gatewayID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "gateway deleted"})
}

// UpdatePolicy 处理 PUT /api/v1/gateways/:id/policy
func (h *GatewayHandler) UpdatePolicy(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	gatewayID := c.Param("id")

	var req model.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	policy, err := h.gatewayService.UpdatePolicy(tenantID, gatewayID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, policy)
}
