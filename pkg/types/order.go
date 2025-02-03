package types

type OrderType string
type OrderSide string

const (
	LimitOrder  OrderType = "limit"
	MarketOrder OrderType = "market"
	
	BuyOrder  OrderSide = "buy"
	SellOrder OrderSide = "sell"
)

type Order struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Side      OrderSide `json:"side"`
	Type      OrderType `json:"type"`
	Amount    float64   `json:"amount"`
	Price     float64   `json:"price"`
	Timestamp int64     `json:"timestamp"`
	Exchange  string    `json:"exchange"`
}
