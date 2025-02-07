package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaClient struct {
	endpoint string
	model    string
}

type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func NewOllamaClient(endpoint, model string) *OllamaClient {
	return &OllamaClient{
		endpoint: endpoint,
		model:    model,
	}
}

func (c *OllamaClient) AnalyzeMarket(data MarketData) (*Analysis, error) {
	prompt := fmt.Sprintf(
		"Analyze the following market data for %s:\nPrice: %.2f\nVolume: %.2f\nTrend: %s\n",
		data.Symbol, data.Price, data.Volume, data.Trend,
	)

	req := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
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

	return &Analysis{
		Symbol:     data.Symbol,
		Action:     "BUY",
		Confidence: 0.8,
		Reasoning:  result.Response,
		Model:      c.model,
		Timestamp:  time.Now(),
		Signals: []Signal{{
			Type:       "TREND",
			Symbol:     data.Symbol,
			Action:     "BUY",
			Confidence: 0.8,
		}},
	}, nil
}
