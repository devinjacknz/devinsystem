package models

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/stretchr/testify/assert"
)

func TestOllamaIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/chat", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"message": {
				"content": "BUY\nconfidence: 85\nStrong buy signal based on volume increase"
			}
		}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-r1")
	marketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	decision, err := client.GenerateTradeDecision(context.Background(), marketData)
	assert.NoError(t, err)
	assert.Equal(t, "BUY", decision.Action)
	assert.Equal(t, 85.0, decision.Confidence)
	assert.Contains(t, decision.Reasoning, "Strong buy signal")
}

func TestOllamaErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "deepseek-r1")
	marketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	_, err := client.GenerateTradeDecision(context.Background(), marketData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}
