package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"sync"
)

type HSMKeyStore struct {
	mu       sync.RWMutex
	keys     map[string][]byte
	masterKey []byte
}

func NewHSMKeyStore() (*HSMKeyStore, error) {
	masterKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, masterKey); err != nil {
		return nil, err
	}
	
	return &HSMKeyStore{
		keys:      make(map[string][]byte),
		masterKey: masterKey,
	}, nil
}

func (ks *HSMKeyStore) Store(id string, key []byte) error {
	block, err := aes.NewCipher(ks.masterKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	
	ks.keys[id] = gcm.Seal(nonce, nonce, key, nil)
	return nil
}

func (ks *HSMKeyStore) Retrieve(id string) ([]byte, error) {
	ks.mu.RLock()
	encryptedKey, exists := ks.keys[id]
	ks.mu.RUnlock()
	
	if !exists {
		return nil, errors.New("key not found")
	}

	block, err := aes.NewCipher(ks.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedKey) < nonceSize {
		return nil, errors.New("invalid key data")
	}

	nonce, ciphertext := encryptedKey[:nonceSize], encryptedKey[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
