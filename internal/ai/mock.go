package ai

type MockService struct{}

func (m *MockService) AnalyzeMarket(data MarketData) (*Analysis, error) {
	return &Analysis{
		Symbol:     data.Symbol,
		Trend:      "neutral",
		Confidence: 0.5,
		Signals:    []Signal{
			{
				Type:       "price",
				Symbol:     data.Symbol,
				Action:     "hold",
				Confidence: 0.5,
			},
		},
	}, nil
}

func (m *MockService) AnalyzeRisk(data MarketData) (*RiskAnalysis, error) {
	return &RiskAnalysis{
		Symbol:        data.Symbol,
		StopLossPrice: data.Price * 0.95,
		RiskLevel:     "LOW",
		Confidence:    0.95,
	}, nil
}
