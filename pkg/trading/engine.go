package trading

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/devinjacknz/devinsystem/internal/exchange"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

type Engine struct {
	mu          sync.RWMutex
	marketData  market.Client
	ollama      models.Client
	riskMgr     risk.Manager
	tokenCache  *utils.TokenCache
	jupiter     *exchange.JupiterDEX
	isRunning   bool
	stopChan    chan struct{}
	positions   map[string]float64
}

func NewEngine(marketData market.Client, ollama models.Client, riskMgr risk.Manager, tokenCache *utils.TokenCache) *Engine {
	return &Engine{
		marketData: marketData,
		ollama:    ollama,
		riskMgr:   riskMgr,
		tokenCache: tokenCache,
		jupiter:    exchange.NewJupiterDEX(),
		stopChan:  make(chan struct{}),
		positions: make(map[string]float64),
	}
}

func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return nil
	}
	e.isRunning = true
	e.mu.Unlock()

	go e.monitorMarkets(ctx)
	return nil
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.isRunning {
		return
	}
	close(e.stopChan)
	e.isRunning = false
}

func (e *Engine) ExecuteTrade(ctx context.Context, token string, amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get market data
	data, err := e.marketData.GetMarketData(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get market data: %w", err)
	}

	// Get AI decision
	decision, err := e.ollama.GenerateTradeDecision(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to generate trade decision: %w", err)
	}

	// Create trade with risk validation
	trade := &risk.Trade{
		Token:     token,
		Amount:    amount,
		Direction: decision.Action,
		Price:     data.Price,
	}

	if err := e.riskMgr.ValidateTrade(ctx, trade); err != nil {
		return fmt.Errorf("trade validation failed: %w", err)
	}

	// Execute order through Jupiter DEX
	if err := e.jupiter.ExecuteOrder(exchange.Order{
		Symbol:    token,
		Side:      decision.Action,
		Amount:    amount,
		Price:     data.Price,
		OrderType: "MARKET",
		Wallet:    os.Getenv("WALLET"),
	}); err != nil {
		return fmt.Errorf("swap execution failed: %w", err)
	}

	// Update position tracking
	switch decision.Action {
	case "BUY":
		e.positions[token] += amount
	case "SELL":
		e.positions[token] = 0
	}

	return nil
}

func (e *Engine) monitorMarkets(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			if err := e.processMarketData(ctx); err != nil {
				continue
			}
		}
	}
}

func (e *Engine) processMarketData(ctx context.Context) error {
	tokens, err := e.tokenCache.GetTopTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to get top tokens: %w", err)
	}

	for _, token := range tokens {
		data, err := e.marketData.GetMarketData(ctx, token.Symbol)
		if err != nil {
			continue
		}

		decision, err := e.ollama.GenerateTradeDecision(ctx, data)
		if err != nil {
			continue
		}

		if decision.Action == "BUY" && decision.Confidence > 0.7 {
			if err := e.ExecuteTrade(ctx, token.Symbol, calculateTradeAmount(data.Price)); err != nil {
				continue
			}
		}
	}

	return nil
}

func calculateTradeAmount(price float64) float64 {
	maxAmount := 3.0 // Max amount in USD
	return maxAmount / price
}
