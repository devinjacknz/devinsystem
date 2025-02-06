package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/trading"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB
	mongoClient, err := mongo.NewClient("mongodb://localhost:27017", "trading")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	mongoRepo := mongo.NewRepository(mongoClient)

	// Create MongoDB indexes
	if err := mongoRepo.CreateIndexes(ctx); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Initialize components
	marketData := market.NewHeliusClient(os.Getenv("RPC_ENDPOINT"), mongoRepo)
	ollama := models.NewOllamaClient("http://localhost:11434", "deepseek-r1")
	riskMgr := trading.NewRiskManager()

	// Initialize token cache with 1-hour TTL and 30 token limit
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

	// Initialize monitor
	monitor, err := trading.NewMonitor("/home/ubuntu/repos/devinsystem/trading.log")
	if err != nil {
		log.Fatalf("Failed to initialize monitor: %v", err)
	}

	// Initialize trading engine
	engine := trading.NewEngine(marketData, ollama, riskMgr, tokenCache, mongoRepo)

	// Start trading engine
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start trading engine: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down trading system...")
	engine.Stop()
}
