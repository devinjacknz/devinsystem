package risk

import (
	"errors"
	"sync"
)

type SlippageProtection struct {
	mu              sync.RWMutex
	maxSlippageBps  map[string]int
	defaultSlippage int
}

func NewSlippageProtection(defaultSlippageBps int) *SlippageProtection {
	return &SlippageProtection{
		maxSlippageBps:  make(map[string]int),
		defaultSlippage: defaultSlippageBps,
	}
}

func (sp *SlippageProtection) SetMaxSlippage(symbol string, basisPoints int) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.maxSlippageBps[symbol] = basisPoints
	return nil
}

func (sp *SlippageProtection) ValidateSlippage(symbol string, expectedPrice, actualPrice float64) error {
	sp.mu.RLock()
	maxBps, exists := sp.maxSlippageBps[symbol]
	sp.mu.RUnlock()

	if !exists {
		maxBps = sp.defaultSlippage
	}

	slippageBps := int((actualPrice-expectedPrice)/expectedPrice * 10000)
	if slippageBps > maxBps {
		return errors.New("slippage exceeds maximum allowed")
	}

	return nil
}
