package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/devinjacknz/devintrade/internal/trading"
	"github.com/devinjacknz/devintrade/internal/wallet"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	router      *mux.Router
	trading     trading.Engine
	walletMgr   wallet.Manager
	jwtSecret   []byte
	upgrader    websocket.Upgrader
}

func NewServer(trading trading.Engine, walletMgr wallet.Manager, jwtSecret []byte) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		trading:   trading,
		walletMgr: walletMgr,
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure based on environment
			},
		},
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Auth routes
	s.router.HandleFunc("/api/auth/login", s.handleLogin).Methods("POST")
	
	// Protected routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(s.authMiddleware)
	
	// Trading routes
	api.HandleFunc("/orders", s.handlePlaceOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", s.handleCancelOrder).Methods("DELETE")
	
	// WebSocket routes
	api.HandleFunc("/ws/prices", s.handlePriceUpdates)
	api.HandleFunc("/ws/positions", s.handlePositionUpdates)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement actual authentication logic
	if creds.Username != "admin" || creds.Password != "password" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": creds.Username,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var order trading.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.trading.PlaceOrder(order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	symbol := r.URL.Query().Get("symbol")

	if err := s.trading.CancelOrder(orderID, symbol); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePriceUpdates(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// TODO: Implement price update streaming
	for {
		if err := conn.WriteJSON(map[string]interface{}{
			"symbol": "SOL/USD",
			"price":  100.0,
			"time":   time.Now(),
		}); err != nil {
			break
		}
		time.Sleep(time.Second)
	}
}

func (s *Server) handlePositionUpdates(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// TODO: Implement position update streaming
	for {
		if err := conn.WriteJSON(map[string]interface{}{
			"symbol":      "SOL/USD",
			"size":        10.0,
			"entryPrice": 100.0,
			"pnl":        50.0,
		}); err != nil {
			break
		}
		time.Sleep(time.Second)
	}
}
