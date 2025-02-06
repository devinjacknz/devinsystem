package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenCache_TTL(t *testing.T) {
	updateFunc := func(ctx context.Context, token string) (*TokenInfo, error) {
		return &TokenInfo{
			Symbol:    token,
			Price:     100.0,
			Volume:    1000.0,
			UpdatedAt: time.Now(),
		}, nil
	}

	cache := NewTokenCache(100*time.Millisecond, 30, updateFunc)

	// Initial get should trigger update
	info, err := cache.Get(context.Background(), "TEST")
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Get within TTL should return cached value
	info2, err := cache.Get(context.Background(), "TEST")
	assert.NoError(t, err)
	assert.Equal(t, info, info2)

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Get after TTL should trigger update
	info3, err := cache.Get(context.Background(), "TEST")
	assert.NoError(t, err)
	assert.NotEqual(t, info.UpdatedAt, info3.UpdatedAt)
}

func TestTokenCache_GetTopTokens(t *testing.T) {
	updateFunc := func(ctx context.Context, token string) (*TokenInfo, error) {
		volumes := map[string]float64{
			"TEST1": 1000.0,
			"TEST2": 2000.0,
			"TEST3": 500.0,
		}
		return &TokenInfo{
			Symbol:    token,
			Price:     100.0,
			Volume:    volumes[token],
			UpdatedAt: time.Now(),
		}, nil
	}

	cache := NewTokenCache(time.Hour, 2, updateFunc)

	// Add test tokens
	cache.Set("TEST1", &TokenInfo{Symbol: "TEST1", Volume: 1000.0})
	cache.Set("TEST2", &TokenInfo{Symbol: "TEST2", Volume: 2000.0})
	cache.Set("TEST3", &TokenInfo{Symbol: "TEST3", Volume: 500.0})

	// Get top 2 tokens
	tokens, err := cache.GetTopTokens(context.Background())
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "TEST2", tokens[0].Symbol)
	assert.Equal(t, "TEST1", tokens[1].Symbol)
}
