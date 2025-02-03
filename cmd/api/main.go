package main

import (
	"log"
	"net/http"
	"os"
	
	"github.com/devinjacknz/devintrade/internal/api"
	"github.com/devinjacknz/devintrade/internal/risk"
	"github.com/devinjacknz/devintrade/internal/trading"
	"github.com/devinjacknz/devintrade/internal/wallet"
)

func main() {
	// Initialize wallet manager
	walletManager, err := wallet.NewWalletManager()
	if err != nil {
		log.Fatalf("Failed to initialize wallet manager: %v", err)
	}

	// Initialize risk manager
	riskManager := risk.NewManager()

	// Initialize trading engine with dependencies
	tradingEngine := trading.NewTradingEngine(riskManager, walletManager)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Create and start server
	server := api.NewServer(tradingEngine, walletManager, jwtSecret)
	log.Fatal(http.ListenAndServe(":8080", server))
}
