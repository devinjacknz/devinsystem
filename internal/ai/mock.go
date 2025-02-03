package ai

type MockService struct{}

func (m *MockService) AnalyzeRisk(data MarketData) (RiskAnalysis, error) {
	// Mock implementation that sets stop loss 5% below current price
	return RiskAnalysis{
		Symbol:        data.Symbol,
		StopLossPrice: data.Price * 0.95,
		RiskLevel:     "LOW",
		Confidence:    0.95,
	}, nil
}
