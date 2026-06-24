package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	defaultClientTTL       = 5 * time.Minute
	defaultCleanupInterval = time.Minute
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a per-IP token-bucket limiter held in memory. Sufficient for a
// single node; a shared (e.g. Redis-backed) limiter is needed across many nodes.
type RateLimiter struct {
	mu              sync.Mutex
	limit           rate.Limit
	burst           int
	clients         map[string]*clientLimiter
	clientTTL       time.Duration
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// NewRateLimiter builds a limiter allowing requestsPerSecond requests/sec with
// bursts up to burst tokens per client IP.
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limit:           rate.Limit(requestsPerSecond),
		burst:           burst,
		clients:         make(map[string]*clientLimiter),
		clientTTL:       defaultClientTTL,
		cleanupInterval: defaultCleanupInterval,
	}
}

// Middleware rejects requests that exceed the limit with 429 Too Many Requests.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.lastCleanup.IsZero() {
		rl.lastCleanup = now
	} else if now.Sub(rl.lastCleanup) >= rl.cleanupInterval {
		rl.cleanupLocked(now)
		rl.lastCleanup = now
	}

	client, ok := rl.clients[ip]
	if !ok {
		client = &clientLimiter{
			limiter: rate.NewLimiter(rl.limit, rl.burst),
		}
		rl.clients[ip] = client
	}

	client.lastSeen = now
	
	return client.limiter.Allow()
}

func (rl *RateLimiter) cleanupLocked(now time.Time) {
	for ip, client := range rl.clients {
		if now.Sub(client.lastSeen) > rl.clientTTL {
			delete(rl.clients, ip)
		}
	}
}
