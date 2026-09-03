package server

import (
	"sync"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol/messages"
)

// HeartbeatProcessor processes heartbeat messages from agents.
type HeartbeatProcessor struct {
	mu         sync.RWMutex
	heartbeats map[string]*HeartbeatState
	timeout    time.Duration
}

// HeartbeatState tracks heartbeat information for a session.
type HeartbeatState struct {
	LastHeartbeat time.Time
	LastSeen      time.Time
	Workloads     struct {
		Assigned int
		Ready    int
		Failed   int
		Updating int
		Unknown  int
	}
	RuntimeAvailable bool
}

// NewHeartbeatProcessor creates a new heartbeat processor.
func NewHeartbeatProcessor() *HeartbeatProcessor {
	return &HeartbeatProcessor{
		heartbeats: make(map[string]*HeartbeatState),
		timeout:    2 * time.Minute,
	}
}

// WithTimeout sets the heartbeat timeout.
func (p *HeartbeatProcessor) WithTimeout(timeout time.Duration) *HeartbeatProcessor {
	p.timeout = timeout
	return p
}

// ProcessHeartbeat processes a heartbeat message.
func (p *HeartbeatProcessor) ProcessHeartbeat(sessionID string, msg *messages.Heartbeat) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	state, exists := p.heartbeats[sessionID]
	if !exists {
		state = &HeartbeatState{}
		p.heartbeats[sessionID] = state
	}

	now := time.Now()
	state.LastHeartbeat = now
	state.LastSeen = now

	// Update runtime status
	state.RuntimeAvailable = msg.Runtime.Available

	// Update workload counts
	state.Workloads.Assigned = msg.Workloads.Assigned
	state.Workloads.Ready = msg.Workloads.Ready
	state.Workloads.Failed = msg.Workloads.Failed
	state.Workloads.Updating = msg.Workloads.Updating
	state.Workloads.Unknown = msg.Workloads.Unknown

	return nil
}

// IsSessionAlive checks if a session is still alive based on heartbeats.
func (p *HeartbeatProcessor) IsSessionAlive(sessionID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, exists := p.heartbeats[sessionID]
	if !exists {
		return false
	}

	return time.Since(state.LastHeartbeat) <= p.timeout
}

// GetSessionState returns the heartbeat state for a session.
func (p *HeartbeatProcessor) GetSessionState(sessionID string) (*HeartbeatState, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, exists := p.heartbeats[sessionID]
	return state, exists
}

// GetDisconnectedSessions returns sessions that have not sent heartbeats recently.
func (p *HeartbeatProcessor) GetDisconnectedSessions() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var disconnected []string
	now := time.Now()

	for sessionID, state := range p.heartbeats {
		if now.Sub(state.LastHeartbeat) > p.timeout {
			disconnected = append(disconnected, sessionID)
		}
	}

	return disconnected
}

// CleanupOldSessions removes sessions that have been disconnected for too long.
func (p *HeartbeatProcessor) CleanupOldSessions(maxAge time.Duration) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var cleaned []string
	now := time.Now()

	for sessionID, state := range p.heartbeats {
		if now.Sub(state.LastSeen) > maxAge {
			delete(p.heartbeats, sessionID)
			cleaned = append(cleaned, sessionID)
		}
	}

	return cleaned
}
