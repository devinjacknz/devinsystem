package exchange

type Manager interface {
	GetExchange(name string) (Exchange, error)
}

type Exchange interface {
	GetMarketPrice(symbol string) (float64, error)
	ExecuteOrder(order Order) error
}

type Order struct {
	Symbol    string  // Token mint address
	Side      string  // BUY or SELL
	Amount    float64 // Amount in base token
	Price     float64 // Price in USDC
	OrderType string  // MARKET only for Jupiter
	Wallet    string  // Wallet public key for trade execution
}

type SolanaAdapter struct {
	client interface{} // Solana client interface
	config Config
}

type Config struct {
	Endpoint string
	APIKey   string
}
