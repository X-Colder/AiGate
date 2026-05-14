// Package service 认证服务：用户注册、登录、租户管理
package service

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/auth"
)

// AuthService 认证业务服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Register 注册新用户（用户可独立存在，无需绑定租户）
func (s *AuthService) Register(req *model.RegisterRequest) (*model.LoginResponse, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	user := model.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Phone:    req.Phone,
		Password: hashPassword(req.Password),
		Role:     "user",
		Status:   1,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}

	token, err := auth.GenerateToken(user.ID, user.TenantID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token error: %w", err)
	}

	return &model.LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Username:    user.Username,
		TenantID:    user.TenantID,
		Role:        user.Role,
		Permissions: model.Permissions{TenantAccess: false, GatewayAccess: false, MonitorAccess: false, ModelAccess: false, APIAccess: true, TeamAccess: false},
	}, nil
}

// Login 用户登录（查询角色权限并返回）
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	var user model.User
	if err := s.db.Where("username = ? AND status = 1", req.Username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if user.Password != hashPassword(req.Password) {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 查询用户关联的角色权限
	permissions := model.Permissions{}
	if user.RoleID != "" {
		var role model.Role
		if err := s.db.Where("id = ?", user.RoleID).First(&role).Error; err == nil {
			permissions.TenantAccess = role.TenantAccess
			permissions.GatewayAccess = role.GatewayAccess
			permissions.MonitorAccess = role.MonitorAccess
			permissions.ModelAccess = role.ModelAccess
			permissions.APIAccess = role.APIAccess
			permissions.TeamAccess = role.TeamAccess
		}
	} else if user.Role == "admin" {
		permissions = model.Permissions{TenantAccess: true, GatewayAccess: true, MonitorAccess: true, ModelAccess: true, APIAccess: true, TeamAccess: true}
	} else {
		permissions = model.Permissions{TenantAccess: false, GatewayAccess: false, MonitorAccess: false, ModelAccess: false, APIAccess: true, TeamAccess: false}
	}

	token, err := auth.GenerateToken(user.ID, user.TenantID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token error: %w", err)
	}

	return &model.LoginResponse{
		Token:       token,
		UserID:      user.ID,
		Username:    user.Username,
		TenantID:    user.TenantID,
		Role:        user.Role,
		Permissions: permissions,
	}, nil
}

// hashPassword SHA256 哈希密码
func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", h)
}
