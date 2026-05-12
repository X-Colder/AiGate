package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	RoleCachePrefix = "aigate:role:"
	RoleCacheTTL    = 60 * time.Second

	PolicyCachePrefix = "aigate:policy:"
	PolicyCacheTTL    = 30 * time.Second
)

type RolePermissions struct {
	TenantAccess  bool `json:"tenant_access"`
	GatewayAccess bool `json:"gateway_access"`
	MonitorAccess bool `json:"monitor_access"`
}

func CacheRolePermissions(ctx context.Context, roleID string, perms *RolePermissions) error {
	if RDB == nil {
		return nil
	}
	data, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, RoleCachePrefix+roleID, data, RoleCacheTTL).Err()
}

func GetCachedRolePermissions(ctx context.Context, roleID string) (*RolePermissions, error) {
	if RDB == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	data, err := RDB.Get(ctx, RoleCachePrefix+roleID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var perms RolePermissions
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil, err
	}
	return &perms, nil
}

func InvalidateRoleCache(ctx context.Context, roleID string) error {
	if RDB == nil {
		return nil
	}
	return RDB.Del(ctx, RoleCachePrefix+roleID).Err()
}

func CacheGatewayPolicy(ctx context.Context, gatewayID string, policy interface{}) error {
	if RDB == nil {
		return nil
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, PolicyCachePrefix+gatewayID, data, PolicyCacheTTL).Err()
}

func GetCachedGatewayPolicy(ctx context.Context, gatewayID string) ([]byte, error) {
	if RDB == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	data, err := RDB.Get(ctx, PolicyCachePrefix+gatewayID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

func InvalidatePolicyCache(ctx context.Context, gatewayID string) error {
	if RDB == nil {
		return nil
	}
	return RDB.Del(ctx, PolicyCachePrefix+gatewayID).Err()
}
