package service

import (
	"fmt"
	"time"

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
		ID:                  uuid.New().String(),
		Name:                req.Name,
		Provider:            req.Provider,
		ModelID:             req.ModelID,
		Description:         req.Description,
		GatewayID:           req.GatewayID,
		BillingMode:         req.BillingMode,
		InputPricePer1K:     req.InputPricePer1K,
		OutputPricePer1K:    req.OutputPricePer1K,
		RequestPrice:        req.RequestPrice,
		UpstreamInputPer1K:  req.UpstreamInputPer1K,
		UpstreamOutputPer1K: req.UpstreamOutputPer1K,
		AlertThreshold:      req.AlertThreshold,
		FreeQuota:           req.FreeQuota,
		MonthlyQuota:        req.MonthlyQuota,
		MaxContextLength:    req.MaxContextLength,
		DocContent:          req.DocContent,
		Status:              1,
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
	if err := s.db.Where("status = ? AND gateway_id != '' AND gateway_id IN (SELECT id FROM gateways WHERE status = 1)", 1).
		Order("sort_order ASC, created_at DESC").Find(&models).Error; err != nil {
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
	if req.GatewayID != nil {
		updates["gateway_id"] = *req.GatewayID
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
	if req.UpstreamInputPer1K != nil {
		updates["upstream_input_per_1k"] = *req.UpstreamInputPer1K
	}
	if req.UpstreamOutputPer1K != nil {
		updates["upstream_output_per_1k"] = *req.UpstreamOutputPer1K
	}
	if req.AlertThreshold != nil {
		updates["alert_threshold"] = *req.AlertThreshold
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

func (s *ModelService) RechargeModel(modelID string, amount float64, desc string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var mc model.ModelCatalog
		if err := tx.Where("id = ?", modelID).First(&mc).Error; err != nil {
			return fmt.Errorf("model not found")
		}
		mc.UpstreamBalance += amount
		mc.UpstreamTotalRecharge += amount
		if err := tx.Save(&mc).Error; err != nil {
			return err
		}
		if desc == "" {
			desc = "模型上游充值"
		}
		log := model.ModelRechargeLog{
			ModelCatalogID: modelID,
			Amount:         amount,
			Balance:        mc.UpstreamBalance,
			Description:    desc,
			CreatedAt:      time.Now(),
		}
		return tx.Create(&log).Error
	})
}

func (s *ModelService) GetFinanceSummary() ([]model.ModelFinanceSummary, error) {
	var models []model.ModelCatalog
	if err := s.db.Order("sort_order ASC, created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	var summaries []model.ModelFinanceSummary
	for _, mc := range models {
		var revenue float64
		var upstreamCost float64
		var userCount int64

		s.db.Model(&model.UsageRecord{}).Where("model_catalog_id = ?", mc.ID).
			Select("COALESCE(SUM(cost), 0)").Scan(&revenue)
		s.db.Model(&model.UsageRecord{}).Where("model_catalog_id = ?", mc.ID).
			Select("COALESCE(SUM(upstream_cost), 0)").Scan(&upstreamCost)
		s.db.Model(&model.UsageRecord{}).Where("model_catalog_id = ?", mc.ID).
			Select("COUNT(DISTINCT user_id)").Scan(&userCount)

		summaries = append(summaries, model.ModelFinanceSummary{
			ModelID:          mc.ID,
			ModelName:        mc.Name,
			Provider:         mc.Provider,
			UserCount:        userCount,
			TotalRevenue:     revenue,
			UpstreamBalance:  mc.UpstreamBalance,
			UpstreamRecharge: mc.UpstreamTotalRecharge,
			UpstreamCost:     upstreamCost,
			Profit:           revenue - upstreamCost,
			AlertThreshold:   mc.AlertThreshold,
			Status:           mc.Status,
		})
	}
	return summaries, nil
}

func (s *ModelService) GetRechargeHistory(modelID string) ([]model.ModelRechargeLog, error) {
	var logs []model.ModelRechargeLog
	if err := s.db.Where("model_catalog_id = ?", modelID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *ModelService) CalculateUpstreamCost(mc *model.ModelCatalog, inputTokens, outputTokens int64) float64 {
	return mc.UpstreamInputPer1K*float64(inputTokens)/1000 + mc.UpstreamOutputPer1K*float64(outputTokens)/1000
}

func (s *ModelService) DeductUpstreamBalance(modelID string, cost float64) error {
	return s.db.Model(&model.ModelCatalog{}).Where("id = ?", modelID).
		Updates(map[string]interface{}{
			"upstream_balance":    gorm.Expr("upstream_balance - ?", cost),
			"upstream_total_cost": gorm.Expr("upstream_total_cost + ?", cost),
		}).Error
}
