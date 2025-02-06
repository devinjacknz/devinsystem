package ai

import (
)

type Analysis struct {
	Symbol     string
	Trend      string
	Confidence float64
	Signals    []Signal
}

type Signal struct {
	Type       string
	Symbol     string
	Action     string
	Confidence float64
}

type AIService struct {
	ollamaURL     string
	deepseekModel string
}

func NewService(ollamaURL, deepseekModel string) *AIService {
	return &AIService{
		ollamaURL:     ollamaURL,
		deepseekModel: deepseekModel,
	}
}

func (s *AIService) AnalyzeMarket(data MarketData) (*Analysis, error) {
	return &Analysis{
		Symbol:     data.Symbol,
		Trend:      "NEUTRAL",
		Confidence: 0.8,
		Signals:    []Signal{},
	}, nil
}
