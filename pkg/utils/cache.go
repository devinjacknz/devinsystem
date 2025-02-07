package utils

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
	"golang.org/x/time/rate"
)

type TokenInfo struct {
	Symbol    string
	Price     float64
	Volume    float64
	UpdatedAt time.Time
}

type TokenCache struct {
	mu          sync.RWMutex
	tokens      map[string]*TokenInfo
	expires     time.Time
	ttl         time.Duration
	limiter     *rate.Limiter
	maxTokens   int
	updateFunc  func(context.Context, string) (*TokenInfo, error)
}

func NewTokenCache(ttl time.Duration, maxTokens int, updateFunc func(context.Context, string) (*TokenInfo, error)) *TokenCache {
	return &TokenCache{
		tokens:     make(map[string]*TokenInfo),
		expires:    time.Now().Add(ttl),
		ttl:        ttl,
		limiter:    rate.NewLimiter(rate.Every(time.Minute), 60),
		maxTokens:  maxTokens,
		updateFunc: updateFunc,
	}
}

func (c *TokenCache) Get(ctx context.Context, token string) (*TokenInfo, error) {
	c.mu.RLock()
	info, exists := c.tokens[token]
	expired := time.Now().After(c.expires)
	c.mu.RUnlock()

	if !exists || expired {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}

		info, err := c.updateFunc(ctx, token)
		if err != nil {
			return nil, err
		}

		c.Set(token, info)
		return info, nil
	}

	return info, nil
}

func (c *TokenCache) Set(token string, info *TokenInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens[token] = info
}

func (c *TokenCache) GetTopTokens(ctx context.Context) ([]string, error) {
	start := time.Now()
	defer func() {
		log.Printf("%s Token cache operation took %v", logging.LogMarkerPerf, time.Since(start))
	}()

	log.Printf("%s Checking token cache status", logging.LogMarkerSystem)
	c.mu.RLock()
	expired := time.Now().After(c.expires)
	c.mu.RUnlock()

	if expired {
		log.Printf("%s Token cache expired, refreshing...", logging.LogMarkerSystem)
		if err := c.Refresh(ctx); err != nil {
			log.Printf("%s Failed to refresh token cache: %v", logging.LogMarkerError, err)
			return nil, err
		}
		log.Printf("%s Token cache refreshed successfully", logging.LogMarkerSystem)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	tokens := make([]string, 0, len(c.tokens))
	for symbol := range c.tokens {
		tokens = append(tokens, symbol)
	}

	if len(tokens) > c.maxTokens {
		tokens = tokens[:c.maxTokens]
	}

	return tokens, nil
}

func (c *TokenCache) Refresh(ctx context.Context) error {
	start := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() {
		log.Printf("%s Token cache refresh took %v", logging.LogMarkerPerf, time.Since(start))
	}()

	if !time.Now().After(c.expires) {
		log.Printf("%s Token cache still valid, skipping refresh", logging.LogMarkerSystem)
		return nil
	}
	
	log.Printf("%s Starting token cache refresh", logging.LogMarkerSystem)

	for token := range c.tokens {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}

		info, err := c.updateFunc(ctx, token)
		if err != nil {
			continue
		}
		c.tokens[token] = info
	}

	c.expires = time.Now().Add(c.ttl)
	return nil
}
