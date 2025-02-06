package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	JupiterBaseURL = "https://api.jup.ag"
	QuoteEndpoint  = "/swap/v1/quote"
	SwapEndpoint   = "/swap/v1/swap"
	PriceEndpoint  = "/price/v2"
)

type JupiterDEX struct {
	client     *RateLimitedClient
	name       string
	tokenCache *TokenCache
	updateMu   sync.Mutex
}

func NewJupiterDEX() *JupiterDEX {
	return &JupiterDEX{
		client: NewRateLimitedClient(1.0), // 1 request per second for free plan
		name:   "Jupiter",
		tokenCache: &TokenCache{
			tokens: make(map[string]TokenInfo),
		},
	}
}

func (j *JupiterDEX) Name() string {
	return j.name
}

func (j *JupiterDEX) GetMarketData() ([]*MarketData, error) {
	if err := j.updateTokenList(); err != nil {
		return nil, fmt.Errorf("failed to update token list: %w", err)
	}

	j.tokenCache.mu.RLock()
	tokens := make([]TokenInfo, 0, len(j.tokenCache.tokens))
	for _, token := range j.tokenCache.tokens {
		tokens = append(tokens, token)
	}
	j.tokenCache.mu.RUnlock()

	var marketData []*MarketData
	for _, token := range tokens {
		data, err := j.getTokenMarketData(token.Mint)
		if err != nil {
			continue // Skip failed tokens but continue
		}
		marketData = append(marketData, data)
		time.Sleep(time.Second) // Respect rate limit
	}
	return marketData, nil
}

func (j *JupiterDEX) getTokenMarketData(mint string) (*MarketData, error) {
	url := fmt.Sprintf("%s%s?inputMint=%s&outputMint=%s", 
		JupiterBaseURL, PriceEndpoint, mint,
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
	
	resp, err := j.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data for %s: %w", mint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code for %s: %d", mint, resp.StatusCode)
	}

	var priceData struct {
		Data struct {
			Price  float64 `json:"price"`
			Volume float64 `json:"volume24h"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		return nil, fmt.Errorf("failed to decode market data for %s: %w", mint, err)
	}

	j.tokenCache.mu.RLock()
	tokenInfo, exists := j.tokenCache.tokens[mint]
	j.tokenCache.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("token info not found for %s", mint)
	}

	return &MarketData{
		Symbol: tokenInfo.Symbol,
		Price:  priceData.Data.Price,
		Volume: priceData.Data.Volume,
	}, nil
}

func (j *JupiterDEX) GetMarketPrice(symbol string) (float64, error) {
	// Convert symbol to mint addresses (e.g., "SOL/USDC" -> solMint/usdcMint)
	inputMint, outputMint := "So11111111111111111111111111111111111111112", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	
	url := fmt.Sprintf("%s%s?inputMint=%s&outputMint=%s", JupiterBaseURL, PriceEndpoint, inputMint, outputMint)
	resp, err := j.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, symbol)
	}

	var priceData struct {
		Data struct {
			Price float64 `json:"price"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		return 0, fmt.Errorf("failed to decode price data for %s: %w", symbol, err)
	}

	return priceData.Data.Price, nil
}

func (j *JupiterDEX) updateTokenList() error {
	j.updateMu.Lock()
	defer j.updateMu.Unlock()

	// Check if update needed (every 1 hour)
	if time.Since(j.tokenCache.updatedAt) < time.Hour {
		return nil
	}

	resp, err := j.client.Get("https://token.jup.ag/strict")
	if err != nil {
		return fmt.Errorf("failed to fetch token list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokens []TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to decode token list: %w", err)
	}

	// Sort by volume and take top 30
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Volume24h > tokens[j].Volume24h
	})
	if len(tokens) > 30 {
		tokens = tokens[:30]
	}

	j.tokenCache.mu.Lock()
	defer j.tokenCache.mu.Unlock()
	j.tokenCache.tokens = make(map[string]TokenInfo)
	for _, token := range tokens {
		j.tokenCache.tokens[token.Mint] = token
	}
	j.tokenCache.updatedAt = time.Now()
	return nil
}

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	// Get quote first
	quoteURL := fmt.Sprintf("%s%s", JupiterBaseURL, QuoteEndpoint)
	quoteReq := JupiterQuoteRequest{
		InputMint:   "So11111111111111111111111111111111111111112", // SOL
		OutputMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		Amount:      fmt.Sprintf("%.0f", order.Amount * 1e9), // Convert to lamports
		SlippageBps: 100, // 1% slippage
	}

	quoteBody, err := json.Marshal(quoteReq)
	if err != nil {
		return fmt.Errorf("failed to marshal quote request: %w", err)
	}

	resp, err := j.client.Post(quoteURL, "application/json", quoteBody)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for quote: %d", resp.StatusCode)
	}

	var quoteResp JupiterQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return fmt.Errorf("failed to decode quote response: %w", err)
	}

	// Execute swap
	swapURL := fmt.Sprintf("%s%s", JupiterBaseURL, SwapEndpoint)
	swapReq := JupiterSwapRequest{
		QuoteResponse:  quoteResp,
		UserPublicKey:  os.Getenv("wallet"),
	}

	swapBody, err := json.Marshal(swapReq)
	if err != nil {
		return fmt.Errorf("failed to marshal swap request: %w", err)
	}

	resp, err = j.client.Post(swapURL, "application/json", swapBody)
	if err != nil {
		return fmt.Errorf("failed to execute swap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for swap: %d", resp.StatusCode)
	}

	var swapResp JupiterSwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResp); err != nil {
		return fmt.Errorf("failed to decode swap response: %w", err)
	}

	return nil
}
