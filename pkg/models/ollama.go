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
	systemPrompt := `You are an aggressive meme coin trading bot. Given market data, respond ONLY with one of these exact formats:
BUY
{confidence}
{reasoning}

OR

SELL
{confidence}
{reasoning}

OR

NOTHING
0.1
Market conditions unfavorable

Confidence must be between 0.1 and 0.9. For meme coins:
- BUY aggressively (0.4-0.6 confidence) on volume spikes >10%
- BUY strongly (0.6-0.8 confidence) on price increase >2% with volume
- SELL quickly (0.5-0.7 confidence) on volume drop >5%
- SELL immediately (0.7-0.9 confidence) on price drop >1%
- Use 0.1-0.3 confidence for NOTHING decisions
- Focus on momentum and quick profits
- Look for scalping opportunities in volatile markets`

	prompt := fmt.Sprintf(`Analyze this real-time market data and make an aggressive trading decision:
Token: %s
Current Price: %.8f SOL
24h Volume: %.2f SOL
Last Update: %s

Consider:
- Volume spikes indicate potential momentum
- Price movements >2%% warrant action
- Look for quick scalping opportunities
- Be aggressive with meme tokens
- Use tight stop losses`, 
		marketData.Symbol, marketData.Price, marketData.Volume, 
		marketData.Timestamp.Format(time.RFC3339))

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

	log.Printf("%s Generated decision for %s: action=%s confidence=%.2f reasoning=%s", logging.LogMarkerAI, 
		marketData.Symbol, tradeDecision.Action, tradeDecision.Confidence, tradeDecision.Reasoning)
	return tradeDecision, nil
}

func parseTradeDecision(content string) (string, float64, string) {
	lines := bytes.Split([]byte(content), []byte("\n"))
	if len(lines) == 0 {
		return "NOTHING", 0.1, "No valid market data available for decision"
	}

	decision := string(bytes.TrimSpace(lines[0]))
	switch decision {
	case "BUY", "SELL":
		// Valid trading decision
	case "NOTHING":
		return "NOTHING", 0.1, "Market conditions unfavorable for trading"
	default:
		return "NOTHING", 0.1, "Invalid decision format received"
	}

	var confidence float64
	var reasoning string

	// Parse confidence from second line
	if len(lines) > 1 {
		confStr := string(bytes.TrimSpace(lines[1]))
		if _, err := fmt.Sscanf(confStr, "%f", &confidence); err != nil || confidence == 0 {
			confidence = 0.3 // Default to minimum trading confidence
		}
	}

	// Ensure confidence is within valid range
	switch {
	case confidence < 0.1:
		confidence = 0.1
	case confidence > 0.9:
		confidence = 0.9
	}

	// Extract reasoning from remaining lines
	if len(lines) > 2 {
		reasoningLines := lines[2:]
		reasoning = string(bytes.TrimSpace(bytes.Join(reasoningLines, []byte("\n"))))
		if reasoning == "" {
			reasoning = "Market analysis complete, confidence level indicates potential opportunity"
		}
	} else {
		reasoning = "Decision based on current market conditions"
	}

	return decision, confidence, reasoning
}
