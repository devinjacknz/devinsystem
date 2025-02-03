package ai

type MarketData struct {
	Symbol string
	Price  float64
	Volume float64
	Trend  string
}

type RiskAnalysis struct {
	Symbol      string
	StopLossPrice float64
	RiskLevel     string
	Confidence    float64
}

type Service interface {
	AnalyzeRisk(data MarketData) (RiskAnalysis, error)
}
