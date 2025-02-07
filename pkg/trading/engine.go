package trading

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/devinjacknz/devinsystem/internal/exchange"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/logging"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

type Engine struct {
	mu          sync.RWMutex
	marketData  market.Client
	ollama      models.Client
	riskMgr     risk.Manager
	tokenCache  *utils.TokenCache // Keep utils for TokenCache type
	jupiter     *exchange.JupiterDEX
	isRunning   bool
	stopChan    chan struct{}
	positions   map[string]float64
}

func NewEngine(marketData market.Client, ollama models.Client, riskMgr risk.Manager, tokenCache *utils.TokenCache) *Engine {
	engine := &Engine{
		marketData: marketData,
		ollama:    ollama,
		riskMgr:   riskMgr,
		tokenCache: tokenCache,
		jupiter:    exchange.NewJupiterDEX(),
		stopChan:  make(chan struct{}),
		positions: make(map[string]float64),
	}

	// Initialize token cache with default tokens
	defaultTokens := []string{
		"So11111111111111111111111111111111111111112", // Wrapped SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
	}

	for _, token := range defaultTokens {
		if info, err := marketData.GetMarketData(context.Background(), token); err == nil {
			tokenCache.Set(token, &utils.TokenInfo{
				Symbol:    token,
				Price:     info.Price,
				Volume:    info.Volume,
				UpdatedAt: time.Now(),
			})
		}
	}

	return engine
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
	start := time.Now()
	e.mu.Lock()
	defer e.mu.Unlock()
	defer func() {
		log.Printf("%s Trade execution for %s took %v", logging.LogMarkerPerf, token, time.Since(start))
	}()

	wallet := os.Getenv("WALLET")
	if wallet == "" {
		return fmt.Errorf("wallet address not configured")
	}
	log.Printf("%s Starting trade execution for %s, amount: %.4f using wallet: %s", logging.LogMarkerTrade, 
		token, amount, wallet)

	// Get market data with timing
	marketStart := time.Now()
	data, err := e.marketData.GetMarketData(ctx, token)
	if err != nil {
		log.Printf("%s Failed to get market data: %v", logging.LogMarkerError, err)
		return fmt.Errorf("failed to get market data: %w", err)
	}
	log.Printf("%s Retrieved data for %s: price=%.4f volume=%.2f (took %v)", logging.LogMarkerMarket, 
		token, data.Price, data.Volume, time.Since(marketStart))

	// Get AI decision with timing
	aiStart := time.Now()
	decision, err := e.ollama.GenerateTradeDecision(ctx, data)
	if err != nil {
		log.Printf("%s Failed to generate trade decision: %v", logging.LogMarkerError, err)
		return fmt.Errorf("failed to generate trade decision: %w", err)
	}
	log.Printf("%s Generated decision: action=%s confidence=%.2f reasoning=%s (took %v)", logging.LogMarkerAI, 
		decision.Action, decision.Confidence, decision.Reasoning, time.Since(aiStart))

	// Create trade with risk validation
	trade := &risk.Trade{
		Token:     token,
		Amount:    amount,
		Direction: decision.Action,
		Price:     data.Price,
	}

	riskStart := time.Now()
	if err := e.riskMgr.ValidateTrade(ctx, trade); err != nil {
		log.Printf("%s Trade validation failed: %v", logging.LogMarkerRisk, err)
		return fmt.Errorf("trade validation failed: %w", err)
	}
	log.Printf("%s Trade validated successfully (took %v)", logging.LogMarkerRisk, time.Since(riskStart))

	// Track swap execution time
	swapStart := time.Now()
	// Execute order through Jupiter DEX with wallet
	order := exchange.Order{
		Symbol:    token,
		Side:      decision.Action,
		Amount:    amount,
		Price:     data.Price,
		OrderType: "MARKET",
		Wallet:    wallet,
	}
	log.Printf("%s Executing order: %+v", logging.LogMarkerTrade, order)
	if err := e.jupiter.ExecuteOrder(order); err != nil {
		log.Printf("%s Swap execution failed: %v", logging.LogMarkerError, err)
		return fmt.Errorf("swap execution failed: %w", err)
	}
	log.Printf("%s Successfully executed %s order for %s (swap took %v)", logging.LogMarkerTrade, 
		decision.Action, token, time.Since(swapStart))

	// Update position tracking
	switch decision.Action {
	case "BUY":
		e.positions[token] += amount
		log.Printf("%s Updated %s position to %.4f", logging.LogMarkerTrade, token, e.positions[token])
	case "SELL":
		e.positions[token] = 0
		log.Printf("%s Closed position for %s", logging.LogMarkerTrade, token)
	}

	return nil
}

func (e *Engine) monitorMarkets(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // More frequent monitoring for testing
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
	start := time.Now()
	defer func() {
		log.Printf("%s Market data processing took %v", logging.LogMarkerPerf, time.Since(start))
	}()

	log.Printf("%s Starting market data processing cycle", logging.LogMarkerMarket)
	tokens, err := e.tokenCache.GetTopTokens(ctx)
	if err != nil {
		log.Printf("%s Failed to get top tokens: %v", logging.LogMarkerError, err)
		return fmt.Errorf("failed to get top tokens: %w", err)
	}

	if len(tokens) == 0 {
		log.Printf("%s No tokens available for analysis, initializing defaults", logging.LogMarkerMarket)
		defaultTokens := []string{
			"So11111111111111111111111111111111111111112", // Wrapped SOL
		}
		for _, token := range defaultTokens {
			if info, err := e.marketData.GetMarketData(ctx, token); err == nil {
				e.tokenCache.Set(token, &utils.TokenInfo{
					Symbol:    token,
					Price:     info.Price,
					Volume:    info.Volume,
					UpdatedAt: time.Now(),
				})
			}
		}
		tokens, _ = e.tokenCache.GetTopTokens(ctx)
	}

	log.Printf("%s Retrieved %d tokens for analysis", logging.LogMarkerMarket, len(tokens))

	for _, token := range tokens {
		data, err := e.marketData.GetMarketData(ctx, token)
		if err != nil {
			log.Printf("%s Failed to get market data for %s: %v", logging.LogMarkerError, token, err)
			continue
		}

		if data.Price <= 0 {
			log.Printf("%s Invalid price for %s: %.8f", logging.LogMarkerError, token, data.Price)
			continue
		}

		decision, err := e.ollama.GenerateTradeDecision(ctx, data)
		if err != nil {
			log.Printf("%s Failed to generate decision for %s: %v", logging.LogMarkerError, token.Symbol, err)
			continue
		}

		if (decision.Action == "BUY" || decision.Action == "SELL") && decision.Confidence > 0.15 {
			amount := calculateTradeAmount(data.Price, data.Volume)
			if decision.Action == "SELL" {
				if position, exists := e.positions[token.Symbol]; exists && position > 0 {
					amount = position
				} else {
					continue
				}
			}

			var executed bool
			for attempt := 1; attempt <= 3; attempt++ {
				log.Printf("%s Attempting trade %d/3 for %s %s", logging.LogMarkerRetry, attempt, decision.Action, token.Symbol)
				if err := e.ExecuteTrade(ctx, token.Symbol, amount); err != nil {
					log.Printf("%s Trade attempt %d failed: %v", logging.LogMarkerRetry, attempt, err)
					time.Sleep(time.Second)
					continue
				}
				log.Printf("%s Successfully executed %s order for %s, amount: %.4f, confidence: %.2f", logging.LogMarkerTrade, 
					decision.Action, token.Symbol, amount, decision.Confidence)
				executed = true
				break
			}
			if !executed {
				log.Printf("%s All retry attempts failed for %s %s", logging.LogMarkerError, decision.Action, token.Symbol)
			}
		}
	}
	return nil
}

func calculateTradeAmount(price float64, volume float64) float64 {
	if price <= 0 {
		log.Printf("%s Invalid price for trade amount calculation: %.8f", logging.LogMarkerError, price)
		return 0
	}

	// Base amount in SOL (0.1 SOL)
	baseAmount := 0.1
	if volume <= 0 {
		log.Printf("%s Invalid volume for trade amount calculation: %.2f", logging.LogMarkerError, volume)
		return baseAmount
	}

	// Scale amount based on volume (max 1.0 SOL)
	liquidityFactor := math.Min(1.0, volume/10000.0)
	amount := baseAmount * (1 + liquidityFactor)

	if amount <= 0 || math.IsInf(amount, 0) || math.IsNaN(amount) {
		log.Printf("%s Invalid calculated amount: %.8f", logging.LogMarkerError, amount)
		return baseAmount
	}

	log.Printf("%s Calculated trade amount: %.8f SOL (volume: %.2f)", logging.LogMarkerTrade, amount, volume)
	return amount
}
