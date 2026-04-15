package model

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
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
}

// ============ 租户管理 DTO ============

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name   string `json:"name"`
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
