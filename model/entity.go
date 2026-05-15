package model

import (
	"time"
)

// Tenant 团队实体（原租户概念），用户升级后创建
type Tenant struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	Name      string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	OwnerID   string    `json:"owner_id" gorm:"size:36;not null"`
	Email     string    `json:"email" gorm:"size:150"`
	Phone     string    `json:"phone" gorm:"size:30"`
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User 用户实体，可独立存在或归属团队
type User struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID  string    `json:"tenant_id" gorm:"size:36;index"`
	Username  string    `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Phone     string    `json:"phone" gorm:"size:30;index"`
	Password  string    `json:"-" gorm:"size:128;not null"`
	RoleID    string    `json:"role_id" gorm:"size:36"`
	Role      string    `json:"role" gorm:"size:20;default:user"`
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role RBAC 角色实体
type Role struct {
	ID            string    `json:"id" gorm:"primaryKey;size:36"`
	Name          string    `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Description   string    `json:"description" gorm:"size:200"`
	TenantAccess  bool      `json:"tenant_access" gorm:"default:false"`
	GatewayAccess bool      `json:"gateway_access" gorm:"default:false"`
	MonitorAccess bool      `json:"monitor_access" gorm:"default:false"`
	ModelAccess   bool      `json:"model_access" gorm:"default:false"`
	APIAccess     bool      `json:"api_access" gorm:"default:false"`
	TeamAccess    bool      `json:"team_access" gorm:"default:false"`
	IsSystem      bool      `json:"is_system" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Gateway 网关实体（全局级，代表一个上游 AI 服务连接）
type Gateway struct {
	ID        string         `json:"id" gorm:"primaryKey;size:36"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Provider  string         `json:"provider" gorm:"size:50;not null"`
	BaseURL   string         `json:"base_url" gorm:"size:255"`
	APIKey    string         `json:"api_key" gorm:"size:255"`
	Timeout   int            `json:"timeout" gorm:"default:60"`
	Status    int            `json:"status" gorm:"default:1"`
	Policy    *GatewayPolicy `json:"policy,omitempty" gorm:"foreignKey:GatewayID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GatewayPolicy 网关策略（限流/熔断/降级）
type GatewayPolicy struct {
	ID                      string    `json:"id" gorm:"primaryKey;size:36"`
	GatewayID               string    `json:"gateway_id" gorm:"size:36;not null;uniqueIndex"`
	RateLimitEnabled        bool      `json:"rate_limit_enabled" gorm:"default:false"`
	RateLimitQPS            int       `json:"rate_limit_qps" gorm:"default:100"`
	RateLimitBurst          int       `json:"rate_limit_burst" gorm:"default:200"`
	CircuitBreakerEnabled   bool      `json:"circuit_breaker_enabled" gorm:"default:false"`
	CircuitBreakerThreshold float64   `json:"circuit_breaker_threshold" gorm:"default:0.5"`
	CircuitBreakerTimeout   int       `json:"circuit_breaker_timeout" gorm:"default:30"`
	CircuitBreakerMinReqs   int       `json:"circuit_breaker_min_reqs" gorm:"default:10"`
	FallbackEnabled         bool      `json:"fallback_enabled" gorm:"default:false"`
	FallbackGatewayID       string    `json:"fallback_gateway_id" gorm:"size:36"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// MetricRecord 监控指标记录
type MetricRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID     string    `json:"tenant_id" gorm:"size:36;index:idx_metric_query"`
	GatewayID    string    `json:"gateway_id" gorm:"size:36;index:idx_metric_query"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null;index:idx_metric_query"`
	RequestCount int64     `json:"request_count"`
	TokensUsed   int64     `json:"tokens_used"`
	AvgLatencyMs float64   `json:"avg_latency_ms"`
	ErrorCount   int64     `json:"error_count"`
	UniqueUsers  int64     `json:"unique_users"`
}

// ModelCatalog 模型目录，管理员配置可用模型和定价
type ModelCatalog struct {
	ID                    string    `json:"id" gorm:"primaryKey;size:36"`
	Name                  string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Provider              string    `json:"provider" gorm:"size:50;not null;index"`
	ModelID               string    `json:"model_id" gorm:"size:100;not null"`
	Description           string    `json:"description" gorm:"size:500"`
	GatewayID             string    `json:"gateway_id" gorm:"size:36;index"`
	BillingMode           string    `json:"billing_mode" gorm:"size:50;default:prepaid"`
	InputPricePer1K       float64   `json:"input_price_per_1k" gorm:"default:0"`
	OutputPricePer1K      float64   `json:"output_price_per_1k" gorm:"default:0"`
	RequestPrice          float64   `json:"request_price" gorm:"default:0"`
	UpstreamInputPer1K    float64   `json:"upstream_input_per_1k" gorm:"default:0"`
	UpstreamOutputPer1K   float64   `json:"upstream_output_per_1k" gorm:"default:0"`
	UpstreamBalance       float64   `json:"upstream_balance" gorm:"default:0"`
	UpstreamTotalRecharge float64   `json:"upstream_total_recharge" gorm:"default:0"`
	UpstreamTotalCost     float64   `json:"upstream_total_cost" gorm:"default:0"`
	AlertThreshold        float64   `json:"alert_threshold" gorm:"default:10"`
	FreeQuota             int64     `json:"free_quota" gorm:"default:0"`
	MonthlyQuota          int64     `json:"monthly_quota" gorm:"default:0"`
	MaxContextLength      int       `json:"max_context_length" gorm:"default:4096"`
	Status                int       `json:"status" gorm:"default:1"`
	DocContent            string    `json:"doc_content,omitempty" gorm:"type:text"`
	SortOrder             int       `json:"sort_order" gorm:"default:0"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// APIKey 开发者 API 密钥
type APIKey struct {
	ID           string     `json:"id" gorm:"primaryKey;size:36"`
	UserID       string     `json:"user_id" gorm:"size:36;not null;index"`
	TenantID     string     `json:"tenant_id" gorm:"size:36;index"`
	Name         string     `json:"name" gorm:"size:100;not null"`
	KeyHash      string     `json:"-" gorm:"size:64;not null;uniqueIndex"`
	KeyPrefix    string     `json:"key_prefix" gorm:"size:16"`
	RateLimitQPM int        `json:"rate_limit_qpm" gorm:"default:60"`
	Models       string     `json:"models" gorm:"size:500;not null"`
	Status       int        `json:"status" gorm:"default:1"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserBalance 用户/团队钱包
type UserBalance struct {
	ID             string    `json:"id" gorm:"primaryKey;size:36"`
	UserID         string    `json:"user_id" gorm:"size:36;uniqueIndex"`
	TenantID       string    `json:"tenant_id" gorm:"size:36;index"`
	Balance        float64   `json:"balance" gorm:"default:0"`
	FreeBalance    float64   `json:"free_balance" gorm:"default:0"`
	TotalRecharged float64   `json:"total_recharged" gorm:"default:0"`
	TotalConsumed  float64   `json:"total_consumed" gorm:"default:0"`
	MonthlyUsed    int64     `json:"monthly_used" gorm:"default:0"`
	MonthlyResetAt time.Time `json:"monthly_reset_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BalanceTransaction 余额变动流水
type BalanceTransaction struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      string    `json:"user_id" gorm:"size:36;not null;index"`
	TenantID    string    `json:"tenant_id" gorm:"size:36;index"`
	Type        string    `json:"type" gorm:"size:20;not null"`
	Amount      float64   `json:"amount"`
	Balance     float64   `json:"balance"`
	RelatedID   string    `json:"related_id" gorm:"size:36"`
	Description string    `json:"description" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// UsageRecord C端调用记录
type UsageRecord struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         string    `json:"user_id" gorm:"size:36;not null;index"`
	TenantID       string    `json:"tenant_id" gorm:"size:36;index"`
	APIKeyID       string    `json:"api_key_id" gorm:"size:36;not null;index"`
	ModelCatalogID string    `json:"model_catalog_id" gorm:"size:36;index"`
	ModelName      string    `json:"model_name" gorm:"size:100"`
	Provider       string    `json:"provider" gorm:"size:50"`
	InputTokens    int64     `json:"input_tokens" gorm:"default:0"`
	OutputTokens   int64     `json:"output_tokens" gorm:"default:0"`
	TotalTokens    int64     `json:"total_tokens" gorm:"default:0"`
	Cost           float64   `json:"cost" gorm:"default:0"`
	UpstreamCost   float64   `json:"upstream_cost" gorm:"default:0"`
	LatencyMs      int64     `json:"latency_ms" gorm:"default:0"`
	StatusCode     int       `json:"status_code" gorm:"default:200"`
	CreatedAt      time.Time `json:"created_at" gorm:"index:idx_usage_time"`
}

// TeamInvitation 团队邀请
type TeamInvitation struct {
	ID        string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID  string    `json:"tenant_id" gorm:"size:36;not null;index"`
	InviterID string    `json:"inviter_id" gorm:"size:36;not null"`
	Phone     string    `json:"phone" gorm:"size:30;not null;index"`
	Status    int       `json:"status" gorm:"default:0"` // 0=待接受 1=已接受 2=已拒绝
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (Tenant) TableName() string              { return "tenants" }
func (User) TableName() string                { return "users" }
func (Role) TableName() string                { return "roles" }
func (Gateway) TableName() string             { return "gateways" }
func (GatewayPolicy) TableName() string       { return "gateway_policies" }
func (MetricRecord) TableName() string        { return "metric_records" }
func (ModelCatalog) TableName() string        { return "model_catalogs" }
func (APIKey) TableName() string              { return "api_keys" }
func (UserBalance) TableName() string         { return "user_balances" }
func (BalanceTransaction) TableName() string  { return "balance_transactions" }
func (UsageRecord) TableName() string         { return "usage_records" }
func (TeamInvitation) TableName() string      { return "team_invitations" }
func (ModelRechargeLog) TableName() string    { return "model_recharge_logs" }
func (Notification) TableName() string        { return "notifications" }

// ModelRechargeLog 模型上游充值记录
type ModelRechargeLog struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelCatalogID string    `json:"model_catalog_id" gorm:"size:36;not null;index"`
	Amount         float64   `json:"amount"`
	Balance        float64   `json:"balance"`
	Description    string    `json:"description" gorm:"size:200"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"`
}

// Notification 系统通知（余额告警等）
type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    string    `json:"user_id" gorm:"size:36;index"`
	Type      string    `json:"type" gorm:"size:30;not null"`
	Title     string    `json:"title" gorm:"size:100"`
	Content   string    `json:"content" gorm:"size:500"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
