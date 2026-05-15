package model

import "time"

// ============ 认证相关 DTO ============

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Phone    string `json:"phone"`
}

type LoginResponse struct {
	Token       string      `json:"token"`
	UserID      string      `json:"user_id"`
	Username    string      `json:"username"`
	TenantID    string      `json:"tenant_id"`
	Role        string      `json:"role"`
	Permissions Permissions `json:"permissions"`
}

type Permissions struct {
	TenantAccess  bool `json:"tenant_access"`
	GatewayAccess bool `json:"gateway_access"`
	MonitorAccess bool `json:"monitor_access"`
	ModelAccess   bool `json:"model_access"`
	APIAccess     bool `json:"api_access"`
	TeamAccess    bool `json:"team_access"`
}

// ============ 团队管理 DTO ============

type CreateTeamRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type InviteMemberRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type InvitationResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	TeamName  string    `json:"team_name"`
	InviterID string    `json:"inviter_id"`
	Inviter   string    `json:"inviter"`
	Phone     string    `json:"phone"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMemberDetail struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Phone     string  `json:"phone"`
	Role      string  `json:"role"`
	Status    int     `json:"status"`
	UsedQuota float64 `json:"used_quota"`
	JoinedAt  string  `json:"joined_at"`
}

// ============ 租户管理 DTO (admin) ============

type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	AdminUser string `json:"admin_user" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6,max=50"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type UpdateTenantRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status *int   `json:"status"`
}

type TenantDetail struct {
	Tenant
	UserCount    int64 `json:"user_count"`
	GatewayCount int64 `json:"gateway_count"`
}

type TenantUsage struct {
	TenantID    string         `json:"tenant_id"`
	TenantName  string         `json:"tenant_name"`
	Gateways    []GatewayUsage `json:"gateways"`
	TotalTokens int64          `json:"total_tokens"`
	TotalReqs   int64          `json:"total_requests"`
}

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

type CreateGatewayRequest struct {
	Name     string `json:"name" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Timeout  int    `json:"timeout"`
}

type UpdateGatewayRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Timeout  int    `json:"timeout"`
	Status   *int   `json:"status"`
}

type UpdatePolicyRequest struct {
	RateLimitEnabled        *bool    `json:"rate_limit_enabled"`
	RateLimitQPS            *int     `json:"rate_limit_qps"`
	RateLimitBurst          *int     `json:"rate_limit_burst"`
	CircuitBreakerEnabled   *bool    `json:"circuit_breaker_enabled"`
	CircuitBreakerThreshold *float64 `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   *int     `json:"circuit_breaker_timeout"`
	CircuitBreakerMinReqs   *int     `json:"circuit_breaker_min_reqs"`
	FallbackEnabled         *bool    `json:"fallback_enabled"`
	FallbackGatewayID       *string  `json:"fallback_gateway_id"`
}

// ============ 监控相关 DTO ============

type MetricQuery struct {
	GatewayID string `form:"gateway_id"`
	TenantID  string `form:"tenant_id"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

type MetricSummary struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	TotalErrors   int64   `json:"total_errors"`
	UniqueUsers   int64   `json:"unique_users"`
	ErrorRate     float64 `json:"error_rate"`
}

// ============ RBAC 角色管理 DTO ============

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=50"`
	Description   string `json:"description"`
	TenantAccess  bool   `json:"tenant_access"`
	GatewayAccess bool   `json:"gateway_access"`
	MonitorAccess bool   `json:"monitor_access"`
	ModelAccess   bool   `json:"model_access"`
	APIAccess     bool   `json:"api_access"`
	TeamAccess    bool   `json:"team_access"`
}

type UpdateRoleRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	TenantAccess  *bool   `json:"tenant_access"`
	GatewayAccess *bool   `json:"gateway_access"`
	MonitorAccess *bool   `json:"monitor_access"`
	ModelAccess   *bool   `json:"model_access"`
	APIAccess     *bool   `json:"api_access"`
	TeamAccess    *bool   `json:"team_access"`
}

type RoleDetail struct {
	Role
	UserCount int64 `json:"user_count"`
}

// ============ 用户管理 DTO ============

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Phone    string `json:"phone"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id" binding:"required"`
}

type UpdateUserRequest struct {
	Password *string `json:"password"`
	RoleID   *string `json:"role_id"`
	Status   *int    `json:"status"`
}

type UserDetail struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Username   string `json:"username"`
	Phone      string `json:"phone"`
	RoleID     string `json:"role_id"`
	RoleName   string `json:"role_name"`
	Role       string `json:"role"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ============ 模型管理 DTO ============

type CreateModelRequest struct {
	Name                string  `json:"name" binding:"required,min=2,max=100"`
	Provider            string  `json:"provider" binding:"required"`
	ModelID             string  `json:"model_id" binding:"required"`
	Description         string  `json:"description"`
	GatewayID           string  `json:"gateway_id"`
	BillingMode         string  `json:"billing_mode"`
	InputPricePer1K     float64 `json:"input_price_per_1k"`
	OutputPricePer1K    float64 `json:"output_price_per_1k"`
	RequestPrice        float64 `json:"request_price"`
	UpstreamInputPer1K  float64 `json:"upstream_input_per_1k"`
	UpstreamOutputPer1K float64 `json:"upstream_output_per_1k"`
	AlertThreshold      float64 `json:"alert_threshold"`
	FreeQuota           int64   `json:"free_quota"`
	MonthlyQuota        int64   `json:"monthly_quota"`
	MaxContextLength    int     `json:"max_context_length"`
	DocContent          string  `json:"doc_content"`
}

type UpdateModelRequest struct {
	Name                *string  `json:"name"`
	Provider            *string  `json:"provider"`
	ModelID             *string  `json:"model_id"`
	Description         *string  `json:"description"`
	GatewayID           *string  `json:"gateway_id"`
	BillingMode         *string  `json:"billing_mode"`
	InputPricePer1K     *float64 `json:"input_price_per_1k"`
	OutputPricePer1K    *float64 `json:"output_price_per_1k"`
	RequestPrice        *float64 `json:"request_price"`
	UpstreamInputPer1K  *float64 `json:"upstream_input_per_1k"`
	UpstreamOutputPer1K *float64 `json:"upstream_output_per_1k"`
	AlertThreshold      *float64 `json:"alert_threshold"`
	FreeQuota           *int64   `json:"free_quota"`
	MonthlyQuota        *int64   `json:"monthly_quota"`
	MaxContextLength    *int     `json:"max_context_length"`
	Status              *int     `json:"status"`
	SortOrder           *int     `json:"sort_order"`
}

type UpdateModelDocRequest struct {
	DocContent string `json:"doc_content" binding:"required"`
}

// ============ API Key DTO ============

type CreateAPIKeyRequest struct {
	Name         string     `json:"name" binding:"required,min=1,max=100"`
	Models       string     `json:"models" binding:"required"`
	RateLimitQPM int        `json:"rate_limit_qpm"`
	ExpiresAt    *time.Time `json:"expires_at"`
}

type UpdateAPIKeyRequest struct {
	Models    *string    `json:"models"`
	ExpiresAt *time.Time `json:"expires_at"`
	Status    *int       `json:"status"`
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

// ============ 模型财务 DTO ============

type ModelRechargeRequest struct {
	ModelID     string  `json:"model_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type ModelFinanceSummary struct {
	ModelID           string  `json:"model_id"`
	ModelName         string  `json:"model_name"`
	Provider          string  `json:"provider"`
	UserCount         int64   `json:"user_count"`
	TotalRevenue      float64 `json:"total_revenue"`
	UpstreamBalance   float64 `json:"upstream_balance"`
	UpstreamRecharge  float64 `json:"upstream_total_recharge"`
	UpstreamCost      float64 `json:"upstream_total_cost"`
	Profit            float64 `json:"profit"`
	AlertThreshold    float64 `json:"alert_threshold"`
	Status            int     `json:"status"`
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
	ID       string `json:"id"`
	Object   string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

type OpenAIModelList struct {
	Object string            `json:"object"`
	Data   []OpenAIModelItem `json:"data"`
}
