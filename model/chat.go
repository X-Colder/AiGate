package model

// ChatRequest 聊天请求
type ChatRequest struct {
	Provider string    `json:"provider" binding:"required"` // AI 提供者: openai, anthropic, etc.
	Model    string    `json:"model"`                       // 模型名称，可选，使用配置中的默认模型
	Messages []Message `json:"messages" binding:"required"` // 消息列表
	Stream   bool      `json:"stream"`                      // 是否流式响应
}

// Message 消息
type Message struct {
	Role    string `json:"role" binding:"required"`    // system, user, assistant
	Content string `json:"content" binding:"required"` // 消息内容
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Content  string `json:"content"`
	Usage    *Usage `json:"usage,omitempty"`
}

// Usage token 用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Delta    string `json:"delta"`
	Done     bool   `json:"done"`
}

// ProviderInfo 提供者信息
type ProviderInfo struct {
	Name    string   `json:"name"`
	Enabled bool     `json:"enabled"`
	Models  []string `json:"models,omitempty"`
}
