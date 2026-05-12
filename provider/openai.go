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

// OpenAIProvider OpenAI 官方 API 提供者实现。
// 通过 OpenAI Chat Completions API 与 GPT 系列模型进行交互。
// 接口文档: https://platform.openai.com/docs/api-reference/chat
type OpenAIProvider struct {
	cfg    config.ProviderConfig // OpenAI 的连接配置
	client *http.Client          // 带超时的 HTTP 客户端
}

// NewOpenAIProvider 创建 OpenAI Provider 实例，使用指定配置初始化 HTTP 客户端
func NewOpenAIProvider(cfg config.ProviderConfig) *OpenAIProvider {
	return &OpenAIProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				MaxConnsPerHost:     50,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Name 返回提供者标识 "openai"
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Chat 向 OpenAI /chat/completions 接口发送聊天请求。
// 支持通过 req.Model 指定模型，未指定时使用配置中的默认模型。
func (p *OpenAIProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	modelName := req.Model
	if modelName == "" {
		modelName = p.cfg.Model
	}

	// 构建 OpenAI 标准请求体
	body := map[string]interface{}{
		"model":    modelName,
		"messages": req.Messages,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.cfg.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// 解析 OpenAI 响应并提取第一个 choice 的内容
	var openaiResp openAIChatResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	content := ""
	if len(openaiResp.Choices) > 0 {
		content = openaiResp.Choices[0].Message.Content
	}

	return &model.ChatResponse{
		ID:       openaiResp.ID,
		Provider: "openai",
		Model:    openaiResp.Model,
		Content:  content,
		Usage: &model.Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream 流式聊天接口（SSE），当前为预留接口
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	return fmt.Errorf("stream not implemented yet for OpenAI provider")
}

// openAIChatResponse OpenAI Chat Completions API 的响应结构
type openAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
