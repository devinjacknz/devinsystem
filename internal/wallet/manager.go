package wallet

import (
	"errors"
	"sync"
)

type walletManager struct {
	mu       sync.RWMutex
	keyStore *HSMKeyStore
	wallets  map[WalletType]*SolanaWallet
}

func NewWalletManager() (*walletManager, error) {
	keyStore, err := NewHSMKeyStore()
	if err != nil {
		return nil, err
	}

	return &walletManager{
		keyStore: keyStore,
		wallets:  make(map[WalletType]*SolanaWallet),
	}, nil
}

func (m *walletManager) CreateWallet(walletType WalletType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.wallets[walletType]; exists {
		return errors.New("wallet already exists")
	}

	wallet, err := NewSolanaWallet(string(walletType), m.keyStore)
	if err != nil {
		return err
	}

	m.wallets[walletType] = wallet
	return nil
}

func (m *walletManager) TransferFunds(from, to WalletType, amount float64) error {
	m.mu.RLock()
	fromWallet, fromExists := m.wallets[from]
	toWallet, toExists := m.wallets[to]
	m.mu.RUnlock()

	if !fromExists || !toExists {
		return errors.New("wallet not found")
	}

	return fromWallet.Transfer(toWallet, amount)
}

func (m *walletManager) GetWallet(walletType WalletType) (*SolanaWallet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wallet, exists := m.wallets[walletType]
	if !exists {
		return nil, errors.New("wallet not found")
	}

	return wallet, nil
}
