package utils

import (
	"sync"
	"time"
)

type TokenInfo struct {
	Symbol    string
	Price     float64
	Volume    float64
	UpdatedAt time.Time
}

type TokenCache struct {
	mu      sync.RWMutex
	tokens  map[string]*TokenInfo
	expires time.Time
}

func NewTokenCache(ttl time.Duration) *TokenCache {
	return &TokenCache{
		tokens:  make(map[string]*TokenInfo),
		expires: time.Now().Add(ttl),
	}
}

func (c *TokenCache) Get(token string) (*TokenInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().After(c.expires) {
		return nil, false
	}

	info, ok := c.tokens[token]
	return info, ok
}

func (c *TokenCache) Set(token string, info *TokenInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens[token] = info
}
