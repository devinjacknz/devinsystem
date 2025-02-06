package exchange

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJupiterDEX_RateLimiting(t *testing.T) {
	jupiter := NewJupiterDEX()
	ctx := context.Background()

	start := time.Now()
	_, _ = jupiter.GetMarketPrice(ctx, "SOL")
	_, _ = jupiter.GetMarketPrice(ctx, "SOL")
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(1),
		"Rate limiting should enforce 1 second delay between requests")
}

func TestJupiterDEX_GetQuote(t *testing.T) {
	jupiter := NewJupiterDEX()
	ctx := context.Background()

	quote, err := jupiter.GetQuote(ctx, "SOL", "USDC", "1000000000")
	if err != nil {
		t.Skip("Skipping test due to API error")
	}

	assert.NotNil(t, quote)
	assert.NotEmpty(t, quote.Price)
}

func TestJupiterDEX_GetMarketPrice(t *testing.T) {
	jupiter := NewJupiterDEX()
	ctx := context.Background()

	price, err := jupiter.GetMarketPrice(ctx, "SOL")
	if err != nil {
		t.Skip("Skipping test due to API error")
	}

	assert.Greater(t, price, float64(0))
}
