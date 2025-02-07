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

	"github.com/devinjacknz/devinsystem/pkg/utils"
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
	log.Printf("%s Requesting quote: input=%s output=%s amount=%s", utils.LogMarkerTrade, inputMint, outputMint, amount)

	if err := j.limiter.Wait(ctx); err != nil {
		log.Printf("%s Rate limit exceeded for quote request: %v", utils.LogMarkerError, err)
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
		log.Printf("%s Failed to marshal quote request: %v", utils.LogMarkerError, err)
		return nil, fmt.Errorf("failed to marshal quote request: %w", err)
	}

	log.Printf("%s Starting quote request with wallet: %s", utils.LogMarkerWallet, os.Getenv("WALLET"))

		resp, err := j.client.Post("https://quote-api.jup.ag/v6/quote", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to get quote: %w", err)
			backoff := time.Duration(attempt) * retryDelay
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("%s Quote attempt %d/%d failed: %v, retrying in %v", utils.LogMarkerRetry, 
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

		log.Printf("%s Quote successful on attempt %d/%d", utils.LogMarkerTrade, attempt, retryAttempts)
		return &quote, nil
	}
	
	log.Printf("%s All quote attempts failed: %v", utils.LogMarkerError, lastErr)
	return nil, fmt.Errorf("failed to get quote after %d attempts: %w", retryAttempts, lastErr)
}

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	start := time.Now()
	ctx := context.Background()
	log.Printf("%s Starting order execution: %+v", utils.LogMarkerTrade, order)
	
	defer func() {
		log.Printf("%s Order execution took %v", utils.LogMarkerPerf, time.Since(start))
	}()

	if order.Amount <= 0 || order.Price <= 0 {
		log.Printf("%s Invalid order amount or price: amount=%.8f price=%.8f", utils.LogMarkerError, order.Amount, order.Price)
		return fmt.Errorf("invalid order amount or price")
	}

	if err := j.limiter.Wait(ctx); err != nil {
		log.Printf("%s Rate limit exceeded: %v", utils.LogMarkerError, err)
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Get quote and execute swap
	swapReq := &SwapRequest{
		QuoteResponse: QuoteResponse{
			InputMint:  order.Symbol,
			OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
			InAmount:   fmt.Sprintf("%.0f", order.Amount * 1e9), // Convert to lamports
		},
		UserPublicKey: order.Wallet,
		SlippageBps:  100, // 1% slippage tolerance
	}

	body, err := json.Marshal(swapReq)
	if err != nil {
		log.Printf("%s Failed to marshal swap request: %v", utils.LogMarkerError, err)
		return fmt.Errorf("failed to marshal swap request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		resp, err := j.client.Post("https://quote-api.jup.ag/v6/swap", "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			backoff := time.Duration(attempt) * retryDelay
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("%s Swap attempt %d/%d failed: %v, retrying in %v", utils.LogMarkerRetry, 
				attempt, retryAttempts, err, backoff)
			time.Sleep(backoff)
			continue
		}

		var result SwapResponse
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("swap API returned status %d: %s", resp.StatusCode, respBody)
			continue
		}

		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to decode swap response: %w", err)
			continue
		}

		log.Printf("%s Swap executed successfully: txHash=%s", utils.LogMarkerTrade, result.TxHash)
		return nil
	}

	log.Printf("%s All swap attempts failed: %v", utils.LogMarkerError, lastErr)
	return fmt.Errorf("failed to execute swap after %d attempts: %w", retryAttempts, lastErr)
}

func (j *JupiterDEX) GetMarketPrice(token string) (float64, error) {
	ctx := context.Background()
	if err := j.limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		var httpResp *http.Response
		httpResp, err := j.client.Get(fmt.Sprintf("https://price-api.jup.ag/v6/price?ids=%s&vsToken=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", token))
		if err != nil {
			lastErr = err
			backoff := time.Duration(attempt) * retryDelay
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("%s Price fetch attempt %d/%d failed: %v, retrying in %v", utils.LogMarkerRetry,
				attempt, retryAttempts, err, backoff)
			time.Sleep(backoff)
			continue
		}

		var priceResp struct {
			Data map[string]struct {
				Price float64 `json:"price"`
			} `json:"data"`
		}
		err = json.NewDecoder(httpResp.Body).Decode(&priceResp)
		httpResp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to decode price response: %w", err)
			continue
		}

		if priceData, ok := priceResp.Data[token]; ok {
			if priceData.Price <= 0 {
				lastErr = fmt.Errorf("invalid price: %.8f", priceData.Price)
				continue
			}
			log.Printf("%s Retrieved price for %s: %.8f USDC", utils.LogMarkerMarket, token, priceData.Price)
			return priceData.Price, nil
		}
		lastErr = fmt.Errorf("token %s not found in price response", token)
		continue
	}

	return 0, fmt.Errorf("failed to get market price after %d attempts: %w", retryAttempts, lastErr)
}
