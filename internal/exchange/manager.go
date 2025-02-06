package exchange

import (
	"errors"
	"sync"
)

type ExchangeManager struct {
	mu        sync.RWMutex
	exchanges map[string]Exchange
}

func NewExchangeManager(solanaURL, pumpURL string) *ExchangeManager {
	manager := &ExchangeManager{
		exchanges: make(map[string]Exchange),
	}

	manager.exchanges["solana"] = NewSolanaDEX(solanaURL)
	manager.exchanges["pump"] = NewPumpFun(pumpURL)
	manager.exchanges["jupiter"] = NewJupiterDEX()

	return manager
}

func (m *ExchangeManager) GetExchange(name string) (Exchange, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exchange, exists := m.exchanges[name]
	if !exists {
		return nil, errors.New("exchange not found")
	}

	return exchange, nil
}
