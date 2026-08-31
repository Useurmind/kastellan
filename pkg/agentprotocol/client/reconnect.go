// Package client provides reconnection logic for the Kastellan Agent Protocol.
package client

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Reconnector handles reconnection with exponential backoff.
type Reconnector struct {
	// Configuration
	initialDelay time.Duration
	maxDelay     time.Duration
	maxRetries   int

	// State
	currentDelay time.Duration
	retryCount   int
	mu           sync.Mutex

	// Callbacks
	onReconnect func() error
	onFail      func(error)
}

// ReconnectorConfig configures the reconnector.
type ReconnectorConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	MaxRetries   int
}

// DefaultReconnectorConfig returns the default reconnector configuration.
func DefaultReconnectorConfig() *ReconnectorConfig {
	return &ReconnectorConfig{
		InitialDelay: time.Second,
		MaxDelay:     time.Minute,
		MaxRetries:   0, // 0 means infinite retries
	}
}

// NewReconnector creates a new reconnector.
func NewReconnector(config *ReconnectorConfig) *Reconnector {
	if config == nil {
		config = DefaultReconnectorConfig()
	}

	return &Reconnector{
		initialDelay: config.InitialDelay,
		maxDelay:     config.MaxDelay,
		maxRetries:   config.MaxRetries,
		currentDelay: config.InitialDelay,
	}
}

// WithOnReconnect sets the callback for successful reconnection.
func (r *Reconnector) WithOnReconnect(callback func() error) *Reconnector {
	r.onReconnect = callback
	return r
}

// WithOnFail sets the callback for reconnection failure.
func (r *Reconnector) WithOnFail(callback func(error)) *Reconnector {
	r.onFail = callback
	return r
}

// NextDelay returns the next delay with jitter.
func (r *Reconnector) NextDelay() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add jitter (0-50% of current delay)
	jitter := time.Duration(rand.Float64() * float64(r.currentDelay) * 0.5)
	delay := r.currentDelay + jitter

	// Exponential backoff
	r.currentDelay = min(r.currentDelay*2, r.maxDelay)

	return delay
}

// ResetDelay resets the delay to the initial value.
func (r *Reconnector) ResetDelay() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.currentDelay = r.initialDelay
	r.retryCount = 0
}

// ShouldRetry returns whether we should retry.
func (r *Reconnector) ShouldRetry() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.maxRetries == 0 {
		return true
	}

	return r.retryCount < r.maxRetries
}

// IncrementRetry increments the retry count.
func (r *Reconnector) IncrementRetry() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.retryCount++
}

// ReconnectLoop runs the reconnection loop.
func (r *Reconnector) ReconnectLoop(ctx context.Context, connectFunc func(context.Context) error) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := connectFunc(ctx); err != nil {
			if r.onFail != nil {
				r.onFail(err)
			}

			if !r.ShouldRetry() {
				return fmt.Errorf("reconnection failed after %d attempts: %w", r.retryCount, err)
			}

			nextDelay := r.NextDelay()
			fmt.Printf("Connection failed: %v, retrying in %v\n", err, nextDelay)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(nextDelay):
				// Continue to next retry
			}
			continue
		}

		// Connection successful
		if r.onReconnect != nil {
			if err := r.onReconnect(); err != nil {
				return fmt.Errorf("reconnect callback failed: %w", err)
			}
		}

		r.ResetDelay()

		// Wait for connection to be closed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.waitConnectionClosed(ctx):
			// Connection was closed, restart
			continue
		}
	}
}

// waitConnectionClosed waits for the connection to be closed.
func (r *Reconnector) waitConnectionClosed(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		// In a real implementation, this would wait for a connection close event
		// For now, we'll just return immediately
		<-ctx.Done()
	}()

	return ch
}

// ReconnectWithBackoff runs a function with exponential backoff.
func ReconnectWithBackoff(ctx context.Context, fn func(context.Context) error, config *ReconnectorConfig) error {
	reconnector := NewReconnector(config)
	return reconnector.ReconnectLoop(ctx, fn)
}

// BackoffDelay calculates the backoff delay with jitter.
func BackoffDelay(attempt int, initial, max time.Duration) time.Duration {
	// Exponential backoff: initial * 2^attempt
	delay := initial * time.Duration(1<<attempt)

	// Cap at max
	if delay > max {
		delay = max
	}

	// Add jitter (0-50% of delay)
	jitter := time.Duration(rand.Float64() * float64(delay) * 0.5)
	delay += jitter

	return delay
}

// CalculateBackoffDelay calculates the delay for a given attempt number.
func CalculateBackoffDelay(attempt int, initial, max time.Duration) time.Duration {
	return BackoffDelay(attempt, initial, max)
}

// min returns the minimum of two durations.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// init initializes the random seed.
func init() {
	rand.Seed(time.Now().UnixNano())
}
