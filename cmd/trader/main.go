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
	"github.com/devinjacknz/devinsystem/pkg/logging"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure logging
	log.Printf("%s Starting trading system...", logging.LogMarkerSystem)
	log.Printf("%s Configuration:", logging.LogMarkerSystem)
	log.Printf("  • Wallet: %s", os.Getenv("WALLET"))
	log.Printf("  • RPC Endpoint: %s", os.Getenv("RPC_ENDPOINT"))
	log.Printf("  • Ollama URL: %s", os.Getenv("OLLAMA_URL"))
	log.Printf("  • Ollama Model: %s", os.Getenv("OLLAMA_MODEL"))
	
	if os.Getenv("WALLET") == "" || os.Getenv("RPC_ENDPOINT") == "" || 
		os.Getenv("OLLAMA_URL") == "" || os.Getenv("OLLAMA_MODEL") == "" {
		log.Fatal("Missing required environment variables")
	}
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
	log.Printf("%s Initializing market data client with primary RPC: %s", logging.LogMarkerSystem, os.Getenv("RPC_ENDPOINT"))
	marketData := market.NewFallbackClient(os.Getenv("RPC_ENDPOINT"))
	if err := marketData.ValidateConnection(ctx); err != nil {
		log.Printf("%s Failed to validate RPC connection: %v, switching to backup RPC", logging.LogMarkerError, err)
		marketData = market.NewFallbackClient("https://eclipse.helius-rpc.com/")
	}
	
	// Initialize Ollama with DeepSeek R1 model
	ollama := models.NewOllamaClient(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))

	// Initialize risk manager with 3M max exposure for meme coins
	riskMgr := risk.NewRiskManager(nil, 3_000_000)

	// Initialize token cache with test tokens
	tokenCache := utils.NewTokenCache(time.Hour, 30, func(ctx context.Context, token string) (*utils.TokenInfo, error) {
		data, err := marketData.GetMarketData(ctx, token)
		if err != nil {
			log.Printf("%s Failed to get market data for %s: %v", logging.LogMarkerError, token, err)
			return nil, err
		}
		log.Printf("%s Retrieved market data for %s: price=%.8f volume=%.2f", logging.LogMarkerMarket,
			token, data.Price, data.Volume)
		return &utils.TokenInfo{
			Symbol:    data.Symbol,
			Price:     data.Price,
			Volume:    data.Volume,
			UpdatedAt: data.Timestamp,
		}, nil
	})

	// Pre-populate cache with test tokens
	testTokens := []string{
		"So11111111111111111111111111111111111111112", // Wrapped SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
	}
	for _, token := range testTokens {
		if _, err := tokenCache.Get(ctx, token); err != nil {
			log.Printf("%s Failed to initialize token %s: %v", logging.LogMarkerError, token, err)
		}
	}

	// Initialize trading engine with Jupiter DEX only
	engine := trading.NewEngine(marketData, ollama, riskMgr, tokenCache)

	log.Printf("%s Starting trading engine with Jupiter DEX integration", logging.LogMarkerSystem)
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start trading engine: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Printf("%s Shutting down trading system...", logging.LogMarkerSystem)
	engine.Stop()
}
