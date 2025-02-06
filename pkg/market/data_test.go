package market

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeliusClient_RateLimiting(t *testing.T) {
	client := NewHeliusClient("https://mainnet.helius-rpc.com/?api-key=test")

	start := time.Now()
	_, _ = client.GetMarketData(context.Background(), "TEST")
	_, _ = client.GetMarketData(context.Background(), "TEST")

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(1), "Rate limiting should enforce 1 second delay")
}

func TestHeliusClient_TokenList(t *testing.T) {
	client := NewHeliusClient("https://mainnet.helius-rpc.com/?api-key=test")

	tokens, err := client.GetTokenList(context.Background())
	if err != nil {
		t.Skip("Skipping test due to RPC error")
	}

	assert.NotEmpty(t, tokens)
	for _, token := range tokens {
		assert.NotEmpty(t, token)
	}
}

func TestHeliusClient_GetMarketData(t *testing.T) {
	client := NewHeliusClient("https://mainnet.helius-rpc.com/?api-key=test")

	data, err := client.GetMarketData(context.Background(), "TEST")
	if err != nil {
		t.Skip("Skipping test due to RPC error")
	}

	assert.NotNil(t, data)
	assert.NotEmpty(t, data.Symbol)
	assert.Greater(t, data.Price, float64(0))
	assert.GreaterOrEqual(t, data.Volume, float64(0))
	assert.False(t, data.Timestamp.IsZero())
}
