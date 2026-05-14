// Package service 租户管理服务（仅管理员可用）
package service

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/store"
)

// TenantService 租户管理业务服务
type TenantService struct {
	db *gorm.DB
}

// NewTenantService 创建租户管理服务
func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{db: db}
}

// Create 创建新租户（同时创建该租户的管理员用户）
func (s *TenantService) Create(req *model.CreateTenantRequest) (*model.Tenant, error) {
	var count int64
	s.db.Model(&model.Tenant{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("tenant name already exists")
	}

	adminID := uuid.New().String()
	tenant := model.Tenant{
		ID:      uuid.New().String(),
		Name:    req.Name,
		OwnerID: adminID,
		Email:   req.Email,
		Phone:   req.Phone,
		Status:  1,
	}

	// 事务：创建租户 + 创建该租户的管理员用户
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return fmt.Errorf("create tenant error: %w", err)
		}
		admin := model.User{
			ID:       adminID,
			TenantID: tenant.ID,
			Username: req.AdminUser,
			Password: fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password))),
			RoleID:   store.DefaultTeamAdminRoleID,
			Role:     "user",
			Status:   1,
		}
		if err := tx.Create(&admin).Error; err != nil {
			return fmt.Errorf("create tenant admin user error: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// List 获取所有租户列表（含统计数据）
func (s *TenantService) List() ([]model.TenantDetail, error) {
	var tenants []model.Tenant
	if err := s.db.Order("created_at ASC").Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("list tenants error: %w", err)
	}

	if len(tenants) == 0 {
		return []model.TenantDetail{}, nil
	}

	// 批量查询各租户的用户数和网关数
	tenantIDs := make([]string, len(tenants))
	for i, t := range tenants {
		tenantIDs[i] = t.ID
	}

	type countResult struct {
		TenantID string
		Count    int64
	}

	var userCounts []countResult
	s.db.Model(&model.User{}).Select("tenant_id, COUNT(*) as count").
		Where("tenant_id IN ?", tenantIDs).Group("tenant_id").Scan(&userCounts)

	// 网关现在是全局资源，统计全局网关数
	var totalGWCount int64
	s.db.Model(&model.Gateway{}).Count(&totalGWCount)

	userMap := make(map[string]int64)
	for _, uc := range userCounts {
		userMap[uc.TenantID] = uc.Count
	}

	details := make([]model.TenantDetail, len(tenants))
	for i, t := range tenants {
		details[i] = model.TenantDetail{
			Tenant:       t,
			UserCount:    userMap[t.ID],
			GatewayCount: totalGWCount,
		}
	}
	return details, nil
}

// GetByID 获取单个租户
func (s *TenantService) GetByID(id string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := s.db.Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, fmt.Errorf("tenant not found")
	}
	return &tenant, nil
}

// Update 更新租户
func (s *TenantService) Update(id string, req *model.UpdateTenantRequest) (*model.Tenant, error) {
	tenant, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) > 0 {
		if err := s.db.Model(tenant).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update tenant error: %w", err)
		}
	}
	return s.GetByID(id)
}

// Delete 删除租户（同时删除关联的用户）
func (s *TenantService) Delete(id string) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("tenant_id = ?", id).Delete(&model.User{})
		return tx.Where("id = ?", id).Delete(&model.Tenant{}).Error
	})
}

// GetUsage 获取租户使用详情（管理员查看）
func (s *TenantService) GetUsage(tenantID string) (*model.TenantUsage, error) {
	tenant, err := s.GetByID(tenantID)
	if err != nil {
		return nil, err
	}

	var gateways []model.Gateway
	s.db.Where("status = 1").Find(&gateways)

	usage := &model.TenantUsage{TenantID: tenant.ID, TenantName: tenant.Name}

	if len(gateways) == 0 {
		return usage, nil
	}

	// 批量聚合所有网关的 metrics
	gwIDs := make([]string, len(gateways))
	for i, gw := range gateways {
		gwIDs[i] = gw.ID
	}

	type metricAgg struct {
		GatewayID string
		Tokens    int64
		Reqs      int64
		Errs      int64
		Latency   float64
	}
	var aggs []metricAgg
	s.db.Model(&model.MetricRecord{}).
		Select("gateway_id, COALESCE(SUM(tokens_used),0) as tokens, COALESCE(SUM(request_count),0) as reqs, COALESCE(SUM(error_count),0) as errs, COALESCE(AVG(avg_latency_ms),0) as latency").
		Where("gateway_id IN ?", gwIDs).
		Group("gateway_id").
		Scan(&aggs)

	aggMap := make(map[string]metricAgg)
	for _, a := range aggs {
		aggMap[a.GatewayID] = a
	}

	for _, gw := range gateways {
		agg := aggMap[gw.ID]
		errRate := 0.0
		if agg.Reqs > 0 {
			errRate = float64(agg.Errs) / float64(agg.Reqs)
		}
		usage.Gateways = append(usage.Gateways, model.GatewayUsage{
			GatewayID: gw.ID, GatewayName: gw.Name, Provider: gw.Provider, Status: gw.Status,
			TokensUsed: agg.Tokens, RequestCount: agg.Reqs, AvgLatencyMs: agg.Latency, ErrorRate: errRate,
		})
		usage.TotalTokens += agg.Tokens
		usage.TotalReqs += agg.Reqs
	}
	return usage, nil
}
