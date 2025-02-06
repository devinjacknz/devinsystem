package exchange

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestJupiterDEX_GetMarketData(t *testing.T) {
	dex := NewJupiterDEX()
	data, err := dex.GetMarketData()
	assert.NoError(t, err)
	assert.NotNil(t, data)
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
