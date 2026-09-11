package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides Redis-backed rate limiting that works across multiple instances
type RateLimiter struct {
	rdb       *redis.Client
	rpm       int
	keyPrefix string
}

// NewRateLimiter creates a new Redis-backed rate limiter
func NewRateLimiter(redisURL string, rpm int) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RateLimiter{
		rdb:       rdb,
		rpm:       rpm,
		keyPrefix: "ledgerly:ratelimit:",
	}, nil
}

// Close closes the Redis connection
func (rl *RateLimiter) Close() error {
	return rl.rdb.Close()
}

// Middleware returns an HTTP middleware that enforces rate limits per user
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get user identifier (from auth context or IP)
		identifier := r.Header.Get("X-User-ID")
		if identifier == "" {
			identifier = getClientIP(r)
		}

		key := fmt.Sprintf("%s%s", rl.keyPrefix, identifier)

		// Sliding window rate limiter using Redis
		allowed, remaining, resetAt, err := rl.checkRateLimit(ctx, key)
		if err != nil {
			// If Redis fails, allow the request (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rpm))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetAt).Seconds())+1))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","message":"Too many requests. Please try again later."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkRateLimit implements a sliding window rate limiter
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	pipe := rl.rdb.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))

	// Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: now.UnixNano(),
	})

	// Set expiry on the key
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, rl.rpm, now.Add(time.Minute), err
	}

	count := int(countCmd.Val())
	remaining := rl.rpm - count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Add(time.Minute)

	if count > rl.rpm {
		return false, 0, resetAt, nil
	}

	return true, remaining, resetAt, nil
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For (from load balancer/proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// NewInMemoryRateLimiter creates a simple in-memory rate limiter for testing
// Not suitable for production with multiple instances
func NewInMemoryRateLimiter(rpm int) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		rpm:     rpm,
		clients: make(map[string][]time.Time),
	}
}

type InMemoryRateLimiter struct {
	rpm     int
	clients map[string][]time.Time
}

func (rl *InMemoryRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := getClientIP(r)
		now := time.Now()
		windowStart := now.Add(-1 * time.Minute)

		// Clean old entries
		var valid []time.Time
		for _, t := range rl.clients[identifier] {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		valid = append(valid, now)
		rl.clients[identifier] = valid

		if len(valid) > rl.rpm {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
