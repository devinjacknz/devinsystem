package integration

import (
	"context"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/db/mongo"
	"github.com/stretchr/testify/assert"
)

func TestMongoDBIntegration(t *testing.T) {
	ctx := context.Background()
	client, err := mongo.NewClient("mongodb://localhost:27017", "trading_test")
	assert.NoError(t, err)
	defer client.Close(ctx)

	repo := mongo.NewRepository(client)
	err = repo.CreateIndexes(ctx)
	assert.NoError(t, err)

	// Test market data persistence
	marketData := &mongo.MarketData{
		Token:     "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"liquidity": 500.0,
			"volatility": 0.05,
		},
	}
	err = repo.SaveMarketData(ctx, marketData)
	assert.NoError(t, err)

	// Test AI decision persistence
	aiDecision := &mongo.AIDecision{
		Token:       "TEST",
		FinalAction: "BUY",
		Confidence:  0.85,
		Timestamp:   time.Now(),
		Models: []mongo.ModelDecision{{
			Name:       "ollama",
			Action:     "BUY",
			Confidence: 0.85,
			Reasoning:  "Strong buy signal based on volume",
		}},
	}
	err = repo.SaveAIDecision(ctx, aiDecision)
	assert.NoError(t, err)

	// Test performance metrics
	perf := &mongo.Performance{
		Token:      "TEST",
		Start:      time.Now().Add(-24 * time.Hour),
		End:        time.Now(),
		AIAccuracy: 0.75,
		Metrics: map[string]interface{}{
			"roi":          0.15,
			"trades":       10,
			"success_rate": 0.80,
		},
	}
	err = repo.SavePerformance(ctx, perf)
	assert.NoError(t, err)

	// Test risk events
	riskEvent := &mongo.RiskEvent{
		Token:     "TEST",
		Type:      "STOP_LOSS",
		Timestamp: time.Now(),
		AIDecision: aiDecision,
		Details: map[string]interface{}{
			"price":    95.0,
			"exposure": 1000.0,
		},
	}
	err = repo.SaveRiskEvent(ctx, riskEvent)
	assert.NoError(t, err)
}

func TestRateLimiting(t *testing.T) {
	ctx := context.Background()
	client, err := mongo.NewClient("mongodb://localhost:27017", "trading_test")
	assert.NoError(t, err)
	defer client.Close(ctx)

	repo := mongo.NewRepository(client)

	start := time.Now()
	for i := 0; i < 3; i++ {
		marketData := &mongo.MarketData{
			Token:     "TEST",
			Price:     100.0,
			Volume:    1000.0,
			Timestamp: time.Now(),
		}
		err := repo.SaveMarketData(ctx, marketData)
		assert.NoError(t, err)
	}

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(2),
		"Rate limiting should enforce 1 RPS")
}
