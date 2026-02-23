// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/yourorg/riskengine/internal/metrics"
)

// RateLimitConfig controls the two-tier token-bucket limiter.
type RateLimitConfig struct {
	// GlobalRPS is the maximum requests per second accepted by the whole service.
	GlobalRPS float64
	// GlobalBurst is the maximum burst size at the global level.
	GlobalBurst int
	// PerIPRPS is the maximum requests per second from a single IP address.
	PerIPRPS float64
	// PerIPBurst is the maximum burst size per IP.
	PerIPBurst int
	// CleanupInterval controls how often idle per-IP limiters are evicted.
	// Defaults to 5 minutes.
	CleanupInterval time.Duration
	// IdleTimeout is the duration after which an unused per-IP limiter is removed.
	// Defaults to 10 minutes.
	IdleTimeout time.Duration
}

// DefaultRateLimitConfig returns a sensible production default:
// 5000 RPS globally, 100 RPS per IP.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:       5000,
		GlobalBurst:     500,
		PerIPRPS:        100,
		PerIPBurst:      20,
		CleanupInterval: 5 * time.Minute,
		IdleTimeout:     10 * time.Minute,
	}
}

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiter holds state for the two-tier limiter.
type rateLimiter struct {
	global  *rate.Limiter
	mu      sync.Mutex
	perIP   map[string]*ipEntry
	cfg     RateLimitConfig
}

func newRateLimiter(cfg RateLimitConfig) *rateLimiter {
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 10 * time.Minute
	}
	rl := &rateLimiter{
		global: rate.NewLimiter(rate.Limit(cfg.GlobalRPS), cfg.GlobalBurst),
		perIP:  make(map[string]*ipEntry),
		cfg:    cfg,
	}
	go rl.cleanup()
	return rl
}

// allow returns true if the request from ip should be allowed.
func (rl *rateLimiter) allow(ip string) bool {
	if !rl.global.Allow() {
		return false
	}
	rl.mu.Lock()
	e, ok := rl.perIP[ip]
	if !ok {
		e = &ipEntry{
			limiter: rate.NewLimiter(rate.Limit(rl.cfg.PerIPRPS), rl.cfg.PerIPBurst),
		}
		rl.perIP[ip] = e
	}
	e.lastSeen = time.Now()
	allowed := e.limiter.Allow()
	rl.mu.Unlock()
	return allowed
}

// cleanup periodically removes per-IP limiters that have been idle.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		threshold := time.Now().Add(-rl.cfg.IdleTimeout)
		rl.mu.Lock()
		for ip, e := range rl.perIP {
			if e.lastSeen.Before(threshold) {
				delete(rl.perIP, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a gin middleware that enforces a two-tier token-bucket
// rate limit: a global limiter and a per-IP limiter.
//
// Rejected requests receive HTTP 429 with a JSON body and the
// riskengine_rate_limited_total counter is incremented.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	rl := newRateLimiter(cfg)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			path := c.FullPath()
			if path == "" {
				path = c.Request.URL.Path
			}
			metrics.RateLimitedTotal.WithLabelValues(path).Inc()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "RATE_LIMITED",
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}
