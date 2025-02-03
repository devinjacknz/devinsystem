package ai

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDeepSeekClient struct {
	mock.Mock
}

func (m *MockDeepSeekClient) GeneratePrediction(data string) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}

func TestDeepSeekService_AnalyzeRisk(t *testing.T) {
	mockClient := new(MockDeepSeekClient)
	service := NewDeepSeekService("http://localhost:8080")
	service.client = mockClient

	tests := []struct {
		name       string
		portfolio  Portfolio
		mockResult string
		mockErr    error
		wantErr    bool
		wantRisk   RiskAnalysis
	}{
		{
			name: "low risk portfolio",
			portfolio: Portfolio{
				TotalValue:    100000.0,
				Positions: []Position{
					{Symbol: "SOL/USD", Amount: 100.0, Value: 10000.0},
					{Symbol: "BTC/USD", Amount: 0.5, Value: 20000.0},
				},
				DailyVolume: 50000.0,
			},
			mockResult: `{"risk_level": "low", "score": 0.2, "recommendations": ["maintain current positions"]}`,
			mockErr:    nil,
			wantErr:    false,
			wantRisk: RiskAnalysis{
				RiskLevel:       "low",
				Score:           0.2,
				Recommendations: []string{"maintain current positions"},
			},
		},
		{
			name: "high risk portfolio",
			portfolio: Portfolio{
				TotalValue:    200000.0,
				Positions: []Position{
					{Symbol: "PEPE/USD", Amount: 1000000.0, Value: 150000.0},
				},
				DailyVolume: 180000.0,
			},
			mockResult: `{"risk_level": "high", "score": 0.8, "recommendations": ["reduce exposure", "diversify portfolio"]}`,
			mockErr:    nil,
			wantErr:    false,
			wantRisk: RiskAnalysis{
				RiskLevel:       "high",
				Score:           0.8,
				Recommendations: []string{"reduce exposure", "diversify portfolio"},
			},
		},
		{
			name: "api error",
			portfolio: Portfolio{
				TotalValue: 100000.0,
				Positions:  []Position{},
			},
			mockResult: "",
			mockErr:    ErrAPIError,
			wantErr:    true,
			wantRisk:   RiskAnalysis{},
		},
		{
			name: "invalid response format",
			portfolio: Portfolio{
				TotalValue: 100000.0,
				Positions:  []Position{},
			},
			mockResult: "invalid json",
			mockErr:    nil,
			wantErr:    true,
			wantRisk:   RiskAnalysis{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := formatPortfolioData(tt.portfolio)
			mockClient.On("GeneratePrediction", data).Return(tt.mockResult, tt.mockErr).Once()

			risk, err := service.AnalyzeRisk(tt.portfolio)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRisk.RiskLevel, risk.RiskLevel)
				assert.Equal(t, tt.wantRisk.Score, risk.Score)
				assert.Equal(t, tt.wantRisk.Recommendations, risk.Recommendations)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestDeepSeekService_ValidateResponse(t *testing.T) {
	service := NewDeepSeekService("http://localhost:8080")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json response",
			input:   `{"risk_level": "medium", "score": 0.5, "recommendations": ["diversify"]}`,
			wantErr: false,
		},
		{
			name:    "invalid json format",
			input:   "invalid json",
			wantErr: true,
		},
		{
			name:    "missing required fields",
			input:   `{"risk_level": "low"}`,
			wantErr: true,
		},
		{
			name:    "invalid risk level",
			input:   `{"risk_level": "invalid", "score": 0.5, "recommendations": []}`,
			wantErr: true,
		},
		{
			name:    "invalid score range",
			input:   `{"risk_level": "low", "score": 1.5, "recommendations": []}`,
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
