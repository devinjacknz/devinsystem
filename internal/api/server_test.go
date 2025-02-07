package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/internal/trading"
	"github.com/devinjacknz/devinsystem/internal/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTradingEngine struct {
	mock.Mock
}

func (m *MockTradingEngine) PlaceOrder(order trading.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockTradingEngine) CancelOrder(orderID string, symbol string) error {
	args := m.Called(orderID, symbol)
	return args.Error(0)
}

func (m *MockTradingEngine) Start() error {
	args := m.Called()
	return args.Error(0)
}

type MockWalletManager struct {
	mock.Mock
}

func (m *MockWalletManager) CreateWallet(walletType wallet.WalletType) error {
	args := m.Called(walletType)
	return args.Error(0)
}

func (m *MockWalletManager) GetWallet(walletType wallet.WalletType) (*wallet.SolanaWallet, error) {
	args := m.Called(walletType)
	return args.Get(0).(*wallet.SolanaWallet), args.Error(1)
}

func TestServer_HandleLogin(t *testing.T) {
	mockTrading := new(MockTradingEngine)
	mockWallet := new(MockWalletManager)
	server := NewServer(mockTrading, mockWallet, []byte("test-secret"))

	tests := []struct {
		name           string
		credentials    map[string]string
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid credentials",
			credentials: map[string]string{
				"username": "admin",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid credentials",
			credentials: map[string]string{
				"username": "wrong",
				"password": "wrong",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.credentials)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			server.handleLogin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectToken {
				var response map[string]string
				json.NewDecoder(w.Body).Decode(&response)
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"])
			}
		})
	}
}

func TestServer_HandlePlaceOrder(t *testing.T) {
	mockTrading := new(MockTradingEngine)
	mockWallet := new(MockWalletManager)
	server := NewServer(mockTrading, mockWallet, []byte("test-secret"))

	order := trading.Order{
		Symbol:    "SOL/USD",
		Side:      "buy",
		Amount:    1.0,
		Price:     100.0,
		OrderType: "limit",
	}

	mockTrading.On("PlaceOrder", order).Return(nil)

	body, _ := json.Marshal(order)
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", createTestToken(t, server.jwtSecret))
	w := httptest.NewRecorder()

	server.handlePlaceOrder(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockTrading.AssertExpectations(t)
}

func TestServer_HandleCancelOrder(t *testing.T) {
	mockTrading := new(MockTradingEngine)
	mockWallet := new(MockWalletManager)
	server := NewServer(mockTrading, mockWallet, []byte("test-secret"))

	orderID := "test-order"
	symbol := "SOL/USD"

	mockTrading.On("CancelOrder", orderID, symbol).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/orders/"+orderID+"?symbol="+symbol, nil)
	req.Header.Set("Authorization", createTestToken(t, server.jwtSecret))
	w := httptest.NewRecorder()

	// Set up router to handle path parameters
	router := server.router
	router.HandleFunc("/api/orders/{id}", server.handleCancelOrder).Methods("DELETE")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTrading.AssertExpectations(t)
}

func TestServer_HandleWebSocket(t *testing.T) {
	mockTrading := new(MockTradingEngine)
	mockWallet := new(MockWalletManager)
	server := NewServer(mockTrading, mockWallet, []byte("test-secret"))

	s := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	header := http.Header{}
	header.Add("Authorization", createTestToken(t, server.jwtSecret))

	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	assert.NoError(t, err)
	defer ws.Close()

	tests := []struct {
		name    string
		message map[string]interface{}
		want    map[string]interface{}
	}{
		{
			name: "subscribe to market data",
			message: map[string]interface{}{
				"type":   "subscribe",
				"symbol": "SOL/USD",
			},
			want: map[string]interface{}{
				"type":    "subscribed",
				"symbol":  "SOL/USD",
				"status":  "success",
			},
		},
		{
			name: "invalid subscription message",
			message: map[string]interface{}{
				"type": "subscribe",
			},
			want: map[string]interface{}{
				"type":    "error",
				"message": "invalid subscription request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.WriteJSON(tt.message)
			assert.NoError(t, err)

			var response map[string]interface{}
			err = ws.ReadJSON(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.want["type"], response["type"])
		})
	}
}

func TestServer_AuthMiddleware(t *testing.T) {
	mockTrading := new(MockTradingEngine)
	mockWallet := new(MockWalletManager)
	server := NewServer(mockTrading, mockWallet, []byte("test-secret"))

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          createTestToken(t, server.jwtSecret),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			token:          "Invalid-Token-Format",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			token:          createExpiredToken(t, server.jwtSecret),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			server.authMiddleware(handler).ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func createTestToken(t *testing.T, secret []byte) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "test",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)
	return "Bearer " + tokenString
}

func createExpiredToken(t *testing.T, secret []byte) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "test",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)
	return "Bearer " + tokenString
}
