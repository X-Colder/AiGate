package store

import (
	"fmt"
	"time"

	"github.com/aigate/config"
	"github.com/aigate/model"
	"github.com/aigate/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

const DefaultAdminRoleID = "00000000-0000-0000-0000-000000000010"
const DefaultUserRoleID = "00000000-0000-0000-0000-000000000011"
const DefaultTeamAdminRoleID = "00000000-0000-0000-0000-000000000012"
const DefaultTeamMemberRoleID = "00000000-0000-0000-0000-000000000013"

func InitDB(cfg *config.DatabaseConfig) error {
	dsn := buildMySQLDSN(cfg)
	dialector := mysql.Open(dsn)

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
		&model.ModelCatalog{},
		&model.APIKey{},
		&model.UserBalance{},
		&model.BalanceTransaction{},
		&model.UsageRecord{},
		&model.TeamInvitation{},
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

func InitDefaultData() error {
	initDefaultRoles()

	var count int64
	DB.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return nil
	}

	admin := model.User{
		ID:       "00000000-0000-0000-0000-000000000001",
		Username: "admin",
		Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
		RoleID:   DefaultAdminRoleID,
		Role:     "admin",
		Status:   1,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	logger.Infof("Default admin user created")
	return nil
}

func initDefaultRoles() {
	roles := []model.Role{
		{ID: DefaultAdminRoleID, Name: "系统管理员", Description: "系统全部权限", TenantAccess: true, GatewayAccess: true, MonitorAccess: true, ModelAccess: true, APIAccess: true, TeamAccess: true, IsSystem: true},
		{ID: DefaultUserRoleID, Name: "普通用户", Description: "C端用户，API调用", TenantAccess: false, GatewayAccess: false, MonitorAccess: false, ModelAccess: false, APIAccess: true, TeamAccess: false, IsSystem: true},
		{ID: DefaultTeamAdminRoleID, Name: "团队管理员", Description: "团队管理+API调用", TenantAccess: false, GatewayAccess: false, MonitorAccess: false, ModelAccess: false, APIAccess: true, TeamAccess: true, IsSystem: true},
		{ID: DefaultTeamMemberRoleID, Name: "团队成员", Description: "团队成员，共享配额", TenantAccess: false, GatewayAccess: false, MonitorAccess: false, ModelAccess: false, APIAccess: true, TeamAccess: false, IsSystem: true},
	}
	for _, r := range roles {
		var existing model.Role
		if DB.Where("id = ?", r.ID).First(&existing).Error != nil {
			DB.Create(&r)
		} else {
			DB.Model(&existing).Updates(map[string]interface{}{
				"name":           r.Name,
				"description":    r.Description,
				"tenant_access":  r.TenantAccess,
				"gateway_access": r.GatewayAccess,
				"monitor_access": r.MonitorAccess,
				"model_access":   r.ModelAccess,
				"api_access":     r.APIAccess,
				"team_access":    r.TeamAccess,
			})
		}
	}
}
