package exchange

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSolanaDEX_AddMarket(t *testing.T) {
	dex := NewSolanaDEX()

	tests := []struct {
		name          string
		symbol        string
		baseDecimals  uint8
		quoteDecimals uint8
		wantErr       bool
	}{
		{
			name:          "add new market",
			symbol:        "SOL/USD",
			baseDecimals:  9,
			quoteDecimals: 6,
			wantErr:       false,
		},
		{
			name:          "add duplicate market",
			symbol:        "SOL/USD",
			baseDecimals:  9,
			quoteDecimals: 6,
			wantErr:       true,
		},
		{
			name:          "add market with different decimals",
			symbol:        "BTC/USD",
			baseDecimals:  8,
			quoteDecimals: 6,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dex.AddMarket(tt.symbol, tt.baseDecimals, tt.quoteDecimals)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				market, exists := dex.markets[tt.symbol]
				assert.True(t, exists)
				assert.Equal(t, tt.symbol, market.Symbol)
				assert.Equal(t, tt.baseDecimals, market.BaseDecimals)
				assert.Equal(t, tt.quoteDecimals, market.QuoteDecimals)
				assert.Empty(t, market.OrderBook.Bids)
				assert.Empty(t, market.OrderBook.Asks)
			}
		})
	}
}

func TestSolanaDEX_UpdateOrderBook(t *testing.T) {
	dex := NewSolanaDEX()
	err := dex.AddMarket("SOL/USD", 9, 6)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		symbol  string
		bids    []PriceLevel
		asks    []PriceLevel
		wantErr bool
	}{
		{
			name:   "update existing market",
			symbol: "SOL/USD",
			bids: []PriceLevel{
				{Price: 100.0, Size: 10.0, Orders: 5},
				{Price: 99.5, Size: 20.0, Orders: 8},
			},
			asks: []PriceLevel{
				{Price: 101.0, Size: 15.0, Orders: 6},
				{Price: 101.5, Size: 25.0, Orders: 10},
			},
			wantErr: false,
		},
		{
			name:    "update non-existent market",
			symbol:  "BTC/USD",
			bids:    []PriceLevel{},
			asks:    []PriceLevel{},
			wantErr: true,
		},
		{
			name:   "update with empty order book",
			symbol: "SOL/USD",
			bids:   []PriceLevel{},
			asks:   []PriceLevel{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dex.UpdateOrderBook(tt.symbol, tt.bids, tt.asks)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				market := dex.markets[tt.symbol]
				assert.Equal(t, tt.bids, market.OrderBook.Bids)
				assert.Equal(t, tt.asks, market.OrderBook.Asks)
			}
		})
	}
}

func TestSolanaDEX_GetMarketPrice(t *testing.T) {
	dex := NewSolanaDEX()
	err := dex.AddMarket("SOL/USD", 9, 6)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		symbol    string
		setupBook func()
		wantPrice float64
		wantErr   bool
	}{
		{
			name:   "get price from non-empty order book",
			symbol: "SOL/USD",
			setupBook: func() {
				dex.UpdateOrderBook("SOL/USD",
					[]PriceLevel{{Price: 100.0, Size: 10.0, Orders: 5}},
					[]PriceLevel{{Price: 101.0, Size: 15.0, Orders: 6}},
				)
			},
			wantPrice: 101.0,
			wantErr:   false,
		},
		{
			name:      "get price from empty order book",
			symbol:    "SOL/USD",
			setupBook: func() {
				dex.UpdateOrderBook("SOL/USD", []PriceLevel{}, []PriceLevel{})
			},
			wantPrice: 0,
			wantErr:   true,
		},
		{
			name:      "get price from non-existent market",
			symbol:    "BTC/USD",
			setupBook: func() {},
			wantPrice: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupBook()
			price, err := dex.GetMarketPrice(tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPrice, price)
			}
		})
	}
}
