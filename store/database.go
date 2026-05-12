package store

import (
	"fmt"
	"time"

	"github.com/aigate/config"
	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

const DefaultTenantID = "00000000-0000-0000-0000-000000000001"
const DefaultAdminRoleID = "00000000-0000-0000-0000-000000000010"
const DefaultUserRoleID = "00000000-0000-0000-0000-000000000011"

func InitDB(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dsn := buildMySQLDSN(cfg)
		dialector = mysql.Open(dsn)
	case "sqlite":
		path := cfg.Path
		if path == "" {
			path = "aigate.db"
		}
		dialector = sqlite.Open(path)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB error: %w", err)
	}

	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	maxLifetime := cfg.ConnMaxLifetime
	if maxLifetime <= 0 {
		maxLifetime = 300
	}
	maxIdleTime := cfg.ConnMaxIdleTime
	if maxIdleTime <= 0 {
		maxIdleTime = 60
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)

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

	logger.Infof("Database initialized: driver=%s", cfg.Driver)
	return nil
}

func buildMySQLDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func GetDB() *gorm.DB {
	return DB
}

func SetDB(db *gorm.DB) {
	DB = db
}

func InitDefaultTenant() error {
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

	admin := model.User{
		ID:       "00000000-0000-0000-0000-000000000001",
		TenantID: DefaultTenantID,
		Username: "admin",
		Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
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
