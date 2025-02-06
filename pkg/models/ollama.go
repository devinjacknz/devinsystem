package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

type ollamaRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	Options   Options   `json:"options"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Options struct {
	Temperature float64 `json:"temperature"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OllamaClient) GenerateTradeDecision(ctx context.Context, data *market.MarketData) (*TradeDecision, error) {
	systemPrompt := `You are a trading assistant. Analyze the market data and make a trading decision.
Consider:
1. Price trends
2. Volume patterns
3. Risk factors

Respond with one of: BUY, SELL, or NOTHING followed by your reasoning.
Include a confidence score (0-100) in your analysis.`

	marketData := fmt.Sprintf(`Market Data:
Symbol: %s
Price: %.8f
Volume: %.2f
Timestamp: %s`, data.Symbol, data.Price, data.Volume, data.Timestamp.Format(time.RFC3339))

	request := ollamaRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: marketData},
		},
		Stream: false,
		Options: Options{
			Temperature: 0.7,
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	decision, confidence, reasoning := parseTradeDecision(response.Message.Content)
	return &TradeDecision{
		Action:     decision,
		Confidence: confidence,
		Reasoning:  reasoning,
		Metadata: map[string]interface{}{
			"model":      c.model,
			"timestamp": time.Now(),
		},
	}, nil
}

func parseTradeDecision(content string) (string, float64, string) {
	lines := bytes.Split([]byte(content), []byte("\n"))
	if len(lines) == 0 {
		return "NOTHING", 0, "No decision could be parsed"
	}

	decision := string(bytes.TrimSpace(lines[0]))
	switch decision {
	case "BUY", "SELL", "NOTHING":
	default:
		decision = "NOTHING"
	}

	var confidence float64
	var reasoning string

	for _, line := range lines[1:] {
		if bytes.Contains(line, []byte("confidence")) {
			fmt.Sscanf(string(line), "confidence: %f", &confidence)
		}
	}

	if len(lines) > 1 {
		reasoning = string(bytes.Join(lines[1:], []byte("\n")))
	}

	return decision, confidence, reasoning
}
