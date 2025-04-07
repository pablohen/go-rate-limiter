package main

import (
	"context"
	"log"
	"net/http"

	"go-rate-limiter/internal/config"
	"go-rate-limiter/internal/usecase/limiter"

	"github.com/redis/go-redis/v9"

	custom_http "go-rate-limiter/internal/delivery/http"
	redis_repo "go-rate-limiter/internal/repository/redis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error: ", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection failed: ", err)
	}

	repo := redis_repo.NewRedisRepository(rdb)
	limiter := limiter.NewRateLimiter(
		repo,
		cfg.LimiterIPMaxRequests,
		cfg.LimiterIPBlockDuration,
		cfg.LimiterTokenMaxRequests,
		cfg.LimiterTokenBlockDuration,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: custom_http.RateLimiterMiddleware(limiter)(mux),
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
