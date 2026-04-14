package provider

import (
	"context"
	"fmt"

	"github.com/aigate/config"
	"github.com/aigate/model"
)

// Provider AI 服务提供者接口
type Provider interface {
	// Name 提供者名称
	Name() string
	// Chat 发送聊天请求
	Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error)
	// ChatStream 发送流式聊天请求
	ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error
}

// Registry 提供者注册表
type Registry struct {
	providers map[string]Provider
}

// NewRegistry 创建提供者注册表
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register 注册提供者
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get 获取提供者
func (r *Registry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}

// List 列出所有提供者
func (r *Registry) List() []model.ProviderInfo {
	var list []model.ProviderInfo
	for name, _ := range r.providers {
		list = append(list, model.ProviderInfo{
			Name:    name,
			Enabled: true,
		})
	}
	return list
}

// InitProviders 根据配置初始化所有启用的提供者
func InitProviders(cfg *config.Config) *Registry {
	registry := NewRegistry()

	for name, providerCfg := range cfg.Providers {
		if !providerCfg.Enabled {
			continue
		}
		switch name {
		case "openai":
			registry.Register(NewOpenAIProvider(providerCfg))
		case "anthropic":
			registry.Register(NewAnthropicProvider(providerCfg))
		}
	}

	return registry
}
