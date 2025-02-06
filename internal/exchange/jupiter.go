package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	JupiterBaseURL = "https://api.jup.ag"
	QuoteEndpoint  = "/swap/v1/quote"
	SwapEndpoint   = "/swap/v1/swap"
	PriceEndpoint  = "/price/v2"
)

type JupiterDEX struct {
	client  *http.Client
	name    string
}

func NewJupiterDEX() *JupiterDEX {
	return &JupiterDEX{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		name: "Jupiter",
	}
}

func (j *JupiterDEX) Name() string {
	return j.name
}

func (j *JupiterDEX) GetMarketData() (*MarketData, error) {
	// Get market data for SOL/USDC as default pair
	const (
		solMint = "So11111111111111111111111111111111111111112"
		usdcMint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	)
	
	url := fmt.Sprintf("%s%s?inputMint=%s&outputMint=%s", JupiterBaseURL, PriceEndpoint, solMint, usdcMint)
	resp, err := j.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var priceData struct {
		Data struct {
			Price  float64 `json:"price"`
			Volume float64 `json:"volume24h"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		return nil, fmt.Errorf("failed to decode market data: %w", err)
	}

	return &MarketData{
		Symbol: "SOL/USDC",
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

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	quoteURL := fmt.Sprintf("%s%s", JupiterBaseURL, QuoteEndpoint)
	swapURL := fmt.Sprintf("%s%s", JupiterBaseURL, SwapEndpoint)

	// Implementation will be expanded in next steps
	return fmt.Errorf("not implemented")
}
