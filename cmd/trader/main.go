package main

import (
	"log"
	
	"github.com/devinjacknz/devinsystem/internal/ai"
	"github.com/devinjacknz/devinsystem/internal/exchange"
	"github.com/devinjacknz/devinsystem/internal/monitoring"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/internal/trading"
	"github.com/devinjacknz/devinsystem/pkg/utils"
)

func main() {
	config, err := utils.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize AI service with DeepSeek R1:1.5b
	aiService := ai.NewService(config.OllamaURL, config.DeepSeekModel)
	
	// Initialize risk manager with meme coin parameters
	riskMgr := risk.NewManager()
	
	// Initialize monitoring
	monitor := monitoring.NewService()
	
	// Initialize exchanges
	solanaDEX := exchange.NewSolanaDEX(config.SolanaRPCURL)
	pumpFun := exchange.NewPumpFun(config.PumpFunURL)
	
	// Initialize trading engine with both exchanges
	engine := trading.NewTradingEngine(
		riskMgr,
		[]exchange.Exchange{solanaDEX, pumpFun},
		aiService,
		monitor,
	)
	
	log.Fatal(engine.Start())
}
