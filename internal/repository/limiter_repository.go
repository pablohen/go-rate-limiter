package repository

import "time"

type LimiterRepository interface {
	IsBlocked(key string) (bool, error)
	Increment(key string, window time.Duration) (int, error)
	Block(key string, duration time.Duration) error
}
