// Package model 定义 AiGate 网关的核心数据结构，
// 包括聊天请求/响应、消息格式、流式块和提供者信息等。
package model

// ChatRequest 聊天请求体，由客户端发送到 POST /api/v1/chat
type ChatRequest struct {
	Provider string    `json:"provider" binding:"required"` // AI 提供者标识: openai, deepseek, doubao, qwen, kimi, anthropic
	Model    string    `json:"model"`                       // 模型名称（可选），未指定时使用提供者的默认模型
	Messages []Message `json:"messages" binding:"required"` // 对话消息列表，至少包含一条
	Stream   bool      `json:"stream"`                      // 是否启用流式响应（SSE），默认 false
	TenantID string    `json:"-"`                           // 内部字段，由中间件注入
}

// Message 单条对话消息
type Message struct {
	Role    string `json:"role" binding:"required"`    // 消息角色: system, user, assistant
	Content string `json:"content" binding:"required"` // 消息文本内容
}

// ChatResponse 聊天响应体，统一所有提供者的返回格式
type ChatResponse struct {
	ID       string `json:"id"`              // 响应唯一标识（由上游 AI 服务生成）
	Provider string `json:"provider"`        // 实际处理请求的提供者名称
	Model    string `json:"model"`           // 实际使用的模型名称
	Content  string `json:"content"`         // AI 生成的回复内容
	Usage    *Usage `json:"usage,omitempty"` // Token 用量统计（部分提供者可能不返回）
}

// Usage Token 消耗统计，统一了不同提供者的字段命名
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 输入提示词消耗的 Token 数
	CompletionTokens int `json:"completion_tokens"` // AI 生成内容消耗的 Token 数
	TotalTokens      int `json:"total_tokens"`      // 总 Token 消耗
}

// StreamChunk 流式响应中的单个数据块（SSE event）
type StreamChunk struct {
	ID       string `json:"id"`       // 块标识
	Provider string `json:"provider"` // 提供者名称
	Model    string `json:"model"`    // 模型名称
	Delta    string `json:"delta"`    // 本次增量文本内容
	Done     bool   `json:"done"`     // 是否为最后一个块
}

// ProviderInfo 提供者信息，用于 GET /api/v1/providers 接口返回
type ProviderInfo struct {
	Name    string   `json:"name"`             // 提供者名称
	Enabled bool     `json:"enabled"`          // 是否已启用
	Models  []string `json:"models,omitempty"` // 可用模型列表（可选）
}
