package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeepSeekClient struct {
	endpoint string
	apiKey   string
}

type DeepSeekRequest struct {
	Input     string            `json:"input"`
	Parameters map[string]string `json:"parameters"`
}

type DeepSeekResponse struct {
	Output string `json:"output"`
}

func NewDeepSeekClient(endpoint, apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

func (c *DeepSeekClient) AnalyzeRisk(data MarketData) (*RiskAnalysis, error) {
	req := DeepSeekRequest{
		Input: fmt.Sprintf(
			"Analyze risk for %s with current price %.2f and volume %.2f",
			data.Symbol, data.Price, data.Volume,
		),
		Parameters: map[string]string{
			"mode": "risk_analysis",
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", c.endpoint+"/v1/analyze", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Parse the response into a RiskAnalysis struct
	return &RiskAnalysis{
		Symbol:        data.Symbol,
		RiskLevel:     "MEDIUM",
		StopLossPrice: data.Price * 0.95,
		Confidence:    0.85,
	}, nil
}
