package auth

import (
	"sync"
	"time"
)

type tokenCache struct {
	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

func newTokenCache() *tokenCache {
	return &tokenCache{}
}

func (c *tokenCache) set(token string, expiresInSeconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	// Buffer expiration by 10% to trigger proactive refresh early
	safeDuration := time.Duration(float64(expiresInSeconds)*0.9) * time.Second
	c.expiresAt = time.Now().Add(safeDuration)
}

func (c *tokenCache) get() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token == "" || time.Now().After(c.expiresAt) {
		return "", false
	}
	return c.token, true
}

func (c *tokenCache) timeUntilExpiry() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token == "" {
		return 0
	}

	d := time.Until(c.expiresAt)
	if d < 0 {
		return 0
	}
	return d
}
