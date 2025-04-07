package http

import (
	"net/http"
	"strings"
	"time"

	"go-rate-limiter/internal/usecase/limiter"
)

func RateLimiterMiddleware(rl limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var identifier string
			var maxReq int
			var blockDur time.Duration

			if token := r.Header.Get("API_KEY"); token != "" {
				identifier = strings.TrimSpace(token)
				maxReq, blockDur = rl.GetTokenConfig()
			} else {
				identifier = getIP(r)
				maxReq, blockDur = rl.GetIPConfig()
			}

			limited, err := rl.Allow(identifier, maxReq, blockDur)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if limited {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getIP(r *http.Request) string {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	}
	return ip
}
