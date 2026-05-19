package middleware

import(
	"net/http"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

// uint16 -> 0 - 65_535
func NewRateLimiter(requestsPerSecond uint16, burst uint16) *RateLimiter{
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), int(burst)),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if !rl.limiter.Allow(){
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w,r)
	})
}