package exchange

import (
	"sync"
	"time"
)

type JupiterQuoteRequest struct {
	InputMint   string `json:"inputMint"`
	OutputMint  string `json:"outputMint"`
	Amount      string `json:"amount"`
	SlippageBps int    `json:"slippageBps"`
}

type JupiterQuoteResponse struct {
	InputAmount    string       `json:"inputAmount"`
	OutputAmount   string       `json:"outputAmount"`
	PriceImpactPct float64     `json:"priceImpactPct"`
	MarketInfos    []MarketInfo `json:"marketInfos"`
}

type MarketInfo struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	InAmount  string `json:"inAmount"`
	OutAmount string `json:"outAmount"`
}

type JupiterSwapRequest struct {
	QuoteResponse JupiterQuoteResponse `json:"quoteResponse"`
	UserPublicKey string              `json:"userPublicKey"`
}

type JupiterSwapResponse struct {
	SwapTransaction string `json:"swapTransaction"`
	Message         string `json:"message,omitempty"`
}

type TokenInfo struct {
	Symbol    string  `json:"symbol"`
	Mint      string  `json:"mint"`
	Volume24h float64 `json:"volume24h"`
}

type TokenList struct {
	Tokens    []TokenInfo
	UpdatedAt time.Time
}

type TokenCache struct {
	mu        sync.RWMutex
	tokens    map[string]TokenInfo  // mint -> info
	updatedAt time.Time
}
