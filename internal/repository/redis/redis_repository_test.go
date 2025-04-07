package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func setupRedis() (*redis.Client, func()) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})

	client.FlushDB(context.Background())

	return client, func() {
		client.FlushDB(context.Background())
		client.Close()
	}
}

func TestRedisRepository_IsBlocked(t *testing.T) {
	client, cleanup := setupRedis()
	defer cleanup()

	repo := NewRedisRepository(client)
	testKey := "test-key-isblocked"

	err := client.Set(context.Background(), "limiter:block:"+testKey, "1", time.Minute).Err()
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}

	blocked, err := repo.IsBlocked(testKey)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !blocked {
		t.Errorf("Expected key to be blocked, but it wasn't")
	}

	blocked, err = repo.IsBlocked("nonexistent-key")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if blocked {
		t.Errorf("Expected key to not be blocked, but it was")
	}
}

func TestRedisRepository_Increment(t *testing.T) {
	client, cleanup := setupRedis()
	defer cleanup()

	repo := NewRedisRepository(client)
	testKey := "test-key-increment"

	count1, err := repo.Increment(testKey, time.Minute)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	count2, err := repo.Increment(testKey, time.Minute)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	if count1 != 1 {
		t.Errorf("Expected first count to be 1, got %d", count1)
	}
	if count2 != 2 {
		t.Errorf("Expected second count to be 2, got %d", count2)
	}

	ttl := client.TTL(context.Background(), "limiter:count:"+testKey).Val()
	if ttl <= 0 {
		t.Errorf("Expected TTL to be set, got %v", ttl)
	}
}

func TestRedisRepository_Block(t *testing.T) {
	client, cleanup := setupRedis()
	defer cleanup()

	repo := NewRedisRepository(client)
	testKey := "test-key-block"
	blockDuration := 1 * time.Minute

	err := repo.Block(testKey, blockDuration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	blocked, err := repo.IsBlocked(testKey)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !blocked {
		t.Errorf("Expected key to be blocked, but it wasn't")
	}

	ttl := client.TTL(context.Background(), "limiter:block:"+testKey).Val()
	if ttl <= 0 {
		t.Errorf("Expected TTL to be set, got %v", ttl)
	}
}
