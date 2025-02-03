package risk

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/devinjacknz/devintrade/internal/ai"
)

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) AnalyzeMarket(data ai.MarketData) (*ai.Analysis, error) {
	args := m.Called(data)
	return args.Get(0).(*ai.Analysis), args.Error(1)
}

func (m *MockAIService) AnalyzeRisk(data ai.MarketData) (*ai.RiskAnalysis, error) {
	args := m.Called(data)
	return args.Get(0).(*ai.RiskAnalysis), args.Error(1)
}

func TestRiskManager_ValidateOrder(t *testing.T) {
	mockAI := new(MockAIService)
	manager := NewRiskManager(mockAI, 1000.0) // Max exposure of 1000 units

	tests := []struct {
		name      string
		order     Order
		setupMock func()
		wantErr   bool
	}{
		{
			name: "valid order",
			order: Order{
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1.0,
				Price:     100.0,
				OrderType: "limit",
			},
			setupMock: func() {
				mockAI.On("AnalyzeRisk", mock.Anything).Return(&ai.RiskAnalysis{
					Symbol:        "SOL/USD",
					RiskLevel:     "low",
					StopLossPrice: 95.0,
					Confidence:    0.3,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "slippage too high",
			order: Order{
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1.0,
				Price:     120.0,
				OrderType: "market",
			},
			setupMock: func() {
				mockAI.On("AnalyzeRisk", mock.Anything).Return(&ai.RiskAnalysis{
					Symbol:        "SOL/USD",
					RiskLevel:     "high",
					StopLossPrice: 115.0,
					Confidence:    0.8,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "stop loss triggered",
			order: Order{
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1.0,
				Price:     90.0,
				OrderType: "limit",
			},
			setupMock: func() {
				mockAI.On("AnalyzeRisk", mock.Anything).Return(&ai.RiskAnalysis{
					Symbol:        "SOL/USD",
					RiskLevel:     "medium",
					StopLossPrice: 85.0,
					Confidence:    0.6,
				}, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := manager.ValidateOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRiskManager_CheckExposure(t *testing.T) {
	mockAI := new(MockAIService)
	manager := NewRiskManager(mockAI, 1000.0)

	tests := []struct {
		name       string
		symbol     string
		positions  map[string]float64
		wantAmount float64
		wantErr    bool
	}{
		{
			name:       "existing position",
			symbol:     "SOL/USD",
			positions:  map[string]float64{"SOL/USD": 100.0},
			wantAmount: 100.0,
			wantErr:    false,
		},
		{
			name:       "no position",
			symbol:     "SOL/USD",
			positions:  map[string]float64{},
			wantAmount: 0.0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.exposures = tt.positions
			amount, err := manager.CheckExposure(tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAmount, amount)
			}
		})
	}
}

func TestRiskManager_UpdateStopLoss(t *testing.T) {
	mockAI := new(MockAIService)
	manager := NewRiskManager(mockAI, 1000.0)

	tests := []struct {
		name      string
		symbol    string
		price     float64
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "successful update",
			symbol: "SOL/USD",
			price:  100.0,
			setupMock: func() {
				mockAI.On("AnalyzeRisk", mock.Anything).Return(&ai.Analysis{
					StopLossPrice: 95.0,
					Risk:          0.3,
				}, nil)
			},
			wantErr: false,
		},
		{
			name:   "update failure",
			symbol: "SOL/USD",
			price:  100.0,
			setupMock: func() {
				mockAI.On("AnalyzeRisk", mock.Anything).Return(&ai.Analysis{
					StopLossPrice: 95.0,
					Risk:          0.3,
				}, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := manager.UpdateStopLoss(tt.symbol, tt.price)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
