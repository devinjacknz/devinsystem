package ai

import "time"

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

type Analysis struct {
	Symbol     string
	Action     string
	Confidence float64
	Reasoning  string
	Model      string
	Timestamp  time.Time
	Signals    []Signal
}

type Signal struct {
	Type       string
	Symbol     string
	Action     string
	Confidence float64
}

type Service interface {
	AnalyzeMarket(data MarketData) (*Analysis, error)
	AnalyzeRisk(data MarketData) (*RiskAnalysis, error)
}
