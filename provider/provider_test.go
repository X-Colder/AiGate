package provider

import (
	"context"
	"testing"

	"github.com/aigate/config"
	"github.com/aigate/model"
)

// MockProvider 模拟 Provider 接口
type MockProvider struct {
	name         string
	chatResp     *model.ChatResponse
	chatErr      error
	streamErr    error
	streamChunks []*model.StreamChunk
}

func (m *MockProvider) Name() string { return m.name }

func (m *MockProvider) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	if m.chatErr != nil {
		return nil, m.chatErr
	}
	return m.chatResp, nil
}

func (m *MockProvider) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	if m.streamErr != nil {
		return m.streamErr
	}
	for _, chunk := range m.streamChunks {
		if err := callback(chunk); err != nil {
			return err
		}
	}
	return nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.providers) != 0 {
		t.Errorf("expected empty providers, got %d", len(r.providers))
	}
}

func TestRegistry_Register_And_Get(t *testing.T) {
	r := NewRegistry()
	mock := &MockProvider{
		name: "test-provider",
		chatResp: &model.ChatResponse{
			ID:       "test-id",
			Provider: "test-provider",
			Content:  "hello",
		},
	}

	r.Register(mock)

	p, err := r.Get("test-provider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "test-provider" {
		t.Errorf("expected test-provider, got %s", p.Name())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&MockProvider{name: "provider-a"})
	r.Register(&MockProvider{name: "provider-b"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 providers, got %d", len(list))
	}

	names := make(map[string]bool)
	for _, p := range list {
		names[p.Name] = true
		if !p.Enabled {
			t.Errorf("expected provider %s to be enabled", p.Name)
		}
	}
	if !names["provider-a"] || !names["provider-b"] {
		t.Error("expected both provider-a and provider-b in list")
	}
}

func TestRegistry_List_Empty(t *testing.T) {
	r := NewRegistry()
	list := r.List()
	if list != nil && len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestInitProviders_NoneEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"openai":    {Enabled: false},
			"anthropic": {Enabled: false},
		},
	}

	registry := InitProviders(cfg)
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("expected 0 providers, got %d", len(list))
	}
}

func TestInitProviders_OpenAIEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"openai": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-4",
				Timeout: 30,
			},
		},
	}

	registry := InitProviders(cfg)
	p, err := registry.Get("openai")
	if err != nil {
		t.Fatalf("expected openai provider, got error: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestInitProviders_AnthropicEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"anthropic": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://api.anthropic.com/v1",
				Model:   "claude-3-sonnet",
				Timeout: 30,
			},
		},
	}

	registry := InitProviders(cfg)
	p, err := registry.Get("anthropic")
	if err != nil {
		t.Fatalf("expected anthropic provider, got error: %v", err)
	}
	if p.Name() != "anthropic" {
		t.Errorf("expected anthropic, got %s", p.Name())
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := NewOpenAIProvider(config.ProviderConfig{Timeout: 10})
	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	p := NewAnthropicProvider(config.ProviderConfig{Timeout: 10})
	if p.Name() != "anthropic" {
		t.Errorf("expected anthropic, got %s", p.Name())
	}
}

// =============================================================
// OpenAICompatibleProvider 测试（DeepSeek、豆包、Qwen、Kimi 共用）
// =============================================================

func TestOpenAICompatibleProvider_Name(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"deepseek"},
		{"doubao"},
		{"qwen"},
		{"kimi"},
	}
	for _, tt := range tests {
		p := NewOpenAICompatibleProvider(tt.name, config.ProviderConfig{Timeout: 10})
		if p.Name() != tt.name {
			t.Errorf("expected %s, got %s", tt.name, p.Name())
		}
	}
}

func TestOpenAICompatibleProvider_ChatStream_NotImplemented(t *testing.T) {
	p := NewOpenAICompatibleProvider("deepseek", config.ProviderConfig{Timeout: 10})
	err := p.ChatStream(context.Background(), &model.ChatRequest{}, func(chunk *model.StreamChunk) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for unimplemented stream")
	}
}

func TestInitProviders_DeepSeekEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"deepseek": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://api.deepseek.com/v1",
				Model:   "deepseek-chat",
				Timeout: 60,
			},
		},
	}
	registry := InitProviders(cfg)
	p, err := registry.Get("deepseek")
	if err != nil {
		t.Fatalf("expected deepseek provider, got error: %v", err)
	}
	if p.Name() != "deepseek" {
		t.Errorf("expected deepseek, got %s", p.Name())
	}
}

func TestInitProviders_DoubaoEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"doubao": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
				Model:   "doubao-pro-32k",
				Timeout: 60,
			},
		},
	}
	registry := InitProviders(cfg)
	p, err := registry.Get("doubao")
	if err != nil {
		t.Fatalf("expected doubao provider, got error: %v", err)
	}
	if p.Name() != "doubao" {
		t.Errorf("expected doubao, got %s", p.Name())
	}
}

func TestInitProviders_QwenEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"qwen": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				Model:   "qwen-turbo",
				Timeout: 60,
			},
		},
	}
	registry := InitProviders(cfg)
	p, err := registry.Get("qwen")
	if err != nil {
		t.Fatalf("expected qwen provider, got error: %v", err)
	}
	if p.Name() != "qwen" {
		t.Errorf("expected qwen, got %s", p.Name())
	}
}

func TestInitProviders_KimiEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"kimi": {
				Enabled: true,
				APIKey:  "test-key",
				BaseURL: "https://api.moonshot.cn/v1",
				Model:   "moonshot-v1-8k",
				Timeout: 60,
			},
		},
	}
	registry := InitProviders(cfg)
	p, err := registry.Get("kimi")
	if err != nil {
		t.Fatalf("expected kimi provider, got error: %v", err)
	}
	if p.Name() != "kimi" {
		t.Errorf("expected kimi, got %s", p.Name())
	}
}

func TestInitProviders_AllEnabled(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"openai":    {Enabled: true, Timeout: 30},
			"anthropic": {Enabled: true, Timeout: 30},
			"deepseek":  {Enabled: true, Timeout: 60},
			"doubao":    {Enabled: true, Timeout: 60},
			"qwen":      {Enabled: true, Timeout: 60},
			"kimi":      {Enabled: true, Timeout: 60},
		},
	}
	registry := InitProviders(cfg)
	list := registry.List()
	if len(list) != 6 {
		t.Errorf("expected 6 providers, got %d", len(list))
	}
}
