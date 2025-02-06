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
	url := fmt.Sprintf("%s%s", JupiterBaseURL, PriceEndpoint)
	resp, err := j.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}
	defer resp.Body.Close()

	var priceData struct {
		Price  float64 `json:"price"`
		Volume float64 `json:"volume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		return nil, fmt.Errorf("failed to decode market data: %w", err)
	}

	return &MarketData{
		Price:  priceData.Price,
		Volume: priceData.Volume,
	}, nil
}

func (j *JupiterDEX) GetMarketPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s%s/%s", JupiterBaseURL, PriceEndpoint, symbol)
	resp, err := j.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	var priceData struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&priceData); err != nil {
		return 0, fmt.Errorf("failed to decode price: %w", err)
	}

	return priceData.Price, nil
}

func (j *JupiterDEX) ExecuteOrder(order Order) error {
	quoteURL := fmt.Sprintf("%s%s", JupiterBaseURL, QuoteEndpoint)
	swapURL := fmt.Sprintf("%s%s", JupiterBaseURL, SwapEndpoint)

	// Implementation will be expanded in next steps
	return fmt.Errorf("not implemented")
}
