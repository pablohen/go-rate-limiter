package limiter

import (
	"time"

	"go-rate-limiter/internal/repository"
)

type RateLimiter interface {
	Allow(identifier string, maxRequests int, blockDuration time.Duration) (bool, error)
	GetIPConfig() (int, time.Duration)
	GetTokenConfig() (int, time.Duration)
}

type rateLimiter struct {
	repo               repository.LimiterRepository
	ipMaxRequests      int
	ipBlockDuration    time.Duration
	tokenMaxRequests   int
	tokenBlockDuration time.Duration
}

func NewRateLimiter(repo repository.LimiterRepository, ipMax int, ipBlock time.Duration, tokenMax int, tokenBlock time.Duration) RateLimiter {
	return &rateLimiter{
		repo:               repo,
		ipMaxRequests:      ipMax,
		ipBlockDuration:    ipBlock,
		tokenMaxRequests:   tokenMax,
		tokenBlockDuration: tokenBlock,
	}
}

func (l *rateLimiter) Allow(identifier string, maxReq int, blockDur time.Duration) (bool, error) {
	blocked, err := l.repo.IsBlocked(identifier)
	if err != nil || blocked {
		return true, err
	}

	count, err := l.repo.Increment(identifier, time.Second)
	if err != nil {
		return true, err
	}

	if count > maxReq {
		if err := l.repo.Block(identifier, blockDur); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}

func (l *rateLimiter) GetIPConfig() (int, time.Duration) {
	return l.ipMaxRequests, l.ipBlockDuration
}

func (l *rateLimiter) GetTokenConfig() (int, time.Duration) {
	return l.tokenMaxRequests, l.tokenBlockDuration
}
