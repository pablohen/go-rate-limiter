package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockRateLimiter struct {
	allowFunc          func(identifier string, maxRequests int, blockDuration time.Duration) (bool, error)
	getIPConfigFunc    func() (int, time.Duration)
	getTokenConfigFunc func() (int, time.Duration)
}

func (m *MockRateLimiter) Allow(identifier string, maxRequests int, blockDuration time.Duration) (bool, error) {
	return m.allowFunc(identifier, maxRequests, blockDuration)
}

func (m *MockRateLimiter) GetIPConfig() (int, time.Duration) {
	return m.getIPConfigFunc()
}

func (m *MockRateLimiter) GetTokenConfig() (int, time.Duration) {
	return m.getTokenConfigFunc()
}

func TestRateLimiterMiddleware_IP(t *testing.T) {
	mockLimiter := &MockRateLimiter{
		allowFunc: func(identifier string, maxRequests int, blockDuration time.Duration) (bool, error) {
			return false, nil
		},
		getIPConfigFunc: func() (int, time.Duration) {
			return 5, time.Minute
		},
		getTokenConfigFunc: func() (int, time.Duration) {
			return 10, 5 * time.Minute
		},
	}

	middleware := RateLimiterMiddleware(mockLimiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != "success" {
		t.Errorf("handler returned wrong body: got %v want %v", rr.Body.String(), "success")
	}
}

func TestRateLimiterMiddleware_Token(t *testing.T) {
	mockLimiter := &MockRateLimiter{
		allowFunc: func(identifier string, maxRequests int, blockDuration time.Duration) (bool, error) {
			if identifier != "test-token" {
				t.Errorf("expected identifier to be test-token, got %s", identifier)
			}
			if maxRequests != 10 {
				t.Errorf("expected max requests to be 10, got %d", maxRequests)
			}
			return false, nil
		},
		getIPConfigFunc: func() (int, time.Duration) {
			return 5, time.Minute
		},
		getTokenConfigFunc: func() (int, time.Duration) {
			return 10, 5 * time.Minute
		},
	}

	middleware := RateLimiterMiddleware(mockLimiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("API_KEY", "test-token")

	rr := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestRateLimiterMiddleware_Limited(t *testing.T) {
	mockLimiter := &MockRateLimiter{
		allowFunc: func(identifier string, maxRequests int, blockDuration time.Duration) (bool, error) {
			return true, nil
		},
		getIPConfigFunc: func() (int, time.Duration) {
			return 5, time.Minute
		},
		getTokenConfigFunc: func() (int, time.Duration) {
			return 10, 5 * time.Minute
		},
	}

	middleware := RateLimiterMiddleware(mockLimiter)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not have been called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	rr := httptest.NewRecorder()

	middleware(testHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusTooManyRequests {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTooManyRequests)
	}

	expectedBody := "you have reached the maximum number of requests or actions allowed within a certain time frame"
	if rr.Body.String() != expectedBody {
		t.Errorf("handler returned wrong body: got %v want %v",
			rr.Body.String(), expectedBody)
	}
}

func TestGetIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	ip := getIP(req)

	if ip != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", ip)
	}

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	ip = getIP(req)

	if ip != "192.168.1.1:12345" {
		t.Errorf("expected IP 192.168.1.1:12345, got %s", ip)
	}
}
