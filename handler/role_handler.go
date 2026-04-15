// Package handler 角色管理请求处理器（仅管理员可用）
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色管理处理器
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List 处理 GET /api/v1/admin/roles
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.roleService.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, roles)
}

// Create 处理 POST /api/v1/admin/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	role, err := h.roleService.Create(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, role)
}

// GetByID 处理 GET /api/v1/admin/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	role, err := h.roleService.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, role)
}

// Update 处理 PUT /api/v1/admin/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	role, err := h.roleService.Update(id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, role)
}

// Delete 处理 DELETE /api/v1/admin/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.roleService.Delete(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "role deleted"})
}
