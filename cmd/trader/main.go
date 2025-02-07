package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/trading"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure logging
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "trading.log"
	}
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Initialize components
	marketData := market.NewFallbackClient(os.Getenv("RPC_ENDPOINT"))
	
	// Initialize Ollama with DeepSeek R1 model
	ollama := models.NewOllamaClient(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))

	// Initialize risk manager with 3M max exposure for meme coins
	riskMgr := risk.NewRiskManager(nil, 3_000_000)

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

	// Initialize trading engine with Jupiter DEX only
	engine := trading.NewEngine(marketData, ollama, riskMgr, tokenCache)

	log.Println("[SYSTEM] Starting trading engine with Jupiter DEX integration")
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start trading engine: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("[SYSTEM] Shutting down trading system...")
	engine.Stop()
}
