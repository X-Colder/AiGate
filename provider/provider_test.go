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
