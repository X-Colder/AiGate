package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

type APIKeyHandler struct {
	apikeyService *service.APIKeyService
}

func NewAPIKeyHandler(apikeyService *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apikeyService: apikeyService}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	key, err := h.apikeyService.Create(userID, tenantID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, key)
}

func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	keys, err := h.apikeyService.List(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, keys)
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	userID := c.GetString("user_id")
	keyID := c.Param("id")
	if err := h.apikeyService.Revoke(userID, keyID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "api key revoked"})
}
