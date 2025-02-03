package ai

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOllamaClient struct {
	mock.Mock
}

func (m *MockOllamaClient) AnalyzeMarket(data MarketData) (*Analysis, error) {
	args := m.Called(data)
	return args.Get(0).(*Analysis), args.Error(1)
}

func (m *MockOllamaClient) AnalyzeRisk(data MarketData) (*RiskAnalysis, error) {
	args := m.Called(data)
	return args.Get(0).(*RiskAnalysis), args.Error(1)
}

type MockDeepSeekClient struct {
	mock.Mock
}

func (m *MockDeepSeekClient) AnalyzeMarket(data MarketData) (*Analysis, error) {
	args := m.Called(data)
	return args.Get(0).(*Analysis), args.Error(1)
}

func (m *MockDeepSeekClient) AnalyzeRisk(data MarketData) (*RiskAnalysis, error) {
	args := m.Called(data)
	return args.Get(0).(*RiskAnalysis), args.Error(1)
}

func TestAIService_AnalyzeMarket(t *testing.T) {
	mockOllama := new(MockOllamaClient)
	mockDeepSeek := new(MockDeepSeekClient)
	service := &AIService{
		ollamaClient:   mockOllama,
		deepseekClient: mockDeepSeek,
	}
	marketData := MarketData{
		Symbol:    "SOL/USD",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: 1234567890,
	}

	tests := []struct {
		name      string
		data      MarketData
		setupMock func()
		wantErr   bool
	}{
		{
			name: "successful analysis",
			data: marketData,
			setupMock: func() {
				analysis := &Analysis{
					Symbol:       "SOL/USD",
					Sentiment:    "bullish",
					Confidence:   0.85,
					PriceTarget: 120.0,
					RiskScore:   0.3,
					Timestamp:   1234567890,
				}
				mockOllama.On("AnalyzeMarket", marketData).Return(analysis, nil)
				mockDeepSeek.On("AnalyzeMarket", marketData).Return(analysis, nil)
			},
			wantErr: false,
		},
		{
			name: "ollama failure",
			data: marketData,
			setupMock: func() {
				mockOllama.On("AnalyzeMarket", marketData).Return((*Analysis)(nil), assert.AnError)
				analysis := &Analysis{
					Symbol:       "SOL/USD",
					Sentiment:    "neutral",
					Confidence:   0.6,
					PriceTarget:  110.0,
					RiskScore:    0.5,
					Timestamp:    1234567890,
				}
				mockDeepSeek.On("AnalyzeMarket", marketData).Return(analysis, nil)
			},
			wantErr: false,
		},
		{
			name: "both clients fail",
			data: marketData,
			setupMock: func() {
				mockOllama.On("AnalyzeMarket", marketData).Return((*Analysis)(nil), assert.AnError)
				mockDeepSeek.On("AnalyzeMarket", marketData).Return((*Analysis)(nil), assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			analysis, err := service.AnalyzeMarket(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, analysis)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, analysis)
			}
		})
	}
}

func TestAIService_AnalyzeRisk(t *testing.T) {
	marketData := MarketData{
		Symbol:    "SOL/USD",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: 1234567890,
	}

	tests := []struct {
		name      string
		data      MarketData
		setupMock func()
		wantErr   bool
	}{
		{
			name: "successful risk analysis",
			data: marketData,
			setupMock: func() {
				analysis := &RiskAnalysis{
					Symbol:        "SOL/USD",
					RiskLevel:     "low",
					StopLossPrice: 95.0,
					Confidence:    0.85,
					MaxExposure:   1000.0,
					Timestamp:     1234567890,
				}
				mockDeepSeek.On("AnalyzeRisk", marketData).Return(analysis, nil)
			},
			wantErr: false,
		},
		{
			name: "failed risk analysis",
			data: marketData,
			setupMock: func() {
				service.deepseekClient = &MockDeepSeekClient{}
				service.deepseekClient.(*MockDeepSeekClient).On("AnalyzeRisk", marketData).Return((*RiskAnalysis)(nil), assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			analysis, err := service.AnalyzeRisk(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, analysis)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, analysis)
				assert.Equal(t, tt.data.Symbol, analysis.Symbol)
			}
		})
	}
}
