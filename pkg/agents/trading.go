package agents

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/devinjacknz/devinsystem/internal/exchange"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
)

type TradingAgent struct {
	*BaseAgent
	marketData market.Client
	ollama    models.Client
	riskMgr   risk.Manager
	jupiter   *exchange.JupiterDEX
	stopChan  chan struct{}
}

func NewTradingAgent(marketData market.Client, ollama models.Client, riskMgr *risk.Manager) *TradingAgent {
	return &TradingAgent{
		BaseAgent:  NewBaseAgent("trading"),
		marketData: marketData,
		ollama:    ollama,
		riskMgr:   riskMgr,
		jupiter:   exchange.NewJupiterDEX(),
		stopChan:  make(chan struct{}),
	}
}

func (t *TradingAgent) Initialize(ctx context.Context) error {
	if err := t.BaseAgent.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize base agent: %w", err)
	}

	log.Println("[SYSTEM] Initializing trading agent")
	log.Println("[SYSTEM] Initializing exchange: Jupiter DEX")
	return nil
}

func (t *TradingAgent) Run(ctx context.Context) error {
	if err := t.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize trading agent: %w", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.stopChan:
			return nil
		case <-ticker.C:
			if err := t.processMarketData(ctx); err != nil {
				log.Printf("[ERROR] Failed to process market data: %v", err)
				continue
			}
		}
	}
}

func (t *TradingAgent) Stop() {
	close(t.stopChan)
}

func (t *TradingAgent) processMarketData(ctx context.Context) error {
	tokens, err := t.marketData.GetTopTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to get top tokens: %w", err)
	}

	for _, token := range tokens {
		data, err := t.marketData.GetMarketData(ctx, token.Symbol)
		if err != nil {
			log.Printf("[ERROR] Failed to get market data for %s: %v", token.Symbol, err)
			continue
		}

		decision, err := t.ollama.GenerateTradeDecision(ctx, data)
		if err != nil {
			log.Printf("[ERROR] Failed to generate trade decision for %s: %v", token.Symbol, err)
			continue
		}

		if decision.Action == "BUY" && decision.Confidence > 0.7 {
			if err := t.executeTrade(ctx, token.Symbol, calculateTradeAmount(data.Price)); err != nil {
				log.Printf("[ERROR] Failed to execute trade for %s: %v", token.Symbol, err)
				continue
			}
		}
	}

	return nil
}

func (t *TradingAgent) executeTrade(ctx context.Context, token string, amount float64) error {
	data, err := t.marketData.GetMarketData(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get market data: %w", err)
	}

	if err := t.riskMgr.ValidateTrade(ctx, &risk.Trade{
		Token:     token,
		Amount:    amount,
		Direction: "BUY",
		Price:     data.Price,
	}); err != nil {
		return fmt.Errorf("trade validation failed: %w", err)
	}

	quote, err := t.jupiter.GetQuote(ctx, token, "USDC", fmt.Sprintf("%.0f", amount))
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	if err := t.jupiter.ExecuteOrder(exchange.Order{
		Symbol:    token,
		Side:      "BUY",
		Amount:    amount,
		Price:     quote.Price,
		OrderType: "MARKET",
	}); err != nil {
		return fmt.Errorf("swap execution failed: %w", err)
	}

	return nil
}

func calculateTradeAmount(price float64) float64 {
	maxAmount := 3.0 // Max amount in USD
	return maxAmount / price
}
