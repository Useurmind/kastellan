package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/kastellan/kastellan/pkg/agentprotocol"
)

// Session represents an active agent connection session.
type Session struct {
	ID                   string
	ExternalHost         string
	AgentID              string
	AgentVersion         string
	ProtocolVersion      string
	RuntimeName          string
	RuntimeVersion       string
	Capabilities         []string
	ConnectedAt          time.Time
	LastHeartbeat        time.Time
	LastSeen             time.Time
	LastReceivedRevision int64
	LastAppliedRevision  int64
	CurrentRevision      int64
	HostStatus           struct {
		RuntimeAvailable bool
		Workloads        struct {
			Assigned int
			Ready    int
			Failed   int
			Updating int
			Unknown  int
		}
	}
	closeCh chan struct{}
	mu      sync.RWMutex
}

// NewSession creates a new session.
func NewSession(id, externalHost, agentID, agentVersion string) *Session {
	return &Session{
		ID:                   id,
		ExternalHost:         externalHost,
		AgentID:              agentID,
		AgentVersion:         agentVersion,
		ProtocolVersion:      ProtocolVersionV1Alpha1,
		ConnectedAt:          time.Now(),
		LastHeartbeat:        time.Now(),
		LastSeen:             time.Now(),
		CurrentRevision:      0,
		LastReceivedRevision: 0,
		LastAppliedRevision:  0,
		closeCh:              make(chan struct{}),
	}
}

// UpdateHeartbeat updates the last heartbeat time and workload status.
func (s *Session) UpdateHeartbeat(workloads struct {
	Assigned int
	Ready    int
	Failed   int
	Updating int
	Unknown  int
}, runtimeReady bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastHeartbeat = time.Now()
	s.LastSeen = time.Now()
	s.HostStatus.RuntimeAvailable = runtimeReady
	s.HostStatus.Workloads.Assigned = workloads.Assigned
	s.HostStatus.Workloads.Ready = workloads.Ready
	s.HostStatus.Workloads.Failed = workloads.Failed
	s.HostStatus.Workloads.Updating = workloads.Updating
	s.HostStatus.Workloads.Unknown = workloads.Unknown
}

// UpdateRevisions updates the revision tracking.
func (s *Session) UpdateRevisions(received, applied int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastReceivedRevision = received
	s.LastAppliedRevision = applied
}

// GetCurrentRevision returns the current revision being sent.
func (s *Session) GetCurrentRevision() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentRevision
}

// SetCurrentRevision sets the current revision being sent.
func (s *Session) SetCurrentRevision(revision int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentRevision = revision
}

// GetLastAppliedRevision returns the last applied revision.
func (s *Session) GetLastAppliedRevision() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastAppliedRevision
}

// IsConnected returns whether the session is still active.
func (s *Session) IsConnected() bool {
	select {
	case <-s.closeCh:
		return false
	default:
		return true
	}
}

// Close marks the session as closed.
func (s *Session) Close() {
	select {
	case <-s.closeCh:
	default:
		close(s.closeCh)
	}
}

// SessionManager manages active agent sessions.
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session and stores it.
func (sm *SessionManager) CreateSession(session *Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[session.ID]; exists {
		return agentprotocol.NewProtocolError(
			ErrorCodeSessionExists,
			fmt.Sprintf("session %s already exists", session.ID),
		)
	}

	for _, s := range sm.sessions {
		if s.ExternalHost == session.ExternalHost && s.ID != session.ID {
			s.Close()
			delete(sm.sessions, s.ID)
		}
	}

	sm.sessions[session.ID] = session
	return nil
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	return session, exists
}

// GetSessionByHost retrieves a session by host name.
func (sm *SessionManager) GetSessionByHost(host string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.sessions {
		if session.ExternalHost == host {
			return session, true
		}
	}

	return nil, false
}

// ListSessions returns all active sessions.
func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetConnectedHosts returns the count of connected hosts.
func (sm *SessionManager) GetConnectedHosts() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, session := range sm.sessions {
		if session.IsConnected() {
			count++
		}
	}

	return count
}
