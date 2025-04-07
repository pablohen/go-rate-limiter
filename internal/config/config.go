package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LimiterIPMaxRequests      int
	LimiterIPBlockDuration    time.Duration
	LimiterTokenMaxRequests   int
	LimiterTokenBlockDuration time.Duration
	RedisAddress              string
	RedisPassword             string
	RedisDB                   int
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("LIMITER_IP_MAX_REQUESTS", 5)
	viper.SetDefault("LIMITER_IP_BLOCK_DURATION", "5m")
	viper.SetDefault("LIMITER_TOKEN_MAX_REQUESTS", 10)
	viper.SetDefault("LIMITER_TOKEN_BLOCK_DURATION", "5m")
	viper.SetDefault("REDIS_ADDRESS", "redis:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return &Config{
		LimiterIPMaxRequests:      viper.GetInt("LIMITER_IP_MAX_REQUESTS"),
		LimiterIPBlockDuration:    viper.GetDuration("LIMITER_IP_BLOCK_DURATION"),
		LimiterTokenMaxRequests:   viper.GetInt("LIMITER_TOKEN_MAX_REQUESTS"),
		LimiterTokenBlockDuration: viper.GetDuration("LIMITER_TOKEN_BLOCK_DURATION"),
		RedisAddress:              viper.GetString("REDIS_ADDRESS"),
		RedisPassword:             viper.GetString("REDIS_PASSWORD"),
		RedisDB:                   viper.GetInt("REDIS_DB"),
	}, nil
}
