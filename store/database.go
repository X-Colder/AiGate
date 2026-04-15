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

// InitDefaultTenant 创建默认租户和管理员账户（首次启动时）
func InitDefaultTenant() error {
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
		Role:     "admin",
		Status:   1,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	logger.Infof("Default tenant and admin user created")
	return nil
}
