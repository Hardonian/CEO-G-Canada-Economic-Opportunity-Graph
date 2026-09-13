package middleware

import (
	"context"
	"math/rand"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// RetryConfig holds configuration for retry-with-backoff.
type RetryConfig struct {
	MaxAttempts  int           // including the first try; zero => 3
	BaseDelay    time.Duration // zero => 100ms
	MaxDelay     time.Duration // zero => 2s
	JitterFactor float64       // fraction of delay; zero => 0.1
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: 2 * time.Second, JitterFactor: 0.1}
}

// Retry returns a middleware that retries failed Fetch calls with
// exponential backoff + jitter. Parse is delegated directly.
func Retry(config RetryConfig) AdapterMiddleware {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 2 * time.Second
	}
	if config.JitterFactor <= 0 {
		config.JitterFactor = 0.1
	}
	return func(next adapters.Adapter) adapters.Adapter {
		return retryAdapter{next: next, config: config}
	}
}

type retryAdapter struct {
	next   adapters.Adapter
	config RetryConfig
}

func (r retryAdapter) Name() string { return r.next.Name() }
func (r retryAdapter) Tier() adapters.SourceTier { return r.next.Tier() }
func (r retryAdapter) Health() *adapters.SourceHealth { return r.next.Health() }

func (r retryAdapter) Fetch(ctx context.Context) ([]byte, error) {
	var lastErr error
	delay := r.config.BaseDelay
	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		data, err := r.next.Fetch(ctx)
		if err == nil {
			return data, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if attempt == r.config.MaxAttempts-1 {
			break
		}
		jitter := float64(delay) * r.config.JitterFactor * (2*rand.Float64() - 1)
		sleep := time.Duration(float64(delay) + jitter)
		if sleep < 0 {
			sleep = 0
		}
		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		delay *= 2
		if delay > r.config.MaxDelay {
			delay = r.config.MaxDelay
		}
	}
	return nil, lastErr
}

func (r retryAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return r.next.Parse(data)
}