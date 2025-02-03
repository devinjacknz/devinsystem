package ai

import (
	"sync"
)

type MarketData struct {
	Symbol string
	Price  float64
	Volume float64
	Trend  string
}

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

type RiskAnalysis struct {
	Symbol        string
	RiskLevel     string
	StopLossPrice float64
	Confidence    float64
}

type Service interface {
	AnalyzeMarket(data MarketData) (*Analysis, error)
	AnalyzeRisk(data MarketData) (*RiskAnalysis, error)
}

type AIService struct {
	mu            sync.RWMutex
	ollamaClient  *OllamaClient
	deepseekClient *DeepSeekClient
}

func NewAIService(ollamaEndpoint, ollamaModel, deepseekEndpoint, deepseekKey string) *AIService {
	return &AIService{
		ollamaClient:   NewOllamaClient(ollamaEndpoint, ollamaModel),
		deepseekClient: NewDeepSeekClient(deepseekEndpoint, deepseekKey),
	}
}

func (s *AIService) AnalyzeMarket(data MarketData) (*Analysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ollamaClient.AnalyzeMarket(data)
}

func (s *AIService) AnalyzeRisk(data MarketData) (*RiskAnalysis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deepseekClient.AnalyzeRisk(data)
}
