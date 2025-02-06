package exchange

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitedClient struct {
	limiter *rate.Limiter
	client  *http.Client
}

func NewRateLimitedClient(rps float64) *RateLimitedClient {
	return &RateLimitedClient{
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	err := c.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *RateLimitedClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *RateLimitedClient) Post(url string, contentType string, body []byte) (*http.Response, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
