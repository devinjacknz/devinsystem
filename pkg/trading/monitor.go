package trading

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Monitor struct {
	logger *log.Logger
}

func NewMonitor(logFile string) (*Monitor, error) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(f, "", log.LstdFlags)
	return &Monitor{logger: logger}, nil
}

func (m *Monitor) LogTrade(trade *Trade) {
	m.logger.Printf("[TRADE] %s %s %.8f @ %.8f", 
		trade.Token, trade.Direction, trade.Amount, trade.Price)
}

func (m *Monitor) LogVolatility(token string, volatility float64) {
	m.logger.Printf("[VOLATILITY] %s %.2f%%", token, volatility*100)
}

func (m *Monitor) LogAISignal(token string, signal string, confidence float64) {
	m.logger.Printf("[AI] %s %s %.2f", token, signal, confidence)
}

func (m *Monitor) LogExposure(token string, exposure float64) {
	m.logger.Printf("[EXPOSURE] %s %.2f", token, exposure)
}

func (m *Monitor) LogSystem(msg string) {
	m.logger.Printf("[SYSTEM] %s", msg)
}

func (m *Monitor) LogError(msg string) {
	m.logger.Printf("[ERROR] %s", msg)
}

func (m *Monitor) LogJupiterSwap(inputToken, outputToken string, inputAmount, outputAmount float64, priceImpact float64) {
	m.logger.Printf("[JUPITER] Swap %f %s -> %f %s (Impact: %.2f%%)",
		inputAmount, inputToken, outputAmount, outputToken, priceImpact*100)
}

func (m *Monitor) LogRiskLimit(limitType string, currentValue float64, limit float64) {
	m.logger.Printf("[RISK] %s limit reached: current=%.2f limit=%.2f",
		limitType, currentValue, limit)
}

func (m *Monitor) LogBalance(balance float64) {
	m.logger.Printf("[BALANCE] Current balance: $%.2f", balance)
}

func (m *Monitor) LogMarketData(token string, price float64, volume float64) {
	m.logger.Printf("[MARKET] %s Price: %.8f Volume: %.2f Time: %s",
		token, price, volume, time.Now().Format(time.RFC3339))
}
