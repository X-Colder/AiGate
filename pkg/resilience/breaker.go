package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	gobreaker "github.com/sony/gobreaker/v2"

	"github.com/aigate/model"
)

type CircuitBreaker struct {
	breakers sync.Map
	rdb      *redis.Client
}

func NewCircuitBreaker(rdb *redis.Client) *CircuitBreaker {
	return &CircuitBreaker{rdb: rdb}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, gatewayID string, policy *model.GatewayPolicy, fn func() (interface{}, error)) (interface{}, error) {
	if policy == nil || !policy.CircuitBreakerEnabled {
		return fn()
	}

	breaker := cb.getOrCreate(gatewayID, policy)
	result, err := breaker.Execute(func() (interface{}, error) {
		return fn()
	})

	// 同步状态到 Redis（用于跨实例可见性）
	if cb.rdb != nil {
		state := breaker.State()
		_ = cb.rdb.Set(ctx, fmt.Sprintf("cb:state:%s", gatewayID), state.String(), 60*time.Second).Err()
	}

	return result, err
}

func (cb *CircuitBreaker) State(gatewayID string) gobreaker.State {
	if v, ok := cb.breakers.Load(gatewayID); ok {
		return v.(*gobreaker.CircuitBreaker[interface{}]).State()
	}
	return gobreaker.StateClosed
}

func (cb *CircuitBreaker) getOrCreate(gatewayID string, policy *model.GatewayPolicy) *gobreaker.CircuitBreaker[interface{}] {
	if v, ok := cb.breakers.Load(gatewayID); ok {
		return v.(*gobreaker.CircuitBreaker[interface{}])
	}

	threshold := policy.CircuitBreakerThreshold
	if threshold <= 0 {
		threshold = 0.5
	}
	timeout := policy.CircuitBreakerTimeout
	if timeout <= 0 {
		timeout = 30
	}
	minReqs := policy.CircuitBreakerMinReqs
	if minReqs <= 0 {
		minReqs = 10
	}

	settings := gobreaker.Settings{
		Name:    gatewayID,
		Timeout: time.Duration(timeout) * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < uint32(minReqs) {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= threshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			if cb.rdb != nil {
				_ = cb.rdb.Set(context.Background(), fmt.Sprintf("cb:state:%s", name), to.String(), 60*time.Second).Err()
			}
		},
	}

	breaker := gobreaker.NewCircuitBreaker[interface{}](settings)
	actual, _ := cb.breakers.LoadOrStore(gatewayID, breaker)
	return actual.(*gobreaker.CircuitBreaker[interface{}])
}
