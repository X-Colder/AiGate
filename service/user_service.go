// Package service 用户管理服务（管理员管理所有用户）
package service

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/model"
)

// UserService 用户管理业务服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// List 获取所有用户列表（含角色名和租户名）
func (s *UserService) List() ([]model.UserDetail, error) {
	var users []model.User
	if err := s.db.Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users error: %w", err)
	}

	details := make([]model.UserDetail, len(users))
	for i, u := range users {
		detail := model.UserDetail{
			ID:        u.ID,
			TenantID:  u.TenantID,
			Username:  u.Username,
			RoleID:    u.RoleID,
			Role:      u.Role,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		// 查询租户名
		var tenant model.Tenant
		if s.db.Where("id = ?", u.TenantID).First(&tenant).Error == nil {
			detail.TenantName = tenant.Name
		}
		// 查询角色名
		var role model.Role
		if s.db.Where("id = ?", u.RoleID).First(&role).Error == nil {
			detail.RoleName = role.Name
		}
		details[i] = detail
	}
	return details, nil
}

// Create 创建新用户（管理员操作）
func (s *UserService) Create(req *model.CreateUserRequest) (*model.UserDetail, error) {
	// 检查用户名唯一
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查租户存在
	var tenant model.Tenant
	if err := s.db.Where("id = ? AND status = 1", req.TenantID).First(&tenant).Error; err != nil {
		return nil, fmt.Errorf("tenant not found or disabled")
	}

	// 检查角色存在
	var role model.Role
	if err := s.db.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found")
	}

	// 根据角色权限判断 role 字段（兼容旧逻辑）
	roleStr := "user"
	if role.TenantAccess {
		roleStr = "admin"
	}

	user := model.User{
		ID:       uuid.New().String(),
		TenantID: req.TenantID,
		Username: req.Username,
		Password: fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password))),
		RoleID:   req.RoleID,
		Role:     roleStr,
		Status:   1,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}

	return &model.UserDetail{
		ID: user.ID, TenantID: user.TenantID, TenantName: tenant.Name,
		Username: user.Username, RoleID: user.RoleID, RoleName: role.Name,
		Role: user.Role, Status: user.Status, CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Update 更新用户（密码、角色、状态）
func (s *UserService) Update(id string, req *model.UpdateUserRequest) (*model.UserDetail, error) {
	var user model.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	updates := map[string]interface{}{}
	if req.Password != nil && *req.Password != "" {
		updates["password"] = fmt.Sprintf("%x", sha256.Sum256([]byte(*req.Password)))
	}
	if req.RoleID != nil {
		var role model.Role
		if err := s.db.Where("id = ?", *req.RoleID).First(&role).Error; err != nil {
			return nil, fmt.Errorf("role not found")
		}
		updates["role_id"] = *req.RoleID
		if role.TenantAccess {
			updates["role"] = "admin"
		} else {
			updates["role"] = "user"
		}
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update user error: %w", err)
		}
	}

	// 重新查询返回
	return s.getUserDetail(id)
}

// Delete 删除用户
func (s *UserService) Delete(id string) error {
	var user model.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}
	return s.db.Where("id = ?", id).Delete(&model.User{}).Error
}

func (s *UserService) getUserDetail(id string) (*model.UserDetail, error) {
	var user model.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	detail := &model.UserDetail{
		ID: user.ID, TenantID: user.TenantID, Username: user.Username,
		RoleID: user.RoleID, Role: user.Role, Status: user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	var tenant model.Tenant
	if s.db.Where("id = ?", user.TenantID).First(&tenant).Error == nil {
		detail.TenantName = tenant.Name
	}
	var role model.Role
	if s.db.Where("id = ?", user.RoleID).First(&role).Error == nil {
		detail.RoleName = role.Name
	}
	return detail, nil
}
