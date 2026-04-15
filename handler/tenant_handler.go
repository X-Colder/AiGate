// Package handler 租户管理请求处理器（仅管理员可用）
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// TenantHandler 租户管理处理器
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler 创建租户管理处理器
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// Create 处理 POST /api/v1/admin/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req model.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	tenant, err := h.tenantService.Create(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, tenant)
}

// List 处理 GET /api/v1/admin/tenants
func (h *TenantHandler) List(c *gin.Context) {
	tenants, err := h.tenantService.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, tenants)
}

// GetByID 处理 GET /api/v1/admin/tenants/:id
func (h *TenantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	tenant, err := h.tenantService.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, tenant)
}

// Update 处理 PUT /api/v1/admin/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	tenant, err := h.tenantService.Update(id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, tenant)
}

// Delete 处理 DELETE /api/v1/admin/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.tenantService.Delete(id); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "tenant deleted"})
}

// GetUsage 处理 GET /api/v1/admin/tenants/:id/usage
func (h *TenantHandler) GetUsage(c *gin.Context) {
	id := c.Param("id")
	usage, err := h.tenantService.GetUsage(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, usage)
}
