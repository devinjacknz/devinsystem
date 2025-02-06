package trading

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRiskManager_ValidateTrade(t *testing.T) {
	riskMgr := NewRiskManager()

	tests := []struct {
		name      string
		trade     *Trade
		wantError bool
	}{
		{
			name: "Valid trade within limits",
			trade: &Trade{
				Token:     "TEST",
				Amount:    1.0,
				Direction: "BUY",
				Price:     100.0,
			},
			wantError: false,
		},
		{
			name: "Trade exceeds max exposure",
			trade: &Trade{
				Token:     "TEST",
				Amount:    50000.0,
				Direction: "BUY",
				Price:     100000.0,
			},
			wantError: true,
		},
		{
			name: "Trade exceeds per-token limit",
			trade: &Trade{
				Token:     "TEST",
				Amount:    20000.0,
				Direction: "BUY",
				Price:     100000.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := riskMgr.ValidateTrade(context.Background(), tt.trade)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRiskManager_GetCurrentRisk(t *testing.T) {
	riskMgr := NewRiskManager()
	ctx := context.Background()

	// Add some risk
	trade := &Trade{
		Token:     "TEST",
		Amount:    1.0,
		Direction: "BUY",
		Price:     100.0,
	}
	err := riskMgr.ValidateTrade(ctx, trade)
	assert.NoError(t, err)

	risk := riskMgr.GetCurrentRisk("TEST")
	assert.Equal(t, 100.0, risk)
}

func TestRiskManager_GetTotalRisk(t *testing.T) {
	riskMgr := NewRiskManager()
	ctx := context.Background()

	// Add risks for multiple tokens
	trades := []*Trade{
		{Token: "TEST1", Amount: 1.0, Direction: "BUY", Price: 100.0},
		{Token: "TEST2", Amount: 2.0, Direction: "BUY", Price: 200.0},
	}

	for _, trade := range trades {
		err := riskMgr.ValidateTrade(ctx, trade)
		assert.NoError(t, err)
	}

	totalRisk := riskMgr.GetTotalRisk()
	assert.Equal(t, 500.0, totalRisk)
}
