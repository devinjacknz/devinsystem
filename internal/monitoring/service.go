package monitoring

import (
	"fmt"
	"log"
	"os"
)

type Service struct {
	logFile string
}

func NewService() *Service {
	s := &Service{
		logFile: "/home/ubuntu/repos/devinsystem/trading.log",
	}
	if err := s.init(); err != nil {
		log.Printf("[ERROR] Failed to initialize monitoring service: %v", err)
	}
	return s
}

func (s *Service) init() error {
	f, err := os.OpenFile(s.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	log.SetOutput(f)
	return nil
}

func (s *Service) LogTrade(symbol string, side string, amount float64, price float64) {
	log.Printf("[TRADE] %s %s %.8f @ %.8f", symbol, side, amount, price)
}

func (s *Service) LogVolatility(symbol string, volatility float64) {
	log.Printf("[VOLATILITY] %s %.2f%%", symbol, volatility*100)
}

func (s *Service) LogAISignal(symbol string, signal string, confidence float64) {
	log.Printf("[AI] %s %s %.2f", symbol, signal, confidence)
}

func (s *Service) LogExposure(symbol string, exposure float64) {
	log.Printf("[EXPOSURE] %s %.2f", symbol, exposure)
}

func (s *Service) LogSystem(msg string) {
	log.Printf("[SYSTEM] %s", msg)
}

func (s *Service) LogError(msg string) {
	log.Printf("[ERROR] %s", msg)
}

func (s *Service) LogJupiterSwap(inputToken, outputToken string, inputAmount, outputAmount float64, priceImpact float64) {
	log.Printf("[JUPITER] Swap %f %s -> %f %s (Impact: %.2f%%)",
		inputAmount, inputToken, outputAmount, outputToken, priceImpact*100)
}
