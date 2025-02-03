package ai

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOllamaClient struct {
	mock.Mock
}

func (m *MockOllamaClient) GenerateAnalysis(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}

func TestOllamaService_AnalyzeMarket(t *testing.T) {
	mockClient := new(MockOllamaClient)
	service := NewOllamaService("http://localhost:11434")
	service.client = mockClient

	tests := []struct {
		name        string
		marketData  MarketData
		mockResult  string
		mockErr     error
		wantErr     bool
		wantSignal  TradingSignal
	}{
		{
			name: "bullish analysis",
			marketData: MarketData{
				Symbol:    "SOL/USD",
				Price:     100.0,
				Volume:    1000000.0,
				BidDepth:  500000.0,
				AskDepth:  400000.0,
			},
			mockResult: `{"signal": "buy", "confidence": 0.85, "reason": "Strong buying pressure with high volume"}`,
			mockErr:    nil,
			wantErr:    false,
			wantSignal: TradingSignal{
				Signal:     "buy",
				Confidence: 0.85,
				Reason:     "Strong buying pressure with high volume",
			},
		},
		{
			name: "bearish analysis",
			marketData: MarketData{
				Symbol:    "SOL/USD",
				Price:     95.0,
				Volume:    800000.0,
				BidDepth:  300000.0,
				AskDepth:  600000.0,
			},
			mockResult: `{"signal": "sell", "confidence": 0.75, "reason": "Increasing sell pressure"}`,
			mockErr:    nil,
			wantErr:    false,
			wantSignal: TradingSignal{
				Signal:     "sell",
				Confidence: 0.75,
				Reason:     "Increasing sell pressure",
			},
		},
		{
			name: "api error",
			marketData: MarketData{
				Symbol: "SOL/USD",
				Price:  100.0,
			},
			mockResult: "",
			mockErr:    ErrAPIError,
			wantErr:    true,
			wantSignal: TradingSignal{},
		},
		{
			name: "invalid response format",
			marketData: MarketData{
				Symbol: "SOL/USD",
				Price:  100.0,
			},
			mockResult: "invalid json",
			mockErr:    nil,
			wantErr:    true,
			wantSignal: TradingSignal{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := formatMarketDataPrompt(tt.marketData)
			mockClient.On("GenerateAnalysis", prompt).Return(tt.mockResult, tt.mockErr).Once()

			signal, err := service.AnalyzeMarket(tt.marketData)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSignal.Signal, signal.Signal)
				assert.Equal(t, tt.wantSignal.Confidence, signal.Confidence)
				assert.Equal(t, tt.wantSignal.Reason, signal.Reason)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestOllamaService_ValidateResponse(t *testing.T) {
	service := NewOllamaService("http://localhost:11434")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json response",
			input:   `{"signal": "buy", "confidence": 0.85, "reason": "Strong market"}`,
			wantErr: false,
		},
		{
			name:    "invalid json format",
			input:   "invalid json",
			wantErr: true,
		},
		{
			name:    "missing required fields",
			input:   `{"signal": "buy"}`,
			wantErr: true,
		},
		{
			name:    "invalid signal value",
			input:   `{"signal": "invalid", "confidence": 0.85, "reason": "test"}`,
			wantErr: true,
		},
		{
			name:    "invalid confidence range",
			input:   `{"signal": "buy", "confidence": 1.5, "reason": "test"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.validateResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
