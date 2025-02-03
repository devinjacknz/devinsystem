package risk

import (
	"sync"
)

type StopLoss struct {
	mu           sync.RWMutex
	stopLevels   map[string]float64
	trailingGaps map[string]float64
}

func NewStopLoss() *StopLoss {
	return &StopLoss{
		stopLevels:   make(map[string]float64),
		trailingGaps: make(map[string]float64),
	}
}

func (sl *StopLoss) SetStopLoss(symbol string, price float64) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.stopLevels[symbol] = price
	return nil
}

func (sl *StopLoss) SetTrailingStop(symbol string, gap float64) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.trailingGaps[symbol] = gap
	return nil
}

func (sl *StopLoss) CheckStopLoss(symbol string, currentPrice float64) (bool, error) {
	sl.mu.RLock()
	stopLevel, exists := sl.stopLevels[symbol]
	sl.mu.RUnlock()

	if !exists {
		return false, nil
	}

	return currentPrice <= stopLevel, nil
}

func (sl *StopLoss) UpdateTrailingStop(symbol string, currentPrice float64) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	gap, exists := sl.trailingGaps[symbol]
	if !exists {
		return nil
	}

	newStopLevel := currentPrice - gap
	currentStop := sl.stopLevels[symbol]

	if newStopLevel > currentStop {
		sl.stopLevels[symbol] = newStopLevel
	}

	return nil
}
