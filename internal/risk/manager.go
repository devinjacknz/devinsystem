package risk

import (
	"fmt"
	"sync"

	"github.com/devinjacknz/devintrade/internal/ai"
)

type Order struct {
	Symbol       string
	Side         string
	Amount       float64
	Price        float64
	OrderType    string
}

type Manager interface {
	ValidateOrder(order Order) error
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

func NewRiskManager(aiService ai.Service, maxExposure float64) *RiskManager {
	return &RiskManager{
		stopLoss:    NewStopLoss(),
		slippage:    NewSlippageProtection(50), // 0.5% default max slippage
		aiService:   aiService,
		exposures:   make(map[string]float64),
		maxExposure: maxExposure,
	}
}

func (rm *RiskManager) ValidateOrder(order Order) error {
	// Check AI risk analysis
	riskAnalysis, err := rm.aiService.AnalyzeRisk(ai.MarketData{
		Symbol: order.Symbol,
		Price:  order.Price,
	})
	if err != nil {
		return fmt.Errorf("failed to analyze risk: %w", err)
	}

	// Set stop loss based on AI recommendation
	if err := rm.stopLoss.SetStopLoss(order.Symbol, riskAnalysis.StopLossPrice); err != nil {
		return fmt.Errorf("failed to set stop loss: %w", err)
	}

	// Check exposure
	exposure, err := rm.CheckExposure(order.Symbol)
	if err != nil {
		return fmt.Errorf("failed to check exposure: %w", err)
	}

	if exposure+order.Amount > rm.maxExposure {
		return fmt.Errorf("order would exceed maximum exposure")
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
