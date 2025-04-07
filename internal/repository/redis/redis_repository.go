package redis

import (
	"context"
	"fmt"
	"time"

	"go-rate-limiter/internal/repository"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) repository.LimiterRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) IsBlocked(key string) (bool, error) {
	ctx := context.Background()
	blockKey := fmt.Sprintf("limiter:block:%s", key)
	exists, err := r.client.Exists(ctx, blockKey).Result()
	return exists > 0, err
}

func (r *RedisRepository) Increment(key string, window time.Duration) (int, error) {
	ctx := context.Background()
	countKey := fmt.Sprintf("limiter:count:%s", key)

	script := `
	local current = redis.call('INCR', KEYS[1])
	if current == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end
	return current
	`
	result, err := r.client.Eval(ctx, script, []string{countKey}, window.Seconds()).Result()
	if err != nil {
		return 0, err
	}
	return int(result.(int64)), nil
}

func (r *RedisRepository) Block(key string, duration time.Duration) error {
	ctx := context.Background()
	blockKey := fmt.Sprintf("limiter:block:%s", key)
	return r.client.Set(ctx, blockKey, "1", duration).Err()
}
