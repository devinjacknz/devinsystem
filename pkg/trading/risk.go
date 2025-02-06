package trading

import (
	"context"
	"fmt"
	"sync"
)

type RiskManager struct {
	mu           sync.RWMutex
	maxExposure  float64
	stopLoss     float64
	slippage     float64
	currentRisks map[string]float64
	totalRisk    float64
}

func NewRiskManager() *RiskManager {
	return &RiskManager{
		maxExposure:  3_000_000, // 3M max exposure
		stopLoss:     0.50,      // 50% stop loss for meme coins
		slippage:     0.02,      // 2% slippage tolerance
		currentRisks: make(map[string]float64),
	}
}

func (r *RiskManager) ValidateTrade(ctx context.Context, trade *Trade) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Calculate trade value
	tradeValue := trade.Amount * trade.Price

	// Check if trade would exceed max exposure
	if r.totalRisk+tradeValue > r.maxExposure {
		return fmt.Errorf("trade would exceed max exposure of %.2f (current: %.2f, trade: %.2f)",
			r.maxExposure, r.totalRisk, tradeValue)
	}

	// Check token-specific exposure
	currentTokenRisk := r.currentRisks[trade.Token]
	if currentTokenRisk+tradeValue > r.maxExposure*0.3 { // Max 30% per token
		return fmt.Errorf("trade would exceed per-token limit of %.2f (current: %.2f, trade: %.2f)",
			r.maxExposure*0.3, currentTokenRisk, tradeValue)
	}

	// Check slippage for the trade
	if trade.Direction == "BUY" {
		maxPrice := trade.Price * (1 + r.slippage)
		if trade.Price > maxPrice {
			return fmt.Errorf("price %.8f exceeds max allowed price %.8f (slippage: %.2f%%)",
				trade.Price, maxPrice, r.slippage*100)
		}
	}

	// Update risk tracking on successful validation
	r.currentRisks[trade.Token] = currentTokenRisk + tradeValue
	r.totalRisk += tradeValue

	return nil
}

func (r *RiskManager) GetCurrentRisk(token string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentRisks[token]
}

func (r *RiskManager) GetTotalRisk() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalRisk
}

func (r *RiskManager) UnwindPosition(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentRisk := r.currentRisks[token]
	if currentRisk == 0 {
		return nil
	}

	r.currentRisks[token] = 0
	r.totalRisk -= currentRisk
	return nil
}

func (r *RiskManager) GetPositionSize(token string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentRisks[token]
}

type Trade struct {
	Token     string
	Amount    float64
	Direction string
	Price     float64
}
