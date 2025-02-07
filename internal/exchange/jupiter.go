package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

type JupiterDEX struct {
	limiter *rate.Limiter
	client  *http.Client
}

type QuoteRequest struct {
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	Amount     string `json:"amount"`
}

type QuoteResponse struct {
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	Price      float64 `json:"price"`
}

type SwapRequest struct {
	QuoteResponse
	UserPublicKey string `json:"userPublicKey"`
}

type SwapResponse struct {
	TxHash string `json:"txHash"`
}

func NewJupiterDEX() *JupiterDEX {
	return &JupiterDEX{
		limiter: rate.NewLimiter(rate.Every(time.Second), 1), // 1 RPS
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (j *JupiterDEX) GetQuote(ctx context.Context, inputMint, outputMint string, amount string) (*QuoteResponse, error) {
	if err := j.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	reqBody := &QuoteRequest{
		InputMint:  inputMint,
		OutputMint: outputMint,
		Amount:     amount,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal quote request: %w", err)
	}

	resp, err := j.client.Post("https://quote-api.jup.ag/v1/quote", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	defer resp.Body.Close()

	var quote QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	return &quote, nil
}

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	ctx := context.Background()
	if err := j.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Get quote for the order
	quote, err := j.GetQuote(ctx, order.Symbol, "USDC", fmt.Sprintf("%.0f", order.Amount))
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	// Execute swap with wallet
	swapReq := &SwapRequest{
		QuoteResponse:  *quote,
		UserPublicKey: os.Getenv("WALLET"),
	}

	body, err := json.Marshal(swapReq)
	if err != nil {
		return fmt.Errorf("failed to marshal swap request: %w", err)
	}

	resp, err := j.client.Post("https://swap-api.jup.ag/v1/swap", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to execute swap: %w", err)
	}
	defer resp.Body.Close()

	var swapResult SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResult); err != nil {
		return fmt.Errorf("failed to decode swap response: %w", err)
	}

	return nil
}

func (j *JupiterDEX) GetMarketPrice(token string) (float64, error) {
	ctx := context.Background()
	if err := j.limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limit exceeded: %w", err)
	}

	resp, err := j.client.Get(fmt.Sprintf("https://price-api.jup.ag/v1/price/%s", token))
	if err != nil {
		return 0, fmt.Errorf("failed to get market price: %w", err)
	}
	defer resp.Body.Close()

	var price struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return 0, fmt.Errorf("failed to decode price response: %w", err)
	}

	return price.Price, nil
}
