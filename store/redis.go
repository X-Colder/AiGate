package store

import (
	"context"
	"fmt"

	"github.com/aigate/config"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	if cfg.Addr == "" {
		return nil
	}

	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 50
	}
	minIdle := cfg.MinIdleConns
	if minIdle <= 0 {
		minIdle = 10
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}

func GetRedis() *redis.Client {
	return RDB
}

func CloseRedis() {
	if RDB != nil {
		RDB.Close()
	}
}
