package exchange

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestJupiterDEX_GetMarketData_MultipleTokens(t *testing.T) {
	dex := NewJupiterDEX()
	data, err := dex.GetMarketData()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	
	// Verify rate limiting
	start := time.Now()
	data, err = dex.GetMarketData()
	assert.NoError(t, err)
	assert.True(t, time.Since(start) >= time.Second)

	// Verify token data
	for _, d := range data {
		assert.NotEmpty(t, d.Symbol)
		assert.Greater(t, d.Price, float64(0))
		assert.GreaterOrEqual(t, d.Volume, float64(0))
	}
}

func TestJupiterDEX_GetMarketPrice(t *testing.T) {
	dex := NewJupiterDEX()
	price, err := dex.GetMarketPrice("SOL/USDC")
	assert.NoError(t, err)
	assert.Greater(t, price, float64(0))
}

func TestJupiterDEX_Name(t *testing.T) {
	dex := NewJupiterDEX()
	assert.Equal(t, "Jupiter", dex.Name())
}
