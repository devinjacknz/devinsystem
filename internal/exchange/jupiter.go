package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	Price      string `json:"price"`
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

	req := &QuoteRequest{
		InputMint:  inputMint,
		OutputMint: outputMint,
		Amount:     amount,
	}

	resp, err := j.client.Post("https://quote-api.jup.ag/v1/quote", "application/json", nil)
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

func (j *JupiterDEX) ExecuteSwap(ctx context.Context, quote *QuoteResponse, wallet Wallet) error {
	if err := j.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	req := &SwapRequest{
		QuoteResponse:  *quote,
		UserPublicKey: wallet.GetPublicKey(),
	}

	resp, err := j.client.Post("https://swap-api.jup.ag/v1/swap", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to execute swap: %w", err)
	}
	defer resp.Body.Close()

	var swap SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swap); err != nil {
		return fmt.Errorf("failed to decode swap response: %w", err)
	}

	return nil
}

func (j *JupiterDEX) GetMarketPrice(ctx context.Context, token string) (float64, error) {
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
