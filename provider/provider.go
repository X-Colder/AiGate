package provider

import (
	"context"
	"fmt"

	"github.com/aigate/config"
	"github.com/aigate/model"
)

// Provider AI 服务提供者统一接口。
// 所有 AI 服务（OpenAI、Anthropic、DeepSeek、豆包、Qwen、Kimi）都需实现此接口，
// 以便 Service 层通过接口调用，实现提供者的可插拔切换。
type Provider interface {
	// Name 返回提供者的唯一标识名称（如 "openai"、"deepseek"）
	Name() string
	// Chat 发送聊天请求，返回完整的响应结果
	Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error)
	// ChatStream 发送流式聊天请求，通过 callback 逐块返回内容
	ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error
}

// Registry 提供者注册表，管理所有已注册的 AI Provider。
// 通过 name -> Provider 的映射，实现请求时的动态路由分发。
type Registry struct {
	providers map[string]Provider
}

// NewRegistry 创建一个空的提供者注册表
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register 将一个 Provider 注册到注册表中，以 Provider.Name() 为 key
func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

// Get 根据名称获取对应的 Provider，不存在时返回 error
func (r *Registry) Get(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}

// List 返回所有已注册的提供者信息列表，用于 /api/v1/providers 接口
func (r *Registry) List() []model.ProviderInfo {
	var list []model.ProviderInfo
	for name := range r.providers {
		list = append(list, model.ProviderInfo{
			Name:    name,
			Enabled: true,
		})
	}
	return list
}

// InitProviders 根据配置文件初始化所有启用的 AI 提供者，并注册到 Registry。
// 支持的提供者：
//   - openai:    OpenAI (GPT 系列)
//   - anthropic: Anthropic (Claude 系列)
//   - deepseek:  DeepSeek (深度求索，兼容 OpenAI 协议)
//   - doubao:    豆包 (字节跳动，兼容 OpenAI 协议)
//   - qwen:      通义千问 (阿里云，兼容 OpenAI 协议)
//   - kimi:      Kimi / Moonshot (月之暗面，兼容 OpenAI 协议)
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
		// 以下四个国内大模型均兼容 OpenAI 接口协议，复用 OpenAICompatibleProvider
		case "deepseek":
			registry.Register(NewOpenAICompatibleProvider("deepseek", providerCfg))
		case "doubao":
			registry.Register(NewOpenAICompatibleProvider("doubao", providerCfg))
		case "qwen":
			registry.Register(NewOpenAICompatibleProvider("qwen", providerCfg))
		case "kimi":
			registry.Register(NewOpenAICompatibleProvider("kimi", providerCfg))
		}
	}

	return registry
}
