package integration

import (
	"context"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/trading"
	"github.com/devinjacknz/devinsystem/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestEndToEndTrading(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize components
	marketData := market.NewHeliusClient("https://mainnet.helius-rpc.com/?api-key=70b9a67a-f177-4881-9976-1e7d95facf43")
	ollama := models.NewOllamaClient("http://localhost:11434", "deepseek-r1")
	riskMgr := trading.NewRiskManager()

	tokenCache := utils.NewTokenCache(time.Hour, 30, func(ctx context.Context, token string) (*utils.TokenInfo, error) {
		data, err := marketData.GetMarketData(ctx, token)
		if err != nil {
			return nil, err
		}
		return &utils.TokenInfo{
			Symbol:    data.Symbol,
			Price:     data.Price,
			Volume:    data.Volume,
			UpdatedAt: data.Timestamp,
		}, nil
	})

	// Initialize engine
	engine := trading.NewEngine(marketData, ollama, riskMgr, tokenCache)

	// Start engine
	err := engine.Start(ctx)
	assert.NoError(t, err)
	defer engine.Stop()

	// Wait for market data collection
	time.Sleep(5 * time.Second)

	// Test token list retrieval
	tokens, err := tokenCache.GetTopTokens(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens)

	// Test trade execution
	err = engine.ExecuteTrade(ctx, tokens[0].Symbol, 0.1)
	assert.NoError(t, err)

	// Verify risk limits
	risk := riskMgr.GetCurrentRisk(tokens[0].Symbol)
	assert.Greater(t, risk, float64(0))
	assert.Less(t, risk, riskMgr.GetTotalRisk())
}

func TestRateLimiting(t *testing.T) {
	ctx := context.Background()
	marketData := market.NewHeliusClient("https://mainnet.helius-rpc.com/?api-key=70b9a67a-f177-4881-9976-1e7d95facf43")

	start := time.Now()
	_, _ = marketData.GetMarketData(ctx, "TEST1")
	_, _ = marketData.GetMarketData(ctx, "TEST2")

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(1), "Rate limiting should enforce 1 second delay")
}

func TestOllamaIntegration(t *testing.T) {
	ctx := context.Background()
	ollama := models.NewOllamaClient("http://localhost:11434", "deepseek-r1")

	marketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	decision, err := ollama.GenerateTradeDecision(ctx, marketData)
	if err != nil {
		t.Skip("Skipping test due to Ollama service error")
	}

	assert.NotNil(t, decision)
	assert.Contains(t, []string{"BUY", "SELL", "NOTHING"}, decision.Action)
	assert.GreaterOrEqual(t, decision.Confidence, float64(0))
	assert.LessOrEqual(t, decision.Confidence, float64(1))
}
