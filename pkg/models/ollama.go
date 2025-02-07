package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	TopP        float64 `json:"top_p"`
	TopK        int     `json:"top_k"`
	Seed        int     `json:"seed"`
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
	systemPrompt := `You are a trading bot. Respond with exactly 3 lines:
Line 1: BUY or SELL or NOTHING
Line 2: A number between 0.1 and 0.9
Line 3: A reason

Example response:
BUY
0.6
Price up 2%

Do not include any other text. Only output these 3 lines.`

	prompt := fmt.Sprintf(`Analyze this data and respond with exactly 3 lines:
Token: %s
Price: %.8f SOL
Volume: %.2f SOL

Remember:
- Line 1: BUY/SELL/NOTHING
- Line 2: 0.1-0.9
- Line 3: reason`,
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
			Temperature: 0.1,  // Very low temperature for consistent format
			TopP:        0.1,  // Reduce randomness
			TopK:        10,   // Limit token choices
			Seed:        1234, // Fixed seed for reproducibility
		},
	}

	// Try to load model first
	loadBody, err := json.Marshal(map[string]interface{}{
		"name":    c.model,
		"stream":  false,
		"timeout": 30000, // 30 seconds
	})
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
	defer loadResp.Body.Close()

	// Read response body for debugging
	respBody, err := io.ReadAll(loadResp.Body)
	if err != nil {
		log.Printf("%s Failed to read model load response: %v", logging.LogMarkerError, err)
		return nil, err
	}
	log.Printf("%s Model load response: %s", logging.LogMarkerAI, string(respBody))

	log.Printf("%s Generating trade decision for %s using %s model", logging.LogMarkerAI, marketData.Symbol, c.model)
	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("%s Failed to marshal Ollama request: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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
	// Read and log raw response for debugging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s Failed to read response body: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	log.Printf("%s Raw API response: %s", logging.LogMarkerAI, string(respBody))

	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&response); err != nil {
		log.Printf("%s Failed to decode Ollama response: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("%s Model response content: %s", logging.LogMarkerAI, response.Message.Content)

	// Clean up response by removing any markdown formatting and extra whitespace
	cleanContent := strings.ReplaceAll(response.Message.Content, "```", "")
	cleanContent = strings.TrimSpace(cleanContent)

	// Split into lines and ensure we have exactly 3 non-empty lines
	lines := strings.Split(cleanContent, "\n")
	var validLines []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			validLines = append(validLines, trimmed)
		}
	}

	if len(validLines) < 3 {
		log.Printf("%s Invalid response format - expected 3 lines, got %d", logging.LogMarkerError, len(validLines))
		return &TradeDecision{
			Action:     "NOTHING",
			Confidence: 0.1,
			Reasoning:  fmt.Sprintf("Invalid response format - got %d lines", len(validLines)),
			Model:      c.model,
			Timestamp:  time.Now(),
		}, nil
	}

	// Parse the cleaned response using first 3 valid lines
	decision, confidence, reasoning := parseTradeDecision(strings.Join(validLines[:3], "\n"))
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
	if len(lines) < 3 {
		return "NOTHING", 0.1, "Insufficient response data"
	}

	// Get decision from first non-empty line
	var decision string
	for _, line := range lines {
		if trimmed := string(bytes.TrimSpace(line)); trimmed != "" {
			decision = trimmed
			break
		}
	}

	switch decision {
	case "BUY", "SELL":
		// Valid trading decision
	case "NOTHING":
		return "NOTHING", 0.1, "Market conditions unfavorable"
	default:
		return "NOTHING", 0.1, "Market analysis inconclusive"
	}

	// Find confidence value (first number between 0.1 and 0.9)
	var confidence float64
	for _, line := range lines {
		trimmed := string(bytes.TrimSpace(line))
		if _, err := fmt.Sscanf(trimmed, "%f", &confidence); err == nil {
			if confidence >= 0.1 && confidence <= 0.9 {
				break
			}
		}
	}

	if confidence < 0.1 {
		confidence = 0.3 // Default trading confidence
	}

	// Extract reasoning (all non-empty lines after decision and confidence)
	var reasoningLines []string
	foundConfidence := false
	for _, line := range lines {
		trimmed := string(bytes.TrimSpace(line))
		if trimmed == "" || trimmed == decision {
			continue
		}
		
		var testConf float64
		if _, err := fmt.Sscanf(trimmed, "%f", &testConf); err == nil {
			foundConfidence = true
			continue
		}
		
		if foundConfidence && trimmed != "" {
			reasoningLines = append(reasoningLines, trimmed)
		}
	}

	reasoning := strings.Join(reasoningLines, " ")
	if reasoning == "" {
		reasoning = "Analysis based on current market conditions"
	}
	
	return decision, confidence, reasoning
}
