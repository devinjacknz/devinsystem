package wallet

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) StoreKey(id string, key []byte) error {
	args := m.Called(id, key)
	return args.Error(0)
}

func (m *MockKeyStore) GetKey(id string) ([]byte, error) {
	args := m.Called(id)
	return args.Get(0).([]byte), args.Error(1)
}

func TestWalletManager_CreateWallet(t *testing.T) {
	mockKeyStore := new(MockKeyStore)
	manager, err := NewWalletManager()
	assert.NoError(t, err)
	manager.keyStore = mockKeyStore

	tests := []struct {
		name        string
		walletType  WalletType
		setupMock   func()
		wantErr     bool
		checkWallet func(*testing.T, *WalletManager)
	}{
		{
			name:      "create trading wallet",
			walletType: TradingWallet,
			setupMock: func() {
				mockKeyStore.On("StoreKey", "wallet-A", mock.Anything).Return(nil)
			},
			wantErr: false,
			checkWallet: func(t *testing.T, m *WalletManager) {
				wallet, err := m.GetWallet(TradingWallet)
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, "wallet-A", wallet.ID())
			},
		},
		{
			name:      "create profit wallet",
			walletType: ProfitWallet,
			setupMock: func() {
				mockKeyStore.On("StoreKey", "wallet-B", mock.Anything).Return(nil)
			},
			wantErr: false,
			checkWallet: func(t *testing.T, m *WalletManager) {
				wallet, err := m.GetWallet(ProfitWallet)
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, "wallet-B", wallet.ID())
			},
		},
		{
			name:      "duplicate wallet",
			walletType: TradingWallet,
			setupMock: func() {
				mockKeyStore.On("StoreKey", "wallet-A", mock.Anything).Return(nil)
			},
			wantErr: true,
			checkWallet: func(t *testing.T, m *WalletManager) {
				wallet, err := m.GetWallet(TradingWallet)
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := manager.CreateWallet(tt.walletType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkWallet != nil {
				tt.checkWallet(t, manager)
			}
		})
	}
}

func TestWalletManager_TransferFunds(t *testing.T) {
	mockKeyStore := new(MockKeyStore)
	manager, err := NewWalletManager()
	assert.NoError(t, err)
	manager.keyStore = mockKeyStore

	// Setup wallets
	mockKeyStore.On("StoreKey", "wallet-A", mock.Anything).Return(nil)
	mockKeyStore.On("StoreKey", "wallet-B", mock.Anything).Return(nil)
	err = manager.CreateWallet(TradingWallet)
	assert.NoError(t, err)
	err = manager.CreateWallet(ProfitWallet)
	assert.NoError(t, err)

	// Set initial balances
	tradingWallet, _ := manager.GetWallet(TradingWallet)
	tradingWallet.(*SolanaWallet).balance = 1000.0

	tests := []struct {
		name        string
		from        WalletType
		to          WalletType
		amount      float64
		wantErr     bool
		checkWallet func(*testing.T, *WalletManager)
	}{
		{
			name:    "transfer profits from trading to profit wallet",
			from:    TradingWallet,
			to:      ProfitWallet,
			amount:  100.0,
			wantErr: false,
			checkWallet: func(t *testing.T, m *WalletManager) {
				tradingWallet, _ := m.GetWallet(TradingWallet)
				profitWallet, _ := m.GetWallet(ProfitWallet)
				assert.Equal(t, 900.0, tradingWallet.(*SolanaWallet).balance)
				assert.Equal(t, 100.0, profitWallet.(*SolanaWallet).balance)
			},
		},
		{
			name:    "transfer with insufficient funds",
			from:    TradingWallet,
			to:      ProfitWallet,
			amount:  1500.0,
			wantErr: true,
			checkWallet: func(t *testing.T, m *WalletManager) {
				tradingWallet, _ := m.GetWallet(TradingWallet)
				profitWallet, _ := m.GetWallet(ProfitWallet)
				assert.Equal(t, 900.0, tradingWallet.(*SolanaWallet).balance)
				assert.Equal(t, 100.0, profitWallet.(*SolanaWallet).balance)
			},
		},
		{
			name:    "transfer with invalid source wallet",
			from:    "invalid",
			to:      ProfitWallet,
			amount:  100.0,
			wantErr: true,
		},
		{
			name:    "transfer with invalid destination wallet",
			from:    TradingWallet,
			to:      "invalid",
			amount:  100.0,
			wantErr: true,
		},
		{
			name:    "transfer zero amount",
			from:    TradingWallet,
			to:      ProfitWallet,
			amount:  0.0,
			wantErr: true,
			checkWallet: func(t *testing.T, m *WalletManager) {
				tradingWallet, _ := m.GetWallet(TradingWallet)
				profitWallet, _ := m.GetWallet(ProfitWallet)
				assert.Equal(t, 900.0, tradingWallet.(*SolanaWallet).balance)
				assert.Equal(t, 100.0, profitWallet.(*SolanaWallet).balance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.TransferFunds(tt.from, tt.to, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkWallet != nil {
				tt.checkWallet(t, manager)
			}
		})
	}
}

func TestWalletManager_GetWallet(t *testing.T) {
	mockKeyStore := new(MockKeyStore)
	manager, err := NewWalletManager()
	assert.NoError(t, err)
	manager.keyStore = mockKeyStore

	// Setup both A and B wallets
	mockKeyStore.On("StoreKey", "wallet-A", mock.Anything).Return(nil)
	mockKeyStore.On("StoreKey", "wallet-B", mock.Anything).Return(nil)
	err = manager.CreateWallet(TradingWallet)
	assert.NoError(t, err)
	err = manager.CreateWallet(ProfitWallet)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		walletType  WalletType
		wantErr     bool
		checkWallet func(*testing.T, Wallet)
	}{
		{
			name:      "get trading wallet (A)",
			walletType: TradingWallet,
			wantErr:   false,
			checkWallet: func(t *testing.T, w Wallet) {
				solanaWallet, ok := w.(*SolanaWallet)
				assert.True(t, ok)
				assert.Equal(t, "wallet-A", solanaWallet.ID())
				assert.NotEmpty(t, solanaWallet.GetAddress())
			},
		},
		{
			name:      "get profit wallet (B)",
			walletType: ProfitWallet,
			wantErr:   false,
			checkWallet: func(t *testing.T, w Wallet) {
				solanaWallet, ok := w.(*SolanaWallet)
				assert.True(t, ok)
				assert.Equal(t, "wallet-B", solanaWallet.ID())
				assert.NotEmpty(t, solanaWallet.GetAddress())
			},
		},
		{
			name:      "get non-existent wallet",
			walletType: "invalid",
			wantErr:   true,
		},
		{
			name:      "get wallet with empty type",
			walletType: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := manager.GetWallet(tt.walletType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				if tt.checkWallet != nil {
					tt.checkWallet(t, wallet)
				}
			}
		})
	}
}
