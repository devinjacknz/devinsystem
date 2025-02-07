package risk

import (
	"context"
	"fmt"
	"sync"

	"github.com/devinjacknz/devinsystem/internal/ai"
)

type Order struct {
	Symbol       string
	Side         string
	Amount       float64
	Price        float64
	OrderType    string
}

type Trade struct {
	Token     string
	Amount    float64
	Direction string
	Price     float64
}

type Manager interface {
	ValidateTrade(ctx context.Context, trade *Trade) error
	CheckExposure(symbol string) (float64, error)
	UpdateStopLoss(symbol string, currentPrice float64) error
}

type RiskManager struct {
	mu          sync.RWMutex
	stopLoss    *StopLoss
	slippage    *SlippageProtection
	aiService   ai.Service
	exposures   map[string]float64
	maxExposure float64
}

func NewManager() *RiskManager {
	return &RiskManager{
		stopLoss:    NewStopLoss(),
		slippage:    NewSlippageProtection(50),
		aiService:   &ai.MockService{},
		exposures:   make(map[string]float64),
		maxExposure: 1000000, // 1M default max exposure
	}
}

func NewRiskManager(aiService ai.Service, maxExposure float64) *RiskManager {
	return &RiskManager{
		stopLoss:    NewStopLoss(),
		slippage:    NewSlippageProtection(50),
		aiService:   aiService,
		exposures:   make(map[string]float64),
		maxExposure: maxExposure,
	}
}

func (rm *RiskManager) ValidateTrade(ctx context.Context, trade *Trade) error {
	// Check AI risk analysis
	riskAnalysis, err := rm.aiService.AnalyzeRisk(ai.MarketData{
		Symbol: trade.Token,
		Price:  trade.Price,
	})
	if err != nil {
		return fmt.Errorf("failed to analyze risk: %w", err)
	}

	// Set stop loss based on AI recommendation
	if err := rm.stopLoss.SetStopLoss(trade.Token, riskAnalysis.StopLossPrice); err != nil {
		return fmt.Errorf("failed to set stop loss: %w", err)
	}

	// Check exposure
	exposure, err := rm.CheckExposure(trade.Token)
	if err != nil {
		return fmt.Errorf("failed to check exposure: %w", err)
	}

	if exposure+trade.Amount > rm.maxExposure {
		return fmt.Errorf("trade would exceed maximum exposure")
	}

	return nil
}

func (rm *RiskManager) CheckExposure(symbol string) (float64, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.exposures[symbol], nil
}

func (rm *RiskManager) UpdateExposure(symbol string, amount float64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.exposures[symbol] = amount
	return nil
}

func (rm *RiskManager) UpdateStopLoss(symbol string, currentPrice float64) error {
	return rm.stopLoss.UpdateTrailingStop(symbol, currentPrice)
}
