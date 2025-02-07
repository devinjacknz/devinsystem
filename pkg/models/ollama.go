package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
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

func (c *OllamaClient) GenerateTradeDecision(ctx context.Context, data interface{}) (*TradeDecision, error) {
	start := time.Now()
	defer func() {
		log.Printf("%s AI decision generation took %v", logging.LogMarkerPerf, time.Since(start))
	}()
	marketData, ok := data.(*market.MarketData)
	if !ok {
		log.Printf("%s Invalid data type provided to AI model", logging.LogMarkerError)
		return nil, fmt.Errorf("invalid data type: expected *market.MarketData")
	}
	systemPrompt := `You are an aggressive trading bot focused on meme coins. Given market data, respond ONLY with one of these exact formats:
BUY
{confidence}
{reasoning}

OR

SELL
{confidence}
{reasoning}

OR

NOTHING
0
No trade opportunity

Confidence must be between 0 and 1. For volatile meme coins:
- BUY when price is rising with increasing volume
- SELL when price is dropping or volume decreasing
- Use higher confidence (0.6-0.8) for strong trends
- Use lower confidence (0.3-0.5) for early trends`

	prompt := fmt.Sprintf(`Market Data:
Symbol: %s
Price: %.8f
Volume: %.2f
Timestamp: %s`, marketData.Symbol, marketData.Price, marketData.Volume, marketData.Timestamp.Format(time.RFC3339))

	request := ollamaRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Options: Options{
			Temperature: 0.7,
		},
	}

	// Try to load model first
	loadBody, err := json.Marshal(map[string]string{"name": c.model})
	if err != nil {
		log.Printf("%s Failed to marshal model load request: %v", logging.LogMarkerError, err)
		return nil, err
	}

	loadReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewReader(loadBody))
	if err != nil {
		log.Printf("%s Failed to create model load request: %v", logging.LogMarkerError, err)
		return nil, err
	}
	loadReq.Header.Set("Content-Type", "application/json")

	log.Printf("%s Loading model %s...", logging.LogMarkerAI, c.model)
	loadResp, err := c.httpClient.Do(loadReq)
	if err != nil {
		log.Printf("%s Failed to load model: %v", logging.LogMarkerError, err)
		return nil, err
	}
	loadResp.Body.Close()

	log.Printf("%s Generating trade decision for %s using %s model", logging.LogMarkerAI, marketData.Symbol, c.model)
	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("%s Failed to marshal Ollama request: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
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
		log.Printf("%s Ollama API returned non-200 status: %d", logging.LogMarkerError, resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("%s Failed to decode Ollama response: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	decision, confidence, reasoning := parseTradeDecision(response.Message.Content)
	tradeDecision := &TradeDecision{
		Action:     decision,
		Confidence: confidence,
		Reasoning:  reasoning,
		Model:     c.model,
		Timestamp: time.Now(),
	}

	log.Printf("%s Generated decision for %s: action=%s confidence=%.2f", logging.LogMarkerAI, 
		marketData.Symbol, tradeDecision.Action, tradeDecision.Confidence)
	return tradeDecision, nil
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
	if confidence == 0 && len(bytes.TrimSpace(line)) > 0 {
			fmt.Sscanf(string(bytes.TrimSpace(line)), "%f", &confidence)
		}
	}

	if len(lines) > 1 {
		reasoning = string(bytes.Join(lines[1:], []byte("\n")))
	}

	return decision, confidence, reasoning
}
