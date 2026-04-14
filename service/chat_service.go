// Package service 实现业务逻辑层，负责协调 Provider 和 Handler 之间的调用。
// ChatService 作为聊天业务的核心，通过 Provider Registry 实现多 AI 服务的动态路由。
package service

import (
	"context"
	"fmt"

	"github.com/aigate/model"
	"github.com/aigate/provider"
)

// ChatService 聊天业务服务，持有 Provider 注册表，
// 根据请求中的 provider 字段将请求路由到对应的 AI 服务。
type ChatService struct {
	registry *provider.Registry // 已注册的提供者注册表
}

// NewChatService 创建聊天服务实例
func NewChatService(registry *provider.Registry) *ChatService {
	return &ChatService{
		registry: registry,
	}
}

// Chat 处理普通聊天请求：根据 req.Provider 找到对应 Provider 并转发请求
func (s *ChatService) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("get provider error: %w", err)
	}
	return p.Chat(ctx, req)
}

// ChatStream 处理流式聊天请求：根据 req.Provider 找到对应 Provider 并以流式方式转发
func (s *ChatService) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return fmt.Errorf("get provider error: %w", err)
	}
	return p.ChatStream(ctx, req, callback)
}

// ListProviders 返回所有已注册的提供者信息，供 GET /api/v1/providers 接口使用
func (s *ChatService) ListProviders() []model.ProviderInfo {
	return s.registry.List()
}
