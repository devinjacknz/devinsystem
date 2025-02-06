package risk

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/devinjacknz/devinsystem/pkg/types"
)

func TestSlippageChecker_CheckSlippage(t *testing.T) {
	checker := NewSlippageChecker(0.02) // 2% max slippage

	tests := []struct {
		name        string
		order       types.Order
		marketPrice float64
		wantResult  bool
		wantErr     bool
	}{
		{
			name: "within slippage limit - buy",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.BuyOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 101.0, // 1% higher
			wantResult:  true,
			wantErr:     false,
		},
		{
			name: "within slippage limit - sell",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.SellOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 99.0, // 1% lower
			wantResult:  true,
			wantErr:     false,
		},
		{
			name: "exceeds slippage limit - buy",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.BuyOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 103.0, // 3% higher
			wantResult:  false,
			wantErr:     false,
		},
		{
			name: "exceeds slippage limit - sell",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.SellOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 97.0, // 3% lower
			wantResult:  false,
			wantErr:     false,
		},
		{
			name: "limit order - no slippage check",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.BuyOrder,
				Type:     types.LimitOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 105.0,
			wantResult:  true,
			wantErr:     false,
		},
		{
			name: "zero price",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.BuyOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    0.0,
			},
			marketPrice: 100.0,
			wantResult:  false,
			wantErr:     true,
		},
		{
			name: "zero market price",
			order: types.Order{
				Symbol:    "SOL/USD",
				Side:     types.BuyOrder,
				Type:     types.MarketOrder,
				Amount:   1.0,
				Price:    100.0,
			},
			marketPrice: 0.0,
			wantResult:  false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.CheckSlippage(tt.order, tt.marketPrice)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestSlippageChecker_CalculateSlippage(t *testing.T) {
	checker := NewSlippageChecker(0.02)

	tests := []struct {
		name        string
		orderPrice  float64
		marketPrice float64
		want        float64
		wantErr     bool
	}{
		{
			name:        "positive slippage",
			orderPrice:  100.0,
			marketPrice: 102.0,
			want:        0.02,
			wantErr:     false,
		},
		{
			name:        "negative slippage",
			orderPrice:  100.0,
			marketPrice: 98.0,
			want:        0.02,
			wantErr:     false,
		},
		{
			name:        "no slippage",
			orderPrice:  100.0,
			marketPrice: 100.0,
			want:        0.0,
			wantErr:     false,
		},
		{
			name:        "zero order price",
			orderPrice:  0.0,
			marketPrice: 100.0,
			want:        0.0,
			wantErr:     true,
		},
		{
			name:        "zero market price",
			orderPrice:  100.0,
			marketPrice: 0.0,
			want:        0.0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slippage, err := checker.calculateSlippage(tt.orderPrice, tt.marketPrice)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.want, slippage, 0.0001)
			}
		})
	}
}
