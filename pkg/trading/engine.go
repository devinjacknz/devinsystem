package trading

import (
	"context"
	"fmt"
	"log"
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

	log.Printf("[TRADE] Starting trade execution for %s, amount: %.4f", token, amount)

	// Get market data
	data, err := e.marketData.GetMarketData(ctx, token)
	if err != nil {
		log.Printf("[ERROR] Failed to get market data: %v", err)
		return fmt.Errorf("failed to get market data: %w", err)
	}
	log.Printf("[MARKET] Retrieved market data for %s: price=%.4f volume=%.2f", 
		token, data.Price, data.Volume)

	// Get AI decision
	decision, err := e.ollama.GenerateTradeDecision(ctx, data)
	if err != nil {
		log.Printf("[ERROR] Failed to generate trade decision: %v", err)
		return fmt.Errorf("failed to generate trade decision: %w", err)
	}
	log.Printf("[AI] Generated trade decision: action=%s confidence=%.2f reasoning=%s", 
		decision.Action, decision.Confidence, decision.Reasoning)

	// Create trade with risk validation
	trade := &risk.Trade{
		Token:     token,
		Amount:    amount,
		Direction: decision.Action,
		Price:     data.Price,
	}

	if err := e.riskMgr.ValidateTrade(ctx, trade); err != nil {
		log.Printf("[RISK] Trade validation failed: %v", err)
		return fmt.Errorf("trade validation failed: %w", err)
	}
	log.Printf("[RISK] Trade validated successfully")

	// Execute order through Jupiter DEX
	if err := e.jupiter.ExecuteOrder(exchange.Order{
		Symbol:    token,
		Side:      decision.Action,
		Amount:    amount,
		Price:     data.Price,
		OrderType: "MARKET",
		Wallet:    os.Getenv("WALLET"),
	}); err != nil {
		log.Printf("[ERROR] Swap execution failed: %v", err)
		return fmt.Errorf("swap execution failed: %w", err)
	}
	log.Printf("[TRADE] Successfully executed %s order for %s", decision.Action, token)

	// Update position tracking
	switch decision.Action {
	case "BUY":
		e.positions[token] += amount
		log.Printf("[POSITION] Updated %s position to %.4f", token, e.positions[token])
	case "SELL":
		e.positions[token] = 0
		log.Printf("[POSITION] Closed position for %s", token)
	}

	return nil
}

func (e *Engine) monitorMarkets(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
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
			log.Printf("[ERROR] Failed to generate decision for %s: %v", token.Symbol, err)
			continue
		}

		if (decision.Action == "BUY" || decision.Action == "SELL") && decision.Confidence > 0.1 {
			amount := calculateTradeAmount(data.Price)
			if decision.Action == "SELL" {
				if position, exists := e.positions[token.Symbol]; exists && position > 0 {
					amount = position
				} else {
					continue
				}
			}

			var executed bool
			for attempt := 1; attempt <= 3; attempt++ {
				log.Printf("[RETRY] Attempting trade %d/3 for %s %s", attempt, decision.Action, token.Symbol)
				if err := e.ExecuteTrade(ctx, token.Symbol, amount); err != nil {
					log.Printf("[RETRY] Trade attempt %d failed: %v", attempt, err)
					time.Sleep(time.Second)
					continue
				}
				log.Printf("[TRADE] Successfully executed %s order for %s, amount: %.4f, confidence: %.2f", 
					decision.Action, token.Symbol, amount, decision.Confidence)
				executed = true
				break
			}
			if !executed {
				log.Printf("[ERROR] All retry attempts failed for %s %s", decision.Action, token.Symbol)
			}
		}
	}
	return nil
}

func calculateTradeAmount(price float64) float64 {
	maxAmount := 100.0 // Max amount in USD for each trade
	return maxAmount / price
}
