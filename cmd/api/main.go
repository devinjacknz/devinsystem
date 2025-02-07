package main

import (
	"log"
	"net/http"
	"os"
	
	"github.com/devinjacknz/devinsystem/internal/api"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/internal/trading"
	"github.com/devinjacknz/devinsystem/internal/wallet"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
