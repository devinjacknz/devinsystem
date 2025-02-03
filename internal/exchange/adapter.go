package exchange

type Manager interface {
	GetExchange(name string) (Exchange, error)
}

type Exchange interface {
	GetMarketPrice(symbol string) (float64, error)
	ExecuteOrder(order Order) error
}

type Order struct {
	Symbol    string
	Side      string
	Amount    float64
	Price     float64
	OrderType string
}

type SolanaAdapter struct {
	client interface{} // Solana client interface
	config Config
}

type Config struct {
	Endpoint string
	APIKey   string
}
