package api

import (
	"github.com/devinjacknz/devintrade/internal/trading"
	"github.com/devinjacknz/devintrade/internal/wallet"
	"github.com/gorilla/mux"
)

type Server struct {
	*mux.Router
	tradingEngine trading.Engine
	walletManager wallet.Manager
	jwtSecret     []byte
}

func NewServer(tradingEngine trading.Engine, walletManager wallet.Manager, jwtSecret []byte) *Server {
	s := &Server{
		Router:        mux.NewRouter(),
		tradingEngine: tradingEngine,
		walletManager: walletManager,
		jwtSecret:     jwtSecret,
	}

	s.Router.HandleFunc("/api/health", s.handleHealth).Methods("GET")
	return s
}
