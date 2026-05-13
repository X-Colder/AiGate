package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
	"github.com/aigate/service"
)

type ModelHandler struct {
	modelService   *service.ModelService
	billingService *service.BillingService
}

func NewModelHandler(modelService *service.ModelService, billingService *service.BillingService) *ModelHandler {
	return &ModelHandler{modelService: modelService, billingService: billingService}
}

func (h *ModelHandler) Create(c *gin.Context) {
	var req model.CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	m, err := h.modelService.Create(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, m)
}

func (h *ModelHandler) List(c *gin.Context) {
	models, err := h.modelService.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, models)
}

func (h *ModelHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.modelService.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "model not found")
		return
	}
	response.Success(c, m)
}

func (h *ModelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	m, err := h.modelService.Update(id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, m)
}

func (h *ModelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.modelService.Delete(id); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "model deleted"})
}

func (h *ModelHandler) UpdateDoc(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateModelDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := h.modelService.UpdateDoc(id, req.DocContent); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "doc updated"})
}

func (h *ModelHandler) ListUserBalances(c *gin.Context) {
	balances, err := h.billingService.ListUserBalances()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, balances)
}

func (h *ModelHandler) Recharge(c *gin.Context) {
	var req model.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := h.billingService.Recharge(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "recharge success"})
}
