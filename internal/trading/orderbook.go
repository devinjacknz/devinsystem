package trading

import (
	"sort"
	"sync"
)

type OrderBook struct {
	mu     sync.RWMutex
	orders map[string]Order
	bids   []PriceLevel
	asks   []PriceLevel
}

type PriceLevel struct {
	Price     float64
	Size      float64
	OrdersNum int
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		orders: make(map[string]Order),
		bids:   make([]PriceLevel, 0),
		asks:   make([]PriceLevel, 0),
	}
}

func (ob *OrderBook) AddOrder(order Order) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.orders[order.ID] = order

	level := PriceLevel{
		Price:     order.Price,
		Size:      order.Amount,
		OrdersNum: 1,
	}

	if order.Side == "buy" {
		ob.bids = insertPriceLevel(ob.bids, level, true)
	} else {
		ob.asks = insertPriceLevel(ob.asks, level, false)
	}

	return nil
}

func (ob *OrderBook) RemoveOrder(orderID string) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	delete(ob.orders, orderID)
	return nil
}

func insertPriceLevel(levels []PriceLevel, level PriceLevel, isBid bool) []PriceLevel {
	idx := sort.Search(len(levels), func(i int) bool {
		if isBid {
			return levels[i].Price <= level.Price
		}
		return levels[i].Price >= level.Price
	})

	levels = append(levels, PriceLevel{})
	copy(levels[idx+1:], levels[idx:])
	levels[idx] = level
	return levels
}
