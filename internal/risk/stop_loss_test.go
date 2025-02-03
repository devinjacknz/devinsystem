package risk

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestStopLossManager_SetStopLoss(t *testing.T) {
	manager := NewStopLossManager()

	tests := []struct {
		name    string
		symbol  string
		price   float64
		wantErr bool
	}{
		{
			name:    "valid stop loss",
			symbol:  "SOL/USD",
			price:   95.0,
			wantErr: false,
		},
		{
			name:    "zero price",
			symbol:  "SOL/USD",
			price:   0.0,
			wantErr: true,
		},
		{
			name:    "negative price",
			symbol:  "SOL/USD",
			price:   -10.0,
			wantErr: true,
		},
		{
			name:    "empty symbol",
			symbol:  "",
			price:   95.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SetStopLoss(tt.symbol, tt.price)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				price, exists := manager.stopLosses[tt.symbol]
				assert.True(t, exists)
				assert.Equal(t, tt.price, price)
			}
		})
	}
}

func TestStopLossManager_CheckStopLoss(t *testing.T) {
	manager := NewStopLossManager()
	
	// Set up initial stop losses
	manager.stopLosses = map[string]float64{
		"SOL/USD": 95.0,
		"BTC/USD": 40000.0,
	}

	tests := []struct {
		name        string
		symbol      string
		price       float64
		wantTripped bool
		wantErr     bool
	}{
		{
			name:        "above stop loss",
			symbol:      "SOL/USD",
			price:       100.0,
			wantTripped: false,
			wantErr:     false,
		},
		{
			name:        "at stop loss",
			symbol:      "SOL/USD",
			price:       95.0,
			wantTripped: true,
			wantErr:     false,
		},
		{
			name:        "below stop loss",
			symbol:      "SOL/USD",
			price:       90.0,
			wantTripped: true,
			wantErr:     false,
		},
		{
			name:        "no stop loss set",
			symbol:      "ETH/USD",
			price:       3000.0,
			wantTripped: false,
			wantErr:     false,
		},
		{
			name:        "zero price",
			symbol:      "SOL/USD",
			price:       0.0,
			wantTripped: false,
			wantErr:     true,
		},
		{
			name:        "empty symbol",
			symbol:      "",
			price:       100.0,
			wantTripped: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tripped, err := manager.CheckStopLoss(tt.symbol, tt.price)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTripped, tripped)
			}
		})
	}
}

func TestStopLossManager_RemoveStopLoss(t *testing.T) {
	manager := NewStopLossManager()
	
	// Set up initial stop losses
	manager.stopLosses = map[string]float64{
		"SOL/USD": 95.0,
		"BTC/USD": 40000.0,
	}

	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{
			name:    "existing stop loss",
			symbol:  "SOL/USD",
			wantErr: false,
		},
		{
			name:    "non-existent stop loss",
			symbol:  "ETH/USD",
			wantErr: false,
		},
		{
			name:    "empty symbol",
			symbol:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RemoveStopLoss(tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				_, exists := manager.stopLosses[tt.symbol]
				assert.False(t, exists)
			}
		})
	}
}
