package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

const (
	maxRequestsPerSecond = 1 // Free plan limit
	requestTimeout      = 10 * time.Second
	retryAttempts      = 5 // Increased retries
	retryDelay         = 500 * time.Millisecond // Reduced delay for faster recovery
	maxBackoff         = 5 * time.Second // Maximum backoff time
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
	SlippageBps  int    `json:"slippageBps"`
}

type SwapResponse struct {
	TxHash string `json:"txHash"`
}

func NewJupiterDEX() *JupiterDEX {
	return &JupiterDEX{
		limiter: rate.NewLimiter(rate.Every(time.Second/maxRequestsPerSecond), 1),
		client:  &http.Client{Timeout: requestTimeout},
	}
}

func (j *JupiterDEX) GetQuote(ctx context.Context, inputMint, outputMint string, amount string) (*QuoteResponse, error) {
	log.Printf("[JUPITER] Requesting quote: input=%s output=%s amount=%s", inputMint, outputMint, amount)

	if err := j.limiter.Wait(ctx); err != nil {
		log.Printf("[JUPITER] Rate limit exceeded for quote request: %v", err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= retryAttempts; attempt++ {

	reqBody := &QuoteRequest{
		InputMint:  inputMint,
		OutputMint: outputMint,
		Amount:     amount,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[JUPITER] Failed to marshal quote request: %v", err)
		return nil, fmt.Errorf("failed to marshal quote request: %w", err)
	}

	log.Printf("[JUPITER] Starting quote request with wallet: %s", os.Getenv("WALLET"))

		resp, err := j.client.Post("https://quote-api.jup.ag/v1/quote", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to get quote: %w", err)
			backoff := time.Duration(attempt) * retryDelay
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("[JUPITER] Quote attempt %d/%d failed: %v, retrying in %v", 
				attempt, retryAttempts, err, backoff)
			time.Sleep(backoff)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("quote API returned status %d: %s", resp.StatusCode, respBody)
			continue
		}

		var quote QuoteResponse
		if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
			lastErr = fmt.Errorf("failed to decode quote response: %w", err)
			continue
		}

		log.Printf("[JUPITER] Quote successful on attempt %d/%d", attempt, retryAttempts)
		return &quote, nil
	}
	
	log.Printf("[JUPITER] All quote attempts failed: %v", lastErr)
	return nil, fmt.Errorf("failed to get quote after %d attempts: %w", retryAttempts, lastErr)

	log.Printf("[JUPITER] Successfully received quote: input=%s output=%s price=%.4f", 
		quote.InputMint, quote.OutputMint, quote.Price)
	return &quote, nil
}

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	ctx := context.Background()
	log.Printf("[JUPITER] Starting order execution: %+v", order)

	if err := j.limiter.Wait(ctx); err != nil {
		log.Printf("[JUPITER] Rate limit exceeded: %v", err)
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Get quote for the order
	quote, err := j.GetQuote(ctx, order.Symbol, "USDC", fmt.Sprintf("%.0f", order.Amount))
	if err != nil {
		log.Printf("[JUPITER] Failed to get quote: %v", err)
		return fmt.Errorf("failed to get quote: %w", err)
	}
	log.Printf("[JUPITER] Received quote: inAmount=%s outAmount=%s price=%.4f", 
		quote.InAmount, quote.OutAmount, quote.Price)

	// Execute swap with wallet
	swapReq := &SwapRequest{
		QuoteResponse:  *quote,
		UserPublicKey: order.Wallet,
		SlippageBps:  50, // 0.5% slippage tolerance
	}

	body, err := json.Marshal(swapReq)
	if err != nil {
		log.Printf("[JUPITER] Failed to marshal swap request: %v", err)
		return fmt.Errorf("failed to marshal swap request: %w", err)
	}

	var (
		lastErr error
		swapResult SwapResponse
	)

	for attempt := 1; attempt <= retryAttempts; attempt++ {
		var resp *http.Response
		resp, err = j.client.Post("https://swap-api.jup.ag/v1/swap", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			backoff := time.Duration(attempt) * retryDelay
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("[JUPITER] Swap attempt %d/%d failed: %v, retrying in %v", 
				attempt, retryAttempts, err, backoff)
			time.Sleep(backoff)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("swap API returned status %d: %s", resp.StatusCode, respBody)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&swapResult); err != nil {
			lastErr = fmt.Errorf("failed to decode swap response: %w", err)
			continue
		}

		log.Printf("[JUPITER] Swap executed successfully: txHash=%s", swapResult.TxHash)
		return nil
	}

	log.Printf("[JUPITER] All swap attempts failed: %v", lastErr)
	return fmt.Errorf("failed to execute swap after %d attempts: %w", retryAttempts, lastErr)
	defer resp.Body.Close()

	var swapResult SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResult); err != nil {
		log.Printf("[JUPITER] Failed to decode swap response: %v", err)
		return fmt.Errorf("failed to decode swap response: %w", err)
	}
	log.Printf("[JUPITER] Swap executed successfully: txHash=%s", swapResult.TxHash)

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
