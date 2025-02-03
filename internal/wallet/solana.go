package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
)

type SolanaWallet struct {
	id        string
	keyStore  *HSMKeyStore
	publicKey ed25519.PublicKey
	balance   float64
}

func NewSolanaWallet(id string, keyStore *HSMKeyStore) (*SolanaWallet, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	if err := keyStore.Store(id, privateKey); err != nil {
		return nil, err
	}

	return &SolanaWallet{
		id:        id,
		keyStore:  keyStore,
		publicKey: publicKey,
		balance:   0,
	}, nil
}

func (w *SolanaWallet) GetBalance() float64 {
	return w.balance
}

func (w *SolanaWallet) Transfer(to Wallet, amount float64) error {
	if amount > w.balance {
		return errors.New("insufficient funds")
	}
	
	// In production, this would interact with Solana blockchain
	w.balance -= amount
	if err := to.ReceiveFunds(amount); err != nil {
		w.balance += amount // rollback on failure
		return err
	}
	
	return nil
}

func (w *SolanaWallet) ReceiveFunds(amount float64) error {
	w.balance += amount
	return nil
}

func (w *SolanaWallet) GetAddress() string {
	return hex.EncodeToString(w.publicKey)
}
