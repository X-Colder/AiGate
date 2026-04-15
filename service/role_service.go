// Package service 角色管理服务（RBAC）
package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/model"
)

// RoleService 角色管理业务服务
type RoleService struct {
	db *gorm.DB
}

// NewRoleService 创建角色管理服务
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// List 获取所有角色列表（含用户数统计）
func (s *RoleService) List() ([]model.RoleDetail, error) {
	var roles []model.Role
	if err := s.db.Order("is_system DESC, created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list roles error: %w", err)
	}

	details := make([]model.RoleDetail, len(roles))
	for i, r := range roles {
		var count int64
		s.db.Model(&model.User{}).Where("role_id = ?", r.ID).Count(&count)
		details[i] = model.RoleDetail{Role: r, UserCount: count}
	}
	return details, nil
}

// Create 创建新角色
func (s *RoleService) Create(req *model.CreateRoleRequest) (*model.Role, error) {
	var count int64
	s.db.Model(&model.Role{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("role name already exists")
	}

	role := model.Role{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Description:   req.Description,
		TenantAccess:  req.TenantAccess,
		GatewayAccess: req.GatewayAccess,
		MonitorAccess: req.MonitorAccess,
		IsSystem:      false,
	}
	if err := s.db.Create(&role).Error; err != nil {
		return nil, fmt.Errorf("create role error: %w", err)
	}
	return &role, nil
}

// GetByID 获取单个角色
func (s *RoleService) GetByID(id string) (*model.Role, error) {
	var role model.Role
	if err := s.db.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}
	return &role, nil
}

// Update 更新角色（系统角色不可修改权限字段名称）
func (s *RoleService) Update(id string, req *model.UpdateRoleRequest) (*model.Role, error) {
	role, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		// 检查名称是否重复
		var count int64
		s.db.Model(&model.Role{}).Where("name = ? AND id != ?", *req.Name, id).Count(&count)
		if count > 0 {
			return nil, fmt.Errorf("role name already exists")
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if !role.IsSystem {
		// 系统角色不可修改权限
		if req.TenantAccess != nil {
			updates["tenant_access"] = *req.TenantAccess
		}
		if req.GatewayAccess != nil {
			updates["gateway_access"] = *req.GatewayAccess
		}
		if req.MonitorAccess != nil {
			updates["monitor_access"] = *req.MonitorAccess
		}
	}

	if len(updates) > 0 {
		if err := s.db.Model(role).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update role error: %w", err)
		}
	}
	return s.GetByID(id)
}

// Delete 删除角色（系统角色不可删）
func (s *RoleService) Delete(id string) error {
	role, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return fmt.Errorf("system role cannot be deleted")
	}

	// 检查是否有用户使用该角色
	var count int64
	s.db.Model(&model.User{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("role is in use by %d users, cannot delete", count)
	}

	return s.db.Where("id = ?", id).Delete(&model.Role{}).Error
}
