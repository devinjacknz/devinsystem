package exchange

import (
	"errors"
	"sync"
)

type ExchangeManager struct {
	mu      sync.RWMutex
	jupiter *JupiterDEX
}

func NewExchangeManager() *ExchangeManager {
	return &ExchangeManager{
		jupiter: NewJupiterDEX(),
	}
}

func (m *ExchangeManager) GetJupiterDEX() *JupiterDEX {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jupiter
}

func (m *ExchangeManager) GetExchange(name string) (Exchange, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name != "jupiter" {
		return nil, errors.New("only Jupiter DEX is supported")
	}

	return m.jupiter, nil
}
