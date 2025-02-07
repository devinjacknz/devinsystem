package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/agents"
	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/trading"
	"github.com/devinjacknz/devinsystem/pkg/utils"
	"github.com/devinjacknz/devinsystem/internal/wallet"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	marketData := market.NewHeliusClient(os.Getenv("RPC_ENDPOINT"))
	
	// Initialize Ollama with DeepSeek R1 model
	modelFactory := models.NewModelFactory()
	ollama := models.NewOllamaModel("deepseek-r1")
	if err := modelFactory.RegisterModel("ollama", ollama); err != nil {
		log.Fatalf("Failed to register Ollama model: %v", err)
	}

	// Initialize risk manager
	riskMgr := trading.NewRiskManager()
	
	// Initialize wallet manager
	walletMgr, err := wallet.NewWalletManager()
	if err != nil {
		log.Fatalf("Failed to initialize wallet manager: %v", err)
	}
	
	if err := walletMgr.CreateWallet(wallet.TradingWallet); err != nil {
		log.Fatalf("Failed to create trading wallet: %v", err)
	}

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

	// Initialize trading engine
	engine := trading.NewEngine(marketData, ollama, riskMgr, tokenCache, walletMgr)

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
