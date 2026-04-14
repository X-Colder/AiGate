package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aigate/config"
	"github.com/aigate/model"
)

// AnthropicProvider Anthropic 官方 API 提供者实现。
// Anthropic 使用独立的 Messages API（非 OpenAI 兼容），需要单独实现。
// 接口文档: https://docs.anthropic.com/en/docs/api-reference
type AnthropicProvider struct {
	cfg    config.ProviderConfig // Anthropic 的连接配置
	client *http.Client          // 带超时的 HTTP 客户端
}

// NewAnthropicProvider 创建 Anthropic Provider 实例
func NewAnthropicProvider(cfg config.ProviderConfig) *AnthropicProvider {
	return &AnthropicProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// Name 返回提供者标识 "anthropic"
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Chat 向 Anthropic Messages API 发送聊天请求。
// 注意：Anthropic 的 system 消息需要单独提取为顶层字段，不能放在 messages 数组中。
func (p *AnthropicProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	modelName := req.Model
	if modelName == "" {
		modelName = p.cfg.Model
	}

	// Anthropic 要求 system 消息从 messages 数组中分离出来
	messages := make([]map[string]string, 0, len(req.Messages))
	systemPrompt := ""
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建 Anthropic 格式的请求体
	body := map[string]interface{}{
		"model":      modelName,
		"messages":   messages,
		"max_tokens": 4096,
	}
	if systemPrompt != "" {
		body["system"] = systemPrompt
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.cfg.BaseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	// Anthropic 使用 x-api-key 头和版本号认证，而非 Bearer Token
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicChatResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	// 提取第一个 content block 的文本内容
	content := ""
	if len(anthropicResp.Content) > 0 {
		content = anthropicResp.Content[0].Text
	}

	// Anthropic 使用 input_tokens / output_tokens，转换为统一的 Usage 格式
	return &model.ChatResponse{
		ID:       anthropicResp.ID,
		Provider: "anthropic",
		Model:    anthropicResp.Model,
		Content:  content,
		Usage: &model.Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

// ChatStream 流式聊天接口（SSE），当前为预留接口
func (p *AnthropicProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	return fmt.Errorf("stream not implemented yet for Anthropic provider")
}

// anthropicChatResponse Anthropic Messages API 的响应结构
type anthropicChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"` // 内容类型，通常为 "text"
		Text string `json:"text"` // 文本内容
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`  // 输入 Token 数
		OutputTokens int `json:"output_tokens"` // 输出 Token 数
	} `json:"usage"`
}
