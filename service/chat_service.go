package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/resilience"
	"github.com/aigate/provider"
)

type ChatService struct {
	registry *provider.Registry
	breaker  *resilience.CircuitBreaker
	db       *gorm.DB
}

func NewChatService(registry *provider.Registry, breaker *resilience.CircuitBreaker, db *gorm.DB) *ChatService {
	return &ChatService{
		registry: registry,
		breaker:  breaker,
		db:       db,
	}
}

func (s *ChatService) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("get provider error: %w", err)
	}

	// 查找该 provider 对应的网关策略
	policy := s.findPolicy(req.Provider, req.TenantID)

	if s.breaker == nil || policy == nil || !policy.CircuitBreakerEnabled {
		return p.Chat(ctx, req)
	}

	result, err := s.breaker.Execute(ctx, policy.GatewayID, policy, func() (interface{}, error) {
		return p.Chat(ctx, req)
	})

	if err != nil {
		// 熔断触发，尝试降级
		if policy.FallbackEnabled && policy.FallbackProvider != "" {
			fallbackP, fbErr := s.registry.Get(policy.FallbackProvider)
			if fbErr == nil {
				fallbackReq := *req
				fallbackReq.Provider = policy.FallbackProvider
				if policy.FallbackModel != "" {
					fallbackReq.Model = policy.FallbackModel
				}
				return fallbackP.Chat(ctx, &fallbackReq)
			}
		}
		return nil, err
	}

	return result.(*model.ChatResponse), nil
}

func (s *ChatService) ChatStream(ctx context.Context, req *model.ChatRequest, callback func(chunk *model.StreamChunk) error) error {
	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return fmt.Errorf("get provider error: %w", err)
	}
	return p.ChatStream(ctx, req, callback)
}

func (s *ChatService) ListProviders() []model.ProviderInfo {
	return s.registry.List()
}

func (s *ChatService) findPolicy(providerName string, tenantID string) *model.GatewayPolicy {
	if s.db == nil {
		return nil
	}
	var gateway model.Gateway
	query := s.db.Where("provider = ? AND status = 1", providerName)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if query.First(&gateway).Error != nil {
		return nil
	}

	var policy model.GatewayPolicy
	if s.db.Where("gateway_id = ?", gateway.ID).First(&policy).Error != nil {
		return nil
	}
	return &policy
}
