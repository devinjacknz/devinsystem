package monitoring

import (
	"log"
	"time"
)

type Service struct {
	logFile string
}

func NewService() *Service {
	return &Service{
		logFile: "trading.log",
	}
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
