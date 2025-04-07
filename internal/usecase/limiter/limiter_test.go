package limiter

import (
	"errors"
	"testing"
	"time"
)

type MockLimiterRepository struct {
	isBlockedFunc func(key string) (bool, error)
	incrementFunc func(key string, window time.Duration) (int, error)
	blockFunc     func(key string, duration time.Duration) error
}

func (m *MockLimiterRepository) IsBlocked(key string) (bool, error) {
	return m.isBlockedFunc(key)
}

func (m *MockLimiterRepository) Increment(key string, window time.Duration) (int, error) {
	return m.incrementFunc(key, window)
}

func (m *MockLimiterRepository) Block(key string, duration time.Duration) error {
	return m.blockFunc(key, duration)
}

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		maxRequests    int
		blockDuration  time.Duration
		setupMock      func(mock *MockLimiterRepository)
		expectedResult bool
		expectedErr    error
	}{
		{
			name:          "Should allow when under limit",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return false, nil
				}
				mock.incrementFunc = func(key string, window time.Duration) (int, error) {
					return 3, nil
				}
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name:          "Should block when over limit",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return false, nil
				}
				mock.incrementFunc = func(key string, window time.Duration) (int, error) {
					return 6, nil
				}
				mock.blockFunc = func(key string, duration time.Duration) error {
					return nil
				}
			},
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name:          "Should return blocked when already blocked",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return true, nil
				}
			},
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name:          "Should return error when IsBlocked fails",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return false, errors.New("redis error")
				}
			},
			expectedResult: true,
			expectedErr:    errors.New("redis error"),
		},
		{
			name:          "Should return error when Increment fails",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return false, nil
				}
				mock.incrementFunc = func(key string, window time.Duration) (int, error) {
					return 0, errors.New("redis error")
				}
			},
			expectedResult: true,
			expectedErr:    errors.New("redis error"),
		},
		{
			name:          "Should return error when Block fails",
			identifier:    "test-ip",
			maxRequests:   5,
			blockDuration: time.Minute,
			setupMock: func(mock *MockLimiterRepository) {
				mock.isBlockedFunc = func(key string) (bool, error) {
					return false, nil
				}
				mock.incrementFunc = func(key string, window time.Duration) (int, error) {
					return 6, nil
				}
				mock.blockFunc = func(key string, duration time.Duration) error {
					return errors.New("redis error")
				}
			},
			expectedResult: true,
			expectedErr:    errors.New("redis error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockLimiterRepository{}
			tt.setupMock(mockRepo)

			limiter := NewRateLimiter(
				mockRepo,
				10,
				time.Minute,
				20,
				time.Minute*5,
			)

			blocked, err := limiter.Allow(tt.identifier, tt.maxRequests, tt.blockDuration)

			if blocked != tt.expectedResult {
				t.Errorf("expected blocked=%v, got %v", tt.expectedResult, blocked)
			}

			if (err != nil && tt.expectedErr == nil) ||
				(err == nil && tt.expectedErr != nil) ||
				(err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("expected error=%v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRateLimiter_GetIPConfig(t *testing.T) {
	mockRepo := &MockLimiterRepository{}
	expectedMaxRequests := 5
	expectedBlockDuration := 2 * time.Minute

	limiter := NewRateLimiter(
		mockRepo,
		expectedMaxRequests,
		expectedBlockDuration,
		10,
		5*time.Minute,
	)

	maxReq, blockDur := limiter.GetIPConfig()

	if maxReq != expectedMaxRequests {
		t.Errorf("expected max requests=%d, got %d", expectedMaxRequests, maxReq)
	}

	if blockDur != expectedBlockDuration {
		t.Errorf("expected block duration=%v, got %v", expectedBlockDuration, blockDur)
	}
}

func TestRateLimiter_GetTokenConfig(t *testing.T) {
	mockRepo := &MockLimiterRepository{}
	expectedMaxRequests := 10
	expectedBlockDuration := 5 * time.Minute

	limiter := NewRateLimiter(
		mockRepo,
		5,
		2*time.Minute,
		expectedMaxRequests,
		expectedBlockDuration,
	)

	maxReq, blockDur := limiter.GetTokenConfig()

	if maxReq != expectedMaxRequests {
		t.Errorf("expected max requests=%d, got %d", expectedMaxRequests, maxReq)
	}

	if blockDur != expectedBlockDuration {
		t.Errorf("expected block duration=%v, got %v", expectedBlockDuration, blockDur)
	}
}
