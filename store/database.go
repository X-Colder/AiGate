// Package store 数据库连接和初始化
package store

import (
	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// DefaultTenantID 默认租户 ID
const DefaultTenantID = "00000000-0000-0000-0000-000000000001"

// InitDB 初始化 SQLite 数据库并自动迁移表结构
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return err
	}

	// 自动迁移所有实体表
	if err := DB.AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Role{},
		&model.Gateway{},
		&model.GatewayPolicy{},
		&model.MetricRecord{},
	); err != nil {
		return err
	}

	logger.Infof("Database initialized: %s", dbPath)
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// SetDB 设置数据库实例（用于测试注入）
func SetDB(db *gorm.DB) {
	DB = db
}

// DefaultAdminRoleID 默认管理员角色 ID
const DefaultAdminRoleID = "00000000-0000-0000-0000-000000000010"

// DefaultUserRoleID 默认普通用户角色 ID
const DefaultUserRoleID = "00000000-0000-0000-0000-000000000011"

// InitDefaultTenant 创建默认租户、默认角色和管理员账户（首次启动时）
func InitDefaultTenant() error {
	// 初始化默认角色
	initDefaultRoles()

	var count int64
	DB.Model(&model.Tenant{}).Count(&count)
	if count > 0 {
		return nil
	}

	tenant := model.Tenant{
		ID:     DefaultTenantID,
		Name:   "Default",
		Status: 1,
	}
	if err := DB.Create(&tenant).Error; err != nil {
		return err
	}

	// 创建默认管理员: admin/admin123
	admin := model.User{
		ID:       "00000000-0000-0000-0000-000000000001",
		TenantID: DefaultTenantID,
		Username: "admin",
		Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9", // sha256("admin123")
		RoleID:   DefaultAdminRoleID,
		Role:     "admin",
		Status:   1,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	logger.Infof("Default tenant and admin user created")
	return nil
}

// initDefaultRoles 初始化系统内置角色
func initDefaultRoles() {
	roles := []model.Role{
		{ID: DefaultAdminRoleID, Name: "超级管理员", Description: "全部权限", TenantAccess: true, GatewayAccess: true, MonitorAccess: true, IsSystem: true},
		{ID: DefaultUserRoleID, Name: "普通用户", Description: "网关管理和监控", TenantAccess: false, GatewayAccess: true, MonitorAccess: true, IsSystem: true},
	}
	for _, r := range roles {
		var count int64
		DB.Model(&model.Role{}).Where("id = ?", r.ID).Count(&count)
		if count == 0 {
			DB.Create(&r)
		}
	}
}
