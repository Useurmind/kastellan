// Package client provides heartbeat handling for the Kastellan Agent Protocol.
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/useurmind/kastellan/pkg/agentprotocol/messages"
)

// HeartbeatManager manages heartbeat sending.
type HeartbeatManager struct {
	// Configuration
	interval time.Duration
	timeout  time.Duration

	// State
	mu          sync.Mutex
	running     bool
	stopChan    chan struct{}
	wg          sync.WaitGroup
	lastSend    time.Time
	lastAck     time.Time
	missedCount int

	// Callbacks
	sendFunc  func(ctx context.Context) error
	ackFunc   func()
	errorFunc func(error)
}

// HeartbeatConfig configures the heartbeat manager.
type HeartbeatConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

// DefaultHeartbeatConfig returns the default heartbeat configuration.
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}
}

// NewHeartbeatManager creates a new heartbeat manager.
func NewHeartbeatManager(config *HeartbeatConfig) *HeartbeatManager {
	if config == nil {
		config = DefaultHeartbeatConfig()
	}

	return &HeartbeatManager{
		interval: config.Interval,
		timeout:  config.Timeout,
		stopChan: make(chan struct{}),
	}
}

// WithSendFunc sets the function to send heartbeats.
func (h *HeartbeatManager) WithSendFunc(fn func(ctx context.Context) error) *HeartbeatManager {
	h.sendFunc = fn
	return h
}

// WithAckFunc sets the function to handle heartbeat acknowledgments.
func (h *HeartbeatManager) WithAckFunc(fn func()) *HeartbeatManager {
	h.ackFunc = fn
	return h
}

// WithErrorFunc sets the function to handle errors.
func (h *HeartbeatManager) WithErrorFunc(fn func(error)) *HeartbeatManager {
	h.errorFunc = fn
	return h
}

// Start starts the heartbeat manager.
func (h *HeartbeatManager) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("heartbeat already running")
	}
	h.running = true
	h.mu.Unlock()

	h.wg.Add(1)
	go h.run(ctx)

	return nil
}

// Stop stops the heartbeat manager.
func (h *HeartbeatManager) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()

	close(h.stopChan)
	h.wg.Wait()
}

// run runs the heartbeat loop.
func (h *HeartbeatManager) run(ctx context.Context) {
	defer h.wg.Done()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			if err := h.sendHeartbeat(ctx); err != nil {
				if h.errorFunc != nil {
					h.errorFunc(err)
				}
			}
		}
	}
}

// sendHeartbeat sends a heartbeat message.
func (h *HeartbeatManager) sendHeartbeat(ctx context.Context) error {
	h.mu.Lock()
	h.lastSend = time.Now()
	h.mu.Unlock()

	if h.sendFunc == nil {
		return fmt.Errorf("send function not set")
	}

	if err := h.sendFunc(ctx); err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return nil
}

// RecordAck records a heartbeat acknowledgment.
func (h *HeartbeatManager) RecordAck() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastAck = time.Now()
	h.missedCount = 0
}

// RecordMiss records a missed heartbeat.
func (h *HeartbeatManager) RecordMiss() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.missedCount++
}

// IsConnected returns whether the heartbeat connection is healthy.
func (h *HeartbeatManager) IsConnected() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if we've missed too many heartbeats
	if h.missedCount > 3 {
		return false
	}

	// Check if we've received an acknowledgment recently
	if h.lastAck.IsZero() {
		return false
	}

	return time.Since(h.lastAck) < h.timeout*3
}

// GetLastSend returns the last send time.
func (h *HeartbeatManager) GetLastSend() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastSend
}

// GetLastAck returns the last acknowledgment time.
func (h *HeartbeatManager) GetLastAck() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastAck
}

// GetMissedCount returns the number of missed heartbeats.
func (h *HeartbeatManager) GetMissedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.missedCount
}

// CreateHeartbeatMessage creates a heartbeat message.
func CreateHeartbeatMessage(sessionID, hostName string) *messages.Heartbeat {
	return &messages.Heartbeat{
		Type:      messages.MessageTypeHeartbeat,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Runtime: struct {
			Available bool   `json:"available"`
			Error     string `json:"error,omitempty"`
		}{
			Available: true,
		},
		Workloads: struct {
			Assigned int `json:"assigned"`
			Ready    int `json:"ready"`
			Failed   int `json:"failed"`
			Updating int `json:"updating,omitempty"`
			Unknown  int `json:"unknown,omitempty"`
		}{
			Assigned: 0,
			Ready:    0,
			Failed:   0,
		},
		Host: struct {
			Name string `json:"name"`
		}{
			Name: hostName,
		},
	}
}

// HeartbeatSender is an interface for sending heartbeat messages.
type HeartbeatSender interface {
	SendHeartbeat(ctx context.Context, msg *messages.Heartbeat) error
}

// NewHeartbeatSender creates a new heartbeat sender.
func NewHeartbeatSender() HeartbeatSender {
	return &heartbeatSender{}
}

type heartbeatSender struct{}

// SendHeartbeat sends a heartbeat message.
func (s *heartbeatSender) SendHeartbeat(ctx context.Context, msg *messages.Heartbeat) error {
	// This would use the actual gRPC stream to send the message
	// For now, we'll just return nil
	return nil
}
