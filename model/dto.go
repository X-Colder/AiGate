package model

import "time"

// ============ 认证相关 DTO ============

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	TenantID string `json:"tenant_id" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token       string      `json:"token"`
	UserID      string      `json:"user_id"`
	Username    string      `json:"username"`
	TenantID    string      `json:"tenant_id"`
	Role        string      `json:"role"`
	Permissions Permissions `json:"permissions"`
}

// Permissions 模块访问权限
type Permissions struct {
	TenantAccess  bool `json:"tenant_access"`
	GatewayAccess bool `json:"gateway_access"`
	MonitorAccess bool `json:"monitor_access"`
	APIAccess     bool `json:"api_access"`
}

// ============ 租户管理 DTO ============

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	AdminUser string `json:"admin_user" binding:"required,min=3,max=50"` // 管理员用户名
	Password  string `json:"password" binding:"required,min=6,max=50"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status *int   `json:"status"`
}

// TenantDetail 租户详情（含统计数据，管理员视图）
type TenantDetail struct {
	Tenant
	UserCount    int64 `json:"user_count"`
	GatewayCount int64 `json:"gateway_count"`
}

// TenantUsage 租户使用详情
type TenantUsage struct {
	TenantID    string         `json:"tenant_id"`
	TenantName  string         `json:"tenant_name"`
	Gateways    []GatewayUsage `json:"gateways"`
	TotalTokens int64          `json:"total_tokens"`
	TotalReqs   int64          `json:"total_requests"`
}

// GatewayUsage 单个网关使用数据
type GatewayUsage struct {
	GatewayID    string  `json:"gateway_id"`
	GatewayName  string  `json:"gateway_name"`
	Provider     string  `json:"provider"`
	Status       int     `json:"status"`
	TokensUsed   int64   `json:"tokens_used"`
	RequestCount int64   `json:"request_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
}

// ============ 网关相关 DTO ============

// CreateGatewayRequest 创建网关请求
type CreateGatewayRequest struct {
	Name     string `json:"name" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	Timeout  int    `json:"timeout"`
}

// UpdateGatewayRequest 更新网关请求
type UpdateGatewayRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	Timeout  int    `json:"timeout"`
	Status   *int   `json:"status"` // 指针类型以区分零值
}

// UpdatePolicyRequest 更新网关策略请求
type UpdatePolicyRequest struct {
	RateLimitEnabled        *bool    `json:"rate_limit_enabled"`
	RateLimitQPS            *int     `json:"rate_limit_qps"`
	RateLimitBurst          *int     `json:"rate_limit_burst"`
	CircuitBreakerEnabled   *bool    `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold *float64 `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   *int     `json:"circuit_breaker_timeout"`
	CircuitBreakerMinReqs   *int     `json:"circuit_breaker_min_reqs"`
	FallbackEnabled         *bool    `json:"fallback_enabled"`
	FallbackProvider        *string  `json:"fallback_provider"`
	FallbackModel           *string  `json:"fallback_model"`
}

// ============ 监控相关 DTO ============

// MetricQuery 监控数据查询参数
type MetricQuery struct {
	GatewayID string `form:"gateway_id"`
	TenantID  string `form:"tenant_id"`                     // admin 按租户筛选
	StartDate string `form:"start_date" binding:"required"` // 格式: 2006-01-02
	EndDate   string `form:"end_date" binding:"required"`   // 格式: 2006-01-02
}

// MetricSummary 监控汇总数据
type MetricSummary struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	TotalErrors   int64   `json:"total_errors"`
	UniqueUsers   int64   `json:"unique_users"`
	ErrorRate     float64 `json:"error_rate"`
}

// ============ RBAC 角色管理 DTO ============

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=50"`
	Description   string `json:"description"`
	TenantAccess  bool   `json:"tenant_access"`
	GatewayAccess bool   `json:"gateway_access"`
	MonitorAccess bool   `json:"monitor_access"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	TenantAccess  *bool   `json:"tenant_access"`
	GatewayAccess *bool   `json:"gateway_access"`
	MonitorAccess *bool   `json:"monitor_access"`
}

// RoleDetail 角色详情（含用户数）
type RoleDetail struct {
	Role
	UserCount int64 `json:"user_count"`
}

// ============ 用户管理 DTO ============

// CreateUserRequest 创建用户请求（管理员创建）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	TenantID string `json:"tenant_id" binding:"required"`
	RoleID   string `json:"role_id" binding:"required"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Password *string `json:"password"`
	RoleID   *string `json:"role_id"`
	Status   *int    `json:"status"`
}

// UserDetail 用户详情（含角色信息和租户名）
type UserDetail struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Username   string `json:"username"`
	RoleID     string `json:"role_id"`
	RoleName   string `json:"role_name"`
	Role       string `json:"role"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ============ 模型管理 DTO ============

type CreateModelRequest struct {
	Name             string  `json:"name" binding:"required,min=2,max=100"`
	Provider         string  `json:"provider" binding:"required"`
	ModelID          string  `json:"model_id" binding:"required"`
	Description      string  `json:"description"`
	BillingMode      string  `json:"billing_mode"`
	InputPricePer1K  float64 `json:"input_price_per_1k"`
	OutputPricePer1K float64 `json:"output_price_per_1k"`
	RequestPrice     float64 `json:"request_price"`
	FreeQuota        int64   `json:"free_quota"`
	MonthlyQuota     int64   `json:"monthly_quota"`
	MaxContextLength int     `json:"max_context_length"`
	DocContent       string  `json:"doc_content"`
}

type UpdateModelRequest struct {
	Name             *string  `json:"name"`
	Provider         *string  `json:"provider"`
	ModelID          *string  `json:"model_id"`
	Description      *string  `json:"description"`
	BillingMode      *string  `json:"billing_mode"`
	InputPricePer1K  *float64 `json:"input_price_per_1k"`
	OutputPricePer1K *float64 `json:"output_price_per_1k"`
	RequestPrice     *float64 `json:"request_price"`
	FreeQuota        *int64   `json:"free_quota"`
	MonthlyQuota     *int64   `json:"monthly_quota"`
	MaxContextLength *int     `json:"max_context_length"`
	Status           *int     `json:"status"`
	SortOrder        *int     `json:"sort_order"`
}

type UpdateModelDocRequest struct {
	DocContent string `json:"doc_content" binding:"required"`
}

// ============ API Key DTO ============

type CreateAPIKeyRequest struct {
	Name         string     `json:"name" binding:"required,min=1,max=100"`
	RateLimitQPM int        `json:"rate_limit_qpm"`
	Models       string     `json:"models"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type APIKeyResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Key          string     `json:"key,omitempty"`
	RateLimitQPM int        `json:"rate_limit_qpm"`
	Models       string     `json:"models"`
	Status       int        `json:"status"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ============ 余额/计费 DTO ============

type RechargeRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type BalanceResponse struct {
	Balance        float64 `json:"balance"`
	FreeBalance    float64 `json:"free_balance"`
	TotalRecharged float64 `json:"total_recharged"`
	TotalConsumed  float64 `json:"total_consumed"`
	MonthlyUsed    int64   `json:"monthly_used"`
	MonthlyQuota   int64   `json:"monthly_quota"`
}

type UserBalanceDetail struct {
	UserID         string  `json:"user_id"`
	Username       string  `json:"username"`
	TenantName     string  `json:"tenant_name"`
	Balance        float64 `json:"balance"`
	FreeBalance    float64 `json:"free_balance"`
	TotalRecharged float64 `json:"total_recharged"`
	TotalConsumed  float64 `json:"total_consumed"`
}

// ============ 用量查询 DTO ============

type UsageQuery struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
	ModelID   string `form:"model_id"`
	APIKeyID  string `form:"api_key_id"`
}

type UsageSummary struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
}

// ============ OpenAI 兼容推理 DTO ============

type OpenAICompletionRequest struct {
	Model       string    `json:"model" binding:"required"`
	Messages    []Message `json:"messages" binding:"required,min=1"`
	Stream      bool      `json:"stream"`
	Temperature *float64  `json:"temperature"`
	TopP        *float64  `json:"top_p"`
	MaxTokens   int       `json:"max_tokens"`
}

type OpenAICompletionResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

type OpenAIChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIModelItem struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

type OpenAIModelList struct {
	Object string            `json:"object"`
	Data   []OpenAIModelItem `json:"data"`
}
