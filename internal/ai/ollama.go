package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaClient struct {
	endpoint    string
	model       string
	temperature float64
}

type OllamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func NewOllamaClient(endpoint, model string, temperature float64) *OllamaClient {
	return &OllamaClient{
		endpoint:    endpoint,
		model:       model,
		temperature: temperature,
	}
}

func (c *OllamaClient) AnalyzeMarket(data MarketData) (*Analysis, error) {
	prompt := fmt.Sprintf(
		"Analyze the following market data for %s:\nPrice: %.2f\nVolume: %.2f\nTrend: %s\n",
		data.Symbol, data.Price, data.Volume, data.Trend,
	)

	req := OllamaRequest{
		Model:       c.model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: c.temperature,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.endpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Parse the response into an Analysis struct
	// This is a simplified example - in production we would use NLP to parse the response
	return &Analysis{
		Symbol:     data.Symbol,
		Trend:      data.Trend,
		Confidence: 0.8,
		Signals: []Signal{{
			Type:       "TREND",
			Action:     "BUY",
			Confidence: 0.8,
		}},
	}, nil
}
