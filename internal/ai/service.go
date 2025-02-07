package ai

import (
	"sync"
)

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
