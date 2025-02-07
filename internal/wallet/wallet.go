package wallet

type WalletType string

const (
	TradingWallet WalletType = "A"
	ProfitWallet  WalletType = "B"
)

type KeyStore interface {
	Store(key []byte) error
	Retrieve() ([]byte, error)
}

type Manager interface {
	CreateWallet(walletType WalletType) error
	GetWallet(walletType WalletType) (*SolanaWallet, error)
	TransferFunds(from, to WalletType, amount float64) error
}

type WalletManager struct {
	keyStore KeyStore
	wallets  map[WalletType]Wallet
}

type Wallet interface {
	GetBalance() float64
	Transfer(to Wallet, amount float64) error
	ReceiveFunds(amount float64) error
}
