package utils

import (
	"encoding/json"
	"os"
)

type Config struct {
	APIPort     int    `json:"api_port"`
	JWTSecret   string `json:"jwt_secret"`
	Environment string `json:"environment"`
	
	// Exchange configurations
	SolanaRPCURL string `json:"solana_rpc_url"`
	PumpFunURL   string `json:"pump_fun_url"`
	
	// AI model configurations
	OllamaURL    string `json:"ollama_url"`
	DeepSeekURL  string `json:"deepseek_url"`
}

// Log markers for structured logging
const (
	LogMarkerSystem   = "[SYSTEM]"
	LogMarkerTrade    = "[TRADE]"
	LogMarkerMarket   = "[MARKET]"
	LogMarkerAI       = "[AI]"
	LogMarkerWallet   = "[WALLET]"
	LogMarkerRisk     = "[RISK]"
	LogMarkerPerf     = "[PERF]"
	LogMarkerError    = "[ERROR]"
	LogMarkerRetry    = "[RETRY]"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}
