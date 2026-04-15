// Package model 定义 AiGate 网关的所有数据模型。
// entity.go 包含数据库实体（用户、租户、网关、网关策略、监控指标）。
package model

import (
	"time"
)

// Tenant 租户实体，多租户隔离的顶层单元
type Tenant struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Name      string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Email     string    `json:"email" gorm:"size:150"`
	Phone     string    `json:"phone" gorm:"size:30"`
	Status    int       `json:"status" gorm:"default:1"` // 1=启用 0=禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User 用户实体，归属于某个租户
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID  string    `json:"tenant_id" gorm:"size:36;not null;index"`
	Username  string    `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"size:128;not null"`       // json 忽略密码输出
	RoleID    string    `json:"role_id" gorm:"size:36"`           // 关联角色
	Role      string    `json:"role" gorm:"size:20;default:user"` // admin, user（兼容旧逻辑）
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role RBAC 角色实体，定义模块访问权限
type Role struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	Name          string    `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description   string    `json:"description" gorm:"size:200"`
	TenantAccess  bool      `json:"tenant_access" gorm:"default:false"`  // 租户管理权限
	GatewayAccess bool      `json:"gateway_access" gorm:"default:false"` // 网关管理权限
	MonitorAccess bool      `json:"monitor_access" gorm:"default:false"` // 监控面板权限
	IsSystem      bool      `json:"is_system" gorm:"default:false"`      // 系统内置角色不可删
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Gateway AI 网关实体，每个网关对应一个 AI 服务提供者
type Gateway struct {
	ID        string         `json:"id" gorm:"primaryKey;size:36"`
	TenantID  string         `json:"tenant_id" gorm:"size:36;not null;index"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Provider  string         `json:"provider" gorm:"size:50;not null"` // openai, deepseek, doubao, qwen, kimi, anthropic
	BaseURL   string         `json:"base_url" gorm:"size:255"`
	APIKey    string         `json:"api_key" gorm:"size:255"`
	Model     string         `json:"model" gorm:"size:100"`
	Timeout   int            `json:"timeout" gorm:"default:60"`
	Status    int            `json:"status" gorm:"default:1"` // 1=启用 0=禁用
	Policy    *GatewayPolicy `json:"policy,omitempty" gorm:"foreignKey:GatewayID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GatewayPolicy 网关策略（熔断/限流/降级）
type GatewayPolicy struct {
	ID        string `json:"id" gorm:"primaryKey;size:36"`
	GatewayID string `json:"gateway_id" gorm:"size:36;not null;uniqueIndex"`
	// 限流策略
	RateLimitEnabled bool `json:"rate_limit_enabled" gorm:"default:false"`
	RateLimitQPS     int  `json:"rate_limit_qps" gorm:"default:100"`   // 每秒最大请求数
	RateLimitBurst   int  `json:"rate_limit_burst" gorm:"default:200"` // 突发最大请求数
	// 熔断策略
	CircuitBreakerEnabled   bool    `json:"circuit_breaker_enabled" gorm:"default:false"`
	CircuitBreakerThreshold float64 `json:"circuit_breaker_threshold" gorm:"default:0.5"` // 错误率阈值 (0-1)
	CircuitBreakerTimeout   int     `json:"circuit_breaker_timeout" gorm:"default:30"`    // 熔断恢复时间（秒）
	CircuitBreakerMinReqs   int     `json:"circuit_breaker_min_reqs" gorm:"default:10"`   // 触发熔断的最小请求数
	// 降级策略
	FallbackEnabled  bool      `json:"fallback_enabled" gorm:"default:false"`
	FallbackProvider string    `json:"fallback_provider" gorm:"size:50"` // 降级到的备用 Provider
	FallbackModel    string    `json:"fallback_model" gorm:"size:100"`   // 降级使用的模型
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MetricRecord 监控指标记录（每分钟聚合一条）
type MetricRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID     string    `json:"tenant_id" gorm:"size:36;not null;index:idx_metric_query"`
	GatewayID    string    `json:"gateway_id" gorm:"size:36;not null;index:idx_metric_query"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null;index:idx_metric_query"`
	RequestCount int64     `json:"request_count"`  // 请求总数
	TokensUsed   int64     `json:"tokens_used"`    // Token 消耗总量
	AvgLatencyMs float64   `json:"avg_latency_ms"` // 平均响应时间（毫秒）
	ErrorCount   int64     `json:"error_count"`    // 错误请求数
	UniqueUsers  int64     `json:"unique_users"`   // 独立访问用户数
}

// TableName 自定义表名
func (Tenant) TableName() string        { return "tenants" }
func (User) TableName() string          { return "users" }
func (Role) TableName() string          { return "roles" }
func (Gateway) TableName() string       { return "gateways" }
func (GatewayPolicy) TableName() string { return "gateway_policies" }
func (MetricRecord) TableName() string  { return "metric_records" }
