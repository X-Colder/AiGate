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

// OpenAICompatibleProvider 兼容 OpenAI 接口协议的通用 Provider 基础实现。
// DeepSeek、豆包(Doubao)、通义千问(Qwen)、Kimi(Moonshot) 等国内大模型
// 均提供与 OpenAI 兼容的 /chat/completions 接口，因此可复用此实现。
type OpenAICompatibleProvider struct {
	providerName string                // 提供者唯一标识，如 "deepseek"、"doubao"
	cfg          config.ProviderConfig // 该提供者的配置信息（API Key、BaseURL 等）
	client       *http.Client          // 带超时的 HTTP 客户端
}

// NewOpenAICompatibleProvider 创建一个兼容 OpenAI 协议的 Provider 实例。
// providerName: 提供者标识名称
// cfg: 从配置文件读取的 ProviderConfig
func NewOpenAICompatibleProvider(providerName string, cfg config.ProviderConfig) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		providerName: providerName,
		cfg:          cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// Name 返回提供者的唯一标识名称
func (p *OpenAICompatibleProvider) Name() string {
	return p.providerName
}

// Chat 向兼容 OpenAI 的 /chat/completions 接口发送聊天请求，返回完整响应。
// 请求格式和响应格式均遵循 OpenAI Chat Completions API 标准。
func (p *OpenAICompatibleProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	// 优先使用请求中指定的模型，否则使用配置中的默认模型
	modelName := req.Model
	if modelName == "" {
		modelName = p.cfg.Model
	}

	// 构建符合 OpenAI 格式的请求体
	body := map[string]interface{}{
		"model":    modelName,
		"messages": req.Messages,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("[%s] marshal request error: %w", p.providerName, err)
	}

	// 所有兼容 OpenAI 的服务都使用 /chat/completions 端点
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.cfg.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("[%s] create request error: %w", p.providerName, err)
	}

	// 设置通用请求头：JSON 格式 + Bearer Token 认证
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("[%s] request error: %w", p.providerName, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[%s] read response error: %w", p.providerName, err)
	}

	// 非 200 状态码视为 API 错误，返回原始响应便于排查
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] API error (status %d): %s", p.providerName, resp.StatusCode, string(respBody))
	}

	// 解析 OpenAI 兼容的响应结构
	var apiResp openAICompatibleResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("[%s] unmarshal response error: %w", p.providerName, err)
	}

	// 提取第一个 choice 的内容作为回复
	content := ""
	if len(apiResp.Choices) > 0 {
		content = apiResp.Choices[0].Message.Content
	}

	return &model.ChatResponse{
		ID:       apiResp.ID,
		Provider: p.providerName,
		Model:    apiResp.Model,
		Content:  content,
		Usage: &model.Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream 流式聊天接口（SSE），当前为预留接口，暂未实现。
func (p *OpenAICompatibleProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	return fmt.Errorf("[%s] stream not implemented yet", p.providerName)
}

// openAICompatibleResponse OpenAI 兼容 API 的通用响应结构体，
// 适用于所有兼容 OpenAI Chat Completions 格式的服务
type openAICompatibleResponse struct {
	ID      string `json:"id"`    // 响应唯一 ID
	Model   string `json:"model"` // 实际使用的模型名称
	Choices []struct {
		Message struct {
			Content string `json:"content"` // AI 回复内容
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`     // 提示词 Token 数
		CompletionTokens int `json:"completion_tokens"` // 生成内容 Token 数
		TotalTokens      int `json:"total_tokens"`      // 总 Token 数
	} `json:"usage"`
}
