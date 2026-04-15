// Package handler 用户管理请求处理器（仅管理员可用）
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 处理 GET /api/v1/admin/users
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userService.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, users)
}

// Create 处理 POST /api/v1/admin/users
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	user, err := h.userService.Create(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, user)
}

// GetByID 处理 GET /api/v1/admin/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	users, err := h.userService.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	for _, u := range users {
		if u.ID == id {
			response.Success(c, u)
			return
		}
	}
	response.Error(c, http.StatusNotFound, "user not found")
}

// Update 处理 PUT /api/v1/admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	user, err := h.userService.Update(id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, user)
}

// Delete 处理 DELETE /api/v1/admin/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.userService.Delete(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "user deleted"})
}
