package service

import (
	"github.com/aigate/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModelService struct {
	db *gorm.DB
}

func NewModelService(db *gorm.DB) *ModelService {
	return &ModelService{db: db}
}

func (s *ModelService) Create(req *model.CreateModelRequest) (*model.ModelCatalog, error) {
	m := &model.ModelCatalog{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Provider:         req.Provider,
		ModelID:          req.ModelID,
		Description:      req.Description,
		BillingMode:      req.BillingMode,
		InputPricePer1K:  req.InputPricePer1K,
		OutputPricePer1K: req.OutputPricePer1K,
		RequestPrice:     req.RequestPrice,
		FreeQuota:        req.FreeQuota,
		MonthlyQuota:     req.MonthlyQuota,
		MaxContextLength: req.MaxContextLength,
		DocContent:       req.DocContent,
		Status:           1,
	}
	if m.BillingMode == "" {
		m.BillingMode = "prepaid"
	}
	if m.MaxContextLength == 0 {
		m.MaxContextLength = 4096
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ModelService) List() ([]model.ModelCatalog, error) {
	var models []model.ModelCatalog
	if err := s.db.Order("sort_order ASC, created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *ModelService) ListEnabled() ([]model.ModelCatalog, error) {
	var models []model.ModelCatalog
	if err := s.db.Where("status = ?", 1).Order("sort_order ASC, created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (s *ModelService) GetByID(id string) (*model.ModelCatalog, error) {
	var m model.ModelCatalog
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *ModelService) Update(id string, req *model.UpdateModelRequest) (*model.ModelCatalog, error) {
	var m model.ModelCatalog
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.ModelID != nil {
		updates["model_id"] = *req.ModelID
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.BillingMode != nil {
		updates["billing_mode"] = *req.BillingMode
	}
	if req.InputPricePer1K != nil {
		updates["input_price_per_1k"] = *req.InputPricePer1K
	}
	if req.OutputPricePer1K != nil {
		updates["output_price_per_1k"] = *req.OutputPricePer1K
	}
	if req.RequestPrice != nil {
		updates["request_price"] = *req.RequestPrice
	}
	if req.FreeQuota != nil {
		updates["free_quota"] = *req.FreeQuota
	}
	if req.MonthlyQuota != nil {
		updates["monthly_quota"] = *req.MonthlyQuota
	}
	if req.MaxContextLength != nil {
		updates["max_context_length"] = *req.MaxContextLength
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) > 0 {
		if err := s.db.Model(&m).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	s.db.Where("id = ?", id).First(&m)
	return &m, nil
}

func (s *ModelService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&model.ModelCatalog{}).Error
}

func (s *ModelService) UpdateDoc(id string, content string) error {
	return s.db.Model(&model.ModelCatalog{}).Where("id = ?", id).Update("doc_content", content).Error
}

func (s *ModelService) GetByModelID(provider, modelID string) (*model.ModelCatalog, error) {
	var m model.ModelCatalog
	if err := s.db.Where("provider = ? AND model_id = ? AND status = 1", provider, modelID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *ModelService) GetByName(name string) (*model.ModelCatalog, error) {
	var m model.ModelCatalog
	if err := s.db.Where("(name = ? OR model_id = ?) AND status = 1", name, name).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
