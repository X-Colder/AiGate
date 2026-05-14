// Package service 网关管理服务：增删改查网关及策略
package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/model"
)

// GatewayService 网关管理业务服务
type GatewayService struct {
	db *gorm.DB
}

// NewGatewayService 创建网关管理服务
func NewGatewayService(db *gorm.DB) *GatewayService {
	return &GatewayService{db: db}
}

// Create 创建网关，同时创建默认策略
func (s *GatewayService) Create(req *model.CreateGatewayRequest) (*model.Gateway, error) {
	gw := model.Gateway{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Provider: req.Provider,
		BaseURL:  req.BaseURL,
		APIKey:   req.APIKey,
		Timeout:  req.Timeout,
		Status:   1,
	}
	if gw.Timeout == 0 {
		gw.Timeout = 60
	}

	// 使用事务同时创建网关和默认策略
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&gw).Error; err != nil {
			return err
		}
		policy := model.GatewayPolicy{
			ID:        uuid.New().String(),
			GatewayID: gw.ID,
		}
		return tx.Create(&policy).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create gateway error: %w", err)
	}

	// 重新查询带策略的完整数据
	return s.GetByID(gw.ID)
}

// List 获取所有网关列表（网关为全局资源）
func (s *GatewayService) List() ([]model.Gateway, error) {
	var gateways []model.Gateway
	db := s.db.Preload("Policy")
	if err := db.Find(&gateways).Error; err != nil {
		return nil, fmt.Errorf("list gateways error: %w", err)
	}
	return gateways, nil
}

// GetByID 获取单个网关
func (s *GatewayService) GetByID(gatewayID string) (*model.Gateway, error) {
	var gw model.Gateway
	if err := s.db.Where("id = ?", gatewayID).Preload("Policy").First(&gw).Error; err != nil {
		return nil, fmt.Errorf("gateway not found")
	}
	return &gw, nil
}

// Update 更新网关信息
func (s *GatewayService) Update(gatewayID string, req *model.UpdateGatewayRequest) (*model.Gateway, error) {
	gw, err := s.GetByID(gatewayID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Provider != "" {
		updates["provider"] = req.Provider
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Timeout > 0 {
		updates["timeout"] = req.Timeout
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.db.Model(gw).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update gateway error: %w", err)
		}
	}

	return s.GetByID(gatewayID)
}

// Delete 删除网关及其策略
func (s *GatewayService) Delete(gatewayID string) error {
	_, err := s.GetByID(gatewayID)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("gateway_id = ?", gatewayID).Delete(&model.GatewayPolicy{})
		tx.Where("gateway_id = ?", gatewayID).Delete(&model.MetricRecord{})
		return tx.Where("id = ?", gatewayID).Delete(&model.Gateway{}).Error
	})
}

// UpdatePolicy 更新网关策略
func (s *GatewayService) UpdatePolicy(gatewayID string, req *model.UpdatePolicyRequest) (*model.GatewayPolicy, error) {
	// 先验证网关存在
	if _, err := s.GetByID(gatewayID); err != nil {
		return nil, err
	}

	var policy model.GatewayPolicy
	if err := s.db.Where("gateway_id = ?", gatewayID).First(&policy).Error; err != nil {
		return nil, fmt.Errorf("policy not found")
	}

	updates := buildPolicyUpdates(req)
	if len(updates) > 0 {
		if err := s.db.Model(&policy).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update policy error: %w", err)
		}
	}

	s.db.Where("gateway_id = ?", gatewayID).First(&policy)
	return &policy, nil
}

// buildPolicyUpdates 构建策略更新字段映射
func buildPolicyUpdates(req *model.UpdatePolicyRequest) map[string]interface{} {
	u := map[string]interface{}{}
	if req.RateLimitEnabled != nil {
		u["rate_limit_enabled"] = *req.RateLimitEnabled
	}
	if req.RateLimitQPS != nil {
		u["rate_limit_qps"] = *req.RateLimitQPS
	}
	if req.RateLimitBurst != nil {
		u["rate_limit_burst"] = *req.RateLimitBurst
	}
	if req.CircuitBreakerEnabled != nil {
		u["circuit_breaker_enabled"] = *req.CircuitBreakerEnabled
	}
	if req.CircuitBreakerThreshold != nil {
		u["circuit_breaker_threshold"] = *req.CircuitBreakerThreshold
	}
	if req.CircuitBreakerTimeout != nil {
		u["circuit_breaker_timeout"] = *req.CircuitBreakerTimeout
	}
	if req.CircuitBreakerMinReqs != nil {
		u["circuit_breaker_min_reqs"] = *req.CircuitBreakerMinReqs
	}
	if req.FallbackEnabled != nil {
		u["fallback_enabled"] = *req.FallbackEnabled
	}
	if req.FallbackGatewayID != nil {
		u["fallback_gateway_id"] = *req.FallbackGatewayID
	}
	return u
}
